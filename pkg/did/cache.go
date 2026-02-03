package did

import (
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached entry with metadata
type CacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
	CachedAt  time.Time
	Valid     bool
	HitCount  int
}

// Cache provides thread-safe caching with TTL support
type Cache struct {
	mu       sync.RWMutex
	entries  map[string]*CacheEntry
	config   *CacheConfig
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewCache creates a new cache instance
func NewCache(config *CacheConfig) *Cache {
	cache := &Cache{
		entries:  make(map[string]*CacheEntry),
		config:   config,
		stopChan: make(chan struct{}),
	}

	// Start cleanup goroutine if enabled
	if config.Enabled {
		cache.wg.Add(1)
		go cache.cleanupRoutine()
	}

	return cache
}

// Set stores a value in the cache with TTL
func (c *Cache) Set(key string, value interface{}) error {
	if !c.config.Enabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache size limit
	if len(c.entries) >= c.config.MaxSize && c.entries[key] == nil {
		return NewDIDError(ErrCodeCacheError, "Cache is full", nil)
	}

	now := time.Now()
	c.entries[key] = &CacheEntry{
		Data:      value,
		ExpiresAt: now.Add(c.config.TTL),
		CachedAt:  now,
		Valid:     true,
		HitCount:  0,
	}

	return nil
}

// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, bool, error) {
	if !c.config.Enabled {
		return nil, false, nil
	}

	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false, nil
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		c.Delete(key)
		return nil, false, nil
	}

	// Update hit count
	c.mu.Lock()
	entry.HitCount++
	c.mu.Unlock()

	return entry.Data, true, nil
}

// Delete removes an entry from the cache
func (c *Cache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
	return nil
}

// Invalidate marks all entries as invalid (used when DID documents are updated)
func (c *Cache) Invalidate(pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If pattern is empty, invalidate all entries
	if pattern == "" {
		c.entries = make(map[string]*CacheEntry)
		return nil
	}

	// Invalidate entries matching the pattern
	for key, entry := range c.entries {
		if len(key) >= len(pattern) && key[:len(pattern)] == pattern {
			entry.Valid = false
		}
	}

	return nil
}

// Clear clears all cache entries
func (c *Cache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	return nil
}

// GetStats returns cache statistics
func (c *Cache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := map[string]interface{}{
		"enabled":       c.config.Enabled,
		"total_entries": len(c.entries),
		"ttl":           c.config.TTL.String(),
		"max_size":      c.config.MaxSize,
	}

	// Count valid and expired entries
	validCount := 0
	expiredCount := 0
	totalHits := 0
	now := time.Now()

	for _, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			expiredCount++
		} else if entry.Valid {
			validCount++
		}
		totalHits += entry.HitCount
	}

	stats["valid_entries"] = validCount
	stats["expired_entries"] = expiredCount
	stats["total_hits"] = totalHits

	return stats
}

// CacheWithExpiry caches a DID document with custom expiry
func (c *Cache) SetWithExpiry(key string, value interface{}, ttl time.Duration) error {
	if !c.config.Enabled {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache size limit
	if len(c.entries) >= c.config.MaxSize && c.entries[key] == nil {
		return NewDIDError(ErrCodeCacheError, "Cache is full", nil)
	}

	now := time.Now()
	c.entries[key] = &CacheEntry{
		Data:      value,
		ExpiresAt: now.Add(ttl),
		CachedAt:  now,
		Valid:     true,
		HitCount:  0,
	}

	return nil
}

// GetDID retrieves a cached DID document
func (c *Cache) GetDID(did string) (*DIDDocument, *CachedDID, bool) {
	if !c.config.Enabled {
		return nil, nil, false
	}

	c.mu.RLock()
	entry, exists := c.entries[did]
	c.mu.RUnlock()

	if !exists {
		return nil, nil, false
	}

	// Check if entry has expired
	now := time.Now()
	if now.After(entry.ExpiresAt) {
		c.Delete(did)
		return nil, nil, false
	}

	// Update hit count
	c.mu.Lock()
	entry.HitCount++
	c.mu.Unlock()

	// Type assert to DIDDocument
	if didDoc, ok := entry.Data.(*DIDDocument); ok {
		cached := &CachedDID{
			Document:  didDoc,
			ExpiresAt: entry.ExpiresAt,
			CachedAt:  entry.CachedAt,
			Valid:     entry.Valid,
		}
		return didDoc, cached, true
	}

	return nil, nil, false
}

// SetDID caches a DID document
func (c *Cache) SetDID(did string, document *DIDDocument) error {
	return c.Set(did, document)
}

// InvalidateDID invalidates all cached DIDs for a specific pattern
func (c *Cache) InvalidateDID(didPattern string) error {
	return c.Invalidate(didPattern)
}

// cleanupRoutine periodically removes expired entries
func (c *Cache) cleanupRoutine() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopChan:
			return
		}
	}
}

// cleanup removes expired entries from the cache
func (c *Cache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		delete(c.entries, key)
	}

	if len(expiredKeys) > 0 {
		fmt.Printf("[Cache] Cleaned up %d expired entries\n", len(expiredKeys))
	}
}

// Stop gracefully stops the cache cleanup routine
func (c *Cache) Stop() error {
	if c.config.Enabled {
		close(c.stopChan)
		c.wg.Wait()
	}
	return nil
}

// Size returns the current number of entries in the cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// Contains checks if a key exists in the cache
func (c *Cache) Contains(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return false
	}

	// Check expiration
	return time.Now().Before(entry.ExpiresAt)
}

// GetExpiry returns the expiration time of a cached entry
func (c *Cache) GetExpiry(key string) (time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return time.Time{}, false
	}

	return entry.ExpiresAt, true
}
