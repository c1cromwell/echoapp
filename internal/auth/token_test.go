package auth

import (
	"testing"
	"time"
)

func newTestTokenService(t *testing.T) *TokenService {
	t.Helper()
	ts, err := NewTokenService()
	if err != nil {
		t.Fatalf("create token service: %v", err)
	}
	return ts
}

func TestTokenService_IssueAndValidate(t *testing.T) {
	ts := newTestTokenService(t)

	token, claims, err := ts.IssueAccessToken("did:prism:abc", "device-hash-1", 2, "messaging payments")
	if err != nil {
		t.Fatalf("issue error: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}

	validated, err := ts.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}

	if validated.Subject != "did:prism:abc" {
		t.Errorf("expected subject did:prism:abc, got %s", validated.Subject)
	}
	if validated.DeviceID != "device-hash-1" {
		t.Errorf("expected device-hash-1, got %s", validated.DeviceID)
	}
	if validated.TrustTier != 2 {
		t.Errorf("expected trust tier 2, got %d", validated.TrustTier)
	}
	if validated.TokenID != claims.TokenID {
		t.Errorf("token ID mismatch")
	}
}

func TestTokenService_ExpiredToken(t *testing.T) {
	ts := newTestTokenService(t)

	// Issue token with immediate expiry
	now := time.Now()
	claims := &TokenClaims{
		Subject:   "did:prism:abc",
		IssuedAt:  now.Add(-10 * time.Minute).Unix(),
		ExpiresAt: now.Add(-5 * time.Minute).Unix(), // Expired 5 min ago
		TokenID:   "test-jti",
		DeviceID:  "device-1",
		TrustTier: 0,
		Scope:     "messaging",
	}

	token, err := ts.signClaims(claims)
	if err != nil {
		t.Fatalf("sign error: %v", err)
	}

	_, err = ts.ValidateAccessToken(token)
	if err == nil {
		t.Error("expired token should fail validation")
	}
}

func TestTokenService_BlocklistedToken(t *testing.T) {
	ts := newTestTokenService(t)

	token, claims, _ := ts.IssueAccessToken("did:prism:abc", "device-1", 0, "messaging")

	// Blocklist it
	ts.BlocklistToken(claims.TokenID, time.Now().Add(time.Hour))

	_, err := ts.ValidateAccessToken(token)
	if err == nil {
		t.Error("blocklisted token should fail validation")
	}
}

func TestTokenService_TamperedToken(t *testing.T) {
	ts := newTestTokenService(t)

	token, _, _ := ts.IssueAccessToken("did:prism:abc", "device-1", 0, "messaging")

	// Tamper with the payload
	parts := splitToken(token)
	tampered := parts[0] + "." + parts[1] + "x" + "." + parts[2]

	_, err := ts.ValidateAccessToken(tampered)
	if err == nil {
		t.Error("tampered token should fail validation")
	}
}

func TestTokenService_TempAccessToken(t *testing.T) {
	ts := newTestTokenService(t)

	token, claims, err := ts.IssueTempAccessToken("did:prism:pending", "device-1")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if claims.Scope != "passkey_registration" {
		t.Errorf("expected passkey_registration scope, got %s", claims.Scope)
	}
}

func TestTokenService_ElevatedToken(t *testing.T) {
	ts := newTestTokenService(t)

	token, claims, err := ts.IssueElevatedToken("did:prism:abc", "device-1", 2, "revoke_device", 5*time.Minute)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if token == "" {
		t.Fatal("token should not be empty")
	}
	if !claims.Elevated {
		t.Error("elevated should be true")
	}
	if claims.ElevatedAction != "revoke_device" {
		t.Errorf("expected revoke_device, got %s", claims.ElevatedAction)
	}
}

func TestRefreshToken_Generation(t *testing.T) {
	token := GenerateRefreshToken()
	if token[:3] != "rt_" {
		t.Errorf("refresh token should start with 'rt_', got %s", token[:3])
	}
	if len(token) < 10 {
		t.Error("refresh token too short")
	}
}

func TestRefreshToken_HashDeterministic(t *testing.T) {
	token := "rt_test-token-123"
	h1 := HashRefreshToken(token)
	h2 := HashRefreshToken(token)
	if h1 != h2 {
		t.Error("same token should produce same hash")
	}

	h3 := HashRefreshToken("rt_different-token")
	if h1 == h3 {
		t.Error("different tokens should produce different hashes")
	}
}

func TestTokenService_RefreshTokenRotation(t *testing.T) {
	ts := newTestTokenService(t)

	// Store initial token
	token := GenerateRefreshToken()
	ts.StoreRefreshToken("user-1", token, "device-1")

	// Rotate
	newToken, record, err := ts.RotateRefreshToken(token, "device-1")
	if err != nil {
		t.Fatalf("rotate error: %v", err)
	}
	if newToken == "" {
		t.Fatal("new token should not be empty")
	}
	if record.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", record.UserID)
	}

	// Old token should no longer work
	_, _, err = ts.RotateRefreshToken(token, "device-1")
	if err == nil {
		t.Error("old token should be rejected (already used)")
	}
}

