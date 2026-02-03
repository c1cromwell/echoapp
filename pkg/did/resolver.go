package did

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Resolver provides DID resolution with caching capabilities
type Resolver struct {
	client   *AtalaClient
	cache    *Cache
	repo     Repository
	config   *DIDConfig
	mu       sync.RWMutex
	inflight map[string]*resolutionInFlight
}

// resolutionInFlight tracks concurrent resolution requests
type resolutionInFlight struct {
	result *DIDDocument
	err    error
	done   chan struct{}
	wg     sync.WaitGroup
}

// NewResolver creates a new DID resolver
func NewResolver(client *AtalaClient, cache *Cache, repo Repository, config *DIDConfig) *Resolver {
	return &Resolver{
		client:   client,
		cache:    cache,
		repo:     repo,
		config:   config,
		inflight: make(map[string]*resolutionInFlight),
	}
}

// Resolve resolves a DID document with caching and concurrency control
func (r *Resolver) Resolve(ctx context.Context, did string) (*DIDDocument, error) {
	// Validate DID format
	if !isValidDIDFormat(did) {
		return nil, NewDIDError(ErrCodeInvalidDID, fmt.Sprintf("Invalid DID format: %s", did), nil)
	}

	// Check cache first
	if cachedDoc, _, found := r.cache.GetDID(did); found {
		return cachedDoc, nil
	}

	// Check if resolution is already in-flight to avoid duplicate work
	inf := r.getOrCreateInflight(did)
	defer r.removeInflight(did)

	// If another goroutine is resolving, wait for it
	if inf.done != nil && inf != r.getOrCreateInflight(did) {
		<-inf.done
		return inf.result, inf.err
	}

	// Perform resolution with timeout
	ctx, cancel := context.WithTimeout(ctx, r.config.ResolutionTimeout)
	defer cancel()

	document, err := r.resolveFromBlockchain(ctx, did)
	if err != nil {
		inf.err = err
		close(inf.done)
		return nil, err
	}

	// Cache the resolved document
	if err := r.cache.SetDID(did, document); err != nil {
		// Log but don't fail if caching fails
		fmt.Printf("[Resolver] Failed to cache DID %s: %v\n", did, err)
	}

	// Also store in repository if available
	if err := r.repo.StoreDIDDocument(ctx, did, document); err != nil {
		fmt.Printf("[Resolver] Failed to store DID document %s in repository: %v\n", did, err)
	}

	inf.result = document
	close(inf.done)
	return document, nil
}

// ResolveWithMetadata resolves a DID and returns metadata about the resolution
func (r *Resolver) ResolveWithMetadata(ctx context.Context, did string) (*DIDDocument, *ResolutionMetadata, error) {
	startTime := time.Now()

	// Check cache first
	if cachedDoc, cached, found := r.cache.GetDID(did); found {
		metadata := &ResolutionMetadata{
			ResolutionTimestamp: startTime,
			CachedAt:            cached.CachedAt,
			CacheValid:          true,
			BlockchainAnchored:  false,
		}
		return cachedDoc, metadata, nil
	}

	// Resolve from blockchain
	document, err := r.Resolve(ctx, did)
	if err != nil {
		return nil, nil, err
	}

	// Get anchor information
	txHash, _, _, anchErr := r.repo.GetAnchor(ctx, did)

	metadata := &ResolutionMetadata{
		ResolutionTimestamp: startTime,
		CacheValid:          false,
		BlockchainAnchored:  anchErr == nil,
		TransactionHash:     txHash,
	}

	return document, metadata, nil
}

// ResolveMultiple resolves multiple DIDs concurrently
func (r *Resolver) ResolveMultiple(ctx context.Context, dids []string) (map[string]*DIDDocument, map[string]error) {
	results := make(map[string]*DIDDocument)
	errors := make(map[string]error)
	mu := sync.Mutex{}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrent resolutions

	for _, did := range dids {
		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			doc, err := r.Resolve(ctx, d)

			mu.Lock()
			if err != nil {
				errors[d] = err
			} else {
				results[d] = doc
			}
			mu.Unlock()
		}(did)
	}

	wg.Wait()
	return results, errors
}

// resolveFromBlockchain resolves a DID from the Cardano blockchain via Atala PRISM
func (r *Resolver) resolveFromBlockchain(ctx context.Context, did string) (*DIDDocument, error) {
	// Try to resolve via Atala PRISM first
	document, err := r.client.ResolveDID(ctx, did)
	if err == nil {
		return document, nil
	}

	// Fallback: Try to retrieve from repository if available
	repoDocs, repErr := r.repo.GetDIDDocument(ctx, did)
	if repErr == nil && repoDocs != nil {
		return repoDocs, nil
	}

	// If both fail, return the original Atala error
	return nil, err
}

// InvalidateCache invalidates the cache for a specific DID
func (r *Resolver) InvalidateCache(did string) error {
	return r.cache.InvalidateDID(did)
}

// InvalidateCachePattern invalidates cache entries matching a pattern
func (r *Resolver) InvalidateCachePattern(pattern string) error {
	return r.cache.Invalidate(pattern)
}

// ClearCache clears all cached entries
func (r *Resolver) ClearCache() error {
	return r.cache.Clear()
}

// CacheStats returns cache statistics
func (r *Resolver) CacheStats() map[string]interface{} {
	return r.cache.GetStats()
}

// getOrCreateInflight gets or creates an in-flight resolution tracking entry
func (r *Resolver) getOrCreateInflight(did string) *resolutionInFlight {
	r.mu.Lock()
	defer r.mu.Unlock()

	if inf, exists := r.inflight[did]; exists {
		return inf
	}

	inf := &resolutionInFlight{
		done: make(chan struct{}),
	}
	r.inflight[did] = inf
	return inf
}

// removeInflight removes an in-flight resolution tracking entry
func (r *Resolver) removeInflight(did string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.inflight, did)
}

// isValidDIDFormat validates DID format
func isValidDIDFormat(did string) bool {
	// Check for basic DID format: did:method:identifier
	if len(did) < 7 {
		return false
	}

	if did[0:4] != "did:" {
		return false
	}

	parts := countColons(did)
	return parts >= 2
}

// countColons counts the number of colons in a string
func countColons(s string) int {
	count := 0
	for _, ch := range s {
		if ch == ':' {
			count++
		}
	}
	return count
}

// BulkResolve resolves multiple DIDs with timeout handling
func (r *Resolver) BulkResolve(ctx context.Context, dids []string, timeout time.Duration) (map[string]*DIDDocument, map[string]error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.ResolveMultiple(ctx, dids)
}

// Health checks the resolver's dependencies
func (r *Resolver) Health(ctx context.Context) (bool, error) {
	// Check Atala client connectivity
	healthy, err := r.client.Health(ctx)
	if !healthy || err != nil {
		return false, fmt.Errorf("Atala PRISM client unhealthy: %w", err)
	}

	// Check repository connectivity
	if err := r.repo.Health(ctx); err != nil {
		return false, fmt.Errorf("Repository unhealthy: %w", err)
	}

	return true, nil
}
