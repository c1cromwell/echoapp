package wallet

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// ChallengeStore issues short-lived, single-use nonces bound to a DID so a
// proof-of-ownership signature cannot be replayed. In-memory (single instance);
// production multi-instance deployments should back this with Redis.
type ChallengeStore struct {
	mu      sync.Mutex
	entries map[string]challengeEntry
	ttl     time.Duration
	now     func() time.Time
}

type challengeEntry struct {
	challenge string
	expires   time.Time
}

// NewChallengeStore returns a store with a 5-minute challenge TTL.
func NewChallengeStore() *ChallengeStore {
	return &ChallengeStore{
		entries: make(map[string]challengeEntry),
		ttl:     5 * time.Minute,
		now:     time.Now,
	}
}

// Issue mints a fresh challenge for the DID (overwriting any prior one) and
// returns it with its expiry.
func (s *ChallengeStore) Issue(did string) (string, time.Time, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", time.Time{}, err
	}
	challenge := "echo-wallet:" + did + ":" + hex.EncodeToString(buf)
	exp := s.now().Add(s.ttl)
	s.mu.Lock()
	s.entries[did] = challengeEntry{challenge: challenge, expires: exp}
	s.mu.Unlock()
	return challenge, exp, nil
}

// Consume validates and burns a challenge for the DID. Returns false if it does
// not match the outstanding challenge or has expired.
func (s *ChallengeStore) Consume(did, challenge string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[did]
	if !ok || e.challenge != challenge || s.now().After(e.expires) {
		return false
	}
	delete(s.entries, did)
	return true
}
