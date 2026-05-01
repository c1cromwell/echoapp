package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/didkey"
)

// DIDBinding records the canonical association of a did:key with its
// raw public-key hex representation and the time it was first registered.
type DIDBinding struct {
	DID          string    `json:"did"`
	PublicKeyHex string    `json:"public_key_hex"`
	RegisteredAt time.Time `json:"registered_at"`
}

// DIDRegistry is the storage contract used by POST /identity/register.
//
// Phase-1 (WO-230 + WO-278) ships an in-memory implementation suitable for
// the local testnet. The canonical Postgres schema for the persistent
// implementation is defined in migrations/007_did_registry.sql; the
// did_registry table mirrors the (DID, PublicKeyHex, RegisteredAt) tuple
// 1:1 so the same handler can be re-pointed at a pgx-backed registry by
// swapping the rt.DIDRegistry field — no contract change required.
type DIDRegistry interface {
	// Register stores a (did, publicKeyHex) binding. If the DID is already
	// bound to the same public key the operation is idempotent and returns
	// isNew=false. If the DID is bound to a different public key,
	// ErrDIDConflict is returned.
	Register(ctx context.Context, did, publicKeyHex string) (binding *DIDBinding, isNew bool, err error)

	// Lookup retrieves a previously registered binding.
	Lookup(ctx context.Context, did string) (*DIDBinding, error)
}

// ErrDIDConflict indicates a DID is already bound to a different public key.
var ErrDIDConflict = errors.New("did already registered with a different public key")

// ErrBindingNotFound indicates a Lookup for an unknown DID.
var ErrBindingNotFound = errors.New("did binding not found")

// MemoryDIDRegistry is the default in-memory DIDRegistry used by tests and
// the local testnet. Safe for concurrent use.
type MemoryDIDRegistry struct {
	mu       sync.RWMutex
	bindings map[string]*DIDBinding
}

// NewMemoryDIDRegistry constructs an empty in-memory registry.
func NewMemoryDIDRegistry() *MemoryDIDRegistry {
	return &MemoryDIDRegistry{bindings: make(map[string]*DIDBinding)}
}

// Register implements DIDRegistry.
func (r *MemoryDIDRegistry) Register(_ context.Context, did, publicKeyHex string) (*DIDBinding, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.bindings[did]; ok {
		if existing.PublicKeyHex == publicKeyHex {
			return existing, false, nil
		}
		return existing, false, ErrDIDConflict
	}

	b := &DIDBinding{
		DID:          did,
		PublicKeyHex: publicKeyHex,
		RegisteredAt: time.Now().UTC(),
	}
	r.bindings[did] = b
	return b, true, nil
}

// Lookup implements DIDRegistry.
func (r *MemoryDIDRegistry) Lookup(_ context.Context, did string) (*DIDBinding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.bindings[did]
	if !ok {
		return nil, ErrBindingNotFound
	}
	return b, nil
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

// handleIdentityRegister implements POST /identity/register, the entry point
// referenced by Step 2 of scripts/validate-phase1.sh and by WO-230's
// "In Scope" item #5.
//
// Contract:
//   - Body: { "did": "did:key:z…", "public_key_hex": "<hex>" }
//   - The handler re-derives the canonical did:key from public_key_hex and
//     refuses any request where the supplied DID does not match — this
//     prevents callers from registering a DID for a key they do not control
//     at the *encoding* level. Possession of the private key is proven later
//     during passkey signing (handleRegisterPasskey) and for every
//     subsequent authenticated request.
//   - Idempotent: re-registering an existing (did, public_key_hex) pair
//     returns 200 with existing=true.
//   - Conflict: registering a DID already bound to a different key returns
//     409 DID_ALREADY_REGISTERED.
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

// handleIdentityResolve implements GET /identity/{did} for clients that want
// to confirm a binding exists and read back the registered public key.
// Public, no auth — the same key is embedded in the DID itself, so this is
// just a convenience read.
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

	binding, err := rt.DIDRegistry.Lookup(r.Context(), did)
	if err != nil {
		if errors.Is(err, ErrBindingNotFound) {
			WriteError(w, http.StatusNotFound, "DID_NOT_REGISTERED", "did has not been registered", r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusInternalServerError, "REGISTRY_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusOK, IdentityRegisterResponse{
		DID:          binding.DID,
		PublicKeyHex: binding.PublicKeyHex,
		RegisteredAt: binding.RegisteredAt.Format(time.RFC3339Nano),
		Existing:     true,
	})
}
