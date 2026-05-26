package api

// Wave 12: SMS OTP recovery endpoints.
//
// Privacy model: the backend stores only H(E.164 phone) — never the raw number.
// Three endpoints:
//
//   POST /v1/auth/sms-recovery/register  — register phone commitment during onboarding
//   POST /v1/auth/sms-recovery/verify    — verify the OTP (single-use)
//   POST /v1/auth/sms-recovery/challenge — initiate recovery (find DID by phone hash, send new OTP)

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/infra"
)

// e164Regexp matches E.164 phone numbers: +[1-9][0-9]{9,14}
var e164Regexp = regexp.MustCompile(`^\+[1-9]\d{9,14}$`)

// smsOTPSession is stored in Redis keyed by the session token.
type smsOTPSession struct {
	PhoneHash string    `json:"phone_hash"`
	OTP       string    `json:"otp"`
	DID       string    `json:"did,omitempty"`
	ExpiresAt time.Time `json:"expires_at"`
	Attempts  int       `json:"attempts"`           // failed verification count (S6 brute-force lockout)
	OPRFKey   string    `json:"oprf_key,omitempty"` // D2: hex(OPRF_k(phone)), committed to the discovery index on verify
}

// --- POST /v1/auth/sms-recovery/register ---

// SMSRecoveryRegisterRequest is the JSON body for the register endpoint.
type SMSRecoveryRegisterRequest struct {
	PhoneHash string `json:"phone_hash"` // "sha256:" + lowercase hex
	PhoneRaw  string `json:"phone_raw"`  // E.164 — used only for OTP send, not stored
	DID       string `json:"did"`
}

// handleSMSRecoveryRegister stores the phone commitment and dispatches an OTP.
func (rt *Router) handleSMSRecoveryRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req SMSRecoveryRegisterRequest
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &req); err != nil || req.PhoneHash == "" || req.PhoneRaw == "" || req.DID == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "phone_hash, phone_raw, and did are required", r.Header.Get("X-Request-ID"))
		return
	}
	if !e164Regexp.MatchString(req.PhoneRaw) {
		WriteError(w, http.StatusBadRequest, "INVALID_PHONE", "phone_raw must be E.164 format (+12125551234)", r.Header.Get("X-Request-ID"))
		return
	}

	// Verify the caller's claimed hash matches the raw phone.
	expectedHash := phoneHash(req.PhoneRaw)
	if !constantTimeStringEqual(req.PhoneHash, expectedHash) {
		WriteError(w, http.StatusBadRequest, "HASH_MISMATCH", "phone_hash does not match sha256(phone_raw)", r.Header.Get("X-Request-ID"))
		return
	}

	if rt.V3 == nil || rt.V3.DB == nil {
		WriteError(w, http.StatusServiceUnavailable, "DB_NOT_CONFIGURED", "database unavailable", r.Header.Get("X-Request-ID"))
		return
	}
	if err := rt.V3.DB.SetSMSRecovery(r.Context(), req.DID, expectedHash); err != nil {
		WriteError(w, http.StatusInternalServerError, "DB_ERROR", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	sessionToken, otp, err := rt.dispatchOTP(r.Context(), req.PhoneRaw, req.PhoneHash, req.DID)
	if err != nil {
		if errors.Is(err, errOTPSendRateLimited) {
			WriteError(w, http.StatusTooManyRequests, "OTP_RATE_LIMITED", "too many codes requested for this number; try again later", r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusInternalServerError, "OTP_SEND_FAILED", "failed to send OTP", r.Header.Get("X-Request-ID"))
		return
	}

	resp := map[string]interface{}{
		"session_token": sessionToken,
		"expires_in":    int(infra.SMSOTPSessionTTL.Seconds()),
	}
	// In dev (and only with the explicit ALLOW_DEV_OTP opt-in), echo the OTP in
	// the body so tests don't need an SMS provider. Never via a response header.
	if devOTPEnabled() {
		resp["_dev_otp"] = otp
	}
	WriteJSON(w, http.StatusOK, resp)
}

// --- POST /v1/auth/sms-recovery/verify ---

type SMSRecoveryVerifyRequest struct {
	SessionToken string `json:"session_token"`
	OTP          string `json:"otp"`
}

// handleSMSRecoveryVerify verifies the OTP and confirms the phone commitment.
func (rt *Router) handleSMSRecoveryVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req SMSRecoveryVerifyRequest
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &req); err != nil || req.SessionToken == "" || req.OTP == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "session_token and otp are required", r.Header.Get("X-Request-ID"))
		return
	}

	session, err := rt.getSMSSession(r.Context(), req.SessionToken)
	if err != nil {
		WriteError(w, http.StatusUnauthorized, "INVALID_SESSION", "OTP session not found or expired", r.Header.Get("X-Request-ID"))
		return
	}

	// S6: once the attempt budget is exhausted, lock the session for its lifetime
	// (a new code must be requested). The session is kept, not deleted, so the
	// lockout persists rather than degrading to "session not found".
	if session.Attempts >= maxOTPAttempts {
		WriteError(w, http.StatusTooManyRequests, "OTP_LOCKED", "too many incorrect attempts; request a new code", r.Header.Get("X-Request-ID"))
		return
	}

	if !constantTimeStringEqual(req.OTP, session.OTP) {
		session.Attempts++
		_ = rt.putSMSSession(r.Context(), req.SessionToken, *session)
		if session.Attempts >= maxOTPAttempts {
			WriteError(w, http.StatusTooManyRequests, "OTP_LOCKED", "too many incorrect attempts; request a new code", r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusUnauthorized, "INVALID_OTP", "OTP does not match", r.Header.Get("X-Request-ID"))
		return
	}

	// D2: phone ownership is now proven — make it discoverable (OPRF key -> DID).
	if session.OPRFKey != "" && rt.V3 != nil && rt.V3.Contacts != nil {
		if err := rt.V3.Contacts.CommitDiscoveryKey(r.Context(), session.OPRFKey, session.DID); err != nil {
			log.Printf("contacts: commit discovery key failed for %s: %v", session.DID, err)
		}
	}

	// Single-use: delete session immediately after successful verification.
	rt.deleteSMSSession(r.Context(), req.SessionToken)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"verified": true,
		"did":      session.DID,
	})
}

