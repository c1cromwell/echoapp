package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/thechadcromwell/echoapp/internal/database"
)

var safetyReportReasons = map[string]bool{
	"spam":          true,
	"harassment":    true,
	"impersonation": true,
	"other":         true,
}

// safetyReportMaxIDLen bounds the optional message/conversation/channel ids.
const safetyReportMaxIDLen = 256

// handleSafetyReports implements POST and GET /v1/safety/reports.
// A signed-in user files a report against another DID. GET returns the
// reports that caller filed, so the action is readable after submit; the
// free-text details are not echoed back (operators read them through
// GET /v1/admin/safety/reports).
func (rt *Router) handleSafetyReports(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	caller, _ := r.Context().Value(ContextKeyUserID).(string)
	if caller == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", reqID)
		return
	}

	switch r.Method {
	case http.MethodPost:
		rt.saveSafetyReport(w, r, caller, reqID)
	case http.MethodGet:
		reports, err := rt.listSafetyReports(r, caller)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "REPORT_LIST_FAILED", "could not load reports", reqID)
			return
		}
		if reports == nil {
			reports = []database.SafetyReport{}
		}
		for i := range reports {
			reports[i].Details = ""
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"reports":    reports,
			"request_id": reqID,
		})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET and POST are allowed", reqID)
	}
}

