package passport

import (
	"testing"

	"github.com/thechadcromwell/echoapp/pkg/storage/encblob"
)

func TestSyncServicePushAndPull(t *testing.T) {
	store := NewMemSyncStore()
	svc := NewSyncService(store, encblob.NewStubStorage())
	ctx := t.Context()

	plain := []byte("aes-gcm-client-encrypted-wallet-blob")
	b64 := EncodeBase64(plain)

	meta, err := svc.Push(ctx, "did:key:zHolder", PushSyncRequest{CiphertextBase64: b64})
	if err != nil {
		t.Fatalf("Push: %v", err)
	}
	if meta.ByteSize != len(plain) {
		t.Fatalf("byte_size = %d, want %d", meta.ByteSize, len(plain))
	}
	if meta.StorageURI == "" {
		t.Fatal("expected storage_uri")
	}

	got, err := svc.Pull(ctx, "did:key:zHolder")
	if err != nil {
		t.Fatalf("Pull: %v", err)
	}
	if string(got.Ciphertext) != string(plain) {
		t.Fatalf("ciphertext mismatch")
	}
}

func TestSyncServicePushRejectsHashMismatch(t *testing.T) {
	svc := NewSyncService(NewMemSyncStore(), encblob.NewStubStorage())
	_, err := svc.Push(t.Context(), "did:key:zHolder", PushSyncRequest{
		CiphertextBase64: EncodeBase64([]byte("blob")),
		ContentHash:      "aabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccddaabbccdd",
	})
	if err != ErrSyncHashMismatch {
		t.Fatalf("expected ErrSyncHashMismatch, got %v", err)
	}
}
