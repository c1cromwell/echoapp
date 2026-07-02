package vip

import (
	"context"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/infra"
)

func TestAuthorizeVIP(t *testing.T) {
	svc := NewSubscriptionService("did:key:z6MkEchoBackend")
	rec, err := svc.AuthorizeVIP(context.Background(), "did:key:z6MkVIPUser")
	if err != nil {
		t.Fatal(err)
	}
	if rec.MaxAmount != VIPMonthlyECHO {
		t.Errorf("max amount: got %d want %d", rec.MaxAmount, VIPMonthlyECHO)
	}
	if rec.Purpose != PurposeVIPSubscription {
		t.Errorf("purpose: got %s", rec.Purpose)
	}
	if svc.SubscriptionTier(rec.OwnerDID) != infra.TierVIP {
		t.Error("expected VIP tier after authorize")
	}
}

func TestRenewSpend_RequiresAllowSpend(t *testing.T) {
	svc := NewSubscriptionService("did:key:z6MkEchoBackend")
	_, err := svc.RenewSpend(context.Background(), "did:key:z6MkUnknown")
	if err != ErrAllowSpendNotFound {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestAllowSpendExpiry(t *testing.T) {
	svc := NewSubscriptionService("did:key:z6MkEchoBackend")
	svc.mu.Lock()
	svc.records["did:key:z6MkExpired"] = &AllowSpendRecord{
		OwnerDID:  "did:key:z6MkExpired",
		MaxAmount: VIPMonthlyECHO,
		ExpiresAt: time.Now().Add(-time.Hour),
		Active:    true,
	}
	svc.mu.Unlock()
	_, err := svc.RenewSpend(context.Background(), "did:key:z6MkExpired")
	if err != ErrAllowSpendExpired {
		t.Errorf("expected expired, got %v", err)
	}
}
