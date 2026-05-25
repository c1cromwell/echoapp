package auth

import (
	"context"
	"sync"
	"testing"
	"time"
)

// fakeRedisBackend is a shared in-memory stand-in for *infra.RedisClient. Two
// TokenService instances pointed at the same fake simulate a process restart
// against the same durable store.
type fakeRedisBackend struct {
	mu      sync.Mutex
	blocked map[string]time.Time
	nonces  map[string]struct{}
}

func newFakeRedisBackend() *fakeRedisBackend {
	return &fakeRedisBackend{blocked: map[string]time.Time{}, nonces: map[string]struct{}{}}
}

func (f *fakeRedisBackend) BlocklistToken(_ context.Context, jti string, exp time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.blocked[jti] = exp
	return nil
}

func (f *fakeRedisBackend) IsBlocklisted(_ context.Context, jti string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.blocked[jti]
	return ok, nil
}

func (f *fakeRedisBackend) SetNX(_ context.Context, key string, _ []byte, _ time.Duration) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.nonces[key]; ok {
		return false, nil
	}
	f.nonces[key] = struct{}{}
	return true, nil
}

func TestDurableBlocklist_SurvivesRestart(t *testing.T) {
	shared := newFakeRedisBackend()

	ts1, _ := NewTokenService()
	ts1.SetRedisBackend(shared)
	ts1.BlocklistToken("jti-revoked", time.Now().Add(time.Hour))

	ts2, _ := NewTokenService() // "restart": new in-memory maps, same durable store
	ts2.SetRedisBackend(shared)

	if !ts2.IsBlocklisted("jti-revoked") {
		t.Fatal("revoked jti must remain blocklisted across a restart (durable store)")
	}
	if ts2.IsBlocklisted("jti-never-seen") {
		t.Fatal("unknown jti must not be blocklisted")
	}
}

func TestDurableNonce_SingleUseAcrossInstances(t *testing.T) {
	shared := newFakeRedisBackend()

	ts1, _ := NewTokenService()
	ts1.SetRedisBackend(shared)
	ts2, _ := NewTokenService()
	ts2.SetRedisBackend(shared)

	if !ts1.CheckAndStoreNonce("nonce-A") {
		t.Fatal("first use of a nonce must be accepted")
	}
	if ts2.CheckAndStoreNonce("nonce-A") {
		t.Fatal("replayed nonce must be rejected even on a different instance (durable single-use)")
	}
}

func TestInMemoryFallback_WhenNoBackend(t *testing.T) {
	ts, _ := NewTokenService() // no backend set
	ts.BlocklistToken("jti-x", time.Now().Add(time.Hour))
	if !ts.IsBlocklisted("jti-x") {
		t.Fatal("in-memory blocklist must work when no redis backend is set")
	}
	if !ts.CheckAndStoreNonce("n1") || ts.CheckAndStoreNonce("n1") {
		t.Fatal("in-memory nonce single-use must work when no redis backend is set")
	}
}
