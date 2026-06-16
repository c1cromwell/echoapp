package comply

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func TestBlocksDeletion_LitigationHold(t *testing.T) {
	db := database.NewMemoryDB()
	svc := NewServiceLegacy(db, db, "tok")
	ctx := context.Background()

	_, err := svc.CreateRetentionPolicy(ctx, CreatePolicyInput{
		OrgDID:         "did:org:acme",
		PolicyType:     database.PolicyLitigationHold,
		ConversationID: "c1",
		CreatedByDID:   "did:admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !svc.BlocksDeletion(ctx, "c1") {
		t.Fatal("litigation hold should block deletion")
	}
	if !svc.BlocksDisappearing(ctx, "c1") {
		t.Fatal("litigation hold should block disappearing messages")
	}
}

func TestReleaseConversation_ClearsEnforcement(t *testing.T) {
	db := database.NewMemoryDB()
	svc := NewServiceLegacy(db, db, "tok")
	ctx := context.Background()

	_, _ = svc.CreateRetentionPolicy(ctx, CreatePolicyInput{
		OrgDID:         "did:org:acme",
		PolicyType:     database.PolicyPermanent,
		ConversationID: "c2",
		CreatedByDID:   "did:admin",
	})
	if err := svc.ReleaseConversation(ctx, "c2"); err != nil {
		t.Fatal(err)
	}
	if svc.BlocksDeletion(ctx, "c2") {
		t.Fatal("released conversation should not block deletion")
	}
}