func (rt *Router) saveSafetyReport(w http.ResponseWriter, r *http.Request, caller, reqID string) {
	var req struct {
		SubjectDID     string `json:"subject_did"`
		Reason         string `json:"reason"`
		Context        string `json:"context"`
		MessageID      string `json:"message_id"`
		ConversationID string `json:"conversation_id"`
		ChannelID      string `json:"channel_id"`
		Details        string `json:"details"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON body", reqID)
		return
	}
	req.SubjectDID = strings.TrimSpace(req.SubjectDID)
	req.Reason = strings.TrimSpace(req.Reason)
	req.Context = strings.TrimSpace(req.Context)
	if req.Context == "" {
		req.Context = "contact"
	}
	if req.SubjectDID == "" || req.SubjectDID == caller {
		WriteError(w, http.StatusBadRequest, "INVALID_SUBJECT", "subject_did must be someone else", reqID)
		return
	}
	if !safetyReportReasons[req.Reason] {
		WriteError(w, http.StatusBadRequest, "INVALID_REASON", "reason must be spam, harassment, impersonation, or other", reqID)
		return
	}
	if req.Context != "contact" && req.Context != "message" && req.Context != "channel" {
		WriteError(w, http.StatusBadRequest, "INVALID_CONTEXT", "context must be contact, message, or channel", reqID)
		return
	}

	req.MessageID = strings.TrimSpace(req.MessageID)
	req.ConversationID = strings.TrimSpace(req.ConversationID)
	req.ChannelID = strings.TrimSpace(req.ChannelID)
	req.Details = strings.TrimSpace(req.Details)
	if len(req.MessageID) > safetyReportMaxIDLen || len(req.ConversationID) > safetyReportMaxIDLen || len(req.ChannelID) > safetyReportMaxIDLen {
		WriteError(w, http.StatusBadRequest, "INVALID_REFERENCE", "message_id, conversation_id and channel_id must be at most 256 bytes", reqID)
		return
	}
	if utf8.RuneCountInString(req.Details) > database.SafetyReportMaxDetails {
		WriteError(w, http.StatusBadRequest, "DETAILS_TOO_LONG", "details must be at most 1000 characters", reqID)
		return
	}

	report := database.SafetyReport{
		ID:             uuid.NewString(),
		ReporterDID:    caller,
		SubjectDID:     req.SubjectDID,
		Reason:         req.Reason,
		Context:        req.Context,
		MessageID:      req.MessageID,
		ConversationID: req.ConversationID,
		ChannelID:      req.ChannelID,
		Details:        req.Details,
		CreatedAt:      time.Now().UTC(),
	}
	if err := rt.persistSafetyReport(r, report); err != nil {
		WriteError(w, http.StatusInternalServerError, "REPORT_SAVE_FAILED", "could not save the report", reqID)
		return
	}
	WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         report.ID,
		"received":   true,
		"request_id": reqID,
	})
}

func (rt *Router) persistSafetyReport(r *http.Request, report database.SafetyReport) error {
	if rt.V3 != nil && rt.V3.DB != nil {
		return rt.V3.DB.SaveSafetyReport(r.Context(), report)
	}
	rt.safetyMu.Lock()
	rt.safetyReports = append(rt.safetyReports, report)
	rt.safetyMu.Unlock()
	return nil
}

func (rt *Router) listSafetyReports(r *http.Request, caller string) ([]database.SafetyReport, error) {
	if rt.V3 != nil && rt.V3.DB != nil {
		return rt.V3.DB.ListSafetyReportsByReporter(r.Context(), caller)
	}
	rt.safetyMu.Lock()
	defer rt.safetyMu.Unlock()
	var out []database.SafetyReport
	for _, report := range rt.safetyReports {
		if report.ReporterDID == caller {
			out = append(out, report)
		}
	}
	return out, nil
}

// --- Moderation read path (operators only) ---

const (
	adminSafetyReportsPath     = "/v1/admin/safety/reports"
	adminAPITokenEnv           = "ADMIN_API_TOKEN"
	safetyReportsDefaultLimit  = 50
	safetyReportsMaxLimit      = 200
	adminAPITokenMinimumLength = 32
)

// adminAPIRequest reports whether the request targets an operator route that
// authenticates itself with ADMIN_API_TOKEN instead of a user session, so the
// auth middleware lets it through to the handler (the same arrangement the
// trust-registry admin mutations use with their own key).
func adminAPIRequest(path string) bool {
	return path == adminSafetyReportsPath
}

// adminAPIToken returns the configured operator token, or "" when the operator
// API is switched off. A token shorter than 32 characters is treated as unset
// rather than leaving a guessable secret in front of moderation data.
func adminAPIToken() string {
	token := os.Getenv(adminAPITokenEnv)
	if len(token) < adminAPITokenMinimumLength {
		return ""
	}
	return token
}

// handleAdminSafetyReports implements GET /v1/admin/safety/reports.
//
//	Authorization: Bearer <ADMIN_API_TOKEN>
//	Query: subject_did (optional), limit (default 50, max 200), offset
//
// Operators only. It is the one place the reporter's free-text details are
// returned. With no ADMIN_API_TOKEN configured the route does not exist (404);
// a missing or wrong token is 403.
func (rt *Router) handleAdminSafetyReports(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	token := adminAPIToken()
	if token == "" {
		WriteError(w, http.StatusNotFound, "ENDPOINT_NOT_FOUND", "Endpoint not found", reqID)
		return
	}
	presented, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok || !constantTimeStringEqual(presented, token) {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "operator token required", reqID)
		return
	}
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", reqID)
		return
	}

	query := r.URL.Query()
	limit := safetyReportsDefaultLimit
	if raw := query.Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 {
			WriteError(w, http.StatusBadRequest, "INVALID_LIMIT", "limit must be a positive integer", reqID)
			return
		}
		limit = min(n, safetyReportsMaxLimit)
	}
	offset := 0
	if raw := query.Get("offset"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			WriteError(w, http.StatusBadRequest, "INVALID_OFFSET", "offset must be zero or a positive integer", reqID)
			return
		}
		offset = n
	}
	subject := strings.TrimSpace(query.Get("subject_did"))

	// One extra row tells us whether another page exists.
	var reports []database.SafetyReport
	var err error
	if rt.V3 != nil && rt.V3.DB != nil {
		reports, err = rt.V3.DB.ListSafetyReports(r.Context(), subject, limit+1, offset)
	} else {
		rt.safetyMu.Lock()
		reports = database.PageSafetyReports(rt.safetyReports, subject, limit+1, offset)
		rt.safetyMu.Unlock()
	}
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "REPORT_LIST_FAILED", "could not load reports", reqID)
		return
	}
	resp := map[string]interface{}{
		"limit":      limit,
		"offset":     offset,
		"request_id": reqID,
	}
	if len(reports) > limit {
		reports = reports[:limit]
		resp["next_offset"] = offset + limit
	}
	if reports == nil {
		reports = []database.SafetyReport{}
	}
	resp["reports"] = reports
	WriteJSON(w, http.StatusOK, resp)
}