func TestTokenService_RefreshTokenReplayDetection(t *testing.T) {
	ts := newTestTokenService(t)

	// Store initial token
	token := GenerateRefreshToken()
	ts.StoreRefreshToken("user-1", token, "device-1")

	// First rotation succeeds
	newToken, _, err := ts.RotateRefreshToken(token, "device-1")
	if err != nil {
		t.Fatalf("first rotation should succeed: %v", err)
	}

	// Replay of old token — should revoke ALL sessions for this user
	_, _, err = ts.RotateRefreshToken(token, "device-1")
	if err == nil {
		t.Fatal("replay should fail")
	}
	if err.Code != ErrCodeRefreshInvalid {
		t.Errorf("expected AUTH_006, got %s", err.Code)
	}

	// Even the legitimate new token should be revoked (nuclear option)
	_, _, err = ts.RotateRefreshToken(newToken, "device-1")
	if err == nil {
		t.Error("all tokens should be revoked after replay detection")
	}
}

func TestTokenService_RefreshTokenDeviceBinding(t *testing.T) {
	ts := newTestTokenService(t)

	token := GenerateRefreshToken()
	ts.StoreRefreshToken("user-1", token, "device-1")

	// Try rotating with wrong device
	_, _, err := ts.RotateRefreshToken(token, "device-WRONG")
	if err == nil {
		t.Error("wrong device should be rejected")
	}
	if err.Code != ErrCodeUnknownDevice {
		t.Errorf("expected AUTH_007, got %s", err.Code)
	}
}

func TestTokenService_RevokeAllUserTokens(t *testing.T) {
	ts := newTestTokenService(t)

	// Create multiple tokens for same user
	for i := 0; i < 5; i++ {
		token := GenerateRefreshToken()
		ts.StoreRefreshToken("user-1", token, "device-1")
	}

	if ts.GetActiveRefreshTokenCount("user-1") != 5 {
		t.Fatalf("expected 5 active tokens")
	}

	count := ts.RevokeAllUserTokens("user-1")
	if count != 5 {
		t.Errorf("expected 5 revoked, got %d", count)
	}

	if ts.GetActiveRefreshTokenCount("user-1") != 0 {
		t.Error("should have 0 active tokens after revoke all")
	}
}

func TestTokenService_NonceAntiReplay(t *testing.T) {
	ts := newTestTokenService(t)

	// First use should succeed
	if !ts.CheckAndStoreNonce("nonce-1") {
		t.Error("first use of nonce should succeed")
	}

	// Second use should fail
	if ts.CheckAndStoreNonce("nonce-1") {
		t.Error("reused nonce should fail")
	}

	// Different nonce should succeed
	if !ts.CheckAndStoreNonce("nonce-2") {
		t.Error("different nonce should succeed")
	}
}

func TestTokenService_ConcurrentRefresh(t *testing.T) {
	ts := newTestTokenService(t)

	token := GenerateRefreshToken()
	ts.StoreRefreshToken("user-1", token, "device-1")

	// Simulate two concurrent rotations
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			_, _, err := ts.RotateRefreshToken(token, "device-1")
			if err != nil {
				results <- err
			} else {
				results <- nil
			}
		}()
	}

	// Collect results
	successes := 0
	failures := 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			successes++
		} else {
			failures++
		}
	}

	// Exactly one should succeed, one should fail
	if successes != 1 {
		t.Errorf("expected 1 success in concurrent refresh, got %d", successes)
	}
	if failures != 1 {
		t.Errorf("expected 1 failure in concurrent refresh, got %d", failures)
	}
}
