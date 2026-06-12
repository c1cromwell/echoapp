package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type enrollmentMDLSession struct {
	ID              string
	Transport       string
	RequestedClaims enrollmentClaimsRequest
	ExpiresAt       time.Time
	VerifierBaseURL string
}

func (rt *Router) storeEnrollmentMDLSession(id string, sess enrollmentMDLSession) {
	rt.enrollmentMDLMu.Lock()
	defer rt.enrollmentMDLMu.Unlock()
	if rt.enrollmentMDLSessions == nil {
		rt.enrollmentMDLSessions = make(map[string]enrollmentMDLSession)
	}
	rt.enrollmentMDLSessions[id] = sess
}

func (rt *Router) loadEnrollmentMDLSession(id string) (enrollmentMDLSession, bool) {
	rt.enrollmentMDLMu.Lock()
	defer rt.enrollmentMDLMu.Unlock()
	s, ok := rt.enrollmentMDLSessions[id]
	if !ok || time.Now().After(s.ExpiresAt) {
		return enrollmentMDLSession{}, false
	}
	return s, true
}

func (rt *Router) deleteEnrollmentMDLSession(id string) {
	rt.enrollmentMDLMu.Lock()
	defer rt.enrollmentMDLMu.Unlock()
	delete(rt.enrollmentMDLSessions, id)
}

func (rt *Router) enrollmentVerifierBaseURL(r *http.Request) string {
	if base := strings.TrimSuffix(rt.OIDCVerifierBaseURL, "/"); base != "" {
		return base
	}
	if base := strings.TrimSuffix(os.Getenv("ENROLLMENT_VERIFIER_BASE_URL"), "/"); base != "" {
		return base
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// handleEnrollmentMDL routes POST /v1/enrollment/mdl/start|finish.
func (rt *Router) handleEnrollmentMDL(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v1/enrollment/mdl/start":
		rt.handleEnrollmentMDLStart(w, r)
	case "/v1/enrollment/mdl/finish":
		rt.handleEnrollmentMDLFinish(w, r)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown mDL enrollment path", r.Header.Get("X-Request-ID"))
	}
}

func (rt *Router) handleEnrollmentMDLUI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_SESSION", "session_id is required", r.Header.Get("X-Request-ID"))
		return
	}
	if _, ok := rt.loadEnrollmentMDLSession(sessionID); !ok {
		WriteError(w, http.StatusBadRequest, "SESSION_EXPIRED", "mDL session expired or unknown", r.Header.Get("X-Request-ID"))
		return
	}

	devDR := base64.StdEncoding.EncodeToString([]byte("dev-mdl-device-response"))
	devST := base64.StdEncoding.EncodeToString([]byte("dev-mdl-session-transcript"))
	callback := "echo-enroll://callback?session_id=" + url.QueryEscape(sessionID) +
		"&device_response_b64=" + url.QueryEscape(devDR) +
		"&session_transcript_b64=" + url.QueryEscape(devST)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><meta charset="utf-8"><title>ECHO mDL</title></head><body>
