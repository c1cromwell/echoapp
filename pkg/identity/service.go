package identity

import (
	"context"
	"log"
	"time"

	"github.com/thechadcromwell/echoapp/pkg/cardano"
	"github.com/thechadcromwell/echoapp/pkg/identity/cache"
	"github.com/thechadcromwell/echoapp/pkg/identity/trust"
	"github.com/thechadcromwell/echoapp/pkg/identity/vc"
)

// Service manages identity operations
type Service struct {
	cardanoClient     *cardano.Client
	trustLevelService *trust.TrustLevelService
	vcService         *vc.VerifiableCredentialService
	identityCache     *cache.IdentityCache
	logger            *log.Logger
}

// NewService creates a new identity service
func NewService(cardanoClient *cardano.Client) *Service {
	logger := log.New(nil, "identity", 0)

	return &Service{
		cardanoClient: cardanoClient,
		trustLevelService: trust.NewTrustLevelService(cardanoClient, trust.TrustLevelServiceConfig{
			Logger: logger,
		}),
		vcService: vc.NewVerifiableCredentialService(cardanoClient, vc.VerifiableCredentialServiceConfig{
			Logger: logger,
		}),
		identityCache: cache.NewIdentityCache(300),
		logger:        logger,
	}
}

// Close closes the service and releases resources
func (s *Service) Close() error {
	s.identityCache.Clear()
	return nil
}

// GetTrustLevel retrieves the trust level for a user
func (s *Service) GetTrustLevel(ctx context.Context, userID string) (string, error) {
	trustLevel, err := s.trustLevelService.GetTrustLevel(ctx, userID)
	if err != nil {
		return "", err
	}
	if trustLevel == nil {
		return "unverified", nil
	}
	return trustLevel.Level, nil
}

// UpdateTrustLevel updates a user's trust level
func (s *Service) UpdateTrustLevel(ctx context.Context, userID, newLevel, verificationMethod, verifierID, reason string) error {
	_, err := s.trustLevelService.UpdateTrustLevel(ctx, userID, newLevel, verificationMethod, verifierID, reason)
	return err
}

// StoreCredential stores a credential for a user
func (s *Service) StoreCredential(ctx context.Context, userID, credentialID string, data map[string]interface{}) error {
	_, err := s.vcService.StoreCredential(ctx, userID, credentialID, data, "", "", time.Now().Add(24*time.Hour), nil)
	return err
}

// GetCredential retrieves a credential
func (s *Service) GetCredential(ctx context.Context, credentialID string) (interface{}, error) {
	return s.vcService.GetCredential(ctx, credentialID)
}

// GetUserCredentials retrieves all credentials for a user
func (s *Service) GetUserCredentials(ctx context.Context, userID string) ([]interface{}, error) {
	creds, err := s.vcService.GetUserCredentials(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]interface{}, len(creds))
	for i, cred := range creds {
		result[i] = cred
	}
	return result, nil
}

// GetStorageHealth checks the health of the storage system
func (s *Service) GetStorageHealth(ctx context.Context) error {
	return s.cardanoClient.Health(ctx)
}
