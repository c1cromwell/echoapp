package api

// stepup_login_handlers.go
// Phase 1 authentication endpoints needed by GlacialLoginScreen + StepUpSheetView.
//
// POST /v1/auth/login/challenge
//   Public — returns a 32-byte random nonce for the iOS PasskeyManager to sign.
//   Optional for Phase 1: iOS step-up uses X-Sender-DID + X-Signature directly,
//   but the challenge endpoint is available for WebAuthn-style flows.
//
// POST /v1/auth/step-up
//   Authenticated (X-Sender-DID + X-Signature, validated by authMiddleware).
//   Issues a short-lived elevated JWT tied to a specific action.
//   The backend's existing StepUpService enforces per-action TTL and method rules.

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/auth"
)

// handleLoginChallenge handles POST /v1/auth/login/challenge.
// Returns a random base64-encoded nonce for the client to sign.
func (rt *Router) handleLoginChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	// Generate 32-byte random nonce
	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		WriteError(w, http.StatusInternalServerError, "CHALLENGE_GEN", "could not generate challenge", r.Header.Get("X-Request-ID"))
		return
	}
	challengeB64 := base64.StdEncoding.EncodeToString(nonce)

	// Store in Redis if available (90-second TTL) for later verification
	if rt.Redis != nil {
		_ = rt.Redis.CacheSet(r.Context(), "challenge:"+challengeB64[:16], nonce, 90*time.Second)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"challenge":  challengeB64,
		"timeout":    60,
		"rp_id":     "echo.app",
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

// handleStepUp handles POST /v1/auth/step-up.
// Requires a valid authenticated request (X-Sender-DID + X-Signature enforced by authMiddleware).
// Issues an elevated JWT for the requested action.
func (rt *Router) handleStepUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	// DID is guaranteed to be present — authMiddleware already validated the signature.
	did, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "READ_BODY", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	req.Action = strings.TrimSpace(req.Action)
	if req.Action == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_ACTION", "action is required", r.Header.Get("X-Request-ID"))
		return
	}

	// Validate action via StepUpService
	stepUpSvc := auth.NewStepUpService()
	if !stepUpSvc.RequiresStepUp(req.Action) {
		WriteError(w, http.StatusBadRequest, "UNKNOWN_ACTION", "action is not a valid step-up action", r.Header.Get("X-Request-ID"))
		return
	}

	// Determine TTL from the action requirement
	ttl := 5 * time.Minute
	if req := stepUpSvc.GetRequirement(auth.StepUpAction(req.Action)); req != nil && req.TokenTTL > 0 {
		ttl = req.TokenTTL
	}

	// Issue elevated token
	elevatedToken, claims, err := rt.tokenService.IssueElevatedToken(did, "", 0, req.Action, ttl)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "TOKEN_ISSUE", "could not issue elevated token", r.Header.Get("X-Request-ID"))
		return
	}

	expiresAt := time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"elevated_token": elevatedToken,
		"expires_at":     expiresAt,
		"action":         req.Action,
		"request_id":     r.Header.Get("X-Request-ID"),
	})
}
