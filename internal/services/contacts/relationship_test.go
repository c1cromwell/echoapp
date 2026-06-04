package contacts

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func TestMutualContacts(t *testing.T) {
	db := database.NewMemoryDB()
	ctx := context.Background()
	svc := NewService(db)

	_ = db.CreateUser(ctx, &database.User{DID: "did:key:alice", Username: "alice", TrustTier: 2})
	_ = db.CreateUser(ctx, &database.User{DID: "did:key:bob", Username: "bob", TrustTier: 2})
	_ = db.CreateUser(ctx, &database.User{DID: "did:key:carol", Username: "carol", TrustTier: 2})
	_ = db.CreateUser(ctx, &database.User{DID: "did:key:dave", Username: "dave", TrustTier: 2})

	_, _ = svc.AddContact(ctx, "did:key:alice", "did:key:bob", "manual")
	_, _ = svc.AddContact(ctx, "did:key:alice", "did:key:carol", "manual")
	_, _ = svc.AddContact(ctx, "did:key:bob", "did:key:alice", "manual")
	_, _ = svc.AddContact(ctx, "did:key:bob", "did:key:carol", "manual")
	_, _ = svc.AddContact(ctx, "did:key:bob", "did:key:dave", "manual")

	mutual, err := svc.MutualContacts(ctx, "did:key:alice", "did:key:bob", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(mutual) != 1 || mutual[0].DID != "did:key:carol" {
		t.Fatalf("expected mutual carol, got %+v", mutual)
	}
	if mutual[0].Username != "carol" {
		t.Fatalf("expected username carol, got %q", mutual[0].Username)
	}
}
