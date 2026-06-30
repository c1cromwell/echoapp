package wallet

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// fakeSubmitter is a CurrencySubmitter that only serves validator snapshots.
type fakeSubmitter struct {
	validators []metagraph.ValidatorSnapshot
}

func (f *fakeSubmitter) SubmitTokenLock(context.Context, metagraph.TokenLockUpdate) (string, error) {
	return "", nil
}
func (f *fakeSubmitter) SubmitWithdrawLock(context.Context, metagraph.WithdrawLockUpdate) (string, error) {
	return "", nil
}
func (f *fakeSubmitter) SubmitStakeDelegation(context.Context, metagraph.StakeDelegationUpdate) (string, error) {
	return "", nil
}
func (f *fakeSubmitter) SubmitCurrencyL1(context.Context, metagraph.CurrencyL1Transaction) (string, error) {
	return "", nil
}
func (f *fakeSubmitter) QueryValidators(context.Context) ([]metagraph.ValidatorSnapshot, error) {
	return f.validators, nil
}

func TestReconcileOnceRefreshesValidatorCache(t *testing.T) {
	store := NewMemStore()
	sub := &fakeSubmitter{validators: []metagraph.ValidatorSnapshot{
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
