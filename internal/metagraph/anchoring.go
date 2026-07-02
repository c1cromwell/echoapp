package metagraph

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

const (
	BatchInterval = 5 * time.Minute
	MaxBatchSize  = 1000
)

// Commitment represents a single message commitment hash.
type Commitment struct {
	MessageID string
	Hash      []byte
	Timestamp time.Time
}

// AnchoringBatch represents a completed batch ready for Data L1 submission.
type AnchoringBatch struct {
	MerkleRoot      string       `json:"merkle_root"`
	CommitmentCount int          `json:"commitment_count"`
	BatchHash       string       `json:"batch_hash"`
	TimeRange       TimeRange    `json:"time_range"`
	Commitments     []Commitment `json:"-"` // internal, not serialized
	CreatedAt       time.Time    `json:"created_at"`
}

// TimeRange represents the temporal bounds of a batch.
type TimeRange struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// AnchoringBatcher collects message commitment hashes and periodically
// builds a Merkle tree, producing batches ready for Data L1 submission.
type AnchoringBatcher struct {
	mu          sync.Mutex
	commitments []Commitment
	maxBatch    int
	batches     []AnchoringBatch // completed batches awaiting submission
	senders     map[string]string
}

// NewAnchoringBatcher creates a new batcher with default settings.
func NewAnchoringBatcher() *AnchoringBatcher {
	return &AnchoringBatcher{
		maxBatch: MaxBatchSize,
		senders:  make(map[string]string),
	}
}

// SetSender records the sender DID for a message ID (used for WS confirmations).
func (b *AnchoringBatcher) SetSender(messageID, senderDID string) {
	if messageID == "" || senderDID == "" {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.senders == nil {
		b.senders = make(map[string]string)
	}
	b.senders[messageID] = senderDID
}

func (b *AnchoringBatcher) senderFor(messageID string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.senders[messageID]
}

// AddCommitment adds a message commitment to the current batch.
// If the batch reaches MaxBatchSize, it is automatically flushed.
func (b *AnchoringBatcher) AddCommitment(messageID string, hash []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.commitments = append(b.commitments, Commitment{
		MessageID: messageID,
		Hash:      hash,
		Timestamp: time.Now(),
	})
}

// Flush manually triggers a batch flush. Returns the batch if one was created.
func (b *AnchoringBatcher) Flush() *AnchoringBatch {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.flushLocked()
}

// flushLocked builds a Merkle tree from current commitments. Must be called with lock held.
func (b *AnchoringBatcher) flushLocked() *AnchoringBatch {
	if len(b.commitments) == 0 {
		return nil
	}

	batch := b.commitments
	b.commitments = nil

	// Compute Merkle root (aligned with BuildMerkleTree for proof generation).
	leafHashes := make([]string, len(batch))
	for i, c := range batch {
		leafHashes[i] = hex.EncodeToString(c.Hash)
	}
	tree := BuildMerkleTree(leafHashes)
	root := tree.Root

	// Compute batch hash (hash of all commitment hashes concatenated)
	batchData := ""
	for _, h := range leafHashes {
		batchData += h
	}
	batchHash := sha256.Sum256([]byte(batchData))

	result := AnchoringBatch{
		MerkleRoot:      root,
		CommitmentCount: len(batch),
		BatchHash:       hex.EncodeToString(batchHash[:]),
		TimeRange: TimeRange{
			From: batch[0].Timestamp,
			To:   batch[len(batch)-1].Timestamp,
		},
		Commitments: batch,
		CreatedAt:   time.Now(),
	}

	b.batches = append(b.batches, result)
	return &result
}

// PendingCount returns the number of commitments awaiting flush.
func (b *AnchoringBatcher) PendingCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.commitments)
}

// CompletedBatches returns all flushed batches awaiting Data L1 submission.
func (b *AnchoringBatcher) CompletedBatches() []AnchoringBatch {
	b.mu.Lock()
	defer b.mu.Unlock()
	result := make([]AnchoringBatch, len(b.batches))
	copy(result, b.batches)
	return result
}

// AckBatch removes a submitted batch from the completed list.
func (b *AnchoringBatcher) AckBatch(merkleRoot string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i, batch := range b.batches {
		if batch.MerkleRoot == merkleRoot {
			b.batches = append(b.batches[:i], b.batches[i+1:]...)
			return true
		}
	}
	return false
}
