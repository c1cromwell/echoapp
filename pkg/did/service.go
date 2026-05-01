package did

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/crypto"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// Service provides core DID management operations using the W3C did:key
// method (P-256 / multicodec p256-pub / base58btc multibase). Phase 1
// per ADR-0001: identifiers are self-certifying — the DID *is* the
// public key, so there is no on-chain anchoring step.
type Service struct {
	resolver      *Resolver
	deviceManager *DeviceManager
	repo          Repository
	cache         *Cache
	config        *DIDConfig
	cryptoUtils   *crypto.CryptoUtils
	mu            sync.RWMutex
	generation    map[string]*GenerationProgress
}

// GenerationProgress tracks DID generation progress. The TransactionHash
// field is retained for backward-compatibility with consumers and is
// always empty for did:key (no on-chain anchor).
type GenerationProgress struct {
	DID             string
	Status          string
	Progress        int
	StartedAt       time.Time
	EstimatedEnd    time.Time
	TransactionHash string
}

// NewService creates a new DID service backed by the canonical
// pkg/didkey derivation library.
func NewService(
	resolver *Resolver,
	deviceManager *DeviceManager,
	repo Repository,
	cache *Cache,
	config *DIDConfig,
	cryptoUtils *crypto.CryptoUtils,
) *Service {
	return &Service{
		resolver:      resolver,
		deviceManager: deviceManager,
		repo:          repo,
		cache:         cache,
		config:        config,
		cryptoUtils:   cryptoUtils,
		generation:    make(map[string]*GenerationProgress),
	}
}

// CreateDID derives a new did:key from the request's public key and
// records the user/device mapping. did:key is self-certifying so there
// is no separate "anchor" step — the DID exists as soon as the public
// key is published off-chain in the issuance record (WO-272).
func (s *Service) CreateDID(ctx context.Context, req *DIDCreationRequest) (*DIDCreationResponse, error) {
	if err := s.validateDIDCreationRequest(req); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.GenerationTimeout)
	defer cancel()
	_ = ctx

	publicKey := req.PublicKey
	if publicKey == "" {
		return nil, NewDIDError(ErrCodeInvalidPublicKey, "public_key is required for did:key derivation", nil)
	}

	did, err := didkey.DeriveFromPublicKeyHex(publicKey)
	if err != nil {
		return nil, NewDIDError(ErrCodeInvalidPublicKey, "Failed to derive did:key from public key", err)
	}

	progress := &GenerationProgress{
		DID:          did,
		Status:       "deriving",
		Progress:     50,
		StartedAt:    time.Now(),
		EstimatedEnd: time.Now().Add(s.config.GenerationTimeout),
	}
	s.trackGenerationProgress(did, progress)
	defer s.removeGenerationProgress(did)

	document := s.createDIDDocument(did, publicKey, req)

	if err := s.validateDIDDocument(document); err != nil {
		return nil, NewDIDError(ErrCodeInvalidDocument, "Invalid DID document", err)
	}

	mapping := &DIDMapping{
		DID:           did,
		UserID:        req.UserID,
		AccountID:     uuid.New().String(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		IsActive:      true,
		PrimaryDevice: req.DeviceID,
		Devices: []DeviceRegistration{
			{
				DeviceID:   req.DeviceID,
				DeviceName: req.DeviceName,
				PublicKey:  publicKey,
				CreatedAt:  time.Now(),
				LastUsedAt: time.Now(),
				IsActive:   true,
			},
		},
	}

	if err := s.repo.CreateDIDMapping(context.Background(), mapping); err != nil {
		return nil, err
	}

	if err := s.repo.StoreDIDDocument(context.Background(), did, document); err != nil {
		fmt.Printf("[Service] Failed to store DID document: %v\n", err)
	}

	if err := s.cache.SetDID(did, document); err != nil {
		fmt.Printf("[Service] Failed to cache DID document: %v\n", err)
	}

	progress.Progress = 100
	progress.Status = "completed"

	return &DIDCreationResponse{
		DID:              did,
		Document:         document,
		AnchoredAt:       time.Now(),
		ResolutionStatus: "resolved",
	}, nil
}

// ResolveDID resolves a DID document
func (s *Service) ResolveDID(ctx context.Context, did string) (*DIDDocument, error) {
	return s.resolver.Resolve(ctx, did)
}

// ResolveDIDWithMetadata resolves a DID with metadata
func (s *Service) ResolveDIDWithMetadata(ctx context.Context, did string) (*DIDDocument, *ResolutionMetadata, error) {
	return s.resolver.ResolveWithMetadata(ctx, did)
}

// UpdateDID updates a stored DID document. did:key documents are
// immutable in W3C semantics (the DID *is* the key), so this only
// updates the cached/persisted document — it never re-anchors.
func (s *Service) UpdateDID(ctx context.Context, did string, document *DIDDocument) error {
	if err := s.validateDIDDocument(document); err != nil {
		return NewDIDError(ErrCodeInvalidDocument, "Invalid DID document", err)
	}

	if err := s.repo.UpdateDIDDocument(ctx, did, document); err != nil {
		fmt.Printf("[Service] Failed to update DID in repository: %v\n", err)
	}

	s.cache.InvalidateDID(did)
	return nil
}

// RegisterDevice adds a new device to a DID
func (s *Service) RegisterDevice(did string, device *DeviceRegistration) error {
	return s.deviceManager.RegisterDevice(did, device)
}

// UnregisterDevice removes a device from a DID
func (s *Service) UnregisterDevice(did, deviceID string) error {
	return s.deviceManager.UnregisterDevice(did, deviceID)
}

