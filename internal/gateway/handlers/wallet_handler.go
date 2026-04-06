// Package handlers provides HTTP handlers for the API gateway.
package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/thechadcromwell/echoapp/internal/wallet"
)

// WalletHandler handles wallet-related HTTP endpoints.
type WalletHandler struct {
	service *wallet.WalletService
}

// NewWalletHandler creates a new wallet handler.
func NewWalletHandler(service *wallet.WalletService) *WalletHandler {
	return &WalletHandler{service: service}
}

// GetWalletState handles GET /api/v1/wallet/:did
func (h *WalletHandler) GetWalletState(w http.ResponseWriter, r *http.Request) {
	did := r.URL.Query().Get("did")
	if did == "" {
		writeError(w, http.StatusBadRequest, "did is required")
		return
	}

	state, err := h.service.GetWalletState(r.Context(), did)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, state)
}

// Stake handles POST /api/v1/wallet/stake
func (h *WalletHandler) Stake(w http.ResponseWriter, r *http.Request) {
	var req wallet.StakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.StakeEcho(r.Context(), req)
	if err != nil {
		status := http.StatusInternalServerError
		if err == wallet.ErrInvalidTier {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Delegate handles POST /api/v1/wallet/delegate
func (h *WalletHandler) Delegate(w http.ResponseWriter, r *http.Request) {
	var req wallet.DelegateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.DelegateToValidator(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// Unstake handles POST /api/v1/wallet/unstake
func (h *WalletHandler) Unstake(w http.ResponseWriter, r *http.Request) {
	var req wallet.UnstakeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.Unstake(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ClaimRewards handles POST /api/v1/wallet/claim
func (h *WalletHandler) ClaimRewards(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DID   string   `json:"did"`
		Types []string `json:"types"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.service.ClaimRewards(r.Context(), req.DID, req.Types)
	if err != nil {
		status := http.StatusInternalServerError
		if err == wallet.ErrNoPendingRewards {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetValidators handles GET /api/v1/wallet/validators
func (h *WalletHandler) GetValidators(w http.ResponseWriter, r *http.Request) {
	validators, err := h.service.GetValidators(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, validators)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
