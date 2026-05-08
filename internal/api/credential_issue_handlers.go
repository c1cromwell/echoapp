package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/thechadcromwell/echoapp/pkg/credentials"
)

// handleIdentityCredentials serves GET/POST /identity/credentials (WO-273).
func (rt *Router) handleIdentityCredentials(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		rt.handleIdentityCredentialsPost(w, r)
	case http.MethodGet:
		rt.handleIdentityCredentialsList(w, r)
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET and POST are allowed", r.Header.Get("X-Request-ID"))
	}
}

func userIDFromContext(r *http.Request) string {
	if v := r.Context().Value(ContextKeyUserID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func parseCredentialTypeFromSchemaID(schemaID string) (credentials.CredentialType, error) {
	switch strings.TrimSpace(schemaID) {
	case string(credentials.ProofOfHumanity):
		return credentials.ProofOfHumanity, nil
	case string(credentials.KYCLite):
		return credentials.KYCLite, nil
	case string(credentials.HighAssurance):
		return credentials.HighAssurance, nil
	case string(credentials.Professional):
		return credentials.Professional, nil
	default:
		return "", credentials.NewCredentialError(credentials.ErrCodeInvalidRequest, "unsupported schemaId / credentialType")
	}
}

func parseCredentialFormat(s string) credentials.CredentialFormat {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "jwt":
		return credentials.JWTFormat
	case "sd-jwt", "sdjwt":
		return credentials.SDJWTFormat
	default:
		return credentials.JSONLDFormat
	}
}

func (rt *Router) handleIdentityCredentialsPost(w http.ResponseWriter, r *http.Request) {
	if rt.CredentialService == nil {
		WriteError(w, http.StatusServiceUnavailable, "CREDENTIALS_NOT_CONFIGURED", "credential issuance is not enabled on this deployment", r.Header.Get("X-Request-ID"))
		return
	}

	principal := userIDFromContext(r)
	if principal == "" {
		WriteError(w, http.StatusUnauthorized, "MISSING_PRINCIPAL", "authenticated subject required", r.Header.Get("X-Request-ID"))
		return
	}

	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		WriteError(w, http.StatusBadRequest, "READ_BODY", "failed to read body", r.Header.Get("X-Request-ID"))
		return
	}

	var body struct {
		SchemaID           string                          `json:"schemaId"`
		SubjectDID         string                          `json:"subjectDid,omitempty"`
		Claims             map[string]interface{}          `json:"claims"`
		VerificationClaims []credentials.VerificationClaim `json:"verificationClaims"`
		PreferredFormat    string                          `json:"preferredFormat,omitempty"`
		ExpirationYears    int                             `json:"expirationYears,omitempty"`
		RawCredentialType  string                          `json:"credentialType,omitempty"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		WriteError(w, http.StatusBadRequest, "MALFORMED_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
		return
	}

	schema := strings.TrimSpace(body.SchemaID)
	if schema == "" {
		schema = strings.TrimSpace(body.RawCredentialType)
	}
	if schema == "" {
		WriteError(w, http.StatusBadRequest, "MISSING_SCHEMA", "schemaId is required", r.Header.Get("X-Request-ID"))
		return
	}

	credType, err := parseCredentialTypeFromSchemaID(schema)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_SCHEMA", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	subject := strings.TrimSpace(body.SubjectDID)
	if subject == "" {
		subject = principal
	}
	if subject != principal {
		WriteError(w, http.StatusForbidden, "SUBJECT_MISMATCH", "subjectDid must match the authenticated principal", r.Header.Get("X-Request-ID"))
		return
	}

	req := &credentials.CredentialIssuanceRequest{
		SubjectDID:         subject,
		CredentialType:     credType,
		Claims:             body.Claims,
		VerificationClaims: body.VerificationClaims,
		PreferredFormat:    parseCredentialFormat(body.PreferredFormat),
		ExpirationYears:    body.ExpirationYears,
	}

	resp, err := rt.CredentialService.IssueCredential(r.Context(), req)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "ISSUANCE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}

	WriteJSON(w, http.StatusAccepted, resp)
}

func (rt *Router) handleIdentityCredentialsList(w http.ResponseWriter, r *http.Request) {
	if rt.CredentialService == nil {
		WriteError(w, http.StatusServiceUnavailable, "CREDENTIALS_NOT_CONFIGURED", "credential service is not enabled on this deployment", r.Header.Get("X-Request-ID"))
		return
	}
	principal := userIDFromContext(r)
	if principal == "" {
		WriteError(w, http.StatusUnauthorized, "MISSING_PRINCIPAL", "authenticated subject required", r.Header.Get("X-Request-ID"))
		return
	}
	list, err := rt.CredentialService.ListCredentials(r.Context(), principal)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "LIST_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]interface{}{"credentials": list})
}
