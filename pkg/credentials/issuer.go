package credentials

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Issuer manages credential issuance
type Issuer struct {
	config          *Config
	cryptoUtils     *CryptoUtils
	storage         Storage
	progressTracker map[string]*IssuanceProgress
	progressMutex   sync.RWMutex
	semaphore       chan struct{} // Limit concurrent issuance
}

// NewIssuer creates new credential issuer
func NewIssuer(config *Config, cryptoUtils *CryptoUtils, storage Storage) *Issuer {
	return &Issuer{
		config:          config,
		cryptoUtils:     cryptoUtils,
		storage:         storage,
		progressTracker: make(map[string]*IssuanceProgress),
		semaphore:       make(chan struct{}, config.IssuerConfig.MaxConcurrentIssues),
	}
}

// IssueCredential issues a verifiable credential
func (i *Issuer) IssueCredential(ctx context.Context, req *CredentialIssuanceRequest) (*CredentialIssuanceResponse, error) {
	// Acquire semaphore slot
	select {
	case i.semaphore <- struct{}{}:
		defer func() { <-i.semaphore }()
	case <-ctx.Done():
		return nil, NewCredentialError(ErrCodeTimeoutError, "context cancelled during issuance")
	}

	// Create credential ID
	credentialID := uuid.New().String()

	// Track progress
	progress := &IssuanceProgress{
		CredentialID: credentialID,
		Status:       "initiated",
		Progress:     0,
		StartedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		EstimatedEnd: time.Now().Add(i.config.CredentialConfig.IssuanceTimeout),
		CurrentStep:  "Validating request",
	}

	i.setProgress(credentialID, progress)
	defer i.cleanupProgress(credentialID)

	// Set context timeout
	ctx, cancel := context.WithTimeout(ctx, i.config.CredentialConfig.IssuanceTimeout)
	defer cancel()

	// Validate request
	if err := i.validateIssuanceRequest(req); err != nil {
		progress.Status = "failed"
		progress.ErrorMessage = err.Error()
		i.setProgress(credentialID, progress)
		return nil, err
	}
	progress.Progress = 10
	progress.CurrentStep = "Generating credential"
	i.setProgress(credentialID, progress)

	// Create credential
	vc := i.createWC3VerifiableCredential(req, credentialID)
	progress.Progress = 30
	progress.CurrentStep = "Signing credential"
	progress.UpdatedAt = time.Now()
	i.setProgress(credentialID, progress)

	// Convert to selected format
	encodedCredential, err := i.encodeCredential(vc, req.PreferredFormat, credentialID)
	if err != nil {
		progress.Status = "failed"
		progress.ErrorMessage = err.Error()
		i.setProgress(credentialID, progress)
		return nil, err
	}
	progress.Progress = 60
	progress.CurrentStep = "Storing credential"
	progress.UpdatedAt = time.Now()
	i.setProgress(credentialID, progress)

	// Store credential
	err = i.storage.StoreCredential(ctx, credentialID, vc, req.PreferredFormat, encodedCredential)
	if err != nil {
		progress.Status = "failed"
		progress.ErrorMessage = err.Error()
		i.setProgress(credentialID, progress)
		return nil, NewCredentialErrorWithDetails(
			ErrCodeStorageFailed,
			"failed to store credential",
			err.Error(),
		)
	}

	// Anchor to blockchain if enabled
	if i.config.CredentialConfig.EnableBlockchainStorage {
		progress.Progress = 70
		progress.CurrentStep = "Anchoring to blockchain"
		progress.UpdatedAt = time.Now()
		i.setProgress(credentialID, progress)

		_, err := i.storage.AnchorCredential(ctx, credentialID, vc)
		if err != nil {
			// Log error but continue - credential is still valid
			fmt.Printf("Warning: failed to anchor credential: %v\n", err)
		} else {
			progress.Status = "anchored"
		}
	}

	// Calculate expiration
	expirationYears := getExpirationYears(req.CredentialType)
	if req.ExpirationYears > 0 {
		expirationYears = req.ExpirationYears
	}
	expiresAt := time.Now().AddDate(expirationYears, 0, 0)

	progress.Progress = 100
	progress.Status = "issued"
	progress.CurrentStep = "Credential issued"
	progress.UpdatedAt = time.Now()
	i.setProgress(credentialID, progress)

	// Create response
	response := &CredentialIssuanceResponse{
		CredentialID:         credentialID,
		VerifiableCredential: encodedCredential,
		Format:               req.PreferredFormat,
		IssuedAt:             time.Now(),
		ExpiresAt:            expiresAt,
		Status:               "issued",
	}

	return response, nil
}

// validateIssuanceRequest validates credential issuance request
func (i *Issuer) validateIssuanceRequest(req *CredentialIssuanceRequest) error {
	errors := ValidationErrors{}

	if req.SubjectDID == "" {
		errors.Add("subject_did", "subject DID is required", "MISSING_SUBJECT_DID")
	}

	if req.CredentialType == "" {
		errors.Add("credential_type", "credential type is required", "MISSING_CREDENTIAL_TYPE")
	}

	if len(req.Claims) == 0 && len(req.VerificationClaims) == 0 {
		errors.Add("claims", "at least one claim is required", "MISSING_CLAIMS")
	}

	if errors.HasErrors() {
		return errors
	}

	return nil
}

