package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// vipTierRecord is an in-memory store for trust tiers. A Postgres/Redis
// backing store can be wired up when infrastructure is available.
type vipTierRecord struct {
	DID          string    `json:"did"`
	TrustTier    int       `json:"trust_tier"`
	EvidenceType string    `json:"evidence_type"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// handleVIPVerify records a completed VIP identity verification.
//
//	POST /v1/auth/vip-verify
//	Body: { "did": "did:key:z…", "trust_tier": 2, "evidence_type": "standard_idv" }
//
// Public path — the DID itself is the identity claim; no separate JWT needed
// because Phase 1 devices reach this immediately after key enrollment.
// The router stores the record; future messages/contact-resolution endpoints
// will read the tier from the DIDRegistry or this cache.
func (rt *Router) handleVIPVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 4096))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "READ_BODY", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		DID          string `json:"did"`
		TrustTier    int    `json:"trust_tier"`
		EvidenceType string `json:"evidence_type"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	req.DID = strings.TrimSpace(req.DID)
	if req.DID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_DID", "did is required", r.Header.Get("X-Request-ID"))
		return
	}
	if !strings.HasPrefix(req.DID, "did:") {
		WriteError(w, http.StatusBadRequest, "INVALID_DID", "did must be a valid DID string", r.Header.Get("X-Request-ID"))
		return
	}
	if req.TrustTier < 1 || req.TrustTier > 5 {
		WriteError(w, http.StatusBadRequest, "INVALID_TIER", "trust_tier must be 1–5", r.Header.Get("X-Request-ID"))
		return
	}
	if req.EvidenceType == "" {
		req.EvidenceType = "unknown"
	}

	record := vipTierRecord{
		DID:          req.DID,
		TrustTier:    req.TrustTier,
		EvidenceType: req.EvidenceType,
		UpdatedAt:    time.Now().UTC(),
	}

	// Persist to Redis if available; fall through silently if not configured.
	if rt.Redis != nil {
		data, _ := json.Marshal(record)
		_ = rt.Redis.CacheSet(r.Context(), "vip:tier:"+req.DID, data, 90*24*time.Hour)
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"did":          record.DID,
		"trust_tier":   record.TrustTier,
		"evidence_type": record.EvidenceType,
		"updated_at":   record.UpdatedAt.Format(time.RFC3339),
		"request_id":   r.Header.Get("X-Request-ID"),
	})
}
