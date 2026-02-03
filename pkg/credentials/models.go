package credentials

import (
	"time"
)

// CredentialType defines supported credential types
type CredentialType string

const (
	ProofOfHumanity CredentialType = "ProofOfHumanity"
	KYCLite         CredentialType = "KYCLite"
	HighAssurance   CredentialType = "HighAssurance"
	Professional    CredentialType = "Professional"
)

// CredentialFormat represents supported credential formats
type CredentialFormat string

const (
	JSONLDFormat CredentialFormat = "json-ld"
	JWTFormat    CredentialFormat = "jwt"
	SDJWTFormat  CredentialFormat = "sd-jwt"
)

// VerifiableCredential represents a W3C Verifiable Credential
type VerifiableCredential struct {
	Context           []string          `json:"@context"`
	Type              []string          `json:"type"`
	ID                string            `json:"id"`
	Issuer            string            `json:"issuer"`
	IssuanceDate      time.Time         `json:"issuanceDate"`
	ExpirationDate    *time.Time        `json:"expirationDate,omitempty"`
	CredentialSubject CredentialSubject `json:"credentialSubject"`
	CredentialStatus  *CredentialStatus `json:"credentialStatus,omitempty"`
	Proof             Proof             `json:"proof"`
}

// CredentialSubject contains claims about the credential subject
type CredentialSubject struct {
	ID                 string                 `json:"id"`
	Claims             map[string]interface{} `json:"claims"`
	VerificationClaims []VerificationClaim    `json:"verificationClaims"`
}

// VerificationClaim represents a verified claim
type VerificationClaim struct {
	Type              string    `json:"type"`
	Value             string    `json:"value"`
	VerifiedAt        time.Time `json:"verifiedAt"`
	VerificationLevel string    `json:"verificationLevel"` // "basic", "intermediate", "high"
}

// CredentialStatus represents revocation status
type CredentialStatus struct {
	ID     string `json:"id"`
	Type   string `json:"type"`   // "CardanoRevocationRegistry2024"
	Status string `json:"status"` // "active", "revoked", "suspended"
}

// Proof represents the cryptographic proof of a credential
type Proof struct {
	Type               string    `json:"type"` // "Ed25519Signature2018", "JsonWebSignature2020"
	Created            time.Time `json:"created"`
	VerificationMethod string    `json:"verificationMethod"`
	ProofPurpose       string    `json:"proofPurpose"` // "assertionMethod"
	SignatureAlgorithm string    `json:"signatureAlgorithm,omitempty"`
	ChallengeNonce     string    `json:"challengeNonce,omitempty"`
	ProofValue         string    `json:"proofValue,omitempty"` // Signature in base64 or JWS format
}

// CredentialIssuanceRequest represents a request to issue a credential
type CredentialIssuanceRequest struct {
	SubjectDID         string                 `json:"subjectDid"`
	CredentialType     CredentialType         `json:"credentialType"`
	Claims             map[string]interface{} `json:"claims"`
	VerificationClaims []VerificationClaim    `json:"verificationClaims"`
	PreferredFormat    CredentialFormat       `json:"preferredFormat"`
	ExpirationYears    int                    `json:"expirationYears,omitempty"`
}

// CredentialIssuanceResponse represents issued credential
type CredentialIssuanceResponse struct {
	CredentialID         string           `json:"credentialId"`
	VerifiableCredential string           `json:"verifiableCredential"`
	Format               CredentialFormat `json:"format"`
	IssuedAt             time.Time        `json:"issuedAt"`
	ExpiresAt            time.Time        `json:"expiresAt"`
	Status               string           `json:"status"` // "pending", "issued", "failed"
}

// CredentialVerificationRequest represents a request to verify a credential
type CredentialVerificationRequest struct {
	Credential string           `json:"credential"`
	Format     CredentialFormat `json:"format"`
	IssuerDID  string           `json:"issuerDid"`
}

// CredentialVerificationResult represents verification results
type CredentialVerificationResult struct {
	CredentialID     string              `json:"credentialId"`
	IsValid          bool                `json:"isValid"`
	VerifiedAt       time.Time           `json:"verifiedAt"`
	Issuer           string              `json:"issuer"`
	Subject          string              `json:"subject"`
	CredentialType   CredentialType      `json:"credentialType"`
	ExpirationDate   *time.Time          `json:"expirationDate,omitempty"`
	SignatureValid   bool                `json:"signatureValid"`
	NotExpired       bool                `json:"notExpired"`
	NotRevoked       bool                `json:"notRevoked"`
	RevocationStatus string              `json:"revocationStatus"`
	Errors           []VerificationError `json:"errors"`
}

// VerificationError represents a verification error
type VerificationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RevocationRecord represents a credential revocation
type RevocationRecord struct {
	CredentialID     string    `json:"credentialId"`
	IssuerDID        string    `json:"issuerDid"`
	SubjectDID       string    `json:"subjectDid"`
	RevokedAt        time.Time `json:"revokedAt"`
	RevocationReason string    `json:"revocationReason"`
	ChainTxHash      string    `json:"chainTxHash,omitempty"`
}

// RevocationStatus represents current revocation status
type RevocationStatus struct {
	CredentialID     string     `json:"credentialId"`
	IsRevoked        bool       `json:"isRevoked"`
	RevokedAt        *time.Time `json:"revokedAt,omitempty"`
	RevocationReason string     `json:"revocationReason,omitempty"`
	ChainIndex       uint64     `json:"chainIndex,omitempty"` // Position in Cardano revocation registry
}

// IssuanceProgress tracks credential issuance progress
type IssuanceProgress struct {
	CredentialID string    `json:"credentialId"`
	Status       string    `json:"status"`   // "initiated", "verified", "signed", "anchored", "issued", "failed"
	Progress     int       `json:"progress"` // 0-100
	StartedAt    time.Time `json:"startedAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	EstimatedEnd time.Time `json:"estimatedEnd,omitempty"`
	CurrentStep  string    `json:"currentStep"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
}

// CredentialMetadata contains credential metadata
type CredentialMetadata struct {
	CredentialID     string           `json:"credentialId"`
	IssuerDID        string           `json:"issuerDid"`
	SubjectDID       string           `json:"subjectDid"`
	CredentialType   CredentialType   `json:"credentialType"`
	Format           CredentialFormat `json:"format"`
	IssuedAt         time.Time        `json:"issuedAt"`
	ExpiresAt        *time.Time       `json:"expiresAt,omitempty"`
	ChainAnchorHash  string           `json:"chainAnchorHash,omitempty"`
	RevocationStatus string           `json:"revocationStatus"`
	TrustScore       float64          `json:"trustScore"` // 0-100
}

// TrustScoreInput represents input for trust score calculation
type TrustScoreInput struct {
	CredentialType    CredentialType
	Age               time.Duration
	VerificationLevel string  // "basic", "intermediate", "high"
	IssuerReputation  float64 // 0-100
	IsRevoked         bool
	SignatureValid    bool
	ExpirationValid   bool
}

// CredentialClaimSet represents a set of claims in a credential
type CredentialClaimSet struct {
	CredentialID string
	Claims       []Claim
}

// Claim represents a single claim in credential
type Claim struct {
	Property string      `json:"property"`
	Value    interface{} `json:"value"`
	Verified bool        `json:"verified"`
}
