package cardano

import (
	"encoding/json"
	"time"
)

// CredentialMetadata provides structured metadata for credentials
type CredentialMetadata struct {
	Issuer      string    `json:"issuer"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	SchemaID    string    `json:"schema_id"`
	CredType    string    `json:"cred_type"`
	Description string    `json:"description"`
}

// CredentialSubject represents the subject of a credential
type CredentialSubject struct {
	ID     string                 `json:"id"`
	Claims map[string]interface{} `json:"claims"`
}

// VerifiableCredential represents a W3C-compliant verifiable credential
type VerifiableCredential struct {
	Context           []string          `json:"@context"`
	Type              []string          `json:"type"`
	Issuer            string            `json:"issuer"`
	IssuanceDate      time.Time         `json:"issuanceDate"`
	ExpirationDate    time.Time         `json:"expirationDate"`
	CredentialSubject CredentialSubject `json:"credentialSubject"`
	Proof             interface{}       `json:"proof"`
}

// TrustLevelEnum represents the allowed trust levels
type TrustLevelEnum string

const (
	TrustLevelUnverified           TrustLevelEnum = "unverified"
	TrustLevelDeviceVerified       TrustLevelEnum = "device-verified"
	TrustLevelKYCVerified          TrustLevelEnum = "kyc-verified"
	TrustLevelOrganizationVerified TrustLevelEnum = "organization-verified"
)

// IsValid checks if the trust level is valid
func (t TrustLevelEnum) IsValid() bool {
	switch t {
	case TrustLevelUnverified, TrustLevelDeviceVerified, TrustLevelKYCVerified, TrustLevelOrganizationVerified:
		return true
	}
	return false
}

// String returns the string representation of the trust level
func (t TrustLevelEnum) String() string {
	return string(t)
}

// CredentialPortability defines how credentials can be referenced across applications
type CredentialPortability struct {
	CredentialID   string    `json:"credential_id"`
	AllowedApps    []string  `json:"allowed_apps"` // Applications that can reference this credential
	RefCount       int       `json:"ref_count"`    // Number of times referenced
	LastReferenced time.Time `json:"last_referenced"`
	Public         bool      `json:"public"` // Can any app reference this credential
}

// CardanoAddress represents a Cardano blockchain address
type CardanoAddress struct {
	Address    string `json:"address"`
	Network    string `json:"network"`
	Type       string `json:"type"` // enterprise, base, pointer, etc.
	StakingKey string `json:"staking_key,omitempty"`
}

// UTxO represents an Unspent Transaction Output on Cardano
type UTxO struct {
	TxHash    string                 `json:"tx_hash"`
	Index     int                    `json:"index"`
	Address   CardanoAddress         `json:"address"`
	Amount    int64                  `json:"amount"` // Lovelace (smallest unit)
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Datum     string                 `json:"datum,omitempty"`
	DatumHash string                 `json:"datum_hash,omitempty"`
}

// BlockchainMetadata represents metadata stored on the blockchain for an identity
type BlockchainMetadata struct {
	UserID              string               `json:"user_id"`
	IdentityAnchor      string               `json:"identity_anchor"` // IPFS hash or URL
	CredentialCount     int                  `json:"credential_count"`
	CurrentTrustLevel   string               `json:"current_trust_level"`
	LastTrustUpdate     time.Time            `json:"last_trust_update"`
	VerificationHistory []VerificationRecord `json:"verification_history"`
	LinkedDIDs          []string             `json:"linked_dids"`
}

// VerificationRecord represents a single verification event
type VerificationRecord struct {
	Type        string    `json:"type"` // device, kyc, organization
	Timestamp   time.Time `json:"timestamp"`
	Verifier    string    `json:"verifier"`
	ExpiresAt   time.Time `json:"expires_at"`
	Fingerprint string    `json:"fingerprint"` // Device or biometric fingerprint
}

// CacheStats provides statistics about cache performance
type CacheStats struct {
	TotalHits      int64         `json:"total_hits"`
	TotalMisses    int64         `json:"total_misses"`
	HitRate        float64       `json:"hit_rate"`
	CachedEntries  int           `json:"cached_entries"`
	AverageLatency time.Duration `json:"average_latency"`
}

// CredentialQuery represents a query for credentials with filters
type CredentialQuery struct {
	UserID        string            `json:"user_id,omitempty"`
	CredentialID  string            `json:"credential_id,omitempty"`
	CredType      string            `json:"cred_type,omitempty"`
	Issuer        string            `json:"issuer,omitempty"`
	MinTrustLevel string            `json:"min_trust_level,omitempty"`
	IsExpired     *bool             `json:"is_expired,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Limit         int               `json:"limit"`
	Offset        int               `json:"offset"`
}

