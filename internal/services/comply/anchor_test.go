package comply

import (
	"context"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

func TestMetagraphAnchor_SubmitAndHealth(t *testing.T) {
	var posted bool
	client := metagraph.SubmitterFunc(func(_ context.Context, tx interface{}) (string, error) {
		// A compliance anchor is a trust commitment: a 32-byte hash and an epoch, nothing else.
		if _, ok := tx.(metagraph.DataL1TrustCommitmentUpdate); !ok {
			t.Errorf("anchored a %T, want a trust commitment", tx)
		}
		posted = true
		return "tx-abc", nil
	})
	anchor := NewMetagraphAnchor(client)

	commitment := policyAnchorRef("did:org:1", database.PolicyPermanent, "c1", "scope")
	tx, err := anchor.SubmitComplianceAnchor(context.Background(), commitment, time.Now().Unix())
	if err != nil {
		t.Fatal(err)
	}
	if !posted {
		t.Fatal("expected a chain submission")
	}
	if tx != "tx-abc" {
		t.Fatalf("tx = %q", tx)
	}
	if anchor.Status() != "healthy" {
		t.Fatalf("status = %q", anchor.Status())
	}
}

func TestComplianceAnchor_FallsBackWithoutDataL1(t *testing.T) {
	svc := NewService(database.NewMemoryDB(), Deps{})
	ref := svc.complianceAnchor(context.Background(), "did:org:1", "retention", "p1", "scope")
	if len(ref) != 64 {
		t.Fatalf("expected 64-char commitment, got %q", ref)
	}
}

func TestAnchorHealthStatus_NoDataL1(t *testing.T) {
	svc := NewService(database.NewMemoryDB(), Deps{})
	if svc.anchorHealthStatus() != "healthy" {
		t.Fatalf("expected healthy local mode")
	}
}
