package wallet

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/rewards"
)

// fakeRewardsService records every Claim so we can assert the adapter clears
// each requested reward type, not just the first one.
type fakeRewardsService struct {
	claimed []string
}

func (f *fakeRewardsService) GetPending(_ context.Context, _ string, _ int) (*rewards.PendingRewards, error) {
	return &rewards.PendingRewards{}, nil
}

func (f *fakeRewardsService) GetDailyStats(_ context.Context) (*rewards.DailyStats, error) {
	return &rewards.DailyStats{}, nil
}

func (f *fakeRewardsService) Claim(_ context.Context, req rewards.ClaimRequest) (*rewards.ClaimResult, error) {
	f.claimed = append(f.claimed, req.RewardType)
	return &rewards.ClaimResult{}, nil
}

func TestClearPendingClearsAllTypes(t *testing.T) {
	fake := &fakeRewardsService{}
	adapter := NewRewardsAdapter(fake)

	if err := adapter.ClearPending(context.Background(), "did:echo:test", []string{"staking", "messaging"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.claimed) != 2 {
		t.Fatalf("expected 2 claims, got %d (%v)", len(fake.claimed), fake.claimed)
	}
	if fake.claimed[0] != "staking" || fake.claimed[1] != "messaging" {
		t.Errorf("expected both types claimed in order, got %v", fake.claimed)
	}
}
