package trust

import (
	"context"
	"log"

	"github.com/thechadcromwell/echoapp/pkg/cardano"
)

type TrustLevelServiceConfig struct {
	Logger *log.Logger
}

type TrustLevelService struct {
	cardanoClient *cardano.Client
	logger        *log.Logger
}

func NewTrustLevelService(cardanoClient *cardano.Client, config TrustLevelServiceConfig) *TrustLevelService {
	if config.Logger == nil {
		config.Logger = log.New(nil, "", 0)
	}
	return &TrustLevelService{
		cardanoClient: cardanoClient,
		logger:        config.Logger,
	}
}

func (ts *TrustLevelService) GetTrustLevel(ctx context.Context, userID string) (*cardano.TrustLevel, error) {
	return ts.cardanoClient.GetTrustLevel(ctx, userID)
}

func (ts *TrustLevelService) UpdateTrustLevel(ctx context.Context, userID, newTrustLevel, verificationMethod, verifierID, reason string) (*cardano.TrustLevelUpdateResult, error) {
	return &cardano.TrustLevelUpdateResult{
		UserID:          userID,
		NewTrustLevel:   newTrustLevel,
		Status:          "submitted",
		TransactionHash: "tx_stub",
		Timestamp:       nil,
	}, nil
}

func (ts *TrustLevelService) PromoteTrustLevel(ctx context.Context, userID, targetLevel, reason, verifierID string) (*cardano.TrustLevelUpdateResult, error) {
	return ts.UpdateTrustLevel(ctx, userID, targetLevel, "promotion", verifierID, reason)
}

func (ts *TrustLevelService) GetTrustLevelAuditTrail(ctx context.Context, userID string) ([]*cardano.AuditEntry, error) {
	return ts.cardanoClient.GetAuditTrail(ctx, userID)
}

func (ts *TrustLevelService) GetMetrics() map[string]interface{} {
	return map[string]interface{}{}
}

func (ts *TrustLevelService) MonitorPendingUpdates(ctx context.Context)          {}
func (ts *TrustLevelService) GetPendingUpdatesForUser(userID string) interface{} { return nil }
