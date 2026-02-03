package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Verifier manages credential verification
type Verifier struct {
	config          *Config
	cryptoUtils     *CryptoUtils
	storage         Storage
	revocationMgr   *RevocationManager
	semaphore       chan struct{}
	trustScoreCache map[string]float64
	cacheMutex      sync.RWMutex
}

// NewVerifier creates new credential verifier
func NewVerifier(config *Config, cryptoUtils *CryptoUtils, storage Storage, revocationMgr *RevocationManager) *Verifier {
	return &Verifier{
		config:          config,
		cryptoUtils:     cryptoUtils,
		storage:         storage,
		revocationMgr:   revocationMgr,
		semaphore:       make(chan struct{}, config.VerifierConfig.MaxConcurrentVerifications),
		trustScoreCache: make(map[string]float64),
	}
}

// VerifyCredential verifies a credential
func (v *Verifier) VerifyCredential(ctx context.Context, req *CredentialVerificationRequest) (*CredentialVerificationResult, error) {
	// Acquire semaphore slot
	select {
	case v.semaphore <- struct{}{}:
		defer func() { <-v.semaphore }()
	case <-ctx.Done():
		return nil, NewCredentialError(ErrCodeTimeoutError, "context cancelled during verification")
	}

	// Set context timeout
	ctx, cancel := context.WithTimeout(ctx, v.config.CredentialConfig.VerificationTimeout)
	defer cancel()

	result := &CredentialVerificationResult{
		VerifiedAt: time.Now(),
		Errors:     []VerificationError{},
	}

	// Parse credential based on format
	var vc *VerifiableCredential
	var err error

	switch req.Format {
	case JSONLDFormat:
		vc, err = v.parseJSONLDCredential(req.Credential)
	case JWTFormat:
		vc, err = v.parseJWTCredential(req.Credential)
	case SDJWTFormat:
		vc, err = v.parseSDJWTCredential(req.Credential)
	default:
		return nil, NewCredentialError(ErrCodeUnsupportedFormat, fmt.Sprintf("unsupported format: %s", req.Format))
	}

	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, VerificationError{
			Code:    "PARSE_ERROR",
			Message: err.Error(),
		})
		return result, nil
	}

	// Validate structure
	if v.config.VerifierConfig.StrictSignature {
		if err := v.validateCredentialStructure(vc); err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, VerificationError{
				Code:    "STRUCTURE_INVALID",
				Message: err.Error(),
			})
			return result, nil
		}
	}

	// Set result metadata
	result.CredentialID = vc.ID
	result.Issuer = vc.Issuer
	result.Subject = vc.CredentialSubject.ID
	result.CredentialType = CredentialType(vc.Type[1]) // First type is VerifiableCredential

	// Verify signature
	result.SignatureValid, err = v.verifySignature(vc, req.IssuerDID)
	if err != nil {
		result.Errors = append(result.Errors, VerificationError{
			Code:    "SIGNATURE_ERROR",
			Message: err.Error(),
		})
	}

	// Check expiration
	if v.config.VerifierConfig.CheckExpiration {
		result.NotExpired = v.checkExpiration(vc)
		result.ExpirationDate = vc.ExpirationDate
		if !result.NotExpired {
			result.Errors = append(result.Errors, VerificationError{
				Code:    "EXPIRED",
				Message: "Credential has expired",
			})
		}
	} else {
		result.NotExpired = true
	}

	// Check revocation status
	if v.config.VerifierConfig.EnableRevocation {
		revokeStatus, err := v.revocationMgr.CheckRevocationStatus(ctx, vc.ID)
		if err != nil {
			result.Errors = append(result.Errors, VerificationError{
				Code:    "REVOCATION_CHECK_ERROR",
				Message: err.Error(),
			})
			result.RevocationStatus = "unknown"
		} else {
			result.NotRevoked = !revokeStatus.IsRevoked
			result.RevocationStatus = "active"
			if revokeStatus.IsRevoked {
				result.RevocationStatus = "revoked"
				result.Errors = append(result.Errors, VerificationError{
					Code:    "REVOKED",
					Message: fmt.Sprintf("Credential revoked: %s", revokeStatus.RevocationReason),
				})
			}
		}
	} else {
		result.NotRevoked = true
		result.RevocationStatus = "active"
	}

	// Overall validity
	result.IsValid = result.SignatureValid && result.NotExpired && result.NotRevoked && len(result.Errors) == 0

	// Calculate trust score
	trustInput := TrustScoreInput{
		CredentialType:   result.CredentialType,
		Age:              time.Since(vc.IssuanceDate),
		IssuerReputation: 85.0, // In production, lookup actual reputation
		IsRevoked:        !result.NotRevoked,
		SignatureValid:   result.SignatureValid,
		ExpirationValid:  result.NotExpired,
	}

	// Determine verification level from credential claims
	verificationLevel := "basic"
	if len(vc.CredentialSubject.VerificationClaims) > 0 {
		verificationLevel = vc.CredentialSubject.VerificationClaims[0].VerificationLevel
	}
	trustInput.VerificationLevel = verificationLevel

	// Cache and use trust score (in production, store this)
	trustScore := v.calculateTrustScore(trustInput)
	v.cacheTrustScore(vc.ID, trustScore)

	return result, nil
}

