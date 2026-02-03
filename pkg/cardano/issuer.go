package cardano

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// IssuerStatus represents the verification status of an issuer
type IssuerStatus string

const (
	IssuerStatusPending   IssuerStatus = "pending"
	IssuerStatusVerified  IssuerStatus = "verified"
	IssuerStatusSuspended IssuerStatus = "suspended"
	IssuerStatusRevoked   IssuerStatus = "revoked"
)

// AuthorityLevel represents the authority level of an issuer
type AuthorityLevel string

const (
	AuthorityLevelBasic    AuthorityLevel = "basic"
	AuthorityLevelStandard AuthorityLevel = "standard"
	AuthorityLevelPlatform AuthorityLevel = "platform"
)

// IssuerRegistration represents an issuer on the blockchain
type IssuerRegistration struct {
	IssuerDID       string                 `json:"issuer_did"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Status          IssuerStatus           `json:"status"`
	AuthorityLevel  AuthorityLevel         `json:"authority_level"`
	RegisteredAt    time.Time              `json:"registered_at"`
	VerifiedAt      *time.Time             `json:"verified_at,omitempty"`
	VerificationURL string                 `json:"verification_url,omitempty"`
	PublicKey       string                 `json:"public_key"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Credentials     []string               `json:"issued_credentials,omitempty"`
	CredentialCount int                    `json:"credential_count"`
	Rating          float64                `json:"rating,omitempty"`          // 0-5.0
	TrustScore      float64                `json:"trust_score,omitempty"`     // 0-100
	RevocationList  string                 `json:"revocation_list,omitempty"` // URI
	OnChainHash     string                 `json:"on_chain_hash,omitempty"`
	TxHash          string                 `json:"tx_hash,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
}

// IssuerAuditEntry tracks issuer status changes
type IssuerAuditEntry struct {
	ID        string    `json:"id"`
	IssuerDID string    `json:"issuer_did"`
	Action    string    `json:"action"` // registered, verified, suspended, revoked
	OldStatus string    `json:"old_status,omitempty"`
	NewStatus string    `json:"new_status"`
	Reason    string    `json:"reason,omitempty"`
	ActorDID  string    `json:"actor_did"` // Who performed the action
	Timestamp time.Time `json:"timestamp"`
	TxHash    string    `json:"tx_hash,omitempty"`
}

// RegisterIssuer registers an issuer on the Cardano blockchain
func (c *Client) RegisterIssuer(ctx context.Context, issuer *IssuerRegistration) (string, error) {
	if err := validateIssuer(issuer); err != nil {
		c.logger.Printf("Issuer validation failed: %v", err)
		return "", err
	}

	// Generate on-chain hash
	contentBytes, _ := json.Marshal(map[string]interface{}{
		"issuer_did":      issuer.IssuerDID,
		"name":            issuer.Name,
		"authority_level": issuer.AuthorityLevel,
	})
	hash := sha256.Sum256(contentBytes)
	issuer.OnChainHash = fmt.Sprintf("%x", hash)[:32]

	// Set registration timestamp
	issuer.RegisteredAt = time.Now()
	issuer.Status = IssuerStatusPending

	payload := map[string]interface{}{
		"issuer_did":      issuer.IssuerDID,
		"name":            issuer.Name,
		"description":     issuer.Description,
		"authority_level": issuer.AuthorityLevel,
		"public_key":      issuer.PublicKey,
		"on_chain_hash":   issuer.OnChainHash,
		"registered_at":   issuer.RegisteredAt.Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "register-issuer", payload)
	if err != nil {
		c.logger.Printf("Failed to register issuer: %v", err)
		return "", err
	}

	issuer.TxHash = txHash
	issuer.Timestamp = time.Now()

	c.logger.Printf("Issuer %s registered with tx %s", issuer.IssuerDID, txHash)
	return txHash, nil
}

// GetIssuer retrieves issuer information
func (c *Client) GetIssuer(ctx context.Context, issuerDID string) (*IssuerRegistration, error) {
	cacheKey := fmt.Sprintf("issuer_%s", issuerDID)

	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		c.logger.Printf("Issuer cache hit: %s", issuerDID)
		return cached.(*IssuerRegistration), nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, nil
	}
}

// VerifyIssuer verifies an issuer and updates status
func (c *Client) VerifyIssuer(ctx context.Context, issuerDID string, verificationURL string, authorityLevel AuthorityLevel) (string, error) {
	now := time.Now()

	payload := map[string]interface{}{
		"issuer_did":       issuerDID,
		"action":           "verify",
		"verification_url": verificationURL,
		"authority_level":  authorityLevel,
		"verified_at":      now.Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "verify-issuer", payload)
	if err != nil {
		return "", err
	}

	// Create audit entry
	c.createIssuerAuditEntry(ctx, &IssuerAuditEntry{
		IssuerDID: issuerDID,
		Action:    "verified",
		OldStatus: string(IssuerStatusPending),
		NewStatus: string(IssuerStatusVerified),
		Timestamp: now,
		TxHash:    txHash,
	})

	c.logger.Printf("Issuer %s verified with tx %s", issuerDID, txHash)
	return txHash, nil
}

// SuspendIssuer suspends an issuer
func (c *Client) SuspendIssuer(ctx context.Context, issuerDID string, reason string) (string, error) {
	payload := map[string]interface{}{
		"issuer_did": issuerDID,
		"action":     "suspend",
		"reason":     reason,
		"timestamp":  time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "suspend-issuer", payload)
	if err != nil {
		return "", err
	}

	c.createIssuerAuditEntry(ctx, &IssuerAuditEntry{
		IssuerDID: issuerDID,
		Action:    "suspended",
		NewStatus: string(IssuerStatusSuspended),
		Reason:    reason,
		Timestamp: time.Now(),
		TxHash:    txHash,
	})

	c.logger.Printf("Issuer %s suspended: %s", issuerDID, reason)
	return txHash, nil
}

// RevokeIssuer revokes an issuer
func (c *Client) RevokeIssuer(ctx context.Context, issuerDID string, reason string) (string, error) {
	payload := map[string]interface{}{
		"issuer_did": issuerDID,
		"action":     "revoke",
		"reason":     reason,
		"timestamp":  time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "revoke-issuer", payload)
	if err != nil {
		return "", err
	}

	c.createIssuerAuditEntry(ctx, &IssuerAuditEntry{
		IssuerDID: issuerDID,
		Action:    "revoked",
		NewStatus: string(IssuerStatusRevoked),
		Reason:    reason,
		Timestamp: time.Now(),
		TxHash:    txHash,
	})

	c.logger.Printf("Issuer %s revoked: %s", issuerDID, reason)
	return txHash, nil
}

// GetIssuerAuditTrail retrieves all audit entries for an issuer
func (c *Client) GetIssuerAuditTrail(ctx context.Context, issuerDID string) ([]*IssuerAuditEntry, error) {
	cacheKey := fmt.Sprintf("issuer_audit_%s", issuerDID)

	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		return cached.([]*IssuerAuditEntry), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		c.logger.Printf("Retrieved audit trail for issuer %s", issuerDID)
		return make([]*IssuerAuditEntry, 0), nil
	}
}

// GetIssuerCredentials retrieves all credentials issued by an issuer
func (c *Client) GetIssuerCredentials(ctx context.Context, issuerDID string) ([]*Credential, error) {
	cacheKey := fmt.Sprintf("issuer_creds_%s", issuerDID)

	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		return cached.([]*Credential), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return make([]*Credential, 0), nil
	}
}

// UpdateIssuerMetadata updates issuer metadata on-chain
func (c *Client) UpdateIssuerMetadata(ctx context.Context, issuerDID string, metadata map[string]interface{}) (string, error) {
	payload := map[string]interface{}{
		"issuer_did": issuerDID,
		"metadata":   metadata,
		"timestamp":  time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "update-issuer-metadata", payload)
	if err != nil {
		return "", err
	}

	c.logger.Printf("Updated metadata for issuer %s", issuerDID)
	return txHash, nil
}

// createIssuerAuditEntry creates an audit entry for issuer actions
func (c *Client) createIssuerAuditEntry(ctx context.Context, entry *IssuerAuditEntry) {
	// In production, this would be stored on-chain
	c.logger.Printf("Audit entry created for issuer %s: %s -> %s", entry.IssuerDID, entry.OldStatus, entry.NewStatus)
}

// validateIssuer validates issuer registration data
func validateIssuer(issuer *IssuerRegistration) error {
	if issuer.IssuerDID == "" {
		return fmt.Errorf("issuer DID is required")
	}
	if issuer.Name == "" {
		return fmt.Errorf("issuer name is required")
	}
	if issuer.PublicKey == "" {
		return fmt.Errorf("public key is required")
	}
	if issuer.AuthorityLevel == "" {
		issuer.AuthorityLevel = AuthorityLevelBasic
	}
	return nil
}
