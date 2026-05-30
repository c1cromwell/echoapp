package recovery

import (
	"context"
	"sync"
	"time"
)

// MemStore is an in-memory recovery Store for tests.
type MemStore struct {
	mu            sync.Mutex
	policies      map[string]Policy
	shareholders  map[string][]Shareholder
	sessions      map[string]Session
}

func NewMemStore() *MemStore {
	return &MemStore{
		policies:     make(map[string]Policy),
		shareholders: make(map[string][]Shareholder),
		sessions:     make(map[string]Session),
	}
}

func (m *MemStore) UpsertPolicy(_ context.Context, policy Policy) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.policies[policy.HolderDID] = policy
	return nil
}

func (m *MemStore) GetPolicy(_ context.Context, holderDID string) (*Policy, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.policies[holderDID]
	if !ok {
		return nil, nil
	}
	cp := p
	return &cp, nil
}

func (m *MemStore) ReplaceShareholders(_ context.Context, holderDID string, shareholders []Shareholder) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]Shareholder, len(shareholders))
	copy(cp, shareholders)
	m.shareholders[holderDID] = cp
	return nil
}

func (m *MemStore) ListShareholders(_ context.Context, holderDID string) ([]Shareholder, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := m.shareholders[holderDID]
	cp := make([]Shareholder, len(out))
	copy(cp, out)
	return cp, nil
}

func (m *MemStore) InsertSession(_ context.Context, session Session) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[session.SessionID] = session
	return nil
}

func (m *MemStore) GetSession(_ context.Context, holderDID, sessionID string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok || s.HolderDID != holderDID {
		return nil, nil
	}
	cp := s
	return &cp, nil
}

func (m *MemStore) CompleteSession(_ context.Context, holderDID, sessionID, commitment string, completedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[sessionID]
	if !ok || s.HolderDID != holderDID {
		return ErrSessionNotFound
	}
	s.Status = SessionCompleted
	s.RootKeyCommitment = commitment
	s.CompletedAt = &completedAt
	m.sessions[sessionID] = s
	return nil
}
