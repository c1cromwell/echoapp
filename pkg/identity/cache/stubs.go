package cache

import (
	"context"
	"sync"
	"time"
)

// IdentityCache represents a cache for identity data
type IdentityCache struct {
	mu    sync.RWMutex
	cache map[string]*CacheEntry
	ttl   time.Duration
}

// CacheEntry represents a single cache entry
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewIdentityCache creates a new identity cache with the specified TTL
func NewIdentityCache(ttl time.Duration) *IdentityCache {
	cache := &IdentityCache{
		cache: make(map[string]*CacheEntry),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Set stores a value in the cache
func (ic *IdentityCache) Set(ctx context.Context, key string, value interface{}) error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.cache[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ic.ttl),
	}

	return nil
}

// Get retrieves a value from the cache
func (ic *IdentityCache) Get(ctx context.Context, key string) (interface{}, error) {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	entry, exists := ic.cache[key]
	if !exists {
		return nil, nil
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, nil
	}

	return entry.Value, nil
}

// GetTrustLevel retrieves a trust level from cache
func (ic *IdentityCache) GetTrustLevel(userID string) (interface{}, bool) {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	key := "trust_" + userID
	entry, exists := ic.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// GetTrustLevelWithFallback retrieves a trust level with a fallback function
func (ic *IdentityCache) GetTrustLevelWithFallback(userID string, fallback func() (interface{}, error)) (interface{}, error) {
	if val, exists := ic.GetTrustLevel(userID); exists {
		return val, nil
	}

	val, err := fallback()
	if err != nil {
		return nil, err
	}

	key := "trust_" + userID
	ic.mu.Lock()
	ic.cache[key] = &CacheEntry{
		Value:     val,
		ExpiresAt: time.Now().Add(ic.ttl),
	}
	ic.mu.Unlock()

	return val, nil
}

// InvalidateTrustLevel invalidates a trust level cache entry
func (ic *IdentityCache) InvalidateTrustLevel(userID string) {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	key := "trust_" + userID
	delete(ic.cache, key)
}

// GetUserCredentials retrieves user credentials from cache
func (ic *IdentityCache) GetUserCredentials(userID string) (interface{}, bool) {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	key := "creds_" + userID
	entry, exists := ic.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// GetUserCredentialsWithFallback retrieves user credentials with fallback
func (ic *IdentityCache) GetUserCredentialsWithFallback(userID string, fallback func() (interface{}, error)) (interface{}, error) {
	if val, exists := ic.GetUserCredentials(userID); exists {
		return val, nil
	}

	val, err := fallback()
	if err != nil {
		return nil, err
	}

	key := "creds_" + userID
	ic.mu.Lock()
	ic.cache[key] = &CacheEntry{
		Value:     val,
		ExpiresAt: time.Now().Add(ic.ttl),
	}
	ic.mu.Unlock()

	return val, nil
}

// GetMetrics returns cache metrics
func (ic *IdentityCache) GetMetrics() map[string]interface{} {
	ic.mu.RLock()
	defer ic.mu.RUnlock()

	return map[string]interface{}{
		"total_entries": len(ic.cache),
		"ttl":           ic.ttl.String(),
	}
}

// Delete removes a value from the cache
func (ic *IdentityCache) Delete(ctx context.Context, key string) error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	delete(ic.cache, key)
	return nil
}

// Clear clears all entries from the cache
func (ic *IdentityCache) Clear() error {
	ic.mu.Lock()
	defer ic.mu.Unlock()

	ic.cache = make(map[string]*CacheEntry)
	return nil
}

// cleanupExpired periodically removes expired entries
func (ic *IdentityCache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ic.mu.Lock()
		now := time.Now()
		for key, entry := range ic.cache {
			if now.After(entry.ExpiresAt) {
				delete(ic.cache, key)
			}
		}
		ic.mu.Unlock()
	}
}
