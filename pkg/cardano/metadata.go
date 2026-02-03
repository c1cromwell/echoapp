package cardano

import (
	"encoding/json"
	"fmt"
	"time"
)

// MetadataLabel represents the Cardano metadata label for identity data
// Using standard labels: 721 (NFT metadata), 777 (custom apps)
const (
	MetadataLabelCredentials = 777
	MetadataLabelTrustLevel  = 778
	MetadataLabelAuditTrail  = 779
	MetadataLabelRevocation  = 780
)

// CardanoMetadata represents the structure of metadata stored in Cardano transactions
type CardanoMetadata struct {
	Label   uint64      `json:"label"`
	Version int         `json:"version"`
	Type    string      `json:"type"` // "credential", "trust-level", "audit", "revocation"
	Data    interface{} `json:"data"`
}

// OnChainCredentialMetadata represents the standardized format for storing credentials on-chain
type OnChainCredentialMetadata struct {
	// Core identifiers
	CredentialID   string `json:"credential_id"`
	UserID         string `json:"user_id"`
	SchemaID       string `json:"schema_id"`
	CredentialType string `json:"credential_type"`

	// Issuer information
	Issuer          string `json:"issuer"`
	IssuerPublicKey string `json:"issuer_public_key"`

	// Temporal information
	IssuedAt  int64 `json:"issued_at"`  // Unix timestamp
	ExpiresAt int64 `json:"expires_at"` // Unix timestamp, 0 if non-expiring
	NotBefore int64 `json:"not_before"` // Unix timestamp

	// Credential content hash
	ContentHash       string `json:"content_hash"`        // SHA256 hash of credential content
	CredentialDataURI string `json:"credential_data_uri"` // IPFS or HTTP URI to full credential

	// Metadata
	Metadata map[string]string `json:"metadata"`
	Tags     []string          `json:"tags"`

	// Portability
	Portable    bool     `json:"portable"`
	AllowedApps []string `json:"allowed_apps"`

	// Revocation
	Revoked          bool   `json:"revoked"`
	RevocationReason string `json:"revocation_reason,omitempty"`
	RevocationDate   int64  `json:"revocation_date,omitempty"`

	// Verification
	VerificationProof string `json:"verification_proof"` // Proof of credential on-chain
	ProofType         string `json:"proof_type"`         // "ed25519Signature2018", etc.
}

// OnChainTrustLevelMetadata represents the standardized format for storing trust levels on-chain
type OnChainTrustLevelMetadata struct {
	// Core identifiers
	UserID          string `json:"user_id"`
	TrustLevelValue string `json:"trust_level"` // unverified, device-verified, kyc-verified, organization-verified

	// Temporal information
	UpdatedAt int64 `json:"updated_at"` // Unix timestamp
	ExpiresAt int64 `json:"expires_at"` // Unix timestamp, 0 if non-expiring

	// Verification details
	VerificationMethod string `json:"verification_method"` // device, kyc, organization, multi-sig
	VerifierID         string `json:"verifier_id"`

	// Verification evidence
	VerificationProof string                 `json:"verification_proof"`
	VerificationData  map[string]interface{} `json:"verification_data"`
	BiometricHash     string                 `json:"biometric_hash,omitempty"` // For device verification
	DeviceFingerprint string                 `json:"device_fingerprint,omitempty"`

	// Trust level details
	Confidence int64             `json:"confidence"` // 0-100 confidence score
	RiskScore  int64             `json:"risk_score"` // 0-100 risk score
	Metadata   map[string]string `json:"metadata"`

	// Previous state for audit trail
	PreviousTrustLevel string `json:"previous_trust_level,omitempty"`
	ChangeReason       string `json:"change_reason"`

	// Multi-sig support
	RequiredSignatures int      `json:"required_signatures,omitempty"`
	Signers            []string `json:"signers,omitempty"`
}

// OnChainAuditMetadata represents audit trail entries stored on-chain
type OnChainAuditMetadata struct {
	// Event identifiers
	AuditID   string `json:"audit_id"`
	UserID    string `json:"user_id"`
	EventType string `json:"event_type"` // trust-level-change, credential-added, credential-revoked, etc.

	// Temporal information
	EventTimestamp int64 `json:"event_timestamp"` // Unix timestamp
	RecordedAt     int64 `json:"recorded_at"`     // When recorded on blockchain

	// Event details
	OldValue     string                 `json:"old_value,omitempty"`
	NewValue     string                 `json:"new_value"`
	ChangeReason string                 `json:"change_reason"`
	Metadata     map[string]interface{} `json:"metadata"`

	// Actor information
	ActorID        string `json:"actor_id"`   // Who initiated the change
	ActorType      string `json:"actor_type"` // user, system, verifier, admin
	ActorPublicKey string `json:"actor_public_key"`

	// Verification
	VerificationProof string `json:"verification_proof"`
	SignatureAlgo     string `json:"signature_algo"`

	// Chain of custody
	PreviousAuditID string `json:"previous_audit_id,omitempty"`
	AuditChainHash  string `json:"audit_chain_hash"` // Hash linking to previous audit entry
}

