package leaderboard

import (
	"context"
	"errors"
	"sort"
	"sync"
)

// Store persists per-window, per-user accumulated leaderboard score. Implementations
// must be safe for concurrent use from multiple HTTP handlers.
//
// R0 ships the in-memory implementation below (the broadcast_channels service sets
// the same precedent). A Postgres-backed Store — mirroring
// internal/services/groups/postgres_store.go — is the persistence follow-up before
// leaderboard state must survive restarts.
type Store interface {
	// Add accumulates amountDatum onto (bucketKey, did) and records the latest tier.
	Add(ctx context.Context, bucketKey, did string, trustTier int, amountDatum int64) error
	// Top returns up to limit entries for bucketKey ranked by score desc, excluding
	// users below minTier, with Rank assigned (1-based).
	Top(ctx context.Context, bucketKey string, limit, minTier int) ([]Entry, error)
}

type aggregate struct {
	tier  int
	score int64
}

type memoryStore struct {
	mu      sync.RWMutex
	buckets map[string]map[string]*aggregate // bucketKey -> did -> aggregate
}

func newMemoryStore() *memoryStore {
	return &memoryStore{buckets: make(map[string]map[string]*aggregate)}
}

func (s *memoryStore) Add(_ context.Context, bucketKey, did string, trustTier int, amountDatum int64) error {
	if bucketKey == "" || did == "" {
		return errors.New("bucket key and did required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	bucket := s.buckets[bucketKey]
	if bucket == nil {
		bucket = make(map[string]*aggregate)
		s.buckets[bucketKey] = bucket
	}
	agg := bucket[did]
	if agg == nil {
		agg = &aggregate{}
		bucket[did] = agg
	}
	agg.tier = trustTier // latest tier wins
	agg.score += amountDatum
	return nil
}

func (s *memoryStore) Top(_ context.Context, bucketKey string, limit, minTier int) ([]Entry, error) {
	s.mu.RLock()
	bucket := s.buckets[bucketKey]
	entries := make([]Entry, 0, len(bucket))
	for did, agg := range bucket {
		if agg.tier < minTier {
			continue
		}
		entries = append(entries, Entry{DID: did, TrustTier: agg.tier, Score: agg.score})
	}
	s.mu.RUnlock()

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score != entries[j].Score {
			return entries[i].Score > entries[j].Score
		}
		return entries[i].DID < entries[j].DID // stable tiebreak
	})
	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries, nil
}
