package auth

import (
	"context"
	"fmt"
	"time"
)

// AttestationType represents the type of Web of Trust attestation
type AttestationType string

const (
	AttestationVouch   AttestationType = "vouch"   // Basic endorsement
	AttestationEndorse AttestationType = "endorse" // Strong endorsement
	AttestationVerify  AttestationType = "verify"  // Direct verification
)

// AttestationContext represents the domain of the attestation
type AttestationContext string

const (
	ContextPersonal     AttestationContext = "personal"
	ContextProfessional AttestationContext = "professional"
	ContextCommunity    AttestationContext = "community"
)

// WebOfTrustAttestation represents a peer vouching for another user
type WebOfTrustAttestation struct {
	AttestationID   string             `json:"attestation_id"`
	AttesterDID     string             `json:"attester_did"`
	SubjectDID      string             `json:"subject_did"`
	Type            AttestationType    `json:"type"`
	Confidence      int                `json:"confidence"` // 1-10
	Context         AttestationContext `json:"context"`
	Statement       string             `json:"statement"` // Optional narrative
	CreatedAt       time.Time          `json:"created_at"`
	ExpiresAt       time.Time          `json:"expires_at"`
	Signature       string             `json:"signature"` // Cryptographic signature
	MetagraphTxHash string             `json:"metagraph_tx_hash"`
	Revoked         bool               `json:"revoked"`
	RevokedAt       *time.Time         `json:"revoked_at,omitempty"`
}

// WebOfTrustConfig defines the rules for trust propagation
type WebOfTrustConfig struct {
	// Minimum trust score to vouch for others
	MinVouchingTrustScore int `json:"min_vouching_trust_score"` // 50

	// Maximum vouches per user per period
	MaxVouchesPerWeek int `json:"max_vouches_per_week"` // 5

	// Trust score boost from vouches
	VouchBoostPoints map[AttestationType]int `json:"vouch_boost_points"`
	// vouch: 1, endorse: 2, verify: 5

	// Maximum boost from web of trust
	MaxWebOfTrustBoost int `json:"max_web_of_trust_boost"` // 15

	// Validity period for attestations
	AttestationValidityDays int `json:"attestation_validity_days"` // 180 (6 months)

	// Decay rate for vouches
	VouchDecayDays int `json:"vouch_decay_days"` // 180

	// Minimum attestations for trust levels
	MinAttestationsForTrusted  int `json:"min_attestations_trusted"`  // 3
	MinAttestationsForVerified int `json:"min_attestations_verified"` // 10

	// ECHO rewards for vouching
	ECHORewardPerVouch    int64 `json:"echo_reward_per_vouch"`    // 5 ECHO
	ECHORewardWhenVouched int64 `json:"echo_reward_when_vouched"` // 10 ECHO
}

// WebOfTrustService manages decentralized trust attestations
type WebOfTrustService struct {
	config *WebOfTrustConfig
	// In production: database, blockchain client, etc.
}

// NewWebOfTrustService creates a new Web of Trust service
func NewWebOfTrustService(config *WebOfTrustConfig) *WebOfTrustService {
	return &WebOfTrustService{
		config: config,
	}
}

// CreateAttestation allows a user to vouch for another
func (s *WebOfTrustService) CreateAttestation(
	ctx context.Context,
	attesterDID string,
	subjectDID string,
	attestationType AttestationType,
	confidence int,
	attestationContext AttestationContext,
	statement string,
) (*WebOfTrustAttestation, *ECHOReward, error) {

	// 1. Validate confidence level
	if confidence < 1 || confidence > 10 {
		return nil, nil, fmt.Errorf("confidence must be 1-10, got %d", confidence)
	}

	// 2. Prevent self-vouching
	if attesterDID == subjectDID {
		return nil, nil, fmt.Errorf("cannot vouch for yourself")
	}

	// 3. Create attestation
	attestation := &WebOfTrustAttestation{
		AttestationID: generateAttestationID(),
		AttesterDID:   attesterDID,
		SubjectDID:    subjectDID,
		Type:          attestationType,
		Confidence:    confidence,
		Context:       attestationContext,
		Statement:     statement,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().AddDate(0, 6, 0), // 6 months
	}

	// 4. Calculate ECHO reward for attester
	reward := &ECHOReward{
		Amount:     5,
		Source:     fmt.Sprintf("vouch:%s", attestationType),
		Multiplier: 1.0,
		RewardType: "vouch",
		EarnedAt:   time.Now(),
	}

	return attestation, reward, nil
}

// RevokeAttestation allows an attester to revoke their attestation
func (s *WebOfTrustService) RevokeAttestation(
	ctx context.Context,
	attestationID string,
	attesterDID string,
) error {

	// In production: verify attestationID belongs to attesterDID
	// Fetch attestation from database
	// Check that attester matches

	// Mark as revoked
	now := time.Now()
	_ = &WebOfTrustAttestation{
		AttestationID: attestationID,
		Revoked:       true,
		RevokedAt:     &now,
	}

	return nil
}

// GetAttestationsForUser retrieves all attestations for a user
func (s *WebOfTrustService) GetAttestationsForUser(
	ctx context.Context,
	userDID string,
) ([]*WebOfTrustAttestation, error) {

	// In production: query database for active attestations
	// Filter out revoked ones
	// Filter out expired ones

	attestations := []*WebOfTrustAttestation{}
	return attestations, nil
}

