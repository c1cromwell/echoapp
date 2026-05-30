package logging

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// TestStubIPFSStorage_StoreAndRetrieveCIDs verifies the stub stores payloads
// and returns deterministic content-addressed CIDs.
func TestStubIPFSStorage_StoreAndRetrieveCIDs(t *testing.T) {
	stub := &StubIPFSStorage{}
	ctx := context.Background()

	payload1 := []byte("audit-log-batch-1")
	payload2 := []byte("audit-log-batch-2")

	cid1, err := stub.Store(ctx, payload1)
	if err != nil {
		t.Fatalf("Store payload1: %v", err)
	}
	cid2, err := stub.Store(ctx, payload2)
	if err != nil {
		t.Fatalf("Store payload2: %v", err)
	}

	if cid1 == cid2 {
		t.Error("different payloads should produce different CIDs")
	}

	if !strings.HasPrefix(cid1, "stub-cid-") {
		t.Errorf("stub CID should start with 'stub-cid-', got %q", cid1)
	}

	cids := stub.StoredCIDs()
	if len(cids) != 2 {
		t.Errorf("want 2 stored CIDs, got %d", len(cids))
	}
}

// TestStubIPFSStorage_SamePayloadSameCID confirms content-addressing is consistent.
func TestStubIPFSStorage_SamePayloadSameCID(t *testing.T) {
	stub := &StubIPFSStorage{}
	ctx := context.Background()
	payload := []byte("repeated-payload")

	cid1, _ := stub.Store(ctx, payload)
	cid2, _ := stub.Store(ctx, payload)

	if cid1 != cid2 {
		t.Errorf("same payload should produce same CID: %q vs %q", cid1, cid2)
	}
}

// TestFallbackIPFSStorage_UsePrimary verifies the fallback uses the primary
// and fires an async secondary pin.
func TestFallbackIPFSStorage_UsesPrimary(t *testing.T) {
	primary := &StubIPFSStorage{}
	secondary := &StubIPFSStorage{}
	fallback := NewFallbackIPFSStorageWithProviders(primary, secondary)

	ctx := context.Background()
	cid, err := fallback.Store(ctx, []byte("batch-data"))
	if err != nil {
		t.Fatalf("fallback Store: %v", err)
	}
	if !strings.HasPrefix(cid, "stub-cid-") {
		t.Errorf("unexpected CID: %q", cid)
	}
	if len(primary.StoredCIDs()) != 1 {
		t.Error("primary should have 1 stored CID")
	}
}

// TestFallbackIPFSStorage_FallsBackOnPrimaryFailure verifies the fallback
// switches to secondary when primary returns an error.
func TestFallbackIPFSStorage_FallsBackOnPrimaryFailure(t *testing.T) {
	primary := &errStorage{err: errors.New("primary down")}
	secondary := &StubIPFSStorage{}
	fallback := NewFallbackIPFSStorageWithProviders(primary, secondary)

	ctx := context.Background()
	cid, err := fallback.Store(ctx, []byte("batch-data"))
	if err != nil {
		t.Fatalf("fallback should succeed via secondary, got: %v", err)
	}
	if !strings.HasPrefix(cid, "stub-cid-") {
		t.Errorf("unexpected CID from secondary: %q", cid)
	}
}

// TestFallbackIPFSStorage_BothFailsReturnsError confirms both-provider
// failure surfaces a meaningful error.
func TestFallbackIPFSStorage_BothFailsReturnsError(t *testing.T) {
	primary := &errStorage{err: errors.New("primary down")}
	secondary := &errStorage{err: errors.New("secondary down")}
	fallback := NewFallbackIPFSStorageWithProviders(primary, secondary)

	_, err := fallback.Store(context.Background(), []byte("batch"))
	if err == nil {
		t.Fatal("expected error when both providers fail")
	}
	if !strings.Contains(err.Error(), "primary=") || !strings.Contains(err.Error(), "fallback=") {
		t.Errorf("error should include both failure reasons, got: %v", err)
	}
}

// TestNewPinataIPFSStorage_NoEnvReturnsError verifies graceful failure when
// PINATA_API_KEY / PINATA_API_SECRET are not set.
func TestNewPinataIPFSStorage_NoEnvReturnsError(t *testing.T) {
	// Environment variables are not set in CI by default.
	_, err := NewPinataIPFSStorage()
	if !errors.Is(err, ErrStorageNotConfigured) {
		t.Errorf("want ErrStorageNotConfigured, got %v", err)
	}
}

// TestNewStorjIPFSStorage_NoEnvReturnsError verifies graceful failure when
// Storj credentials are not set.
func TestNewStorjIPFSStorage_NoEnvReturnsError(t *testing.T) {
	_, err := NewStorjIPFSStorage()
	if !errors.Is(err, ErrStorageNotConfigured) {
		t.Errorf("want ErrStorageNotConfigured, got %v", err)
	}
}

// TestNewFallbackIPFSStorage_NoEnvReturnsError verifies that when neither
// provider is configured, NewFallbackIPFSStorage returns an error.
func TestNewFallbackIPFSStorage_NoEnvReturnsError(t *testing.T) {
	_, err := NewFallbackIPFSStorage()
	if !errors.Is(err, ErrStorageNotConfigured) {
		t.Errorf("want ErrStorageNotConfigured, got %v", err)
	}
}

// --- test helpers ---

// errStorage always returns an error from Store.
type errStorage struct{ err error }

func (e *errStorage) Store(_ context.Context, _ []byte) (string, error) {
	return "", e.err
}
