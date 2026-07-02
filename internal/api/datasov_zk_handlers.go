package api

import (
	"net/http"

	"github.com/thechadcromwell/echoapp/internal/services/zk"
)

// handleDataSovSettings manages opt-in for the data sovereignty layer (WO-248).
func (h *V3Handlers) handleDataSovSettings(w http.ResponseWriter, r *http.Request) {
	if h.DataSov == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "data sovereignty not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	switch r.Method {
	case http.MethodGet:
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"opted_in": h.DataSov.IsOptedIn(did),
		})
	case http.MethodPost:
		var req struct {
			OptedIn bool `json:"opted_in"`
		}
		if err := h.readJSON(r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", r.Header.Get("X-Request-ID"))
			return
		}
		h.DataSov.SetOptIn(did, req.OptedIn)
		WriteJSON(w, http.StatusOK, map[string]interface{}{"opted_in": req.OptedIn})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or POST only", r.Header.Get("X-Request-ID"))
	}
}

// handleDataSovContribute accepts anonymized stats hashes (WO-249).
func (h *V3Handlers) handleDataSovContribute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST", r.Header.Get("X-Request-ID"))
		return
	}
	if h.DataSov == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "data sovereignty not configured", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		StatsHash string `json:"stats_hash"`
	}
	if err := h.readJSON(r, &req); err != nil || req.StatsHash == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "stats_hash required", r.Header.Get("X-Request-ID"))
		return
	}
	c, err := h.DataSov.Contribute(h.getDID(r), req.StatsHash)
	if err != nil {
		WriteError(w, http.StatusForbidden, "NOT_OPTED_IN", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusCreated, c)
}

// handleDataSovQuery returns aggregate sovereignty stats (WO-249).
func (h *V3Handlers) handleDataSovQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET", r.Header.Get("X-Request-ID"))
		return
	}
	if h.DataSov == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "data sovereignty not configured", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, h.DataSov.Query())
}

// handleZKVerify verifies a ZK proof with commitment fallback (WO-236).
func (h *V3Handlers) handleZKVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST", r.Header.Get("X-Request-ID"))
		return
	}
	if h.ZKVerifier == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "zk verifier not configured", r.Header.Get("X-Request-ID"))
		return
	}
	var req zk.VerifyRequest
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "invalid json", r.Header.Get("X-Request-ID"))
		return
	}
	res, err := h.ZKVerifier.Verify(req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VERIFY_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, res)
}

// handlePQModeStatus returns post-quantum mode availability (WO-257/258 stub).
func (h *V3Handlers) handlePQModeStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"available": false,
		"mode":      "stub",
		"detail":    "Post-quantum cryptography mode is not yet enabled in this build (WO-257).",
	})
}

// WireDataSovZK registers Phase 5 privacy extension routes.
func (h *V3Handlers) WireDataSovZK(mux *http.ServeMux) {
	mux.HandleFunc("/v3/datasov/settings", h.handleDataSovSettings)
	mux.HandleFunc("/v3/datasov/contribute", h.handleDataSovContribute)
	mux.HandleFunc("/v3/datasov/query", h.handleDataSovQuery)
	mux.HandleFunc("/v3/zk/verify", h.handleZKVerify)
	mux.HandleFunc("/v3/pq/status", h.handlePQModeStatus)
}
