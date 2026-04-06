package messaging

import (
	"sync"
	"time"
)

const (
	// ReportsForWarning is the threshold for issuing a warning
	ReportsForWarning = 1
	// ReportsForSuspension is the threshold for suspending a feature
	ReportsForSuspension = 3
	// SuspensionDuration is how long a feature is suspended after abuse
	SuspensionDuration = 7 * 24 * time.Hour
	// TrustPenaltyPerReport is the trust score deduction per abuse report
	TrustPenaltyPerReport = 5
	// TrustPenaltyScheduledReport is the deduction for a reported scheduled message
	TrustPenaltyScheduledReport = 3
)

// AbuseReportType categorizes the type of abuse
type AbuseReportType string

const (
	AbuseTypeSilentSpam     AbuseReportType = "silent_spam"
	AbuseTypeScheduledSpam  AbuseReportType = "scheduled_spam"
	AbuseTypeHarassment     AbuseReportType = "harassment"
	AbuseTypeInappropriate  AbuseReportType = "inappropriate_content"
)

// AbuseReport represents a single abuse report
type AbuseReport struct {
	ID           string
	ReporterID   string
	ReportedID   string
	MessageID    string
	Type         AbuseReportType
	Description  string
	CreatedAt    time.Time
}

// Suspension represents a feature suspension
type Suspension struct {
	UserID    string
	Feature   string // "silent" or "scheduled"
	Reason    string
	StartedAt time.Time
	ExpiresAt time.Time
}

// SilentBlockEntry represents a user blocking silent messages from a contact
type SilentBlockEntry struct {
	UserID    string
	BlockedID string
	CreatedAt time.Time
}

// AbuseTracker monitors and enforces abuse prevention rules
type AbuseTracker struct {
	mu          sync.RWMutex
	reports     map[string][]*AbuseReport   // reportedID -> reports
	suspensions map[string][]*Suspension    // userID -> suspensions
	silentBlocks map[string]map[string]bool // userID -> set of blocked senderIDs
}

// NewAbuseTracker creates a new abuse tracker
func NewAbuseTracker() *AbuseTracker {
	return &AbuseTracker{
		reports:      make(map[string][]*AbuseReport),
		suspensions:  make(map[string][]*Suspension),
		silentBlocks: make(map[string]map[string]bool),
	}
}

// Report files an abuse report against a user
func (a *AbuseTracker) Report(report *AbuseReport) (*Suspension, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	report.CreatedAt = time.Now()
	a.reports[report.ReportedID] = append(a.reports[report.ReportedID], report)

	// Count unique reporters for the relevant feature
	feature := "silent"
	if report.Type == AbuseTypeScheduledSpam {
		feature = "scheduled"
	}

	uniqueReporters := a.countUniqueReporters(report.ReportedID, feature)

	// Auto-suspend if threshold met
	if uniqueReporters >= ReportsForSuspension {
		suspension := &Suspension{
			UserID:    report.ReportedID,
			Feature:   feature,
			Reason:    "multiple abuse reports",
			StartedAt: time.Now(),
			ExpiresAt: time.Now().Add(SuspensionDuration),
		}
		a.suspensions[report.ReportedID] = append(a.suspensions[report.ReportedID], suspension)
		return suspension, nil
	}

	return nil, nil
}

// countUniqueReporters counts distinct users who reported a user for a feature
func (a *AbuseTracker) countUniqueReporters(reportedID, feature string) int {
	seen := make(map[string]bool)
	for _, r := range a.reports[reportedID] {
		isSilent := r.Type == AbuseTypeSilentSpam || r.Type == AbuseTypeHarassment || r.Type == AbuseTypeInappropriate
		isScheduled := r.Type == AbuseTypeScheduledSpam

		if (feature == "silent" && isSilent) || (feature == "scheduled" && isScheduled) {
			seen[r.ReporterID] = true
		}
	}
	return len(seen)
}

// IsSuspended checks if a user's feature is currently suspended
func (a *AbuseTracker) IsSuspended(userID, feature string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	now := time.Now()
	for _, s := range a.suspensions[userID] {
		if s.Feature == feature && now.Before(s.ExpiresAt) {
			return true
		}
	}
	return false
}

// GetActiveSuspension returns the active suspension for a user's feature, if any
func (a *AbuseTracker) GetActiveSuspension(userID, feature string) *Suspension {
	a.mu.RLock()
	defer a.mu.RUnlock()

	now := time.Now()
	for _, s := range a.suspensions[userID] {
		if s.Feature == feature && now.Before(s.ExpiresAt) {
			return s
		}
	}
	return nil
}

// BlockSilent adds a sender to a user's silent message block list
func (a *AbuseTracker) BlockSilent(userID, senderID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.silentBlocks[userID] == nil {
		a.silentBlocks[userID] = make(map[string]bool)
	}
	a.silentBlocks[userID][senderID] = true
}

// UnblockSilent removes a sender from a user's silent message block list
func (a *AbuseTracker) UnblockSilent(userID, senderID string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.silentBlocks[userID] != nil {
		delete(a.silentBlocks[userID], senderID)
	}
}

// IsSilentBlocked checks if a sender is blocked from sending silent messages to a user
func (a *AbuseTracker) IsSilentBlocked(userID, senderID string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if blocks, ok := a.silentBlocks[userID]; ok {
		return blocks[senderID]
	}
	return false
}

// GetReportCount returns the total reports against a user
func (a *AbuseTracker) GetReportCount(userID string) int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return len(a.reports[userID])
}

// GetReports returns all reports against a user
func (a *AbuseTracker) GetReports(userID string) []*AbuseReport {
	a.mu.RLock()
	defer a.mu.RUnlock()

	src := a.reports[userID]
	reports := make([]*AbuseReport, len(src))
	for i, r := range src {
		copied := *r
		reports[i] = &copied
	}
	return reports
}

// CalculateTrustPenalty returns the trust score penalty for a user's abuse reports
func (a *AbuseTracker) CalculateTrustPenalty(userID string) int {
	a.mu.RLock()
	defer a.mu.RUnlock()

	penalty := 0
	for _, r := range a.reports[userID] {
		switch r.Type {
		case AbuseTypeSilentSpam, AbuseTypeHarassment, AbuseTypeInappropriate:
			penalty += TrustPenaltyPerReport
		case AbuseTypeScheduledSpam:
			penalty += TrustPenaltyScheduledReport
		}
	}
	return penalty
}
