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

// backgroundSync syncs revocation status with blockchain
func (rm *RevocationManager) backgroundSync() {
	for {
		select {
		case <-rm.syncTicker.C:
			rm.syncWithBlockchain()
		case <-rm.stopChan:
			rm.syncTicker.Stop()
			return
		}
	}
}

// syncWithBlockchain syncs with blockchain revocation registry
func (rm *RevocationManager) syncWithBlockchain() {
	// In production, sync with Cardano revocation registry
	// For now, this is a placeholder
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Clear cache periodically for fresh data
	rm.cacheMutex.Lock()
	rm.cache = make(map[string]*RevocationStatus)
	rm.cacheMutex.Unlock()

	// Log sync
	_ = ctx
}

// GetRevocationRegistry gets the revocation registry
func (rm *RevocationManager) GetRevocationRegistry(ctx context.Context) ([]RevocationRecord, error) {
	// In production, fetch from blockchain
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

// RevocationRegistry manages revocation registry on blockchain
type RevocationRegistry struct {
	storage     Storage
	chainClient interface{} // Cardano client in production
	indexPath   string
	cacheTTL    time.Duration
}

// NewRevocationRegistry creates new revocation registry
func NewRevocationRegistry(storage Storage, indexPath string, cacheTTL time.Duration) *RevocationRegistry {
	return &RevocationRegistry{
		storage:   storage,
		indexPath: indexPath,
		cacheTTL:  cacheTTL,
	}
}

// RegisterRevocation registers revocation in blockchain registry
func (rr *RevocationRegistry) RegisterRevocation(ctx context.Context, credentialID, issuerDID, reason string) (txHash string, err error) {
	// In production, write to Cardano blockchain
	txHash = "tx_revocation_" + credentialID
	return txHash, nil
}

// QueryRevocationStatus queries revocation status from blockchain
func (rr *RevocationRegistry) QueryRevocationStatus(ctx context.Context, credentialID string) (bool, error) {
	// In production, query Cardano blockchain
	// For now, query storage
	record, err := rr.storage.GetRevocationRecord(ctx, credentialID)
	if err != nil {
		return false, err
	}

	return record != nil, nil
}

// BuildRevocationIndex builds local revocation index
func (rr *RevocationRegistry) BuildRevocationIndex(ctx context.Context) error {
	// In production, download revocation registry from blockchain and build index
	// For now, placeholder
	return nil
}

// UpdateRevocationIndex updates revocation index from blockchain
func (rr *RevocationRegistry) UpdateRevocationIndex(ctx context.Context) error {
	// In production, sync with blockchain every period
	return nil
}

// GetRevocationIndexSize gets size of revocation index
func (rr *RevocationRegistry) GetRevocationIndexSize() int {
	// In production, return size of indexed revocations
	return 0
}
