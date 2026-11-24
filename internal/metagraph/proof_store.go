package metagraph

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/infra"
)

// MessageAnchorProof is returned by GET /v1/messages/{id}/merkle-proof (WO-15).
type MessageAnchorProof struct {
	MessageID       string   `json:"messageId"`
	Commitment      string   `json:"commitment"`
	Siblings        []string `json:"siblings"`
	MerkleLeafIndex int      `json:"merkleLeafIndex,omitempty"`
	SnapshotHash    string   `json:"snapshotHash"`
	SnapshotHeight  int64    `json:"snapshotHeight,omitempty"`
	MerkleRoot      string   `json:"merkleRoot"`
}

// ProofStore persists message→proof mappings for later client verification.
type ProofStore interface {
	Put(ctx context.Context, proof MessageAnchorProof) error
	Get(ctx context.Context, messageID string) (MessageAnchorProof, bool, error)
}

// SnapshotRef describes one signed snapshot captured as finality evidence.
// Ordinal and Hash are as reported by the node; SHA256 is the digest of the
// archived JSON body (not the Tessellation snapshot hash). ArchiveKey/SHA256
// are empty when no object storage was configured.
type SnapshotRef struct {
	Ordinal    int64    `json:"ordinal"`
	Hash       string   `json:"hash"`
	Signers    []string `json:"signers"`
	ArchiveKey string   `json:"archiveKey,omitempty"`
	SHA256     string   `json:"sha256,omitempty"`
	// ContainsRoot is set for Metagraph L0 snapshots only: true when the root
	// was found in the snapshot's data-application on-chain state, false when
	// the snapshot only bounds finality ("at or before this ordinal").
	ContainsRoot *bool `json:"containsRoot,omitempty"`
}

// AnchoredRoot is the durable record of one Merkle root submitted to Data L1.
type AnchoredRoot struct {
	MerkleRoot  string
	Network     string
	LeafCount   int
	SubmittedAt time.Time
	Finalized   bool
	FinalizedAt *time.Time
	ML0         *SnapshotRef
	GL0         *SnapshotRef
}

// RootStore persists anchored roots and what was observed about their finality.
type RootStore interface {
	// RecordRootSubmitted inserts the root if it is not already known.
	RecordRootSubmitted(ctx context.Context, root string, leafCount int, network string, at time.Time) error
	GetRoot(ctx context.Context, root string) (AnchoredRoot, bool, error)
	// ListIncompleteRoots returns roots that are not finalized or have no
	// metagraph snapshot recorded yet, newest first.
	ListIncompleteRoots(ctx context.Context, limit int) ([]AnchoredRoot, error)
	// MarkRootFinalized sets finalized; the first finalized_at wins.
	MarkRootFinalized(ctx context.Context, root string, at time.Time) error
	// SaveRootSnapshots records snapshot evidence; a nil ref leaves that side untouched.
	SaveRootSnapshots(ctx context.Context, root string, ml0, gl0 *SnapshotRef) error
}

// AnchorStore is a ProofStore that also tracks anchored roots.
type AnchorStore interface {
	ProofStore
	RootStore
}

// MemoryProofStore is an in-process proof index.
type MemoryProofStore struct {
	mu     sync.RWMutex
	proofs map[string]MessageAnchorProof
	roots  map[string]AnchoredRoot
}

// NewMemoryProofStore creates an empty in-memory proof store.
func NewMemoryProofStore() *MemoryProofStore {
	return &MemoryProofStore{
		proofs: make(map[string]MessageAnchorProof),
		roots:  make(map[string]AnchoredRoot),
	}
}

// RecordRootSubmitted implements RootStore.
func (s *MemoryProofStore) RecordRootSubmitted(_ context.Context, root string, leafCount int, network string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.roots[root]; !ok {
		s.roots[root] = AnchoredRoot{MerkleRoot: root, Network: network, LeafCount: leafCount, SubmittedAt: at}
	}
	return nil
}

// GetRoot implements RootStore.
func (s *MemoryProofStore) GetRoot(_ context.Context, root string) (AnchoredRoot, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.roots[root]
	return r, ok, nil
}

// ListIncompleteRoots implements RootStore.
func (s *MemoryProofStore) ListIncompleteRoots(_ context.Context, limit int) ([]AnchoredRoot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []AnchoredRoot
	for _, r := range s.roots {
		if !r.Finalized || r.ML0 == nil {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SubmittedAt.After(out[j].SubmittedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// MarkRootFinalized implements RootStore.
func (s *MemoryProofStore) MarkRootFinalized(_ context.Context, root string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.roots[root]
	if !ok {
		return nil
	}
	r.Finalized = true
	if r.FinalizedAt == nil {
		r.FinalizedAt = &at
	}
	s.roots[root] = r
	return nil
}

// SaveRootSnapshots implements RootStore.
func (s *MemoryProofStore) SaveRootSnapshots(_ context.Context, root string, ml0, gl0 *SnapshotRef) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.roots[root]
	if !ok {
		return nil
	}
	if ml0 != nil {
		c := *ml0
		r.ML0 = &c
	}
	if gl0 != nil {
		c := *gl0
		r.GL0 = &c
	}
	s.roots[root] = r
	return nil
}

// Put stores a proof keyed by message ID.
func (s *MemoryProofStore) Put(_ context.Context, proof MessageAnchorProof) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proofs[proof.MessageID] = proof
	return nil
}

// Get loads a proof by message ID.
func (s *MemoryProofStore) Get(_ context.Context, messageID string) (MessageAnchorProof, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.proofs[messageID]
	return p, ok, nil
}

// RedisProofStore caches proofs in Redis when available (7-day TTL).
// It is the non-durable fallback used when no database is configured: anchored
// roots (RootStore) are tracked by the embedded memory store only.
type RedisProofStore struct {
	redis *infra.RedisClient
	*MemoryProofStore
	mem *MemoryProofStore
	ttl time.Duration
}

// NewRedisProofStore wraps Redis with an in-memory fallback.
func NewRedisProofStore(redis *infra.RedisClient, mem *MemoryProofStore) *RedisProofStore {
	if mem == nil {
		mem = NewMemoryProofStore()
	}
	return &RedisProofStore{redis: redis, MemoryProofStore: mem, mem: mem, ttl: 7 * 24 * time.Hour}
}

// Put writes to memory and Redis.
func (s *RedisProofStore) Put(ctx context.Context, proof MessageAnchorProof) error {
	if err := s.mem.Put(ctx, proof); err != nil {
		return err
	}
	if s.redis == nil {
		return nil
	}
	b, err := json.Marshal(proof)
	if err != nil {
		return err
	}
	return s.redis.CacheSet(ctx, "anchor:"+proof.MessageID, b, s.ttl)
}

// Get reads Redis first, then memory.
func (s *RedisProofStore) Get(ctx context.Context, messageID string) (MessageAnchorProof, bool, error) {
	if s.redis != nil {
		b, err := s.redis.CacheGet(ctx, "anchor:"+messageID)
		if err != nil {
			return MessageAnchorProof{}, false, err
		}
		if len(b) > 0 {
			var p MessageAnchorProof
			if err := json.Unmarshal(b, &p); err == nil {
				return p, true, nil
			}
		}
	}
	return s.mem.Get(ctx, messageID)
}
