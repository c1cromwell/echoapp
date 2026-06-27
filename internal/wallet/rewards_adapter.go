package wallet

import (
	"context"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/rewards"
)

// RewardsAdapter implements RewardsQuerier using the rewards service.
type RewardsAdapter struct {
	svc *rewards.Service
}

// NewRewardsAdapter wraps the v3 rewards service for wallet aggregation.
func NewRewardsAdapter(svc *rewards.Service) *RewardsAdapter {
	return &RewardsAdapter{svc: svc}
}

func (a *RewardsAdapter) GetPending(ctx context.Context, did string) (int64, error) {
	pending, err := a.svc.GetPending(ctx, did, 1)
	if err != nil {
		return 0, err
	}
	return pending.TotalPending, nil
}

func (a *RewardsAdapter) GetPendingByType(ctx context.Context, did, rewardType string) (int64, error) {
	pending, err := a.svc.GetPending(ctx, did, 1)
	if err != nil {
		return 0, err
	}
	if pending.Pending == nil {
		return 0, nil
	}
	return pending.Pending[rewardType], nil
}

func (a *RewardsAdapter) GetAutoScaleState(ctx context.Context, did string) (*AutoScaleState, error) {
	stats, err := a.svc.GetDailyStats(ctx)
	if err != nil {
		return nil, err
	}
	_ = did
	return &AutoScaleState{
		CurrentRate:          stats.AutoScaleRate,
		DailyBudget:          stats.DailyBudget,
		EffectiveDailyBudget: stats.EffectiveDailyBudget,
		BudgetUsedToday:      stats.TotalDistributed,
		RemainingToday:       stats.RemainingBudget,
		LastUpdated:          stats.Timestamp.UTC().Format(time.RFC3339),
	}, nil
}

func (a *RewardsAdapter) ClearPending(ctx context.Context, did string, types []string) error {
	_, err := a.svc.Claim(ctx, rewards.ClaimRequest{
		DID:        did,
		RewardType: firstOr(types, "messaging"),
		TrustTier:  1,
	})
	return err
}

func firstOr(types []string, fallback string) string {
	if len(types) > 0 {
		return types[0]
	}
	return fallback
}

var _ RewardsQuerier = (*RewardsAdapter)(nil)
