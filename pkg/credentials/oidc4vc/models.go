package oidc4vc

import (
	"time"
)

// IssuerMetadata represents OIDC4VC Credential Issuer metadata
type IssuerMetadata struct {
	CredentialIssuer                  string                         `json:"credential_issuer"`
	AuthorizationServers              []string                       `json:"authorization_servers,omitempty"`
	TokenEndpoint                     string                         `json:"token_endpoint"`
	CredentialEndpoint                string                         `json:"credential_endpoint"`
	DeferredCredentialEndpoint        string                         `json:"deferred_credential_endpoint,omitempty"`
	NotificationEndpoint              string                         `json:"notification_endpoint,omitempty"`
	CredentialConfigurationsSupported map[string]CredentialConfig    `json:"credential_configurations_supported"`
	AuthDisplay                       []AuthDisplay                  `json:"auth_display,omitempty"`
}

// CredentialConfig represents credential configuration in issuer metadata
type CredentialConfig struct {
	Format                         string                `json:"format"`
	Scope                          string                `json:"scope"`
	CredentialType                 []string              `json:"credential_type"`
	CredentialDefinition           map[string]interface{} `json:"credential_definition"`
	ProofTypesSupported            []string              `json:"proof_types_supported"`
	ProofSigningAlgValuesSupported []string              `json:"proof_signing_alg_values_supported,omitempty"`
	Claims                         map[string]ClaimInfo  `json:"claims,omitempty"`
	Order                          []string              `json:"order,omitempty"`
}

// ClaimInfo provides information about a claim
type ClaimInfo struct {
	Mandatory bool   `json:"mandatory,omitempty"`
	ValueType string `json:"value_type,omitempty"`
}

// AuthDisplay represents authentication display info
type AuthDisplay struct {
	Name   string `json:"name"`
	Locale string `json:"locale,omitempty"`
}

// DisplayProperty represents display properties
type DisplayProperty struct {
	Name   string `json:"name"`
	Locale string `json:"locale,omitempty"`
}

// VerifierMetadata represents OIDC4VC Verifier metadata
type VerifierMetadata struct {
	VerifierID                     string   `json:"verifier_id"`
	VerificationEndpoint           string   `json:"verification_endpoint"`
	PresentationDefinitionEndpoint string   `json:"presentation_definition_endpoint,omitempty"`
	PresentationSubmissionEndpoint string   `json:"presentation_submission_endpoint"`
	CredentialTypesSupported       []string `json:"credential_types_supported"`
	ProofTypesSupported            []string `json:"proof_types_supported"`
	FormatSupported                []string `json:"format_supported"`
	ResponseTypesSupported         []string `json:"response_types_supported"`
	ScopesSupported                []string `json:"scopes_supported"`
	SubjectTypesSupported          []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
}

// PresentationRequest represents OIDC4VC presentation request (Verifier to Holder)
type PresentationRequest struct {
	ClientID            string              `json:"client_id"`
	RedirectURI         string              `json:"redirect_uri"`
	ResponseType        string              `json:"response_type"`
	ResponseMode        string              `json:"response_mode,omitempty"`
	Scope               string              `json:"scope,omitempty"`
	State               string              `json:"state"`
	Nonce               string              `json:"nonce,omitempty"`
	PresentationDef     *PresentationDef    `json:"presentation_definition,omitempty"`
	ClientMetadata      *ClientMetadata     `json:"client_metadata,omitempty"`
}

// PresentationDef represents presentation definition
type PresentationDef struct {
	ID                   string              `json:"id"`
	InputDescriptors     []InputDescriptor   `json:"input_descriptors"`
	SubmissionRequirements []SubmissionRequirement `json:"submission_requirements,omitempty"`
}

// InputDescriptor describes credential requirements
type InputDescriptor struct {
	ID          string       `json:"id"`
	Name        string       `json:"name,omitempty"`
	Purpose     string       `json:"purpose,omitempty"`
	Format      FormatSpec   `json:"format,omitempty"`
	Constraints *Constraints `json:"constraints,omitempty"`
}

