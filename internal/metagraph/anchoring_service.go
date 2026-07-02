package metagraph

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/relay"
)

// CommitmentSource drains pending message commitments (typically the WS hub).
type CommitmentSource interface {
	FlushCommitments() []relay.CommitmentEntry
}

// ConfirmationPublisher pushes anchoring confirmations to connected clients.
type ConfirmationPublisher interface {
	PublishConfirmation(to string, payload AnchorConfirmation) bool
}

// AnchorConfirmation is the WebSocket confirmation payload (WO-15).
type AnchorConfirmation struct {
	Type           string   `json:"type"` // "confirmation"
	MessageID      string   `json:"messageId"`
	SnapshotHash   string   `json:"snapshotHash"`
	SnapshotHeight int64    `json:"snapshotHeight,omitempty"`
	MerkleProof    []string `json:"merkleProof,omitempty"`
	MerkleRoot     string   `json:"merkleRoot,omitempty"`
}

// AnchoringService batches commitments and anchors Merkle roots on Data L1 (WO-15).
type AnchoringService struct {
	dataL1      *MetagraphClient
	source      CommitmentSource
	confirm     ConfirmationPublisher
	proofs      ProofStore
	batcher     *AnchoringBatcher
	interval    time.Duration
	maxBatch    int
	lastFlush   time.Time
	mu          sync.Mutex
	submissions int
}

// AnchoringConfig wires the anchoring worker.
type AnchoringConfig struct {
	DataL1    *MetagraphClient
	Source    CommitmentSource
	Confirm   ConfirmationPublisher
	Proofs    ProofStore
	Interval  time.Duration
	LastFlush time.Time // optional; zero defaults to time.Now()
}

// NewAnchoringService creates a WO-15 anchoring worker.
func NewAnchoringService(cfg AnchoringConfig) *AnchoringService {
	interval := cfg.Interval
	if interval <= 0 {
		interval = BatchInterval
	}
	proofs := cfg.Proofs
	if proofs == nil {
		proofs = NewMemoryProofStore()
	}
	return &AnchoringService{
		dataL1:    cfg.DataL1,
		source:    cfg.Source,
		confirm:   cfg.Confirm,
		proofs:    proofs,
		batcher:   NewAnchoringBatcher(),
		interval:  interval,
		maxBatch:  MaxBatchSize,
		lastFlush: defaultLastFlush(cfg.LastFlush),
	}
}

func defaultLastFlush(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

// Run starts the periodic flush loop until ctx is cancelled.
func (s *AnchoringService) Run(ctx context.Context) {
	if s == nil || s.dataL1 == nil || s.source == nil {
		return
	}
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Tick(ctx)
		}
	}
}

// Tick ingests pending commitments and submits when due.
func (s *AnchoringService) Tick(ctx context.Context) {
	if s == nil {
		return
	}
	s.ingest(s.source.FlushCommitments())
	pending := s.batcher.PendingCount()
	s.mu.Lock()
	due := pending > 0 && (pending >= s.maxBatch || time.Since(s.lastFlush) >= s.interval)
	s.mu.Unlock()
	if !due {
		return
	}
	if err := s.submitPending(ctx); err != nil {
		log.Printf("anchoring: submit failed: %v", err)
	}
}

func (s *AnchoringService) ingest(entries []relay.CommitmentEntry) {
	if len(entries) == 0 {
		return
	}
	for _, e := range entries {
		if len(e.Hash) == 0 {
			continue
		}
		s.batcher.AddCommitment(e.MessageID, e.Hash)
		if e.SenderDID != "" && e.MessageID != "" {
			s.batcher.SetSender(e.MessageID, e.SenderDID)
		}
	}
}

func (s *AnchoringService) submitPending(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	batch := s.batcher.Flush()
	if batch == nil {
		s.lastFlush = time.Now()
		return nil
	}
	return s.submitBatch(ctx, batch)
}

func (s *AnchoringService) submitBatch(ctx context.Context, batch *AnchoringBatch) error {
	if batch == nil || batch.CommitmentCount == 0 {
		return nil
	}
	leafHashes := make([]string, len(batch.Commitments))
	for i, c := range batch.Commitments {
		leafHashes[i] = hex.EncodeToString(c.Hash)
	}
	tree := BuildMerkleTree(leafHashes)
	if tree.Root == "" {
		return nil
	}
	batch.MerkleRoot = tree.Root

	tx := DataL1MerkleRootUpdate{Root: tree.Root, LeafCount: batch.CommitmentCount}
	txID, err := s.dataL1.SubmitDataL1(ctx, tx)
	if err != nil {
		// Re-queue commitments for a later attempt.
		for _, c := range batch.Commitments {
			s.batcher.AddCommitment(c.MessageID, c.Hash)
		}
		return err
	}
	s.lastFlush = time.Now()
	s.submissions++

	for i, c := range batch.Commitments {
		if c.MessageID == "" {
			continue
		}
		proof := MessageAnchorProof{
			MessageID:      c.MessageID,
			Commitment:     leafHashes[i],
			Siblings:       tree.ProofSiblings(i),
			SnapshotHash:   txID,
			MerkleRoot:     tree.Root,
			SnapshotHeight: 0,
		}
		_ = s.proofs.Put(ctx, proof)
		if s.confirm != nil {
			sender := s.batcher.senderFor(c.MessageID)
			if sender != "" {
				s.confirm.PublishConfirmation(sender, AnchorConfirmation{
					Type:         "confirmation",
					MessageID:    c.MessageID,
					SnapshotHash: txID,
					MerkleProof:  proof.Siblings,
					MerkleRoot:   tree.Root,
				})
			}
		}
	}
	s.batcher.AckBatch(tree.Root)
	log.Printf("anchoring: submitted %d commitments root=%s tx=%s", batch.CommitmentCount, tree.Root, txID)
	return nil
}

// ProofStore exposes the configured proof index (for handlers/tests).
func (s *AnchoringService) ProofStore() ProofStore {
	if s == nil {
		return nil
	}
	return s.proofs
}

// Submissions returns the number of successful Data L1 batch submissions.
func (s *AnchoringService) Submissions() int {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.submissions
}

// hubConfirmationAdapter adapts the API hub to ConfirmationPublisher.
type hubConfirmationAdapter struct {
	publish func(to string, payload json.RawMessage) bool
}

func (a hubConfirmationAdapter) PublishConfirmation(to string, c AnchorConfirmation) bool {
	if a.publish == nil {
		return false
	}
	b, err := json.Marshal(c)
	if err != nil {
		return false
	}
	return a.publish(to, b)
}

// HubConfirmationPublisher builds a publisher that sends WS confirmation signals.
func HubConfirmationPublisher(publish func(to string, payload json.RawMessage) bool) ConfirmationPublisher {
	return hubConfirmationAdapter{publish: publish}
}
