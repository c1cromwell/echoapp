// Package vip implements WO-206 AllowSpend VIP subscription payment rails.
package vip

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/infra"
	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

const (
	PurposeVIPSubscription = "vip_subscription"
	// VIPMonthlyECHO is $9.99/month equivalent in datum (9.99 ECHO).
	VIPMonthlyECHO int64 = 999_00000000
)

var (
	ErrAllowSpendNotFound = errors.New("allow spend authorization not found")
	ErrAllowSpendExpired  = errors.New("allow spend authorization expired")
	ErrSpendExceedsAllow  = errors.New("spend amount exceeds allowance")
)

// AllowSpendRecord mirrors an on-chain AllowSpend authorization.
type AllowSpendRecord struct {
	OwnerDID   string    `json:"ownerDid"`
	SpenderDID string    `json:"spenderDid"`
	MaxAmount  int64     `json:"maxAmount"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Purpose    string    `json:"purpose"`
	TxHash     string    `json:"txHash,omitempty"`
	Active     bool      `json:"active"`
}

// SubscriptionService manages VIP AllowSpend and SpendTransaction renewals.
type SubscriptionService struct {
	mu          sync.RWMutex
	backendDID  string
	records     map[string]*AllowSpendRecord
	submitAllow func(ctx context.Context, tx metagraph.CurrencyL1Transaction) (string, error)
	submitSpend func(ctx context.Context, ownerDID string, amount int64) (string, error)
	redis       *infra.RedisClient
}

// NewSubscriptionService creates a VIP subscription service.
func NewSubscriptionService(backendDID string) *SubscriptionService {
	return &SubscriptionService{
		backendDID: backendDID,
		records:    make(map[string]*AllowSpendRecord),
	}
}

// WithCurrencyL1 wires AllowSpend submission to Currency L1.
func (s *SubscriptionService) WithCurrencyL1(client *metagraph.MetagraphClient) *SubscriptionService {
	if client == nil {
		return s
	}
	s.submitAllow = func(ctx context.Context, tx metagraph.CurrencyL1Transaction) (string, error) {
		return client.SubmitCurrencyL1(ctx, tx)
	}
	return s
}

// WithRedis enables VIP tier rate-limit cache updates.
func (s *SubscriptionService) WithRedis(r *infra.RedisClient) *SubscriptionService {
	s.redis = r
	return s
}

// AuthorizeVIP creates a 30-day AllowSpend for VIP subscription billing.
func (s *SubscriptionService) AuthorizeVIP(ctx context.Context, ownerDID string) (*AllowSpendRecord, error) {
	expires := time.Now().UTC().Add(30 * 24 * time.Hour)
	rec := &AllowSpendRecord{
		OwnerDID:   ownerDID,
		SpenderDID: s.backendDID,
		MaxAmount:  VIPMonthlyECHO,
		ExpiresAt:  expires,
		Purpose:    PurposeVIPSubscription,
		Active:     true,
	}

	if s.submitAllow != nil {
		as, err := metagraph.NewAllowSpend(
			"vip-"+ownerDID,
			ownerDID,
			s.backendDID,
			PurposeVIPSubscription,
			big.NewInt(VIPMonthlyECHO),
			expires,
		)
		if err != nil {
			return nil, err
		}
		txHash, err := s.submitAllow(ctx, metagraph.CurrencyL1Transaction{
			Type:       "allowSpend",
			AllowSpend: as,
		})
		if err != nil {
			return nil, err
		}
		rec.TxHash = txHash
	}

	s.mu.Lock()
	s.records[ownerDID] = rec
	s.mu.Unlock()

	s.activateVIPTier(ctx, ownerDID)
	return rec, nil
}

// RenewSpend executes monthly SpendTransaction against an active AllowSpend.
func (s *SubscriptionService) RenewSpend(ctx context.Context, ownerDID string) (string, error) {
	rec, err := s.GetAllowSpend(ownerDID)
	if err != nil {
		return "", err
	}
	if time.Now().After(rec.ExpiresAt) {
		return "", ErrAllowSpendExpired
	}
	if s.submitSpend != nil {
		return s.submitSpend(ctx, ownerDID, VIPMonthlyECHO)
	}
	s.activateVIPTier(ctx, ownerDID)
	return "spend_local", nil
}

// GetAllowSpend returns the active authorization for ownerDID.
func (s *SubscriptionService) GetAllowSpend(ownerDID string) (*AllowSpendRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.records[ownerDID]
	if !ok || !rec.Active {
		return nil, ErrAllowSpendNotFound
	}
	cp := *rec
	return &cp, nil
}

func (s *SubscriptionService) activateVIPTier(ctx context.Context, did string) {
	if s.redis == nil {
		return
	}
	payload, _ := json.Marshal(map[string]interface{}{
		"did":        did,
		"tier":       string(infra.TierVIP),
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	})
	_ = s.redis.CacheSet(ctx, "rate_limit:tier:"+did, payload, 90*24*time.Hour)
}

// SubscriptionTier returns infra.TierVIP when an active AllowSpend exists.
func (s *SubscriptionService) SubscriptionTier(ownerDID string) infra.SubscriptionTier {
	rec, err := s.GetAllowSpend(ownerDID)
	if err != nil || time.Now().After(rec.ExpiresAt) {
		return infra.TierBase
	}
	return infra.TierVIP
}
