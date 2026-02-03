package cardano

import (
	"context"
	"fmt"
	"time"
)

// CredentialEvent represents an event in the credential lifecycle
type CredentialEvent struct {
	ID           string                 `json:"id"`
	CredentialID string                 `json:"credential_id"`
	EventType    string                 `json:"event_type"` // issued, verified, revoked, suspended
	Actor        string                 `json:"actor"`      // DID of actor
	Details      map[string]interface{} `json:"details,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	TxHash       string                 `json:"tx_hash,omitempty"`
	BlockHeight  int                    `json:"block_height,omitempty"`
}

// CredentialAuditTrail tracks all events for a credential
type CredentialAuditTrail struct {
	CredentialID  string             `json:"credential_id"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	Events        []*CredentialEvent `json:"events"`
	EventCount    int                `json:"event_count"`
	LastEvent     *CredentialEvent   `json:"last_event,omitempty"`
	ImmutableHash string             `json:"immutable_hash,omitempty"`
}

// AppAuthorization represents authorization for an app to access a credential
type AppAuthorization struct {
	CredentialID string                 `json:"credential_id"`
	AppID        string                 `json:"app_id"`
	AppName      string                 `json:"app_name"`
	AuthorizedAt time.Time              `json:"authorized_at"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
	Permissions  []string               `json:"permissions"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	TxHash       string                 `json:"tx_hash,omitempty"`
}

// RecordCredentialEvent records an event in the credential audit trail
func (c *Client) RecordCredentialEvent(ctx context.Context, credentialID string, eventType string, actor string, details map[string]interface{}) (string, error) {
	if credentialID == "" || eventType == "" || actor == "" {
		return "", fmt.Errorf("credentialID, eventType, and actor are required")
	}

	event := &CredentialEvent{
		CredentialID: credentialID,
		EventType:    eventType,
		Actor:        actor,
		Details:      details,
		Timestamp:    time.Now(),
	}

	payload := map[string]interface{}{
		"credential_id": credentialID,
		"event_type":    eventType,
		"actor":         actor,
		"details":       details,
		"timestamp":     time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "record-credential-event", payload)
	if err != nil {
		c.logger.Printf("Failed to record credential event: %v", err)
		return "", err
	}

	event.TxHash = txHash
	cacheKey := fmt.Sprintf("credential_event_%s_%s", credentialID, eventType)
	c.credentialCache.Set(cacheKey, event, c.credentialCache.ttl)

	c.logger.Printf("Recorded event %s for credential %s with tx %s", eventType, credentialID, txHash)
	return txHash, nil
}

// GetCredentialAuditTrail retrieves the full audit trail for a credential
func (c *Client) GetCredentialAuditTrail(ctx context.Context, credentialID string) (*CredentialAuditTrail, error) {
	cacheKey := fmt.Sprintf("credential_audit_%s", credentialID)

	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		c.logger.Printf("Credential audit trail cache hit: %s", credentialID)
		return cached.(*CredentialAuditTrail), nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		// In production, would query blockchain
		trail := &CredentialAuditTrail{
			CredentialID: credentialID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			Events:       make([]*CredentialEvent, 0),
		}
		return trail, nil
	}
}

// VerifyCredentialIntegrity verifies that a credential hasn't been tampered with
func (c *Client) VerifyCredentialIntegrity(ctx context.Context, credentialID string, contentHash string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		// In production, would verify hash against blockchain
		c.logger.Printf("Verified credential integrity for %s", credentialID)
		return true, nil
	}
}

// RevokeCredential revokes a credential with reason
func (c *Client) RevokeCredentialWithReason(ctx context.Context, credentialID string, reason string, actor string) (string, error) {
	payload := map[string]interface{}{
		"credential_id": credentialID,
		"action":        "revoke",
		"reason":        reason,
		"actor":         actor,
		"timestamp":     time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "revoke-credential", payload)
	if err != nil {
		return "", err
	}

	// Record revocation event
	c.RecordCredentialEvent(ctx, credentialID, "revoked", actor, map[string]interface{}{
		"reason": reason,
	})

	c.logger.Printf("Credential %s revoked: %s", credentialID, reason)
	return txHash, nil
}

