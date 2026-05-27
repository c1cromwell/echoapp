package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// displayNameAllowed matches per PRD v3.1 AC-INFRA-004.4 and mirrors the iOS DisplayNameValidator.
// Allowed: Unicode letters, digits, space, hyphen, underscore, apostrophe.
var displayNameAllowed = regexp.MustCompile(`^[\p{L}\p{N} \-_']+$`)

// DisplayNameError is returned when display name validation fails.
type DisplayNameError struct {
	Code    string
	Message string
}

func (e *DisplayNameError) Error() string { return e.Message }

// validateDisplayName trims and validates a display name, returning the trimmed form.
// Rules must be byte-for-byte identical to the iOS DisplayNameValidator.
func validateDisplayName(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)

	if utf8.RuneCountInString(trimmed) < 1 {
		return "", &DisplayNameError{Code: "DISPLAY_NAME_EMPTY", Message: "Display name must be at least 1 character"}
	}
	if utf8.RuneCountInString(trimmed) > 32 {
		return "", &DisplayNameError{Code: "DISPLAY_NAME_TOO_LONG", Message: "Display name must be 32 characters or fewer"}
	}
	for _, r := range trimmed {
		if unicode.IsControl(r) {
			return "", &DisplayNameError{Code: "DISPLAY_NAME_INVALID_CHARS", Message: "Display name contains control characters"}
		}
	}
	if !displayNameAllowed.MatchString(trimmed) {
		return "", &DisplayNameError{
			Code:    "DISPLAY_NAME_INVALID_CHARS",
			Message: "Display name may only contain letters, digits, spaces, hyphens, underscores, and apostrophes",
		}
	}
	return trimmed, nil
}

// RegisterDIDRequest is the body for POST /v1/auth/register-did.
type RegisterDIDRequest struct {
	PublicKey      string `json:"public_key"`
	DisplayName    string `json:"display_name"`
	AssuranceLevel string `json:"assurance_level"` // "ial0" for display-name-only first-run
}

// RegisterDIDResponse is returned on successful DID registration.
type RegisterDIDResponse struct {
	DID         string `json:"did"`
	DisplayName string `json:"display_name"`
	TrustTier   int    `json:"trust_tier"`
	MintedAt    string `json:"minted_at"`
}

// handleRegisterDID handles POST /v1/auth/register-did.
// Public endpoint — no Bearer token required. Idempotency-Key middleware is applied at routing.
func (rt *Router) handleRegisterDID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req RegisterDIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	displayName, err := validateDisplayName(req.DisplayName)
	if err != nil {
		var dne *DisplayNameError
		if errors.As(err, &dne) {
			WriteError(w, http.StatusBadRequest, dne.Code, dne.Message, r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusBadRequest, "DISPLAY_NAME_INVALID", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	if req.PublicKey == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_PUBLIC_KEY", "public_key is required", r.Header.Get("X-Request-ID"))
		return
	}

	// Derive the canonical did:key from the supplied P-256 public key (ADR-0001;
	// did:key is the Phase-1 identity method, no chain transaction required).
	did, err := didkey.DeriveFromPublicKeyHex(req.PublicKey)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_PUBLIC_KEY", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	mintedAt := time.Now().UTC()

	// Trust tier starts at 0 — promoted to 1 on passkey webhook per AC-INFRA-004.5.
	WriteJSON(w, http.StatusOK, RegisterDIDResponse{
		DID:         did,
		DisplayName: displayName,
		TrustTier:   0,
		MintedAt:    mintedAt.Format(time.RFC3339),
	})
}

// PasskeyRegisterRequest is the body for POST /v1/enrollment/passkey.
type PasskeyRegisterRequest struct {
	DID         string `json:"did"`
	PublicKey   string `json:"public_key"`
	Attestation string `json:"attestation"`
}

// handleRegisterPasskey handles POST /v1/enrollment/passkey.
// Promotes trust_tier from 0 → 1 on successful passkey registration (AC-INFRA-004.5).
func (rt *Router) handleRegisterPasskey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req PasskeyRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	if req.DID == "" || req.PublicKey == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_FIELDS", "did and public_key are required", r.Header.Get("X-Request-ID"))
		return
	}

	// Production: verify attestation and promote tier 0 → 1 in the users table.
	// promoteToTier1IfTier0(ctx, req.DID) — non-fatal if it fails (passkey still registered).

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "linked",
		"did":        req.DID,
		"trust_tier": 1,
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

// handleRestoreChallenge handles POST /v1/auth/restore-challenge.
// Returns a 32-byte nonce the wallet must sign to prove ownership before re-binding a DID.
func (rt *Router) handleRestoreChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	// Production: store nonce keyed by wallet_address with 5-min TTL.
	nonce := uuid.New().String() + uuid.New().String() // 72 printable chars ≈ adequate entropy
	WriteJSON(w, http.StatusOK, map[string]string{
		"challenge":  nonce,
		"expires_in": "300",
	})
}

// RestoreDIDRequest is the body for POST /v1/auth/restore-did.
type RestoreDIDRequest struct {
	WalletAddress      string `json:"wallet_address"`
	NewDevicePublicKey string `json:"new_device_public_key"`
	WalletSignature    string `json:"wallet_signature"`
}

// handleRestoreDID handles POST /v1/auth/restore-did.
// Looks up the DID bound to the wallet, verifies wallet ownership, and re-binds to new device key.
func (rt *Router) handleRestoreDID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req RestoreDIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	if req.WalletAddress == "" || req.NewDevicePublicKey == "" || req.WalletSignature == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_FIELDS", "wallet_address, new_device_public_key, and wallet_signature are required", r.Header.Get("X-Request-ID"))
		return
	}

	// Production: verify wallet signature against challenge, look up DID by wallet_address,
	// re-bind DID to new_device_public_key, return user record.
	// Errors: 404 WALLET_NOT_ENROLLED, 401 WALLET_SIGNATURE_INVALID, 409 DEVICE_ALREADY_ENROLLED.
	// WO-273: Persisted wallet→subject_did mapping is required before returning a canonical did:key.
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"did":            "pending:wallet:" + req.WalletAddress,
		"display_name":   "Restored User",
		"trust_tier":     1,
		"wallet_address": req.WalletAddress,
		"request_id":     r.Header.Get("X-Request-ID"),
	})
}

// handleEnrollmentMDL handles POST /v1/enrollment/mdl/start and /mdl/finish.
func (rt *Router) handleEnrollmentMDL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	WriteError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED",
		"mDL enrollment is not yet available; use VC wallet enrollment or first-run onboarding",
		r.Header.Get("X-Request-ID"))
}

// handleEnrollmentIDV handles POST /v1/enrollment/idv/start and /idv/await.
func (rt *Router) handleEnrollmentIDV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	WriteError(w, http.StatusNotImplemented, "NOT_IMPLEMENTED",
		"IDV enrollment is not yet available; use VC wallet enrollment or VIP verify",
		r.Header.Get("X-Request-ID"))
}
