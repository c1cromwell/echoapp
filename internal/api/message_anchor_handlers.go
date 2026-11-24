package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// anchorProofStore returns the proof index to read from: the durable store
// when one is wired (it works with the chain and the anchoring worker off),
// otherwise the anchoring service's own store.
func (rt *Router) anchorProofStore() metagraph.ProofStore {
	if rt.AnchorStore != nil {
		return rt.AnchorStore
	}
	if rt.Anchoring != nil {
		return rt.Anchoring.ProofStore()
	}
	return nil
}

// handleMessageMerkleProof serves GET /v1/messages/{messageId}/merkle-proof (WO-15).
func (rt *Router) handleMessageMerkleProof(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	store := rt.anchorProofStore()
	if store == nil {
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
	proof, ok, err := store.Get(r.Context(), messageID)
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

// handleAnchorProofBundle serves GET /v1/anchors/{root}/proof[?messageId=...].
//
// It returns a self-contained bundle built from the gateway's durable record
// only (no chain call), so it keeps working with the chain switched off. Same
// auth level as GET /v1/messages/{id}/merkle-proof. The bundle can be checked
// offline with cmd/verifyproof.
func (rt *Router) handleAnchorProofBundle(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", reqID)
		return
	}
	if rt.AnchorStore == nil {
		WriteError(w, http.StatusServiceUnavailable, "ANCHORING_NOT_CONFIGURED", "Message anchoring is not enabled", reqID)
		return
	}
	root := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/v1/anchors/"), "/proof")
	if !isAnchorRoot(root) {
		WriteError(w, http.StatusBadRequest, "INVALID_ROOT", "root must be 64 lowercase hex chars", reqID)
		return
	}
	bundle, err := metagraph.BuildProofBundle(r.Context(), rt.AnchorStore, rt.ChainNetwork, root, r.URL.Query().Get("messageId"))
	switch {
	case errors.Is(err, metagraph.ErrRootNotFound):
		WriteError(w, http.StatusNotFound, "ROOT_NOT_FOUND", "No anchor record for this root", reqID)
	case errors.Is(err, metagraph.ErrMessageNotInRoot):
		WriteError(w, http.StatusNotFound, "PROOF_NOT_FOUND", "No Merkle proof for this message under this root", reqID)
	case err != nil:
		WriteError(w, http.StatusInternalServerError, "PROOF_LOOKUP_FAILED", "Anchor proof lookup failed", reqID)
	default:
		WriteJSON(w, http.StatusOK, bundle)
	}
}

func isAnchorRoot(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9') && !(c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}