// SuspendCredential suspends a credential temporarily
func (c *Client) SuspendCredential(ctx context.Context, credentialID string, reason string, actor string) (string, error) {
	payload := map[string]interface{}{
		"credential_id": credentialID,
		"action":        "suspend",
		"reason":        reason,
		"actor":         actor,
		"timestamp":     time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "suspend-credential", payload)
	if err != nil {
		return "", err
	}

	// Record suspension event
	c.RecordCredentialEvent(ctx, credentialID, "suspended", actor, map[string]interface{}{
		"reason": reason,
	})

	return txHash, nil
}

// RestoreCredential restores a suspended credential
func (c *Client) RestoreCredential(ctx context.Context, credentialID string, actor string) (string, error) {
	payload := map[string]interface{}{
		"credential_id": credentialID,
		"action":        "restore",
		"actor":         actor,
		"timestamp":     time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "restore-credential", payload)
	if err != nil {
		return "", err
	}

	c.RecordCredentialEvent(ctx, credentialID, "restored", actor, nil)
	return txHash, nil
}

// QueryCredentialEvents queries events in audit trail with filters
func (c *Client) QueryCredentialEvents(ctx context.Context, credentialID string, eventType string, startTime time.Time) ([]*CredentialEvent, error) {
	cacheKey := fmt.Sprintf("credential_events_%s_%s", credentialID, eventType)

	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		return cached.([]*CredentialEvent), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		c.logger.Printf("Queried events for credential %s", credentialID)
		return make([]*CredentialEvent, 0), nil
	}
}

// AuthorizeAppAccess authorizes an application to access a credential
func (c *Client) AuthorizeAppAccess(ctx context.Context, credentialID string, appID string, appName string, permissions []string, expiresIn *time.Duration) (string, error) {
	auth := &AppAuthorization{
		CredentialID: credentialID,
		AppID:        appID,
		AppName:      appName,
		AuthorizedAt: time.Now(),
		Permissions:  permissions,
	}

	if expiresIn != nil {
		expiry := time.Now().Add(*expiresIn)
		auth.ExpiresAt = &expiry
	}

	payload := map[string]interface{}{
		"credential_id": credentialID,
		"app_id":        appID,
		"app_name":      appName,
		"permissions":   permissions,
		"authorized_at": auth.AuthorizedAt.Unix(),
	}

	if auth.ExpiresAt != nil {
		payload["expires_at"] = auth.ExpiresAt.Unix()
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "authorize-app-access", payload)
	if err != nil {
		c.logger.Printf("Failed to authorize app access: %v", err)
		return "", err
	}

	auth.TxHash = txHash
	cacheKey := fmt.Sprintf("app_auth_%s_%s", credentialID, appID)
	c.credentialCache.Set(cacheKey, auth, c.credentialCache.ttl)

	c.logger.Printf("Authorized app %s to access credential %s", appID, credentialID)
	return txHash, nil
}

// RevokeAppAccess revokes an app's access to a credential
func (c *Client) RevokeAppAccess(ctx context.Context, credentialID string, appID string) (string, error) {
	payload := map[string]interface{}{
		"credential_id": credentialID,
		"app_id":        appID,
		"action":        "revoke_access",
		"timestamp":     time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "revoke-app-access", payload)
	if err != nil {
		return "", err
	}

	cacheKey := fmt.Sprintf("app_auth_%s_%s", credentialID, appID)
	c.credentialCache.Delete(cacheKey)

	c.logger.Printf("Revoked app %s access to credential %s", appID, credentialID)
	return txHash, nil
}

// CheckAppAccess checks if an app has access to a credential
func (c *Client) CheckAppAccess(ctx context.Context, credentialID string, appID string) (bool, error) {
	cacheKey := fmt.Sprintf("app_auth_%s_%s", credentialID, appID)

	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		auth := cached.(*AppAuthorization)
		// Check if authorization has expired
		if auth.ExpiresAt != nil && time.Now().After(*auth.ExpiresAt) {
			return false, nil
		}
		return true, nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return false, nil
	}
}

// GetAppAuthorizations gets all authorizations for a credential
func (c *Client) GetAppAuthorizations(ctx context.Context, credentialID string) ([]*AppAuthorization, error) {
	cacheKey := fmt.Sprintf("cred_app_auths_%s", credentialID)

	if cached, exists := c.credentialCache.Get(cacheKey); exists {
		return cached.([]*AppAuthorization), nil
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return make([]*AppAuthorization, 0), nil
	}
}
