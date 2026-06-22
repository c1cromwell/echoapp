package infra

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeRateStore is a correct in-memory sliding-window stand-in for *RedisClient,
// so the delegation + fallback behaviour can be tested without a real Redis.
type fakeRateStore struct {
	mu      sync.Mutex
	hits    map[string][]time.Time
	failErr error
}

func newFakeRateStore() *fakeRateStore { return &fakeRateStore{hits: map[string][]time.Time{}} }

func (f *fakeRateStore) AllowRate(_ context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error) {
	if f.failErr != nil {
		return false, 0, 0, f.failErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-window)
	kept := f.hits[key][:0]
	for _, t := range f.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= limit {
		f.hits[key] = kept
		return false, 0, kept[0].Add(window).Sub(now), nil
	}
	kept = append(kept, now)
	f.hits[key] = kept
	return true, limit - len(kept), 0, nil
}

func TestRateLimiter_StoreEnforcesLimitAcrossKeys(t *testing.T) {
	rl := NewRateLimiter(map[string]RateLimitConfig{"act": {MaxRequests: 2, Window: time.Minute}})
	rl.SetStore(newFakeRateStore())

	if err := rl.Check("did1", "act"); err != nil {
		t.Fatalf("1st should pass: %v", err)
	}
	if err := rl.Check("did1", "act"); err != nil {
		t.Fatalf("2nd should pass: %v", err)
	}
	if err := rl.Check("did1", "act"); err == nil {
		t.Fatal("3rd should be rate limited by the store")
	}
	// A different DID has its own window.
	if err := rl.Check("did2", "act"); err != nil {
		t.Fatalf("other DID should pass: %v", err)
	}
}

func TestRateLimiter_FallsBackToInMemoryOnStoreError(t *testing.T) {
	rl := NewRateLimiter(map[string]RateLimitConfig{"act": {MaxRequests: 1, Window: time.Minute}})
	store := newFakeRateStore()
	store.failErr = errors.New("redis down")
	rl.SetStore(store)

	// Store errors on every call → in-process window still enforces the limit.
	if err := rl.Check("did1", "act"); err != nil {
		t.Fatalf("1st should pass via in-memory fallback: %v", err)
	}
	if err := rl.Check("did1", "act"); err == nil {
		t.Fatal("2nd should be blocked by in-memory fallback despite store error")
	}
}
