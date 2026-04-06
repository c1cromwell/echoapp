package onboarding

import (
	"sync"
	"time"
)

// OnboardingAnalyticsEvent represents an onboarding event
type OnboardingAnalyticsEvent struct {
	EventType         string // step_started, step_completed, step_skipped, session_completed, etc.
	SessionID         string
	UserID            string
	StepName          string
	Timestamp         time.Time
	DurationMS        int64 // How long the step took
	Data              map[string]interface{}
	IPAddress         string
	UserAgent         string
	DeviceFingerprint string
}

// OnboardingFunnelMetrics tracks conversion at each step
type OnboardingFunnelMetrics struct {
	TotalSessionsStarted   int
	CarouselViewed         int
	PhoneEntryStarted      int
	PhoneVerified          int
	PasskeySetup           int
	RecoverySetup          int
	ProfileSetup           int
	CompletedOnboarding    int
	SkippedAtSteps         map[string]int // step -> count
	AverageSessionDuration time.Duration
	ConversionRate         float64 // 0.0-1.0
}

// CredentialUsageStats tracks which credentials are used
type CredentialUsageStats struct {
	CredentialType      CredentialType
	Count               int
	AverageTrustScore   float64
	IssuerDistribution  map[string]int // issuer -> count
	VerificationSuccess int
	VerificationFailure int
}

// OnboardingAnalyticsService tracks onboarding metrics and events
type OnboardingAnalyticsService struct {
	mu              sync.RWMutex
	events          []OnboardingAnalyticsEvent
	sessionMetrics  map[string]*OnboardingFunnelMetrics
	credentialStats map[CredentialType]*CredentialUsageStats
	eventBuffer     chan OnboardingAnalyticsEvent
	stopChan        chan bool
}

// NewOnboardingAnalyticsService creates the analytics service
func NewOnboardingAnalyticsService() *OnboardingAnalyticsService {
	oas := &OnboardingAnalyticsService{
		events:          make([]OnboardingAnalyticsEvent, 0),
		sessionMetrics:  make(map[string]*OnboardingFunnelMetrics),
		credentialStats: make(map[CredentialType]*CredentialUsageStats),
		eventBuffer:     make(chan OnboardingAnalyticsEvent, 100),
		stopChan:        make(chan bool),
	}

	// Start background event processor
	go oas.processEvents()

	return oas
}

// RecordEvent logs an onboarding event
func (oas *OnboardingAnalyticsService) RecordEvent(event OnboardingAnalyticsEvent) {
	event.Timestamp = time.Now()
	select {
	case oas.eventBuffer <- event:
	case <-oas.stopChan:
	default:
		// Buffer full, log synchronously
		oas.logEvent(event)
	}
}

// RecordStepStarted logs when a user starts a step
func (oas *OnboardingAnalyticsService) RecordStepStarted(sessionID, userID, stepName string) {
	oas.RecordEvent(OnboardingAnalyticsEvent{
		EventType: "step_started",
		SessionID: sessionID,
		UserID:    userID,
		StepName:  stepName,
	})
}

// RecordStepCompleted logs when a user completes a step
func (oas *OnboardingAnalyticsService) RecordStepCompleted(sessionID, userID, stepName string, durationMS int64) {
	oas.RecordEvent(OnboardingAnalyticsEvent{
		EventType:  "step_completed",
		SessionID:  sessionID,
		UserID:     userID,
		StepName:   stepName,
		DurationMS: durationMS,
	})
}

// RecordStepSkipped logs when a user skips a step
func (oas *OnboardingAnalyticsService) RecordStepSkipped(sessionID, userID, stepName string) {
	oas.RecordEvent(OnboardingAnalyticsEvent{
		EventType: "step_skipped",
		SessionID: sessionID,
		UserID:    userID,
		StepName:  stepName,
	})
}

// RecordOnboardingCompleted logs session completion
func (oas *OnboardingAnalyticsService) RecordOnboardingCompleted(
	sessionID, userID string,
	durationMS int64,
	credentialsUsed []CredentialType,
) {
	data := map[string]interface{}{
		"credentials_count": len(credentialsUsed),
		"credentials_used":  credentialsUsed,
	}

	oas.RecordEvent(OnboardingAnalyticsEvent{
		EventType:  "session_completed",
		SessionID:  sessionID,
		UserID:     userID,
		DurationMS: durationMS,
		Data:       data,
	})
}

// RecordCredentialVerification logs credential verification result
func (oas *OnboardingAnalyticsService) RecordCredentialVerification(
	sessionID, userID string,
	credType CredentialType,
	issuer *TrustedIssuer,
	success bool,
) {
	data := map[string]interface{}{
		"credential_type": credType,
		"success":         success,
	}

	if issuer != nil {
		data["issuer_id"] = issuer.ID
		data["issuer_name"] = issuer.Name
	}

	oas.RecordEvent(OnboardingAnalyticsEvent{
		EventType: "credential_verified",
		SessionID: sessionID,
		UserID:    userID,
		StepName:  string(credType),
		Data:      data,
	})

	oas.updateCredentialStats(credType, issuer, success)
}

