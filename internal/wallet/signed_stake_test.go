package wallet

import (
	"context"
	"testing"
)

// SubmitSignedTokenLock must relay the client-signed payload (as "tokenLock") to
// the metagraph and mirror the position into the store, without the backend
// re-originating the transaction.
func TestSubmitSignedTokenLockForwardsAndMirrors(t *testing.T) {
	t.Setenv("ECHO_WALLET_GENESIS_AUTO", "1")
	ctx := context.Background()
	store := NewMemStore()
	sub := &fakeSubmitter{signedHash: "tl_signed_1"}
	q := NewLedgerQuerier(store, sub)
	did := "did:echo:signed"

	// Seed balance.
	if _, err := store.GetBalance(ctx, did); err != nil {
		t.Fatalf("seed balance: %v", err)
	}
	tier, err := ValidateTier("bronze")
	if err != nil {
		t.Fatalf("tier: %v", err)
	}
	signed := []byte(`{"value":{"amount":10000000000,"source":"DAGx"},"proofs":[{"id":"ab","signature":"3045"}]}`)

	hash, err := q.SubmitSignedTokenLock(ctx, did, 100*DatumPerECHO, tier, signed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash != "tl_signed_1" {
		t.Fatalf("hash = %q", hash)
	}
	if sub.lastSignedType != "tokenLock" || string(sub.lastSigned) != string(signed) {
		t.Fatalf("forward mismatch: type=%q body=%s", sub.lastSignedType, sub.lastSigned)
	}
	locks, _ := store.ListLocks(ctx, did)
	if len(locks) != 1 || locks[0].ID != hash {
		t.Fatalf("expected mirrored lock with id %q, got %+v", hash, locks)
	}
}

// Without a Currency L1 submitter, signed submission must fail (not silently
// fall back to a local hash).
func TestSubmitSignedTokenLockRequiresSubmitter(t *testing.T) {
	q := NewLedgerQuerier(NewMemStore(), nil)
	tier, _ := ValidateTier("bronze")
	_, err := q.SubmitSignedTokenLock(context.Background(), "did:echo:x", 100*DatumPerECHO, tier, []byte(`{}`))
	if err != ErrSignedSubmitUnavailable {
		t.Fatalf("expected ErrSignedSubmitUnavailable, got %v", err)
	}
}
