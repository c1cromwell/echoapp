package did

import (
	"context"
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// Resolver provides DID resolution for did:key. did:key is self-resolving
// (the public key is encoded in the identifier itself), so resolution is
// a pure-function operation against pkg/didkey.Parse — no network call,
// no chain query for the cryptographic resolution path.
//
// The cache + repository remain to (a) hold pre-built DIDDocuments and
// (b) let the wider service surface (controllers, multi-device records)
// stay attached to a DID. They are entirely optional for the cryptographic
// resolution path.
type Resolver struct {
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

// NewResolver creates a new did:key resolver.
func NewResolver(cache *Cache, repo Repository, config *DIDConfig) *Resolver {
	return &Resolver{
		cache:    cache,
		repo:     repo,
		config:   config,
		inflight: make(map[string]*resolutionInFlight),
	}
}

// Resolve resolves a did:key document. Hot path: cache hit. Cold path:
// pkg/didkey.Parse rebuilds the key, then we synthesize a minimal W3C
// document and cache it.
func (r *Resolver) Resolve(ctx context.Context, did string) (*DIDDocument, error) {
	if !isValidDIDFormat(did) {
		return nil, NewDIDError(ErrCodeInvalidDID, fmt.Sprintf("Invalid DID format: %s", did), nil)
	}

	if cachedDoc, _, found := r.cache.GetDID(did); found {
		return cachedDoc, nil
	}

	inf := r.getOrCreateInflight(did)
	defer r.removeInflight(did)

	if inf.done != nil && inf != r.getOrCreateInflight(did) {
		<-inf.done
		return inf.result, inf.err
	}

	ctx, cancel := context.WithTimeout(ctx, r.config.ResolutionTimeout)
	defer cancel()
	_ = ctx

	document, err := r.resolveFromKey(did)
	if err != nil {
		inf.err = err
		close(inf.done)
		return nil, err
	}

	if err := r.cache.SetDID(did, document); err != nil {
		fmt.Printf("[Resolver] Failed to cache DID %s: %v\n", did, err)
	}

	if err := r.repo.StoreDIDDocument(context.Background(), did, document); err != nil {
		fmt.Printf("[Resolver] Failed to store DID document %s in repository: %v\n", did, err)
	}

	inf.result = document
	close(inf.done)
	return document, nil
}

// ResolveWithMetadata resolves a DID and returns metadata about the
// resolution. BlockchainAnchored is always false for did:key.
func (r *Resolver) ResolveWithMetadata(ctx context.Context, did string) (*DIDDocument, *ResolutionMetadata, error) {
	startTime := time.Now()

	if cachedDoc, cached, found := r.cache.GetDID(did); found {
		metadata := &ResolutionMetadata{
			ResolutionTimestamp: startTime,
			CachedAt:            cached.CachedAt,
			CacheValid:          true,
			BlockchainAnchored:  false,
		}
		return cachedDoc, metadata, nil
	}

	document, err := r.Resolve(ctx, did)
	if err != nil {
		return nil, nil, err
	}

	metadata := &ResolutionMetadata{
		ResolutionTimestamp: startTime,
		CacheValid:          false,
		BlockchainAnchored:  false,
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

// resolveFromKey rebuilds the DIDDocument by re-deriving the public key
// from the did:key identifier itself (pkg/didkey.Parse). On parse error
// we fall back to the repository in case a multi-device controller
// document exists for this DID.
func (r *Resolver) resolveFromKey(did string) (*DIDDocument, error) {
	pub, err := didkey.Parse(did)
	if err == nil {
		// SEC1 uncompressed encoding for the public key, hex-encoded so
		// it round-trips cleanly through the existing PublicKeyBase64 field.
		// (The field name is legacy; we now store hex per the didkey API.)
		raw := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
		now := time.Now()
		keyID := fmt.Sprintf("%s#%s", did, did[8:])
		return &DIDDocument{
			Context: []string{
				"https://www.w3.org/ns/did/v1",
				"https://w3id.org/security/multikey/v1",
			},
			ID:         did,
			Controller: []string{did},
			PublicKey: []PublicKey{
				{
					ID:                 keyID,
					Type:               "JsonWebKey2020",
					Controller:         did,
					PublicKeyBase64:    hex.EncodeToString(raw),
					PublicKeyMultibase: did[8:],
				},
			},
			Authentication: []Authentication{
				{Type: "JsonWebKey2020", PublicKey: keyID},
			},
			AssertionMethod: []AssertionMethod{
				{Type: "JsonWebKey2020", PublicKey: keyID},
			},
			Created: now,
			Updated: now,
		}, nil
	}

	// Fallback: the DID is structurally valid (did:method:identifier) but
	// not a did:key. Try the repository for any cached document (e.g. a
	// multi-device controller record).
	repoDoc, repErr := r.repo.GetDIDDocument(context.Background(), did)
	if repErr == nil && repoDoc != nil {
		return repoDoc, nil
	}

	return nil, NewDIDError(ErrCodeResolutionFailed, "Failed to resolve did:key and no repository fallback", err)
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

// Health checks the resolver's dependencies. did:key resolution is a
// pure function so the only thing to verify is the repository (used
// for controller-document fallback).
func (r *Resolver) Health(ctx context.Context) (bool, error) {
	if err := r.repo.Health(ctx); err != nil {
		return false, fmt.Errorf("repository unhealthy: %w", err)
	}
	return true, nil
}