// CalculateWebOfTrustBoost computes trust score boost from attestations
func (s *WebOfTrustService) CalculateWebOfTrustBoost(
	ctx context.Context,
	attestations []*WebOfTrustAttestation,
) (int, error) {

	if len(attestations) == 0 {
		return 0, nil
	}

	boost := 0

	for _, att := range attestations {
		// Skip revoked or expired attestations
		if att.Revoked || time.Now().After(att.ExpiresAt) {
			continue
		}

		// Weight by type
		typeBoost := 0
		switch att.Type {
		case AttestationVouch:
			typeBoost = 1
		case AttestationEndorse:
			typeBoost = 2
		case AttestationVerify:
			typeBoost = 5
		}

		// Weight by confidence (1-10)
		confidenceWeight := float64(att.Confidence) / 10.0

		// Calculate contribution
		contribution := int(float64(typeBoost) * confidenceWeight)
		boost += contribution
	}

	// Cap at maximum
	maxBoost := 15
	if boost > maxBoost {
		boost = maxBoost
	}

	return boost, nil
}

// GetAttestationStats returns statistics about a user's attestations
type AttestationStats struct {
	TotalAttestation    int
	VouchCount          int
	EndorseCount        int
	VerifyCount         int
	AverageConfidence   float64
	ActiveAttestations  int
	ExpiringAttestation int
	RevokedCount        int
}

func (s *WebOfTrustService) GetAttestationStats(
	ctx context.Context,
	attestations []*WebOfTrustAttestation,
) *AttestationStats {

	stats := &AttestationStats{}

	for _, att := range attestations {
		stats.TotalAttestation++

		if att.Revoked {
			stats.RevokedCount++
			continue
		}

		if time.Now().After(att.ExpiresAt) {
			continue // Expired, don't count as active
		}

		stats.ActiveAttestations++

		// Check if expiring soon (within 30 days)
		if time.Until(att.ExpiresAt) < 30*24*time.Hour {
			stats.ExpiringAttestation++
		}

		// Count by type
		switch att.Type {
		case AttestationVouch:
			stats.VouchCount++
		case AttestationEndorse:
			stats.EndorseCount++
		case AttestationVerify:
			stats.VerifyCount++
		}
	}

	// Calculate average confidence
	if stats.ActiveAttestations > 0 {
		totalConfidence := 0
		for _, att := range attestations {
			if !att.Revoked && time.Now().Before(att.ExpiresAt) {
				totalConfidence += att.Confidence
			}
		}
		stats.AverageConfidence = float64(totalConfidence) / float64(stats.ActiveAttestations)
	}

	return stats
}

// DetectCircularVouching checks if creating an attestation would create a circle
func (s *WebOfTrustService) DetectCircularVouching(
	ctx context.Context,
	attesterDID string,
	subjectDID string,
	maxDepth int,
) (bool, error) {

	// BFS to check for circular dependency
	visited := make(map[string]bool)
	queue := []string{subjectDID}
	depth := 0

	for len(queue) > 0 && depth < maxDepth {
		current := queue[0]
		queue = queue[1:]

		if current == attesterDID {
			return true, nil // Circular!
		}

		if visited[current] {
			continue
		}
		visited[current] = true

		// In production: get all attestations where current is attester
		// Add subjects to queue

		depth++
	}

	return false, nil
}

// Helper function to generate unique attestation ID
func generateAttestationID() string {
	return fmt.Sprintf("att_%d", time.Now().UnixNano())
}

// TransitiveTrustCalculator computes trust transitive through network
type TransitiveTrustCalculator struct {
	attestations map[string][]*WebOfTrustAttestation
}

// NewTransitiveTrustCalculator creates a new transitive trust calculator
func NewTransitiveTrustCalculator() *TransitiveTrustCalculator {
	return &TransitiveTrustCalculator{
		attestations: make(map[string][]*WebOfTrustAttestation),
	}
}

// CalculateTransitiveTrust computes trust score through network paths
// Uses weighted path algorithm: trust(A→B→C) = trust(A→B) * trust(B→C) * decay
func (ttc *TransitiveTrustCalculator) CalculateTransitiveTrust(
	ctx context.Context,
	sourceDID string,
	targetDID string,
	maxHops int,
) (float64, error) {

	if sourceDID == targetDID {
		return 1.0, nil // Perfect trust in self
	}

	visited := make(map[string]bool)
	maxTrust := 0.0

	// BFS to find best trust path
	type pathNode struct {
		did   string
		trust float64
		hops  int
	}

	queue := []pathNode{{sourceDID, 1.0, 0}}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.hops > maxHops {
			continue
		}

		if visited[current.did] {
			continue
		}
		visited[current.did] = true

		if current.did == targetDID {
			if current.trust > maxTrust {
				maxTrust = current.trust
			}
			continue
		}

		// In production: get attestations where current.did is attester
		// for each attestation:
		//   - calculate decay: 1.0 - (hops / maxHops * 0.2) = decay from distance
		//   - confidence: attestation.Confidence / 10.0
		//   - add to queue: trust * decay * confidence

	}

	return maxTrust, nil
}
