package api

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/credentials"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// DIDBinding records a device key registered for a subject did:key.
type DIDBinding struct {
	DID          string    `json:"did"`
	PublicKeyHex string    `json:"public_key_hex"`
	DeviceLabel  string    `json:"device_label,omitempty"`
	RegisteredAt time.Time `json:"-"`
}

// DIDRegistry is the storage contract for POST /identity/register and
// multi-device flows (WO-273).
type DIDRegistry interface {
	Register(ctx context.Context, did, publicKeyHex string) (binding *DIDBinding, isNew bool, err error)
	Lookup(ctx context.Context, did string) (*DIDBinding, error)
	ListDevices(ctx context.Context, did string) ([]*DIDBinding, error)
	RegisterAdditionalDevice(ctx context.Context, subjectDID, newPublicKeyHex, deviceLabel string) (*DIDBinding, error)
}

// ErrDIDConflict indicates Register was called for a DID that already has
// device keys registered, but none match the supplied public key (use
// POST /identity/devices for additional keys).
var ErrDIDConflict = errors.New("did already registered with a different public key")

// ErrBindingNotFound indicates a Lookup/ListDevices for an unknown DID.
var ErrBindingNotFound = errors.New("did binding not found")

// ErrDuplicateDeviceKey indicates the device public key is already present.
var ErrDuplicateDeviceKey = errors.New("device public key already registered for this did")

// ErrSigningDeviceNotRegistered indicates signing_did does not match any key on file.
var ErrSigningDeviceNotRegistered = errors.New("signing_did does not match a registered device for subject_did")

// MemoryDIDRegistry is the default in-memory DIDRegistry used by tests and
// the local testnet. Safe for concurrent use.
type MemoryDIDRegistry struct {
	mu    sync.RWMutex
	byDID map[string][]*DIDBinding
}

// NewMemoryDIDRegistry constructs an empty in-memory registry.
func NewMemoryDIDRegistry() *MemoryDIDRegistry {
	return &MemoryDIDRegistry{byDID: make(map[string][]*DIDBinding)}
}

func cloneBinding(b *DIDBinding) *DIDBinding {
	if b == nil {
		return nil
	}
	c := *b
	return &c
}

// Register implements DIDRegistry.
func (r *MemoryDIDRegistry) Register(_ context.Context, did, publicKeyHex string) (*DIDBinding, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	list := r.byDID[did]
	for _, b := range list {
		if b.PublicKeyHex == publicKeyHex {
			return cloneBinding(b), false, nil
		}
	}
	if len(list) > 0 {
		return nil, false, ErrDIDConflict
	}
	b := &DIDBinding{
		DID:          did,
		PublicKeyHex: publicKeyHex,
		DeviceLabel:  "primary",
		RegisteredAt: time.Now().UTC(),
	}
	r.byDID[did] = []*DIDBinding{b}
	return cloneBinding(b), true, nil
}

// Lookup returns the oldest device row (primary) for compatibility.
func (r *MemoryDIDRegistry) Lookup(_ context.Context, did string) (*DIDBinding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.byDID[did]
	if len(list) == 0 {
		return nil, ErrBindingNotFound
	}
	return cloneBinding(list[0]), nil
}

// ListDevices implements DIDRegistry.
func (r *MemoryDIDRegistry) ListDevices(_ context.Context, did string) ([]*DIDBinding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.byDID[did]
	if len(list) == 0 {
		return nil, ErrBindingNotFound
	}
	out := make([]*DIDBinding, len(list))
	for i, b := range list {
		out[i] = cloneBinding(b)
	}
	return out, nil
}

// RegisterAdditionalDevice implements DIDRegistry.
func (r *MemoryDIDRegistry) RegisterAdditionalDevice(_ context.Context, subjectDID, newPublicKeyHex, deviceLabel string) (*DIDBinding, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	list := r.byDID[subjectDID]
	if len(list) == 0 {
		return nil, ErrBindingNotFound
	}
	for _, b := range list {
		if b.PublicKeyHex == newPublicKeyHex {
			return nil, ErrDuplicateDeviceKey
		}
	}
	b := &DIDBinding{
		DID:          subjectDID,
		PublicKeyHex: newPublicKeyHex,
		DeviceLabel:  deviceLabel,
		RegisteredAt: time.Now().UTC(),
	}
	r.byDID[subjectDID] = append(list, b)
	return cloneBinding(b), nil
}