<h1>Present mobile driver's license</h1>
<p>On a supported device, the system Digital Credentials API will prompt Apple Wallet.</p>
` + devSection(callback) + `
</body></html>`))
}

func devSection(devCallback string) string {
	if !enrollmentDevModeEnabled() {
		return `<p>Dev completion is disabled. Set DEV_MODE=true for local mDL enrollment testing.</p>`
	}
	return `<p><strong>Dev only:</strong> <a href="` + devCallback + `">Simulate mDL presentation</a></p>`
}

type enrollmentMDLStartRequest struct {
	Transport       string                  `json:"transport"`
	RequestedClaims enrollmentClaimsRequest `json:"requested_claims"`
}

func (rt *Router) handleEnrollmentMDLStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req enrollmentMDLStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}
	transport := strings.TrimSpace(req.Transport)
	if transport == "" {
		transport = "web_dcapi"
	}

	sessionID := uuid.New().String()
	expiresAt := time.Now().Add(5 * time.Minute)
	base := rt.enrollmentVerifierBaseURL(r)
	rt.storeEnrollmentMDLSession(sessionID, enrollmentMDLSession{
		ID:              sessionID,
		Transport:       transport,
		RequestedClaims: req.RequestedClaims,
		ExpiresAt:       expiresAt,
		VerifierBaseURL: base,
	})

	verifierURL := base + "/v1/enrollment/mdl/ui?session_id=" + url.QueryEscape(sessionID)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":          sessionID,
		"verifierURL": verifierURL,
		"expiresAt":   expiresAt.Format(time.RFC3339),
		"request_id":  r.Header.Get("X-Request-ID"),
	})
}

type enrollmentMDLFinishRequest struct {
	SessionID            string `json:"session_id"`
	Transport            string `json:"transport"`
	DeviceResponseB64    string `json:"device_response_b64"`
	SessionTranscriptB64 string `json:"session_transcript_b64"`
	CallbackURL          string `json:"callback_url"`
}

func (rt *Router) handleEnrollmentMDLFinish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}

	var req enrollmentMDLFinishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	sessionID := req.SessionID
	deviceResponseB64 := req.DeviceResponseB64
	sessionTranscriptB64 := req.SessionTranscriptB64

	if req.CallbackURL != "" {
		parsed, err := url.Parse(req.CallbackURL)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_CALLBACK", "Malformed callback URL", r.Header.Get("X-Request-ID"))
			return
		}
		q := parsed.Query()
		if sessionID == "" {
			sessionID = q.Get("session_id")
		}
		if deviceResponseB64 == "" {
			deviceResponseB64 = q.Get("device_response_b64")
		}
		if sessionTranscriptB64 == "" {
			sessionTranscriptB64 = q.Get("session_transcript_b64")
		}
	}

	if sessionID == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_SESSION", "session_id is required", r.Header.Get("X-Request-ID"))
		return
	}
	sess, ok := rt.loadEnrollmentMDLSession(sessionID)
	if !ok {
		WriteError(w, http.StatusBadRequest, "SESSION_EXPIRED", "mDL session expired or unknown", r.Header.Get("X-Request-ID"))
		return
	}

	if deviceResponseB64 == "" || sessionTranscriptB64 == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_MDOC", "device_response_b64 and session_transcript_b64 are required", r.Header.Get("X-Request-ID"))
		return
	}
	if _, err := base64.StdEncoding.DecodeString(deviceResponseB64); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_MDOC", "device_response_b64 is not valid base64", r.Header.Get("X-Request-ID"))
		return
	}
	if _, err := base64.StdEncoding.DecodeString(sessionTranscriptB64); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_MDOC", "session_transcript_b64 is not valid base64", r.Header.Get("X-Request-ID"))
		return
	}

	rt.deleteEnrollmentMDLSession(sessionID)

	claims := disclosedClaimsFromEnrollmentRequest(sess.RequestedClaims)
	rt.issueVerifiedBundle(w, r, verifiedBundleResponse{
		CredentialType:       "org.iso.18013.5.1.mDL",
		AssuranceLevel:       "ial2",
		DisclosedClaims:      claims,
		EvidenceKind:         "mdoc",
		DeviceResponseB64:    deviceResponseB64,
		SessionTranscriptB64: sessionTranscriptB64,
	})
}

func disclosedClaimsFromEnrollmentRequest(claims enrollmentClaimsRequest) map[string]string {
	out := map[string]string{}
	if claims.GivenName {
		out["givenName"] = "DevGiven"
	}
	if claims.FamilyName {
		out["familyName"] = "DevFamily"
	}
	if claims.AgeOver18 {
		out["ageOver18"] = "true"
	}
	if claims.IssuingCountry {
		out["issuingCountry"] = "US"
	}
	return out
}
