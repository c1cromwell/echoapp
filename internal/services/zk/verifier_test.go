package zk

import "testing"

func TestVerifier_CommitmentFallback(t *testing.T) {
	v := NewVerifier()
	res, err := v.Verify(VerifyRequest{
		SubjectDID: "did:key:alice",
		ClaimType:  "tier3",
		Proof:      "ignored",
		Nonce:      "nonce-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Verified || res.Mode != "commitment_fallback" {
		t.Fatalf("unexpected result: %+v", res)
	}
}
