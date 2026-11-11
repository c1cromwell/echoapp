package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/thechadcromwell/echoapp/internal/database"
)

var safetyReportReasons = map[string]bool{
	"spam":          true,
	"harassment":    true,
	"impersonation": true,
	"other":         true,
}

// handleSafetyReports implements POST and GET /v1/safety/reports.
// A signed-in user files a report against another DID. GET returns the
// reports that caller filed, so the action is readable after submit.
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
		SubjectDID string `json:"subject_did"`
		Reason     string `json:"reason"`
		Context    string `json:"context"`
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

	report := database.SafetyReport{
		ID:          uuid.NewString(),
		ReporterDID: caller,
		SubjectDID:  req.SubjectDID,
		Reason:      req.Reason,
		Context:     req.Context,
		CreatedAt:   time.Now().UTC(),
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
