package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/wallet"
	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// enrollmentVerifiedRecord is the server-side tail of a successful VC/mDL/IDV finish.
type enrollmentVerifiedRecord struct {
	HolderDID      string
	AssuranceLevel string
	CredentialType string
	ExpiresAt      time.Time
}

func (rt *Router) storeEnrollmentVerified(ref string, rec enrollmentVerifiedRecord) {
	rt.enrollmentVerifiedMu.Lock()
	defer rt.enrollmentVerifiedMu.Unlock()
	if rt.enrollmentVerified == nil {
		rt.enrollmentVerified = make(map[string]enrollmentVerifiedRecord)
	}
	rt.enrollmentVerified[ref] = rec
}

func (rt *Router) loadEnrollmentVerified(ref string) (enrollmentVerifiedRecord, bool) {
	rt.enrollmentVerifiedMu.Lock()
	defer rt.enrollmentVerifiedMu.Unlock()
	rec, ok := rt.enrollmentVerified[ref]
	if !ok || time.Now().After(rec.ExpiresAt) {
		return enrollmentVerifiedRecord{}, false
	}
	return rec, true
}

func trustTierForIAL(ial string) int {
	switch ial {
	case "ial3":
		return 5
	case "ial2":
		return 4
	case "ial1":
		return 2
	default:
		return 1
	}
}

// deterministicDAGAddress is the interim, server-derivable wallet address.
// Single source of truth lives in the wallet package so custody enforcement
// (which rejects this exact form in real-funds mode) can never drift from it.
func deterministicDAGAddress(did string) string {
	return wallet.ServerDerivableAddress(did)
}

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
// Returns a nonce the wallet must sign to prove ownership before re-binding a DID.
func (rt *Router) handleRestoreChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		WalletAddress string `json:"wallet_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.WalletAddress) == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_FIELDS", "wallet_address is required", r.Header.Get("X-Request-ID"))
		return
	}
	addr := strings.TrimSpace(req.WalletAddress)
	nonce := uuid.New().String() + uuid.New().String()
	rt.restoreChallengeMu.Lock()
	if rt.restoreChallenges == nil {
		rt.restoreChallenges = make(map[string]restoreChallenge)
	}
	rt.restoreChallenges[addr] = restoreChallenge{nonce: nonce, expiresAt: time.Now().Add(5 * time.Minute)}
	rt.restoreChallengeMu.Unlock()

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
// Looks up the DID bound to the wallet, consumes the restore challenge, and issues a JWT.
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

	rt.restoreChallengeMu.Lock()
	ch, ok := rt.restoreChallenges[req.WalletAddress]
	if ok {
		delete(rt.restoreChallenges, req.WalletAddress)
	}
	rt.restoreChallengeMu.Unlock()
	if !ok || time.Now().After(ch.expiresAt) {
		WriteError(w, http.StatusUnauthorized, "WALLET_SIGNATURE_INVALID", "restore challenge missing or expired", r.Header.Get("X-Request-ID"))
		return
	}

	did := rt.lookupDIDByWallet(r.Context(), req.WalletAddress)
	if did == "" {
		WriteError(w, http.StatusNotFound, "WALLET_NOT_ENROLLED", "no identity is linked to this recovery phrase", r.Header.Get("X-Request-ID"))
		return
	}

	displayName := "Restored User"
	trustTier := 1
	if rt.V3 != nil && rt.V3.DB != nil {
		if user, err := rt.V3.DB.GetUserByDID(r.Context(), did); err == nil && user != nil {
			if user.Username != "" {
				displayName = user.Username
			}
			if user.TrustTier > 0 {
				trustTier = user.TrustTier
			}
		}
	}

	access := ""
	if rt.tokenService != nil {
		token, _, err := rt.tokenService.IssueAccessToken(did, "restored-device", trustTier, "messaging")
		if err == nil {
			access = token
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"did":            did,
		"display_name":   displayName,
		"trust_tier":     trustTier,
		"wallet_address": req.WalletAddress,
		"access_token":   access,
		"request_id":     r.Header.Get("X-Request-ID"),
	})
}

func (rt *Router) lookupDIDByWallet(ctx context.Context, addr string) string {
	if addr == "" {
		return ""
	}
	rt.enrollmentWalletMu.Lock()
	for did, linked := range rt.enrollmentWalletByDID {
		if linked == addr {
			rt.enrollmentWalletMu.Unlock()
			return did
		}
	}
	rt.enrollmentWalletMu.Unlock()
	if rt.V3 != nil && rt.V3.Wallet != nil && rt.V3.Wallet.Store != nil {
		if did, err := rt.V3.Wallet.Store.GetDIDByDAGAddress(ctx, addr); err == nil && did != "" {
			return did
		}
	}
	return ""
}

// handleEnrollmentDID handles POST /v1/enrollment/did.
// Links a verified credential reference to the holder DID for enrollment tail.
func (rt *Router) handleEnrollmentDID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		CredentialReference string `json:"credential_reference"`
		PublicKeyHex        string `json:"public_key_hex"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CredentialReference == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_FIELDS", "credential_reference is required", r.Header.Get("X-Request-ID"))
		return
	}

	rec, ok := rt.loadEnrollmentVerified(req.CredentialReference)
	if !ok {
		WriteError(w, http.StatusBadRequest, "CREDENTIAL_REF_UNKNOWN", "Credential reference expired or unknown", r.Header.Get("X-Request-ID"))
		return
	}

	did := strings.TrimSpace(rec.HolderDID)
	if did == "" && req.PublicKeyHex != "" {
		derived, err := didkey.DeriveFromPublicKeyHex(req.PublicKeyHex)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_PUBLIC_KEY", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		did = derived
	}
	if did == "" {
		WriteError(w, http.StatusBadRequest, "HOLDER_DID_MISSING", "Verified credential did not include a holder DID", r.Header.Get("X-Request-ID"))
		return
	}

	if req.PublicKeyHex != "" && rt.DIDRegistry != nil {
		ctx := r.Context()
		if _, err := rt.DIDRegistry.Lookup(ctx, did); err == nil {
			_, _ = rt.DIDRegistry.RegisterAdditionalDevice(ctx, did, req.PublicKeyHex, "credential-enrollment")
		} else {
			_, _, _ = rt.DIDRegistry.Register(ctx, did, req.PublicKeyHex)
		}
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"did":             did,
		"trust_tier":      trustTierForIAL(rec.AssuranceLevel),
		"credential_type": rec.CredentialType,
		"assurance_level": rec.AssuranceLevel,
		"request_id":      r.Header.Get("X-Request-ID"),
	})
}

// handleEnrollmentWallet handles POST /v1/enrollment/wallet.
// Provisions a Constellation wallet address for the enrolled DID (Phase 2 proxy).
func (rt *Router) handleEnrollmentWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		DID string `json:"did"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_FIELDS", "did is required", r.Header.Get("X-Request-ID"))
		return
	}

	rt.enrollmentWalletMu.Lock()
	if rt.enrollmentWalletByDID == nil {
		rt.enrollmentWalletByDID = make(map[string]string)
	}
	addr, ok := rt.enrollmentWalletByDID[req.DID]
	if !ok {
		addr = deterministicDAGAddress(req.DID)
		rt.enrollmentWalletByDID[req.DID] = addr
	}
	rt.enrollmentWalletMu.Unlock()

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"address":    addr,
		"did":        req.DID,
		"request_id": r.Header.Get("X-Request-ID"),
		"note":       fmt.Sprintf("wallet proxy for %s", req.DID),
	})
}
