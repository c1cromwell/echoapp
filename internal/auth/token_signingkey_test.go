package auth

import (
	"testing"
)

// a valid 32-byte (64 hex) P-256 scalar, < curve order N
const testSigningKeyHex = "4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318"

// TestSigningKey_PersistentAcrossInstances is the regression guard for S2:
// two TokenService instances built from the same JWT_SIGNING_KEY must validate
// each other's tokens (i.e. a restart does not invalidate sessions).
func TestSigningKey_PersistentAcrossInstances(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", testSigningKeyHex)

	ts1, err := NewTokenService()
	if err != nil {
		t.Fatalf("ts1: %v", err)
	}
	token, _, err := ts1.IssueAccessToken("did:key:zAlice", "dev1", 3, "messaging")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	ts2, err := NewTokenService() // simulates a restart with the same configured key
	if err != nil {
		t.Fatalf("ts2: %v", err)
	}
	if _, err := ts2.ValidateAccessToken(token); err != nil {
		t.Fatalf("token issued by ts1 must validate on ts2 (same key), got: %v", err)
	}
	if ts1.KeyID() != ts2.KeyID() {
		t.Fatalf("stable keyID expected across instances: %q vs %q", ts1.KeyID(), ts2.KeyID())
	}
}

func TestSigningKey_ProductionRequiresKey(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "")
	t.Setenv("ENVIRONMENT", "production")

	if _, err := NewTokenService(); err == nil {
		t.Fatal("expected NewTokenService to fail in production without JWT_SIGNING_KEY")
	}
}

func TestSigningKey_InvalidRejected(t *testing.T) {
	for _, bad := range []string{"nothex", "abcd", "00"} {
		t.Setenv("JWT_SIGNING_KEY", bad)
		if _, err := NewTokenService(); err == nil {
			t.Fatalf("expected rejection of invalid JWT_SIGNING_KEY %q", bad)
		}
	}
}

func TestSigningKey_EphemeralInDev(t *testing.T) {
	t.Setenv("JWT_SIGNING_KEY", "")
	t.Setenv("ENVIRONMENT", "development")
	if _, err := NewTokenService(); err != nil {
		t.Fatalf("dev should fall back to an ephemeral key, got: %v", err)
	}
}
