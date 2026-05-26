package messaging

import "sync"

// ReactionStore holds emoji reactions on messages. Each (message, reactor) pair
// has at most one emoji — re-reacting replaces it. This is the durable truth for
// reactions; the live WS "reaction" signal is just a latency optimization that
// clients reconcile against this store.
//
// NOTE: in-memory/single-instance for now — a durable/shared store is a follow-up
// (tracked alongside the other Phase 2/3 in-memory stores).
type ReactionStore struct {
	mu sync.RWMutex
	// messageID -> reactorDID -> emoji
	byMessage map[string]map[string]string
}

// NewReactionStore creates an empty reaction store.
func NewReactionStore() *ReactionStore {
	return &ReactionStore{byMessage: make(map[string]map[string]string)}
}

// ReactionCount is an aggregated reaction for a message.
type ReactionCount struct {
	Emoji    string   `json:"emoji"`
	Count    int      `json:"count"`
	Reactors []string `json:"reactors"` // DIDs that reacted with this emoji
}

// Add sets (or replaces) a reactor's emoji on a message.
func (s *ReactionStore) Add(messageID, reactorDID, emoji string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.byMessage[messageID] == nil {
		s.byMessage[messageID] = make(map[string]string)
	}
	s.byMessage[messageID][reactorDID] = emoji
}

// Remove clears a reactor's reaction on a message (no-op if absent).
func (s *ReactionStore) Remove(messageID, reactorDID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m := s.byMessage[messageID]; m != nil {
		delete(m, reactorDID)
		if len(m) == 0 {
			delete(s.byMessage, messageID)
		}
	}
}

// List returns reactions on a message aggregated by emoji, with stable ordering
// (by descending count, then emoji) so responses are deterministic.
func (s *ReactionStore) List(messageID string) []ReactionCount {
	s.mu.RLock()
	defer s.mu.RUnlock()

	byEmoji := make(map[string][]string)
	for did, emoji := range s.byMessage[messageID] {
		byEmoji[emoji] = append(byEmoji[emoji], did)
	}

	out := make([]ReactionCount, 0, len(byEmoji))
	for emoji, reactors := range byEmoji {
		sortStrings(reactors)
		out = append(out, ReactionCount{Emoji: emoji, Count: len(reactors), Reactors: reactors})
	}
	sortReactions(out)
	return out
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

func sortReactions(r []ReactionCount) {
	for i := 1; i < len(r); i++ {
		for j := i; j > 0 && less(r[j], r[j-1]); j-- {
			r[j-1], r[j] = r[j], r[j-1]
		}
	}
}

func less(a, b ReactionCount) bool {
	if a.Count != b.Count {
		return a.Count > b.Count // higher count first
	}
	return a.Emoji < b.Emoji
}
