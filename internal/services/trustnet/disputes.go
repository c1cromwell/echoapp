package trustnet

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// DisputeType represents the category of dispute
type DisputeType string

const (
	DisputeFalseReport      DisputeType = "false_report"
	DisputeCoordinatedAttack DisputeType = "coordinated_attack"
	DisputeSystemError      DisputeType = "system_error"
	DisputeAccountCompromise DisputeType = "account_compromise"
)

// DisputeStatus represents the current state of a dispute
type DisputeStatus string

const (
	DisputeOpen     DisputeStatus = "open"
	DisputeReview   DisputeStatus = "in_review"
	DisputeUpheld   DisputeStatus = "upheld"
	DisputeRejected DisputeStatus = "rejected"
	DisputeExpired  DisputeStatus = "expired"
)

// VoteDecision represents a juror's vote
type VoteDecision string

const (
	VoteUphold  VoteDecision = "uphold"
	VoteReject  VoteDecision = "reject"
	VoteAbstain VoteDecision = "abstain"
)

const (
	// JurorCount is the number of jurors assigned to each dispute
	JurorCount = 5

	// MinJurorTrustScore is the minimum trust score to be a juror
	MinJurorTrustScore = 80

	// ReviewPeriod is the time allowed for juror review
	ReviewPeriod = 72 * time.Hour

	// DisputeCooldown is the minimum time between disputes for the same user
	DisputeCooldown = 90 * 24 * time.Hour

	// JurorTrustBonus is the trust score bonus for juror participation
	JurorTrustBonus = 2.0
)

// Dispute represents a trust dispute
type Dispute struct {
	ID           string
	FiledBy      string
	AgainstDID   string // the entity/system being disputed
	Type         DisputeType
	Status       DisputeStatus
	Evidence     string
	Jurors       []string // juror DIDs
	Votes        map[string]JurorVote
	CreatedAt    time.Time
	ExpiresAt    time.Time
	ResolvedAt   *time.Time
	StakeAmount  float64
	StakeRefunded bool
}

// JurorVote represents a single juror's vote
type JurorVote struct {
	JurorDID  string
	Decision  VoteDecision
	Reasoning string
	VotedAt   time.Time
}

// DisputeService manages trust disputes
type DisputeService struct {
	mu       sync.RWMutex
	disputes map[string]*Dispute  // disputeID -> dispute
	byUser   map[string][]string  // userDID -> []disputeID
}

// NewDisputeService creates a new dispute service
func NewDisputeService() *DisputeService {
	return &DisputeService{
		disputes: make(map[string]*Dispute),
		byUser:   make(map[string][]string),
	}
}

// FileDispute creates a new dispute
func (s *DisputeService) FileDispute(filedBy, againstDID string, disputeType DisputeType, evidence string, stake float64) (*Dispute, error) {
	if !isValidDisputeType(disputeType) {
		return nil, ErrDisputeInvalidType
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check 90-day cooldown
	for _, id := range s.byUser[filedBy] {
		d := s.disputes[id]
		if time.Since(d.CreatedAt) < DisputeCooldown {
			return nil, ErrDisputeRateLimited
		}
	}

	dispute := &Dispute{
		ID:          uuid.New().String(),
		FiledBy:     filedBy,
		AgainstDID:  againstDID,
		Type:        disputeType,
		Status:      DisputeOpen,
		Evidence:    evidence,
		Votes:       make(map[string]JurorVote),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ReviewPeriod),
		StakeAmount: stake,
	}

	s.disputes[dispute.ID] = dispute
	s.byUser[filedBy] = append(s.byUser[filedBy], dispute.ID)
	return dispute, nil
}

// AssignJurors assigns jurors to a dispute
func (s *DisputeService) AssignJurors(disputeID string, jurorDIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.disputes[disputeID]
	if !ok {
		return ErrDisputeNotFound
	}

	if d.Status != DisputeOpen {
		return ErrDisputeAlreadyResolved
	}

	// Validate jurors don't have conflicts
	for _, juror := range jurorDIDs {
		if juror == d.FiledBy || juror == d.AgainstDID {
			return ErrJurorConflict
		}
	}

	d.Jurors = jurorDIDs
	d.Status = DisputeReview
	return nil
}

