package metagraph

import (
	"context"
	"log"
	"math/big"
	"time"
)

// TreasuryAddress is the metagraph treasury DAG address for fee payments.
const TreasuryAddress = "echo_treasury"

// FeeQuerier abstracts metagraph fee and balance queries.
type FeeQuerier interface {
	QueryPendingFees() (int64, error)
	QueryDAGBalance(address string) (int64, error)
	SubmitFeeTransaction(dagAmount int64) (string, error)
}

// FeeManager automatically pays snapshot fees from treasury DAG reserves.
type FeeManager struct {
	client          FeeQuerier
	checkInterval   time.Duration
	lowBalanceAlert func(balance int64)
}

// NewFeeManager creates a FeeManager that checks fees at the given interval.
func NewFeeManager(client FeeQuerier) *FeeManager {
	return &FeeManager{
		client:        client,
		checkInterval: 1 * time.Hour,
	}
}

// SetLowBalanceAlert sets a callback for when treasury balance drops below 2x pending fees.
func (f *FeeManager) SetLowBalanceAlert(fn func(balance int64)) {
	f.lowBalanceAlert = fn
}

// Run starts the automated fee payment loop. Blocks until ctx is cancelled.
func (f *FeeManager) Run(ctx context.Context) {
	ticker := time.NewTicker(f.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := f.CheckAndPayFees(ctx); err != nil {
				log.Printf("Fee payment error: %v", err)
			}
		}
	}
}

// CheckAndPayFees queries pending fees and submits a FeeTransaction if needed.
func (f *FeeManager) CheckAndPayFees(_ context.Context) error {
	pending, err := f.client.QueryPendingFees()
	if err != nil {
		return err
	}

	if pending <= 0 {
		return nil
	}

	balance, err := f.client.QueryDAGBalance(TreasuryAddress)
	if err != nil {
		return err
	}

	if balance < pending*2 {
		if f.lowBalanceAlert != nil {
			f.lowBalanceAlert(balance)
		}
	}

	_, err = f.client.SubmitFeeTransaction(pending)
	return err
}

// SnapshotFeeEstimate estimates the DAG fee for a snapshot given data size.
func SnapshotFeeEstimate(dataSizeBytes int) *big.Int {
	// Base fee + per-KB cost
	baseFee := big.NewInt(100_000) // 0.001 DAG
	perKB := big.NewInt(10_000)    // 0.0001 DAG per KB
	kb := int64(dataSizeBytes / 1024)
	if dataSizeBytes%1024 > 0 {
		kb++
	}
	total := new(big.Int).Add(baseFee, new(big.Int).Mul(perKB, big.NewInt(kb)))
	return total
}
