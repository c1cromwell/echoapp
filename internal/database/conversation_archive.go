package database

import "context"

// ConversationArchiveStore backs per-conversation archive flags (WO-198 / M6).
type ConversationArchiveStore interface {
	SetConversationArchived(ctx context.Context, conversationID string, archived bool) error
	IsConversationArchived(ctx context.Context, conversationID string) (bool, error)
}

func (m *MemoryDB) SetConversationArchived(_ context.Context, conversationID string, archived bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.archived == nil {
		m.archived = make(map[string]bool)
	}
	if archived {
		m.archived[conversationID] = true
	} else {
		delete(m.archived, conversationID)
	}
	return nil
}

func (m *MemoryDB) IsConversationArchived(_ context.Context, conversationID string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.archived == nil {
		return false, nil
	}
	return m.archived[conversationID], nil
}
