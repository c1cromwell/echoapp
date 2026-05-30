package encblob

import (
	"context"
	"testing"
)

func TestStubStorage_StoreAndRetrieve(t *testing.T) {
	stub := NewStubStorage()
	ctx := context.Background()
	payload := []byte("client-encrypted-credential-blob")

	uri, err := stub.Store(ctx, payload)
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if uri == "" {
		t.Fatal("expected non-empty URI")
	}

	got, err := stub.Retrieve(ctx, uri)
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("payload mismatch")
	}
}

func TestNewFallbackStorage_NoEnvReturnsError(t *testing.T) {
	t.Setenv("PINATA_API_KEY", "")
	t.Setenv("PINATA_API_SECRET", "")
	t.Setenv("STORJ_ACCESS_KEY", "")
	t.Setenv("STORJ_SECRET_KEY", "")
	t.Setenv("STORJ_BUCKET", "")

	_, err := NewFallbackStorage()
	if err != ErrStorageNotConfigured {
		t.Fatalf("expected ErrStorageNotConfigured, got %v", err)
	}
}
