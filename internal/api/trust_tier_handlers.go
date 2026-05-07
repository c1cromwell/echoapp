package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

func (rt *Router) handleTrustTierCommitment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.IdentityL1 == nil {
		WriteError(w, http.StatusServiceUnavailable, "L1_UNAVAILABLE", "Identity L1 client is not configured", r.Header.Get("X-Request-ID"))
		return
	}
	uid, ok := r.Context().Value(ContextKeyUserID).(string)
	if !ok || uid == "" {
		WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", r.Header.Get("X-Request-ID"))
		return
	}
	var req struct {
		SubjectDID string `json:"subject_did"`
		Tier       int    `json:"tier"`
		Nonce      string `json:"nonce"`
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "READ_BODY", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if err := json.Unmarshal(body, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if req.SubjectDID == "" {
		req.SubjectDID = uid
	}
	if req.SubjectDID != uid {
		WriteError(w, http.StatusForbidden, "SUBJECT_MISMATCH", "subject_did must match the authenticated principal", r.Header.Get("X-Request-ID"))
		return
	}
	if req.Tier < 1 || req.Tier > 5 {
		WriteError(w, http.StatusBadRequest, "INVALID_TIER", "tier must be between 1 and 5", r.Header.Get("X-Request-ID"))
		return
	}
	if req.Nonce == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_NONCE", "nonce is required", r.Header.Get("X-Request-ID"))
		return
	}
	preimage := make([]byte, 0, 1+len(req.Nonce))
	preimage = append(preimage, byte(req.Tier))
	preimage = append(preimage, []byte(req.Nonce)...)
	sum := sha256.Sum256(preimage)
	commitment := hex.EncodeToString(sum[:])
	anchoredAt := time.Now().UnixMilli()
	tx := metagraph.TrustTierCommitmentUpdate{
		SubjectDID: req.SubjectDID,
		Commitment: commitment,
		AnchoredAt: anchoredAt,
	}
	if _, err := rt.IdentityL1.SubmitIdentityL1(r.Context(), tx); err != nil {
		WriteError(w, http.StatusBadGateway, "L1_SUBMIT_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if rt.Redis != nil {
		payload, _ := json.Marshal(map[string]interface{}{
			"subject_did": req.SubjectDID,
			"tier":        req.Tier,
			"commitment":  commitment,
			"anchored_at": anchoredAt,
		})
		_ = rt.Redis.CacheSet(r.Context(), "tiercommit:"+req.SubjectDID, payload, 60*time.Second)
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"subject_did": req.SubjectDID,
		"tier":        req.Tier,
		"commitment":  commitment,
		"anchored_at": anchoredAt,
		"request_id":  r.Header.Get("X-Request-ID"),
	})
}
