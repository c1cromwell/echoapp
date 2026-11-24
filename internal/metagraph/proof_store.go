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

// ChainAnchor is the chain's evidence for one anchored root: where it was recorded, and a proof
// that can be checked offline (besu.Proof as JSON). The gateway stores it so a proof bundle can be
// served, and verified, with the chain switched off.
type ChainAnchor struct {
	BlockNumber int64           `json:"blockNumber"` // block the root was recorded in
	TxHash      string          `json:"txHash"`
	Contract    string          `json:"contract"`
	Proof       json.RawMessage `json:"proof,omitempty"`
}

// AnchoredRoot is the durable record of one Merkle root submitted to the chain.
type AnchoredRoot struct {
	MerkleRoot  string
	Network     string
	LeafCount   int
	SubmittedAt time.Time
	Finalized   bool
	FinalizedAt *time.Time
	Chain       *ChainAnchor
}

// RootStore persists anchored roots and the chain's evidence for them.
type RootStore interface {
	// RecordRootSubmitted inserts the root if it is not already known.
	RecordRootSubmitted(ctx context.Context, root string, leafCount int, network string, at time.Time) error
	GetRoot(ctx context.Context, root string) (AnchoredRoot, bool, error)
	// ListIncompleteRoots returns roots that are not finalized or have no chain proof yet, newest first.
	ListIncompleteRoots(ctx context.Context, limit int) ([]AnchoredRoot, error)
	// MarkRootFinalized sets finalized; the first finalized_at wins.
	MarkRootFinalized(ctx context.Context, root string, at time.Time) error
	// SaveChainAnchor records the chain's evidence for the root.
	SaveChainAnchor(ctx context.Context, root string, anchor ChainAnchor) error
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
		if !r.Finalized || r.Chain == nil || len(r.Chain.Proof) == 0 {
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

// SaveChainAnchor implements RootStore.
func (s *MemoryProofStore) SaveChainAnchor(_ context.Context, root string, anchor ChainAnchor) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.roots[root]
	if !ok {
		return nil
	}
	c := anchor
	r.Chain = &c
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
