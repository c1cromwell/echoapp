package api

import (
	"net/http"
	"strings"
)

// handleMessageMerkleProof serves GET /v1/messages/{messageId}/merkle-proof (WO-15).
func (rt *Router) handleMessageMerkleProof(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.Anchoring == nil || rt.Anchoring.ProofStore() == nil {
		WriteError(w, http.StatusServiceUnavailable, "ANCHORING_NOT_CONFIGURED", "Message anchoring is not enabled", r.Header.Get("X-Request-ID"))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/v1/messages/")
	messageID := strings.TrimSuffix(path, "/merkle-proof")
	messageID = strings.Trim(messageID, "/")
	if messageID == "" || strings.Contains(messageID, "/") {
		WriteError(w, http.StatusBadRequest, "INVALID_MESSAGE_ID", "message id is required", r.Header.Get("X-Request-ID"))
		return
	}
	proof, ok, err := rt.Anchoring.ProofStore().Get(r.Context(), messageID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PROOF_LOOKUP_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if !ok {
		WriteError(w, http.StatusNotFound, "PROOF_NOT_FOUND", "No Merkle proof for this message yet", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, proof)
}
