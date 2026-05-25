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
	mu        sync.Mutex
	blocked   map[string]time.Time
	nonces    map[string]struct{}
	refresh   map[string][]byte          // tokenHash -> record JSON
	userIndex map[string]map[string]bool // userID -> set of tokenHashes
}

func newFakeRedisBackend() *fakeRedisBackend {
	return &fakeRedisBackend{
		blocked:   map[string]time.Time{},
		nonces:    map[string]struct{}{},
		refresh:   map[string][]byte{},
		userIndex: map[string]map[string]bool{},
	}
}

func (f *fakeRedisBackend) RefreshPut(_ context.Context, tokenHash string, record []byte, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]byte, len(record))
	copy(cp, record)
	f.refresh[tokenHash] = cp
	return nil
}

func (f *fakeRedisBackend) RefreshGet(_ context.Context, tokenHash string) ([]byte, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	rec, ok := f.refresh[tokenHash]
	return rec, ok, nil
}

func (f *fakeRedisBackend) RefreshAddToUser(_ context.Context, userID, tokenHash string, _ time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.userIndex[userID] == nil {
		f.userIndex[userID] = map[string]bool{}
	}
	f.userIndex[userID][tokenHash] = true
	return nil
}

func (f *fakeRedisBackend) RefreshUserHashes(_ context.Context, userID string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	for h := range f.userIndex[userID] {
		out = append(out, h)
	}
	return out, nil
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

// TestDurableRefresh_RotationAndReuseSurviveRestart verifies refresh-token
// rotation and reuse detection work against a shared durable store across two
// TokenService instances (a simulated restart).
func TestDurableRefresh_RotationAndReuseSurviveRestart(t *testing.T) {
	shared := newFakeRedisBackend()

	ts1, _ := NewTokenService()
	ts1.SetRedisBackend(shared)
	refresh := GenerateRefreshToken()
	ts1.StoreRefreshToken("did:key:zUser", refresh, "dev1")

	// "Restart": fresh in-memory state, same durable store.
	ts2, _ := NewTokenService()
	ts2.SetRedisBackend(shared)

	newTok, rec, aerr := ts2.RotateRefreshToken(refresh, "dev1")
	if aerr != nil {
		t.Fatalf("rotate should succeed across restart, got %v", aerr)
	}
	if newTok == "" || newTok == refresh || rec.UserID != "did:key:zUser" {
		t.Fatalf("unexpected rotation result: tok=%q rec=%+v", newTok, rec)
	}

	// Reusing the consumed token must fail and revoke the family.
	if _, _, aerr := ts2.RotateRefreshToken(refresh, "dev1"); aerr == nil {
		t.Fatal("reused refresh token must be rejected")
	}
	if n := ts2.GetActiveRefreshTokenCount("did:key:zUser"); n != 0 {
		t.Fatalf("reuse should revoke the family, got %d active", n)
	}
}

func TestDurableRefresh_RevokeAllSurvivesRestart(t *testing.T) {
	shared := newFakeRedisBackend()
	ts1, _ := NewTokenService()
	ts1.SetRedisBackend(shared)
	ts1.StoreRefreshToken("did:key:zUser", GenerateRefreshToken(), "dev1")
	ts1.StoreRefreshToken("did:key:zUser", GenerateRefreshToken(), "dev2")

	ts2, _ := NewTokenService()
	ts2.SetRedisBackend(shared)
	if n := ts2.GetActiveRefreshTokenCount("did:key:zUser"); n != 2 {
		t.Fatalf("want 2 active across restart, got %d", n)
	}
	if revoked := ts2.RevokeAllUserTokens("did:key:zUser"); revoked != 2 {
		t.Fatalf("want 2 revoked, got %d", revoked)
	}
	if n := ts2.GetActiveRefreshTokenCount("did:key:zUser"); n != 0 {
		t.Fatalf("want 0 active after revoke, got %d", n)
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