// IdentityRegisterRequest is the body for POST /identity/register.
type IdentityRegisterRequest struct {
	DID          string `json:"did"`
	PublicKeyHex string `json:"public_key_hex"`
}

// IdentityRegisterResponse is returned on successful registration.
type IdentityRegisterResponse struct {
	DID          string `json:"did"`
	PublicKeyHex string `json:"public_key_hex"`
	RegisteredAt string `json:"registered_at"`
	Existing     bool   `json:"existing"`
}

// IdentityDeviceDTO is one device row in a resolve response.
type IdentityDeviceDTO struct {
	PublicKeyHex string `json:"public_key_hex"`
	DeviceLabel  string `json:"device_label"`
	RegisteredAt string `json:"registered_at"`
}

// IdentityResolveResponse is returned by GET /identity/resolve/{did} and
// GET /identity/{did} (WO-273).
type IdentityResolveResponse struct {
	DID     string              `json:"did"`
	Devices []IdentityDeviceDTO `json:"devices"`
}

// IdentityAddDeviceRequest is the JSON body for POST /identity/devices.
type IdentityAddDeviceRequest struct {
	SubjectDID      string `json:"subject_did"`
	NewPublicKeyHex string `json:"new_public_key_hex"`
	DeviceLabel     string `json:"device_label"`
	SigningDID      string `json:"signing_did"`
}

const identitySignatureHeader = "X-Identity-Signature"

func writeIdentityResolve(w http.ResponseWriter, did string, devices []*DIDBinding) {
	dtos := make([]IdentityDeviceDTO, 0, len(devices))
	for _, b := range devices {
		dtos = append(dtos, IdentityDeviceDTO{
			PublicKeyHex: b.PublicKeyHex,
			DeviceLabel:  b.DeviceLabel,
			RegisteredAt: b.RegisteredAt.UTC().Format(time.RFC3339Nano),
		})
	}
	WriteJSON(w, http.StatusOK, IdentityResolveResponse{DID: did, Devices: dtos})
}

