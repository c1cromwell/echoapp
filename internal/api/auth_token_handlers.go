package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/auth"
)

// handleAuthSession implements POST /v1/auth/session.
//
// It exchanges a passkey-signed request for a session: a short-lived access
// token (what the WebSocket upgrade and Bearer-only routes need) plus a
// rotating refresh token. Without it a newly enrolled device has a registered
// key but no way to obtain a token, so it can never open the realtime socket;
// only the recovery flow issued one, and that one expires after 15 minutes.
//
// The request MUST be authenticated with X-Sender-DID + X-Signature (verified
// by the auth middleware against the DID's registered device keys). A Bearer
// token is deliberately not accepted here: an access token must not be able to
// mint its own successors.
func (rt *Router) handleAuthSession(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", reqID)
		return
	}
	if rt.tokenService == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKEN_SERVICE_UNAVAILABLE", "token service not configured", reqID)
		return
	}
	did, _ := r.Context().Value(ContextKeyUserID).(string)
	if did == "" || r.Header.Get(headerSenderDID) != did {
		WriteError(w, http.StatusUnauthorized, "PASSKEY_REQUIRED", "a passkey-signed request is required to open a session", reqID)
		return
	}

	var req struct {
		DeviceID string `json:"device_id"`
	}
	if body, err := io.ReadAll(io.LimitReader(r.Body, 4096)); err == nil && len(body) > 0 {
		_ = json.Unmarshal(body, &req)
	}
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.DeviceID == "" || len(req.DeviceID) > 128 {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "device_id is required", reqID)
		return
	}

	trustTier := 0
	if rt.V3 != nil && rt.V3.DB != nil {
		if user, err := rt.V3.DB.GetUserByDID(r.Context(), did); err == nil && user != nil {
			trustTier = user.TrustTier
		}
	}

	access, claims, err := rt.tokenService.IssueAccessToken(did, req.DeviceID, trustTier, "messaging")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "could not issue access token", reqID)
		return
	}
	refresh := auth.GenerateRefreshToken()
	rt.tokenService.StoreRefreshToken(did, refresh, req.DeviceID)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"did":           did,
		"access_token":  access,
		"refresh_token": refresh,
		"token_type":    "Bearer",
		"expires_in":    int(time.Until(time.Unix(claims.ExpiresAt, 0)).Seconds()),
		"request_id":    reqID,
	})
}

// handleAuthRefresh implements POST /v3/auth/refresh (Wave A.2).
//
// The opaque refresh token is itself the credential, so this endpoint is public.
// It rotates the refresh token (single-use, with reuse detection that revokes the
// whole family) and mints a fresh access token.
func (rt *Router) handleAuthRefresh(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", reqID)
		return
	}
	if rt.tokenService == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKEN_SERVICE_UNAVAILABLE", "token service not configured", reqID)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
		DeviceID     string `json:"device_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" || req.DeviceID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "refresh_token and device_id are required", reqID)
		return
	}

	newRefresh, record, aerr := rt.tokenService.RotateRefreshToken(req.RefreshToken, req.DeviceID)
	if aerr != nil {
		WriteError(w, aerr.HTTPStatus, string(aerr.Code), aerr.Message, reqID)
		return
	}

	access, claims, err := rt.tokenService.IssueAccessToken(record.UserID, req.DeviceID, 0, "messaging")
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "TOKEN_ISSUE_FAILED", "could not issue access token", reqID)
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  access,
		"refresh_token": newRefresh,
		"token_type":    "Bearer",
		"expires_in":    int(time.Until(time.Unix(claims.ExpiresAt, 0)).Seconds()),
		"request_id":    reqID,
	})
}

// handleAuthRevoke implements POST /v3/auth/revoke (Wave A.2).
//
// Authenticated ("logout"): revokes all of the caller's refresh tokens so none
// can be rotated again. Access tokens are short-lived (15m); callers needing
// immediate access-token revocation should also rely on the durable blocklist.
func (rt *Router) handleAuthRevoke(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", reqID)
		return
	}
	if rt.tokenService == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKEN_SERVICE_UNAVAILABLE", "token service not configured", reqID)
		return
	}

	userID, _ := r.Context().Value(ContextKeyUserID).(string)
	if userID == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required to revoke", reqID)
		return
	}

	count := rt.tokenService.RevokeAllUserTokens(userID)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"revoked":                true,
		"refresh_tokens_revoked": count,
		"request_id":             reqID,
	})
}
