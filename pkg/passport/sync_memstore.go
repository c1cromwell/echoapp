package passport

import (
	"context"
	"sync"
)

// MemSyncStore is an in-memory SyncStore for tests.
type MemSyncStore struct {
	mu   sync.Mutex
	blob map[string]SyncBlobRecord
}

func NewMemSyncStore() *MemSyncStore {
	return &MemSyncStore{blob: make(map[string]SyncBlobRecord)}
}

func (m *MemSyncStore) UpsertSyncBlob(_ context.Context, record SyncBlobRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blob[record.HolderDID] = record
	return nil
}

func (m *MemSyncStore) GetSyncBlob(_ context.Context, holderDID string) (*SyncBlobRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.blob[holderDID]
	if !ok {
		return nil, nil
	}
	cp := r
	return &cp, nil
}