// CastVote records a juror's vote
func (s *DisputeService) CastVote(disputeID, jurorDID string, decision VoteDecision, reasoning string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.disputes[disputeID]
	if !ok {
		return ErrDisputeNotFound
	}

	if d.Status != DisputeReview {
		return ErrDisputeAlreadyResolved
	}

	// Check juror is assigned
	isJuror := false
	for _, j := range d.Jurors {
		if j == jurorDID {
			isJuror = true
			break
		}
	}
	if !isJuror {
		return ErrJurorIneligible
	}

	// Check not already voted
	if _, voted := d.Votes[jurorDID]; voted {
		return ErrJurorAlreadyVoted
	}

	d.Votes[jurorDID] = JurorVote{
		JurorDID:  jurorDID,
		Decision:  decision,
		Reasoning: reasoning,
		VotedAt:   time.Now(),
	}

	// Check if all jurors have voted — auto-resolve
	if len(d.Votes) == len(d.Jurors) {
		s.resolveLocked(d)
	}

	return nil
}

// resolveLocked resolves a dispute based on votes (must hold write lock)
func (s *DisputeService) resolveLocked(d *Dispute) {
	upholdCount := 0
	rejectCount := 0

	for _, v := range d.Votes {
		switch v.Decision {
		case VoteUphold:
			upholdCount++
		case VoteReject:
			rejectCount++
		}
	}

	now := time.Now()
	d.ResolvedAt = &now

	if upholdCount > rejectCount {
		d.Status = DisputeUpheld
		d.StakeRefunded = true
	} else {
		d.Status = DisputeRejected
		d.StakeRefunded = false
	}
}

// ResolveExpired checks and resolves expired disputes
func (s *DisputeService) ResolveExpired() []*Dispute {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expired []*Dispute
	now := time.Now()

	for _, d := range s.disputes {
		if d.Status == DisputeReview && now.After(d.ExpiresAt) {
			// If some votes exist, resolve based on what we have
			if len(d.Votes) > 0 {
				s.resolveLocked(d)
			} else {
				d.Status = DisputeExpired
				d.ResolvedAt = &now
				d.StakeRefunded = true // refund if no jurors voted
			}
			expired = append(expired, d)
		}
	}
	return expired
}

// GetDispute retrieves a dispute by ID
func (s *DisputeService) GetDispute(id string) (*Dispute, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.disputes[id]
	if !ok {
		return nil, ErrDisputeNotFound
	}
	return d, nil
}

// GetUserDisputes returns all disputes filed by a user
func (s *DisputeService) GetUserDisputes(userDID string) []*Dispute {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Dispute
	for _, id := range s.byUser[userDID] {
		if d, ok := s.disputes[id]; ok {
			result = append(result, d)
		}
	}
	return result
}

// GetActiveDisputes returns all disputes currently in review
func (s *DisputeService) GetActiveDisputes() []*Dispute {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Dispute
	for _, d := range s.disputes {
		if d.Status == DisputeOpen || d.Status == DisputeReview {
			result = append(result, d)
		}
	}
	return result
}

// GetJurorDisputes returns disputes assigned to a juror
func (s *DisputeService) GetJurorDisputes(jurorDID string) []*Dispute {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Dispute
	for _, d := range s.disputes {
		for _, j := range d.Jurors {
			if j == jurorDID {
				result = append(result, d)
				break
			}
		}
	}
	return result
}

// VoteCount returns the current vote tallies for a dispute
func (s *DisputeService) VoteCount(disputeID string) (uphold, reject, abstain int, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.disputes[disputeID]
	if !ok {
		return 0, 0, 0, ErrDisputeNotFound
	}

	for _, v := range d.Votes {
		switch v.Decision {
		case VoteUphold:
			uphold++
		case VoteReject:
			reject++
		case VoteAbstain:
			abstain++
		}
	}
	return
}

func isValidDisputeType(dt DisputeType) bool {
	switch dt {
	case DisputeFalseReport, DisputeCoordinatedAttack, DisputeSystemError, DisputeAccountCompromise:
		return true
	}
	return false
}
