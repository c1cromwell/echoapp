// Package rewards provides batch processing for reward claims.
// The BatchProcessor collects pending claims and submits them to
// the metagraph Currency L1 as AtomicAction transactions.
package rewards

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

const (
	DefaultBatchInterval = 5 * time.Minute
	DefaultMaxBatchSize  = 100
)

// PendingClaim is a reward claim awaiting L1 submission.
type PendingClaim struct {
	ClaimID    string    `json:"claimId"`
	DID        string    `json:"did"`
	RewardType string    `json:"rewardType"`
	Amount     int64     `json:"amount"`
	TrustTier  int       `json:"trustTier"`
	Multiplier float64   `json:"multiplier"`
	Timestamp  time.Time `json:"timestamp"`
}

// BatchResult contains the outcome of a batch submission.
type BatchResult struct {
	BatchID     string         `json:"batchId"`
	TxHash      string         `json:"txHash"`
	ClaimCount  int            `json:"claimCount"`
	TotalECHO   int64          `json:"totalEcho"`
	Claims      []PendingClaim `json:"claims"`
	SubmittedAt time.Time      `json:"submittedAt"`
	Error       error          `json:"-"`
}

// MetagraphSubmitter abstracts Currency L1 submission for testability.
type MetagraphSubmitter interface {
	SubmitCurrencyL1(ctx context.Context, tx metagraph.CurrencyL1Transaction) (string, error)
}

// BatchProcessor collects pending reward claims and periodically submits
// them to the metagraph Currency L1 as AtomicAction transactions.
type BatchProcessor struct {
	submitter     MetagraphSubmitter
	batchInterval time.Duration
	maxBatchSize  int

	mu      sync.Mutex
	pending []PendingClaim
	results []BatchResult

	stopCh chan struct{}
	done   chan struct{}
}

// NewBatchProcessor creates a batch processor with default settings.
func NewBatchProcessor(submitter MetagraphSubmitter) *BatchProcessor {
	return &BatchProcessor{
		submitter:     submitter,
		batchInterval: DefaultBatchInterval,
		maxBatchSize:  DefaultMaxBatchSize,
		stopCh:        make(chan struct{}),
		done:          make(chan struct{}),
	}
}

// SetBatchInterval overrides the default flush interval.
func (bp *BatchProcessor) SetBatchInterval(d time.Duration) {
	bp.batchInterval = d
}

// SetMaxBatchSize overrides the default max batch size.
func (bp *BatchProcessor) SetMaxBatchSize(n int) {
	bp.maxBatchSize = n
}

// Enqueue adds a pending claim. If batch size is reached, an immediate flush is triggered.
func (bp *BatchProcessor) Enqueue(claim PendingClaim) {
	bp.mu.Lock()
	bp.pending = append(bp.pending, claim)
	shouldFlush := len(bp.pending) >= bp.maxBatchSize
	bp.mu.Unlock()

	if shouldFlush {
		bp.Flush()
	}
}

// PendingCount returns the number of claims awaiting submission.
func (bp *BatchProcessor) PendingCount() int {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return len(bp.pending)
}

// Results returns completed batch results.
func (bp *BatchProcessor) Results() []BatchResult {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	out := make([]BatchResult, len(bp.results))
	copy(out, bp.results)
	return out
}

// Start begins the periodic flush loop. Call Stop() to shut down.
func (bp *BatchProcessor) Start() {
	go func() {
		defer close(bp.done)
		ticker := time.NewTicker(bp.batchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				bp.Flush()
			case <-bp.stopCh:
				// Final flush on shutdown
				bp.Flush()
				return
			}
		}
	}()
}

// Stop signals the flush loop to terminate and waits for it to finish.
func (bp *BatchProcessor) Stop() {
	close(bp.stopCh)
	<-bp.done
}

// Flush drains pending claims and submits them as an AtomicAction.
func (bp *BatchProcessor) Flush() {
	bp.mu.Lock()
	if len(bp.pending) == 0 {
		bp.mu.Unlock()
		return
	}
	batch := bp.pending
	bp.pending = nil
	bp.mu.Unlock()

	result := bp.submitBatch(batch)

	bp.mu.Lock()
	bp.results = append(bp.results, result)
	bp.mu.Unlock()
}

// submitBatch builds an AtomicAction from claims and submits to Currency L1.
func (bp *BatchProcessor) submitBatch(claims []PendingClaim) BatchResult {
	batchID := uuid.New().String()

	var totalECHO int64
	for _, c := range claims {
		totalECHO += c.Amount
	}

	// Build AtomicAction operations
	ops := make([]metagraph.AtomicOperation, 0, len(claims)*2+1)

	for _, claim := range claims {
		// Reward claim operation
		claimPayload, _ := json.Marshal(map[string]interface{}{
			"claim_id":    claim.ClaimID,
			"did":         claim.DID,
			"reward_type": claim.RewardType,
			"amount":      claim.Amount,
		})
		ops = append(ops, metagraph.AtomicOperation{
			Type:    metagraph.OpRewardClaim,
			Layer:   metagraph.CurrencyL1,
			Payload: claimPayload,
		})

		// Trust verification for each claim
		trustPayload, _ := json.Marshal(map[string]interface{}{
			"did":        claim.DID,
			"trust_tier": claim.TrustTier,
			"multiplier": claim.Multiplier,
		})
		ops = append(ops, metagraph.AtomicOperation{
			Type:    metagraph.OpTrustVerification,
			Layer:   metagraph.DataL1,
			Payload: trustPayload,
		})
	}

	// Auto-scale rate update operation
	capPayload, _ := json.Marshal(map[string]interface{}{
		"batch_id":             batchID,
		"total_echo":           totalECHO,
		"claim_count":          len(claims),
		"timestamp":            time.Now().UTC(),
		"auto_scale_update_by": "batch_processor",
	})
	ops = append(ops, metagraph.AtomicOperation{
		Type:    metagraph.OpAutoScaleRateUpdate,
		Layer:   metagraph.CurrencyL1,
		Payload: capPayload,
	})

	// Build the AtomicAction
	txID := uuid.New().String()
	action, err := metagraph.NewAtomicAction(txID, "echo:treasury", metagraph.CurrencyL1, ops)
	if err != nil {
		log.Printf("[BatchProcessor] failed to build AtomicAction: %v", err)
		return BatchResult{
			BatchID:     batchID,
			ClaimCount:  len(claims),
			TotalECHO:   totalECHO,
			Claims:      claims,
			SubmittedAt: time.Now(),
			Error:       err,
		}
	}

	// Submit to Currency L1
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	txHash, err := bp.submitter.SubmitCurrencyL1(ctx, metagraph.CurrencyL1Transaction{
		Type:         "atomic_action",
		AtomicAction: action,
	})
	if err != nil {
		log.Printf("[BatchProcessor] L1 submission failed for batch %s: %v", batchID, err)
	} else {
		log.Printf("[BatchProcessor] batch %s submitted: txHash=%s claims=%d totalECHO=%d",
			batchID, txHash, len(claims), totalECHO)
	}

	return BatchResult{
		BatchID:     batchID,
		TxHash:      txHash,
		ClaimCount:  len(claims),
		TotalECHO:   totalECHO,
		Claims:      claims,
		SubmittedAt: time.Now(),
		Error:       err,
	}
}
