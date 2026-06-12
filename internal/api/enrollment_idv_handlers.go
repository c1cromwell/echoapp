package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type enrollmentIDVSession struct {
	ID        string
	Provider  string
	Secret    string
	Status    string // pending | verified
	ExpiresAt time.Time
}

func (rt *Router) storeEnrollmentIDVSession(id string, sess enrollmentIDVSession) {
	rt.enrollmentIDVMu.Lock()
	defer rt.enrollmentIDVMu.Unlock()
	if rt.enrollmentIDVSessions == nil {
		rt.enrollmentIDVSessions = make(map[string]enrollmentIDVSession)
	}
	rt.enrollmentIDVSessions[id] = sess
}

func (rt *Router) loadEnrollmentIDVSession(id string) (enrollmentIDVSession, bool) {
	rt.enrollmentIDVMu.Lock()
	defer rt.enrollmentIDVMu.Unlock()
	s, ok := rt.enrollmentIDVSessions[id]
	if !ok || time.Now().After(s.ExpiresAt) {
		return enrollmentIDVSession{}, false
	}
	return s, true
}

func (rt *Router) updateEnrollmentIDVSession(id string, sess enrollmentIDVSession) {
	rt.storeEnrollmentIDVSession(id, sess)
}

func idvProviderName() string {
	if strings.TrimSpace(os.Getenv("STRIPE_IDENTITY_SECRET")) != "" {
		return "stripe_identity"
	}
	if strings.TrimSpace(os.Getenv("SUMSUB_APP_TOKEN")) != "" {
		return "sumsub"
	}
	return "dev_stub"
}

// handleEnrollmentIDV routes POST /v1/enrollment/idv/start|await.
func (rt *Router) handleEnrollmentIDV(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1/enrollment/idv/start":
		rt.handleEnrollmentIDVStart(w, r)
	case "/v1/enrollment/idv/await":
		rt.handleEnrollmentIDVAwait(w, r)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown IDV enrollment path", r.Header.Get("X-Request-ID"))
	}
}

func (rt *Router) handleEnrollmentIDVStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	provider := idvProviderName()
	if provider != "dev_stub" {
		WriteError(w, http.StatusNotImplemented, "IDV_PROVIDER_NOT_CONFIGURED",
			"IDV provider SDK integration is not wired on the server; configure STRIPE_IDENTITY_SECRET",
			r.Header.Get("X-Request-ID"))
		return
	}
	if os.Getenv("ENVIRONMENT") == "production" && !enrollmentDevModeEnabled() {
		WriteError(w, http.StatusServiceUnavailable, "IDV_UNAVAILABLE",
			"IDV enrollment requires STRIPE_IDENTITY_SECRET or SUMSUB_APP_TOKEN in production",
			r.Header.Get("X-Request-ID"))
		return
	}

	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(15 * time.Minute)
	rt.storeEnrollmentIDVSession(sessionID, enrollmentIDVSession{
		ID:        sessionID,
		Provider:  provider,
		Secret:    "dev_secret_" + sessionID,
		Status:    "pending",
		ExpiresAt: expiresAt,
	})

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":           sessionID,
		"clientSecret": "dev_secret_" + sessionID,
		"provider":     provider,
		"expires_at":   expiresAt.Format(time.RFC3339),
		"request_id":   r.Header.Get("X-Request-ID"),
	})
}

type enrollmentIDVAwaitRequest struct {
	SessionID string `json:"session_id"`
}

func (rt *Router) handleEnrollmentIDVAwait(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req enrollmentIDVAwaitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_SESSION", "session_id is required", r.Header.Get("X-Request-ID"))
		return
	}

	sess, ok := rt.loadEnrollmentIDVSession(req.SessionID)
	if !ok {
		WriteError(w, http.StatusBadRequest, "SESSION_EXPIRED", "IDV session expired or unknown", r.Header.Get("X-Request-ID"))
		return
	}

	if sess.Provider != "dev_stub" && !enrollmentDevModeEnabled() {
		WriteError(w, http.StatusNotImplemented, "IDV_AWAIT_UNSUPPORTED",
			"Server-side IDV await requires dev_stub or provider webhook integration",
			r.Header.Get("X-Request-ID"))
		return
	}

	sess.Status = "verified"
	rt.updateEnrollmentIDVSession(req.SessionID, sess)

	rt.issueVerifiedBundle(w, r, verifiedBundleResponse{
		CredentialType:      "IDVVerification",
		AssuranceLevel:      "ial2",
		DisclosedClaims:     map[string]string{"ageOver18": "true", "documentType": "drivers_license"},
		EvidenceKind:        "idv",
		ProviderReferenceID: "dev_idv_" + req.SessionID,
		ConfidenceScore:     0.97,
	})
}
