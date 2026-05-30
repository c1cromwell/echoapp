package credentials

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/thechadcromwell/echoapp/pkg/passport/disclosure"
)

// FormatHandler handles credential format conversion
type FormatHandler struct {
	cryptoUtils *CryptoUtils
}

// NewFormatHandler creates new format handler
func NewFormatHandler(cryptoUtils *CryptoUtils) *FormatHandler {
	return &FormatHandler{
		cryptoUtils: cryptoUtils,
	}
}

// ToJSONLD converts credential to JSON-LD format
func (fh *FormatHandler) ToJSONLD(vc *VerifiableCredential) (string, error) {
	credentialJSON, err := json.MarshalIndent(vc, "", "  ")
	if err != nil {
		return "", NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to convert credential to JSON-LD",
			err.Error(),
		)
	}
	return string(credentialJSON), nil
}

// ToJWT converts credential to JWT format
func (fh *FormatHandler) ToJWT(vc *VerifiableCredential, issuerPrivateKey string) (string, error) {
	// Create JWT header
	header := map[string]interface{}{
		"alg": "EdDSA",
		"typ": "JWT",
		"kid": vc.Proof.VerificationMethod,
	}

	// Create JWT payload
	payload := map[string]interface{}{
		"vc":  vc,
		"iss": vc.Issuer,
		"sub": vc.CredentialSubject.ID,
		"iat": vc.IssuanceDate.Unix(),
		"jti": vc.ID,
	}

	if vc.ExpirationDate != nil {
		payload["exp"] = vc.ExpirationDate.Unix()
	}

	// Marshal header and payload
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	// Create JWT
	headerStr := string(headerJSON)
	payloadStr := string(payloadJSON)

	jwt, err := fh.cryptoUtils.CreateJWSSignature(headerStr, payloadStr, issuerPrivateKey)
	if err != nil {
		return "", NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to create JWT",
			err.Error(),
		)
	}

	return jwt, nil
}

// ToSDJWT converts credential to SD-JWT format with selective disclosure placeholders.
func (fh *FormatHandler) ToSDJWT(vc *VerifiableCredential, issuerPrivateKey string, disclosureFields []string) (string, error) {
	jwt, err := fh.ToJWT(vc, issuerPrivateKey)
	if err != nil {
		return "", err
	}
	if len(disclosureFields) == 0 {
		return jwt + "~", nil
	}
	claims := make(map[string]interface{})
	raw, _ := json.Marshal(vc.CredentialSubject)
	var sub map[string]interface{}
	if err := json.Unmarshal(raw, &sub); err != nil {
		return "", err
	}
	for _, field := range disclosureFields {
		if v, ok := sub[field]; ok {
			claims[field] = v
		}
	}
	if len(claims) == 0 {
		return jwt + "~", nil
	}
	sd, _, err := disclosure.BuildFromSubjectClaims(jwt, claims)
	if err != nil {
		return "", NewCredentialErrorWithDetails(ErrCodeInvalidCredential, "failed to build SD-JWT", err.Error())
	}
	return sd, nil
}

// FromJSONLD parses JSON-LD credential
func (fh *FormatHandler) FromJSONLD(credentialJSON string) (*VerifiableCredential, error) {
	var vc VerifiableCredential
	err := json.Unmarshal([]byte(credentialJSON), &vc)
	if err != nil {
		return nil, NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to parse JSON-LD credential",
			err.Error(),
		)
	}
	return &vc, nil
}