// --- POST /v1/auth/sms-recovery/challenge ---

type SMSRecoveryChallengeRequest struct {
	PhoneHash string `json:"phone_hash"`
	PhoneRaw  string `json:"phone_raw"`
}

// handleSMSRecoveryChallenge initiates account recovery: finds the DID by phone
// hash and sends a new OTP to the caller's phone.
func (rt *Router) handleSMSRecoveryChallenge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req SMSRecoveryChallengeRequest
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, &req); err != nil || req.PhoneHash == "" || req.PhoneRaw == "" {
		WriteError(w, http.StatusBadRequest, "INVALID_BODY", "phone_hash and phone_raw are required", r.Header.Get("X-Request-ID"))
		return
	}
	if !e164Regexp.MatchString(req.PhoneRaw) {
		WriteError(w, http.StatusBadRequest, "INVALID_PHONE", "phone_raw must be E.164 format", r.Header.Get("X-Request-ID"))
		return
	}

	expectedHash := phoneHash(req.PhoneRaw)
	if !constantTimeStringEqual(req.PhoneHash, expectedHash) {
		WriteError(w, http.StatusBadRequest, "HASH_MISMATCH", "phone_hash does not match sha256(phone_raw)", r.Header.Get("X-Request-ID"))
		return
	}

	if rt.V3 == nil || rt.V3.DB == nil {
		WriteError(w, http.StatusServiceUnavailable, "DB_NOT_CONFIGURED", "database unavailable", r.Header.Get("X-Request-ID"))
		return
	}
	did, err := rt.V3.DB.GetSMSRecoveryByPhoneHash(r.Context(), expectedHash)
	if err != nil {
		// Do not reveal whether the phone is registered (privacy).
		WriteError(w, http.StatusNotFound, "PHONE_NOT_REGISTERED", "phone number is not registered for recovery", r.Header.Get("X-Request-ID"))
		return
	}

	sessionToken, otp, err := rt.dispatchOTP(r.Context(), req.PhoneRaw, expectedHash, did)
	if err != nil {
		if errors.Is(err, errOTPSendRateLimited) {
			WriteError(w, http.StatusTooManyRequests, "OTP_RATE_LIMITED", "too many codes requested for this number; try again later", r.Header.Get("X-Request-ID"))
			return
		}
		WriteError(w, http.StatusInternalServerError, "OTP_SEND_FAILED", "failed to send OTP", r.Header.Get("X-Request-ID"))
		return
	}

	resp := map[string]interface{}{
		"session_token": sessionToken,
		"did":           did,
		"expires_in":    int(infra.SMSOTPSessionTTL.Seconds()),
	}
	if devOTPEnabled() {
		resp["_dev_otp"] = otp
	}
	WriteJSON(w, http.StatusOK, resp)
}

// --- helpers ---

// phoneHash returns "sha256:" + lowercase hex of SHA-256(E.164 phone).
func phoneHash(e164 string) string {
	h := sha256.Sum256([]byte(e164))
	return "sha256:" + hex.EncodeToString(h[:])
}

// constantTimeStringEqual compares two strings in constant time.
func constantTimeStringEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

const (
	// maxOTPAttempts is the number of failed verifications allowed per OTP
	// session before it is invalidated (S6 brute-force lockout).
	maxOTPAttempts = 3
	// maxOTPSendsPerWindow / otpSendWindow bound how many codes may be requested
	// for a single phone number, preventing SMS flooding (S9).
	maxOTPSendsPerWindow = 3
	otpSendWindow        = 15 * time.Minute
	// otpSendAction is the PublicRateLimiter action key for per-phone send limits.
	otpSendAction = "otp_send"
)

// errOTPSendRateLimited signals that a phone number has requested too many codes.
var errOTPSendRateLimited = errors.New("otp send rate limited")

