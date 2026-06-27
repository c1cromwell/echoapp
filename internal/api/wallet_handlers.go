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
	did, _ := r.Context().Value(ContextKeyUserID).(string)
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
	did, _ := r.Context().Value(ContextKeyUserID).(string)
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
	did, _ := r.Context().Value(ContextKeyUserID).(string)
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
	did, _ := r.Context().Value(ContextKeyUserID).(string)
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
	did, _ := r.Context().Value(ContextKeyUserID).(string)
	var req struct {
		Types []string `json:"types"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Types) == 0 {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "types array is required", r.Header.Get("X-Request-ID"))
		return
	}
	result, err := h.Service.ClaimRewards(r.Context(), did, req.Types)
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
	did, _ := r.Context().Value(ContextKeyUserID).(string)
	var req struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Address) == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "address is required", r.Header.Get("X-Request-ID"))
		return
	}
	addr := strings.TrimSpace(req.Address)
	if err := h.Store.LinkDAGAddress(r.Context(), did, addr); err != nil {
		WriteError(w, http.StatusInternalServerError, "LINK_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"did": did, "address": addr})
}
