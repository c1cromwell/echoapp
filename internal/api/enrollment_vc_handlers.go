package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/pkg/credentials/oidc4vc"
)

type enrollmentVCSession struct {
	State          string
	CredentialType string
	RedirectURI    string
	ExpiresAt      time.Time
}

type enrollmentClaimsRequest struct {
	FamilyName     bool `json:"familyName"`
	GivenName      bool `json:"givenName"`
	AgeOver18      bool `json:"ageOver18"`
	IssuingCountry bool `json:"issuingCountry"`
	Portrait       bool `json:"portrait"`
}

type enrollmentVCStartRequest struct {
	RequestedClaims enrollmentClaimsRequest `json:"requested_claims"`
}

type enrollmentVCFinishRequest struct {
	SessionID   string                          `json:"session_id"`
	CallbackURL string                          `json:"callback_url,omitempty"`
	VPToken     string                          `json:"vp_token,omitempty"`
	State       string                          `json:"state,omitempty"`
	Submission  *oidc4vc.PresentationSubmission `json:"presentation_submission,omitempty"`
}

func (rt *Router) handleEnrollmentVC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	switch r.URL.Path {
	case "/v1/enrollment/vc/start":
		rt.handleEnrollmentVCStart(w, r)
	case "/v1/enrollment/vc/finish":
		rt.handleEnrollmentVCFinish(w, r)
	default:
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown enrollment path", r.Header.Get("X-Request-ID"))
	}
}

func (rt *Router) handleEnrollmentVCStart(w http.ResponseWriter, r *http.Request) {
	if rt.OIDCVerifier == nil {
		WriteError(w, http.StatusServiceUnavailable, "OIDC4VC_DISABLED", "OIDC4VC verifier is not enabled", r.Header.Get("X-Request-ID"))
		return
	}

	var req enrollmentVCStartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	credentialType := mapEnrollmentClaimsToCredentialType(req.RequestedClaims)
	redirectURI := "echo-enroll://callback"
	presReq, err := rt.OIDCVerifier.BeginPresentation(credentialType, redirectURI)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "PRESENTATION_REQUEST_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	expiresAt := time.Now().Add(5 * time.Minute)
	rt.storeEnrollmentVCSession(presReq.State, enrollmentVCSession{
		State:          presReq.State,
		CredentialType: credentialType,
		RedirectURI:    redirectURI,
		ExpiresAt:      expiresAt,
	})

	base := strings.TrimSuffix(rt.OIDCVerifierBaseURL, "/")
	verifierURL := base + "/verification/ui?state=" + url.QueryEscape(presReq.State) +
		"&credential_type=" + url.QueryEscape(credentialType)

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":                   presReq.State,
		"verifierURL":          verifierURL,
		"expiresAt":            expiresAt.Format(time.RFC3339),
		"presentation_request": presReq,
		"request_id":           r.Header.Get("X-Request-ID"),
	})
}

