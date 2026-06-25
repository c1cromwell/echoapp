package contacts

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func TestBlockWithoutPriorContact(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.CreateUser(ctx, &database.User{UserID: "u1", DID: "did:alice", Username: "alice"})
	db.CreateUser(ctx, &database.User{UserID: "u2", DID: "did:bob", Username: "bob"})

	if err := svc.BlockContact(ctx, "did:alice", "did:bob"); err != nil {
		t.Fatalf("BlockContact without prior row: %v", err)
	}

	blocked, err := svc.IsBlocked(ctx, "did:alice", "did:bob")
	if err != nil || !blocked {
		t.Fatalf("expected alice blocked bob, got blocked=%v err=%v", blocked, err)
	}

	list, err := svc.GetBlockedContacts(ctx, "did:alice")
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 blocked contact, got %d err=%v", len(list), err)
	}
}

func TestSearchExcludesBlocked(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.CreateUser(ctx, &database.User{UserID: "u1", DID: "did:alice", Username: "alice"})
	db.CreateUser(ctx, &database.User{UserID: "u2", DID: "did:bob", Username: "bob"})

	if err := svc.BlockContact(ctx, "did:alice", "did:bob"); err != nil {
		t.Fatal(err)
	}

	results, err := svc.SearchByUsername(ctx, "did:alice", "bob")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("blocked user should not appear in search, got %d", len(results))
	}
}

func TestProfilePrivacyFilter(t *testing.T) {
	svc, db := setupTest()
	ctx := context.Background()

	db.CreateUser(ctx, &database.User{UserID: "u1", DID: "did:alice", Username: "alice"})
	db.CreateUser(ctx, &database.User{UserID: "u2", DID: "did:bob", Username: "bob"})

	bio := "secret bio"
	_, err := svc.UpdateOwnProfile(ctx, "did:bob", nil, &bio, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.UpdatePrivacy(ctx, "did:bob", PrivacySettings{
		ShowLastSeen:       VisibilityNobody,
		ShowOnlineStatus:   VisibilityNobody,
		ShowProfilePicture: VisibilityNobody,
		ShowStatusMessage:  VisibilityNobody,
		AllowGroupInvites:  VisibilityNobody,
		AllowCalls:         VisibilityNobody,
		ShowTrustScore:     VisibilityNobody,
	})
	if err != nil {
		t.Fatal(err)
	}

	view, err := svc.GetProfile(ctx, "did:alice", "did:bob")
	if err != nil {
		t.Fatal(err)
	}
	if view.Bio != "" || view.StatusMessage != "" || view.AvatarURL != "" {
		t.Fatalf("non-contact should not see private fields: %+v", view)
	}
}
