package bots

import (
	"sync"
	"time"
)

// InstallStore holds per-user bot installations in memory (Stage 4 MVP).
type InstallStore struct {
	mu    sync.RWMutex
	byKey map[string]Installation // key = userDID + "\x00" + botDID
}

// NewInstallStore creates an empty install store.
func NewInstallStore() *InstallStore {
	return &InstallStore{byKey: make(map[string]Installation)}
}

func installKey(userDID, botDID string) string {
	return userDID + "\x00" + botDID
}

// Install records or updates a bot installation for userDID.
func (s *InstallStore) Install(userDID string, botDID string, granted []Permission) Installation {
	if s == nil || userDID == "" || botDID == "" {
		return Installation{}
	}
	now := time.Now().UTC()
	inst := Installation{
		BotDID:             botDID,
		UserDID:            userDID,
		GrantedPermissions: append([]Permission(nil), granted...),
		InstalledAt:        now,
		Active:             true,
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byKey[installKey(userDID, botDID)] = inst
	return inst
}

// Uninstall removes a bot for userDID.
func (s *InstallStore) Uninstall(userDID, botDID string) bool {
	if s == nil || userDID == "" || botDID == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := installKey(userDID, botDID)
	if _, ok := s.byKey[key]; !ok {
		return false
	}
	delete(s.byKey, key)
	return true
}

// SetActive toggles whether a bot may run for userDID.
func (s *InstallStore) SetActive(userDID, botDID string, active bool) (Installation, bool) {
	if s == nil || userDID == "" || botDID == "" {
		return Installation{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	key := installKey(userDID, botDID)
	inst, ok := s.byKey[key]
	if !ok {
		return Installation{}, false
	}
	inst.Active = active
	s.byKey[key] = inst
	return inst, true
}

// ListInstalled returns all installations for userDID.
func (s *InstallStore) ListInstalled(userDID string) []Installation {
	if s == nil || userDID == "" {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Installation
	for _, inst := range s.byKey {
		if inst.UserDID == userDID {
			out = append(out, inst)
		}
	}
	return out
}

// Get returns a single installation.
func (s *InstallStore) Get(userDID, botDID string) (Installation, bool) {
	if s == nil || userDID == "" || botDID == "" {
		return Installation{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	inst, ok := s.byKey[installKey(userDID, botDID)]
	return inst, ok
}
