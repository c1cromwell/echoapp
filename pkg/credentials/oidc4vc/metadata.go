package oidc4vc

import (
	"fmt"
	"time"
)

// MetadataManager manages OIDC4VC metadata generation
type MetadataManager struct {
	issuerDID           string
	verifierDID         string
	issuerBaseURL       string
	verifierBaseURL     string
	supportedFormats    []string
	supportedProofTypes []string
}

// NewMetadataManager creates new metadata manager
func NewMetadataManager(issuerDID, verifierDID, issuerBaseURL, verifierBaseURL string) *MetadataManager {
	return &MetadataManager{
		issuerDID:           issuerDID,
		verifierDID:         verifierDID,
		issuerBaseURL:       issuerBaseURL,
		verifierBaseURL:     verifierBaseURL,
		supportedFormats:    []string{"json-ld+jwt", "jwt_vc_json", "sd-jwt"},
		supportedProofTypes: []string{"jwt", "ldp_vc", "ldp_vp"},
	}
}

// GenerateIssuerMetadata generates OIDC4VC issuer metadata
func (m *MetadataManager) GenerateIssuerMetadata() *IssuerMetadata {
	return &IssuerMetadata{
		CredentialIssuer:                  m.issuerDID,
		AuthorizationServers:              []string{m.issuerBaseURL},
		TokenEndpoint:                     fmt.Sprintf("%s/token", m.issuerBaseURL),
		CredentialEndpoint:                fmt.Sprintf("%s/credential", m.issuerBaseURL),
		DeferredCredentialEndpoint:        fmt.Sprintf("%s/deferred_credential", m.issuerBaseURL),
		NotificationEndpoint:              fmt.Sprintf("%s/notification", m.issuerBaseURL),
		CredentialConfigurationsSupported: m.buildCredentialConfigs(),
		AuthDisplay: []AuthDisplay{
			{
				Name:   "Credential Issuer",
				Locale: "en-US",
			},
		},
	}
}

// buildCredentialConfigs builds credential configurations
func (m *MetadataManager) buildCredentialConfigs() map[string]CredentialConfig {
	configs := make(map[string]CredentialConfig)

	credentialTypes := []string{
		"ProofOfHumanity",
		"KYCLite",
		"HighAssurance",
		"Professional",
	}

	for _, credType := range credentialTypes {
		configs[credType] = CredentialConfig{
			Format:         "json-ld+jwt",
			Scope:          fmt.Sprintf("credential_%s", credType),
			CredentialType: []string{credType},
			CredentialDefinition: map[string]interface{}{
				"@context": []string{
					"https://www.w3.org/2018/credentials/v1",
					"https://www.w3.org/2018/credentials/examples/v1",
				},
				"type": []string{"VerifiableCredential", credType},
			},
			ProofTypesSupported:            []string{"jwt", "ldp_vc"},
			ProofSigningAlgValuesSupported: []string{"EdDSA", "ES256"},
			Claims:                         m.buildClaimsInfo(credType),
		}
	}

	return configs
}

// buildClaimsInfo builds claims information for credential type
func (m *MetadataManager) buildClaimsInfo(credType string) map[string]ClaimInfo {
	claims := make(map[string]ClaimInfo)

	baseClaimsCount := 3
	switch credType {
	case "ProofOfHumanity":
		claims["humanityProof"] = ClaimInfo{
			Mandatory: true,
			ValueType: "boolean",
		}
	case "KYCLite":
		claims["firstName"] = ClaimInfo{Mandatory: true, ValueType: "string"}
		claims["lastName"] = ClaimInfo{Mandatory: true, ValueType: "string"}
		claims["verificationLevel"] = ClaimInfo{Mandatory: true, ValueType: "string"}
	case "HighAssurance":
		claims["firstName"] = ClaimInfo{Mandatory: true, ValueType: "string"}
		claims["lastName"] = ClaimInfo{Mandatory: true, ValueType: "string"}
		claims["dateOfBirth"] = ClaimInfo{Mandatory: true, ValueType: "string"}
		claims["verificationLevel"] = ClaimInfo{Mandatory: true, ValueType: "string"}
	case "Professional":
		claims["profession"] = ClaimInfo{Mandatory: true, ValueType: "string"}
		claims["employer"] = ClaimInfo{Mandatory: false, ValueType: "string"}
		claims["credentials"] = ClaimInfo{Mandatory: true, ValueType: "array"}
	}

	// Add common claims
	for i := baseClaimsCount; i < len(claims)+baseClaimsCount; i++ {
		claims["issuanceDate"] = ClaimInfo{
			Mandatory: true,
			ValueType: "string",
		}
		claims["expirationDate"] = ClaimInfo{
			Mandatory: false,
			ValueType: "string",
		}
		claims["issuer"] = ClaimInfo{
			Mandatory: true,
			ValueType: "string",
		}
	}

	return claims
}

