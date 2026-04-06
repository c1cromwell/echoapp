package contacts

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func setupTest() (*Service, database.DB) {
	db := database.NewMemoryDB()
	svc := NewService(db)
	return svc, db
}

func TestPhoneHashing(t *testing.T) {
	h1 := HashPhone("+1 555-0100")
	h2 := HashPhone("+15550100")
	if string(h1) != string(h2) {
		t.Error("phone normalization should produce same hash for equivalent numbers")
	}
}

func TestPSIDiscovery(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	result, err := svc.PSIDiscovery(ctx, "did:alice", []string{"hash1", "hash2"})
	if err != nil {
		t.Fatalf("PSIDiscovery: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestPSIDiscoveryRateLimit(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	hashes := make([]string, 1001)
	_, err := svc.PSIDiscovery(ctx, "did:alice", hashes)
	if err != ErrRateLimited {
		t.Errorf("expected ErrRateLimited, got %v", err)
	}
}

func TestSearchByUsername(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.CreateUser(ctx, &database.User{
		UserID: "user-1", DID: "did:bob", Username: "bob",
	})

	results, err := svc.SearchByUsername(ctx, "did:alice", "bob")
	if err != nil {
		t.Fatalf("SearchByUsername: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0]["did"] != "did:bob" {
		t.Errorf("expected did:bob, got %v", results[0]["did"])
	}
}

func TestSearchByUsernameSelfFilter(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.CreateUser(ctx, &database.User{
		UserID: "user-1", DID: "did:alice", Username: "alice",
	})

	results, _ := svc.SearchByUsername(ctx, "did:alice", "alice")
	if len(results) != 0 {
		t.Errorf("should not return self in search results")
	}
}

func TestAddContact(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.CreateUser(ctx, &database.User{
		UserID: "user-1", DID: "did:alice", Username: "alice",
	})

	contact, err := svc.AddContact(ctx, "did:alice", "did:bob", "manual")
	if err != nil {
		t.Fatalf("AddContact: %v", err)
	}
	if contact.ContactDID != "did:bob" {
		t.Errorf("expected did:bob, got %s", contact.ContactDID)
	}
}

func TestAddContactSelf(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	_, err := svc.AddContact(ctx, "did:alice", "did:alice", "manual")
	if err != ErrSelfContact {
		t.Errorf("expected ErrSelfContact, got %v", err)
	}
}

func TestTier1Limit(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.CreateUser(ctx, &database.User{
		UserID: "user-1", DID: "did:alice", Username: "alice",
	})

	// Add 10 contacts (tier 1 limit)
	for i := 0; i < 10; i++ {
		did := "did:contact-" + string(rune('a'+i))
		svc.AddContact(ctx, "did:alice", did, "manual")
	}

	// 11th should fail
	_, err := svc.AddContact(ctx, "did:alice", "did:contact-overflow", "manual")
	if err != ErrTier1Limit {
		t.Errorf("expected ErrTier1Limit, got %v", err)
	}
}

func TestBlockAndUnblock(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.CreateUser(ctx, &database.User{UserID: "user-1", DID: "did:alice", Username: "alice"})
	svc.AddContact(ctx, "did:alice", "did:bob", "manual")

	if err := svc.BlockContact(ctx, "did:alice", "did:bob"); err != nil {
		t.Fatalf("BlockContact: %v", err)
	}

	contacts, _ := svc.GetContacts(ctx, "did:alice")
	if len(contacts) == 0 || !contacts[0].Blocked {
		t.Error("expected contact to be blocked")
	}

	if err := svc.UnblockContact(ctx, "did:alice", "did:bob"); err != nil {
		t.Fatalf("UnblockContact: %v", err)
	}
	contacts, _ = svc.GetContacts(ctx, "did:alice")
	if contacts[0].Blocked {
		t.Error("expected contact to be unblocked")
	}
}

func TestGetContactsWithTrustBadge(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.SetTrustScore(ctx, &database.TrustScore{DID: "did:bob", Score: 80, Tier: 4})
	svc.AddContact(ctx, "did:alice", "did:bob", "manual")

	contacts, _ := svc.GetContacts(ctx, "did:alice")
	if len(contacts) == 0 {
		t.Fatal("expected contacts")
	}
	if contacts[0].TrustBadge != "verified" {
		t.Errorf("expected 'verified' badge for tier 4, got %s", contacts[0].TrustBadge)
	}
}

func TestInviteFlow(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	invite, err := svc.CreateInviteLink(ctx, "did:alice")
	if err != nil {
		t.Fatalf("CreateInviteLink: %v", err)
	}
	if invite.Code == "" {
		t.Fatal("expected non-empty invite code")
	}

	// Verify
	verified, err := svc.VerifyInvite(ctx, invite.Code)
	if err != nil {
		t.Fatalf("VerifyInvite: %v", err)
	}
	if verified.CreatorDID != "did:alice" {
		t.Errorf("expected did:alice, got %s", verified.CreatorDID)
	}

	// Accept
	accepted, err := svc.AcceptInvite(ctx, invite.Code, "did:bob")
	if err != nil {
		t.Fatalf("AcceptInvite: %v", err)
	}
	if !accepted.Accepted {
		t.Error("expected invite to be accepted")
	}

	// Double accept should fail
	_, err = svc.AcceptInvite(ctx, invite.Code, "did:charlie")
	if err == nil {
		t.Error("expected error on double accept")
	}
}

func TestSelfInvite(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	invite, _ := svc.CreateInviteLink(ctx, "did:alice")
	_, err := svc.AcceptInvite(ctx, invite.Code, "did:alice")
	if err != ErrSelfContact {
		t.Errorf("expected ErrSelfContact, got %v", err)
	}
}

func TestInvalidInviteCode(t *testing.T) {
	svc, _ := setupTest()
	ctx := context.Background()

	_, err := svc.VerifyInvite(ctx, "nonexistent")
	if err != ErrInvalidInvite {
		t.Errorf("expected ErrInvalidInvite, got %v", err)
	}
}
