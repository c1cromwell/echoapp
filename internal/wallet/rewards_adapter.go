package wallet

import (
	"context"
	"time"

	"github.com/thechadcromwell/echoapp/internal/services/rewards"
)

// rewardsService is the subset of *rewards.Service the adapter depends on,
// extracted so the claim/aggregation logic is unit-testable.
type rewardsService interface {
	GetPending(ctx context.Context, did string, trustTier int) (*rewards.PendingRewards, error)
	GetDailyStats(ctx context.Context) (*rewards.DailyStats, error)
	Claim(ctx context.Context, req rewards.ClaimRequest) (*rewards.ClaimResult, error)
}

// RewardsAdapter implements RewardsQuerier using the rewards service.
type RewardsAdapter struct {
	svc rewardsService
}

// NewRewardsAdapter wraps the v3 rewards service for wallet aggregation.
func NewRewardsAdapter(svc rewardsService) *RewardsAdapter {
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

func (a *RewardsAdapter) ClearPending(ctx context.Context, did string, types []string, trustTier int) error {
	if len(types) == 0 {
		types = []string{"messaging"}
	}
	// Tier 0 (e.g. passkey requests with no JWT claim) falls back to the base
	// tier so a missing claim never zeroes the reward multiplier.
	if trustTier < 1 {
		trustTier = 1
	}
	var firstErr error
	for _, rewardType := range types {
		if _, err := a.svc.Claim(ctx, rewards.ClaimRequest{
			DID:        did,
			RewardType: rewardType,
			TrustTier:  trustTier,
		}); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

var _ RewardsQuerier = (*RewardsAdapter)(nil)
