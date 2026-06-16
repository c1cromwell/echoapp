package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/comply"
)

func (h *ComplyHandlers) handleComplyAuditReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	orgDID, ok := h.authorizeComplyRead(r)
	if !ok || h.Comply == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
		return
	}
	from, to := parseTimeRange(r)
	report, err := h.Comply.GenerateAuditReport(r.Context(), orgDID, from, to)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "AUDIT_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	format := strings.ToLower(r.URL.Query().Get("format"))
	if format == "pdf" {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=comply-audit.pdf")
		_, _ = w.Write(renderAuditPDF(report))
		return
	}
	WriteJSON(w, http.StatusOK, report)
}

func (h *ComplyHandlers) handleComplyLitigationHoldRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		orgDID, ok := h.authorizeComplyRead(r)
		if !ok || h.Comply == nil {
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
			return
		}
		activeOnly := r.URL.Query().Get("status") != "all"
		matters, err := h.Comply.ListLitigationHolds(r.Context(), orgDID, activeOnly)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{"matters": matters})
	case http.MethodPost:
		h.createLitigationHold(w, r)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only", r.Header.Get("X-Request-ID"))
	}
}

func (h *ComplyHandlers) createLitigationHold(w http.ResponseWriter, r *http.Request) {
	orgDID, ok := h.authorizeComply(r)
	if !ok || h.Comply == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		MatterID      string   `json:"matterId"`
		CustodianDIDs []string `json:"custodianDids"`
		Scope         string   `json:"scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	actor := orgDID
	if v := r.Context().Value(ContextKeyUserID); v != nil {
		if s, ok := v.(string); ok && s != "" && s != "comply-service" {
			actor = s
		}
	}
	matter, err := h.Comply.ActivateLitigationHold(r.Context(), comply.ActivateLitigationHoldInput{
		OrgDID:         orgDID,
		MatterID:       req.MatterID,
		CustodianDIDs:  req.CustodianDIDs,
		ScopeLabel:     req.Scope,
		ActivatedByDID: actor,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, "HOLD_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, matter)
}

func (h *ComplyHandlers) handleComplyLitigationHoldSub(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/comply/litigation/hold/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "matter id required", r.Header.Get("X-Request-ID"))
		return
	}
	matterID := parts[0]

	if len(parts) == 2 && parts[1] == "release" && r.Method == http.MethodPut {
		orgDID, ok := h.authorizeComply(r)
		if !ok || h.Comply == nil {
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
			return
		}
		actor := orgDID
		matter, err := h.Comply.ReleaseLitigationHold(r.Context(), orgDID, matterID, actor)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "RELEASE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, matter)
		return
	}
	if r.Method == http.MethodGet && len(parts) == 1 {
		orgDID, ok := h.authorizeComplyRead(r)
		if !ok || h.Comply == nil {
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
			return
		}
		matter, err := h.Comply.GetLitigationHold(r.Context(), orgDID, matterID)
		if err != nil {
			WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusOK, matter)
		return
	}
	WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown litigation route", r.Header.Get("X-Request-ID"))
}

func (h *ComplyHandlers) handleComplyEDiscoveryExportRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createEDiscoveryExport(w, r)
	case http.MethodGet:
		h.listEDiscoveryExports(w, r)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only", r.Header.Get("X-Request-ID"))
	}
}

func (h *ComplyHandlers) createEDiscoveryExport(w http.ResponseWriter, r *http.Request) {
	orgDID, ok := h.authorizeComply(r)
	if !ok || h.Comply == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		MatterID     string   `json:"matterId"`
		CustodianSet []string `json:"custodianSet"`
		DateFrom     *string  `json:"dateFrom"`
		DateTo       *string  `json:"dateTo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	actor := orgDID
	var from, to *time.Time
	if req.DateFrom != nil {
		t, err := time.Parse(time.RFC3339, *req.DateFrom)
		if err == nil {
			from = &t
		}
	}
	if req.DateTo != nil {
		t, err := time.Parse(time.RFC3339, *req.DateTo)
		if err == nil {
			to = &t
		}
	}
	export, err := h.Comply.CreateEDiscoveryExport(r.Context(), comply.CreateExportInput{
		OrgDID:        orgDID,
		MatterID:      req.MatterID,
		RequesterDID:  actor,
		DateFrom:      from,
		DateTo:        to,
		CustodianDIDs: req.CustodianSet,
	})
	if err != nil {
		WriteError(w, http.StatusBadRequest, "EXPORT_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusAccepted, export)
}

func (h *ComplyHandlers) listEDiscoveryExports(w http.ResponseWriter, r *http.Request) {
	orgDID, ok := h.authorizeComplyRead(r)
	if !ok || h.Comply == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
		return
	}
	exports, err := h.Comply.ListEDiscoveryExports(r.Context(), orgDID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"exports": exports})
}

func (h *ComplyHandlers) handleComplyEDiscoveryExportSub(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	orgDID, ok := h.authorizeComplyRead(r)
	if !ok || h.Comply == nil {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Comply authorization required", r.Header.Get("X-Request-ID"))
		return
	}
	exportID := strings.TrimPrefix(r.URL.Path, "/comply/ediscovery/export/")
	export, err := h.Comply.GetEDiscoveryExport(r.Context(), orgDID, exportID)
	if err != nil {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, export)
}

func parseTimeRange(r *http.Request) (from, to time.Time) {
	to = time.Now().UTC()
	from = to.Add(-30 * 24 * time.Hour)
	if raw := r.URL.Query().Get("from"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			from = t
		}
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			to = t
		}
	}
	return from, to
}

// renderAuditPDF returns a minimal PDF stub with zero-PII summary text (WO-252).
func renderAuditPDF(report *comply.AuditReport) []byte {
	text := "Echo Comply Audit Report\nOrg: " + report.OrgDID + "\nGenerated: " + report.GeneratedAt.Format(time.RFC3339)
	text += "\nActive retention: " + itoa(report.ActiveRetention)
	text += "\nActive holds: " + itoa(report.ActiveHolds)
	text += "\nPending exports: " + itoa(report.PendingExports)
	text += "\n" + report.VerificationNotice
	// Minimal valid PDF with one text stream.
	content := "BT /F1 10 Tf 50 750 Td (" + escapePDF(text) + ") Tj ET"
	obj := "1 0 obj<< /Type /Catalog /Pages 2 0 R >>endobj\n"
	obj += "2 0 obj<< /Type /Pages /Kids [3 0 R] /Count 1 >>endobj\n"
	obj += "3 0 obj<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources<< /Font<< /F1 5 0 R >> >> >>endobj\n"
	obj += "4 0 obj<< /Length " + itoa(len(content)) + " >>stream\n" + content + "\nendstream endobj\n"
	obj += "5 0 obj<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>endobj\n"
	xref := "xref\n0 6\n0000000000 65535 f \n"
	body := "%PDF-1.4\n" + obj + xref + "trailer<< /Size 6 /Root 1 0 R >>\nstartxref\n0\n%%EOF"
	return []byte(body)
}

func escapePDF(s string) string {
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return strings.ReplaceAll(s, "\n", " ")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}
