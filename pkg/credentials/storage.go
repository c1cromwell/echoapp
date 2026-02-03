package credentials

import (
	"context"
)

// Storage defines credential storage interface
type Storage interface {
	// Credential operations
	StoreCredential(ctx context.Context, credentialID string, vc *VerifiableCredential, format CredentialFormat, encoded string) error
	RetrieveCredential(ctx context.Context, credentialID string) (*VerifiableCredential, error)
	DeleteCredential(ctx context.Context, credentialID string) error
	ListCredentialsBySubject(ctx context.Context, subjectDID string) ([]*CredentialMetadata, error)
	ListCredentialsByIssuer(ctx context.Context, issuerDID string) ([]*CredentialMetadata, error)

	// Metadata operations
	StoreMetadata(ctx context.Context, metadata *CredentialMetadata) error
	RetrieveMetadata(ctx context.Context, credentialID string) (*CredentialMetadata, error)

	// Blockchain operations
	AnchorCredential(ctx context.Context, credentialID string, vc *VerifiableCredential) (string, error)
	GetAnchorStatus(ctx context.Context, credentialID string) (string, error)

	// Revocation operations
	RecordRevocation(ctx context.Context, record *RevocationRecord) error
	GetRevocationRecord(ctx context.Context, credentialID string) (*RevocationRecord, error)

	// Search operations
	SearchCredentials(ctx context.Context, query string) ([]*CredentialMetadata, error)

	// Health
	Health(ctx context.Context) error
	Close() error
}

// InMemoryStorage implements in-memory credential storage
type InMemoryStorage struct {
	credentials map[string]*VerifiableCredential
	metadata    map[string]*CredentialMetadata
	revocations map[string]*RevocationRecord
	anchors     map[string]string // credentialID -> txHash
}

// NewInMemoryStorage creates new in-memory storage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		credentials: make(map[string]*VerifiableCredential),
		metadata:    make(map[string]*CredentialMetadata),
		revocations: make(map[string]*RevocationRecord),
		anchors:     make(map[string]string),
	}
}

// StoreCredential stores a credential
func (s *InMemoryStorage) StoreCredential(ctx context.Context, credentialID string, vc *VerifiableCredential, format CredentialFormat, encoded string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.credentials[credentialID] = vc

	// Store metadata
	metadata := &CredentialMetadata{
		CredentialID:     credentialID,
		IssuerDID:        vc.Issuer,
		SubjectDID:       vc.CredentialSubject.ID,
		CredentialType:   CredentialType(vc.Type[1]),
		Format:           format,
		IssuedAt:         vc.IssuanceDate,
		ExpiresAt:        vc.ExpirationDate,
		RevocationStatus: "active",
		TrustScore:       85.0,
	}

	if vc.CredentialStatus != nil {
		metadata.RevocationStatus = vc.CredentialStatus.Status
	}

	s.metadata[credentialID] = metadata

	return nil
}

// RetrieveCredential retrieves a credential
func (s *InMemoryStorage) RetrieveCredential(ctx context.Context, credentialID string) (*VerifiableCredential, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	vc, exists := s.credentials[credentialID]
	if !exists {
		return nil, NewCredentialError(ErrCodeCredentialNotFound, "credential not found")
	}

	return vc, nil
}

// DeleteCredential deletes a credential
func (s *InMemoryStorage) DeleteCredential(ctx context.Context, credentialID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	delete(s.credentials, credentialID)
	delete(s.metadata, credentialID)

	return nil
}

// ListCredentialsBySubject lists credentials by subject DID
func (s *InMemoryStorage) ListCredentialsBySubject(ctx context.Context, subjectDID string) ([]*CredentialMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var results []*CredentialMetadata
	for _, metadata := range s.metadata {
		if metadata.SubjectDID == subjectDID {
			results = append(results, metadata)
		}
	}

	return results, nil
}

// ListCredentialsByIssuer lists credentials by issuer DID
func (s *InMemoryStorage) ListCredentialsByIssuer(ctx context.Context, issuerDID string) ([]*CredentialMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	var results []*CredentialMetadata
	for _, metadata := range s.metadata {
		if metadata.IssuerDID == issuerDID {
			results = append(results, metadata)
		}
	}

	return results, nil
}

// StoreMetadata stores credential metadata
func (s *InMemoryStorage) StoreMetadata(ctx context.Context, metadata *CredentialMetadata) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.metadata[metadata.CredentialID] = metadata
	return nil
}

// RetrieveMetadata retrieves credential metadata
func (s *InMemoryStorage) RetrieveMetadata(ctx context.Context, credentialID string) (*CredentialMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	metadata, exists := s.metadata[credentialID]
	if !exists {
		return nil, NewCredentialError(ErrCodeCredentialNotFound, "credential metadata not found")
	}

	return metadata, nil
}

// AnchorCredential anchors credential to blockchain
func (s *InMemoryStorage) AnchorCredential(ctx context.Context, credentialID string, vc *VerifiableCredential) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Simulate blockchain anchoring
	txHash := "tx_" + credentialID
	s.anchors[credentialID] = txHash

	// Update metadata with anchor hash
	if metadata, exists := s.metadata[credentialID]; exists {
		metadata.ChainAnchorHash = txHash
	}

	return txHash, nil
}

// GetAnchorStatus gets anchor status
func (s *InMemoryStorage) GetAnchorStatus(ctx context.Context, credentialID string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if _, exists := s.anchors[credentialID]; exists {
		return "confirmed", nil
	}

	return "not_anchored", nil
}

// RecordRevocation records credential revocation
func (s *InMemoryStorage) RecordRevocation(ctx context.Context, record *RevocationRecord) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.revocations[record.CredentialID] = record

	// Update metadata
	if metadata, exists := s.metadata[record.CredentialID]; exists {
		metadata.RevocationStatus = "revoked"
	}

	// Update credential status
	if vc, exists := s.credentials[record.CredentialID]; exists {
		if vc.CredentialStatus != nil {
			vc.CredentialStatus.Status = "revoked"
		}
	}

	return nil
}

// GetRevocationRecord gets revocation record
func (s *InMemoryStorage) GetRevocationRecord(ctx context.Context, credentialID string) (*RevocationRecord, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	record, exists := s.revocations[credentialID]
	if !exists {
		return nil, nil // No revocation record
	}

	return record, nil
}

// SearchCredentials searches for credentials
func (s *InMemoryStorage) SearchCredentials(ctx context.Context, query string) ([]*CredentialMetadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// Simple substring search
	var results []*CredentialMetadata
	for _, metadata := range s.metadata {
		// Search in CredentialID, IssuerDID, SubjectDID
		if contains(metadata.CredentialID, query) ||
			contains(metadata.IssuerDID, query) ||
			contains(metadata.SubjectDID, query) {
			results = append(results, metadata)
		}
	}

	return results, nil
}

// Health checks storage health
func (s *InMemoryStorage) Health(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// Close closes storage
func (s *InMemoryStorage) Close() error {
	return nil
}

// Helper function
func contains(str, substr string) bool {
	return len(str) >= len(substr) && str[len(str)-len(substr):] == substr
}