// FromJWT parses a compact JWT credential (header.payload.signature).
// Extracts the VC from the `vc` claim of the payload.
// Signature verification is deliberately not performed here — the caller
// must use VerifyCredential / the oidc4vc verifier for trust validation.
func (fh *FormatHandler) FromJWT(jwtToken string) (*VerifiableCredential, error) {
	parts := strings.Split(jwtToken, ".")
	if len(parts) != 3 {
		return nil, NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"invalid JWT format: expected header.payload.signature",
			fmt.Sprintf("got %d parts", len(parts)),
		)
	}

	// Base64url decode the payload (middle part).
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to base64url-decode JWT payload",
			err.Error(),
		)
	}

	// JWT-VC payload: the `vc` claim holds the VerifiableCredential object.
	var claims struct {
		VC      *VerifiableCredential `json:"vc"`
		Issuer  string                `json:"iss"`
		Subject string                `json:"sub"`
		JTI     string                `json:"jti"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to parse JWT payload JSON",
			err.Error(),
		)
	}
	if claims.VC == nil {
		return nil, NewCredentialError(ErrCodeInvalidCredential, "JWT payload missing 'vc' claim")
	}

	// Backfill top-level JWT claims into the VC if not already set.
	if claims.VC.Issuer == "" && claims.Issuer != "" {
		claims.VC.Issuer = claims.Issuer
	}
	if claims.VC.ID == "" && claims.JTI != "" {
		claims.VC.ID = claims.JTI
	}
	return claims.VC, nil
}

// FromSDJWT parses an SD-JWT credential.
// Phase 1: delegates to FromJWT — the selective-disclosure tilde-separated
// disclosures are stripped before JWT parsing (full SD-JWT is Phase 3 scope).
func (fh *FormatHandler) FromSDJWT(sdjwt string) (*VerifiableCredential, error) {
	// SD-JWT format: <JWT>~<disclosure1>~<disclosure2>~...
	// Strip disclosures and parse the base JWT.
	jwtPart := strings.SplitN(sdjwt, "~", 2)[0]
	if jwtPart == "" {
		return nil, NewCredentialError(ErrCodeInvalidCredential, "empty SD-JWT")
	}
	return fh.FromJWT(jwtPart)
}

// NegotiateFormat negotiates best format for wallet
func (fh *FormatHandler) NegotiateFormat(supportedFormats []string, preferredFormat CredentialFormat) CredentialFormat {
	// Check if preferred format is supported
	for _, f := range supportedFormats {
		if CredentialFormat(f) == preferredFormat {
			return preferredFormat
		}
	}

	// Default to JSON-LD
	return JSONLDFormat
}

// CredentialFormatter provides format conversion capabilities
type CredentialFormatter struct {
	formatHandler *FormatHandler
	config        *Config
}

// NewCredentialFormatter creates new credential formatter
func NewCredentialFormatter(formatHandler *FormatHandler, config *Config) *CredentialFormatter {
	return &CredentialFormatter{
		formatHandler: formatHandler,
		config:        config,
	}
}

// ConvertFormat converts credential between formats
func (cf *CredentialFormatter) ConvertFormat(vc *VerifiableCredential, fromFormat, toFormat CredentialFormat, privateKey string) (string, error) {
	// For now, all conversions go through JSON-LD as intermediate
	switch toFormat {
	case JSONLDFormat:
		return cf.formatHandler.ToJSONLD(vc)
	case JWTFormat:
		return cf.formatHandler.ToJWT(vc, privateKey)
	case SDJWTFormat:
		return cf.formatHandler.ToSDJWT(vc, privateKey, []string{})
	default:
		return "", NewCredentialError(
			ErrCodeUnsupportedFormat,
			fmt.Sprintf("unsupported format: %s", toFormat),
		)
	}
}

// CheckFormatSupport checks if format is supported
func (cf *CredentialFormatter) CheckFormatSupport(format CredentialFormat) bool {
	for _, f := range cf.config.CredentialConfig.SupportedFormats {
		if f == format {
			return true
		}
	}
	return false
}

// GetSupportedFormats returns list of supported formats
func (cf *CredentialFormatter) GetSupportedFormats() []string {
	var formats []string
	for _, f := range cf.config.CredentialConfig.SupportedFormats {
		formats = append(formats, string(f))
	}
	return formats
}

// SerializeCredential serializes credential for transmission
func (cf *CredentialFormatter) SerializeCredential(vc *VerifiableCredential, format CredentialFormat) (string, error) {
	return cf.ConvertFormat(vc, JSONLDFormat, format, "")
}

// DeserializeCredential deserializes credential from transmission format
func (cf *CredentialFormatter) DeserializeCredential(data string, format CredentialFormat) (*VerifiableCredential, error) {
	switch format {
	case JSONLDFormat:
		return cf.formatHandler.FromJSONLD(data)
	case JWTFormat:
		return cf.formatHandler.FromJWT(data)
	case SDJWTFormat:
		return cf.formatHandler.FromSDJWT(data)
	default:
		return nil, NewCredentialError(
			ErrCodeUnsupportedFormat,
			fmt.Sprintf("unsupported format: %s", format),
		)
	}
}
