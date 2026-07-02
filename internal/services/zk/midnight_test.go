package zk

import "testing"

func TestVerifier_MidnightEnvelope(t *testing.T) {
	v := NewVerifier()
	nonce := "nonce-midnight"
	proof := BuildMidnightEnvelope("did:key:alice", "tier3", nonce, "circuit-proof")
	res, err := v.Verify(VerifyRequest{
		SubjectDID: "did:key:alice",
		ClaimType:  "tier3",
		Proof:      proof,
		Nonce:      nonce,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Verified || res.Mode != "midnight" {
		t.Fatalf("unexpected: %+v", res)
	}
}

func TestVerifier_MidnightRejectsBadCommitment(t *testing.T) {
	v := NewVerifier()
	res, err := v.Verify(VerifyRequest{
		SubjectDID: "did:key:alice",
		ClaimType:  "tier3",
		Proof:      "midnight:" + `{"commitment":"deadbeef","proof":"x"}`,
		Nonce:      "n1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Verified {
		t.Fatal("expected rejection")
	}
}
