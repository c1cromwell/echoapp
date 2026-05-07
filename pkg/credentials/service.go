package credentials

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Service orchestrates credential operations
type Service struct {
	config              *Config
	issuer              *Issuer
	verifier            *Verifier
	storage             Storage
	revocationManager   *RevocationManager
	statusPublisher     *StatusListPublisher
	formatHandler       *FormatHandler
	credentialFormatter *CredentialFormatter
	cryptoUtils         *CryptoUtils
}

// NewService creates new credentials service. Optional pgPool enables durable StatusList2021
// slots and L1 outbox (WO-274) when combined with Metagraph anchoring config.
func NewService(config *Config, pgPool *pgxpool.Pool) (*Service, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Initialize components
	cryptoUtils := NewCryptoUtils()
	storage := NewInMemoryStorage()
	revocationMgr := NewRevocationManager(storage, config.RevocationConfig.CacheTTL)
	formatHandler := NewFormatHandler(cryptoUtils)
	credentialFormatter := NewCredentialFormatter(formatHandler, config)

	statusPub := NewStatusListPublisher(config.MetagraphConfig)
	if statusPub != nil && pgPool != nil {
		statusPub.AttachPostgres(pgPool)
	}
	if statusPub != nil {
		statusPub.Start()
	}

	issuer := NewIssuer(config, cryptoUtils, storage, statusPub)
	verifier := NewVerifier(config, cryptoUtils, storage, revocationMgr)

	return &Service{
		config:              config,
		issuer:              issuer,
		verifier:            verifier,
		storage:             storage,
		revocationManager:   revocationMgr,
		statusPublisher:     statusPub,
		formatHandler:       formatHandler,
		credentialFormatter: credentialFormatter,
		cryptoUtils:         cryptoUtils,
	}, nil
}

// IssueCredential issues a credential
func (s *Service) IssueCredential(ctx context.Context, req *CredentialIssuanceRequest) (*CredentialIssuanceResponse, error) {
	if req.PreferredFormat == "" {
		req.PreferredFormat = s.config.CredentialConfig.DefaultFormat
	}

	return s.issuer.IssueCredential(ctx, req)
}

// VerifyCredential verifies a credential
func (s *Service) VerifyCredential(ctx context.Context, req *CredentialVerificationRequest) (*CredentialVerificationResult, error) {
	return s.verifier.VerifyCredential(ctx, req)
}

// RevokeCredential revokes a credential. Local revocation is durable via storage; Identity L1
// StatusList2021 publish is asynchronous (coalesced + retries).
func (s *Service) RevokeCredential(ctx context.Context, credentialID, issuerDID, subjectDID, reason string) (*RevokeCredentialResult, error) {
	if err := s.revocationManager.RevokeCredential(ctx, credentialID, issuerDID, subjectDID, reason); err != nil {
		return nil, err
	}
	result := &RevokeCredentialResult{
		Status:       "revoked",
		CredentialID: credentialID,
	}
	if s.statusPublisher != nil {
		enqueued := s.statusPublisher.MarkRevoked(credentialID)
		result.StatusList = &RevokeStatusListMeta{
			PublishEnqueued:           enqueued,
			MetagraphPublishPending:   s.statusPublisher.PublishPending(),
		}
	}
	return result, nil
}

// CheckRevocationStatus checks if credential is revoked
func (s *Service) CheckRevocationStatus(ctx context.Context, credentialID string) (*RevocationStatus, error) {
	return s.revocationManager.CheckRevocationStatus(ctx, credentialID)
}

// GetCredential retrieves a credential
func (s *Service) GetCredential(ctx context.Context, credentialID string) (*VerifiableCredential, error) {
	return s.storage.RetrieveCredential(ctx, credentialID)
}

// ListCredentials lists credentials by subject
func (s *Service) ListCredentials(ctx context.Context, subjectDID string) ([]*CredentialMetadata, error) {
	return s.storage.ListCredentialsBySubject(ctx, subjectDID)
}

// GetIssuanceProgress gets issuance progress
func (s *Service) GetIssuanceProgress(credentialID string) *IssuanceProgress {
	return s.issuer.GetIssuanceProgress(credentialID)
}

// GetTrustScore gets trust score for credential
func (s *Service) GetTrustScore(credentialID string) (float64, bool) {
	return s.verifier.GetTrustScore(credentialID)
}

// ConvertCredentialFormat converts credential between formats
func (s *Service) ConvertCredentialFormat(vc *VerifiableCredential, toFormat CredentialFormat, privateKey string) (string, error) {
	return s.credentialFormatter.ConvertFormat(vc, JSONLDFormat, toFormat, privateKey)
}

// GetStorageHealth checks storage health
func (s *Service) GetStorageHealth(ctx context.Context) error {
	return s.storage.Health(ctx)
}

// GetRevocationCacheStats gets revocation cache statistics
func (s *Service) GetRevocationCacheStats() map[string]interface{} {
	return s.revocationManager.GetCacheStats()
}

// Close closes the service
func (s *Service) Close() error {
	if s.statusPublisher != nil {
		s.statusPublisher.Stop()
	}
	if err := s.revocationManager.Close(); err != nil {
		return fmt.Errorf("failed to close revocation manager: %w", err)
	}

	if err := s.storage.Close(); err != nil {
		return fmt.Errorf("failed to close storage: %w", err)
	}

	return nil
}

// GetComponentStatus gets status of all service components
func (s *Service) GetComponentStatus(ctx context.Context) map[string]interface{} {
	status := make(map[string]interface{})

	status["storage"] = map[string]interface{}{
		"healthy": s.storage.Health(ctx) == nil,
	}

	status["revocation_cache"] = s.revocationManager.GetCacheStats()

	status["supported_formats"] = s.credentialFormatter.GetSupportedFormats()

	status["issuance_timeout"] = fmt.Sprintf("%d seconds", int(s.config.CredentialConfig.IssuanceTimeout.Seconds()))
	status["verification_timeout"] = fmt.Sprintf("%d seconds", int(s.config.CredentialConfig.VerificationTimeout.Seconds()))

	if s.statusPublisher != nil {
		status["status_list_publisher"] = s.statusPublisher.Stats()
	}

	return status
}
