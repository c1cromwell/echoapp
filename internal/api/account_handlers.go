package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/thechadcromwell/echoapp/internal/auth"
)

// AccountPurger erases the server-side personal data held for a DID and reports
// how many rows were removed per table.
type AccountPurger interface {
	PurgeAccount(ctx context.Context, did string) (map[string]int64, error)
}

// elevatedFor reports whether the request carries a live elevated token for
// action, issued to userID. The ordinary session still has to authenticate;
// this is the extra check the step-up route is for.
func (rt *Router) elevatedFor(r *http.Request, userID, action string) bool {
	raw := r.Header.Get("X-Elevated-Token")
	if raw == "" || rt.tokenService == nil || userID == "" {
		return false
	}
	claims, err := rt.tokenService.ValidateAccessToken(raw)
	if err != nil {
		return false
	}
	return claims.Subject == userID && auth.IsElevatedForAction(claims, action)
}

// handleDeleteAccount implements DELETE /v1/users/account (WO-228; App Store
// Review Guideline 5.1.1(v)).
//
// It erases the account rather than merely signing it out:
//
//  1. server-side personal data is purged in one transaction (profile, device
//     and push-token bindings, discovery index, contacts, queued inbound
//     messages, backups, preferences, credentials, wallet link — the full list
//     and the documented exceptions live in database.accountPurgeSteps);
//  2. cached device keys are dropped, so the DID's signatures stop verifying;
//  3. every refresh token is revoked and any live socket is closed.
//
// If the purge fails nothing else is touched and 500 is returned, so the
// caller still holds a working session and can retry.
//
// End-to-end encrypted messages already delivered to other people stay on
// their devices; the server never held them in readable form.
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
	if !rt.elevatedFor(r, userID, string(auth.StepUpDeleteAccount)) {
		WriteError(w, http.StatusForbidden, "STEP_UP_REQUIRED", "delete account requires a step-up token for this action", reqID)
		return
	}

	purged := false
	var rowsRemoved int64
	if rt.AccountPurger != nil {
		removed, err := rt.AccountPurger.PurgeAccount(r.Context(), userID)
		if err != nil {
			log.Printf("account deletion failed (request %s): %v", reqID, err)
			WriteError(w, http.StatusInternalServerError, "DELETION_FAILED", "Could not delete the account. Nothing was removed; please try again.", reqID)
			return
		}
		purged = true
		for _, n := range removed {
			rowsRemoved += n
		}
	}

	if rt.Redis != nil {
		_ = rt.Redis.DeleteDIDDeviceKeys(r.Context(), userID)
		_ = rt.Redis.CacheDelete(r.Context(), "vip:tier:"+userID)
		_, _ = rt.Redis.QueueDrain(r.Context(), userID)
	}

	rt.enrollmentWalletMu.Lock()
	delete(rt.enrollmentWalletByDID, userID)
	rt.enrollmentWalletMu.Unlock()

	revoked := 0
	if rt.tokenService != nil {
		revoked = rt.tokenService.RevokeAllUserTokens(userID)
	}
	rt.WSHub.DisconnectUser(userID)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"deleted":                true,
		"did":                    userID,
		"server_data_purged":     purged,
		"rows_removed":           rowsRemoved,
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

	// Only the DID's own authenticated session may bind a wallet to it. This
	// route used to be unauthenticated, which let anyone rebind any DID.
	if caller, _ := r.Context().Value(ContextKeyUserID).(string); caller == "" || caller != req.DID {
		WriteError(w, http.StatusForbidden, "DID_MISMATCH", "did must match the authenticated caller", reqID)
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
