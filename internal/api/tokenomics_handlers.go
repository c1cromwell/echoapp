package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/governance"
	"github.com/thechadcromwell/echoapp/internal/tokenomics"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/genesis"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/quests"
)

// TokenomicsHandlers serves WO-206/214/215/225/226/271 HTTP endpoints.
type TokenomicsHandlers struct {
	Service *tokenomics.Service
}

func (h *TokenomicsHandlers) handleEmissionStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, h.Service.Emission.Get())
}

func (h *TokenomicsHandlers) handleVesting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	vesting, err := h.Service.VestingInfo(r.Context(), didStr)
	if err != nil {
		WriteError(w, http.StatusNotFound, "VESTING_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	revocations := h.Service.Revocation.Events(didStr)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"did":               didStr,
		"is_founder":        genesis.IsFounderDID(didStr),
		"vesting":           vesting,
		"revocation_events": revocations,
		"explorer_url":      "https://dagexplorer.io/address/" + didStr,
		"request_id":        r.Header.Get("X-Request-ID"),
	})
}

func (h *TokenomicsHandlers) handleQuestsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	catalog, err := h.Service.Quests.ListCatalog(r.Context(), didStr)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "QUESTS_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	starter := make([]quests.CatalogEntry, 0)
	advanced := make([]quests.CatalogEntry, 0)
	for _, e := range catalog {
		if e.Tier == quests.TierAdvanced {
			advanced = append(advanced, e)
		} else {
			starter = append(starter, e)
		}
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"starter":    starter,
		"advanced":   advanced,
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

func (h *TokenomicsHandlers) handleQuestClaim(w http.ResponseWriter, r *http.Request, questID string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	txHash, err := h.Service.Quests.Claim(r.Context(), didStr, questID)
	if err != nil {
		code := http.StatusBadRequest
		errCode := "CLAIM_ERROR"
		if err == quests.ErrAlreadyClaimed {
			code = http.StatusConflict
			errCode = "ALREADY_CLAIMED"
		}
		WriteError(w, code, errCode, err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"quest_id":   questID,
		"tx_hash":    txHash,
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

func (h *TokenomicsHandlers) handleVIPAllowSpend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil || h.Service.VIP == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	rec, err := h.Service.VIP.AuthorizeVIP(r.Context(), didStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ALLOW_SPEND_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, rec)
}

func (h *TokenomicsHandlers) handleGovernanceVotingPower(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	tier := TrustTierFromContext(r.Context(), 1)
	var founderLocked int64
	if h.Service != nil && genesis.IsFounderDID(didStr) {
		founderLocked = h.Service.Revocation.RemainingLocked(didStr)
	}
	staked := int64(0)
	if h.Service != nil && h.Service.Wallet != nil {
		if state, err := h.Service.Wallet.GetWalletState(r.Context(), didStr); err == nil {
			staked = state.Staked + founderLocked
		}
	} else {
		staked = founderLocked
	}
	weight := governance.CalculateWeight(staked, tier)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"did":            didStr,
		"trust_tier":     tier,
		"total_staked":   staked,
		"founder_locked": founderLocked,
		"weight":         weight,
		"can_vote":       governance.CanVote(tier, staked),
		"request_id":     r.Header.Get("X-Request-ID"),
	})
}

func (h *TokenomicsHandlers) handleFounderRevocationInitiate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "READ_BODY", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		ID        string `json:"id"`
		TargetDID string `json:"target_founder_did"`
		Amount    int64  `json:"amount"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if req.ID == "" {
		req.ID = "rev-" + time.Now().UTC().Format("20060102150405")
	}
	rev, err := h.Service.Revocation.Initiate(req.ID, req.TargetDID, req.Amount)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "REVOCATION_INIT_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, rev)
}

func (h *TokenomicsHandlers) handleFounderRevocationStatus(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if h.Service == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	rev, err := h.Service.Revocation.Get(id)
	if err != nil {
		WriteError(w, http.StatusNotFound, "REVOCATION_NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, rev)
}

func (h *TokenomicsHandlers) handleFounderRevocationSign(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	did := r.Context().Value(ContextKeyUserID)
	didStr, _ := did.(string)
	if didStr == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	rev, err := h.Service.Revocation.Sign(id, didStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "SIGN_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, rev)
}

func (h *TokenomicsHandlers) handleFounderRevocationFinalize(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	ev, err := h.Service.Revocation.Finalize(id)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "FINALIZE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, ev)
}

// RegisterTokenomicsRoutes mounts tokenomics HTTP routes on the mux.
func RegisterTokenomicsRoutes(mux *http.ServeMux, h *TokenomicsHandlers) {
	if h == nil {
		return
	}
	mux.HandleFunc("/v1/tokens/emission/status", h.handleEmissionStatus)
	mux.HandleFunc("/v1/tokens/vesting", h.handleVesting)
	mux.HandleFunc("/v1/tokens/governance/voting-power", h.handleGovernanceVotingPower)
	mux.HandleFunc("/v1/tokens/vip/allow-spend", h.handleVIPAllowSpend)
	mux.HandleFunc("/v1/gamification/quests", h.handleQuestsList)
	mux.HandleFunc("/v1/gamification/quests/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/v1/gamification/quests/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) < 2 || parts[1] != "claim" {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown quest route", r.Header.Get("X-Request-ID"))
			return
		}
		h.handleQuestClaim(w, r, parts[0])
	})
	mux.HandleFunc("/v1/admin/founder-revocation/initiate", h.handleFounderRevocationInitiate)
	mux.HandleFunc("/v1/admin/founder-revocation/status/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/v1/admin/founder-revocation/status/")
		if id == "" {
			WriteError(w, http.StatusBadRequest, "MISSING_ID", "revocation id required", r.Header.Get("X-Request-ID"))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/sign") {
			h.handleFounderRevocationSign(w, r, strings.TrimSuffix(id, "/sign"))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/finalize") {
			h.handleFounderRevocationFinalize(w, r, strings.TrimSuffix(id, "/finalize"))
			return
		}
		h.handleFounderRevocationStatus(w, r, strings.Trim(id, "/"))
	})
}

// handleTokenomicsSubroutes is called from Router.handleV1 for tokenomics paths.
func (rt *Router) handleTokenomicsSubroutes(w http.ResponseWriter, r *http.Request) {
	if rt.Tokenomics == nil {
		WriteError(w, http.StatusServiceUnavailable, "TOKENOMICS_UNAVAILABLE", "tokenomics not configured", r.Header.Get("X-Request-ID"))
		return
	}
	path := r.URL.Path
	switch {
	case path == "/v1/tokens/emission/status":
		rt.Tokenomics.handleEmissionStatus(w, r)
	case path == "/v1/tokens/vesting":
		rt.Tokenomics.handleVesting(w, r)
	case path == "/v1/tokens/governance/voting-power":
		rt.Tokenomics.handleGovernanceVotingPower(w, r)
	case path == "/v1/tokens/vip/allow-spend":
		rt.Tokenomics.handleVIPAllowSpend(w, r)
	case path == "/v1/gamification/quests":
		rt.Tokenomics.handleQuestsList(w, r)
	case strings.HasPrefix(path, "/v1/gamification/quests/") && strings.HasSuffix(path, "/claim"):
		questID := strings.TrimSuffix(strings.TrimPrefix(path, "/v1/gamification/quests/"), "/claim")
		questID = strings.Trim(questID, "/")
		rt.Tokenomics.handleQuestClaim(w, r, questID)
	case path == "/v1/admin/founder-revocation/initiate":
		rt.Tokenomics.handleFounderRevocationInitiate(w, r)
	case strings.HasPrefix(path, "/v1/admin/founder-revocation/status/"):
		id := strings.TrimPrefix(path, "/v1/admin/founder-revocation/status/")
		id = strings.TrimSuffix(strings.TrimSuffix(id, "/sign"), "/finalize")
		id = strings.Trim(id, "/")
		if strings.HasSuffix(path, "/sign") {
			rt.Tokenomics.handleFounderRevocationSign(w, r, id)
		} else if strings.HasSuffix(path, "/finalize") {
			rt.Tokenomics.handleFounderRevocationFinalize(w, r, id)
		} else {
			rt.Tokenomics.handleFounderRevocationStatus(w, r, id)
		}
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown tokenomics route", r.Header.Get("X-Request-ID"))
	}
}
