package cardano

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// VerificationMethod represents how a user was verified
type VerificationMethod string

const (
	VerificationMethodAppleDigitalID VerificationMethod = "apple_digital_id"
	VerificationMethodThirdParty     VerificationMethod = "third_party_verification"
	VerificationMethodSelfCertified  VerificationMethod = "self_certified"
	VerificationMethodOrganizational VerificationMethod = "organizational_verification"
	VerificationMethodBiometric      VerificationMethod = "biometric"
	VerificationMethodGovernmentID   VerificationMethod = "government_id"
)

// TrustLevelRecord represents a trust level with verification method
type TrustLevelRecord struct {
	UserID              string                 `json:"user_id"`
	Level               string                 `json:"level"` // unverified, device-verified, kyc-verified, organization-verified
	VerificationMethod  VerificationMethod     `json:"verification_method"`
	VerificationDetails map[string]interface{} `json:"verification_details,omitempty"`
	VerifiedBy          string                 `json:"verified_by"` // DID or identifier of verifier
	VerificationDate    time.Time              `json:"verification_date"`
	ExpiresAt           *time.Time             `json:"expires_at,omitempty"`
	Confidence          float64                `json:"confidence"` // 0.0-1.0
	UpdatedAt           time.Time              `json:"updated_at"`
	UpdatedBy           string                 `json:"updated_by"` // Who made the update
	Reason              string                 `json:"reason"`
	OnChainHash         string                 `json:"on_chain_hash,omitempty"`
	TxHash              string                 `json:"tx_hash,omitempty"`
	Timestamp           time.Time              `json:"timestamp"`
}

// TrustLevelHistory tracks all trust level changes
type TrustLevelHistory struct {
	UserID       string              `json:"user_id"`
	CurrentLevel string              `json:"current_level"`
	HistoryCount int                 `json:"history_count"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Records      []*TrustLevelRecord `json:"records,omitempty"`
	LastUpdate   *TrustLevelRecord   `json:"last_update,omitempty"`
}

// VerificationRequest represents a request to verify a user
type VerificationRequest struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Method      VerificationMethod     `json:"method"`
	Details     map[string]interface{} `json:"details"`
	Status      string                 `json:"status"` // pending, approved, rejected
	RequestedAt time.Time              `json:"requested_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	ProcessedBy string                 `json:"processed_by,omitempty"`
	Reason      string                 `json:"reason,omitempty"`
	TxHash      string                 `json:"tx_hash,omitempty"`
}

// UpdateTrustLevelWithMethod updates a user's trust level with verification method
func (c *Client) UpdateTrustLevelWithMethod(ctx context.Context, userID string, level string, method VerificationMethod, verifiedBy string, confidence float64, reason string) (string, error) {
	if err := validateTrustLevelUpdate(userID, level, confidence); err != nil {
		return "", err
	}

	record := &TrustLevelRecord{
		UserID:             userID,
		Level:              level,
		VerificationMethod: method,
		VerifiedBy:         verifiedBy,
		VerificationDate:   time.Now(),
		Confidence:         confidence,
		UpdatedAt:          time.Now(),
		UpdatedBy:          verifiedBy,
		Reason:             reason,
	}

	// Generate on-chain hash
	contentBytes, _ := json.Marshal(map[string]interface{}{
		"user_id": userID,
		"level":   level,
		"method":  method,
		"date":    record.VerificationDate.Unix(),
	})
	hash := sha256.Sum256(contentBytes)
	record.OnChainHash = fmt.Sprintf("%x", hash)[:32]

	payload := map[string]interface{}{
		"user_id":             userID,
		"level":               level,
		"verification_method": method,
		"verified_by":         verifiedBy,
		"confidence":          confidence,
		"reason":              reason,
		"on_chain_hash":       record.OnChainHash,
		"timestamp":           record.VerificationDate.Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "update-trust-level", payload)
	if err != nil {
		c.logger.Printf("Failed to update trust level: %v", err)
		return "", err
	}

	record.TxHash = txHash
	record.Timestamp = time.Now()

	cacheKey := fmt.Sprintf("trust_%s", userID)
	c.trustLevelCache.Set(cacheKey, record, c.trustLevelCache.ttl)

	c.logger.Printf("Updated trust level for %s to %s via %s with tx %s", userID, level, method, txHash)
	return txHash, nil
}

// GetTrustLevelWithHistory retrieves current trust level and history
func (c *Client) GetTrustLevelWithHistory(ctx context.Context, userID string) (*TrustLevelHistory, error) {
	cacheKey := fmt.Sprintf("trust_history_%s", userID)

	if cached, exists := c.trustLevelCache.Get(cacheKey); exists {
		c.logger.Printf("Trust level history cache hit: %s", userID)
		return cached.(*TrustLevelHistory), nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		history := &TrustLevelHistory{
			UserID:       userID,
			CurrentLevel: "unverified",
			HistoryCount: 0,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			Records:      make([]*TrustLevelRecord, 0),
		}
		return history, nil
	}
}

// GetTrustLevelRecord retrieves the current trust level record
func (c *Client) GetTrustLevelRecord(ctx context.Context, userID string) (*TrustLevelRecord, error) {
	cacheKey := fmt.Sprintf("trust_%s", userID)

	if cached, exists := c.trustLevelCache.Get(cacheKey); exists {
		return cached.(*TrustLevelRecord), nil
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return &TrustLevelRecord{
			UserID:     userID,
			Level:      "unverified",
			UpdatedAt:  time.Now(),
			Confidence: 0.0,
		}, nil
	}
}