// handleIdentityRegister implements POST /identity/register.
func (rt *Router) handleIdentityRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	if rt.DIDRegistry == nil {
		WriteError(w, http.StatusServiceUnavailable, "REGISTRY_NOT_CONFIGURED", "DID registry not initialized", r.Header.Get("X-Request-ID"))
		return
	}

	var req IdentityRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	if req.DID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_DID", "did is required", r.Header.Get("X-Request-ID"))
		return
	}
	if req.PublicKeyHex == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_PUBLIC_KEY", "public_key_hex is required", r.Header.Get("X-Request-ID"))
		return
	}

	canonical, err := didkey.DeriveFromPublicKeyHex(req.PublicKeyHex)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_PUBLIC_KEY", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if canonical != req.DID {
		WriteError(w, http.StatusBadRequest, "DID_KEY_MISMATCH",
			"supplied did does not match canonical derivation from public_key_hex",
			r.Header.Get("X-Request-ID"))
		return
	}

	binding, isNew, err := rt.DIDRegistry.Register(r.Context(), req.DID, req.PublicKeyHex)
	if err != nil {
		switch {
		case errors.Is(err, ErrDIDConflict):
			WriteError(w, http.StatusConflict, "DID_ALREADY_REGISTERED",
				"did is already registered with a different public key",
				r.Header.Get("X-Request-ID"))
		default:
			WriteError(w, http.StatusInternalServerError, "REGISTRY_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		}
		return
	}

	status := http.StatusCreated
	if !isNew {
		status = http.StatusOK
	}
	WriteJSON(w, status, IdentityRegisterResponse{
		DID:          binding.DID,
		PublicKeyHex: binding.PublicKeyHex,
		RegisteredAt: binding.RegisteredAt.Format(time.RFC3339Nano),
		Existing:     !isNew,
	})
}

// handleIdentityResolve returns all device public keys for a subject DID.
func (rt *Router) handleIdentityResolve(w http.ResponseWriter, r *http.Request, did string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.DIDRegistry == nil {
		WriteError(w, http.StatusServiceUnavailable, "REGISTRY_NOT_CONFIGURED", "DID registry not initialized", r.Header.Get("X-Request-ID"))
		return
	}
	if did == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_DID", "did path segment required", r.Header.Get("X-Request-ID"))
		return
	}

	devices, err := rt.DIDRegistry.ListDevices(r.Context(), did)
	if err != nil {
		if errors.Is(err, ErrBindingNotFound) {
			WriteError(w, http.StatusNotFound, "DID_NOT_REGISTERED", "did has not been registered", r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusInternalServerError, "REGISTRY_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	writeIdentityResolve(w, did, devices)
}

// handleIdentityAddDevice implements POST /identity/devices (WO-273).
// The caller signs the exact HTTP body bytes with the private key that matches
// signing_did; the signature is sent as hex in X-Identity-Signature (raw P1363 or ASN.1 DER).
func (rt *Router) handleIdentityAddDevice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.DIDRegistry == nil {
		WriteError(w, http.StatusServiceUnavailable, "REGISTRY_NOT_CONFIGURED", "DID registry not initialized", r.Header.Get("X-Request-ID"))
		return
	}

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "READ_BODY", "failed to read body", r.Header.Get("X-Request-ID"))
		return
	}

	sigHex := r.Header.Get(identitySignatureHeader)
	if sigHex == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_SIGNATURE", identitySignatureHeader+" header required", r.Header.Get("X-Request-ID"))
		return
	}
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_SIGNATURE_HEX", "signature must be lowercase hex", r.Header.Get("X-Request-ID"))
		return
	}

	var req IdentityAddDeviceRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	if req.SubjectDID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_SUBJECT_DID", "subject_did is required", r.Header.Get("X-Request-ID"))
		return
	}
	if req.NewPublicKeyHex == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_NEW_PUBLIC_KEY", "new_public_key_hex is required", r.Header.Get("X-Request-ID"))
		return
	}
	if req.SigningDID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_SIGNING_DID", "signing_did is required", r.Header.Get("X-Request-ID"))
		return
	}
	deviceLabel := req.DeviceLabel
	if deviceLabel == "" {
		deviceLabel = "device"
	}

	if _, err := didkey.DeriveFromPublicKeyHex(req.NewPublicKeyHex); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_PUBLIC_KEY", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	devices, err := rt.DIDRegistry.ListDevices(r.Context(), req.SubjectDID)
	if err != nil {
		if errors.Is(err, ErrBindingNotFound) {
			WriteError(w, http.StatusNotFound, "DID_NOT_REGISTERED", "subject_did has not been registered", r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusInternalServerError, "REGISTRY_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	var signerPub *ecdsa.PublicKey
	for _, b := range devices {
		deviceDID, derErr := didkey.DeriveFromPublicKeyHex(b.PublicKeyHex)
		if derErr != nil {
			continue
		}
		if deviceDID != req.SigningDID {
			continue
		}
		pub, parseErr := didkey.Parse(req.SigningDID)
		if parseErr != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_SIGNING_DID", parseErr.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		signerPub = pub
		break
	}
	if signerPub == nil {
		WriteError(w, http.StatusForbidden, "SIGNING_DEVICE_UNKNOWN", ErrSigningDeviceNotRegistered.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	if err := didkey.VerifyECDSAP256SHA256(signerPub, rawBody, sig); err != nil {
		WriteError(w, http.StatusForbidden, "INVALID_SIGNATURE", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	if _, err := rt.DIDRegistry.RegisterAdditionalDevice(r.Context(), req.SubjectDID, req.NewPublicKeyHex, deviceLabel); err != nil {
		switch {
		case errors.Is(err, ErrDuplicateDeviceKey):
			WriteError(w, http.StatusConflict, "DEVICE_KEY_CONFLICT", err.Error(), r.Header.Get("X-Request-ID"))
		default:
			WriteError(w, http.StatusInternalServerError, "REGISTRY_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		}
		return
	}

	// Invalidate the passkey auth Redis cache so the new device key is picked up immediately.
	if rt.Redis != nil {
		_ = rt.Redis.DeleteDIDDeviceKeys(r.Context(), req.SubjectDID)
	}

	if rt.CredentialService != nil {
		if deviceDID, derr := didkey.DeriveFromPublicKeyHex(req.NewPublicKeyHex); derr == nil {
			_, _ = rt.CredentialService.IssueCredential(r.Context(), &credentials.CredentialIssuanceRequest{
				SubjectDID:     deviceDID,
				CredentialType: credentials.DeviceAttestationCredential,
				Claims: map[string]interface{}{
					"controller":  req.SubjectDID,
					"deviceLabel": deviceLabel,
					"attestedAt":  time.Now().UTC().Format(time.RFC3339Nano),
				},
				PreferredFormat: credentials.JSONLDFormat,
			})
		}
	}

	updated, err := rt.DIDRegistry.ListDevices(r.Context(), req.SubjectDID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "REGISTRY_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	writeIdentityResolve(w, req.SubjectDID, updated)
}
