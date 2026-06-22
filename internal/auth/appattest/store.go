package appattest

import (
	"crypto/ecdsa"
	"sync"
)

// KeyRecord is a stored attested device key and its last verified sign count.
type KeyRecord struct {
	PublicKey *ecdsa.PublicKey
	SignCount uint32
}

// KeyStore persists attested keys by keyID and tracks the monotonic sign count.
// Production backs this with the database; MemoryKeyStore is for tests/dev.
type KeyStore interface {
	Put(keyID string, rec KeyRecord) error
	Get(keyID string) (KeyRecord, bool, error)
	UpdateSignCount(keyID string, signCount uint32) error
}

// MemoryKeyStore is an in-memory KeyStore.
type MemoryKeyStore struct {
	mu   sync.RWMutex
	recs map[string]KeyRecord
}

func NewMemoryKeyStore() *MemoryKeyStore {
	return &MemoryKeyStore{recs: make(map[string]KeyRecord)}
}

func (m *MemoryKeyStore) Put(keyID string, rec KeyRecord) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recs[keyID] = rec
	return nil
}

func (m *MemoryKeyStore) Get(keyID string) (KeyRecord, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	rec, ok := m.recs[keyID]
	return rec, ok, nil
}

func (m *MemoryKeyStore) UpdateSignCount(keyID string, signCount uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rec, ok := m.recs[keyID]
	if !ok {
		return nil
	}
	rec.SignCount = signCount
	m.recs[keyID] = rec
	return nil
}
