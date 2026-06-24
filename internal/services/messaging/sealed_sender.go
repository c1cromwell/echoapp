package messaging

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var errInvalidSealedSender = errors.New("invalid sealed sender")

const sealedTokenTTL = 5 * time.Minute

// SealedTokenTTLSeconds exposes token lifetime for API responses.
func SealedTokenTTLSeconds() int { return int(sealedTokenTTL.Seconds()) }

// SealedTokenStore issues opaque delivery tokens for sealed-sender relay (WO-219).
// Tokens bind a registered sender DID server-side without exposing identity on the wire.
type SealedTokenStore struct {
	mu     sync.Mutex
	tokens map[string]sealedTokenEntry
}

type sealedTokenEntry struct {
	senderDID string
	expiresAt time.Time
}

// NewSealedTokenStore creates an in-memory token store.
func NewSealedTokenStore() *SealedTokenStore {
	return &SealedTokenStore{tokens: make(map[string]sealedTokenEntry)}
}

// Issue creates a single-use delivery token for senderDID.
func (s *SealedTokenStore) Issue(senderDID string) (string, error) {
	if senderDID == "" {
		return "", errInvalidSealedSender
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked(time.Now())
	s.tokens[token] = sealedTokenEntry{
		senderDID: senderDID,
		expiresAt: time.Now().Add(sealedTokenTTL),
	}
	return token, nil
}

// Consume validates and burns a delivery token for the claimed sender DID.
func (s *SealedTokenStore) Consume(token, senderDID string) bool {
	if token == "" || senderDID == "" {
		return false
	}
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.purgeLocked(now)
	entry, ok := s.tokens[token]
	if !ok || entry.senderDID != senderDID || now.After(entry.expiresAt) {
		return false
	}
	delete(s.tokens, token)
	return true
}

func (s *SealedTokenStore) purgeLocked(now time.Time) {
	for token, entry := range s.tokens {
		if now.After(entry.expiresAt) {
			delete(s.tokens, token)
		}
	}
}