// GetFunnelMetrics returns the current funnel metrics
func (oas *OnboardingAnalyticsService) GetFunnelMetrics() *OnboardingFunnelMetrics {
	oas.mu.RLock()
	defer oas.mu.RUnlock()

	metrics := &OnboardingFunnelMetrics{
		SkippedAtSteps: make(map[string]int),
	}

	// Calculate from events
	sessionMap := make(map[string]map[string]bool) // sessionID -> step -> completed

	for _, event := range oas.events {
		if _, ok := sessionMap[event.SessionID]; !ok {
			sessionMap[event.SessionID] = make(map[string]bool)
		}

		switch event.EventType {
		case "session_started":
			if metrics.TotalSessionsStarted == 0 {
				metrics.TotalSessionsStarted = 1
			}
		case "carousel_viewed":
			metrics.CarouselViewed++
		case "phone_entry_started":
			metrics.PhoneEntryStarted++
		case "phone_verified":
			metrics.PhoneVerified++
		case "passkey_setup":
			metrics.PasskeySetup++
		case "recovery_setup":
			metrics.RecoverySetup++
		case "profile_setup":
			metrics.ProfileSetup++
		case "session_completed":
			metrics.CompletedOnboarding++
		case "step_skipped":
			metrics.SkippedAtSteps[event.StepName]++
		}
	}

	// Calculate conversion rate
	if metrics.TotalSessionsStarted > 0 {
		metrics.ConversionRate = float64(metrics.CompletedOnboarding) / float64(metrics.TotalSessionsStarted)
	}

	return metrics
}

// GetCredentialStatistics returns credential usage statistics
func (oas *OnboardingAnalyticsService) GetCredentialStatistics() map[CredentialType]*CredentialUsageStats {
	oas.mu.RLock()
	defer oas.mu.RUnlock()

	// Return a copy
	result := make(map[CredentialType]*CredentialUsageStats)
	for credType, stats := range oas.credentialStats {
		statCopy := *stats
		issuerCopy := make(map[string]int)
		for k, v := range stats.IssuerDistribution {
			issuerCopy[k] = v
		}
		statCopy.IssuerDistribution = issuerCopy
		result[credType] = &statCopy
	}
	return result
}

// GetSessionDuration calculates the total duration of a session
func (oas *OnboardingAnalyticsService) GetSessionDuration(sessionID string) time.Duration {
	oas.mu.RLock()
	defer oas.mu.RUnlock()

	var startTime, endTime *time.Time

	for _, event := range oas.events {
		if event.SessionID != sessionID {
			continue
		}

		if startTime == nil {
			startTime = &event.Timestamp
		}
		endTime = &event.Timestamp
	}

	if startTime == nil || endTime == nil {
		return 0
	}

	return endTime.Sub(*startTime)
}

// GetAvailableCredentials returns suggested credentials based on verified issuers
func (oas *OnboardingAnalyticsService) GetAvailableCredentials(
	registry *TrustRegistryService,
) map[CredentialType]*TrustedIssuer {
	issuers := registry.GetActiveTrustedIssuers()
	result := make(map[CredentialType]*TrustedIssuer)

	// Map credentials to highest-trust issuer
	for _, issuer := range issuers {
		for _, credType := range issuer.CredentialTypes {
			if existing, ok := result[credType]; !ok {
				result[credType] = issuer
			} else {
				// Prefer higher trust level
				if issuer.TrustLevel == TrustLevelHigh && existing.TrustLevel != TrustLevelHigh {
					result[credType] = issuer
				}
			}
		}
	}

	return result
}

// Private helper functions

func (oas *OnboardingAnalyticsService) logEvent(event OnboardingAnalyticsEvent) {
	oas.mu.Lock()
	defer oas.mu.Unlock()

	oas.events = append(oas.events, event)
}

func (oas *OnboardingAnalyticsService) processEvents() {
	for {
		select {
		case event := <-oas.eventBuffer:
			oas.logEvent(event)
		case <-oas.stopChan:
			return
		}
	}
}

func (oas *OnboardingAnalyticsService) updateCredentialStats(
	credType CredentialType,
	issuer *TrustedIssuer,
	success bool,
) {
	oas.mu.Lock()
	defer oas.mu.Unlock()

	if _, ok := oas.credentialStats[credType]; !ok {
		oas.credentialStats[credType] = &CredentialUsageStats{
			CredentialType:     credType,
			IssuerDistribution: make(map[string]int),
		}
	}

	stats := oas.credentialStats[credType]
	stats.Count++

	if success {
		stats.VerificationSuccess++
	} else {
		stats.VerificationFailure++
	}

	if issuer != nil {
		stats.IssuerDistribution[issuer.ID]++
	}
}

// Shutdown gracefully stops the analytics service
func (oas *OnboardingAnalyticsService) Shutdown() {
	// Signal all goroutines to stop
	close(oas.stopChan)

	// Give a brief moment for the goroutine to exit
	time.Sleep(10 * time.Millisecond)

	// Process any remaining events in the buffer (non-blocking reads)
	for {
		select {
		case event := <-oas.eventBuffer:
			oas.logEvent(event)
		default:
			// No more events in the buffer
			return
		}
	}
}
