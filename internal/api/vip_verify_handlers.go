package api

import (
	"context"
	"crypto/x509"
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

// Highest tier each kind of evidence can set through this endpoint.
//
// Anything other than a verified StoreKit transaction is self-asserted: the
// server has no proof behind it, so it can only record the baseline tier that
// silent provisioning gives every new account. Higher tiers come from evidence
// the server checks itself (the StoreKit subscription here; credential
// enrollment sets its own tier in handleEnrollmentDID).
const (
	vipMaxSelfAssertedTier = 1
)

// vipTxnOwner remembers which Apple transaction is bound to which DID so one
// subscription cannot be replayed onto a second account. It is the fallback
// when Redis is not configured (or is unreachable); with Redis the binding is
// shared by every gateway instance and survives a restart.
var (
	vipTxnMu    sync.Mutex
	vipTxnOwner = map[string]string{}
)

// vipTxnStore is the part of infra.RedisClient the replay guard needs.
type vipTxnStore interface {
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
	CacheGet(ctx context.Context, key string) ([]byte, error)
}

// claimVIPTransaction binds txnID to did, or reports that it already belongs
// to a different DID.
func (rt *Router) claimVIPTransaction(ctx context.Context, txnID, did string, expires time.Time) bool {
	if rt.Redis != nil {
		return claimVIPTransactionIn(ctx, rt.Redis, txnID, did, expires)
	}
	return claimVIPTransactionIn(ctx, nil, txnID, did, expires)
}

// claimVIPTransactionIn uses store when it is non-nil and reachable, and the
// process-local map otherwise. A Redis binding lives until the transaction
// expires plus a day of slack: after that the transaction is rejected as not
// current anyway.
func claimVIPTransactionIn(ctx context.Context, store vipTxnStore, txnID, did string, expires time.Time) bool {
	if store != nil {
		ttl := time.Until(expires) + 24*time.Hour
		if ttl < 24*time.Hour {
			ttl = 24 * time.Hour
		}
		// CacheGet reads under the "cache:" prefix; SetNX takes the full key.
		key := "vip:txn:" + txnID
		set, err := store.SetNX(ctx, "cache:"+key, []byte(did), ttl)
		if err == nil && set {
			return true
		}
		if err == nil {
			if owner, gerr := store.CacheGet(ctx, key); gerr == nil && owner != nil {
				return string(owner) == did
			}
		}
		// Redis unreachable (or the key expired between the two calls): fall
		// through to the process-local guard.
	}
	vipTxnMu.Lock()
	defer vipTxnMu.Unlock()
	if owner, ok := vipTxnOwner[txnID]; ok && owner != did {
		return false
	}
	vipTxnOwner[txnID] = did
	return true
}

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
// device key (X-Sender-DID / X-Signature), and may only set its own tier.
//
// The tier that is recorded depends on the evidence, not on what was asked for:
//
//   - evidence_type "vip_subscription" needs signed_transaction, a StoreKit 2
//     JWS that verifies against Apple Root CA - G3 (see verifyStoreKitJWS). It
//     records a VIP membership mark until the subscription's expiry. It does not raise the
//     trust tier.
//   - every other evidence_type is self-asserted and records at most
//     vipMaxSelfAssertedTier.
//
// A request for more than the evidence supports still returns 200, with the
// granted tier in trust_tier, the original ask in requested_trust_tier and
// capped=true, so first-run clients that report a locally computed tier keep
// working without being able to raise their own tier.
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
	maxTier := vipMaxSelfAssertedTier
	membership := MembershipNone
	var membershipExpires time.Time
	if req.EvidenceType == "vip_subscription" {
		var payload []byte
		var err error
		if storeKitUnverifiedAllowed() {
			// Local Xcode StoreKit testing only; never reachable in production.
			payload, err = unverifiedStoreKitPayload(req.SignedTransaction)
		} else {
			var roots *x509.CertPool
			roots, err = rt.storeKitRootPool()
			if err != nil {
				// Fail closed: without a trust anchor nothing can be verified.
				WriteError(w, http.StatusServiceUnavailable, "STOREKIT_UNAVAILABLE", "subscription verification is not available", r.Header.Get("X-Request-ID"))
				return
			}
			payload, err = verifyStoreKitJWS(req.SignedTransaction, roots, time.Now())
		}
		var claims storeKitClaims
		if err == nil {
			claims, err = parseStoreKitClaims(payload, time.Now())
		}
		if err != nil {
			WriteError(w, http.StatusBadRequest, "STOREKIT_REQUIRED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		if !rt.claimVIPTransaction(r.Context(), claims.TransactionID, req.DID, time.UnixMilli(claims.ExpiresDate)) {
			WriteError(w, http.StatusConflict, "TRANSACTION_REUSED", "this subscription is already linked to another account", r.Header.Get("X-Request-ID"))
			return
		}
		// Paying buys a membership mark, not trust. The trust tier is earned by verification
		// and a purchase never raises it.
		membership = MembershipVIP
		membershipExpires = time.UnixMilli(claims.ExpiresDate)
		rt.setMembership(r.Context(), req.DID, membership, membershipExpires)
	}
	requestedTier := req.TrustTier
	if req.TrustTier > maxTier {
		req.TrustTier = maxTier
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

	out := map[string]interface{}{
		"did":                  record.DID,
		"trust_tier":           record.TrustTier,
		"requested_trust_tier": requestedTier,
		"capped":               record.TrustTier < requestedTier,
		"evidence_type":        record.EvidenceType,
		"updated_at":           record.UpdatedAt.Format(time.RFC3339),
		"request_id":           r.Header.Get("X-Request-ID"),
	}
	if membership != MembershipNone {
		out["membership"] = membership
		out["membership_expires_at"] = membershipExpires.UTC().Format(time.RFC3339)
	}
	WriteJSON(w, http.StatusOK, out)
}

// --- Membership marks ---------------------------------------------------------------------
//
// A mark beside a name that says what kind of account this is. It is separate from the trust
// tier on purpose: trust is earned by verification, a mark can be paid for, and the two must
// never be mistaken for each other. Each mark has its own colour, and none uses the blue of the
// verified trust tiers.
//
//	vip           violet   an active ECHO VIP subscription. Says "this account pays", not "this
//	                       account is who it claims to be".
//	organization  gold     a business or organisation Echo has checked. Not issued yet.
//	official      grey     a government or public body Echo has checked. Not issued yet.

const (
	MembershipNone         = ""
	MembershipVIP          = "vip"
	MembershipOrganization = "organization"
	MembershipOfficial     = "official"
)

type membershipRecord struct {
	Mark      string    `json:"mark"`
	ExpiresAt time.Time `json:"expires_at"`
}

func membershipKey(did string) string { return "membership:" + did }

// setMembership records a mark until expires. Kept in Redis when configured, otherwise in memory.
func (rt *Router) setMembership(ctx context.Context, did, mark string, expires time.Time) {
	ttl := time.Until(expires)
	if did == "" || mark == MembershipNone || ttl <= 0 {
		return
	}
	rec := membershipRecord{Mark: mark, ExpiresAt: expires.UTC()}
	if rt.Redis != nil {
		data, _ := json.Marshal(rec)
		_ = rt.Redis.CacheSet(ctx, membershipKey(did), data, ttl)
		return
	}
	rt.memberships.Store(did, rec)
}

// membershipOf returns the account's current mark, or "" when it has none or it has lapsed.
func (rt *Router) membershipOf(ctx context.Context, did string) string {
	var rec membershipRecord
	if rt.Redis != nil {
		data, err := rt.Redis.CacheGet(ctx, membershipKey(did))
		if err != nil || len(data) == 0 || json.Unmarshal(data, &rec) != nil {
			return MembershipNone
		}
	} else {
		v, ok := rt.memberships.Load(did)
		if !ok {
			return MembershipNone
		}
		rec = v.(membershipRecord)
	}
	if !rec.ExpiresAt.After(time.Now()) {
		return MembershipNone
	}
	return rec.Mark
}

func (rt *Router) clearMembership(ctx context.Context, did string) {
	if rt.Redis != nil {
		_ = rt.Redis.CacheDelete(ctx, membershipKey(did))
	}
	rt.memberships.Delete(did)
}

// handleMembership returns the mark shown beside an account's name.
// GET /v1/membership?did=did:key:z… -> {"did":"…","mark":"vip"}
// A mark is public by nature (it is drawn next to the name for everyone who sees the account),
// so any signed-in user may ask.
func (rt *Router) handleMembership(w http.ResponseWriter, r *http.Request) {
	reqID := r.Header.Get("X-Request-ID")
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET required", reqID)
		return
	}
	did := r.URL.Query().Get("did")
	if did == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_FIELD", "did is required", reqID)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"did": did, "mark": rt.membershipOf(r.Context(), did)})
}

type storeKitClaims struct {
	TransactionID string `json:"transactionId"`
	BundleID      string `json:"bundleId"`
	ProductID     string `json:"productId"`
	ExpiresDate   int64  `json:"expiresDate"`
	// "Production", "Sandbox" (TestFlight, App Review, sandbox testers) or "Xcode".
	Environment string `json:"environment"`
	// Set by Apple when the purchase was refunded or revoked.
	RevocationDate int64 `json:"revocationDate"`
}

// parseStoreKitClaims reads the payload of a StoreKit 2 transaction whose
// signature has already been dealt with (verifyStoreKitJWS, or the explicit
// non-production bypass) and checks that it is the Echo VIP monthly
// subscription and still current.
func parseStoreKitClaims(payload []byte, now time.Time) (storeKitClaims, error) {
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
	if claims.RevocationDate != 0 {
		return storeKitClaims{}, errStoreKit("subscription was refunded or revoked")
	}
	// Apple signs Sandbox transactions with the same chain as real ones, and a
	// sandbox purchase is free: without this check any sandbox tester gets VIP
	// on the production gateway.
	if claims.Environment != "Production" && !storeKitSandboxAllowed() {
		return storeKitClaims{}, errStoreKit("transaction is not from the production App Store")
	}
	return claims, nil
}

type storeKitError string

func (e storeKitError) Error() string { return string(e) }

func errStoreKit(msg string) error { return storeKitError(msg) }
