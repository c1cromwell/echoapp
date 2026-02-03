package did

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DeviceManager handles multi-device registration and management
type DeviceManager struct {
	repo                 Repository
	cache                *Cache
	config               *DIDConfig
	mu                   sync.RWMutex
	pendingRegistrations map[string]*PendingDeviceRegistration
}

// PendingDeviceRegistration tracks a device registration in progress
type PendingDeviceRegistration struct {
	DeviceID   string
	DID        string
	Challenge  string
	CreatedAt  time.Time
	ExpiresAt  time.Time
	QRCodeData *QRCodeData
	PublicKey  string
}

// NewDeviceManager creates a new device manager
func NewDeviceManager(repo Repository, cache *Cache, config *DIDConfig) *DeviceManager {
	return &DeviceManager{
		repo:                 repo,
		cache:                cache,
		config:               config,
		pendingRegistrations: make(map[string]*PendingDeviceRegistration),
	}
}

// RegisterDevice adds a new device to a DID document
func (dm *DeviceManager) RegisterDevice(did string, device *DeviceRegistration) error {
	if did == "" || device == nil {
		return NewDIDError(ErrCodeInvalidRequest, "DID and device cannot be empty", nil)
	}

	if device.DeviceID == "" {
		return NewDIDError(ErrCodeInvalidRequest, "Device ID is required", nil)
	}

	if device.PublicKey == "" {
		return NewDIDError(ErrCodeInvalidRequest, "Public key is required", nil)
	}

	// Validate public key format
	if err := validatePublicKeyFormat(device.PublicKey); err != nil {
		return NewDIDError(ErrCodeInvalidPublicKey, fmt.Sprintf("Invalid public key format: %v", err), nil)
	}

	// Add device to repository
	if err := dm.repo.AddDevice(dm.getContext(), did, device); err != nil {
		return err
	}

	// Invalidate cache for this DID
	dm.cache.InvalidateDID(did)

	return nil
}

// UnregisterDevice removes a device from a DID document
func (dm *DeviceManager) UnregisterDevice(did, deviceID string) error {
	if did == "" || deviceID == "" {
		return NewDIDError(ErrCodeInvalidRequest, "DID and device ID are required", nil)
	}

	// Remove device from repository
	if err := dm.repo.RemoveDevice(dm.getContext(), did, deviceID); err != nil {
		return err
	}

	// Invalidate cache for this DID
	dm.cache.InvalidateDID(did)

	return nil
}

// ListDevices lists all devices for a DID
func (dm *DeviceManager) ListDevices(did string) ([]DeviceRegistration, error) {
	if did == "" {
		return nil, NewDIDError(ErrCodeInvalidRequest, "DID is required", nil)
	}

	devices, err := dm.repo.ListDevices(dm.getContext(), did)
	if err != nil {
		return nil, err
	}

	return devices, nil
}

