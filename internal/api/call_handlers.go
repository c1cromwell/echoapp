package api

import (
	"net/http"
	"os"
)

// ICE server config for WebRTC (M4). Returns public STUN; optional TURN via env (M4b).
//
//	GET /v3/calls/ice-servers
//	GET /v3/calls/relays          — Signal-parity alias (Wave S3)
func (h *V3Handlers) handleCallsICEServers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	servers := []map[string]interface{}{
		{"urls": []string{"stun:stun.l.google.com:19302"}},
		{"urls": []string{"stun:stun1.l.google.com:19302"}},
	}
	turnConfigured := false
	if turnURL := os.Getenv("ECHO_TURN_URL"); turnURL != "" {
		turnConfigured = true
		entry := map[string]interface{}{"urls": []string{turnURL}}
		if user := os.Getenv("ECHO_TURN_USERNAME"); user != "" {
			entry["username"] = user
		}
		if cred := os.Getenv("ECHO_TURN_CREDENTIAL"); cred != "" {
			entry["credential"] = cred
		}
		servers = append(servers, entry)
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"ice_servers":      servers,
		"turn_configured":  turnConfigured,
		"relays":           servers, // alias for Signal-style clients / CallICEAPIClient
	})
}
