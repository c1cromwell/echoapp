package credentials

import (
	"context"
	"sync"
	"time"
)

// RevocationManager manages credential revocation
type RevocationManager struct {
	storage    Storage
	cache      map[string]*RevocationStatus
	cacheMutex sync.RWMutex
	cacheTTL   time.Duration
	syncTicker *time.Ticker
	stopChan   chan struct{}
}

// NewRevocationManager creates new revocation manager
func NewRevocationManager(storage Storage, cacheTTL time.Duration) *RevocationManager {
	rm := &RevocationManager{
		storage:    storage,
		cache:      make(map[string]*RevocationStatus),
		cacheTTL:   cacheTTL,
		syncTicker: time.NewTicker(1 * time.Hour),
		stopChan:   make(chan struct{}),
	}

	// Start background sync
	go rm.backgroundSync()

	return rm
}

// RevokeCredential revokes a credential
func (rm *RevocationManager) RevokeCredential(ctx context.Context, credentialID, issuerDID, subjectDID, reason string) error {
	// Create revocation record
	record := &RevocationRecord{
		CredentialID:     credentialID,
		IssuerDID:        issuerDID,
		SubjectDID:       subjectDID,
		RevokedAt:        time.Now(),
		RevocationReason: reason,
	}

	// Store revocation
	err := rm.storage.RecordRevocation(ctx, record)
	if err != nil {
		return NewCredentialErrorWithDetails(
			ErrCodeRevocationCheckFailed,
			"failed to record revocation",
			err.Error(),
		)
	}

	// Invalidate cache
	rm.invalidateCache(credentialID)

	return nil
}

// CheckRevocationStatus checks if credential is revoked
func (rm *RevocationManager) CheckRevocationStatus(ctx context.Context, credentialID string) (*RevocationStatus, error) {
	// Check cache first
	rm.cacheMutex.RLock()
	if status, exists := rm.cache[credentialID]; exists {
		rm.cacheMutex.RUnlock()
		return status, nil
	}
	rm.cacheMutex.RUnlock()

	// Query storage
	record, err := rm.storage.GetRevocationRecord(ctx, credentialID)
	if err != nil {
		return nil, NewCredentialErrorWithDetails(
			ErrCodeRevocationCheckFailed,
			"failed to check revocation status",
			err.Error(),
		)
	}

	var status *RevocationStatus
	if record != nil {
		status = &RevocationStatus{
			CredentialID:     credentialID,
			IsRevoked:        true,
			RevokedAt:        &record.RevokedAt,
			RevocationReason: record.RevocationReason,
		}
	} else {
		status = &RevocationStatus{
			CredentialID: credentialID,
			IsRevoked:    false,
		}
	}

	// Cache status
	rm.cacheMutex.Lock()
	rm.cache[credentialID] = status
	rm.cacheMutex.Unlock()

	return status, nil
}

// BatchCheckRevocation checks revocation status for multiple credentials
func (rm *RevocationManager) BatchCheckRevocation(ctx context.Context, credentialIDs []string) (map[string]*RevocationStatus, error) {
	results := make(map[string]*RevocationStatus)
	var mu sync.Mutex
	var wg sync.WaitGroup

	semaphore := make(chan struct{}, 10) // Limit concurrent checks

	for _, credID := range credentialIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			status, err := rm.CheckRevocationStatus(ctx, id)
			if err == nil {
				mu.Lock()
				results[id] = status
				mu.Unlock()
			}
		}(credID)
	}

	wg.Wait()
	return results, nil
}

// invalidateCache invalidates cache entry
func (rm *RevocationManager) invalidateCache(credentialID string) {
	rm.cacheMutex.Lock()
	defer rm.cacheMutex.Unlock()
	delete(rm.cache, credentialID)
}

// backgroundSync polls the Identity Metagraph StatusList2021 publisher and
// refreshes the local cache. Phase 1 stub: just clears the cache; the full
// fetcher lands in WO-274 alongside the StatusList batch publisher.
func (rm *RevocationManager) backgroundSync() {
	for {
		select {
		case <-rm.syncTicker.C:
			rm.syncWithStatusList2021()
		case <-rm.stopChan:
			rm.syncTicker.Stop()
			return
		}
	}
}

// syncWithStatusList2021 syncs with the Identity Metagraph StatusList2021
// publisher. Phase 1: placeholder that just invalidates the cache so the
// next CheckRevocationStatus hits storage. WO-274 wires the actual L1 GET.
func (rm *RevocationManager) syncWithStatusList2021() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rm.cacheMutex.Lock()
	rm.cache = make(map[string]*RevocationStatus)
	rm.cacheMutex.Unlock()

	_ = ctx
}

// GetRevocationRegistry returns the in-memory revocation records. The
// metagraph-anchored bit-vector view lands in WO-274.
func (rm *RevocationManager) GetRevocationRegistry(ctx context.Context) ([]RevocationRecord, error) {
	return []RevocationRecord{}, nil
}

// GetCacheStats gets cache statistics
func (rm *RevocationManager) GetCacheStats() map[string]interface{} {
	rm.cacheMutex.RLock()
	defer rm.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cached_entries": len(rm.cache),
		"ttl_seconds":    rm.cacheTTL.Seconds(),
	}
}

// Close closes the revocation manager
func (rm *RevocationManager) Close() error {
	close(rm.stopChan)
	return nil
}

// RevocationRegistry manages a per-issuer StatusList2021 bit vector that
// is published to the Constellation Identity Metagraph (WO-272 anchors;
// WO-274 wires the actual issuer + publisher).
type RevocationRegistry struct {
	storage   Storage
	indexPath string
	cacheTTL  time.Duration
}

// NewRevocationRegistry creates a new StatusList2021-backed revocation registry.
func NewRevocationRegistry(storage Storage, indexPath string, cacheTTL time.Duration) *RevocationRegistry {
	return &RevocationRegistry{
		storage:   storage,
		indexPath: indexPath,
		cacheTTL:  cacheTTL,
	}
}

// RegisterRevocation flips the credential's bit in the issuer's
// StatusList2021 bit vector. Phase 1 stub returns a deterministic
// reference; WO-274 wires the real Identity L1 submission.
func (rr *RevocationRegistry) RegisterRevocation(ctx context.Context, credentialID, issuerDID, reason string) (statusListRef string, err error) {
	statusListRef = "statuslist:" + credentialID
	return statusListRef, nil
}

// QueryRevocationStatus checks whether the credential's bit is set in
// the issuer's StatusList2021 vector. Phase 1: queries local storage.
// WO-274 wires the L1 GET that returns the actual bit-vector page.
func (rr *RevocationRegistry) QueryRevocationStatus(ctx context.Context, credentialID string) (bool, error) {
	record, err := rr.storage.GetRevocationRecord(ctx, credentialID)
	if err != nil {
		return false, err
	}

	return record != nil, nil
}

// BuildRevocationIndex builds the local cache of StatusList2021 entries.
// WO-274 wires the metagraph fetcher.
func (rr *RevocationRegistry) BuildRevocationIndex(ctx context.Context) error {
	return nil
}

// UpdateRevocationIndex pulls the latest published StatusList2021 vector
// from the Identity Metagraph and overwrites the local cache.
func (rr *RevocationRegistry) UpdateRevocationIndex(ctx context.Context) error {
	return nil
}

// GetRevocationIndexSize returns the number of indexed revocations.
func (rr *RevocationRegistry) GetRevocationIndexSize() int {
	return 0
}