// OnChainRevocationMetadata represents revocation information stored on-chain
type OnChainRevocationMetadata struct {
	// Revocation identifiers
	RevocationID string `json:"revocation_id"`
	CredentialID string `json:"credential_id"`
	UserID       string `json:"user_id"`

	// Temporal information
	RevokedAt   int64 `json:"revoked_at"`   // Unix timestamp
	EffectiveAt int64 `json:"effective_at"` // When revocation becomes effective

	// Revocation details
	RevocationReason string                 `json:"revocation_reason"`
	RevocationScope  string                 `json:"revocation_scope"` // full, partial, conditional
	ScopedDetails    map[string]interface{} `json:"scoped_details,omitempty"`

	// Actor information
	RevokedBy        string `json:"revoked_by"`
	RevokerPublicKey string `json:"revoker_public_key"`

	// Recovery options
	Recoverable      bool   `json:"recoverable"`
	RecoveryDeadline int64  `json:"recovery_deadline,omitempty"`
	RecoveryMethod   string `json:"recovery_method,omitempty"`

	// Notification
	NotificationURL string            `json:"notification_url,omitempty"`
	Metadata        map[string]string `json:"metadata"`
}

// MetadataBuilder provides utilities for building Cardano metadata
type MetadataBuilder struct {
	metadata map[uint64]interface{}
}

// NewMetadataBuilder creates a new metadata builder
func NewMetadataBuilder() *MetadataBuilder {
	return &MetadataBuilder{
		metadata: make(map[uint64]interface{}),
	}
}

// AddCredentialMetadata adds credential metadata to the builder
func (mb *MetadataBuilder) AddCredentialMetadata(cred *OnChainCredentialMetadata) *MetadataBuilder {
	mb.metadata[MetadataLabelCredentials] = mb.buildMetadataObject("credential", cred)
	return mb
}

// AddTrustLevelMetadata adds trust level metadata to the builder
func (mb *MetadataBuilder) AddTrustLevelMetadata(tl *OnChainTrustLevelMetadata) *MetadataBuilder {
	mb.metadata[MetadataLabelTrustLevel] = mb.buildMetadataObject("trust-level", tl)
	return mb
}

// AddAuditMetadata adds audit metadata to the builder
func (mb *MetadataBuilder) AddAuditMetadata(audit *OnChainAuditMetadata) *MetadataBuilder {
	mb.metadata[MetadataLabelAuditTrail] = mb.buildMetadataObject("audit", audit)
	return mb
}

// AddRevocationMetadata adds revocation metadata to the builder
func (mb *MetadataBuilder) AddRevocationMetadata(rev *OnChainRevocationMetadata) *MetadataBuilder {
	mb.metadata[MetadataLabelRevocation] = mb.buildMetadataObject("revocation", rev)
	return mb
}

// buildMetadataObject builds a CardanoMetadata object
func (mb *MetadataBuilder) buildMetadataObject(mdType string, data interface{}) CardanoMetadata {
	return CardanoMetadata{
		Label:   MetadataLabelCredentials,
		Version: 1,
		Type:    mdType,
		Data:    data,
	}
}

// Build returns the compiled metadata map
func (mb *MetadataBuilder) Build() map[uint64]interface{} {
	return mb.metadata
}

// BuildJSON returns the metadata as JSON
func (mb *MetadataBuilder) BuildJSON() ([]byte, error) {
	return json.Marshal(mb.metadata)
}

// SerializeCredentialMetadata serializes credential metadata to JSON
func SerializeCredentialMetadata(cred *OnChainCredentialMetadata) ([]byte, error) {
	return json.Marshal(cred)
}

// SerializeTrustLevelMetadata serializes trust level metadata to JSON
func SerializeTrustLevelMetadata(tl *OnChainTrustLevelMetadata) ([]byte, error) {
	return json.Marshal(tl)
}

// SerializeAuditMetadata serializes audit metadata to JSON
func SerializeAuditMetadata(audit *OnChainAuditMetadata) ([]byte, error) {
	return json.Marshal(audit)
}

// SerializeRevocationMetadata serializes revocation metadata to JSON
func SerializeRevocationMetadata(rev *OnChainRevocationMetadata) ([]byte, error) {
	return json.Marshal(rev)
}

// DeserializeCredentialMetadata deserializes credential metadata from JSON
func DeserializeCredentialMetadata(data []byte) (*OnChainCredentialMetadata, error) {
	var cred OnChainCredentialMetadata
	if err := json.Unmarshal(data, &cred); err != nil {
		return nil, fmt.Errorf("failed to deserialize credential metadata: %w", err)
	}
	return &cred, nil
}

// DeserializeTrustLevelMetadata deserializes trust level metadata from JSON
func DeserializeTrustLevelMetadata(data []byte) (*OnChainTrustLevelMetadata, error) {
	var tl OnChainTrustLevelMetadata
	if err := json.Unmarshal(data, &tl); err != nil {
		return nil, fmt.Errorf("failed to deserialize trust level metadata: %w", err)
	}
	return &tl, nil
}

