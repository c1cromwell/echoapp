package api

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	vipMonthlyProductID = "com.echo.app.vip.monthly"
	vipBundleID         = "com.echo.app"
)

// vipTransactions remembers which Apple transaction is bound to which DID so
// one subscription cannot be replayed onto a second account in this process.
var (
	vipTxnMu    sync.Mutex
	vipTxnOwner = map[string]string{}
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
// Authenticated: the caller must sign the request with the DID's registered
// device key (X-Sender-DID / X-Signature), and may only set its own tier. The
// tier is still self-asserted evidence until IDV / StoreKit validation backs it,
// so nothing should grant privileges from this record alone.
// The router stores the record; future messages/contact-resolution endpoints
// will read the tier from the DIDRegistry or this cache.
func (rt *Router) handleVIPVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "READ_BODY", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	var req struct {
		DID               string `json:"did"`
		TrustTier         int    `json:"trust_tier"`
		EvidenceType      string `json:"evidence_type"`
		SignedTransaction string `json:"signed_transaction"`
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
	// The caller may only record a tier for its own DID. This route used to be
	// unauthenticated, which let anyone set any DID's tier.
	if caller, _ := r.Context().Value(ContextKeyUserID).(string); caller == "" || caller != req.DID {
		WriteError(w, http.StatusForbidden, "DID_MISMATCH", "did must match the authenticated caller", r.Header.Get("X-Request-ID"))
		return
	}
	if req.TrustTier < 1 || req.TrustTier > 5 {
		WriteError(w, http.StatusBadRequest, "INVALID_TIER", "trust_tier must be 1–5", r.Header.Get("X-Request-ID"))
		return
	}
	if req.EvidenceType == "" {
		req.EvidenceType = "unknown"
	}
	if req.EvidenceType == "vip_subscription" {
		claims, err := parseStoreKitTransaction(req.SignedTransaction, time.Now())
		if err != nil {
			WriteError(w, http.StatusBadRequest, "STOREKIT_REQUIRED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		vipTxnMu.Lock()
		if owner, ok := vipTxnOwner[claims.TransactionID]; ok && owner != req.DID {
			vipTxnMu.Unlock()
			WriteError(w, http.StatusConflict, "TRANSACTION_REUSED", "this subscription is already linked to another account", r.Header.Get("X-Request-ID"))
			return
		}
		vipTxnOwner[claims.TransactionID] = req.DID
		vipTxnMu.Unlock()
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
		"did":           record.DID,
		"trust_tier":    record.TrustTier,
		"evidence_type": record.EvidenceType,
		"updated_at":    record.UpdatedAt.Format(time.RFC3339),
		"request_id":    r.Header.Get("X-Request-ID"),
	})
}

type storeKitClaims struct {
	TransactionID string `json:"transactionId"`
	BundleID      string `json:"bundleId"`
	ProductID     string `json:"productId"`
	ExpiresDate   int64  `json:"expiresDate"`
}

// parseStoreKitTransaction reads the payload of a StoreKit 2 signed transaction
// and checks that it is the Echo VIP monthly subscription and still current.
// Xcode's local StoreKit signer is not Apple's production certificate, so the
// signature chain is not pinned here.
func parseStoreKitTransaction(jws string, now time.Time) (storeKitClaims, error) {
	parts := strings.Split(jws, ".")
	if len(parts) != 3 || parts[1] == "" {
		return storeKitClaims{}, errStoreKit("signed_transaction must be a StoreKit transaction")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return storeKitClaims{}, errStoreKit("signed_transaction payload is not valid")
	}
	var claims storeKitClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return storeKitClaims{}, errStoreKit("signed_transaction payload is not valid")
	}
	if claims.ProductID != vipMonthlyProductID || claims.BundleID != vipBundleID || claims.TransactionID == "" {
		return storeKitClaims{}, errStoreKit("transaction is not an ECHO VIP subscription")
	}
	if claims.ExpiresDate <= now.UnixMilli() {
		return storeKitClaims{}, errStoreKit("subscription is not current")
	}
	return claims, nil
}

type storeKitError string

func (e storeKitError) Error() string { return string(e) }

func errStoreKit(msg string) error { return storeKitError(msg) }
