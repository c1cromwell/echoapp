package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DeviceService manages device fingerprinting, registration, and validation.
type DeviceService struct {
	mu      sync.RWMutex
	devices map[string][]*DeviceRecord // key: user_id -> devices
}

// NewDeviceService creates a new device service.
func NewDeviceService() *DeviceService {
	return &DeviceService{
		devices: make(map[string][]*DeviceRecord),
	}
}

// ComputeDeviceHash computes a SHA-256 hash of stable device fields.
// Only uses fields that don't change on app or OS updates.
func ComputeDeviceHash(info DeviceInfo) string {
	data := fmt.Sprintf("%s:%s:%s", info.DeviceID, info.Platform, info.Model)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ValidateDeviceIntegrity checks for jailbreak and Secure Enclave presence.
func ValidateDeviceIntegrity(info DeviceInfo) *AuthError {
	if info.JailbreakStatus {
		return NewAuthError(ErrCodeDeviceIntegrity, 400)
	}
	if !info.SecureEnclave {
		return NewAuthError(ErrCodeDeviceIntegrity, 400)
	}
	return nil
}

// ValidateDeviceInfo checks that required fields are present.
func ValidateDeviceInfo(info DeviceInfo) *AuthError {
	if info.DeviceID == "" || info.Platform == "" || info.Model == "" || info.OSVersion == "" || info.AppVersion == "" {
		return &AuthError{
			Code:       ErrCodeDeviceIntegrity,
			Message:    "Missing required device information",
			HTTPStatus: 400,
		}
	}
	if info.Platform != "ios" {
		return &AuthError{
			Code:       ErrCodeDeviceIntegrity,
			Message:    "Unsupported platform",
			HTTPStatus: 400,
		}
	}
	return nil
}

// RegisterDevice registers a new device for a user.
func (ds *DeviceService) RegisterDevice(userID string, info DeviceInfo, ip string, credentialID string) *DeviceRecord {
	deviceHash := ComputeDeviceHash(info)
	friendlyName := info.Model

	record := &DeviceRecord{
		ID:           uuid.New().String(),
		UserID:       userID,
		DeviceHash:   deviceHash,
		FriendlyName: friendlyName,
		Platform:     info.Platform,
		OSVersion:    info.OSVersion,
		LastIP:       ip,
		LastActiveAt: time.Now(),
		CreatedAt:    time.Now(),
		CredentialID: credentialID,
	}

	ds.mu.Lock()
	ds.devices[userID] = append(ds.devices[userID], record)
	ds.mu.Unlock()

	return record
}

// GetDevices returns all registered devices for a user.
func (ds *DeviceService) GetDevices(userID string) []*DeviceRecord {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	devices := ds.devices[userID]
	result := make([]*DeviceRecord, len(devices))
	copy(result, devices)
	return result
}

// GetDeviceByHash finds a device by its hash for a user.
func (ds *DeviceService) GetDeviceByHash(userID, deviceHash string) *DeviceRecord {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	for _, d := range ds.devices[userID] {
		if d.DeviceHash == deviceHash {
			return d
		}
	}
	return nil
}

// IsKnownDevice checks if a device hash is registered for a user.
func (ds *DeviceService) IsKnownDevice(userID, deviceHash string) bool {
	return ds.GetDeviceByHash(userID, deviceHash) != nil
}

// UpdateDeviceActivity updates the last active time and IP for a device.
func (ds *DeviceService) UpdateDeviceActivity(userID, deviceHash, ip string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for _, d := range ds.devices[userID] {
		if d.DeviceHash == deviceHash {
			d.LastActiveAt = time.Now()
			d.LastIP = ip
			return
		}
	}
}

// RevokeDevice removes a device registration.
func (ds *DeviceService) RevokeDevice(userID, deviceID string) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	devices := ds.devices[userID]
	for i, d := range devices {
		if d.ID == deviceID {
			ds.devices[userID] = append(devices[:i], devices[i+1:]...)
			return true
		}
	}
	return false
}

// DeviceCount returns the number of registered devices for a user.
func (ds *DeviceService) DeviceCount(userID string) int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.devices[userID])
}
