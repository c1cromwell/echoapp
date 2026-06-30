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
	tiers   []int
}

func (f *fakeRewardsService) GetPending(_ context.Context, _ string, _ int) (*rewards.PendingRewards, error) {
	return &rewards.PendingRewards{}, nil
}

func (f *fakeRewardsService) GetDailyStats(_ context.Context) (*rewards.DailyStats, error) {
	return &rewards.DailyStats{}, nil
}

func (f *fakeRewardsService) Claim(_ context.Context, req rewards.ClaimRequest) (*rewards.ClaimResult, error) {
	f.claimed = append(f.claimed, req.RewardType)
	f.tiers = append(f.tiers, req.TrustTier)
	return &rewards.ClaimResult{}, nil
}

func TestClearPendingClearsAllTypes(t *testing.T) {
	fake := &fakeRewardsService{}
	adapter := NewRewardsAdapter(fake)

	if err := adapter.ClearPending(context.Background(), "did:echo:test", []string{"staking", "messaging"}, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(fake.claimed) != 2 {
		t.Fatalf("expected 2 claims, got %d (%v)", len(fake.claimed), fake.claimed)
	}
	if fake.claimed[0] != "staking" || fake.claimed[1] != "messaging" {
		t.Errorf("expected both types claimed in order, got %v", fake.claimed)
	}
	for _, tier := range fake.tiers {
		if tier != 3 {
			t.Errorf("expected trust tier 3 propagated, got %v", fake.tiers)
		}
	}
}

func TestClearPendingDefaultsTierFloor(t *testing.T) {
	fake := &fakeRewardsService{}
	adapter := NewRewardsAdapter(fake)

	// Tier 0 (no JWT claim) must floor to 1, never zero the multiplier.
	if err := adapter.ClearPending(context.Background(), "did:echo:test", []string{"messaging"}, 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fake.tiers) != 1 || fake.tiers[0] != 1 {
		t.Errorf("expected tier floored to 1, got %v", fake.tiers)
	}
}
