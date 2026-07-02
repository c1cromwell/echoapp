package zk

import (
	"testing"
)

func TestVerifier_CommitmentFallback(t *testing.T) {
	v := NewVerifier()
	nonce := "nonce-1"
	proof := commitmentHash("did:key:alice", "tier3", nonce)
	res, err := v.Verify(VerifyRequest{
		SubjectDID: "did:key:alice",
		ClaimType:  "tier3",
		Proof:      proof,
		Nonce:      nonce,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Verified || res.Mode != "commitment_fallback" {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestVerifier_RejectsBadProof(t *testing.T) {
	v := NewVerifier()
	res, err := v.Verify(VerifyRequest{
		SubjectDID: "did:key:alice",
		ClaimType:  "tier3",
		Proof:      "deadbeef",
		Nonce:      "nonce-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Verified {
		t.Fatal("expected rejection")
	}
}
