package did

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/thechadcromwell/echoapp/internal/crypto"
)

// Service provides core DID management operations
type Service struct {
	client        *AtalaClient
	resolver      *Resolver
	deviceManager *DeviceManager
	repo          Repository
	cache         *Cache
	config        *DIDConfig
	cryptoUtils   *crypto.CryptoUtils
	mu            sync.RWMutex
	generation    map[string]*GenerationProgress
}

// GenerationProgress tracks DID generation progress
type GenerationProgress struct {
	DID             string
	Status          string
	Progress        int
	StartedAt       time.Time
	EstimatedEnd    time.Time
	TransactionHash string
}

// NewService creates a new DID service
func NewService(
	client *AtalaClient,
	resolver *Resolver,
	deviceManager *DeviceManager,
	repo Repository,
	cache *Cache,
	config *DIDConfig,
	cryptoUtils *crypto.CryptoUtils,
) *Service {
	return &Service{
		client:        client,
		resolver:      resolver,
		deviceManager: deviceManager,
		repo:          repo,
		cache:         cache,
		config:        config,
		cryptoUtils:   cryptoUtils,
		generation:    make(map[string]*GenerationProgress),
	}
}

// CreateDID generates a new DID and anchors it to the blockchain
func (s *Service) CreateDID(ctx context.Context, req *DIDCreationRequest) (*DIDCreationResponse, error) {
	// Validate request
	if err := s.validateDIDCreationRequest(req); err != nil {
		return nil, err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.config.GenerationTimeout)
	defer cancel()

	// Generate unique identifier
	uniqueID := uuid.New().String()
	did := fmt.Sprintf("did:prism:cardano:%s", uniqueID)

	// Track generation progress
	progress := &GenerationProgress{
		DID:          did,
		Status:       "initializing",
		Progress:     5,
		StartedAt:    time.Now(),
		EstimatedEnd: time.Now().Add(s.config.GenerationTimeout),
	}
	s.trackGenerationProgress(did, progress)
	defer s.removeGenerationProgress(did)

	// Generate public key if not provided
	publicKey := req.PublicKey
	if publicKey == "" {
		keyPair, err := s.cryptoUtils.GenerateKey()
		if err != nil {
			return nil, NewDIDError(ErrCodeCryptoFailed, "Failed to generate key pair", err)
		}
		publicKey = keyPair.PublicKeyHex()
		progress.Progress = 15
	}

	// Create DID document
	document := s.createDIDDocument(did, publicKey, req)
	progress.Progress = 30

	// Verify document structure
	if err := s.validateDIDDocument(document); err != nil {
		return nil, NewDIDError(ErrCodeInvalidDocument, "Invalid DID document", err)
	}
	progress.Progress = 45

	// Store DID mapping
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

	if err := s.repo.CreateDIDMapping(ctx, mapping); err != nil {
		return nil, err
	}
	progress.Progress = 60

	// Store DID document
	if err := s.repo.StoreDIDDocument(ctx, did, document); err != nil {
		// Log but continue - not critical
		fmt.Printf("[Service] Failed to store DID document: %v\n", err)
	}
	progress.Progress = 70

	// Anchor to blockchain via Atala PRISM
	txHash, err := s.client.AnchorDID(ctx, did, document)
	if err != nil {
		return nil, NewDIDError(ErrCodeAnchoringFailed, "Failed to anchor DID to blockchain", err)
	}
	progress.TransactionHash = txHash
	progress.Progress = 85

	// Record anchor
	if err := s.repo.RecordAnchor(ctx, did, txHash, 0); err != nil {
		// Log but continue - transaction is recorded on-chain
		fmt.Printf("[Service] Failed to record anchor: %v\n", err)
	}
	progress.Progress = 95

	// Cache the document
	if err := s.cache.SetDID(did, document); err != nil {
		fmt.Printf("[Service] Failed to cache DID document: %v\n", err)
	}

	progress.Progress = 100
	progress.Status = "completed"

	return &DIDCreationResponse{
		DID:              did,
		Document:         document,
		TransactionHash:  txHash,
		AnchoredAt:       time.Now(),
		ResolutionStatus: "anchored",
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

// UpdateDID updates a DID document
func (s *Service) UpdateDID(ctx context.Context, did string, document *DIDDocument) error {
	// Validate document
	if err := s.validateDIDDocument(document); err != nil {
		return NewDIDError(ErrCodeInvalidDocument, "Invalid DID document", err)
	}

	// Update in Atala PRISM
	if err := s.client.UpdateDID(ctx, did, document); err != nil {
		return err
	}

	// Update in repository
	if err := s.repo.UpdateDIDDocument(ctx, did, document); err != nil {
		fmt.Printf("[Service] Failed to update DID in repository: %v\n", err)
	}

	// Invalidate cache
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

// VerifyDIDDocument verifies a DID document
func (s *Service) VerifyDIDDocument(ctx context.Context, document *DIDDocument) (bool, error) {
	return s.client.VerifyDIDDocument(ctx, document)
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

// createDIDDocument creates a DID document from a creation request
func (s *Service) createDIDDocument(did, publicKey string, req *DIDCreationRequest) *DIDDocument {
	now := time.Now()

	keyID := fmt.Sprintf("%s#key-1", did)
	verificationKeyID := fmt.Sprintf("%s#auth-key-1", did)
	assertionKeyID := fmt.Sprintf("%s#assertion-key-1", did)

	return &DIDDocument{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2018/v1",
		},
		ID:         did,
		Controller: []string{did},
		PublicKey: []PublicKey{
			{
				ID:              keyID,
				Type:            "Ed25519VerificationKey2018",
				Controller:      did,
				PublicKeyBase64: publicKey,
			},
		},
		Authentication: []Authentication{
			{
				Type:      "Ed25519VerificationKey2018",
				PublicKey: verificationKeyID,
			},
		},
		AssertionMethod: []AssertionMethod{
			{
				Type:      "Ed25519VerificationKey2018",
				PublicKey: assertionKeyID,
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
		return fmt.Errorf("At least one public key is required")
	}

	if len(doc.Authentication) == 0 {
		return fmt.Errorf("At least one authentication method is required")
	}

	if len(doc.AssertionMethod) == 0 {
		return fmt.Errorf("At least one assertion method is required")
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
	if err := s.client.Close(); err != nil {
		return err
	}

	if err := s.cache.Stop(); err != nil {
		return err
	}

	return nil
}
