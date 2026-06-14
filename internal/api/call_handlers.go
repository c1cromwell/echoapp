package api

import (
	"net/http"
)

// ICE server config for WebRTC (M4). Returns public STUN; TURN credentials are client-held.
//
//	GET /v3/calls/ice-servers
func (h *V3Handlers) handleCallsICEServers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"ice_servers": []map[string]interface{}{
			{"urls": []string{"stun:stun.l.google.com:19302"}},
			{"urls": []string{"stun:stun1.l.google.com:19302"}},
		},
	})
}