// parseJSONLDCredential parses JSON-LD credential
func (v *Verifier) parseJSONLDCredential(credentialJSON string) (*VerifiableCredential, error) {
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

// parseJWTCredential parses JWT credential
func (v *Verifier) parseJWTCredential(jwt string) (*VerifiableCredential, error) {
	// In production, use proper JWT parsing library
	// For now, return error
	return nil, NewCredentialError(
		ErrCodeUnsupportedFormat,
		"JWT parsing not fully implemented",
	)
}

// parseSDJWTCredential parses SD-JWT credential
func (v *Verifier) parseSDJWTCredential(sdjwt string) (*VerifiableCredential, error) {
	// In production, use proper SD-JWT parsing library
	return nil, NewCredentialError(
		ErrCodeUnsupportedFormat,
		"SD-JWT parsing not fully implemented",
	)
}

// validateCredentialStructure validates W3C credential structure
func (v *Verifier) validateCredentialStructure(vc *VerifiableCredential) error {
	errors := ValidationErrors{}

	// Validate context
	if len(vc.Context) == 0 {
		errors.Add("context", "credential context is required", "MISSING_CONTEXT")
	}

	// Validate type
	if len(vc.Type) < 2 {
		errors.Add("type", "credential must have at least VerifiableCredential and specific type", "INVALID_TYPE")
	}

	// Check for required VerifiableCredential type
	hasVCType := false
	for _, t := range vc.Type {
		if t == "VerifiableCredential" {
			hasVCType = true
			break
		}
	}
	if !hasVCType {
		errors.Add("type", "VerifiableCredential type is required", "MISSING_VC_TYPE")
	}

	// Validate issuer
	if vc.Issuer == "" {
		errors.Add("issuer", "issuer is required", "MISSING_ISSUER")
	}

	// Validate issuance date
	if vc.IssuanceDate.IsZero() {
		errors.Add("issuance_date", "issuance date is required", "MISSING_ISSUANCE_DATE")
	}

	// Validate credential subject
	if vc.CredentialSubject.ID == "" {
		errors.Add("credential_subject", "credential subject ID is required", "MISSING_SUBJECT_ID")
	}

	// Validate proof
	if vc.Proof.Type == "" {
		errors.Add("proof", "proof type is required", "MISSING_PROOF_TYPE")
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// verifySignature verifies credential signature
func (v *Verifier) verifySignature(vc *VerifiableCredential, issuerDID string) (bool, error) {
	if vc.Proof.ProofValue == "" {
		return false, NewCredentialError(ErrCodeInvalidProof, "proof value is missing")
	}

	// Convert credential to JSON for signature verification
	credentialBytes, err := json.Marshal(vc)
	if err != nil {
		return false, NewCredentialErrorWithDetails(
			ErrCodeInvalidProof,
			"failed to marshal credential for verification",
			err.Error(),
		)
	}

	// In production, retrieve issuer's public key from DID document
	// For now, use issuer DID to verify
	// Get public key from issuer (simplified)
	publicKeyBase64 := issuerDID // In production, resolve from DID document

	// Verify signature
	valid, err := v.cryptoUtils.VerifySignature(
		publicKeyBase64,
		credentialBytes,
		vc.Proof.ProofValue,
	)

	if err != nil {
		return false, NewCredentialErrorWithDetails(
			ErrCodeInvalidProof,
			"signature verification failed",
			err.Error(),
		)
	}

	return valid, nil
}

// checkExpiration checks if credential is expired
func (v *Verifier) checkExpiration(vc *VerifiableCredential) bool {
	if vc.ExpirationDate == nil {
		// No expiration date means credential doesn't expire
		return true
	}

	return time.Now().Before(*vc.ExpirationDate)
}

// calculateTrustScore calculates trust score for credential
func (v *Verifier) calculateTrustScore(input TrustScoreInput) float64 {
	score := 100.0

	// Penalize for revocation
	if input.IsRevoked {
		return 0.0
	}

	// Penalize for invalid signature
	if !input.SignatureValid {
		score -= 50.0
	}

	// Penalize for expired credential
	if !input.ExpirationValid {
		score -= 40.0
	}

	// Adjust based on credential type
	switch input.CredentialType {
	case ProofOfHumanity:
		score *= 0.8
	case KYCLite:
		score *= 0.85
	case HighAssurance:
		score *= 1.0
	case Professional:
		score *= 0.9
	}

	// Adjust based on age (older credentials are less trustworthy)
	ageMonths := input.Age.Hours() / (24 * 30)
	if ageMonths > 6 {
		score *= 0.9
	}
	if ageMonths > 12 {
		score *= 0.8
	}

	// Adjust based on issuer reputation
	score = score * (input.IssuerReputation / 100.0)

	// Adjust based on verification level
	switch input.VerificationLevel {
	case "basic":
		score *= 0.7
	case "intermediate":
		score *= 0.85
	case "high":
		score *= 1.0
	}

	// Clamp to 0-100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// cacheTrustScore caches trust score for credential
func (v *Verifier) cacheTrustScore(credentialID string, score float64) {
	v.cacheMutex.Lock()
	defer v.cacheMutex.Unlock()
	v.trustScoreCache[credentialID] = score
}

// GetTrustScore retrieves cached trust score
func (v *Verifier) GetTrustScore(credentialID string) (float64, bool) {
	v.cacheMutex.RLock()
	defer v.cacheMutex.RUnlock()
	score, exists := v.trustScoreCache[credentialID]
	return score, exists
}

// BatchVerify verifies multiple credentials concurrently
func (v *Verifier) BatchVerify(ctx context.Context, requests []CredentialVerificationRequest) ([]*CredentialVerificationResult, error) {
	results := make([]*CredentialVerificationResult, len(requests))
	errors := make([]error, len(requests))
	var wg sync.WaitGroup

	for i, req := range requests {
		wg.Add(1)
		go func(index int, request CredentialVerificationRequest) {
			defer wg.Done()
			result, err := v.VerifyCredential(ctx, &request)
			results[index] = result
			errors[index] = err
		}(i, req)
	}

	wg.Wait()

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return results, err
		}
	}

	return results, nil
}
