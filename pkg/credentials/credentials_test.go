package credentials

import "testing"

func TestVC(t *testing.T) {
	vc := &VerifiableCredential{
		ID:     "test-vc-1",
		Issuer: "did:test:issuer",
	}
	if vc.ID == "" {
		t.Fatal("VC ID is empty")
	}
}