// GetDevices lists all devices for a DID
func (s *Service) GetDevices(did string) ([]DeviceRegistration, error) {
	return s.deviceManager.ListDevices(did)
}

// InitiateDeviceRegistration initiates a device registration flow
func (s *Service) InitiateDeviceRegistration(did string) (*PendingDeviceRegistration, error) {
	return s.deviceManager.InitiateDeviceRegistration(did)
}

// GenerateQRCodeForDeviceRegistration generates a QR code for device registration
func (s *Service) GenerateQRCodeForDeviceRegistration(did string) (*QRCodeData, string, error) {
	return s.deviceManager.GenerateDeviceRegistrationQRCode(did)
}

// CompleteDeviceRegistration completes the device registration flow
func (s *Service) CompleteDeviceRegistration(deviceID, challenge, publicKey, deviceName string) (*DeviceRegistration, error) {
	return s.deviceManager.CompleteDeviceRegistration(deviceID, challenge, publicKey, deviceName)
}

// GetDIDMapping retrieves a DID mapping by DID
func (s *Service) GetDIDMapping(ctx context.Context, did string) (*DIDMapping, error) {
	return s.repo.GetDIDByID(ctx, did)
}

// GetDIDMappingByUserID retrieves a DID mapping by user ID
func (s *Service) GetDIDMappingByUserID(ctx context.Context, userID string) (*DIDMapping, error) {
	return s.repo.GetDIDByUserID(ctx, userID)
}

// VerifyDIDDocument verifies a DID document by re-deriving the did:key
// from its embedded public key and comparing it to the document's id.
// Returns true iff the embedded public key produces the document id.
func (s *Service) VerifyDIDDocument(ctx context.Context, document *DIDDocument) (bool, error) {
	if document == nil || len(document.PublicKey) == 0 {
		return false, NewDIDError(ErrCodeInvalidDocument, "DID document missing public key", nil)
	}
	derived, err := didkey.DeriveFromPublicKeyHex(document.PublicKey[0].PublicKeyBase64)
	if err != nil {
		return false, NewDIDError(ErrCodeInvalidPublicKey, "Failed to re-derive did:key", err)
	}
	return derived == document.ID, nil
}

// GetGenerationProgress returns the progress of DID generation
func (s *Service) GetGenerationProgress(did string) *GenerationProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if progress, exists := s.generation[did]; exists {
		return progress
	}
	return nil
}

// InvalidateCache invalidates the cache for a specific DID
func (s *Service) InvalidateCache(did string) error {
	return s.cache.InvalidateDID(did)
}

// ClearCache clears all cache entries
func (s *Service) ClearCache() error {
	return s.cache.Clear()
}

// GetCacheStats returns cache statistics
func (s *Service) GetCacheStats() map[string]interface{} {
	return s.cache.GetStats()
}

// Health checks all service dependencies
func (s *Service) Health(ctx context.Context) (bool, error) {
	return s.resolver.Health(ctx)
}

// createDIDDocument creates a W3C DID document from a creation request.
// The verificationMethod is JsonWebKey2020 / Multibase per the did:key
// spec (https://w3c-ccg.github.io/did-method-key/).
func (s *Service) createDIDDocument(did, publicKey string, req *DIDCreationRequest) *DIDDocument {
	now := time.Now()

	keyID := fmt.Sprintf("%s#%s", did, did[8:]) // did:key uses the multibase identifier as the fragment

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
				PublicKeyBase64:    publicKey,
				PublicKeyMultibase: did[8:],
			},
		},
		Authentication: []Authentication{
			{
				Type:      "JsonWebKey2020",
				PublicKey: keyID,
			},
		},
		AssertionMethod: []AssertionMethod{
			{
				Type:      "JsonWebKey2020",
				PublicKey: keyID,
			},
		},
		Service: []ServiceEndpoint{
			{
				ID:              fmt.Sprintf("%s#inbox", did),
				Type:            "DIDCommMessaging",
				ServiceEndpoint: "https://service.example.com",
			},
		},
		Created: now,
		Updated: now,
	}
}

// validateDIDCreationRequest validates a DID creation request
func (s *Service) validateDIDCreationRequest(req *DIDCreationRequest) error {
	ve := &ValidationErrors{}

	if req.UserID == "" {
		ve.Add("user_id", "User ID is required")
	}

	if req.DeviceID == "" {
		ve.Add("device_id", "Device ID is required")
	}

	if req.PublicKey != "" {
		if err := validatePublicKeyFormat(req.PublicKey); err != nil {
			ve.Add("public_key", fmt.Sprintf("Invalid public key format: %v", err))
		}
	}

	if ve.HasErrors() {
		return ve
	}

	return nil
}

// validateDIDDocument validates a DID document
func (s *Service) validateDIDDocument(doc *DIDDocument) error {
	if doc == nil {
		return fmt.Errorf("DID document cannot be nil")
	}

	if doc.ID == "" {
		return fmt.Errorf("DID ID is required")
	}

	if len(doc.PublicKey) == 0 {
		return fmt.Errorf("at least one public key is required")
	}

	if len(doc.Authentication) == 0 {
		return fmt.Errorf("at least one authentication method is required")
	}

	if len(doc.AssertionMethod) == 0 {
		return fmt.Errorf("at least one assertion method is required")
	}

	return nil
}

// trackGenerationProgress tracks DID generation progress
func (s *Service) trackGenerationProgress(did string, progress *GenerationProgress) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.generation[did] = progress
}

// removeGenerationProgress removes DID generation progress tracking
func (s *Service) removeGenerationProgress(did string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.generation, did)
}

// Close gracefully shuts down the service
func (s *Service) Close() error {
	if err := s.cache.Stop(); err != nil {
		return err
	}

	return nil
}
