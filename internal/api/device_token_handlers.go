package api

// WO-273: Device registration token flow.
//
// Primary device calls POST /identity/devices/token (authenticated) to
// obtain a short-lived token that is encoded into a QR code. The secondary
// device scans the QR code and calls POST /identity/devices with
// {token, new_public_key_hex} — no P-256 signature required, because the
// secondary device has not yet been registered.
//
// Token lifecycle:
//   - Created by: POST /identity/devices/token (requires passkey or JWT auth)
//   - Stored as:  Redis key "device_reg_token:{token}" → JSON blob, 5-min TTL
//   - Consumed by: POST /identity/devices (token path)
//   - Single-use: deleted from Redis on first successful consume
//   - Expiry: 5 minutes (300 seconds)

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

const deviceRegTokenTTL = 5 * time.Minute

// deviceRegTokenRecord is the value stored in Redis under the token key.
type deviceRegTokenRecord struct {
	IssuerDID string    `json:"issuer_did"`
	IssuedAt  time.Time `json:"issued_at"`
}

// DeviceTokenResponse is the JSON response for POST /identity/devices/token.
type DeviceTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

// DeviceTokenAddRequest is the body for the token-based add-device flow.
type DeviceTokenAddRequest struct {
	Token        string `json:"token"`
	NewPublicKey string `json:"new_public_key_hex"`
	DeviceLabel  string `json:"device_label,omitempty"`
}

// handleIdentityDeviceToken issues a 5-minute registration token (WO-273).
// The caller must be authenticated; the token encodes the caller's DID.
// POST /identity/devices/token
func (rt *Router) handleIdentityDeviceToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	issuerDID, _ := r.Context().Value(ContextKeyUserID).(string)
	if issuerDID == "" {
		WriteError(w, http.StatusUnauthorized, "AUTH_REQUIRED", "authentication required to issue device token", r.Header.Get("X-Request-ID"))
		return
	}

	if rt.Redis == nil {
		WriteError(w, http.StatusServiceUnavailable, "REDIS_NOT_CONFIGURED", "registration token service unavailable", r.Header.Get("X-Request-ID"))
		return
	}

	// Generate a cryptographically random 32-byte token.
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		WriteError(w, http.StatusInternalServerError, "TOKEN_GEN_FAILED", "failed to generate token", r.Header.Get("X-Request-ID"))
		return
	}
	token := hex.EncodeToString(raw)

	record, _ := json.Marshal(deviceRegTokenRecord{
		IssuerDID: issuerDID,
		IssuedAt:  time.Now().UTC(),
	})

	if err := rt.Redis.SetDeviceRegToken(r.Context(), token, record, deviceRegTokenTTL); err != nil {
		WriteError(w, http.StatusInternalServerError, "TOKEN_STORE_FAILED", "failed to store token", r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, DeviceTokenResponse{
		Token:     token,
		ExpiresIn: int(deviceRegTokenTTL.Seconds()),
	})
}

// handleIdentityListDevices returns all registered device keys for a DID.
// GET /identity/devices/{did}
func (rt *Router) handleIdentityListDevices(w http.ResponseWriter, r *http.Request, did string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if did == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_DID", "DID is required", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.DIDRegistry == nil {
		WriteError(w, http.StatusServiceUnavailable, "REGISTRY_NOT_CONFIGURED", "DID registry not initialized", r.Header.Get("X-Request-ID"))
		return
	}

	devices, err := rt.DIDRegistry.ListDevices(r.Context(), did)
	if err != nil {
		WriteError(w, http.StatusNotFound, "DID_NOT_FOUND", "DID not registered", r.Header.Get("X-Request-ID"))
		return
	}

	type deviceEntry struct {
		DID          string `json:"did"`
		PublicKeyHex string `json:"public_key_hex"`
		DeviceLabel  string `json:"device_label,omitempty"`
	}
	out := make([]deviceEntry, 0, len(devices))
	for _, b := range devices {
		out = append(out, deviceEntry{
			DID:          b.DID,
			PublicKeyHex: b.PublicKeyHex,
			DeviceLabel:  b.DeviceLabel,
		})
	}
	WriteJSON(w, http.StatusOK, out)
}

// tryTokenBasedDeviceAdd attempts the QR-code token path for add-device.
// Returns true if it handled the request (including errors), false if the
// request should fall through to the signature-based path.
func (rt *Router) tryTokenBasedDeviceAdd(w http.ResponseWriter, r *http.Request, rawBody []byte) bool {
	var req DeviceTokenAddRequest
	if err := json.Unmarshal(rawBody, &req); err != nil || req.Token == "" {
		return false // not a token-based request
	}
	if req.NewPublicKey == "" {
		return false
	}

	if rt.Redis == nil {
		WriteError(w, http.StatusServiceUnavailable, "REDIS_NOT_CONFIGURED", "token service unavailable", r.Header.Get("X-Request-ID"))
		return true
	}

	recordBytes, err := rt.Redis.GetDeviceRegToken(r.Context(), req.Token)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "INVALID_TOKEN", "registration token invalid or expired", r.Header.Get("X-Request-ID"))
		return true
	}

	var record deviceRegTokenRecord
	if err := json.Unmarshal(recordBytes, &record); err != nil {
		WriteError(w, http.StatusInternalServerError, "TOKEN_PARSE_FAILED", "malformed token record", r.Header.Get("X-Request-ID"))
		return true
	}

	// Single-use: delete before proceeding so concurrent requests can't replay it.
	_ = rt.Redis.DeleteDeviceRegToken(r.Context(), req.Token)

	deviceLabel := req.DeviceLabel
	if deviceLabel == "" {
		deviceLabel = "secondary-device"
	}

	if _, err := rt.DIDRegistry.RegisterAdditionalDevice(r.Context(), record.IssuerDID, req.NewPublicKey, deviceLabel); err != nil {
		WriteError(w, http.StatusConflict, "DEVICE_KEY_CONFLICT", err.Error(), r.Header.Get("X-Request-ID"))
		return true
	}

	if rt.Redis != nil {
		_ = rt.Redis.DeleteDIDDeviceKeys(context.Background(), record.IssuerDID)
	}

	WriteJSON(w, http.StatusCreated, map[string]string{
		"subject_did":  record.IssuerDID,
		"device_label": deviceLabel,
		"status":       "registered",
	})
	return true
}
