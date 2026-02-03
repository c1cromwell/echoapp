package credentials

import (
	"encoding/json"
	"fmt"
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

// ToSDJWT converts credential to SD-JWT format
func (fh *FormatHandler) ToSDJWT(vc *VerifiableCredential, issuerPrivateKey string, disclosureFields []string) (string, error) {
	// SD-JWT allows selective disclosure of claims
	// For now, return JWT (proper SD-JWT implementation would use sd-jwt library)
	return fh.ToJWT(vc, issuerPrivateKey)
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

// FromJWT parses JWT credential
func (fh *FormatHandler) FromJWT(jwt string) (*VerifiableCredential, error) {
	// In production, use proper JWT library with verification
	// For now, basic parsing
	return nil, fmt.Errorf("JWT parsing not implemented")
}

// FromSDJWT parses SD-JWT credential
func (fh *FormatHandler) FromSDJWT(sdjwt string) (*VerifiableCredential, error) {
	// In production, use proper SD-JWT library
	return nil, fmt.Errorf("SD-JWT parsing not implemented")
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
