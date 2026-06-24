package api

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/services/bots"
)

// handleBotsSubroute dispatches /v3/bots/* (Stage 4 / WO-11 foundation).
func (h *V3Handlers) handleBotsSubroute(w http.ResponseWriter, r *http.Request) {
	if h.Bots == nil {
		WriteError(w, http.StatusServiceUnavailable, "BOTS_UNAVAILABLE", "bot service not configured", r.Header.Get("X-Request-ID"))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/v3/bots")
	path = strings.Trim(path, "/")

	switch {
	case path == "catalog" && r.Method == http.MethodGet:
		h.handleBotsCatalog(w, r)
	case path == "installed" && r.Method == http.MethodGet:
		h.handleBotsInstalled(w, r)
	case path == "install" && r.Method == http.MethodPost:
		h.handleBotsInstall(w, r)
	default:
		if strings.HasSuffix(path, "/install") {
			botDID, err := url.PathUnescape(strings.TrimSuffix(path, "/install"))
			if err != nil || botDID == "" {
				WriteError(w, http.StatusBadRequest, "INVALID_BOT", "bot DID required", r.Header.Get("X-Request-ID"))
				return
			}
			switch r.Method {
			case http.MethodDelete:
				h.handleBotsUninstall(w, r, botDID)
			case http.MethodPatch:
				h.handleBotsSetActive(w, r, botDID)
			default:
				WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "DELETE or PATCH required", r.Header.Get("X-Request-ID"))
			}
			return
		}
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "unknown bots route", r.Header.Get("X-Request-ID"))
	}
}

func (h *V3Handlers) handleBotsCatalog(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"bots": bots.DefaultCatalog(),
	})
}

func (h *V3Handlers) handleBotsInstalled(w http.ResponseWriter, r *http.Request) {
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"installed": h.Bots.ListInstalled(did),
	})
}

func (h *V3Handlers) handleBotsInstall(w http.ResponseWriter, r *http.Request) {
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		BotDID      string            `json:"bot_did"`
		Permissions []bots.Permission `json:"permissions"`
	}
	if err := h.readJSON(r, &req); err != nil || req.BotDID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "bot_did and permissions required", r.Header.Get("X-Request-ID"))
		return
	}
	manifest, ok := bots.LookupManifest(req.BotDID)
	if !ok {
		WriteError(w, http.StatusNotFound, "BOT_NOT_FOUND", "bot not in catalog", r.Header.Get("X-Request-ID"))
		return
	}
	if !bots.ValidateGrant(manifest, req.Permissions) {
		WriteError(w, http.StatusBadRequest, "INVALID_PERMISSIONS", "permissions must match bot manifest", r.Header.Get("X-Request-ID"))
		return
	}
	inst := h.Bots.Install(did, req.BotDID, req.Permissions)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"installation": inst,
	})
}

func (h *V3Handlers) handleBotsUninstall(w http.ResponseWriter, r *http.Request, botDID string) {
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	if !h.Bots.Uninstall(did, botDID) {
		WriteError(w, http.StatusNotFound, "NOT_INSTALLED", "bot not installed", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"removed": true})
}

func (h *V3Handlers) handleBotsSetActive(w http.ResponseWriter, r *http.Request, botDID string) {
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		Active bool `json:"active"`
	}
	if err := h.readJSON(r, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "active boolean required", r.Header.Get("X-Request-ID"))
		return
	}
	inst, ok := h.Bots.SetActive(did, botDID, req.Active)
	if !ok {
		WriteError(w, http.StatusNotFound, "NOT_INSTALLED", "bot not installed", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"installation": inst})
}
