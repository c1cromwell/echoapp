package metagraph

import (
	"context"
	"math/big"
	"testing"
	"time"
)

// --- Mock FeeQuerier ---

type mockFeeQuerier struct {
	pendingFees int64
	dagBalance  int64
	txHash      string
	submitErr   error
	queryErr    error
	feesPaid    int64
}

func (m *mockFeeQuerier) QueryPendingFees() (int64, error) {
	if m.queryErr != nil {
		return 0, m.queryErr
	}
	return m.pendingFees, nil
}

func (m *mockFeeQuerier) QueryDAGBalance(_ string) (int64, error) {
	return m.dagBalance, nil
}

func (m *mockFeeQuerier) SubmitFeeTransaction(dagAmount int64) (string, error) {
	if m.submitErr != nil {
		return "", m.submitErr
	}
	m.feesPaid = dagAmount
	return "fee_tx_hash", nil
}

// --- Tests ---

func TestCheckAndPayFees_NoPending(t *testing.T) {
	mock := &mockFeeQuerier{pendingFees: 0, dagBalance: 1000}
	fm := NewFeeManager(mock)

	err := fm.CheckAndPayFees(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.feesPaid != 0 {
		t.Error("should not have submitted fee when nothing pending")
	}
}

func TestCheckAndPayFees_PaysFees(t *testing.T) {
	mock := &mockFeeQuerier{pendingFees: 100, dagBalance: 1000}
	fm := NewFeeManager(mock)

	err := fm.CheckAndPayFees(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.feesPaid != 100 {
		t.Errorf("expected 100 fees paid, got %d", mock.feesPaid)
	}
}

func TestCheckAndPayFees_LowBalanceAlert(t *testing.T) {
	alertCalled := false
	var alertBalance int64

	mock := &mockFeeQuerier{pendingFees: 600, dagBalance: 1000}
	fm := NewFeeManager(mock)
	fm.SetLowBalanceAlert(func(balance int64) {
		alertCalled = true
		alertBalance = balance
	})

	_ = fm.CheckAndPayFees(context.Background())

	if !alertCalled {
		t.Error("expected low balance alert when balance < 2x pending")
	}
	if alertBalance != 1000 {
		t.Errorf("expected alert balance 1000, got %d", alertBalance)
	}
}

func TestCheckAndPayFees_NoAlertWhenSufficient(t *testing.T) {
	alertCalled := false

	mock := &mockFeeQuerier{pendingFees: 100, dagBalance: 1000}
	fm := NewFeeManager(mock)
	fm.SetLowBalanceAlert(func(_ int64) {
		alertCalled = true
	})

	_ = fm.CheckAndPayFees(context.Background())

	if alertCalled {
		t.Error("should not alert when balance >= 2x pending")
	}
}

func TestFeeManagerRun_CancelContext(t *testing.T) {
	mock := &mockFeeQuerier{pendingFees: 0}
	fm := &FeeManager{
		client:        mock,
		checkInterval: 10 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		fm.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Run exited cleanly
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancellation")
	}
}

func TestSnapshotFeeEstimate(t *testing.T) {
	// 1KB data
	fee := SnapshotFeeEstimate(1024)
	expected := new(big.Int).Add(big.NewInt(100_000), big.NewInt(10_000))
	if fee.Cmp(expected) != 0 {
		t.Errorf("expected fee %s, got %s", expected, fee)
	}

	// 0 bytes
	feeZero := SnapshotFeeEstimate(0)
	if feeZero.Cmp(big.NewInt(100_000)) != 0 {
		t.Errorf("expected base fee 100_000, got %s", feeZero)
	}

	// 2500 bytes = 3KB
	fee2500 := SnapshotFeeEstimate(2500)
	expected2500 := new(big.Int).Add(big.NewInt(100_000), new(big.Int).Mul(big.NewInt(10_000), big.NewInt(3)))
	if fee2500.Cmp(expected2500) != 0 {
		t.Errorf("expected fee %s for 2500 bytes, got %s", expected2500, fee2500)
	}
}