// dispatchOTP generates a 6-digit OTP, stores the session, sends it via the
// configured SMS provider, and returns the session token. Sends per phone number
// are rate limited (S9) when a limiter is configured.
func (rt *Router) dispatchOTP(ctx context.Context, phoneRaw, phoneHash, did string) (sessionToken, otp string, err error) {
	// S9: cap codes requested per phone number within the window.
	if rt.PublicRateLimiter != nil {
		if rerr := rt.PublicRateLimiter.Check(phoneHash, otpSendAction); rerr != nil {
			return "", "", errOTPSendRateLimited
		}
	}

	otp, err = generateOTP()
	if err != nil {
		return "", "", fmt.Errorf("generate OTP: %w", err)
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", "", fmt.Errorf("generate session token: %w", err)
	}
	sessionToken = hex.EncodeToString(tokenBytes)

	session := smsOTPSession{
		PhoneHash: phoneHash,
		OTP:       otp,
		DID:       did,
		ExpiresAt: time.Now().Add(infra.SMSOTPSessionTTL),
	}
	// D2: precompute the OPRF discovery key from the raw number now (it's never
	// stored raw); commit it to the discovery index only once ownership is
	// verified (handleSMSRecoveryVerify).
	if rt.V3 != nil && rt.V3.Contacts != nil {
		if key, kerr := rt.V3.Contacts.DiscoveryKey(phoneRaw); kerr == nil {
			session.OPRFKey = key
		}
	}
	if err := rt.putSMSSession(ctx, sessionToken, session); err != nil {
		return "", "", fmt.Errorf("store OTP session: %w", err)
	}

	// Send SMS — raw phone used only here, never stored.
	msg := fmt.Sprintf("Your Echo verification code is %s. Valid for 5 minutes. Never share this code.", otp)
	if rt.SMSProvider != nil {
		_ = rt.SMSProvider.Send(phoneRaw, msg) // non-fatal: log in prod
	}

	return sessionToken, otp, nil
}

// putSMSSession persists an OTP session: Redis when configured, otherwise an
// in-memory map (dev/test). Storing in-memory keeps the recovery flow usable
// without Redis.
func (rt *Router) putSMSSession(ctx context.Context, sessionToken string, s smsOTPSession) error {
	if rt.Redis != nil {
		raw, _ := json.Marshal(s)
		return rt.Redis.SetSMSOTPSession(ctx, sessionToken, raw)
	}
	rt.smsSessionsMu.Lock()
	defer rt.smsSessionsMu.Unlock()
	if rt.smsSessions == nil {
		rt.smsSessions = make(map[string]smsOTPSession)
	}
	rt.smsSessions[sessionToken] = s
	return nil
}

// deleteSMSSession removes an OTP session (single-use / lockout).
func (rt *Router) deleteSMSSession(ctx context.Context, sessionToken string) {
	if rt.Redis != nil {
		_ = rt.Redis.DeleteSMSOTPSession(ctx, sessionToken)
		return
	}
	rt.smsSessionsMu.Lock()
	defer rt.smsSessionsMu.Unlock()
	delete(rt.smsSessions, sessionToken)
}

// getSMSSession retrieves and deserializes an OTP session (Redis or in-memory).
func (rt *Router) getSMSSession(ctx context.Context, sessionToken string) (*smsOTPSession, error) {
	if rt.Redis != nil {
		raw, err := rt.Redis.GetSMSOTPSession(ctx, sessionToken)
		if err != nil {
			return nil, err
		}
		var session smsOTPSession
		if err := json.Unmarshal(raw, &session); err != nil {
			return nil, fmt.Errorf("malformed session: %w", err)
		}
		return &session, nil
	}

	rt.smsSessionsMu.Lock()
	defer rt.smsSessionsMu.Unlock()
	session, ok := rt.smsSessions[sessionToken]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	if time.Now().After(session.ExpiresAt) {
		delete(rt.smsSessions, sessionToken)
		return nil, fmt.Errorf("session expired")
	}
	return &session, nil
}

// generateOTP generates a cryptographically random 6-digit OTP.
func generateOTP() (string, error) {
	max := big.NewInt(1_000_000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

var devOTPWarnOnce sync.Once

// devOTPEnabled reports whether the OTP may be echoed in the response body for
// local testing. It requires BOTH DEV_MODE and ALLOW_DEV_OTP to be truthy and is
// hard-disabled when ENVIRONMENT=production — so a single mis-set flag (or a
// forgotten DEV_MODE) cannot leak OTPs in production. The OTP is never returned
// in a response header (headers leak to proxies, logs, and browser history).
func devOTPEnabled() bool {
	if os.Getenv("ENVIRONMENT") == "production" {
		return false
	}
	if !envTruthy("DEV_MODE") || !envTruthy("ALLOW_DEV_OTP") {
		return false
	}
	devOTPWarnOnce.Do(func() {
		log.Println("⚠ security: dev OTP echo ENABLED (DEV_MODE+ALLOW_DEV_OTP) — OTPs are returned in API responses; never enable in production")
	})
	return true
}

func envTruthy(name string) bool {
	v := os.Getenv(name)
	return v == "true" || v == "1"
}
