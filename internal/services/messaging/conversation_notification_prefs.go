package messaging

import "sync"

// ConversationNotificationPrefsStore holds per-recipient mute state for push routing (WO-56).
// Device-local prefs are synced by the client; content-blind keys only.
type ConversationNotificationPrefsStore struct {
	mu    sync.RWMutex
	muted map[string]bool // key = recipientDID + "\x00" + conversationID
}

// NewConversationNotificationPrefsStore creates an in-memory prefs store.
func NewConversationNotificationPrefsStore() *ConversationNotificationPrefsStore {
	return &ConversationNotificationPrefsStore{muted: make(map[string]bool)}
}

func convNotifKey(recipientDID, conversationID string) string {
	return recipientDID + "\x00" + conversationID
}

// SetMuted records whether recipientDID has muted conversationID.
func (s *ConversationNotificationPrefsStore) SetMuted(recipientDID, conversationID string, muted bool) {
	if s == nil || recipientDID == "" || conversationID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := convNotifKey(recipientDID, conversationID)
	if muted {
		s.muted[key] = true
	} else {
		delete(s.muted, key)
	}
}

// IsMuted reports whether push alerts should be suppressed for the recipient.
func (s *ConversationNotificationPrefsStore) IsMuted(recipientDID, conversationID string) bool {
	if s == nil || recipientDID == "" || conversationID == "" {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.muted[convNotifKey(recipientDID, conversationID)]
}
