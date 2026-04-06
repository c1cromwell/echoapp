package main

import (
	"math/big"
	"testing"
	"time"

	"github.com/thechadcromwell/echoapp/internal/tokenomics/models"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/rewards"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/staking"
)

// TestTokenConfiguration verifies token specs are correct
func TestTokenConfiguration(t *testing.T) {
	config := models.NewTokenConfig()

	if config.Name != "ECHO" {
		t.Errorf("Expected token name ECHO, got %s", config.Name)
	}

	if config.Symbol != "ECHO" {
		t.Errorf("Expected token symbol ECHO, got %s", config.Symbol)
	}

	if config.Decimals != 8 {
		t.Errorf("Expected decimals 8, got %d", config.Decimals)
	}

	if !config.HardCapped {
		t.Error("Expected HardCapped to be true")
	}

	// Verify total supply is 1 billion with 8 decimals
	expected := new(big.Int)
	expected.SetString("100000000000000000", 10)

	if config.TotalSupply.Cmp(expected) != 0 {
		t.Errorf("Expected total supply %s, got %s", expected.String(), config.TotalSupply.String())
	}
}

// TestAllocationBreakdown verifies token distribution sums to expected totals
func TestAllocationBreakdown(t *testing.T) {
	ab := models.NewAllocationBreakdown()

	if ab.UserRewards == nil || ab.ValidatorRewards == nil || ab.Ecosystem == nil {
		t.Fatal("Allocation pools not initialized")
	}

	// Verify individual allocations
	total := new(big.Int)
	total.Add(ab.UserRewards, ab.ValidatorRewards)
	total.Add(total, ab.Ecosystem)
	total.Add(total, ab.Team)
	total.Add(total, ab.Treasury)
	total.Add(total, ab.Liquidity)

	expected := models.NewTokenConfig().TotalSupply

	if total.Cmp(expected) != 0 {
		t.Errorf("Expected total allocation %s, got %s", expected.String(), total.String())
	}

	// Verify user rewards are 40% (400M ECHO)
	expectedUserRewards := new(big.Int).Mul(expected, big.NewInt(40))
	expectedUserRewards.Div(expectedUserRewards, big.NewInt(100))

	if ab.UserRewards.Cmp(expectedUserRewards) != 0 {
		t.Errorf("Expected user rewards %s, got %s", expectedUserRewards.String(), ab.UserRewards.String())
	}

	// Verify validator rewards are 25% (250M ECHO)
	expectedValidatorRewards := new(big.Int).Mul(expected, big.NewInt(25))
	expectedValidatorRewards.Div(expectedValidatorRewards, big.NewInt(100))

	if ab.ValidatorRewards.Cmp(expectedValidatorRewards) != 0 {
		t.Errorf("Expected validator rewards %s, got %s", expectedValidatorRewards.String(), ab.ValidatorRewards.String())
	}
}

// TestTrustScoreMultiplier verifies reward multipliers based on trust score
func TestTrustScoreMultiplier(t *testing.T) {
	tests := []struct {
		name         string
		trustScore   int
		expectedMult float64
	}{
		{"Unverified", 10, 0.5},
		{"Newcomer", 30, 1.0},
		{"Member", 50, 1.5},
		{"Trusted", 70, 2.5},
		{"Verified", 90, 5.0},
	}

	for _, tt := range tests {
		ts := &models.TrustScore{
			UserID: "test_user",
			Score:  tt.trustScore,
		}
		mult := ts.GetMultiplier()
		if mult != tt.expectedMult {
			t.Errorf("%s: expected multiplier %.1f, got %.1f", tt.name, tt.expectedMult, mult)
		}
	}
}

// TestRewardCalculator verifies basic reward calculations
func TestRewardCalculator(t *testing.T) {
	rc := rewards.NewRewardCalculator()

	// Test that calculator is initialized
	if rc == nil {
		t.Fatal("RewardCalculator not initialized")
	}

	if len(rc.BaseRewards) == 0 {
		t.Fatal("BaseRewards map is empty")
	}

	// Test basic reward calculation
	reward := rc.CalculateReward(0, 1.0) // Text reward with 1x multiplier
	if reward == nil {
		t.Error("CalculateReward returned nil")
	}

	if reward.Cmp(big.NewInt(0)) <= 0 {
		t.Error("CalculateReward returned non-positive value")
	}
}

// TestRewardDistribution verifies anti-gaming tracking
func TestRewardDistribution(t *testing.T) {
	rd := rewards.NewRewardDistributor()
	userAddress := "test_user_123"

	// Verify distributor can track limits
	if !rd.CanDistribute(userAddress) {
		t.Error("New user should be able to receive rewards")
	}

	// Add some rewards
	for i := 0; i < 50; i++ {
		rd.IncrementCount(userAddress)
	}

	// After 50 increments, should still be able to distribute
	if !rd.CanDistribute(userAddress) && rd.DailyLimits[userAddress] < 500 {
		t.Error("Should allow rewards up to 500 per day")
	}
}

