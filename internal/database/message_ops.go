package database

import (
	"context"
	"errors"
	"time"
)

// MaxPinnedPerConversation caps pinned messages per conversation (WO-59).
const MaxPinnedPerConversation = 5

// ErrPinLimitReached is returned when a conversation already has the max pins.
var ErrPinLimitReached = errors.New("pin limit reached for conversation")

// MessageEdit is one immutable edit version of a message (WO-25). It is only
// persisted server-side when the conversation is under retention (Comply /
// litigation hold). Ciphertext is opaque — the server never reads content.
type MessageEdit struct {
	MessageID      string    `json:"messageId"`
	ConversationID string    `json:"conversationId"`
	EditorDID      string    `json:"editorDid"`
	Version        int       `json:"version"`
	Ciphertext     []byte    `json:"ciphertext"`
	EditedAt       time.Time `json:"editedAt"`
}

// PinnedMessage is a message pinned within a conversation (WO-59).
type PinnedMessage struct {
	ConversationID string    `json:"conversationId"`
	MessageID      string    `json:"messageId"`
	PinnerDID      string    `json:"pinnerDid"`
	PinnedAt       time.Time `json:"pinnedAt"`
}

// MessageOpsStore backs the M1 message operations: edit history (retention-gated),
// synchronized-delete tombstones, and pins. Per the hybrid model, edit history is
// only retained when the conversation is flagged for retention; otherwise edits are
// relayed events and clients hold the history.
type MessageOpsStore interface {
	// Retention gate — Comply / litigation hold sets this (default false).
	SetConversationRetention(ctx context.Context, conversationID string, retained bool) error
	IsConversationRetained(ctx context.Context, conversationID string) (bool, error)

	// Edit history (immutable; only persisted for retained conversations).
	AppendEditVersion(ctx context.Context, edit *MessageEdit) (int, error)
	GetEditHistory(ctx context.Context, messageID string) ([]*MessageEdit, error)

	// Synchronized delete (WO-84). retained=true keeps any edit history for eDiscovery.
	MarkMessageDeleted(ctx context.Context, messageID string, retained bool) error
	IsMessageDeleted(ctx context.Context, messageID string) (bool, error)

	// Pins (WO-59), max MaxPinnedPerConversation per conversation.
	PinMessage(ctx context.Context, conversationID, messageID, pinnerDID string) error
	UnpinMessage(ctx context.Context, conversationID, messageID string) error
	GetPinnedMessages(ctx context.Context, conversationID string) ([]*PinnedMessage, error)

	// Disappearing messages: per-conversation TTL in seconds (0 = off). Clients and
	// the relay expiry both enforce it; this is the synced source of the setting.
	SetDisappearingTTL(ctx context.Context, conversationID string, ttlSeconds int) error
	GetDisappearingTTL(ctx context.Context, conversationID string) (int, error)
}

// --- MemoryDB implementation ---

func (m *MemoryDB) SetConversationRetention(_ context.Context, conversationID string, retained bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if retained {
		m.retained[conversationID] = true
	} else {
		delete(m.retained, conversationID)
	}
	return nil
}

func (m *MemoryDB) IsConversationRetained(_ context.Context, conversationID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.retained[conversationID], nil
}

func (m *MemoryDB) AppendEditVersion(_ context.Context, edit *MessageEdit) (int, error) {
	if edit == nil || edit.MessageID == "" {
		return 0, errors.New("edit message id required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	version := len(m.editVersions[edit.MessageID]) + 1
	stored := *edit
	stored.Version = version
	if stored.EditedAt.IsZero() {
		stored.EditedAt = time.Now()
	}
	m.editVersions[edit.MessageID] = append(m.editVersions[edit.MessageID], &stored)
	return version, nil
}

func (m *MemoryDB) GetEditHistory(_ context.Context, messageID string) ([]*MessageEdit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	src := m.editVersions[messageID]
	out := make([]*MessageEdit, len(src))
	copy(out, src)
	return out, nil
}

func (m *MemoryDB) MarkMessageDeleted(_ context.Context, messageID string, retained bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deletedMsgs[messageID] = true
	// Under retention, edit history is preserved for eDiscovery; otherwise drop it.
	if !retained {
		delete(m.editVersions, messageID)
	}
	return nil
}

func (m *MemoryDB) IsMessageDeleted(_ context.Context, messageID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.deletedMsgs[messageID], nil
}

func (m *MemoryDB) PinMessage(_ context.Context, conversationID, messageID, pinnerDID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing := m.pins[conversationID]
	for _, p := range existing {
		if p.MessageID == messageID {
			return nil // already pinned — idempotent
		}
	}
	if len(existing) >= MaxPinnedPerConversation {
		return ErrPinLimitReached
	}
	m.pins[conversationID] = append(existing, &PinnedMessage{
		ConversationID: conversationID,
		MessageID:      messageID,
		PinnerDID:      pinnerDID,
		PinnedAt:       time.Now(),
	})
	return nil
}

func (m *MemoryDB) UnpinMessage(_ context.Context, conversationID, messageID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing := m.pins[conversationID]
	out := existing[:0]
	for _, p := range existing {
		if p.MessageID != messageID {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		delete(m.pins, conversationID)
	} else {
		m.pins[conversationID] = out
	}
	return nil
}

func (m *MemoryDB) GetPinnedMessages(_ context.Context, conversationID string) ([]*PinnedMessage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	src := m.pins[conversationID]
	out := make([]*PinnedMessage, len(src))
	copy(out, src)
	return out, nil
}

func (m *MemoryDB) SetDisappearingTTL(_ context.Context, conversationID string, ttlSeconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if ttlSeconds <= 0 {
		delete(m.disappearing, conversationID)
	} else {
		m.disappearing[conversationID] = ttlSeconds
	}
	return nil
}

func (m *MemoryDB) GetDisappearingTTL(_ context.Context, conversationID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.disappearing[conversationID], nil
}
