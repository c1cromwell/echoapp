package recovery

import (
	"context"
	"crypto/rand"
	"strings"
	"testing"
	"time"
)

func TestShamirSplitCombineRoundTrip(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	shares, err := SplitRecoverySecret(secret, DefaultTotal, DefaultThreshold)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if len(shares) != DefaultTotal {
		t.Fatalf("want %d shares, got %d", DefaultTotal, len(shares))
	}
	got, err := CombineRecoveryShares(shares[:DefaultThreshold])
	if err != nil {
		t.Fatalf("Combine: %v", err)
	}
	if string(got) != string(secret) {
		t.Fatal("reconstructed secret mismatch")
	}
}

func TestShamirMMinusOneSharesRevealNothing(t *testing.T) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	shares, err := SplitRecoverySecret(secret, 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	_, err = CombineRecoveryShares(shares[:1])
	if err == nil {
		t.Fatal("expected error combining fewer than threshold shares")
	}
}

func TestSetupInitiateComplete(t *testing.T) {
	store := NewMemStore()
	svc := NewService(store)
	ctx := context.Background()
	holder := "did:key:zHolderRecovery"

	policy, shareholders, err := svc.Setup(ctx, holder, SetupRequest{
		Threshold: 2,
		Total:     3,
		Shareholders: []ShareholderInput{
			{ShareIndex: 1, GuardianDID: "did:key:zDevice1", Role: RoleDevice, Status: StatusActive},
			{ShareIndex: 2, GuardianDID: "did:key:zDevice2", Role: RoleDevice, Status: StatusActive},
			{ShareIndex: 3, GuardianDID: "did:key:zContact", Role: RoleContact, Status: StatusActive},
		},
	})
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if policy.Threshold != 2 || len(shareholders) != 3 {
		t.Fatalf("unexpected setup result")
	}

	resp, err := svc.Initiate(ctx, holder)
	if err != nil {
		t.Fatalf("Initiate: %v", err)
	}
	if resp.Session.SessionID == "" {
		t.Fatal("expected session id")
	}

	rootKey := []byte("passport-root-key-32-bytes!!!!!")
	commitment := RootKeyCommitment(rootKey)
	session, err := svc.Complete(ctx, holder, CompleteRequest{
		SessionID:         resp.Session.SessionID,
		RootKeyCommitment: commitment,
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if session.Status != SessionCompleted {
		t.Fatalf("status = %q", session.Status)
	}
}

func TestRejectSecretFields(t *testing.T) {
	err := RejectSecretFields(map[string]interface{}{"recovery_secret": "nope"})
	if err == nil {
		t.Fatal("expected honeypot rejection")
	}
}

func TestGuardianCredentialJSON(t *testing.T) {
	raw, err := GuardianCredentialJSON(
		"did:key:zHolder",
		"did:key:zGuardian",
		2,
		mustParseTime(t, "2026-05-29T00:00:00Z"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(raw, "EchoGuardianCredential") || !strings.Contains(raw, "guardianFor") {
		t.Fatalf("unexpected VC json: %s", raw)
	}
}

func mustParseTime(t *testing.T, s string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}
