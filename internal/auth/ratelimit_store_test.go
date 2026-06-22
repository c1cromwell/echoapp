package auth

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// fakeAuthRateStore is a correct in-memory sliding-window stand-in for the Redis
// store, mirroring the structural RateLimitStore interface.
type fakeAuthRateStore struct {
	mu      sync.Mutex
	hits    map[string][]time.Time
	failErr error
}

func newFakeAuthRateStore() *fakeAuthRateStore {
	return &fakeAuthRateStore{hits: map[string][]time.Time{}}
}

func (f *fakeAuthRateStore) AllowRate(_ context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error) {
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

func TestAuthRateLimiter_StoreEnforcesLimit(t *testing.T) {
	rl := NewAuthRateLimiter()
	rl.SetStore(newFakeAuthRateStore())
	cfg := AuthRateLimitConfig{Name: "test", Limit: 2, Window: time.Minute}

	if err := rl.Check("k1", cfg); err != nil {
		t.Fatalf("1st should pass: %v", err)
	}
	if err := rl.Check("k1", cfg); err != nil {
		t.Fatalf("2nd should pass: %v", err)
	}
	if err := rl.Check("k1", cfg); err == nil {
		t.Fatal("3rd should be rate limited by the store")
	} else if err.RetryAfter == nil || *err.RetryAfter < 1 {
		t.Fatalf("expected a positive Retry-After, got %+v", err.RetryAfter)
	}
}

func TestAuthRateLimiter_FallsBackOnStoreError(t *testing.T) {
	rl := NewAuthRateLimiter()
	store := newFakeAuthRateStore()
	store.failErr = errors.New("redis down")
	rl.SetStore(store)
	cfg := AuthRateLimitConfig{Name: "test", Limit: 1, Window: time.Minute}

	if err := rl.Check("k1", cfg); err != nil {
		t.Fatalf("1st should pass via in-memory fallback: %v", err)
	}
	if err := rl.Check("k1", cfg); err == nil {
		t.Fatal("2nd should be blocked by in-memory fallback despite store error")
	}
}
