package wallet

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/ledger"
)

// fakeSubmitter is a CurrencySubmitter for tests: serves validator snapshots and
// records client-signed forwards.
type fakeSubmitter struct {
	validators     []ledger.ValidatorSnapshot
	signedHash     string
	lastSignedType string
	lastSigned     []byte
}

func (f *fakeSubmitter) SubmitTokenLock(context.Context, ledger.TokenLockUpdate) (string, error) {
	return "", nil
}
func (f *fakeSubmitter) SubmitWithdrawLock(context.Context, ledger.WithdrawLockUpdate) (string, error) {
	return "", nil
}
func (f *fakeSubmitter) SubmitStakeDelegation(context.Context, ledger.StakeDelegationUpdate) (string, error) {
	return "", nil
}
func (f *fakeSubmitter) SubmitCurrencyL1(context.Context, ledger.CurrencyL1Transaction) (string, error) {
	return "", nil
}
func (f *fakeSubmitter) QueryValidators(context.Context) ([]ledger.ValidatorSnapshot, error) {
	return f.validators, nil
}
func (f *fakeSubmitter) SubmitSignedByType(_ context.Context, txType string, signed []byte) (string, error) {
	f.lastSignedType = txType
	f.lastSigned = signed
	return f.signedHash, nil
}

func TestReconcileOnceRefreshesValidatorCache(t *testing.T) {
	store := NewMemStore()
	sub := &fakeSubmitter{validators: []ledger.ValidatorSnapshot{
		{ID: "v1", Address: "DAGv1", Layer: "currency_l1"},
		{ID: "v2", Address: "DAGv2", Layer: "currency_l1"},
	}}
	q := NewLedgerQuerier(store, sub)
	r := NewReconciler(q)

	r.reconcileOnce(context.Background())

	cached, err := store.ListValidators(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cached) != 2 {
		t.Fatalf("expected 2 validators cached, got %d", len(cached))
	}
}

// Run must no-op (and return) when there is no Currency L1 source to reconcile.
func TestReconcilerNoOpWithoutSubmitter(t *testing.T) {
	q := NewLedgerQuerier(NewMemStore(), nil)
	// Run returns immediately rather than blocking when submitter is nil.
	NewReconciler(q).Run(context.Background())
}
