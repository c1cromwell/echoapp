package messaging

import "testing"

func TestMinTTLForTrustTier(t *testing.T) {
	if got := MinTTLForTrustTier(1); got != 3600 {
		t.Fatalf("tier1 min = %d, want 3600", got)
	}
	if got := MinTTLForTrustTier(2); got != 300 {
		t.Fatalf("tier2 min = %d, want 300", got)
	}
	if got := MinTTLForTrustTier(3); got != 10 {
		t.Fatalf("tier3 min = %d, want 10", got)
	}
}

func TestValidateDisappearingTTL(t *testing.T) {
	if err := ValidateDisappearingTTL(1, 0); err != nil {
		t.Fatal("off should be allowed")
	}
	if err := ValidateDisappearingTTL(1, 30); err == nil {
		t.Fatal("tier1 should block 30s")
	}
	if err := ValidateDisappearingTTL(2, 300); err != nil {
		t.Fatalf("tier2 should allow 5m: %v", err)
	}
	if err := ValidateDisappearingTTL(3, 10); err != nil {
		t.Fatalf("tier3 should allow 10s: %v", err)
	}
}

func TestDisappearingRestrictionService_Appeal(t *testing.T) {
	svc := NewDisappearingRestrictionService()
	appeal := svc.SubmitAppeal("did:key:alice", "false positive")
	if appeal.Status != "pending" {
		t.Fatalf("status = %s", appeal.Status)
	}
	if len(svc.ListAppeals("did:key:alice")) != 1 {
		t.Fatal("expected one appeal")
	}
}