// FormatSpec specifies supported formats
type FormatSpec struct {
	JSONLD *FormatDetails `json:"json-ld,omitempty"`
	JWT    *FormatDetails `json:"jwt,omitempty"`
	SDJWT  *FormatDetails `json:"sd-jwt,omitempty"`
	LDP_VC *FormatDetails `json:"ldp_vc,omitempty"`
	JWT_VC *FormatDetails `json:"jwt_vc_json,omitempty"`
}

// FormatDetails provides format-specific details
type FormatDetails struct {
	Alg       []string `json:"alg,omitempty"`
	ProofType []string `json:"proof_type,omitempty"`
	Crv       []string `json:"crv,omitempty"`
}

// Constraints represents constraints on credentials
type Constraints struct {
	Fields []FieldConstraint `json:"fields,omitempty"`
}

// FieldConstraint constrains credential fields
type FieldConstraint struct {
	Path            []string               `json:"path"`
	Filter          map[string]interface{} `json:"filter,omitempty"`
	Optional        bool                   `json:"optional,omitempty"`
	Predicate       string                 `json:"predicate,omitempty"`
	Purpose         string                 `json:"purpose,omitempty"`
	LimitDisclosure string                 `json:"limit_disclosure,omitempty"`
	Fields          []FieldConstraint      `json:"fields,omitempty"`
}

// SubmissionRequirement describes submission requirements
type SubmissionRequirement struct {
	Name       string                   `json:"name,omitempty"`
	Purpose    string                   `json:"purpose,omitempty"`
	Rule       string                   `json:"rule"`
	Count      int                      `json:"count,omitempty"`
	Min        int                      `json:"min,omitempty"`
	Max        int                      `json:"max,omitempty"`
	From       string                   `json:"from,omitempty"`
	FromNested []SubmissionRequirement  `json:"from_nested,omitempty"`
}

// PresentationSubmission represents response to presentation request
type PresentationSubmission struct {
	ID              string          `json:"id"`
	DefinitionID    string          `json:"definition_id"`
	DescriptorMap   []DescriptorMap `json:"descriptor_map"`
}

// DescriptorMap maps descriptor to presentation element
type DescriptorMap struct {
	ID        string         `json:"id"`
	Format    string         `json:"format"`
	Path      string         `json:"path"`
	PathNested *DescriptorMap `json:"path_nested,omitempty"`
}

// ClientMetadata represents client metadata in request
type ClientMetadata struct {
	ClientName             string                 `json:"client_name"`
	LogoURI                string                 `json:"logo_uri,omitempty"`
	Contacts               []string               `json:"contacts,omitempty"`
	SubjectType            string                 `json:"subject_type"`
	IDTokenSignedResponseAlg string               `json:"id_token_signed_response_alg,omitempty"`
	ResponseTypesSupported []string               `json:"response_types_supported,omitempty"`
	VPFormatsSupported     map[string]interface{} `json:"vp_formats_supported,omitempty"`
}

// TokenRequest represents OIDC4VC token request
type TokenRequest struct {
	GrantType         string `json:"grant_type"`
	Code              string `json:"code,omitempty"`
	ClientID          string `json:"client_id,omitempty"`
	ClientSecret      string `json:"client_secret,omitempty"`
	RedirectURI       string `json:"redirect_uri,omitempty"`
	CodeVerifier      string `json:"code_verifier,omitempty"`
	PreAuthorizedCode string `json:"pre-authorized_code,omitempty"`
	TxCode            string `json:"tx_code,omitempty"`
	Scope             string `json:"scope,omitempty"`
}

// TokenResponse represents OIDC4VC token response
type TokenResponse struct {
	AccessToken     string `json:"access_token"`
	TokenType       string `json:"token_type"`
	ExpiresIn       int64  `json:"expires_in"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	Scope           string `json:"scope,omitempty"`
	CNonce          string `json:"c_nonce,omitempty"`
	CNonceExpiresIn int64  `json:"c_nonce_expires_in,omitempty"`
}

// CredentialRequest represents an OIDC4VC credential request
type CredentialRequest struct {
	Format             string                 `json:"format"`
	CredentialType     []string               `json:"credential_type"`
	CredentialSubject  map[string]interface{} `json:"credentialSubject,omitempty"`
	Claims             map[string]interface{} `json:"claims,omitempty"`
	Proof              ProofRequest           `json:"proof"`
}

// ProofRequest represents proof in credential request
type ProofRequest struct {
	ProofType string `json:"proof_type"`
	JWT       string `json:"jwt,omitempty"`
	Alg       string `json:"alg,omitempty"`
	Kid       string `json:"kid,omitempty"`
	Nonce     string `json:"nonce,omitempty"`
	Aud       string `json:"aud,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
}

