package credentials

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Issuer manages credential issuance
type Issuer struct {
	config          *Config
	cryptoUtils     *CryptoUtils
	storage         Storage
	statusList      *StatusListPublisher
	progressTracker map[string]*IssuanceProgress
	progressMutex   sync.RWMutex
	semaphore       chan struct{} // Limit concurrent issuance
	privKey         *ecdsa.PrivateKey
	privMu          sync.Mutex
}

// NewIssuer creates new credential issuer
func NewIssuer(config *Config, cryptoUtils *CryptoUtils, storage Storage, statusList *StatusListPublisher) *Issuer {
	return &Issuer{
		config:          config,
		cryptoUtils:     cryptoUtils,
		storage:         storage,
		statusList:      statusList,
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

	statusIdx := 0
	if i.statusList != nil {
		var err error
		statusIdx, err = i.statusList.AllocateIndex(credentialID)
		if err != nil {
			return nil, NewCredentialErrorWithDetails(ErrCodeIssuanceFailed, "status list slot allocation failed", err.Error())
		}
	}
	statusIdxStr := strconv.Itoa(statusIdx)

	// Create credential
	vc := i.createWC3VerifiableCredential(req, credentialID, statusIdxStr)
	progress.Progress = 30
	progress.CurrentStep = "Signing credential"
	progress.UpdatedAt = time.Now()
	i.setProgress(credentialID, progress)

	if i.config.CredentialConfig.UseW3CVC2 {
		if err := i.signVC2DataIntegrity(vc); err != nil {
			progress.Status = "failed"
			progress.ErrorMessage = err.Error()
			i.setProgress(credentialID, progress)
			return nil, err
		}
	}

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

	if i.config.MetagraphConfig.EnableAnchor && i.config.MetagraphConfig.IdentityL1URL != "" {
		progress.Progress = 65
		progress.CurrentStep = "Anchoring VC issuance metadata to Identity L1"
		progress.UpdatedAt = time.Now()
		i.setProgress(credentialID, progress)
		if _, err := i.publishVCIssuanceMetadata(ctx, credentialID, req, vc.IssuanceDate); err != nil {
			fmt.Printf("Warning: Identity L1 VC issuance anchor failed: %v\n", err)
		}
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
func (i *Issuer) createWC3VerifiableCredential(req *CredentialIssuanceRequest, credentialID, statusListIndex string) *VerifiableCredential {
	expirationYears := getExpirationYears(req.CredentialType)
	if req.ExpirationYears > 0 {
		expirationYears = req.ExpirationYears
	}
	expirationDate := time.Now().AddDate(expirationYears, 0, 0)

	now := time.Now()

	baseURL := i.config.CredentialConfig.StatusListCredentialBaseURL
	if baseURL == "" {
		baseURL = "https://identity-metagraph.echo.app/status"
	}
	listID := fmt.Sprintf("%s/0", baseURL)
	if statusListIndex == "" {
		statusListIndex = "0"
	}
	statusURI := fmt.Sprintf("%s#%s", listID, statusListIndex)

	var context []string
	var credStatus *CredentialStatus
	if i.config.CredentialConfig.UseW3CVC2 {
		context = []string{
			"https://www.w3.org/ns/credentials/v2",
			"https://w3id.org/security/multikey/v1",
		}
		credStatus = &CredentialStatus{
			ID:                   statusURI,
			Type:                 "StatusList2021Entry",
			Status:               "active",
			StatusPurpose:        "revocation",
			StatusListIndex:      statusListIndex,
			StatusListCredential: listID,
		}
	} else {
		context = []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://www.w3.org/2018/credentials/examples/v1",
		}
		credStatus = &CredentialStatus{
			ID:     fmt.Sprintf("urn:revocation:%s", credentialID),
			Type:   "StatusList2021Entry",
			Status: "active",
		}
	}

	vc := &VerifiableCredential{
		Context: context,
		Type: []string{
			"VerifiableCredential",
			string(req.CredentialType),
		},
		ID:             fmt.Sprintf("urn:credential:%s", credentialID),
		Issuer:         i.effectiveIssuerDID(),
		IssuanceDate:   now,
		ExpirationDate: &expirationDate,
		CredentialSubject: CredentialSubject{
			ID:                 req.SubjectDID,
			Claims:             req.Claims,
			VerificationClaims: req.VerificationClaims,
		},
		CredentialStatus: credStatus,
		Proof:            i.createCredentialProof(credentialID),
	}

	return vc
}

// createCredentialProof creates proof for credential
func (i *Issuer) createCredentialProof(credentialID string) Proof {
	nonce, _ := i.cryptoUtils.GenerateNonce(32)

	if i.config.CredentialConfig.UseW3CVC2 {
		return Proof{
			Type:                 "DataIntegrityProof",
			Cryptosuite:          "ecdsa-2019",
			Created:              time.Now().UTC(),
			VerificationMethod:   i.verificationMethodForIssuer(),
			ProofPurpose:         "assertionMethod",
			ChallengeNonce:       nonce,
			ProofValue:           "",
		}
	}

	return Proof{
		Type:               i.config.IssuerConfig.ProofType,
		Created:            time.Now(),
		VerificationMethod: i.verificationMethodForIssuer(),
		ProofPurpose:       "assertionMethod",
		SignatureAlgorithm: GetSignatureAlgorithm(i.config.IssuerConfig.SigningAlgorithm),
		ChallengeNonce:     nonce,
	}
}

// encodeCredential encodes credential in specified format
func (i *Issuer) encodeCredential(vc *VerifiableCredential, format CredentialFormat, credentialID string) (string, error) {
	switch format {
	case JSONLDFormat:
		if i.config.CredentialConfig.UseW3CVC2 {
			return i.marshalVC2JSONLD(vc)
		}
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
	issuerDID := i.effectiveIssuerDID()

	if i.config.CredentialConfig.UseW3CVC2 {
		vcMap, err := vc2CredentialDocMap(vc, true)
		if err != nil {
			return "", err
		}
		header := map[string]interface{}{
			"alg": "ES256",
			"typ": "JWT",
			"kid": i.config.IssuerConfig.PublicKeyID,
		}
		if header["kid"] == "" && len(issuerDID) > len("did:key:") {
			header["kid"] = issuerDID[len("did:key:"):]
		}

		payload := map[string]interface{}{
			"vc":  vcMap,
			"iss": issuerDID,
			"sub": vc.CredentialSubject.ID,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(365 * 24 * time.Hour).Unix(),
			"jti": credentialID,
		}

		headerJSON, err := json.Marshal(header)
		if err != nil {
			return "", err
		}
		payloadJSON, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}

		priv, err := i.loadIssuerPrivateKey()
		if err != nil {
			return "", NewCredentialErrorWithDetails(ErrCodeInvalidCredential, "JWT ES256 requires issuer PEM private key", err.Error())
		}

		return i.cryptoUtils.CreateJWS_ES256(string(headerJSON), string(payloadJSON), priv)
	}

	header := map[string]interface{}{
		"alg": GetSignatureAlgorithm(i.config.IssuerConfig.SigningAlgorithm),
		"typ": "JWT",
		"kid": i.config.IssuerConfig.PublicKeyID,
	}

	payload := map[string]interface{}{
		"vc":  vc,
		"iss": issuerDID,
		"sub": vc.CredentialSubject.ID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(365 * 24 * time.Hour).Unix(),
		"jti": credentialID,
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerStr := string(headerJSON)
	payloadStr := string(payloadJSON)

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
		return 2
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