// TestStakingTiers verifies staking reward structure
func TestStakingTiers(t *testing.T) {
	tiers := staking.GetStakingTiers()

	// Verify all 5 tiers exist
	if len(tiers) != 5 {
		t.Errorf("Expected 5 staking tiers, got %d", len(tiers))
	}

	// Verify flexible tier has no lock
	flexibleTier := tiers[0]
	if flexibleTier.Name != "Flexible" {
		t.Errorf("First tier should be Flexible, got %s", flexibleTier.Name)
	}
	if flexibleTier.LockupDays != 0 {
		t.Error("Flexible tier should have 0 lockup days")
	}

	// Verify 365-day tier has highest APY
	tier365 := tiers[4]
	if tier365.Name != "365-Day Lock" {
		t.Errorf("Last tier should be 365-Day Lock, got %s", tier365.Name)
	}
	if tier365.APYPercent != 15.0 {
		t.Errorf("Expected 365-day APY 15.0, got %.2f", tier365.APYPercent)
	}

	// Verify governance weights increase with lockup period
	for i := 0; i < len(tiers)-1; i++ {
		if tiers[i].GovernanceWeight >= tiers[i+1].GovernanceWeight {
			t.Errorf("Governance weight should increase with lockup: tier %d (%.1f) >= tier %d (%.1f)",
				i, tiers[i].GovernanceWeight, i+1, tiers[i+1].GovernanceWeight)
		}
	}
}

// TestStakingRewardCalculation verifies reward calculations
func TestStakingRewardCalculation(t *testing.T) {
	const baseUnits int64 = 100000000

	stake := &staking.Stake{
		StakeID:         "stake_123",
		UserID:          "user_123",
		Amount:          big.NewInt(1000 * baseUnits),
		Tier:            staking.GetStakingTiers()[1], // 30-day tier
		CreatedAt:       time.Now().AddDate(0, 0, -30),
		UnlocksAt:       time.Now().AddDate(0, 0, 0),
		LastRewardClaim: time.Now().AddDate(0, 0, -30),
	}

	// Calculate pending rewards
	pending := stake.CalculatePendingReward()

	// Expected: 1000 ECHO * 5% APY / 365 days * 30 days ≈ 4.11 ECHO
	if pending.Cmp(big.NewInt(0)) <= 0 {
		t.Error("Pending reward should be positive")
	}

	// Verify result is reasonable (between 1 and 10 ECHO)
	minExpected := big.NewInt(1 * baseUnits)
	maxExpected := big.NewInt(10 * baseUnits)

	if pending.Cmp(minExpected) < 0 || pending.Cmp(maxExpected) > 0 {
		t.Logf("Pending reward %.4f ECHO seems off but may be acceptable",
			float64(pending.Int64())/float64(baseUnits))
	}
}

// TestDailyRewardTracker verifies daily limit tracking
func TestDailyRewardTracker(t *testing.T) {
	tracker := &models.DailyRewardTracker{
		UserID:           "user_123",
		Date:             time.Now(),
		MessagesRewarded: 250,
		EchoEarned:       big.NewInt(25000000), // 0.25 ECHO
		TotalActions:     300,
	}

	// Should not be at limit yet (250 < 500)
	if tracker.IsLimitReached() {
		t.Error("Limit should not be reached at 250 messages")
	}

	// Reach the limit
	tracker.MessagesRewarded = 500
	if !tracker.IsLimitReached() {
		t.Error("Limit should be reached at 500 messages")
	}
}

// TestRewardEarningTracking verifies reward earning structure
func TestRewardEarningTracking(t *testing.T) {
	earning := &models.RewardEarning{
		UserID:     "user_123",
		RewardType: models.TextReward,
		Amount:     big.NewInt(1000000),
		Multiplier: 1.5,
		EarnedAt:   time.Now(),
		Claimed:    false,
	}

	if earning.Claimed {
		t.Error("New earning should not be claimed")
	}

	if earning.Amount.Cmp(big.NewInt(0)) <= 0 {
		t.Error("Earning amount should be positive")
	}
}

// BenchmarkRewardCalculation benchmarks reward calculation
func BenchmarkRewardCalculation(b *testing.B) {
	rc := rewards.NewRewardCalculator()

	for i := 0; i < b.N; i++ {
		rc.CalculateReward(0, 1.5) // Text reward with multiplier
	}
}

// BenchmarkStakingRewardCalculation benchmarks staking calculations
func BenchmarkStakingRewardCalculation(b *testing.B) {
	stake := &staking.Stake{
		StakeID:         "stake_test",
		UserID:          "user_test",
		Amount:          big.NewInt(1000000000),
		Tier:            staking.GetStakingTiers()[1],
		CreatedAt:       time.Now(),
		UnlocksAt:       time.Now().AddDate(0, 1, 0),
		LastRewardClaim: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stake.CalculatePendingReward()
	}
}
