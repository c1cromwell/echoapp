package api

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// verifiedBundleResponse is the wire shape consumed by EnrollmentAPIClient.
type verifiedBundleResponse struct {
	CredentialReferenceUUID string
	IssuerDID               string
	CredentialType          string
	AssuranceLevel          string
	DisclosedClaims         map[string]string
	EvidenceKind            string
	DeviceResponseB64       string
	SessionTranscriptB64    string
	VPToken                 string
	PresentationSubmission  []byte
	ProviderReferenceID     string
	ConfidenceScore         float64
	HolderDID               string
}

func (rt *Router) issueVerifiedBundle(w http.ResponseWriter, r *http.Request, in verifiedBundleResponse) {
	reqID := r.Header.Get("X-Request-ID")
	credRef := uuid.New().String()
	holderDID := in.HolderDID
	if holderDID == "" {
		holderDID = "did:key:holder:" + credRef[:8]
	}

	rt.storeEnrollmentVerified(credRef, enrollmentVerifiedRecord{
		HolderDID:      holderDID,
		AssuranceLevel: in.AssuranceLevel,
		CredentialType: in.CredentialType,
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	})

	evidence := map[string]interface{}{"kind": in.EvidenceKind}
	switch in.EvidenceKind {
	case "mdoc":
		evidence["device_response_b64"] = in.DeviceResponseB64
		evidence["session_transcript_b64"] = in.SessionTranscriptB64
	case "openid4vp":
		evidence["vp_token"] = in.VPToken
		if len(in.PresentationSubmission) > 0 {
			evidence["presentation_submission_b64"] = base64.StdEncoding.EncodeToString(in.PresentationSubmission)
		}
	case "idv":
		evidence["provider_reference_id"] = in.ProviderReferenceID
		evidence["confidence_score"] = in.ConfidenceScore
	}

	claims := in.DisclosedClaims
	if claims == nil {
		claims = map[string]string{}
	}
	if _, ok := claims["holder_did"]; !ok {
		claims["holder_did"] = holderDID
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"credential_reference_uuid": credRef,
		"issuer_did":                defaultEnrollmentIssuerDID(rt),
		"credential_type":           in.CredentialType,
		"issued_at":                 time.Now().UTC().Format(time.RFC3339),
		"revocation_index":          nil,
		"assurance_level":           in.AssuranceLevel,
		"disclosed_claims":          claims,
		"evidence":                  evidence,
		"request_id":                reqID,
	})
}

func defaultEnrollmentIssuerDID(rt *Router) string {
	if rt.OIDCVerifier != nil {
		return rt.OIDCVerifier.Metadata().VerifierID
	}
	return "did:key:echo-enrollment-issuer"
}

func enrollmentDevModeEnabled() bool {
	return envTruthy("DEV_MODE") || envTruthy("ALLOW_DEV_ENROLLMENT")
}
