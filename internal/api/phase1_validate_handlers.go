package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// handlePhase1TrustTierCommitment exposes POST /v1/phase1/trust-tier-commitment for
// scripts/validate-phase1.sh (WO-230 Step 3). It forwards a TrustTierCommitmentUpdate
// to Identity L1 without JWT — only when ENVIRONMENT=development or
// PHASE1_ALLOW_OPEN_VALIDATE=true.
func (rt *Router) handlePhase1TrustTierCommitment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if os.Getenv("ENVIRONMENT") != "development" && os.Getenv("PHASE1_ALLOW_OPEN_VALIDATE") != "true" {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "open Phase-1 validate endpoints are disabled", r.Header.Get("X-Request-ID"))
		return
	}
	if rt.IdentityL1 == nil {
		WriteError(w, http.StatusServiceUnavailable, "L1_UNAVAILABLE", "Identity L1 client is not configured", r.Header.Get("X-Request-ID"))
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
	req.SubjectDID = strings.TrimSpace(req.SubjectDID)
	req.Nonce = strings.TrimSpace(req.Nonce)
	if req.SubjectDID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_SUBJECT", "subject_did is required", r.Header.Get("X-Request-ID"))
		return
	}
	if !strings.HasPrefix(req.SubjectDID, "did:key:") {
		WriteError(w, http.StatusBadRequest, "INVALID_SUBJECT", "subject_did must be did:key…", r.Header.Get("X-Request-ID"))
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

	commitment := metagraph.TrustTierCommitmentHex(req.Tier, req.Nonce)
	anchoredAt := time.Now().UnixMilli()
	tx := metagraph.TrustTierCommitmentUpdate{
		SubjectDID: req.SubjectDID,
		Commitment: commitment,
		AnchoredAt: anchoredAt,
	}
	txHash, err := rt.IdentityL1.SubmitIdentityL1(r.Context(), tx)
	if err != nil {
		WriteError(w, http.StatusBadGateway, "L1_SUBMIT_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if txHash == "" {
		txHash = "accepted"
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"subject_did": req.SubjectDID,
		"tier":        req.Tier,
		"commitment":  commitment,
		"anchored_at": anchoredAt,
		"tx_hash":     txHash,
		"request_id":  r.Header.Get("X-Request-ID"),
	})
}
