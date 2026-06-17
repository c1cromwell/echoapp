package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/services/comply"
)

// ComplyHandlers serves /comply/* (WO-250 / WO-252). Portal uses service token + X-Org-DID.
type ComplyHandlers struct {
	Comply *comply.Service
}

// RegisterComplyRoutes mounts Comply REST endpoints on the gateway (:8000/comply/*) or :8011 standalone.
func (h *ComplyHandlers) RegisterComplyRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/comply/dashboard", h.handleComplyDashboard)
	mux.HandleFunc("/comply/segments/summary", h.handleComplySegmentSummary)
	mux.HandleFunc("/comply/org/profile", h.handleComplyOrgProfile)
	mux.HandleFunc("/comply/audit/report", h.handleComplyAuditReport)
	mux.HandleFunc("/comply/retention/policy", h.handleComplyRetentionPolicy)
	mux.HandleFunc("/comply/litigation/hold", h.handleComplyLitigationHoldRoot)
	mux.HandleFunc("/comply/litigation/hold/", h.handleComplyLitigationHoldSub)
	mux.HandleFunc("/comply/ediscovery/export", h.handleComplyEDiscoveryExportRoot)
	mux.HandleFunc("/comply/ediscovery/export/", h.handleComplyEDiscoveryExportSub)
	mux.HandleFunc("/comply/conversations/", h.handleComplyConversationSubroute)
}

func (h *ComplyHandlers) handleComplyDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	orgDID, ok := h.authorizeComplyRead(r)
	if !ok || h.Comply == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
		return
	}
	stats, err := h.Comply.Dashboard(r.Context(), orgDID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "DASHBOARD_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, stats)
}

func (h *ComplyHandlers) handleComplyOrgProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	orgDID, ok := h.authorizeComplyRead(r)
	if !ok || h.Comply == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
		return
	}
	profile, err := h.Comply.GetOrgProfile(r.Context(), orgDID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "ORG_PROFILE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, profile)
}

func (h *ComplyHandlers) handleComplySegmentSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	orgDID, ok := h.authorizeComplyRead(r)
	if !ok || h.Comply == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
		return
	}
	report, err := h.Comply.SegmentDashboard(r.Context(), orgDID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "SEGMENTS_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, report)
}

