package rewards

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// --- In-Memory DailyCapStore ---

type memCapStore struct {
	mu   sync.RWMutex
	caps map[string]*DailyCapEntry // key: "did:rewardType"
}

func newMemCapStore() *memCapStore {
	return &memCapStore{caps: make(map[string]*DailyCapEntry)}
}

func capKey(did string, rt RewardType) string {
	return did + ":" + string(rt)
}

func (s *memCapStore) GetCap(_ context.Context, did string, rewardType RewardType) (*DailyCapEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.caps[capKey(did, rewardType)]
	if !ok {
		return nil, nil
	}
	return entry, nil
}

func (s *memCapStore) SetCap(_ context.Context, entry *DailyCapEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.caps[capKey(entry.DID, entry.RewardType)] = entry
	return nil
}

func (s *memCapStore) IncrementEarned(_ context.Context, did string, rewardType RewardType, amount int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := capKey(did, rewardType)
	entry, ok := s.caps[key]
	if !ok {
		return errors.New("cap entry not found")
	}
	entry.Earned += amount
	return nil
}

func (s *memCapStore) ResetExpired(_ context.Context) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, entry := range s.caps {
		if entry.IsExpired() {
			entry.Earned = 0
			entry.ResetAt = nextUTCMidnight()
			count++
		}
	}
	return count, nil
}

// --- Tests ---

func TestInitializeCaps(t *testing.T) {
	store := newMemCapStore()
	emission := NewEmissionSchedule(time.Now().Add(-24 * time.Hour))
	tracker := NewDailyCapTracker(store, emission)

	err := tracker.InitializeCaps(context.Background(), "did:echo:test", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	caps, err := tracker.GetDailyCaps(context.Background(), "did:echo:test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(caps) != 4 {
		t.Errorf("expected 4 reward types, got %d", len(caps))
	}

	// Tier 2 messaging cap should be 200 ECHO (100 * 2)
	msgCap := caps[RewardMessaging]
	if msgCap == nil {
		t.Fatal("expected messaging cap")
	}
	if msgCap.Cap != 200_00000000 {
		t.Errorf("expected messaging cap 200_00000000, got %d", msgCap.Cap)
	}
}

func TestCheckAndRecord_WithinCap(t *testing.T) {
	store := newMemCapStore()
	emission := NewEmissionSchedule(time.Now().Add(-24 * time.Hour))
	tracker := NewDailyCapTracker(store, emission)

	_ = tracker.InitializeCaps(context.Background(), "did:echo:test", 1)

	err := tracker.CheckAndRecord(context.Background(), "did:echo:test", RewardMessaging, 50_00000000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	caps, _ := tracker.GetDailyCaps(context.Background(), "did:echo:test")
	if caps[RewardMessaging].Earned != 50_00000000 {
		t.Errorf("expected earned 50_00000000, got %d", caps[RewardMessaging].Earned)
	}
}

func TestCheckAndRecord_ExceedsCap(t *testing.T) {
	store := newMemCapStore()
	emission := NewEmissionSchedule(time.Now().Add(-24 * time.Hour))
	tracker := NewDailyCapTracker(store, emission)

	_ = tracker.InitializeCaps(context.Background(), "did:echo:test", 1)

	// Try to claim more than the 100 ECHO cap
	err := tracker.CheckAndRecord(context.Background(), "did:echo:test", RewardMessaging, 101_00000000)
	if !errors.Is(err, ErrDailyCapExceeded) {
		t.Errorf("expected ErrDailyCapExceeded, got %v", err)
	}
}

func TestCheckAndRecord_AccumulateToExact(t *testing.T) {
	store := newMemCapStore()
	emission := NewEmissionSchedule(time.Now().Add(-24 * time.Hour))
	tracker := NewDailyCapTracker(store, emission)

	_ = tracker.InitializeCaps(context.Background(), "did:echo:test", 1)

	// Claim exactly up to cap
	_ = tracker.CheckAndRecord(context.Background(), "did:echo:test", RewardMessaging, 50_00000000)
	err := tracker.CheckAndRecord(context.Background(), "did:echo:test", RewardMessaging, 50_00000000)
	if err != nil {
		t.Fatalf("expected success claiming to exact cap, got %v", err)
	}

	// One more should fail
	err = tracker.CheckAndRecord(context.Background(), "did:echo:test", RewardMessaging, 1)
	if !errors.Is(err, ErrDailyCapExceeded) {
		t.Errorf("expected ErrDailyCapExceeded after reaching cap, got %v", err)
	}
}

func TestDailyCapEntry_Remaining(t *testing.T) {
	entry := &DailyCapEntry{
		Earned: 30_00000000,
		Cap:    100_00000000,
	}

	if entry.Remaining() != 70_00000000 {
		t.Errorf("expected remaining 70_00000000, got %d", entry.Remaining())
	}

	// Over-earned
	entry.Earned = 110_00000000
	if entry.Remaining() != 0 {
		t.Errorf("expected remaining 0 when over-earned, got %d", entry.Remaining())
	}
}

func TestDailyCapEntry_IsExpired(t *testing.T) {
	future := &DailyCapEntry{ResetAt: time.Now().Add(time.Hour)}
	if future.IsExpired() {
		t.Error("expected not expired for future reset")
	}

	past := &DailyCapEntry{ResetAt: time.Now().Add(-time.Hour)}
	if !past.IsExpired() {
		t.Error("expected expired for past reset")
	}
}

func TestDefaultCaps_TierMultiplier(t *testing.T) {
	tier1 := DefaultCaps(1)
	tier3 := DefaultCaps(3)

	if tier3[RewardMessaging] != tier1[RewardMessaging]*3 {
		t.Errorf("tier 3 messaging cap should be 3x tier 1: got %d vs %d",
			tier3[RewardMessaging], tier1[RewardMessaging]*3)
	}
}

func TestAllRewardTypes(t *testing.T) {
	types := AllRewardTypes()
	if len(types) != 4 {
		t.Errorf("expected 4 reward types, got %d", len(types))
	}
}
