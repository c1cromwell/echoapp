package trustnet

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// EndorsementCategory represents what a user is being endorsed for
type EndorsementCategory string

const (
	EndorseReliable     EndorsementCategory = "reliable"
	EndorseResponsive   EndorsementCategory = "responsive"
	EndorseProfessional EndorsementCategory = "professional"
	EndorseHelpful      EndorsementCategory = "helpful"
	EndorseTrustworthy  EndorsementCategory = "trustworthy"
)

const (
	// MinTrustToEndorse is the minimum trust score required to endorse others
	MinTrustToEndorse = 60

	// MaxEndorsementsPerDay is the daily endorsement limit
	MaxEndorsementsPerDay = 5

	// RevocationCooldown is the waiting period after revoking an endorsement
	// before the endorser can re-endorse the same user in the same category
	RevocationCooldown = 30 * 24 * time.Hour // 30 days
)

// Endorsement represents a trust endorsement from one user to another
type Endorsement struct {
	ID          string
	EndorserDID string
	EndorseeDID string
	Category    EndorsementCategory
	Weight      float64 // based on endorser's trust score
	Message     string  // optional endorsement message
	CreatedAt   time.Time
	RevokedAt   *time.Time
	Active      bool
}

// RevocationRecord tracks when an endorsement was revoked for cooldown
type RevocationRecord struct {
	EndorserDID string
	EndorseeDID string
	Category    EndorsementCategory
	RevokedAt   time.Time
}

// EndorsementService manages endorsements between users
type EndorsementService struct {
	mu          sync.RWMutex
	endorsements map[string]*Endorsement   // endorsementID -> endorsement
	byEndorser   map[string][]string       // endorserDID -> []endorsementID
	byEndorsee   map[string][]string       // endorseeDID -> []endorsementID
	dailyCounts  map[string]map[string]int // endorserDID -> date -> count
	revocations  []RevocationRecord
}

// NewEndorsementService creates a new endorsement service
func NewEndorsementService() *EndorsementService {
	return &EndorsementService{
		endorsements: make(map[string]*Endorsement),
		byEndorser:   make(map[string][]string),
		byEndorsee:   make(map[string][]string),
		dailyCounts:  make(map[string]map[string]int),
	}
}

// Endorse creates a new endorsement
func (s *EndorsementService) Endorse(endorserDID, endorseeDID string, category EndorsementCategory, trustScore int, message string) (*Endorsement, error) {
	if endorserDID == endorseeDID {
		return nil, ErrEndorsementSelfEndorse
	}

	if !isValidCategory(category) {
		return nil, ErrEndorsementNotFound
	}

	if trustScore < MinTrustToEndorse {
		return nil, ErrEndorsementInsufficientTrust
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check daily rate limit
	today := time.Now().Format("2006-01-02")
	if s.dailyCounts[endorserDID] == nil {
		s.dailyCounts[endorserDID] = make(map[string]int)
	}
	if s.dailyCounts[endorserDID][today] >= MaxEndorsementsPerDay {
		return nil, ErrEndorsementRateLimited
	}

	// Check for duplicate active endorsement in same category
	for _, id := range s.byEndorser[endorserDID] {
		e := s.endorsements[id]
		if e.Active && e.EndorseeDID == endorseeDID && e.Category == category {
			return nil, ErrEndorsementDuplicate
		}
	}

	// Check revocation cooldown
	for _, r := range s.revocations {
		if r.EndorserDID == endorserDID && r.EndorseeDID == endorseeDID && r.Category == category {
			if time.Since(r.RevokedAt) < RevocationCooldown {
				return nil, ErrEndorsementCooldown
			}
		}
	}

	// Calculate weight based on endorser's trust score (0.6 to 1.0)
	weight := 0.6 + (float64(trustScore-MinTrustToEndorse) / float64(100-MinTrustToEndorse) * 0.4)
	if weight > 1.0 {
		weight = 1.0
	}

	endorsement := &Endorsement{
		ID:          uuid.New().String(),
		EndorserDID: endorserDID,
		EndorseeDID: endorseeDID,
		Category:    category,
		Weight:      weight,
		Message:     message,
		CreatedAt:   time.Now(),
		Active:      true,
	}

	s.endorsements[endorsement.ID] = endorsement
	s.byEndorser[endorserDID] = append(s.byEndorser[endorserDID], endorsement.ID)
	s.byEndorsee[endorseeDID] = append(s.byEndorsee[endorseeDID], endorsement.ID)
	s.dailyCounts[endorserDID][today]++

	return endorsement, nil
}

// Revoke revokes an endorsement
func (s *EndorsementService) Revoke(endorsementID, endorserDID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.endorsements[endorsementID]
	if !ok {
		return ErrEndorsementNotFound
	}

	if e.EndorserDID != endorserDID {
		return ErrEndorsementNotOwner
	}

	if !e.Active {
		return ErrEndorsementNotFound
	}

	now := time.Now()
	e.Active = false
	e.RevokedAt = &now

	s.revocations = append(s.revocations, RevocationRecord{
		EndorserDID: e.EndorserDID,
		EndorseeDID: e.EndorseeDID,
		Category:    e.Category,
		RevokedAt:   now,
	})

	return nil
}

// GetEndorsement retrieves an endorsement by ID
func (s *EndorsementService) GetEndorsement(id string) (*Endorsement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.endorsements[id]
	if !ok {
		return nil, ErrEndorsementNotFound
	}
	return e, nil
}

// GetEndorsementsFor returns all active endorsements for a user
func (s *EndorsementService) GetEndorsementsFor(endorseeDID string) []*Endorsement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Endorsement
	for _, id := range s.byEndorsee[endorseeDID] {
		e := s.endorsements[id]
		if e.Active {
			result = append(result, e)
		}
	}
	return result
}