// CredentialResponse represents OIDC4VC credential response
type CredentialResponse struct {
	Format     string `json:"format"`
	Credential string `json:"credential"`
	CNonc      string `json:"c_nonce,omitempty"`
	ProofType  string `json:"proof_type,omitempty"`
	JWT        string `json:"jwt,omitempty"`
}

// DeferredCredentialRequest represents request for deferred credential
type DeferredCredentialRequest struct {
	AcceptanceToken string `json:"acceptance_token"`
}

// DeferredCredentialResponse represents deferred credential response
type DeferredCredentialResponse struct {
	TransactionID string `json:"transaction_id,omitempty"`
	Credential    string `json:"credential,omitempty"`
	IssuanceDate  string `json:"issuance_date,omitempty"`
}

// NotificationRequest represents notification request (issuer to holder)
type NotificationRequest struct {
	NotificationType string `json:"notification_type"`
	CredentialID     string `json:"credential_id,omitempty"`
	Event            string `json:"event,omitempty"`
	IssuedCredential string `json:"issued_credential,omitempty"`
}

// ProofOfPossession represents proof of possession for credential request
type ProofOfPossession struct {
	ProofType           string    `json:"proof_type"`
	JWT                 string    `json:"jwt"`
	Alg                 string    `json:"alg"`
	Kid                 string    `json:"kid"`
	Nonce               string    `json:"nonce"`
	Aud                 string    `json:"aud"`
	Iat                 int64     `json:"iat"`
	CNonce              string    `json:"c_nonce"`
	CNonceExpiresAt     time.Time `json:"c_nonce_expires_at"`
}

// AccessToken represents OIDC4VC access token
type AccessToken struct {
	Token              string    `json:"token"`
	ClientID           string    `json:"client_id"`
	Scope              string    `json:"scope"`
	ExpiresAt          time.Time `json:"expires_at"`
	CNonce             string    `json:"c_nonce"`
	CNonceExpiresAt    time.Time `json:"c_nonce_expires_at"`
}

// AuthorizationCode represents authorization code for OIDC4VC auth code flow
type AuthorizationCode struct {
	Code                string    `json:"code"`
	ClientID            string    `json:"client_id"`
	RedirectURI         string    `json:"redirect_uri"`
	Scope               string    `json:"scope"`
	State               string    `json:"state"`
	ExpiresAt           time.Time `json:"expires_at"`
	CodeChallenge       string    `json:"code_challenge"`
	CodeChallengeMethod string    `json:"code_challenge_method"`
}

// PreAuthorizedCode represents pre-authorized code
type PreAuthorizedCode struct {
	Code           string    `json:"code"`
	CredentialType string    `json:"credential_type"`
	ExpiresAt      time.Time `json:"expires_at"`
	PINRequired    bool      `json:"pin_required"`
	PINLength      int       `json:"pin_length"`
	MaxAttempts    int       `json:"max_attempts"`
}

// WalletInfo represents wallet capabilities and supported formats
type WalletInfo struct {
	ID                      string   `json:"id"`
	SupportedFormats        []string `json:"supported_formats"`
	SupportedProofTypes     []string `json:"supported_proof_types"`
	SupportedAlgorithms     []string `json:"supported_algorithms"`
	SupportsPreAuthorized   bool     `json:"supports_pre_authorized"`
	SupportsDirectPost      bool     `json:"supports_direct_post"`
	SupportsDirectPostJWT   bool     `json:"supports_direct_post_jwt"`
	PreferredCredentialFormat string `json:"preferred_credential_format"`
}

