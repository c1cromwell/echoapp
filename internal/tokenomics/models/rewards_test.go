package models

import (
	"math/big"
	"testing"
	"time"
)

func TestRewardType(t *testing.T) {
	tests := []struct {
		name string
		typ  RewardType
	}{
		{"Text Reward", TextReward},
		{"Voice Reward", VoiceReward},
		{"Video Reward", VideoReward},
		{"Referral Reward", ReferralReward},
		{"Governance Reward", GovernanceReward},
		{"Staking Reward", StakingReward},
		{"Burn Reward", BurnReward},
		{"Bridge Reward", BridgeReward},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.typ < 0 || tt.typ > 7 {
				t.Errorf("Invalid reward type: %v", tt.typ)
			}
		})
	}
}

func TestRewardEarning(t *testing.T) {
	earning := &RewardEarning{
		UserID:     "user-123",
		RewardType: TextReward,
		Amount:     big.NewInt(1000),
		Multiplier: 1.5,
		EarnedAt:   time.Now(),
		Claimed:    false,
	}

	if earning.UserID != "user-123" {
		t.Errorf("UserID: got %v, want user-123", earning.UserID)
	}

	if earning.Claimed {
		t.Errorf("Claimed: got true, want false")
	}

	if earning.Multiplier != 1.5 {
		t.Errorf("Multiplier: got %v, want 1.5", earning.Multiplier)
	}
}

// Auto-scaling model (PRD v2.5.1): there is no hard daily cap, so even a very
// high message count never reports the limit as reached.
func TestDailyRewardTrackerNoHardCap(t *testing.T) {
	for _, count := range []int{250, 500, 10000} {
		tracker := &DailyRewardTracker{
			UserID:           "user-123",
			Date:             time.Now(),
			MessagesRewarded: count,
			EchoEarned:       big.NewInt(5000000000),
		}
		if tracker.IsLimitReached() {
			t.Errorf("IsLimitReached at %d messages: got true, want false (auto-scale, no cap)", count)
		}
	}
}

func TestTrustScoreMultiplier(t *testing.T) {
	tests := []struct {
		name       string
		score      int
		multiplier float64
	}{
		{"Unverified (0-19)", 10, 1.0},
		{"Newcomer (20-39)", 30, 1.2},
		{"Member (40-59)", 50, 1.5},
		{"Trusted (60-79)", 70, 2.0},
		{"Verified (80-100)", 90, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trustScore := &TrustScore{
				UserID: "user-123",
				Score:  tt.score,
			}

			multiplier := trustScore.GetMultiplier()
			if multiplier != tt.multiplier {
				t.Errorf("Multiplier: got %v, want %v", multiplier, tt.multiplier)
			}
		})
	}
}

func TestTrustScoreBoundaries(t *testing.T) {
	tests := []struct {
		name       string
		score      int
		multiplier float64
	}{
		{"Score 0", 0, 1.0},
		{"Score 19", 19, 1.0},
		{"Score 20", 20, 1.2},
		{"Score 100", 100, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trustScore := &TrustScore{
				UserID: "user-123",
				Score:  tt.score,
			}

			multiplier := trustScore.GetMultiplier()
			if multiplier != tt.multiplier {
				t.Errorf("Score %d: got %v, want %v", tt.score, multiplier, tt.multiplier)
			}
		})
	}
}

func TestReferralInfo(t *testing.T) {
	referral := &ReferralInfo{
		ReferrerID:     "referrer-123",
		RefereeID:      "referee-456",
		SignupBonus:    big.NewInt(500000000),
		VerifyBonus:    big.NewInt(2000000000),
		MilestoneBonus: big.NewInt(2500000000),
		CreatedAt:      time.Now(),
	}

	if referral.ReferrerID != "referrer-123" {
		t.Errorf("ReferrerID: incorrect")
	}

	if referral.RefereeID != "referee-456" {
		t.Errorf("RefereeID: incorrect")
	}
}

func BenchmarkTrustScoreMultiplier(b *testing.B) {
	trustScore := &TrustScore{
		UserID: "user-123",
		Score:  75,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = trustScore.GetMultiplier()
	}
}

func BenchmarkRewardEarning(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &RewardEarning{
			UserID:     "user-123",
			RewardType: TextReward,
			Amount:     big.NewInt(1000),
			Multiplier: 1.5,
			EarnedAt:   time.Now(),
			Claimed:    false,
		}
	}
}
