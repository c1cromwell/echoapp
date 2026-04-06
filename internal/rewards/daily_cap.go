package rewards

import (
	"context"
	"errors"
	"sync"
	"time"
)

// RewardType identifies a reward category.
type RewardType string

const (
	RewardMessaging   RewardType = "messaging"
	RewardReferral    RewardType = "referral"
	RewardStaking     RewardType = "staking"
	RewardPaymentRail RewardType = "payment_rail"
)

// AllRewardTypes returns all reward types for iteration.
func AllRewardTypes() []RewardType {
	return []RewardType{RewardMessaging, RewardReferral, RewardStaking, RewardPaymentRail}
}

// DailyCapEntry tracks how much a user has earned today for a specific reward type.
type DailyCapEntry struct {
	DID        string     `json:"did"`
	RewardType RewardType `json:"rewardType"`
	Earned     int64      `json:"earned"`
	Cap        int64      `json:"cap"`
	ResetAt    time.Time  `json:"resetAt"`
}

// Remaining returns how much more the user can earn today.
func (e *DailyCapEntry) Remaining() int64 {
	r := e.Cap - e.Earned
	if r < 0 {
		return 0
	}
	return r
}

// IsExpired returns whether this cap entry needs resetting (past UTC midnight).
func (e *DailyCapEntry) IsExpired() bool {
	return time.Now().After(e.ResetAt)
}

// DailyCapStore abstracts the persistence layer for daily cap tracking.
type DailyCapStore interface {
	GetCap(ctx context.Context, did string, rewardType RewardType) (*DailyCapEntry, error)
	SetCap(ctx context.Context, entry *DailyCapEntry) error
	IncrementEarned(ctx context.Context, did string, rewardType RewardType, amount int64) error
	ResetExpired(ctx context.Context) (int, error)
}

// DailyCapTracker manages per-DID daily reward caps.
type DailyCapTracker struct {
	store    DailyCapStore
	emission *EmissionSchedule
	mu       sync.RWMutex
}

// NewDailyCapTracker creates a tracker with the given store and emission schedule.
func NewDailyCapTracker(store DailyCapStore, emission *EmissionSchedule) *DailyCapTracker {
	return &DailyCapTracker{
		store:    store,
		emission: emission,
	}
}

// DefaultCaps returns the default daily cap for each reward type based on trust tier.
func DefaultCaps(tier int) map[RewardType]int64 {
	// Base caps scale with trust tier (multiplier: tier * 1.0)
	base := map[RewardType]int64{
		RewardMessaging:   100_00000000, // 100 ECHO
		RewardReferral:    50_00000000,  // 50 ECHO
		RewardStaking:     0,            // Staking rewards have no per-user cap
		RewardPaymentRail: 25_00000000,  // 25 ECHO
	}
	multiplier := int64(tier)
	if multiplier < 1 {
		multiplier = 1
	}
	caps := make(map[RewardType]int64)
	for rt, cap := range base {
		caps[rt] = cap * multiplier
	}
	return caps
}

// CheckAndRecord validates a reward against the daily cap and records it.
func (t *DailyCapTracker) CheckAndRecord(ctx context.Context, did string, rewardType RewardType, amount int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	entry, err := t.store.GetCap(ctx, did, rewardType)
	if err != nil {
		return err
	}

	if entry == nil || entry.IsExpired() {
		return errors.New("daily cap not initialized or expired; call InitializeCaps first")
	}

	if amount > entry.Remaining() {
		return ErrDailyCapExceeded
	}

	return t.store.IncrementEarned(ctx, did, rewardType, amount)
}

// InitializeCaps sets up daily cap entries for a DID based on their trust tier.
func (t *DailyCapTracker) InitializeCaps(ctx context.Context, did string, trustTier int) error {
	caps := DefaultCaps(trustTier)
	midnight := nextUTCMidnight()

	for rt, cap := range caps {
		entry := &DailyCapEntry{
			DID:        did,
			RewardType: rt,
			Earned:     0,
			Cap:        cap,
			ResetAt:    midnight,
		}
		if err := t.store.SetCap(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

// GetDailyCaps returns all cap entries for a DID.
func (t *DailyCapTracker) GetDailyCaps(ctx context.Context, did string) (map[RewardType]*DailyCapEntry, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[RewardType]*DailyCapEntry)
	for _, rt := range AllRewardTypes() {
		entry, err := t.store.GetCap(ctx, did, rt)
		if err != nil {
			return nil, err
		}
		if entry != nil {
			result[rt] = entry
		}
	}
	return result, nil
}

func nextUTCMidnight() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
}

var ErrDailyCapExceeded = errors.New("daily reward cap exceeded")
