package contacts

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func TestCommitDiscoveryKeyRespectsTierPolicy(t *testing.T) {
	ctx := context.Background()
	db := database.NewMemoryDB()
	svc := NewService(db)
	svc.SetOPRF(mustOPRF(t))

	trueVal := true
	_ = db.CreateUser(ctx, &database.User{DID: "did:key:tier3", Username: "tier3user", TrustTier: 3})
	_ = db.CreateUser(ctx, &database.User{DID: "did:key:tier1", Username: "tier1user", TrustTier: 1})
	_ = db.CreateUser(ctx, &database.User{
		DID: "did:key:tier1opt", Username: "tier1opt", TrustTier: 1, PhoneDiscoveryOptIn: &trueVal,
	})

	if err := svc.CommitDiscoveryKey(ctx, "key_tier3", "did:key:tier3"); err != nil {
		t.Fatal(err)
	}
	if err := svc.CommitDiscoveryKey(ctx, "key_tier1", "did:key:tier1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.CommitDiscoveryKey(ctx, "key_tier1opt", "did:key:tier1opt"); err != nil {
		t.Fatal(err)
	}

	index, err := svc.DiscoveryIndex(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if index["key_tier3"] != "did:key:tier3" {
		t.Fatal("tier 3 user should be discoverable")
	}
	if _, ok := index["key_tier1"]; ok {
		t.Fatal("tier 1 user should not be discoverable without opt-in")
	}
	if index["key_tier1opt"] != "did:key:tier1opt" {
		t.Fatal("tier 1 user with opt-in should be discoverable")
	}
}

func mustOPRF(t *testing.T) *OPRFService {
	t.Helper()
	o, err := NewOPRFService()
	if err != nil {
		t.Fatalf("NewOPRFService: %v", err)
	}
	return o
}
