package messaging

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrDisappearingTTLRestricted is returned when a TTL is below the caller's trust tier minimum.
var ErrDisappearingTTLRestricted = errors.New("disappearing ttl below trust tier minimum")

// DisappearingRestrictionPolicy defines minimum TTL by trust tier (WO-115).
type DisappearingRestrictionPolicy struct {
	MinTTLSeconds int      `json:"min_ttl_seconds"`
	Reason        string   `json:"reason"`
	BlockedTTLs   []int    `json:"blocked_ttls,omitempty"`
}

// MinTTLForTrustTier returns the minimum allowed disappearing TTL in seconds.
// Tier 1: 1h, Tier 2: 5m, Tier 3+: no floor (10s presets allowed).
func MinTTLForTrustTier(tier int) int {
	switch {
	case tier <= 1:
		return 3600
	case tier == 2:
		return 300
	default:
		return 10
	}
}

// PolicyForTier returns the restriction policy shown to clients.
func PolicyForTier(tier int) DisappearingRestrictionPolicy {
	min := MinTTLForTrustTier(tier)
	policy := DisappearingRestrictionPolicy{
		MinTTLSeconds: min,
	}
	switch {
	case tier <= 1:
		policy.Reason = "Unverified accounts cannot use timers shorter than 1 hour to prevent harassment."
		policy.BlockedTTLs = []int{10, 30, 60, 300}
	case tier == 2:
		policy.Reason = "Newcomer accounts cannot use timers shorter than 5 minutes."
		policy.BlockedTTLs = []int{10, 30, 60}
	default:
		policy.Reason = "No disappearing-message restrictions for your trust tier."
	}
	return policy
}

// ValidateDisappearingTTL returns nil when ttl is allowed for the given tier (0 = off).
func ValidateDisappearingTTL(tier, ttlSeconds int) error {
	if ttlSeconds <= 0 {
		return nil
	}
	min := MinTTLForTrustTier(tier)
	if ttlSeconds < min {
		return fmt.Errorf("%w: tier=%d min=%ds requested=%ds", ErrDisappearingTTLRestricted, tier, min, ttlSeconds)
	}
	return nil
}

// DisappearingAppeal is a user request to review imposed TTL restrictions.
type DisappearingAppeal struct {
	ID        string    `json:"id"`
	DIDD      string    `json:"did"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// DisappearingRestrictionService tracks violations and appeals (WO-115).
type DisappearingRestrictionService struct {
	mu        sync.Mutex
	violations map[string]int
	appeals    []DisappearingAppeal
}

// NewDisappearingRestrictionService creates an in-memory restriction tracker.
func NewDisappearingRestrictionService() *DisappearingRestrictionService {
	return &DisappearingRestrictionService{
		violations: make(map[string]int),
	}
}

// RecordViolation increments abuse counter for escalation.
func (s *DisappearingRestrictionService) RecordViolation(did string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.violations[did]++
	return s.violations[did]
}

// ViolationCount returns how many blocked TTL attempts were recorded.
func (s *DisappearingRestrictionService) ViolationCount(did string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.violations[did]
}

// SubmitAppeal records a restriction appeal for manual review.
func (s *DisappearingRestrictionService) SubmitAppeal(did, reason string) DisappearingAppeal {
	s.mu.Lock()
	defer s.mu.Unlock()
	appeal := DisappearingAppeal{
		ID:        fmt.Sprintf("appeal_%d", time.Now().UnixNano()),
		DIDD:      did,
		Reason:    reason,
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}
	s.appeals = append(s.appeals, appeal)
	return appeal
}

// ListAppeals returns appeals for a DID.
func (s *DisappearingRestrictionService) ListAppeals(did string) []DisappearingAppeal {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []DisappearingAppeal
	for _, a := range s.appeals {
		if a.DIDD == did {
			out = append(out, a)
		}
	}
	return out
}
