package api

import (
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/validation"
)

// handleGroupBlockchainProof returns anchored group metadata (WO-156).
// GET /v1/groups/{id}/blockchain-proof
func (rt *Router) handleGroupBlockchainProof(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.V3 == nil || rt.V3.GroupAnchoring == nil {
		WriteError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "group anchoring not configured", r.Header.Get("X-Request-ID"))
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/v1/groups/")
	groupID := strings.TrimSuffix(path, "/blockchain-proof")
	if err := validation.ValidateGroupID(groupID); err != nil {
		WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	proof, ok := rt.V3.GroupAnchoring.GetBlockchainProof(groupID)
	if !ok {
		WriteError(w, http.StatusNotFound, "PROOF_NOT_FOUND", "no anchored metadata for group", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, proof)
}
