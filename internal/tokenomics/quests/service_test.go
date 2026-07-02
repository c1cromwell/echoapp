package quests

import (
	"context"
	"testing"
)

func TestClaim_Idempotent(t *testing.T) {
	store := NewMemStore()
	svc := NewService(store)
	did := "did:key:z6MkQuestUser"
	questID := "identity_builder"

	_ = svc.MarkComplete(context.Background(), did, questID)
	tx1, err := svc.Claim(context.Background(), did, questID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Claim(context.Background(), did, questID)
	if err != ErrAlreadyClaimed {
		t.Errorf("expected already claimed, got %v (first tx %s)", err, tx1)
	}
}

func TestListCatalog(t *testing.T) {
	store := NewMemStore()
	svc := NewService(store)
	did := "did:key:z6MkQuestUser2"
	_ = svc.MarkComplete(context.Background(), did, "stack_and_earn")

	entries, err := svc.ListCatalog(context.Background(), did)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != len(All()) {
		t.Errorf("expected %d quests, got %d", len(All()), len(entries))
	}
	var found bool
	for _, e := range entries {
		if e.ID == "stack_and_earn" && e.CompletedAt != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected stack_and_earn completed")
	}
}

func TestClaim_TrustGate(t *testing.T) {
	store := NewMemStore()
	svc := NewService(store).WithTrustTier(func(_ context.Context, _ string) (int, error) {
		return 1, nil
	})
	did := "did:key:z6MkLowTrust"
	_ = svc.MarkComplete(context.Background(), did, "governance_debut")
	_, err := svc.Claim(context.Background(), did, "governance_debut")
	if err != ErrTrustTierTooLow {
		t.Errorf("expected trust tier error, got %v", err)
	}
}
