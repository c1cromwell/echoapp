package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/wallet"
)

// WalletHandlers exposes authenticated /v3/wallet routes.
type WalletHandlers struct {
	Service *wallet.WalletService
	Store   wallet.Store
	// RealFunds enables real-funds custody enforcement (ECHO_WALLET_REAL_FUNDS).
	RealFunds bool
	// Proof verifies client-held proof-of-ownership; nil until the signing SDK
	// ships, which (in real-funds mode) hard-blocks value-moving operations.
	Proof wallet.ProofVerifier
	// Challenges issues single-use nonces for proof-of-ownership (real funds).
	Challenges *wallet.ChallengeStore
}

// requireCustody enforces real-funds custody rules on value-moving operations
// and returns the validated public key (used by /link to bind the account).
// In interim mode (default) it is a no-op so the TestFlight flow is unchanged.
// In real-funds mode it requires a verified proof-of-ownership and rejects
// server-derivable addresses; with no verifier wired it hard-blocks.
func (h *WalletHandlers) requireCustody(w http.ResponseWriter, r *http.Request, did, address, proof string) (string, bool) {
	if !h.RealFunds {
		return "", true
	}
	reqID := r.Header.Get("X-Request-ID")
	if h.Proof == nil {
		WriteError(w, http.StatusServiceUnavailable, "CUSTODY_NOT_READY",
			"real-funds custody requires client-side signing which is not yet available", reqID)
		return "", false
	}
	if address != "" && address == wallet.ServerDerivableAddress(did) {
		WriteError(w, http.StatusBadRequest, "SERVER_DERIVABLE_ADDRESS",
			"address must be user-held, not server-derivable", reqID)
		return "", false
	}
	pubKey, err := h.Proof.VerifyOwnership(did, address, proof)
	if err != nil {
		WriteError(w, http.StatusForbidden, "PROOF_INVALID", "proof of ownership failed", reqID)
		return "", false
	}
	return pubKey, true
}

// walletDID extracts the authenticated DID and rejects empty values, so a
// blank-DID identity can never read or mutate wallet state even if upstream
// auth is misrouted (defense-in-depth; auth middleware is the primary gate).
func walletDID(w http.ResponseWriter, r *http.Request) (string, bool) {
	did, _ := r.Context().Value(ContextKeyUserID).(string)
	if strings.TrimSpace(did) == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return "", false
	}
	return did, true
}

func (h *WalletHandlers) handleWalletRoot(w http.ResponseWriter, r *http.Request) {
	if h.Service == nil {
		WriteError(w, http.StatusServiceUnavailable, "WALLET_UNAVAILABLE", "Wallet service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did, ok := walletDID(w, r)
	if !ok {
		return
	}
	state, err := h.Service.GetWalletState(r.Context(), did)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "WALLET_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, state)
}

func (h *WalletHandlers) handleWalletStake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did, ok := walletDID(w, r)
	if !ok {
		return
	}
	if _, ok := h.requireCustody(w, r, did, "", r.Header.Get("X-Wallet-Proof")); !ok {
		return
	}
	var req struct {
		Amount int64  `json:"amount"`
		Tier   string `json:"tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 || req.Tier == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "amount and tier are required", r.Header.Get("X-Request-ID"))
		return
	}
	result, err := h.Service.StakeEcho(r.Context(), wallet.StakeRequest{
		DID:    did,
		Amount: req.Amount,
		Tier:   req.Tier,
	})
	if err != nil {
		status := http.StatusBadRequest
		if err == wallet.ErrInsufficientBalance {
			status = http.StatusConflict
		}
		WriteError(w, status, "STAKE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

func (h *WalletHandlers) handleWalletUnstake(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did, ok := walletDID(w, r)
	if !ok {
		return
	}
	if _, ok := h.requireCustody(w, r, did, "", r.Header.Get("X-Wallet-Proof")); !ok {
		return
	}
	var req struct {
		StakeID string `json:"stakeId"`
		Amount  int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StakeID == "" || req.Amount <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "stakeId and amount are required", r.Header.Get("X-Request-ID"))
		return
	}
	result, err := h.Service.Unstake(r.Context(), wallet.UnstakeRequest{
		DID:     did,
		StakeID: req.StakeID,
		Amount:  req.Amount,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, "UNSTAKE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

func (h *WalletHandlers) handleWalletDelegate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did, ok := walletDID(w, r)
	if !ok {
		return
	}
	if _, ok := h.requireCustody(w, r, did, "", r.Header.Get("X-Wallet-Proof")); !ok {
		return
	}
	var req struct {
		StakeID     string `json:"stakeId"`
		ValidatorID string `json:"validatorId"`
		Amount      int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StakeID == "" || req.ValidatorID == "" || req.Amount <= 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "stakeId, validatorId, and amount are required", r.Header.Get("X-Request-ID"))
		return
	}
	result, err := h.Service.DelegateToValidator(r.Context(), wallet.DelegateRequest{
		DID:         did,
		StakeID:     req.StakeID,
		ValidatorID: req.ValidatorID,
		Amount:      req.Amount,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, "DELEGATE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

func (h *WalletHandlers) handleWalletClaim(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did, ok := walletDID(w, r)
	if !ok {
		return
	}
	if _, ok := h.requireCustody(w, r, did, "", r.Header.Get("X-Wallet-Proof")); !ok {
		return
	}
	var req struct {
		Types []string `json:"types"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Types) == 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "types array is required", r.Header.Get("X-Request-ID"))
		return
	}
	result, err := h.Service.ClaimRewards(r.Context(), did, req.Types, TrustTierFromContext(r.Context(), 1))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "CLAIM_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, result)
}

func (h *WalletHandlers) handleWalletValidators(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	validators, err := h.Service.GetValidators(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "VALIDATORS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"validators": validators})
}

func (h *WalletHandlers) handleWalletLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Store == nil {
		WriteError(w, http.StatusServiceUnavailable, "WALLET_UNAVAILABLE", "Wallet store not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did, ok := walletDID(w, r)
	if !ok {
		return
	}
	var req struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Address) == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "address is required", r.Header.Get("X-Request-ID"))
		return
	}
	addr := strings.TrimSpace(req.Address)
	pubKey, ok := h.requireCustody(w, r, did, addr, r.Header.Get("X-Wallet-Proof"))
	if !ok {
		return
	}
	// In real-funds mode pubKey is the proof-validated key bound to the address;
	// in interim mode it is empty (server-derivable address, no binding).
	if err := h.Store.LinkDAGAccount(r.Context(), did, addr, pubKey); err != nil {
		WriteError(w, http.StatusInternalServerError, "LINK_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"did": did, "address": addr})
}

// handleWalletChallenge issues a single-use proof-of-ownership challenge for the
// authenticated DID (real-funds custody). The client signs it with the wallet
// key and returns it in the X-Wallet-Proof header on the next mutating call.
func (h *WalletHandlers) handleWalletChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did, ok := walletDID(w, r)
	if !ok {
		return
	}
	if h.Challenges == nil {
		WriteError(w, http.StatusServiceUnavailable, "WALLET_UNAVAILABLE", "Challenge store not configured", r.Header.Get("X-Request-ID"))
		return
	}
	challenge, expires, err := h.Challenges.Issue(did)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "CHALLENGE_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"challenge": challenge,
		"expiresAt": expires.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}
