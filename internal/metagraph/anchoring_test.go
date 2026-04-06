package metagraph

import (
	"testing"
)

func TestAnchoringBatcher_AddAndFlush(t *testing.T) {
	ab := NewAnchoringBatcher()

	ab.AddCommitment("msg-1", []byte("hash1"))
	ab.AddCommitment("msg-2", []byte("hash2"))
	ab.AddCommitment("msg-3", []byte("hash3"))

	if ab.PendingCount() != 3 {
		t.Errorf("expected 3 pending, got %d", ab.PendingCount())
	}

	batch := ab.Flush()
	if batch == nil {
		t.Fatal("batch should not be nil")
	}

	if batch.CommitmentCount != 3 {
		t.Errorf("expected 3 commitments in batch, got %d", batch.CommitmentCount)
	}
	if batch.MerkleRoot == "" {
		t.Error("merkle root should not be empty")
	}
	if batch.BatchHash == "" {
		t.Error("batch hash should not be empty")
	}
	if batch.TimeRange.From.IsZero() || batch.TimeRange.To.IsZero() {
		t.Error("time range should be set")
	}
	if ab.PendingCount() != 0 {
		t.Error("pending should be 0 after flush")
	}
}

func TestAnchoringBatcher_FlushEmpty(t *testing.T) {
	ab := NewAnchoringBatcher()
	batch := ab.Flush()
	if batch != nil {
		t.Error("flushing empty batcher should return nil")
	}
}

func TestAnchoringBatcher_AutoFlushAtMaxBatch(t *testing.T) {
	ab := NewAnchoringBatcher()

	// Add exactly MaxBatchSize commitments
	for i := 0; i < MaxBatchSize; i++ {
		ab.AddCommitment("msg-auto", []byte("hash"))
	}

	// Should have auto-flushed
	if ab.PendingCount() != 0 {
		t.Errorf("expected 0 pending after auto-flush, got %d", ab.PendingCount())
	}

	completed := ab.CompletedBatches()
	if len(completed) != 1 {
		t.Errorf("expected 1 completed batch, got %d", len(completed))
	}

	if completed[0].CommitmentCount != MaxBatchSize {
		t.Errorf("expected %d commitments, got %d", MaxBatchSize, completed[0].CommitmentCount)
	}
}

func TestAnchoringBatcher_CompletedBatches(t *testing.T) {
	ab := NewAnchoringBatcher()

	// Create two batches
	ab.AddCommitment("msg-1", []byte("h1"))
	ab.Flush()
	ab.AddCommitment("msg-2", []byte("h2"))
	ab.Flush()

	completed := ab.CompletedBatches()
	if len(completed) != 2 {
		t.Errorf("expected 2 completed batches, got %d", len(completed))
	}
}

func TestAnchoringBatcher_AckBatch(t *testing.T) {
	ab := NewAnchoringBatcher()

	ab.AddCommitment("msg-1", []byte("h1"))
	batch := ab.Flush()

	found := ab.AckBatch(batch.MerkleRoot)
	if !found {
		t.Error("should find and ack the batch")
	}

	completed := ab.CompletedBatches()
	if len(completed) != 0 {
		t.Errorf("expected 0 batches after ack, got %d", len(completed))
	}
}

func TestAnchoringBatcher_AckNonexistent(t *testing.T) {
	ab := NewAnchoringBatcher()
	found := ab.AckBatch("nonexistent-root")
	if found {
		t.Error("should not find nonexistent batch")
	}
}

func TestAnchoringBatcher_DeterministicRoots(t *testing.T) {
	ab1 := NewAnchoringBatcher()
	ab1.AddCommitment("msg-1", []byte("hash_a"))
	ab1.AddCommitment("msg-2", []byte("hash_b"))
	batch1 := ab1.Flush()

	ab2 := NewAnchoringBatcher()
	ab2.AddCommitment("msg-1", []byte("hash_a"))
	ab2.AddCommitment("msg-2", []byte("hash_b"))
	batch2 := ab2.Flush()

	if batch1.MerkleRoot != batch2.MerkleRoot {
		t.Error("same commitments should produce same merkle root")
	}

	// Different commitments should produce different root
	ab3 := NewAnchoringBatcher()
	ab3.AddCommitment("msg-1", []byte("hash_a"))
	ab3.AddCommitment("msg-2", []byte("hash_c"))
	batch3 := ab3.Flush()

	if batch1.MerkleRoot == batch3.MerkleRoot {
		t.Error("different commitments should produce different merkle root")
	}
}
