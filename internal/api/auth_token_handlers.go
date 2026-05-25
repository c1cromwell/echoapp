package api

import (
	"encoding/json"
	"net/http"
	"time"
)

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