// DeserializeAuditMetadata deserializes audit metadata from JSON
func DeserializeAuditMetadata(data []byte) (*OnChainAuditMetadata, error) {
	var audit OnChainAuditMetadata
	if err := json.Unmarshal(data, &audit); err != nil {
		return nil, fmt.Errorf("failed to deserialize audit metadata: %w", err)
	}
	return &audit, nil
}

// DeserializeRevocationMetadata deserializes revocation metadata from JSON
func DeserializeRevocationMetadata(data []byte) (*OnChainRevocationMetadata, error) {
	var rev OnChainRevocationMetadata
	if err := json.Unmarshal(data, &rev); err != nil {
		return nil, fmt.Errorf("failed to deserialize revocation metadata: %w", err)
	}
	return &rev, nil
}

// CreateCredentialMetadata creates a new credential metadata instance
func CreateCredentialMetadata(credID, userID, schemaID, issuer string, expiresAt time.Time) *OnChainCredentialMetadata {
	return &OnChainCredentialMetadata{
		CredentialID: credID,
		UserID:       userID,
		SchemaID:     schemaID,
		Issuer:       issuer,
		IssuedAt:     time.Now().Unix(),
		ExpiresAt:    expiresAt.Unix(),
		Portable:     true,
		Metadata:     make(map[string]string),
		Tags:         make([]string, 0),
		AllowedApps:  make([]string, 0),
	}
}

// CreateTrustLevelMetadata creates a new trust level metadata instance
func CreateTrustLevelMetadata(userID, trustLevel, verificationMethod, verifierID string) *OnChainTrustLevelMetadata {
	return &OnChainTrustLevelMetadata{
		UserID:             userID,
		TrustLevelValue:    trustLevel,
		UpdatedAt:          time.Now().Unix(),
		VerificationMethod: verificationMethod,
		VerifierID:         verifierID,
		Metadata:           make(map[string]string),
		Confidence:         75,
		RiskScore:          25,
	}
}

// CreateAuditMetadata creates a new audit metadata instance
func CreateAuditMetadata(userID, eventType, newValue, reason, actorID string) *OnChainAuditMetadata {
	return &OnChainAuditMetadata{
		UserID:         userID,
		EventType:      eventType,
		NewValue:       newValue,
		ChangeReason:   reason,
		ActorID:        actorID,
		EventTimestamp: time.Now().Unix(),
		RecordedAt:     time.Now().Unix(),
		Metadata:       make(map[string]interface{}),
	}
}

// ValidateCredentialMetadata validates credential metadata
func ValidateCredentialMetadata(cred *OnChainCredentialMetadata) error {
	if cred.CredentialID == "" {
		return fmt.Errorf("credential ID is required")
	}
	if cred.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if cred.Issuer == "" {
		return fmt.Errorf("issuer is required")
	}
	if cred.IssuedAt == 0 {
		return fmt.Errorf("issued at timestamp is required")
	}
	if cred.IssuedAt > time.Now().Unix() {
		return fmt.Errorf("issued at timestamp cannot be in the future")
	}
	if cred.ExpiresAt > 0 && cred.ExpiresAt <= cred.IssuedAt {
		return fmt.Errorf("expiration date must be after issued date")
	}
	return nil
}

// ValidateTrustLevelMetadata validates trust level metadata
func ValidateTrustLevelMetadata(tl *OnChainTrustLevelMetadata) error {
	if tl.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if tl.TrustLevelValue == "" {
		return fmt.Errorf("trust level is required")
	}
	if tl.VerifierID == "" {
		return fmt.Errorf("verifier ID is required")
	}
	if tl.Confidence < 0 || tl.Confidence > 100 {
		return fmt.Errorf("confidence score must be between 0 and 100")
	}
	if tl.RiskScore < 0 || tl.RiskScore > 100 {
		return fmt.Errorf("risk score must be between 0 and 100")
	}
	return nil
}

// ValidateAuditMetadata validates audit metadata
func ValidateAuditMetadata(audit *OnChainAuditMetadata) error {
	if audit.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if audit.EventType == "" {
		return fmt.Errorf("event type is required")
	}
	if audit.ActorID == "" {
		return fmt.Errorf("actor ID is required")
	}
	return nil
}

// CalculateMetadataSize returns the approximate size of metadata in bytes
func CalculateMetadataSize(metadata CardanoMetadata) (int, error) {
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return 0, err
	}
	return len(jsonData), nil
}

// IsCredentialExpired checks if a credential has expired
func IsCredentialExpired(cred *OnChainCredentialMetadata) bool {
	if cred.ExpiresAt == 0 {
		return false
	}
	return cred.ExpiresAt < time.Now().Unix()
}

// IsTrustLevelExpired checks if a trust level has expired
func IsTrustLevelExpired(tl *OnChainTrustLevelMetadata) bool {
	if tl.ExpiresAt == 0 {
		return false
	}
	return tl.ExpiresAt < time.Now().Unix()
}
