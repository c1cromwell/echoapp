package rewards

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// mockSubmitter records calls and optionally returns an error.
type mockSubmitter struct {
	mu          sync.Mutex
	submissions []metagraph.CurrencyL1Transaction
	err         error
}

func (m *mockSubmitter) SubmitCurrencyL1(_ context.Context, tx metagraph.CurrencyL1Transaction) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.submissions = append(m.submissions, tx)
	if m.err != nil {
		return "", m.err
	}
	return fmt.Sprintf("tx_%d", len(m.submissions)), nil
}

func (m *mockSubmitter) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.submissions)
}

func makeClaim(id string, amount int64) PendingClaim {
	return PendingClaim{
		ClaimID:    id,
		DID:        "did:dag:test_" + id,
		RewardType: "messaging",
		Amount:     amount,
		TrustTier:  3,
		Multiplier: 1.0,
		Timestamp:  time.Now(),
	}
}

func TestBatchProcessor_EnqueueAndFlush(t *testing.T) {
	sub := &mockSubmitter{}
	bp := NewBatchProcessor(sub)

	bp.Enqueue(makeClaim("c1", 100))
	bp.Enqueue(makeClaim("c2", 200))

	if bp.PendingCount() != 2 {
		t.Fatalf("expected 2 pending, got %d", bp.PendingCount())
	}

	bp.Flush()

	if bp.PendingCount() != 0 {
		t.Fatalf("expected 0 pending after flush, got %d", bp.PendingCount())
	}
	if sub.count() != 1 {
		t.Fatalf("expected 1 submission, got %d", sub.count())
	}

	results := bp.Results()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ClaimCount != 2 {
		t.Errorf("expected 2 claims in batch, got %d", results[0].ClaimCount)
	}
	if results[0].TotalECHO != 300 {
		t.Errorf("expected 300 total ECHO, got %d", results[0].TotalECHO)
	}
	if results[0].TxHash != "tx_1" {
		t.Errorf("expected tx_1, got %s", results[0].TxHash)
	}
	if results[0].Error != nil {
		t.Errorf("unexpected error: %v", results[0].Error)
	}
}

func TestBatchProcessor_FlushEmpty(t *testing.T) {
	sub := &mockSubmitter{}
	bp := NewBatchProcessor(sub)

	bp.Flush() // should be no-op

	if sub.count() != 0 {
		t.Fatalf("expected 0 submissions for empty flush, got %d", sub.count())
	}
	if len(bp.Results()) != 0 {
		t.Fatalf("expected 0 results, got %d", len(bp.Results()))
	}
}

func TestBatchProcessor_AutoFlushOnMaxSize(t *testing.T) {
	sub := &mockSubmitter{}
	bp := NewBatchProcessor(sub)
	bp.SetMaxBatchSize(3)

	bp.Enqueue(makeClaim("c1", 10))
	bp.Enqueue(makeClaim("c2", 20))
	// Third enqueue triggers auto-flush
	bp.Enqueue(makeClaim("c3", 30))

	// Give async flush a moment
	time.Sleep(50 * time.Millisecond)

	if sub.count() != 1 {
		t.Fatalf("expected auto-flush submission, got %d", sub.count())
	}
	if bp.PendingCount() != 0 {
		t.Errorf("expected 0 pending after auto-flush, got %d", bp.PendingCount())
	}
}

func TestBatchProcessor_SubmissionError(t *testing.T) {
	sub := &mockSubmitter{err: fmt.Errorf("network timeout")}
	bp := NewBatchProcessor(sub)

	bp.Enqueue(makeClaim("c1", 100))
	bp.Flush()

	results := bp.Results()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error == nil {
		t.Error("expected error in result")
	}
	if results[0].TxHash != "" {
		t.Errorf("expected empty txHash on error, got %s", results[0].TxHash)
	}
}

func TestBatchProcessor_StartStop(t *testing.T) {
	sub := &mockSubmitter{}
	bp := NewBatchProcessor(sub)
	bp.SetBatchInterval(50 * time.Millisecond)

	bp.Enqueue(makeClaim("c1", 100))

	bp.Start()
	time.Sleep(150 * time.Millisecond) // Wait for at least one tick
	bp.Stop()

	if sub.count() < 1 {
		t.Fatalf("expected at least 1 submission from periodic flush, got %d", sub.count())
	}
}

func TestBatchProcessor_AtomicActionStructure(t *testing.T) {
	sub := &mockSubmitter{}
	bp := NewBatchProcessor(sub)

	bp.Enqueue(makeClaim("c1", 100))
	bp.Flush()

	if sub.count() != 1 {
		t.Fatalf("expected 1 submission, got %d", sub.count())
	}

	tx := sub.submissions[0]
	if tx.Type != "atomic_action" {
		t.Errorf("expected atomic_action type, got %s", tx.Type)
	}
	if tx.AtomicAction == nil {
		t.Fatal("expected AtomicAction to be set")
	}
	// 1 claim → 1 reward_claim + 1 trust_verification + 1 daily_cap_update = 3 ops
	if len(tx.AtomicAction.Operations) != 3 {
		t.Errorf("expected 3 operations, got %d", len(tx.AtomicAction.Operations))
	}

	ops := tx.AtomicAction.Operations
	if ops[0].Type != metagraph.OpRewardClaim {
		t.Errorf("expected first op to be reward_claim, got %s", ops[0].Type)
	}
	if ops[1].Type != metagraph.OpTrustVerification {
		t.Errorf("expected second op to be trust_verification, got %s", ops[1].Type)
	}
	if ops[2].Type != metagraph.OpDailyCapUpdate {
		t.Errorf("expected third op to be daily_cap_update, got %s", ops[2].Type)
	}
}

func TestBatchProcessor_MultipleBatches(t *testing.T) {
	sub := &mockSubmitter{}
	bp := NewBatchProcessor(sub)

	bp.Enqueue(makeClaim("c1", 100))
	bp.Flush()

	bp.Enqueue(makeClaim("c2", 200))
	bp.Enqueue(makeClaim("c3", 300))
	bp.Flush()

	results := bp.Results()
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ClaimCount != 1 || results[1].ClaimCount != 2 {
		t.Errorf("unexpected claim counts: %d, %d", results[0].ClaimCount, results[1].ClaimCount)
	}
}
