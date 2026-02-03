package did

import (
	"time"
)

// DIDDocument represents a W3C DID document following the PRISM specification
type DIDDocument struct {
	Context            []string           `json:"@context"`
	ID                 string             `json:"id"`
	Controller         []string           `json:"controller,omitempty"`
	PublicKey          []PublicKey        `json:"publicKey"`
	Authentication     []Authentication   `json:"authentication"`
	AssertionMethod    []AssertionMethod  `json:"assertionMethod"`
	Service            []ServiceEndpoint  `json:"service,omitempty"`
	Created            time.Time          `json:"created"`
	Updated            time.Time          `json:"updated"`
	Proof              *Proof             `json:"proof,omitempty"`
}

// PublicKey represents a public key in a DID document
type PublicKey struct {
	ID              string `json:"id"`
	Type            string `json:"type"` // Ed25519VerificationKey2018
	Controller      string `json:"controller"`
	PublicKeyBase64 string `json:"publicKeyBase64"`
	PublicKeyMultibase string `json:"publicKeyMultibase,omitempty"`
}

// Authentication represents authentication verification
type Authentication struct {
	Type      string `json:"type,omitempty"`
	PublicKey string `json:"publicKey,omitempty"`
	ID        string `json:"id,omitempty"`
}

// AssertionMethod represents assertion capability
type AssertionMethod struct {
	Type      string `json:"type,omitempty"`
	PublicKey string `json:"publicKey,omitempty"`
	ID        string `json:"id,omitempty"`
}

// ServiceEndpoint represents a service endpoint in a DID document
type ServiceEndpoint struct {
	ID              string        `json:"id"`
	Type            string        `json:"type"`
	ServiceEndpoint string        `json:"serviceEndpoint"`
	RecipientKeys   []string      `json:"recipientKeys,omitempty"`
	RoutingKeys     []string      `json:"routingKeys,omitempty"`
	Accept          []string      `json:"accept,omitempty"`
	Priority        int           `json:"priority,omitempty"`
}

// Proof represents a cryptographic proof
type Proof struct {
	Type              string    `json:"type"`
	Created           time.Time `json:"created"`
	VerificationMethod string    `json:"verificationMethod"`
	ProofPurpose      string    `json:"proofPurpose"`
	ProofValue        string    `json:"proofValue"`
}

// DIDCreationRequest represents a request to create a new DID
type DIDCreationRequest struct {
	UserID          string                 `json:"userId" binding:"required"`
	DeviceID        string                 `json:"deviceId" binding:"required"`
	PublicKey       string                 `json:"publicKey" binding:"required"`
	DeviceName      string                 `json:"deviceName"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// DIDCreationResponse represents the response after DID creation
type DIDCreationResponse struct {
	DID              string         `json:"did"`
	Document         *DIDDocument   `json:"document"`
	TransactionHash  string         `json:"transactionHash,omitempty"`
	AnchoredAt       time.Time      `json:"anchoredAt"`
	ResolutionStatus string         `json:"resolutionStatus"`
}

// DIDResolutionRequest represents a request to resolve a DID
type DIDResolutionRequest struct {
	DID string `json:"did" binding:"required"`
}

// DIDResolutionResponse represents the response after DID resolution
type DIDResolutionResponse struct {
	DIDDocument *DIDDocument `json:"didDocument"`
	Metadata    ResolutionMetadata `json:"metadata"`
}

// ResolutionMetadata contains metadata about DID resolution
type ResolutionMetadata struct {
	ContentType         string    `json:"contentType,omitempty"`
	ResolutionTimestamp time.Time `json:"resolutionTimestamp"`
	CachedAt            time.Time `json:"cachedAt,omitempty"`
	CacheValid          bool      `json:"cacheValid"`
	BlockchainAnchored  bool      `json:"blockchainAnchored"`
	TransactionHash     string    `json:"transactionHash,omitempty"`
}

// DeviceRegistration represents a device in a DID document
type DeviceRegistration struct {
	DeviceID    string `json:"deviceId"`
	DeviceName  string `json:"deviceName"`
	PublicKey   string `json:"publicKey"`
	CreatedAt   time.Time `json:"createdAt"`
	LastUsedAt  time.Time `json:"lastUsedAt"`
	IsActive    bool      `json:"isActive"`
	IsSecureEnclave bool `json:"isSecureEnclave"`
}

// MultiDeviceRegistrationRequest represents a request to register a new device
type MultiDeviceRegistrationRequest struct {
	DID         string `json:"did" binding:"required"`
	DeviceID    string `json:"deviceId" binding:"required"`
	DeviceName  string `json:"deviceName"`
	PublicKey   string `json:"publicKey" binding:"required"`
	QRCodeData  string `json:"qrCodeData,omitempty"`
}

// MultiDeviceRegistrationResponse represents the response after device registration
type MultiDeviceRegistrationResponse struct {
	DID          string         `json:"did"`
	UpdatedDocument *DIDDocument `json:"updatedDocument"`
	DeviceID     string         `json:"deviceId"`
	TransactionHash string       `json:"transactionHash,omitempty"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}

// QRCodeData represents data encoded in a QR code for device registration
type QRCodeData struct {
	DID           string `json:"did"`
	DeviceID      string `json:"deviceId"`
	ChallengeName string `json:"challenge"`
	Timestamp     int64  `json:"timestamp"`
	ValidUntil    int64  `json:"validUntil"`
}

// CachedDID represents a cached DID document with expiration
type CachedDID struct {
	Document   *DIDDocument
	ExpiresAt  time.Time
	CachedAt   time.Time
	Valid      bool
}

// DIDMapping represents the mapping between a DID and a user account
type DIDMapping struct {
	DID            string
	UserID         string
	AccountID      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	IsActive       bool
	PrimaryDevice  string
	Devices        []DeviceRegistration
}

// AnchorRequest represents a request to anchor a DID to the blockchain
type AnchorRequest struct {
	DID      string
	Document *DIDDocument
	Proof    *Proof
}

// AnchorResponse represents the response after anchoring a DID
type AnchorResponse struct {
	DID             string
	TransactionHash string
	BlockNumber     int64
	Timestamp       time.Time
	Status          string // "confirmed", "pending", "failed"
}

// AtalaResponse represents a response from Atala PRISM
type AtalaResponse struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// HealthCheckResponse represents the health status of the service
type HealthCheckResponse struct {
	Status              string    `json:"status"`
	Timestamp           time.Time `json:"timestamp"`
	AtalaPRISMConnected bool      `json:"atalaConnected"`
	CardanoConnected    bool      `json:"cardanoConnected"`
	CacheStatus         string    `json:"cacheStatus"`
	DatabaseConnected   bool      `json:"databaseConnected"`
}
