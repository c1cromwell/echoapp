package metagraph

import (
	"context"
	"encoding/json"
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

// MemoryProofStore is an in-process proof index.
type MemoryProofStore struct {
	mu     sync.RWMutex
	proofs map[string]MessageAnchorProof
}

// NewMemoryProofStore creates an empty in-memory proof store.
func NewMemoryProofStore() *MemoryProofStore {
	return &MemoryProofStore{proofs: make(map[string]MessageAnchorProof)}
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
type RedisProofStore struct {
	redis *infra.RedisClient
	mem   *MemoryProofStore
	ttl   time.Duration
}

// NewRedisProofStore wraps Redis with an in-memory fallback.
func NewRedisProofStore(redis *infra.RedisClient, mem *MemoryProofStore) *RedisProofStore {
	return &RedisProofStore{redis: redis, mem: mem, ttl: 7 * 24 * time.Hour}
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