// GenerateVerifierMetadata generates OIDC4VC verifier metadata
func (m *MetadataManager) GenerateVerifierMetadata() *VerifierMetadata {
	return &VerifierMetadata{
		VerifierID:                     m.verifierDID,
		VerificationEndpoint:           fmt.Sprintf("%s/verification", m.verifierBaseURL),
		PresentationDefinitionEndpoint: fmt.Sprintf("%s/presentation_definition", m.verifierBaseURL),
		PresentationSubmissionEndpoint: fmt.Sprintf("%s/presentation_submission", m.verifierBaseURL),
		CredentialTypesSupported: []string{
			"ProofOfHumanity",
			"KYCLite",
			"HighAssurance",
			"Professional",
		},
		ProofTypesSupported: m.supportedProofTypes,
		FormatSupported:     m.supportedFormats,
		ResponseTypesSupported: []string{
			"vp_token",
			"id_token",
		},
		ScopesSupported: []string{
			"credential_ProofOfHumanity",
			"credential_KYCLite",
			"credential_HighAssurance",
			"credential_Professional",
		},
		SubjectTypesSupported: []string{
			"public",
			"pairwise",
		},
		IdTokenSigningAlgValuesSupported: []string{
			"EdDSA",
			"ES256",
			"RS256",
		},
	}
}

// GenerateAuthorizationRequest generates authorization request
func (m *MetadataManager) GenerateAuthorizationRequest(clientID, redirectURI, state string) map[string]string {
	return map[string]string{
		"client_id":     clientID,
		"redirect_uri":  redirectURI,
		"response_type": "code",
		"scope":         "openid",
		"state":         state,
	}
}

// GeneratePresentationRequest generates presentation request with definition
func (m *MetadataManager) GeneratePresentationRequest(clientID, redirectURI, state, credentialType string) *PresentationRequest {
	return &PresentationRequest{
		ClientID:        clientID,
		RedirectURI:     redirectURI,
		ResponseType:    "vp_token",
		ResponseMode:    "direct_post",
		State:           state,
		Nonce:           state, // In production, use separate nonce
		PresentationDef: m.buildPresentationDefinition(credentialType),
		ClientMetadata:  m.buildClientMetadata(),
	}
}

// buildPresentationDefinition builds presentation definition for credential type
func (m *MetadataManager) buildPresentationDefinition(credentialType string) *PresentationDef {
	return &PresentationDef{
		ID: fmt.Sprintf("pd_%s_%d", credentialType, time.Now().Unix()),
		InputDescriptors: []InputDescriptor{
			{
				ID:      fmt.Sprintf("id_%s", credentialType),
				Name:    fmt.Sprintf("%s Credential", credentialType),
				Purpose: fmt.Sprintf("Please provide your %s credential", credentialType),
				Format: FormatSpec{
					JSONLD: &FormatDetails{
						ProofType: []string{"Ed25519Signature2018", "JsonWebSignature2020"},
					},
					JWT: &FormatDetails{
						Alg: []string{"EdDSA", "ES256"},
					},
					SDJWT: &FormatDetails{
						Alg: []string{"EdDSA", "ES256"},
					},
				},
				Constraints: &Constraints{
					Fields: []FieldConstraint{
						{
							Path: []string{"$.type"},
							Filter: map[string]interface{}{
								"type":     "array",
								"contains": credentialType,
							},
						},
						{
							Path: []string{"$.credentialSubject.id"},
						},
					},
				},
			},
		},
	}
}

// buildClientMetadata builds client metadata
func (m *MetadataManager) buildClientMetadata() *ClientMetadata {
	return &ClientMetadata{
		ClientName:  "Credential Verifier",
		SubjectType: "public",
		ResponseTypesSupported: []string{
			"vp_token",
			"id_token",
		},
		VPFormatsSupported: map[string]interface{}{
			"json-ld": map[string]interface{}{
				"proof_type": []string{"Ed25519Signature2018", "JsonWebSignature2020"},
			},
			"jwt": map[string]interface{}{
				"alg": []string{"EdDSA", "ES256"},
			},
			"sd-jwt": map[string]interface{}{
				"alg": []string{"EdDSA", "ES256"},
			},
		},
	}
}

// ValidateCredentialRequest validates credential request against issuer capabilities
func (m *MetadataManager) ValidateCredentialRequest(req *CredentialRequest, configs map[string]CredentialConfig) error {
	if req.Format == "" {
		return fmt.Errorf("credential format is required")
	}

	// Check if format is supported
	formatSupported := false
	for _, fmt := range m.supportedFormats {
		if fmt == req.Format {
			formatSupported = true
			break
		}
	}
	if !formatSupported {
		return fmt.Errorf("unsupported credential format: %s", req.Format)
	}

	// Check credential types
	if len(req.CredentialType) == 0 {
		return fmt.Errorf("credential type is required")
	}

	for _, credType := range req.CredentialType {
		if _, exists := configs[credType]; !exists {
			return fmt.Errorf("unsupported credential type: %s", credType)
		}
	}

	// Validate proof
	if req.Proof.ProofType == "" {
		return fmt.Errorf("proof type is required")
	}

	return nil
}

// ValidatePresentationRequest validates presentation request
func (m *MetadataManager) ValidatePresentationRequest(req *PresentationRequest) error {
	if req.ClientID == "" {
		return fmt.Errorf("client_id is required")
	}

	if req.RedirectURI == "" {
		return fmt.Errorf("redirect_uri is required")
	}

	if req.State == "" {
		return fmt.Errorf("state is required")
	}

	if req.PresentationDef == nil || len(req.PresentationDef.InputDescriptors) == 0 {
		return fmt.Errorf("presentation definition with input descriptors is required")
	}

	return nil
}