// createWC3VerifiableCredential creates W3C-compliant credential
func (i *Issuer) createWC3VerifiableCredential(req *CredentialIssuanceRequest, credentialID string) *VerifiableCredential {
	expirationYears := getExpirationYears(req.CredentialType)
	if req.ExpirationYears > 0 {
		expirationYears = req.ExpirationYears
	}
	expirationDate := time.Now().AddDate(expirationYears, 0, 0)

	now := time.Now()

	vc := &VerifiableCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://www.w3.org/2018/credentials/examples/v1",
		},
		Type: []string{
			"VerifiableCredential",
			string(req.CredentialType),
		},
		ID:             fmt.Sprintf("urn:credential:%s", credentialID),
		Issuer:         i.config.IssuerConfig.IssuerDID,
		IssuanceDate:   now,
		ExpirationDate: &expirationDate,
		CredentialSubject: CredentialSubject{
			ID:                 req.SubjectDID,
			Claims:             req.Claims,
			VerificationClaims: req.VerificationClaims,
		},
		CredentialStatus: &CredentialStatus{
			ID:     fmt.Sprintf("urn:revocation:%s", credentialID),
			Type:   "CardanoRevocationRegistry2024",
			Status: "active",
		},
		Proof: i.createCredentialProof(credentialID),
	}

	return vc
}

// createCredentialProof creates proof for credential
func (i *Issuer) createCredentialProof(credentialID string) Proof {
	nonce, _ := i.cryptoUtils.GenerateNonce(32)

	return Proof{
		Type:               i.config.IssuerConfig.ProofType,
		Created:            time.Now(),
		VerificationMethod: fmt.Sprintf("%s#%s", i.config.IssuerConfig.IssuerDID, i.config.IssuerConfig.PublicKeyID),
		ProofPurpose:       "assertionMethod",
		SignatureAlgorithm: GetSignatureAlgorithm(i.config.IssuerConfig.SigningAlgorithm),
		ChallengeNonce:     nonce,
	}
}

// encodeCredential encodes credential in specified format
func (i *Issuer) encodeCredential(vc *VerifiableCredential, format CredentialFormat, credentialID string) (string, error) {
	switch format {
	case JSONLDFormat:
		credentialBytes, err := json.Marshal(vc)
		if err != nil {
			return "", NewCredentialErrorWithDetails(
				ErrCodeInvalidCredential,
				"failed to encode credential as JSON-LD",
				err.Error(),
			)
		}
		return string(credentialBytes), nil

	case JWTFormat:
		// Simple JWT encoding (in production, use proper JWT library)
		return i.createJWT(vc, credentialID)

	case SDJWTFormat:
		// SD-JWT encoding
		return i.createSDJWT(vc, credentialID)

	default:
		return "", NewCredentialError(
			ErrCodeUnsupportedFormat,
			fmt.Sprintf("unsupported credential format: %s", format),
		)
	}
}

// createJWT creates JWT representation of credential
func (i *Issuer) createJWT(vc *VerifiableCredential, credentialID string) (string, error) {
	// Create header and payload
	header := map[string]interface{}{
		"alg": GetSignatureAlgorithm(i.config.IssuerConfig.SigningAlgorithm),
		"typ": "JWT",
		"kid": i.config.IssuerConfig.PublicKeyID,
	}

	payload := map[string]interface{}{
		"vc":  vc,
		"iss": i.config.IssuerConfig.IssuerDID,
		"sub": vc.CredentialSubject.ID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(365 * 24 * time.Hour).Unix(),
		"jti": credentialID,
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerStr := string(headerJSON)
	payloadStr := string(payloadJSON)

	// Sign JWT
	jwt, err := i.cryptoUtils.CreateJWSSignature(headerStr, payloadStr, i.config.IssuerConfig.PrivateKeyPath)
	if err != nil {
		return "", NewCredentialErrorWithDetails(
			ErrCodeInvalidCredential,
			"failed to sign JWT",
			err.Error(),
		)
	}

	return jwt, nil
}

// createSDJWT creates SD-JWT representation of credential
func (i *Issuer) createSDJWT(vc *VerifiableCredential, credentialID string) (string, error) {
	// SD-JWT is JWT with selectively disclosable claims
	// For now, return JWT (full implementation would use sd-jwt library)
	return i.createJWT(vc, credentialID)
}

// GetIssuanceProgress gets issuance progress
func (i *Issuer) GetIssuanceProgress(credentialID string) *IssuanceProgress {
	i.progressMutex.RLock()
	defer i.progressMutex.RUnlock()
	return i.progressTracker[credentialID]
}

// setProgress sets issuance progress
func (i *Issuer) setProgress(credentialID string, progress *IssuanceProgress) {
	i.progressMutex.Lock()
	defer i.progressMutex.Unlock()
	i.progressTracker[credentialID] = progress
}

// cleanupProgress removes progress tracking after delay
func (i *Issuer) cleanupProgress(credentialID string) {
	time.AfterFunc(5*time.Minute, func() {
		i.progressMutex.Lock()
		defer i.progressMutex.Unlock()
		delete(i.progressTracker, credentialID)
	})
}

// Helper function to get expiration years for credential type
func getExpirationYears(credType CredentialType) int {
	switch credType {
	case ProofOfHumanity:
		return 1
	case KYCLite:
		return 1
	case HighAssurance:
		return 5
	case Professional:
		return 2
	default:
		return 1
	}
}

// RecoverCredential recovers a previously issued credential
func (i *Issuer) RecoverCredential(ctx context.Context, credentialID string) (*VerifiableCredential, error) {
	return i.storage.RetrieveCredential(ctx, credentialID)
}

// ListIssuedCredentials lists credentials issued to subject
func (i *Issuer) ListIssuedCredentials(ctx context.Context, subjectDID string) ([]*CredentialMetadata, error) {
	return i.storage.ListCredentialsBySubject(ctx, subjectDID)
}
