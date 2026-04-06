package auth

import (
	"testing"
)

func testDeviceInfo() DeviceInfo {
	return DeviceInfo{
		DeviceID:      "test-device-123",
		Platform:      "ios",
		OSVersion:     "17.4",
		AppVersion:    "1.0.0",
		Model:         "iPhone 15 Pro",
		Locale:        "en_US",
		Timezone:      "America/New_York",
		SecureEnclave: true,
		BiometricType: "face_id",
	}
}

func TestComputeDeviceHash_Deterministic(t *testing.T) {
	info := testDeviceInfo()
	h1 := ComputeDeviceHash(info)
	h2 := ComputeDeviceHash(info)

	if h1 != h2 {
		t.Error("same device info should produce same hash")
	}
}

func TestComputeDeviceHash_StableFields(t *testing.T) {
	info1 := testDeviceInfo()
	info2 := testDeviceInfo()

	// Changing app version should NOT change the hash (unstable field)
	info2.AppVersion = "2.0.0"
	info2.OSVersion = "18.0"

	h1 := ComputeDeviceHash(info1)
	h2 := ComputeDeviceHash(info2)

	if h1 != h2 {
		t.Error("device hash should only use stable fields (DeviceID, Platform, Model)")
	}
}

func TestComputeDeviceHash_DifferentDevices(t *testing.T) {
	info1 := testDeviceInfo()
	info2 := testDeviceInfo()
	info2.DeviceID = "different-device-456"

	h1 := ComputeDeviceHash(info1)
	h2 := ComputeDeviceHash(info2)

	if h1 == h2 {
		t.Error("different devices should produce different hashes")
	}
}

func TestValidateDeviceIntegrity_JailbreakRejected(t *testing.T) {
	info := testDeviceInfo()
	info.JailbreakStatus = true

	err := ValidateDeviceIntegrity(info)
	if err == nil {
		t.Error("jailbroken device should be rejected")
	}
	if err.Code != ErrCodeDeviceIntegrity {
		t.Errorf("expected AUTH_010, got %s", err.Code)
	}
}

func TestValidateDeviceIntegrity_NoSecureEnclave(t *testing.T) {
	info := testDeviceInfo()
	info.SecureEnclave = false

	err := ValidateDeviceIntegrity(info)
	if err == nil {
		t.Error("device without Secure Enclave should be rejected")
	}
}

func TestValidateDeviceIntegrity_ValidDevice(t *testing.T) {
	info := testDeviceInfo()
	err := ValidateDeviceIntegrity(info)
	if err != nil {
		t.Errorf("valid device should pass: %v", err)
	}
}

func TestValidateDeviceInfo_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		info DeviceInfo
	}{
		{"empty", DeviceInfo{}},
		{"no device_id", DeviceInfo{Platform: "ios", Model: "iPhone", OSVersion: "17", AppVersion: "1.0"}},
		{"no platform", DeviceInfo{DeviceID: "x", Model: "iPhone", OSVersion: "17", AppVersion: "1.0"}},
		{"no model", DeviceInfo{DeviceID: "x", Platform: "ios", OSVersion: "17", AppVersion: "1.0"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeviceInfo(tt.info)
			if err == nil {
				t.Error("should reject missing fields")
			}
		})
	}
}

func TestValidateDeviceInfo_WrongPlatform(t *testing.T) {
	info := testDeviceInfo()
	info.Platform = "android"

	err := ValidateDeviceInfo(info)
	if err == nil {
		t.Error("non-iOS platform should be rejected")
	}
}

func TestDeviceService_RegisterAndLookup(t *testing.T) {
	ds := NewDeviceService()
	info := testDeviceInfo()

	record := ds.RegisterDevice("user-1", info, "192.168.1.1", "cred-1")
	if record.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", record.UserID)
	}
	if record.FriendlyName != "iPhone 15 Pro" {
		t.Errorf("expected iPhone 15 Pro, got %s", record.FriendlyName)
	}

	// Lookup
	deviceHash := ComputeDeviceHash(info)
	if !ds.IsKnownDevice("user-1", deviceHash) {
		t.Error("registered device should be known")
	}
	if ds.IsKnownDevice("user-1", "unknown-hash") {
		t.Error("unknown device should not be known")
	}
}

func TestDeviceService_GetDevices(t *testing.T) {
	ds := NewDeviceService()

	info1 := testDeviceInfo()
	info2 := testDeviceInfo()
	info2.DeviceID = "device-2"
	info2.Model = "iPad Pro"

	ds.RegisterDevice("user-1", info1, "1.1.1.1", "cred-1")
	ds.RegisterDevice("user-1", info2, "2.2.2.2", "cred-2")

	devices := ds.GetDevices("user-1")
	if len(devices) != 2 {
		t.Errorf("expected 2 devices, got %d", len(devices))
	}
}

func TestDeviceService_RevokeDevice(t *testing.T) {
	ds := NewDeviceService()
	info := testDeviceInfo()

	record := ds.RegisterDevice("user-1", info, "1.1.1.1", "cred-1")

	ok := ds.RevokeDevice("user-1", record.ID)
	if !ok {
		t.Error("should revoke existing device")
	}

	if ds.DeviceCount("user-1") != 0 {
		t.Error("should have 0 devices after revoke")
	}
}

func TestDeviceService_RevokeNonexistent(t *testing.T) {
	ds := NewDeviceService()
	ok := ds.RevokeDevice("user-1", "nonexistent-id")
	if ok {
		t.Error("should not revoke nonexistent device")
	}
}

func TestDeviceService_UpdateActivity(t *testing.T) {
	ds := NewDeviceService()
	info := testDeviceInfo()

	ds.RegisterDevice("user-1", info, "1.1.1.1", "cred-1")

	deviceHash := ComputeDeviceHash(info)
	ds.UpdateDeviceActivity("user-1", deviceHash, "9.9.9.9")

	record := ds.GetDeviceByHash("user-1", deviceHash)
	if record.LastIP != "9.9.9.9" {
		t.Errorf("expected updated IP 9.9.9.9, got %s", record.LastIP)
	}
}