// GetEndorsementsBy returns all active endorsements given by a user
func (s *EndorsementService) GetEndorsementsBy(endorserDID string) []*Endorsement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Endorsement
	for _, id := range s.byEndorser[endorserDID] {
		e := s.endorsements[id]
		if e.Active {
			result = append(result, e)
		}
	}
	return result
}

// GetEndorsementsByCategory returns active endorsements for a user in a specific category
func (s *EndorsementService) GetEndorsementsByCategory(endorseeDID string, category EndorsementCategory) []*Endorsement {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Endorsement
	for _, id := range s.byEndorsee[endorseeDID] {
		e := s.endorsements[id]
		if e.Active && e.Category == category {
			result = append(result, e)
		}
	}
	return result
}

// CountActiveEndorsements returns the count of active endorsements for a user
func (s *EndorsementService) CountActiveEndorsements(endorseeDID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, id := range s.byEndorsee[endorseeDID] {
		if s.endorsements[id].Active {
			count++
		}
	}
	return count
}

// CalculateNetworkScore calculates the network component of trust score from endorsements
func (s *EndorsementService) CalculateNetworkScore(endorseeDID string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalWeight := 0.0
	count := 0
	for _, id := range s.byEndorsee[endorseeDID] {
		e := s.endorsements[id]
		if e.Active {
			totalWeight += e.Weight
			count++
		}
	}

	if count == 0 {
		return 0
	}

	// Diminishing returns: each additional endorsement contributes less
	// First endorsement = full weight, subsequent ones diminish
	score := 0.0
	diminishFactor := 1.0
	for _, id := range s.byEndorsee[endorseeDID] {
		e := s.endorsements[id]
		if e.Active {
			score += e.Weight * diminishFactor
			diminishFactor *= 0.85 // 15% diminishing per additional endorsement
		}
	}

	// Normalize to 0-25 range (network component is 25% of total score)
	maxScore := 25.0
	if score > maxScore {
		score = maxScore
	}
	return score
}

// DailyEndorsementsRemaining returns how many endorsements the user can still give today
func (s *EndorsementService) DailyEndorsementsRemaining(endorserDID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	today := time.Now().Format("2006-01-02")
	used := 0
	if s.dailyCounts[endorserDID] != nil {
		used = s.dailyCounts[endorserDID][today]
	}
	remaining := MaxEndorsementsPerDay - used
	if remaining < 0 {
		return 0
	}
	return remaining
}

func isValidCategory(cat EndorsementCategory) bool {
	switch cat {
	case EndorseReliable, EndorseResponsive, EndorseProfessional, EndorseHelpful, EndorseTrustworthy:
		return true
	}
	return false
}