// VerifyUserWithAppleDigitalID verifies a user using Apple Digital ID
func (c *Client) VerifyUserWithAppleDigitalID(ctx context.Context, userID string, appleUserID string, certificationDetails map[string]interface{}) (string, error) {
	return c.UpdateTrustLevelWithMethod(
		ctx,
		userID,
		"device-verified",
		VerificationMethodAppleDigitalID,
		"apple_system",
		0.95,
		"Verified via Apple Digital ID",
	)
}

// VerifyUserWithThirdParty verifies a user through a third-party service
func (c *Client) VerifyUserWithThirdParty(ctx context.Context, userID string, verifierDID string, verificationData map[string]interface{}) (string, error) {
	return c.UpdateTrustLevelWithMethod(
		ctx,
		userID,
		"kyc-verified",
		VerificationMethodThirdParty,
		verifierDID,
		0.90,
		fmt.Sprintf("Verified by %s", verifierDID),
	)
}

// VerifyUserWithOrganization verifies a user through organizational verification
func (c *Client) VerifyUserWithOrganization(ctx context.Context, userID string, organizationDID string, employeeID string) (string, error) {
	return c.UpdateTrustLevelWithMethod(
		ctx,
		userID,
		"organization-verified",
		VerificationMethodOrganizational,
		organizationDID,
		0.98,
		fmt.Sprintf("Verified by organization %s", organizationDID),
	)
}

// CreateVerificationRequest creates a new verification request
func (c *Client) CreateVerificationRequest(ctx context.Context, userID string, method VerificationMethod, details map[string]interface{}) (string, error) {
	request := &VerificationRequest{
		ID:          fmt.Sprintf("vreq_%d", time.Now().UnixNano()),
		UserID:      userID,
		Method:      method,
		Details:     details,
		Status:      "pending",
		RequestedAt: time.Now(),
	}

	payload := map[string]interface{}{
		"request_id": request.ID,
		"user_id":    userID,
		"method":     method,
		"details":    details,
		"timestamp":  request.RequestedAt.Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "create-verification-request", payload)
	if err != nil {
		return "", err
	}

	request.TxHash = txHash
	cacheKey := fmt.Sprintf("vreq_%s", request.ID)
	c.trustLevelCache.Set(cacheKey, request, c.trustLevelCache.ttl)

	c.logger.Printf("Created verification request %s for user %s", request.ID, userID)
	return request.ID, nil
}

// ApproveVerificationRequest approves a verification request
func (c *Client) ApproveVerificationRequest(ctx context.Context, requestID string, approverDID string) (string, error) {
	cacheKey := fmt.Sprintf("vreq_%s", requestID)

	var request *VerificationRequest
	if cached, exists := c.trustLevelCache.Get(cacheKey); exists {
		request = cached.(*VerificationRequest)
	} else {
		return "", fmt.Errorf("verification request not found: %s", requestID)
	}

	now := time.Now()
	request.Status = "approved"
	request.ProcessedAt = &now
	request.ProcessedBy = approverDID

	payload := map[string]interface{}{
		"request_id": requestID,
		"action":     "approve",
		"approver":   approverDID,
		"timestamp":  now.Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "approve-verification", payload)
	if err != nil {
		return "", err
	}

	request.TxHash = txHash
	c.trustLevelCache.Set(cacheKey, request, c.trustLevelCache.ttl)

	// Update trust level based on verification method
	c.UpdateTrustLevelWithMethod(ctx, request.UserID, "kyc-verified", request.Method, approverDID, 0.85, "Verification approved")

	c.logger.Printf("Approved verification request %s", requestID)
	return txHash, nil
}

// RejectVerificationRequest rejects a verification request
func (c *Client) RejectVerificationRequest(ctx context.Context, requestID string, rejectorDID string, reason string) (string, error) {
	cacheKey := fmt.Sprintf("vreq_%s", requestID)

	var request *VerificationRequest
	if cached, exists := c.trustLevelCache.Get(cacheKey); exists {
		request = cached.(*VerificationRequest)
	} else {
		return "", fmt.Errorf("verification request not found: %s", requestID)
	}

	now := time.Now()
	request.Status = "rejected"
	request.ProcessedAt = &now
	request.ProcessedBy = rejectorDID
	request.Reason = reason

	payload := map[string]interface{}{
		"request_id": requestID,
		"action":     "reject",
		"rejector":   rejectorDID,
		"reason":     reason,
		"timestamp":  now.Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "reject-verification", payload)
	if err != nil {
		return "", err
	}

	request.TxHash = txHash
	c.trustLevelCache.Set(cacheKey, request, c.trustLevelCache.ttl)

	c.logger.Printf("Rejected verification request %s: %s", requestID, reason)
	return txHash, nil
}

// DowngradeTrustLevel downgrades a user's trust level
func (c *Client) DowngradeTrustLevel(ctx context.Context, userID string, reason string, actor string) (string, error) {
	payload := map[string]interface{}{
		"user_id":   userID,
		"action":    "downgrade",
		"reason":    reason,
		"actor":     actor,
		"timestamp": time.Now().Unix(),
	}

	txHash, err := c.simulateCardanoSubmission(ctx, "downgrade-trust-level", payload)
	if err != nil {
		return "", err
	}

	cacheKey := fmt.Sprintf("trust_%s", userID)
	c.trustLevelCache.Delete(cacheKey)

	c.logger.Printf("Downgraded trust level for %s: %s", userID, reason)
	return txHash, nil
}

// validateTrustLevelUpdate validates trust level update parameters
func validateTrustLevelUpdate(userID string, level string, confidence float64) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if level == "" {
		return fmt.Errorf("trust level is required")
	}
	validLevels := map[string]bool{
		"unverified":            true,
		"device-verified":       true,
		"kyc-verified":          true,
		"organization-verified": true,
	}
	if !validLevels[level] {
		return fmt.Errorf("invalid trust level: %s", level)
	}
	if confidence < 0 || confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0 and 1.0")
	}
	return nil
}
