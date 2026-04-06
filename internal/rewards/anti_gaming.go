package rewards

import (
	"errors"
	"sync"
	"time"
)

// Anti-gaming limits from the Scala RewardClaimValidator spec.
const (
	MaxClaimsPerHour       = 10
	VelocityWindow         = 1 * time.Hour
	DuplicateWindowSeconds = 60
)

// Errors.
var (
	ErrVelocityExceeded = errors.New("rate limit exceeded: max 10 claims per hour")
	ErrDuplicateClaim   = errors.New("duplicate claim: same type and amount within 60 seconds")
)

// ClaimEvent records a single reward claim for anti-gaming analysis.
type ClaimEvent struct {
	DID        string
	RewardType string
	Amount     int64
	Timestamp  time.Time
}

// AntiGamingDetector provides velocity checks and duplicate detection.
type AntiGamingDetector struct {
	mu     sync.Mutex
	claims map[string][]ClaimEvent // DID -> recent claims
}

// NewAntiGamingDetector creates a new detector.
func NewAntiGamingDetector() *AntiGamingDetector {
	return &AntiGamingDetector{
		claims: make(map[string][]ClaimEvent),
	}
}

// CheckAndRecord validates a claim against anti-gaming rules and records it if valid.
// Returns an error if the claim is rejected.
func (d *AntiGamingDetector) CheckAndRecord(event ClaimEvent) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Prune expired events
	d.pruneExpired(event.DID, event.Timestamp)

	recent := d.claims[event.DID]

	// Velocity check: max 10 claims per hour
	if len(recent) >= MaxClaimsPerHour {
		return ErrVelocityExceeded
	}

	// Duplicate detection: same type + amount within 60 seconds
	for _, prev := range recent {
		if prev.RewardType == event.RewardType &&
			prev.Amount == event.Amount &&
			event.Timestamp.Sub(prev.Timestamp).Seconds() <= DuplicateWindowSeconds {
			return ErrDuplicateClaim
		}
	}

	// Record the claim
	d.claims[event.DID] = append(d.claims[event.DID], event)
	return nil
}

// RecentClaimCount returns how many claims the DID has in the current velocity window.
func (d *AntiGamingDetector) RecentClaimCount(did string) int {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.pruneExpired(did, time.Now())
	return len(d.claims[did])
}

// pruneExpired removes claims older than VelocityWindow. Must be called with lock held.
func (d *AntiGamingDetector) pruneExpired(did string, now time.Time) {
	claims := d.claims[did]
	cutoff := now.Add(-VelocityWindow)

	var pruned []ClaimEvent
	for _, c := range claims {
		if c.Timestamp.After(cutoff) {
			pruned = append(pruned, c)
		}
	}
	d.claims[did] = pruned
}

// Reset clears all tracked claims (used in testing).
func (d *AntiGamingDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.claims = make(map[string][]ClaimEvent)
}
