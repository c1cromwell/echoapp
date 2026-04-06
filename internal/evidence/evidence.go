// Package evidence provides a client for Constellation's Digital Evidence API.
//
// Digital Evidence anchors cryptographic fingerprints (SHA-256 hashes) on the
// Hypergraph, providing a public verification explorer, "Smart Checkmark"
// certification, and enterprise compliance packaging.
//
// This is a premium enterprise feature for Organization tier clients.
// Free-tier users rely on ECHO's core Merkle root pipeline for message integrity.
package evidence

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

// UserTier determines Digital Evidence access level.
type UserTier string

const (
	TierFree         UserTier = "free"         // standard Merkle anchoring only
	TierVIP          UserTier = "vip"          // optional user-initiated media fingerprinting
	TierOrganization UserTier = "organization" // automatic audit + Smart Checkmark + compliance
)

// EvidenceType categorizes the type of evidence being anchored.
type EvidenceType string

const (
	EvidenceAuditTrail    EvidenceType = "audit_trail"
	EvidenceMediaAuth     EvidenceType = "media_authenticity"
	EvidenceSmartCheck    EvidenceType = "smart_checkmark"
	EvidenceRetentionProof EvidenceType = "retention_proof"
)

// FingerprintRequest represents a submission to the Digital Evidence API.
type FingerprintRequest struct {
	Hash           string       `json:"hash"`            // SHA-256 hex string
	EvidenceType   EvidenceType `json:"evidence_type"`
	Metadata       EventMeta    `json:"metadata"`
	OrganizationID string       `json:"organization_id"`
	TenantID       string       `json:"tenant_id"`
}

// EventMeta contains contextual metadata for an evidence fingerprint.
type EventMeta struct {
	Description string `json:"description"`
	SourceDID   string `json:"source_did,omitempty"`
	BatchID     string `json:"batch_id,omitempty"`
	IPFSCid     string `json:"ipfs_cid,omitempty"`
	Timestamp   string `json:"timestamp"`
}

// FingerprintResponse is returned after successful evidence submission.
type FingerprintResponse struct {
	EventID        string `json:"event_id"`
	ExplorerURL    string `json:"explorer_url"`    // public verification URL
	SmartCheckmark bool   `json:"smart_checkmark"` // whether Smart Checkmark was issued
	AnchoredAt     string `json:"anchored_at"`
	Status         string `json:"status"`
}

// EvidenceRecord is a stored record of a Digital Evidence submission.
type EvidenceRecord struct {
	EventID      string       `json:"event_id"`
	Hash         string       `json:"hash"`
	EvidenceType EvidenceType `json:"evidence_type"`
	ExplorerURL  string       `json:"explorer_url"`
	SubmittedAt  time.Time    `json:"submitted_at"`
	SubmittedBy  string       `json:"submitted_by"` // DID of submitter
	Tier         UserTier     `json:"tier"`
}

// ClientConfig holds configuration for the Digital Evidence API client.
type ClientConfig struct {
	APIKey         string `json:"api_key"`
	OrganizationID string `json:"organization_id"`
	TenantID       string `json:"tenant_id"`
	BaseURL        string `json:"base_url"`
}

// Validate checks that the client configuration is complete.
func (c *ClientConfig) Validate() error {
	if c.APIKey == "" {
		return errors.New("api_key is required")
	}
	if c.OrganizationID == "" {
		return errors.New("organization_id is required")
	}
	if c.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if c.BaseURL == "" {
		return errors.New("base_url is required")
	}
	return nil
}

// ComputeFingerprint returns the SHA-256 hex digest of the given data.
func ComputeFingerprint(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// CanAccessEvidence checks whether a user tier has access to Digital Evidence.
func CanAccessEvidence(tier UserTier, evidenceType EvidenceType) bool {
	switch tier {
	case TierFree:
		return false
	case TierVIP:
		return evidenceType == EvidenceMediaAuth
	case TierOrganization:
		return true
	default:
		return false
	}
}

// AuditBatchFingerprint creates a fingerprint request for an IPFS audit batch.
// Used after each IPFS batch push by the Go backend.
func AuditBatchFingerprint(config *ClientConfig, ipfsCID, batchMetadataHash, sourceDID string) (*FingerprintRequest, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	combined := ipfsCID + batchMetadataHash
	hash := ComputeFingerprint([]byte(combined))

	return &FingerprintRequest{
		Hash:         hash,
		EvidenceType: EvidenceAuditTrail,
		Metadata: EventMeta{
			Description: "Encrypted conversation log batch audit fingerprint",
			SourceDID:   sourceDID,
			IPFSCid:     ipfsCID,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		OrganizationID: config.OrganizationID,
		TenantID:       config.TenantID,
	}, nil
}

// MediaFingerprint creates a fingerprint request for media authenticity verification.
// iOS client computes hash before E2E encryption and submits via this endpoint.
func MediaFingerprint(config *ClientConfig, mediaHash, sourceDID string) (*FingerprintRequest, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return &FingerprintRequest{
		Hash:         mediaHash,
		EvidenceType: EvidenceMediaAuth,
		Metadata: EventMeta{
			Description: "Media authenticity fingerprint (pre-encryption hash)",
			SourceDID:   sourceDID,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		OrganizationID: config.OrganizationID,
		TenantID:       config.TenantID,
	}, nil
}

// SmartCheckmarkFingerprint creates a fingerprint for Smart Checkmark verification.
// Hash of message content + timestamp + sender DID.
func SmartCheckmarkFingerprint(config *ClientConfig, messageContentHash, timestamp, senderDID string) (*FingerprintRequest, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	combined := messageContentHash + timestamp + senderDID
	hash := ComputeFingerprint([]byte(combined))

	return &FingerprintRequest{
		Hash:         hash,
		EvidenceType: EvidenceSmartCheck,
		Metadata: EventMeta{
			Description: "Smart Checkmark message verification fingerprint",
			SourceDID:   senderDID,
			Timestamp:   timestamp,
		},
		OrganizationID: config.OrganizationID,
		TenantID:       config.TenantID,
	}, nil
}

// RetentionProofFingerprint creates a fingerprint proving data was retained and deleted per policy.
func RetentionProofFingerprint(config *ClientConfig, auditBatchHash, deletionConfirmation, sourceDID string) (*FingerprintRequest, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	combined := auditBatchHash + deletionConfirmation
	hash := ComputeFingerprint([]byte(combined))

	return &FingerprintRequest{
		Hash:         hash,
		EvidenceType: EvidenceRetentionProof,
		Metadata: EventMeta{
			Description: "Data retention proof at regulatory boundary",
			SourceDID:   sourceDID,
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
		},
		OrganizationID: config.OrganizationID,
		TenantID:       config.TenantID,
	}, nil
}
