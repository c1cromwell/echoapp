package api

import (
	"encoding/json"
	"net/http"
)

// handleDeleteAccount implements DELETE /v1/users/account (WO-228).
// Revokes refresh tokens for the authenticated user; durable identity revocation
// on the Identity Metagraph is a follow-on (Phase 2 Wave B).
func (rt *Router) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodDelete {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only DELETE is allowed", reqID)
		return
	}

	userID, _ := r.Context().Value(ContextKeyUserID).(string)
	if userID == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", reqID)
		return
	}

	revoked := 0
	if rt.tokenService != nil {
		revoked = rt.tokenService.RevokeAllUserTokens(userID)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"deleted":                true,
		"did":                    userID,
		"refresh_tokens_revoked": revoked,
		"request_id":             reqID,
	})
}

// handleLinkWallet implements POST /v1/identity/link-wallet.
// Binds a Constellation wallet address to a did:key during first-run or enrollment.
func (rt *Router) handleLinkWallet(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", reqID)
		return
	}

	var req struct {
		DID           string `json:"did"`
		WalletAddress string `json:"wallet_address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DID == "" || req.WalletAddress == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_FIELDS", "did and wallet_address are required", reqID)
		return
	}

	rt.enrollmentWalletMu.Lock()
	if rt.enrollmentWalletByDID == nil {
		rt.enrollmentWalletByDID = make(map[string]string)
	}
	rt.enrollmentWalletByDID[req.DID] = req.WalletAddress
	rt.enrollmentWalletMu.Unlock()

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"linked":         true,
		"did":            req.DID,
		"wallet_address": req.WalletAddress,
		"request_id":     reqID,
	})
}
