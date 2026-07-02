package api

import (
	"errors"
	"net/http"

	"github.com/thechadcromwell/echoapp/internal/services/messaging"
)

// handleDisappearingRestrictions returns trust-tier TTL policy for the caller (WO-115).
func (h *V3Handlers) handleDisappearingRestrictions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	tier := TrustTierFromContext(r.Context(), 1)
	policy := messaging.PolicyForTier(tier)
	resp := map[string]interface{}{
		"trust_tier": tier,
		"policy":     policy,
	}
	if h.DisappearingRestrictions != nil {
		resp["violation_count"] = h.DisappearingRestrictions.ViolationCount(h.getDID(r))
	}
	WriteJSON(w, http.StatusOK, resp)
}

// handleDisappearingAppeals accepts restriction appeals (WO-115).
func (h *V3Handlers) handleDisappearingAppeals(w http.ResponseWriter, r *http.Request) {
	if h.DisappearingRestrictions == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "appeals not configured", r.Header.Get("X-Request-ID"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		did := h.getDID(r)
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"appeals": h.DisappearingRestrictions.ListAppeals(did),
		})
	case http.MethodPost:
		var req struct {
			Reason string `json:"reason"`
		}
		if err := h.readJSON(r, &req); err != nil || req.Reason == "" {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "reason is required", r.Header.Get("X-Request-ID"))
			return
		}
		appeal := h.DisappearingRestrictions.SubmitAppeal(h.getDID(r), req.Reason)
		WriteJSON(w, http.StatusCreated, appeal)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only", r.Header.Get("X-Request-ID"))
	}
}

func validateDisappearingTTL(w http.ResponseWriter, r *http.Request, tier, ttl int, restrictions *messaging.DisappearingRestrictionService, did string) bool {
	if err := messaging.ValidateDisappearingTTL(tier, ttl); err != nil {
		if restrictions != nil && did != "" {
			restrictions.RecordViolation(did)
		}
		if errors.Is(err, messaging.ErrDisappearingTTLRestricted) {
			WriteError(w, http.StatusForbidden, "DISAPPEARING_TTL_RESTRICTED", err.Error(), r.Header.Get("X-Request-ID"))
			return false
		}
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return false
	}
	return true
}
