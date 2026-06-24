package api

import (
	"net/http"
)

// handleConversationNotifications syncs mute prefs for silent push (WO-56).
// GET/PUT /v3/conversations/{id}/notifications  body {"muted":true}
func (h *V3Handlers) handleConversationNotifications(w http.ResponseWriter, r *http.Request, conversationID string) {
	if h.ConvNotifPrefs == nil {
		WriteError(w, http.StatusServiceUnavailable, "NOTIFICATIONS_UNAVAILABLE", "conversation notifications not configured", r.Header.Get("X-Request-ID"))
		return
	}
	did := h.getDID(r)
	if did == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	switch r.Method {
	case http.MethodGet:
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"conversation_id": conversationID,
			"muted":           h.ConvNotifPrefs.IsMuted(did, conversationID),
		})
	case http.MethodPut:
		var req struct {
			Muted bool `json:"muted"`
		}
		if err := h.readJSON(r, &req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		h.ConvNotifPrefs.SetMuted(did, conversationID, req.Muted)
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"conversation_id": conversationID,
			"muted":           req.Muted,
		})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET or PUT required", r.Header.Get("X-Request-ID"))
	}
}
