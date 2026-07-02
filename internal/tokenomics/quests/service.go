package quests

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrQuestNotFound    = errors.New("quest not found")
	ErrAlreadyClaimed   = errors.New("quest reward already claimed")
	ErrQuestIncomplete  = errors.New("quest not completed")
	ErrTrustTierTooLow  = errors.New("trust tier requirement not met")
	ErrClaimNotEligible = errors.New("not eligible to claim quest reward")
)

// Store persists quest completion state.
type Store interface {
	GetCompletion(ctx context.Context, did, questID string) (*Completion, error)
	SetCompletion(ctx context.Context, c Completion) error
	ListCompletions(ctx context.Context, did string) ([]Completion, error)
}

// MemStore is an in-memory quest completion store for dev/test.
type MemStore struct {
	mu   sync.RWMutex
	data map[string]Completion
}

func NewMemStore() *MemStore {
	return &MemStore{data: make(map[string]Completion)}
}

func key(did, questID string) string { return did + "|" + questID }

func (s *MemStore) GetCompletion(_ context.Context, did, questID string) (*Completion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.data[key(did, questID)]
	if !ok {
		return nil, nil
	}
	cp := c
	return &cp, nil
}

func (s *MemStore) SetCompletion(_ context.Context, c Completion) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key(c.DID, c.QuestID)] = c
	return nil
}

func (s *MemStore) ListCompletions(_ context.Context, did string) ([]Completion, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Completion, 0)
	for k, c := range s.data {
		if len(k) > len(did) && k[:len(did)] == did {
			out = append(out, c)
		}
	}
	return out, nil
}

// Service manages quest catalog and claims (WO-271).
type Service struct {
	store       Store
	trustTier   func(ctx context.Context, did string) (int, error)
	submitClaim func(ctx context.Context, did string, q Definition) (string, error)
}

// NewService creates a quest service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// WithTrustTier wires trust tier lookup for gating.
func (s *Service) WithTrustTier(fn func(ctx context.Context, did string) (int, error)) *Service {
	s.trustTier = fn
	return s
}

// WithClaimSubmitter wires AtomicAction claim submission.
func (s *Service) WithClaimSubmitter(fn func(ctx context.Context, did string, q Definition) (string, error)) *Service {
	s.submitClaim = fn
	return s
}

// ListCatalog returns quests with per-user completion status.
func (s *Service) ListCatalog(ctx context.Context, did string) ([]CatalogEntry, error) {
	completions, _ := s.store.ListCompletions(ctx, did)
	byQuest := make(map[string]Completion, len(completions))
	for _, c := range completions {
		byQuest[c.QuestID] = c
	}

	out := make([]CatalogEntry, 0, len(All()))
	for _, def := range All() {
		entry := CatalogEntry{Definition: def}
		if c, ok := byQuest[def.ID]; ok {
			entry.CompletedAt = c.CompletedAt
			entry.RewardClaimed = c.RewardClaimed
			if c.CompletedAt != "" {
				entry.Progress = def.RequiredCount
			}
		}
		out = append(out, entry)
	}
	return out, nil
}

// MarkComplete records quest completion (event hook).
func (s *Service) MarkComplete(ctx context.Context, did, questID string) error {
	if _, ok := ByID(questID); !ok {
		return ErrQuestNotFound
	}
	existing, _ := s.store.GetCompletion(ctx, did, questID)
	if existing != nil && existing.CompletedAt != "" {
		return nil
	}
	return s.store.SetCompletion(ctx, Completion{
		DID:         did,
		QuestID:     questID,
		CompletedAt: time.Now().UTC().Format(time.RFC3339),
	})
}

// Claim submits quest reward via AtomicAction; idempotent on claimed.
func (s *Service) Claim(ctx context.Context, did, questID string) (string, error) {
	def, ok := ByID(questID)
	if !ok {
		return "", ErrQuestNotFound
	}
	comp, _ := s.store.GetCompletion(ctx, did, questID)
	if comp != nil && comp.RewardClaimed {
		return "", ErrAlreadyClaimed
	}
	if comp == nil || comp.CompletedAt == "" {
		return "", ErrQuestIncomplete
	}
	if def.MinTrustTier > 0 && s.trustTier != nil {
		tier, err := s.trustTier(ctx, did)
		if err != nil {
			return "", err
		}
		if tier < def.MinTrustTier {
			return "", ErrTrustTierTooLow
		}
	}

	txHash := "queued_pre_genesis"
	if s.submitClaim != nil {
		var err error
		txHash, err = s.submitClaim(ctx, did, def)
		if err != nil {
			return "", err
		}
	}

	comp.RewardClaimed = true
	comp.TxHash = txHash
	if err := s.store.SetCompletion(ctx, *comp); err != nil {
		return "", err
	}
	return txHash, nil
}
