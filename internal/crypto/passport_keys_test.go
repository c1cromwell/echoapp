package crypto

import (
	"bytes"
	"testing"
)

func TestDerivePassportCredentialSyncKey_Deterministic(t *testing.T) {
	kds := NewKeyDerivationService()
	root := bytes.Repeat([]byte{0xab}, 32)

	k1, err := kds.DerivePassportCredentialSyncKey(root)
	if err != nil {
		t.Fatal(err)
	}
	k2, err := kds.DerivePassportCredentialSyncKey(root)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(k1, k2) {
		t.Fatal("expected deterministic derivation")
	}
	if len(k1) != 32 {
		t.Fatalf("key len = %d", len(k1))
	}
}

func TestDerivePassportCredentialSyncKey_RejectsShortRoot(t *testing.T) {
	kds := NewKeyDerivationService()
	if _, err := kds.DerivePassportCredentialSyncKey([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected error for short root key")
	}
}