// QueryResult represents the result of a credential query
type QueryResult struct {
	Total       int           `json:"total"`
	Count       int           `json:"count"`
	Offset      int           `json:"offset"`
	Credentials []*Credential `json:"credentials"`
	QueryTime   time.Duration `json:"query_time"`
	CacheHit    bool          `json:"cache_hit"`
}

// SyncState represents the state of synchronization with the blockchain
type SyncState struct {
	LastSyncTime         time.Time     `json:"last_sync_time"`
	NextSyncTime         time.Time     `json:"next_sync_time"`
	SyncInProgress       bool          `json:"sync_in_progress"`
	LastBlockProcessed   int64         `json:"last_block_processed"`
	PendingTransactions  int           `json:"pending_transactions"`
	FailedTransactions   int           `json:"failed_transactions"`
	SyncDuration         time.Duration `json:"sync_duration"`
	CredentialsProcessed int           `json:"credentials_processed"`
	TrustLevelsProcessed int           `json:"trust_levels_processed"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	TraceID   string    `json:"trace_id,omitempty"`
}

// Credential represents a credential stored on the blockchain
type Credential struct {
	ID             string                 `json:"id"`
	CredentialID   string                 `json:"credential_id"`
	UserID         string                 `json:"user_id"`
	SchemaID       string                 `json:"schema_id"`
	CredentialType string                 `json:"credential_type"`
	Issuer         string                 `json:"issuer"`
	IssuedAt       time.Time              `json:"issued_at"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
	ContentHash    string                 `json:"content_hash"`
	TxHash         string                 `json:"tx_hash,omitempty"`
	Status         string                 `json:"status"` // active, revoked, expired
	Data           map[string]interface{} `json:"data,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// TrustLevel represents the trust level of a user
type TrustLevel struct {
	UserID    string    `json:"user_id"`
	Level     string    `json:"level"`
	TxHash    string    `json:"tx_hash,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
	Reason    string    `json:"reason,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// AuditEntry represents an audit trail entry
type AuditEntry struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Action        string    `json:"action"`
	Details       string    `json:"details,omitempty"`
	ActorID       string    `json:"actor_id"`
	PreviousLevel string    `json:"previous_level,omitempty"`
	NewLevel      string    `json:"new_level,omitempty"`
	Reason        string    `json:"reason,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// CredentialStoreResult represents the result of storing a credential
type CredentialStoreResult struct {
	CredentialID    string    `json:"credential_id"`
	UserID          string    `json:"user_id"`
	TransactionHash string    `json:"transaction_hash"`
	Status          string    `json:"status"`
	ContentHash     string    `json:"content_hash"`
	Timestamp       time.Time `json:"timestamp"`
}

// RevocationResult represents the result of revoking a credential
type RevocationResult struct {
	CredentialID    string    `json:"credential_id"`
	RevocationID    string    `json:"revocation_id"`
	TransactionHash string    `json:"transaction_hash"`
	Status          string    `json:"status"`
	Timestamp       time.Time `json:"timestamp"`
}

// TrustLevelUpdateResult represents the result of updating a trust level
type TrustLevelUpdateResult struct {
	UserID          string     `json:"user_id"`
	NewTrustLevel   string     `json:"new_trust_level"`
	OldTrustLevel   string     `json:"old_trust_level,omitempty"`
	Status          string     `json:"status"`
	TransactionHash string     `json:"transaction_hash"`
	Timestamp       *time.Time `json:"timestamp,omitempty"`
}

// MarshalJSON provides custom JSON marshaling
func (c *Credential) MarshalJSON() ([]byte, error) {
	type Alias Credential
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(c),
		Timestamp: c.Timestamp.Format(time.RFC3339),
	})
}

// MarshalJSON provides custom JSON marshaling
func (t *TrustLevel) MarshalJSON() ([]byte, error) {
	type Alias TrustLevel
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(t),
		Timestamp: t.Timestamp.Format(time.RFC3339),
	})
}

// MarshalJSON provides custom JSON marshaling
func (a *AuditEntry) MarshalJSON() ([]byte, error) {
	type Alias AuditEntry
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     (*Alias)(a),
		Timestamp: a.Timestamp.Format(time.RFC3339),
	})
}
