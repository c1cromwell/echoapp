package database

import (
	"context"
	"errors"
	"time"
)

// ErrDeviceRevoked is returned when a device's sync stream has been revoked.
var ErrDeviceRevoked = errors.New("device sync stream revoked")

// MaxSyncEntriesPerDevice caps a single device's pending sync stream to bound
// server storage. Entries are consumed by cursor; a client that never catches up
// will start losing the oldest history (acceptable — it can re-seed from backup).
const MaxSyncEntriesPerDevice = 20000

// SyncEntry is one opaque, content-blind entry in a device's sync stream (WO-CA3).
// The server never reads Ciphertext — it is wrapped to the target device's key by
// the writer (pairwise ECDH). Seq is monotonic per (ControllerDID, TargetDeviceID).
type SyncEntry struct {
	ControllerDID  string    `json:"controllerDid"`
	TargetDeviceID string    `json:"targetDeviceId"`
	Seq            int64     `json:"seq"`
	EntryType      string    `json:"entryType,omitempty"` // opaque hint: "message" | "history" | "tombstone"
	Ciphertext     []byte    `json:"ciphertext"`
	CreatedAt      time.Time `json:"createdAt"`
}

// DeviceSyncStore is the content-blind, per-device-addressed message-history sync
// log (WO-CA3). A device pulls only its own stream (scoped by controller DID);
// revoking a device closes and purges its stream — the "revoke stops sync" gate.
type DeviceSyncStore interface {
	// AppendSyncEntry appends an opaque entry to a target device's stream and
	// returns the assigned monotonic seq. Errors with ErrDeviceRevoked if the
	// target stream has been revoked.
	AppendSyncEntry(ctx context.Context, entry *SyncEntry) (int64, error)
	// PullSyncEntries returns entries with seq > afterSeq for the stream, ordered
	// by seq ascending, up to limit. ErrDeviceRevoked if revoked.
	PullSyncEntries(ctx context.Context, controllerDID, targetDeviceID string, afterSeq int64, limit int) ([]*SyncEntry, error)
	// SyncHead returns the latest assigned seq for the stream (0 if empty).
	SyncHead(ctx context.Context, controllerDID, targetDeviceID string) (int64, error)
	// RevokeDeviceStream closes a device's stream and purges its entries.
	RevokeDeviceStream(ctx context.Context, controllerDID, targetDeviceID string) error
	// AckSyncEntries deletes entries with seq <= throughSeq after the client has applied them.
	AckSyncEntries(ctx context.Context, controllerDID, targetDeviceID string, throughSeq int64) error
	// IsDeviceStreamActive reports whether the stream is active (not revoked).
	IsDeviceStreamActive(ctx context.Context, controllerDID, targetDeviceID string) (bool, error)
}

// --- MemoryDB implementation ---

func (m *MemoryDB) AppendSyncEntry(_ context.Context, entry *SyncEntry) (int64, error) {
	if entry == nil || entry.ControllerDID == "" || entry.TargetDeviceID == "" {
		return 0, errors.New("controller did and target device id required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.syncRevoked[entry.ControllerDID][entry.TargetDeviceID] {
		return 0, ErrDeviceRevoked
	}
	if m.syncSeq[entry.ControllerDID] == nil {
		m.syncSeq[entry.ControllerDID] = make(map[string]int64)
	}
	if m.syncStreams[entry.ControllerDID] == nil {
		m.syncStreams[entry.ControllerDID] = make(map[string][]*SyncEntry)
	}
	seq := m.syncSeq[entry.ControllerDID][entry.TargetDeviceID] + 1
	m.syncSeq[entry.ControllerDID][entry.TargetDeviceID] = seq

	stored := *entry
	stored.Seq = seq
	if stored.CreatedAt.IsZero() {
		stored.CreatedAt = time.Now()
	}
	stream := append(m.syncStreams[entry.ControllerDID][entry.TargetDeviceID], &stored)
	// Bound the pending stream; oldest unconsumed entries fall off.
	if len(stream) > MaxSyncEntriesPerDevice {
		stream = stream[len(stream)-MaxSyncEntriesPerDevice:]
	}
	m.syncStreams[entry.ControllerDID][entry.TargetDeviceID] = stream
	return seq, nil
}

func (m *MemoryDB) AckSyncEntries(_ context.Context, controllerDID, targetDeviceID string, throughSeq int64) error {
	if throughSeq <= 0 {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	stream := m.syncStreams[controllerDID][targetDeviceID]
	if len(stream) == 0 {
		return nil
	}
	kept := stream[:0]
	for _, e := range stream {
		if e.Seq > throughSeq {
			kept = append(kept, e)
		}
	}
	m.syncStreams[controllerDID][targetDeviceID] = kept
	return nil
}

func (m *MemoryDB) PullSyncEntries(_ context.Context, controllerDID, targetDeviceID string, afterSeq int64, limit int) ([]*SyncEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.syncRevoked[controllerDID][targetDeviceID] {
		return nil, ErrDeviceRevoked
	}
	if limit <= 0 {
		limit = 100
	}
	out := make([]*SyncEntry, 0, limit)
	for _, e := range m.syncStreams[controllerDID][targetDeviceID] {
		if e.Seq > afterSeq {
			cp := *e
			out = append(out, &cp)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (m *MemoryDB) SyncHead(_ context.Context, controllerDID, targetDeviceID string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.syncSeq[controllerDID][targetDeviceID], nil
}

func (m *MemoryDB) RevokeDeviceStream(_ context.Context, controllerDID, targetDeviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.syncRevoked[controllerDID] == nil {
		m.syncRevoked[controllerDID] = make(map[string]bool)
	}
	m.syncRevoked[controllerDID][targetDeviceID] = true
	if m.syncStreams[controllerDID] != nil {
		delete(m.syncStreams[controllerDID], targetDeviceID) // purge — stops sync
	}
	return nil
}

func (m *MemoryDB) IsDeviceStreamActive(_ context.Context, controllerDID, targetDeviceID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return !m.syncRevoked[controllerDID][targetDeviceID], nil
}