func (h *ComplyHandlers) handleComplyRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		orgDID, ok := h.authorizeComplyRead(r)
		if !ok || h.Comply == nil {
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
			return
		}
		policies, err := h.Comply.ListPolicies(r.Context(), orgDID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{"policies": policies})
	case http.MethodPost:
		orgDID, ok := h.authorizeComply(r)
		if !ok || h.Comply == nil {
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
			return
		}
		var req struct {
			PolicyType     string  `json:"policy_type"`
			ConversationID string  `json:"conversation_id"`
			ScopeLabel     string  `json:"scope_label"`
			ExpiresAt      *string `json:"expires_at"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		var expires *time.Time
		if req.ExpiresAt != nil && *req.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
			if err != nil {
				WriteError(w, http.StatusBadRequest, "INVALID_EXPIRES", "expires_at must be RFC3339", r.Header.Get("X-Request-ID"))
				return
			}
			expires = &t
		}
		actor := orgDID
		if v := r.Context().Value(ContextKeyUserID); v != nil {
			if s, ok := v.(string); ok && s != "" && s != "comply-service" {
				actor = s
			}
		}
		pol, err := h.Comply.CreateRetentionPolicy(r.Context(), comply.CreatePolicyInput{
			OrgDID:         orgDID,
			PolicyType:     database.RetentionPolicyType(req.PolicyType),
			ConversationID: req.ConversationID,
			ScopeLabel:     req.ScopeLabel,
			ExpiresAt:      expires,
			CreatedByDID:   actor,
		})
		if err != nil {
			WriteError(w, http.StatusBadRequest, "POLICY_CREATE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusCreated, pol)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only", r.Header.Get("X-Request-ID"))
	}
}

func (h *ComplyHandlers) handleComplyConversationSubroute(w http.ResponseWriter, r *http.Request) {
	orgDID, ok := h.authorizeComply(r)
	if !ok || h.Comply == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/comply/conversations/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown comply conversation route", r.Header.Get("X-Request-ID"))
		return
	}
	convID, action := parts[0], parts[1]
	switch action {
	case "retention":
		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
			return
		}
		var req struct {
			PolicyType string `json:"policy_type"`
			ScopeLabel string `json:"scope_label"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		policyType := database.RetentionPolicyType(req.PolicyType)
		if !policyType.Valid() {
			policyType = database.PolicyPermanent
		}
		actor := orgDID
		if v := r.Context().Value(ContextKeyUserID); v != nil {
			if s, ok := v.(string); ok && s != "" && s != "comply-service" {
				actor = s
			}
		}
		pol, err := h.Comply.CreateRetentionPolicy(r.Context(), comply.CreatePolicyInput{
			OrgDID:         orgDID,
			PolicyType:     policyType,
			ConversationID: convID,
			ScopeLabel:     req.ScopeLabel,
			CreatedByDID:   actor,
		})
		if err != nil {
			WriteError(w, http.StatusBadRequest, "RETENTION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"conversation_id": convID,
			"policy_id":       pol.ID,
			"policy_type":     pol.PolicyType,
		})
	case "release":
		if r.Method != http.MethodPost {
			WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
			return
		}
		if err := h.Comply.ReleaseConversation(r.Context(), convID); err != nil {
			WriteError(w, http.StatusInternalServerError, "RELEASE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{"conversation_id": convID, "released": true})
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown comply conversation route", r.Header.Get("X-Request-ID"))
	}
}

func (h *ComplyHandlers) authorizeComply(r *http.Request) (orgDID string, ok bool) {
	return h.authorizeComplyMode(r, false)
}

func (h *ComplyHandlers) authorizeComplyRead(r *http.Request) (orgDID string, ok bool) {
	return h.authorizeComplyMode(r, true)
}

func (h *ComplyHandlers) authorizeComplyMode(r *http.Request, allowUserRead bool) (orgDID string, ok bool) {
	orgDID = strings.TrimSpace(r.Header.Get("X-Org-DID"))
	if orgDID == "" {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	expected := ""
	if h.Comply != nil {
		expected = h.Comply.ServiceToken()
	}
	if expected == "" {
		expected = os.Getenv("COMPLY_SERVICE_TOKEN")
	}
	if expected != "" && token == expected {
		return orgDID, true
	}
	actor := actorDIDFromRequest(r)
	if actor != "" && h.Comply != nil {
		var err error
		if allowUserRead {
			err = h.Comply.AuthorizeOrgRead(r.Context(), orgDID, actor)
		} else {
			err = h.Comply.AuthorizeOrgWrite(r.Context(), orgDID, actor)
		}
		if err == nil {
			return orgDID, true
		}
	}
	if expected == "" && actor != "" {
		return orgDID, true
	}
	return "", false
}

func actorDIDFromRequest(r *http.Request) string {
	if v := r.Context().Value(ContextKeyUserID); v != nil {
		if s, ok := v.(string); ok && s != "" && s != "comply-service" {
			return s
		}
	}
	return ""
}

// ServiceAuth validates portal service token requests (used by auth middleware bypass).
func (h *ComplyHandlers) ServiceAuth(r *http.Request) bool {
	if h == nil || h.Comply == nil {
		return false
	}
	orgDID := strings.TrimSpace(r.Header.Get("X-Org-DID"))
	if orgDID == "" {
		return false
	}
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	expected := h.Comply.ServiceToken()
	if expected == "" {
		expected = os.Getenv("COMPLY_SERVICE_TOKEN")
	}
	return expected != "" && token == expected
}
