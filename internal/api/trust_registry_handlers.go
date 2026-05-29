package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/onboarding"
)

func trustRegistryAdminAuthorized(r *http.Request) bool {
	key := os.Getenv("TRUST_REGISTRY_ADMIN_KEY")
	if key == "" {
		return false
	}
	return r.Header.Get("X-Trust-Registry-Admin-Key") == key
}

func (rt *Router) handleTrustRegistry(w http.ResponseWriter, r *http.Request) {
	if rt.TrustRegistry == nil {
		WriteError(w, http.StatusServiceUnavailable, "TRUST_REGISTRY_UNAVAILABLE", "Trust registry not configured", r.Header.Get("X-Request-ID"))
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v1/trust-registry")
	path = strings.TrimPrefix(path, "/")
	if path == "" || path == "issuers" {
		rt.handleTrustRegistryList(w, r)
		return
	}
	if !strings.HasPrefix(path, "issuers/") {
		WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown trust registry path", r.Header.Get("X-Request-ID"))
		return
	}
	rest := strings.TrimPrefix(path, "issuers/")
	if strings.Contains(rest, "/") {
		parts := strings.SplitN(rest, "/", 2)
		issuerID, action := parts[0], parts[1]
		switch action {
		case "suspend":
			rt.handleTrustRegistrySuspend(w, r, issuerID)
		case "resume":
			rt.handleTrustRegistryResume(w, r, issuerID)
		case "revoke":
			rt.handleTrustRegistryRevoke(w, r, issuerID)
		default:
			WriteError(w, http.StatusNotFound, "NOT_FOUND", "Unknown trust registry action", r.Header.Get("X-Request-ID"))
		}
		return
	}
	rt.handleTrustRegistryGet(w, r, rest)
}

func (rt *Router) handleTrustRegistryList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		issuers := rt.TrustRegistry.ListIssuers()
		out := make([]map[string]interface{}, 0, len(issuers))
		for _, issuer := range issuers {
			out = append(out, issuerToJSON(issuer))
		}
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"issuers": out,
			"count":   len(out),
		})
	case http.MethodPost:
		if !trustRegistryAdminAuthorized(r) {
			WriteError(w, http.StatusForbidden, "FORBIDDEN", "Admin key required", r.Header.Get("X-Request-ID"))
			return
		}
		var req trustRegistryRegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid JSON body", r.Header.Get("X-Request-ID"))
			return
		}
		issuer := req.toTrustedIssuer()
		if err := rt.TrustRegistry.RegisterIssuer(issuer); err != nil {
			WriteError(w, http.StatusConflict, "REGISTER_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
			return
		}
		WriteJSON(w, http.StatusCreated, issuerToJSON(issuer))
	default:
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", r.Header.Get("X-Request-ID"))
	}
}

func (rt *Router) handleTrustRegistryGet(w http.ResponseWriter, r *http.Request, ref string) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only GET is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	issuer, err := rt.TrustRegistry.ResolveIssuer(ref)
	if err != nil || issuer == nil {
		WriteError(w, http.StatusNotFound, "ISSUER_NOT_FOUND", "Issuer not found", r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, issuerToJSON(issuer))
}

func (rt *Router) handleTrustRegistrySuspend(w http.ResponseWriter, r *http.Request, issuerID string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if !trustRegistryAdminAuthorized(r) {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "Admin key required", r.Header.Get("X-Request-ID"))
		return
	}
	if err := rt.TrustRegistry.SuspendIssuer(issuerID); err != nil {
		WriteError(w, http.StatusBadRequest, "SUSPEND_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "suspended", "issuer_id": issuerID})
}

func (rt *Router) handleTrustRegistryResume(w http.ResponseWriter, r *http.Request, issuerID string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if !trustRegistryAdminAuthorized(r) {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "Admin key required", r.Header.Get("X-Request-ID"))
		return
	}
	if err := rt.TrustRegistry.ResumeIssuer(issuerID); err != nil {
		WriteError(w, http.StatusBadRequest, "RESUME_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "active", "issuer_id": issuerID})
}

func (rt *Router) handleTrustRegistryRevoke(w http.ResponseWriter, r *http.Request, issuerID string) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Only POST is allowed", r.Header.Get("X-Request-ID"))
		return
	}
	if !trustRegistryAdminAuthorized(r) {
		WriteError(w, http.StatusForbidden, "FORBIDDEN", "Admin key required", r.Header.Get("X-Request-ID"))
		return
	}
	if err := rt.TrustRegistry.RevokeIssuer(issuerID); err != nil {
		WriteError(w, http.StatusBadRequest, "REVOKE_FAILED", err.Error(), r.Header.Get("X-Request-ID"))
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "revoked", "issuer_id": issuerID})
}

type trustRegistryRegisterRequest struct {
	ID                          string   `json:"issuer_id"`
	Name                        string   `json:"name"`
	DID                         string   `json:"issuer_did"`
	Type                        string   `json:"issuer_type"`
	Jurisdiction                string   `json:"jurisdiction"`
	TrustLevel                  string   `json:"trust_level"`
	Status                      string   `json:"status"`
	CredentialTypes             []string `json:"trusted_credential_types"`
	VerificationPublicKeyBase64 string   `json:"verification_public_key_b64"`
	PublicKeyURL                string   `json:"public_key_url"`
}

func (req trustRegistryRegisterRequest) toTrustedIssuer() *onboarding.TrustedIssuer {
	status := req.Status
	if status == "" {
		status = "active"
	}
	credTypes := make([]onboarding.CredentialType, len(req.CredentialTypes))
	for i, ct := range req.CredentialTypes {
		credTypes[i] = onboarding.CredentialType(ct)
	}
	jurisdiction := onboarding.Jurisdiction(req.Jurisdiction)
	if jurisdiction == "" {
		jurisdiction = onboarding.JurisdictionGlobal
	}
	return &onboarding.TrustedIssuer{
		ID:                          req.ID,
		Name:                        req.Name,
		DID:                         req.DID,
		Type:                        onboarding.IssuerType(req.Type),
		Jurisdiction:                jurisdiction,
		TrustLevel:                  onboarding.TrustLevel(req.TrustLevel),
		Status:                      status,
		CredentialTypes:             credTypes,
		VerificationPublicKeyBase64: req.VerificationPublicKeyBase64,
		PublicKeyURL:                req.PublicKeyURL,
		VerifiedTimestamp:           time.Now(),
		LastVerificationDate:        time.Now(),
	}
}

func issuerToJSON(issuer *onboarding.TrustedIssuer) map[string]interface{} {
	credTypes := make([]string, len(issuer.CredentialTypes))
	for i, ct := range issuer.CredentialTypes {
		credTypes[i] = string(ct)
	}
	return map[string]interface{}{
		"issuer_id":                 issuer.ID,
		"name":                      issuer.Name,
		"issuer_did":                issuer.DID,
		"issuer_type":               string(issuer.Type),
		"jurisdiction":              string(issuer.Jurisdiction),
		"trust_level":               string(issuer.TrustLevel),
		"status":                    issuer.Status,
		"trusted_credential_types":  credTypes,
		"verification_public_key_b64": issuer.VerificationPublicKeyBase64,
		"public_key_url":            issuer.PublicKeyURL,
		"last_verified":             issuer.LastVerificationDate.UTC().Format(time.RFC3339),
	}
}
