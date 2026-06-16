package comply

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func TestDECoverageRate_AfterLitigationHold(t *testing.T) {
	db := database.NewMemoryDB()
	svc := NewServiceLegacy(db, db, "tok")
	ctx := context.Background()

	_ = db.Enqueue(ctx, &database.QueuedMessage{
		MessageID: "m1", ConversationID: "c1",
		SenderDID: "did:custodian", RecipientDID: "did:peer",
	})
	_ = db.Enqueue(ctx, &database.QueuedMessage{
		MessageID: "m2", ConversationID: "c1",
		SenderDID: "did:peer", RecipientDID: "did:custodian",
	})

	_, err := svc.ActivateLitigationHold(ctx, ActivateLitigationHoldInput{
		OrgDID: "did:org:acme", MatterID: "matter-1",
		CustodianDIDs: []string{"did:custodian"}, ActivatedByDID: "did:admin",
	})
	if err != nil {
		t.Fatal(err)
	}

	stats, err := svc.Dashboard(ctx, "did:org:acme")
	if err != nil {
		t.Fatal(err)
	}
	if stats["deCoverageRate"] != "100%" {
		t.Fatalf("expected 100%% DE coverage after hold, got %v", stats["deCoverageRate"])
	}
}

func TestOrgRBAC_MemberCannotWrite(t *testing.T) {
	db := database.NewMemoryDB()
	svc := NewServiceLegacy(db, db, "tok")
	ctx := context.Background()
	_ = svc.EnsureOrgAdmin(ctx, "did:org:acme", "did:admin")
	_ = db.UpsertOrgMember(ctx, &database.OrgMember{
		OrgDID: "did:org:acme", MemberDID: "did:member", Role: database.OrgRoleMember,
	})
	if err := svc.AuthorizeOrgWrite(ctx, "did:org:acme", "did:member"); err == nil {
		t.Fatal("member should not write")
	}
	if err := svc.AuthorizeOrgRead(ctx, "did:org:acme", "did:member"); err != nil {
		t.Fatal("member should read")
	}
}