func (rt *Router) handleEnrollmentVCFinish(w http.ResponseWriter, r *http.Request) {
	if rt.OIDCVerifier == nil {
		WriteError(w, http.StatusServiceUnavailable, "OIDC4VC_DISABLED", "OIDC4VC verifier is not enabled", r.Header.Get("X-Request-ID"))
		return
	}

	var req enrollmentVCFinishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	vpToken := req.VPToken
	state := req.State
	submission := req.Submission

	if req.CallbackURL != "" {
		parsed, err := url.Parse(req.CallbackURL)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_CALLBACK", "Malformed callback URL", r.Header.Get("X-Request-ID"))
			return
		}
		q := parsed.Query()
		if vpToken == "" {
			vpToken = q.Get("vp_token")
		}
		if state == "" {
			state = q.Get("state")
		}
	}

	if state == "" && req.SessionID != "" {
		state = req.SessionID
	}
	if vpToken == "" || state == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_VP", "vp_token and state are required", r.Header.Get("X-Request-ID"))
		return
	}

	sess, ok := rt.loadEnrollmentVCSession(state)
	if !ok || time.Now().After(sess.ExpiresAt) {
		WriteError(w, http.StatusBadRequest, "SESSION_EXPIRED", "Enrollment session expired or unknown", r.Header.Get("X-Request-ID"))
		return
	}

	if submission == nil {
		submission = defaultPresentationSubmission(sess.CredentialType)
	}

	result, err := rt.OIDCVerifier.AcceptPresentation(r.Context(), vpToken, submission, state)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "VERIFICATION_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	if !result.IsValid {
		msg := result.Error
		if msg == "" {
			msg = "credential verification failed"
		}
		WriteError(w, http.StatusBadRequest, "VERIFICATION_FAILED", msg, r.Header.Get("X-Request-ID"))
		return
	}

	rt.deleteEnrollmentVCSession(state)

	credRef := uuid.New().String()
	rt.storeEnrollmentVerified(credRef, enrollmentVerifiedRecord{
		HolderDID:      result.HolderDID,
		AssuranceLevel: mapCredentialTypeToIAL(sess.CredentialType),
		CredentialType: sess.CredentialType,
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	})

	submissionJSON, _ := json.Marshal(submission)
	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"credential_reference_uuid": credRef,
		"issuer_did":                rt.verifierDID(),
		"credential_type":           sess.CredentialType,
		"issued_at":                 time.Now().UTC().Format(time.RFC3339),
		"revocation_index":          nil,
		"assurance_level":           mapCredentialTypeToIAL(sess.CredentialType),
		"disclosed_claims":          map[string]string{"holder_did": result.HolderDID},
		"evidence": map[string]interface{}{
			"kind":                        "openid4vp",
			"vp_token":                    vpToken,
			"presentation_submission_b64": base64.StdEncoding.EncodeToString(submissionJSON),
		},
		"request_id": r.Header.Get("X-Request-ID"),
	})
}

func mapEnrollmentClaimsToCredentialType(claims enrollmentClaimsRequest) string {
	if claims.AgeOver18 && (claims.GivenName || claims.FamilyName) {
		return "KYCLite"
	}
	return "ProofOfHumanity"
}

func mapCredentialTypeToIAL(credentialType string) string {
	switch credentialType {
	case "HighAssurance", "Professional":
		return "ial3"
	case "KYCLite":
		return "ial2"
	default:
		return "ial1"
	}
}

func defaultPresentationSubmission(credentialType string) *oidc4vc.PresentationSubmission {
	return &oidc4vc.PresentationSubmission{
		ID:           uuid.New().String(),
		DefinitionID: "pres_def_" + strings.ToLower(credentialType),
		DescriptorMap: []oidc4vc.DescriptorMap{
			{
				ID:     "credential_0",
				Format: "jwt_vc_json",
				Path:   "$.vp.verifiableCredential[0]",
			},
		},
	}
}

func (rt *Router) verifierDID() string {
	if rt.OIDCVerifier != nil {
		return rt.OIDCVerifier.Metadata().VerifierID
	}
	return "did:key:verifier"
}

func (rt *Router) storeEnrollmentVCSession(state string, sess enrollmentVCSession) {
	rt.enrollmentVCMu.Lock()
	defer rt.enrollmentVCMu.Unlock()
	if rt.enrollmentVCSessions == nil {
		rt.enrollmentVCSessions = make(map[string]enrollmentVCSession)
	}
	rt.enrollmentVCSessions[state] = sess
}

func (rt *Router) loadEnrollmentVCSession(state string) (enrollmentVCSession, bool) {
	rt.enrollmentVCMu.Lock()
	defer rt.enrollmentVCMu.Unlock()
	s, ok := rt.enrollmentVCSessions[state]
	return s, ok
}

func (rt *Router) deleteEnrollmentVCSession(state string) {
	rt.enrollmentVCMu.Lock()
	defer rt.enrollmentVCMu.Unlock()
	delete(rt.enrollmentVCSessions, state)
}