// GetDevice retrieves a specific device
func (dm *DeviceManager) GetDevice(did, deviceID string) (*DeviceRegistration, error) {
	if did == "" || deviceID == "" {
		return nil, NewDIDError(ErrCodeInvalidRequest, "DID and device ID are required", nil)
	}

	device, err := dm.repo.GetDevice(dm.getContext(), did, deviceID)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// InitiateDeviceRegistration initiates a device registration flow
func (dm *DeviceManager) InitiateDeviceRegistration(did string) (*PendingDeviceRegistration, error) {
	if did == "" {
		return nil, NewDIDError(ErrCodeInvalidRequest, "DID is required", nil)
	}

	deviceID := uuid.New().String()
	challenge := uuid.New().String()
	now := time.Now()

	pending := &PendingDeviceRegistration{
		DeviceID:  deviceID,
		DID:       did,
		Challenge: challenge,
		CreatedAt: now,
		ExpiresAt: now.Add(15 * time.Minute), // 15-minute window for registration
		QRCodeData: &QRCodeData{
			DID:           did,
			DeviceID:      deviceID,
			ChallengeName: challenge,
			Timestamp:     now.Unix(),
			ValidUntil:    now.Add(15 * time.Minute).Unix(),
		},
	}

	dm.mu.Lock()
	dm.pendingRegistrations[deviceID] = pending
	dm.mu.Unlock()

	return pending, nil
}

// GenerateDeviceRegistrationQRCode generates QR code data for device registration
// Note: Actual QR code image generation should use a proper QR code library like github.com/skip2/go-qrcode
func (dm *DeviceManager) GenerateDeviceRegistrationQRCode(did string) (*QRCodeData, string, error) {
	pending, err := dm.InitiateDeviceRegistration(did)
	if err != nil {
		return nil, "", err
	}

	// Encode QR code data to JSON
	qrDataJSON, err := json.Marshal(pending.QRCodeData)
	if err != nil {
		return nil, "", NewDIDError(ErrCodeQRCodeGenerationFailed, "Failed to marshal QR code data", err)
	}

	// Return the JSON as the QR code data payload
	// In production, encode this to an actual QR code image
	qrCodePayload := base64.StdEncoding.EncodeToString(qrDataJSON)

	return pending.QRCodeData, qrCodePayload, nil
}

// VerifyDeviceRegistrationChallenge verifies the device registration challenge
func (dm *DeviceManager) VerifyDeviceRegistrationChallenge(deviceID, challenge string) (*PendingDeviceRegistration, error) {
	dm.mu.RLock()
	pending, exists := dm.pendingRegistrations[deviceID]
	dm.mu.RUnlock()

	if !exists {
		return nil, NewDIDError(ErrCodeInvalidRequest, fmt.Sprintf("No pending registration for device: %s", deviceID), nil)
	}

	if time.Now().After(pending.ExpiresAt) {
		dm.mu.Lock()
		delete(dm.pendingRegistrations, deviceID)
		dm.mu.Unlock()
		return nil, NewDIDError(ErrCodeInvalidRequest, "Device registration challenge expired", nil)
	}

	if pending.Challenge != challenge {
		return nil, NewDIDError(ErrCodeUnauthorized, "Invalid challenge", nil)
	}

	return pending, nil
}

// CompleteDeviceRegistration completes the device registration flow
func (dm *DeviceManager) CompleteDeviceRegistration(deviceID, challenge, publicKey, deviceName string) (*DeviceRegistration, error) {
	pending, err := dm.VerifyDeviceRegistrationChallenge(deviceID, challenge)
	if err != nil {
		return nil, err
	}

	// Store public key temporarily
	dm.mu.Lock()
	pending.PublicKey = publicKey
	dm.mu.Unlock()

	// Validate public key
	if err := validatePublicKeyFormat(publicKey); err != nil {
		return nil, NewDIDError(ErrCodeInvalidPublicKey, fmt.Sprintf("Invalid public key format: %v", err), nil)
	}

	// Create device registration
	device := &DeviceRegistration{
		DeviceID:   deviceID,
		DeviceName: deviceName,
		PublicKey:  publicKey,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
		IsActive:   true,
	}

	// Register device
	if err := dm.RegisterDevice(pending.DID, device); err != nil {
		return nil, err
	}

	// Clean up pending registration
	dm.mu.Lock()
	delete(dm.pendingRegistrations, deviceID)
	dm.mu.Unlock()

	return device, nil
}

// UpdateDeviceLastUsed updates the last used timestamp for a device
func (dm *DeviceManager) UpdateDeviceLastUsed(did, deviceID string) error {
	device, err := dm.GetDevice(did, deviceID)
	if err != nil {
		return err
	}

	device.LastUsedAt = time.Now()
	if err := dm.repo.UpdateDevice(dm.getContext(), did, device); err != nil {
		return err
	}

	return nil
}

// SetDeviceActive sets the active status of a device
func (dm *DeviceManager) SetDeviceActive(did, deviceID string, active bool) error {
	device, err := dm.GetDevice(did, deviceID)
	if err != nil {
		return err
	}

	device.IsActive = active
	if err := dm.repo.UpdateDevice(dm.getContext(), did, device); err != nil {
		return err
	}

	// Invalidate cache
	dm.cache.InvalidateDID(did)

	return nil
}

// GetActiveDevices returns all active devices for a DID
func (dm *DeviceManager) GetActiveDevices(did string) ([]DeviceRegistration, error) {
	devices, err := dm.ListDevices(did)
	if err != nil {
		return nil, err
	}

	activeDevices := make([]DeviceRegistration, 0)
	for _, device := range devices {
		if device.IsActive {
			activeDevices = append(activeDevices, device)
		}
	}

	return activeDevices, nil
}

// ValidateDevicePublicKey validates that a public key belongs to a device
func (dm *DeviceManager) ValidateDevicePublicKey(did, deviceID, publicKey string) (bool, error) {
	device, err := dm.GetDevice(did, deviceID)
	if err != nil {
		return false, err
	}

	return device.PublicKey == publicKey, nil
}

// getContext returns a context for repository operations
func (dm *DeviceManager) getContext() context.Context {
	return context.Background()
}

// validatePublicKeyFormat validates the format of a public key
func validatePublicKeyFormat(publicKey string) error {
	if len(publicKey) == 0 {
		return fmt.Errorf("public key cannot be empty")
	}

	// For Ed25519 keys, accept base64-encoded 32-byte keys (44 characters when base64 encoded)
	// or multibase format
	if !isValidBase64(publicKey) && !isValidMultibase(publicKey) {
		return fmt.Errorf("public key must be base64 or multibase encoded")
	}

	return nil
}

// isValidBase64 checks if a string is valid base64
func isValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		// Try with padding
		_, err = base64.RawStdEncoding.DecodeString(s)
	}
	return err == nil
}

// isValidMultibase checks if a string is in multibase format
func isValidMultibase(s string) bool {
	// Multibase format: first character is the base encoding, followed by encoded data
	// Common prefixes: z (base58btc), b (base32), m (base32hex), etc.
	return len(s) > 1 && (s[0] == 'z' || s[0] == 'b' || s[0] == 'm')
}

// GetPendingRegistrations returns all pending registrations (for debugging/admin)
func (dm *DeviceManager) GetPendingRegistrations() map[string]*PendingDeviceRegistration {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	pending := make(map[string]*PendingDeviceRegistration)
	for k, v := range dm.pendingRegistrations {
		pending[k] = v
	}
	return pending
}

// CleanupExpiredRegistrations removes expired pending registrations
func (dm *DeviceManager) CleanupExpiredRegistrations() int {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	now := time.Now()
	expired := 0

	for deviceID, pending := range dm.pendingRegistrations {
		if now.After(pending.ExpiresAt) {
			delete(dm.pendingRegistrations, deviceID)
			expired++
		}
	}

	return expired
}
