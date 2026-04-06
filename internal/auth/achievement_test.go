package auth

import (
	"context"
	"testing"
	"time"
)

func TestCheckBadgeUnlock(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	tests := []struct {
		name         string
		badge        BadgeType
		score        *TrustScoreSnapshot
		shouldUnlock bool
	}{
		{
			name:  "email verified",
			badge: BadgeEmailVerified,
			score: &TrustScoreSnapshot{
				Score: 25,
				Verification: VerificationPoints{
					EmailVerifiedBool: true,
				},
			},
			shouldUnlock: true,
		},
		{
			name:  "email not verified",
			badge: BadgeEmailVerified,
			score: &TrustScoreSnapshot{
				Score: 25,
				Verification: VerificationPoints{
					EmailVerifiedBool: false,
				},
			},
			shouldUnlock: false,
		},
		{
			name:  "phone verified",
			badge: BadgePhoneVerified,
			score: &TrustScoreSnapshot{
				Score: 25,
				Verification: VerificationPoints{
					PhoneVerifiedBool: true,
				},
			},
			shouldUnlock: true,
		},
		{
			name:  "kyc lite verified",
			badge: BadgeKYCVerified,
			score: &TrustScoreSnapshot{
				Score: 50,
				Verification: VerificationPoints{
					KYCLiteVerified: 1,
				},
			},
			shouldUnlock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canUnlock, reason, err := service.CheckBadgeUnlock(ctx, tt.badge, tt.score, nil)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if canUnlock != tt.shouldUnlock {
				t.Errorf("canUnlock = %v, want %v. Reason: %s", canUnlock, tt.shouldUnlock, reason)
			}
		})
	}
}

func TestCheckBadgeUnlockTrustBased(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	tests := []struct {
		name         string
		badge        BadgeType
		score        int
		shouldUnlock bool
	}{
		{
			name:         "trusted user at 41",
			badge:        BadgeTrustedUser,
			score:        41,
			shouldUnlock: true,
		},
		{
			name:         "trusted user at 40",
			badge:        BadgeTrustedUser,
			score:        40,
			shouldUnlock: false,
		},
		{
			name:         "verified user at 61",
			badge:        BadgeVerifiedUser,
			score:        61,
			shouldUnlock: true,
		},
		{
			name:         "verified user at 60",
			badge:        BadgeVerifiedUser,
			score:        60,
			shouldUnlock: false,
		},
		{
			name:         "elite user at 81",
			badge:        BadgeEliteUser,
			score:        81,
			shouldUnlock: true,
		},
		{
			name:         "elite user at 80",
			badge:        BadgeEliteUser,
			score:        80,
			shouldUnlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := &TrustScoreSnapshot{Score: tt.score}
			canUnlock, _, _ := service.CheckBadgeUnlock(ctx, tt.badge, score, nil)

			if canUnlock != tt.shouldUnlock {
				t.Errorf("canUnlock = %v, want %v", canUnlock, tt.shouldUnlock)
			}
		})
	}
}

func TestCheckBadgeUnlockActivityBased(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		badge         BadgeType
		contactCount  int
		groupsCreated int
		streak        int
		shouldUnlock  bool
	}{
		{
			name:         "social butterfly at 50",
			badge:        BadgeSocialButterfly,
			contactCount: 50,
			shouldUnlock: true,
		},
		{
			name:         "social butterfly at 49",
			badge:        BadgeSocialButterfly,
			contactCount: 49,
			shouldUnlock: false,
		},
		{
			name:          "community builder at 5",
			badge:         BadgeCommunityBuilder,
			groupsCreated: 5,
			shouldUnlock:  true,
		},
		{
			name:          "community builder at 4",
			badge:         BadgeCommunityBuilder,
			groupsCreated: 4,
			shouldUnlock:  false,
		},
		{
			name:         "daily streak fire at 30",
			badge:        BadgeDailyStreakFire,
			streak:       30,
			shouldUnlock: true,
		},
		{
			name:         "daily streak fire at 29",
			badge:        BadgeDailyStreakFire,
			streak:       29,
			shouldUnlock: false,
		},
		{
			name:         "streak master at 50",
			badge:        BadgeStreakMaster,
			streak:       50,
			shouldUnlock: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := &TrustScoreSnapshot{
				Behavior: BehaviorPoints{
					ContactCount:     tt.contactCount,
					GroupsCreated:    tt.groupsCreated,
					DailyLoginStreak: tt.streak,
				},
			}
			canUnlock, _, _ := service.CheckBadgeUnlock(ctx, tt.badge, score, nil)

			if canUnlock != tt.shouldUnlock {
				t.Errorf("canUnlock = %v, want %v", canUnlock, tt.shouldUnlock)
			}
		})
	}
}

func TestCheckBadgeUnlockOnChainBased(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		badge         BadgeType
		txCount       int64
		stakedAmount  float64
		voteCount     int64
		referralCount int64
		shouldUnlock  bool
	}{
		{
			name:         "blockchain born",
			badge:        BadgeBlockchainBorn,
			txCount:      1,
			shouldUnlock: true,
		},
		{
			name:         "blockchain born zero",
			badge:        BadgeBlockchainBorn,
			txCount:      0,
			shouldUnlock: false,
		},
		{
			name:         "stakeholder at 1000",
			badge:        BadgeStakeholder,
			stakedAmount: 1000,
			shouldUnlock: true,
		},
		{
			name:         "stakeholder at 999",
			badge:        BadgeStakeholder,
			stakedAmount: 999,
			shouldUnlock: false,
		},
		{
			name:         "governance voter at 3",
			badge:        BadgeGovernanceVoter,
			voteCount:    3,
			shouldUnlock: true,
		},
		{
			name:         "governance voter at 2",
			badge:        BadgeGovernanceVoter,
			voteCount:    2,
			shouldUnlock: false,
		},
		{
			name:          "referral master at 5",
			badge:         BadgeReferralMaster,
			referralCount: 5,
			shouldUnlock:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := &TrustScoreSnapshot{
				OnChain: OnChainPoints{
					TransactionCount:    tt.txCount,
					StakedAmount:        tt.stakedAmount,
					GovernanceVoteCount: tt.voteCount,
					ReferralCount:       tt.referralCount,
				},
			}
			canUnlock, _, _ := service.CheckBadgeUnlock(ctx, tt.badge, score, nil)

			if canUnlock != tt.shouldUnlock {
				t.Errorf("canUnlock = %v, want %v", canUnlock, tt.shouldUnlock)
			}
		})
	}
}

func TestEarnBadge(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	badge, reward, err := service.EarnBadge(ctx, "did:echo:alice", BadgePhoneVerified, LevelBronze)

	if err != nil {
		t.Errorf("EarnBadge failed: %v", err)
	}

	if badge == nil {
		t.Fatal("badge is nil")
	}

	if badge.BadgeType != BadgePhoneVerified {
		t.Errorf("badge type = %s, want %s", badge.BadgeType, BadgePhoneVerified)
	}

	if badge.UserDID != "did:echo:alice" {
		t.Errorf("user DID = %s, want did:echo:alice", badge.UserDID)
	}

	if !badge.IsActive {
		t.Error("badge should be active")
	}

	if reward == nil {
		t.Fatal("reward is nil")
	}

	if reward.Amount != 10 {
		t.Errorf("reward = %d, want 10", reward.Amount)
	}
}

func TestEarnBadgeWithBonus(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	tests := []struct {
		level          AchievementLevel
		expectedReward int64
	}{
		{LevelBronze, 100},
		{LevelSilver, 150},
		{LevelGold, 200},
		{LevelPlatinum, 250},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			badge, reward, _ := service.EarnBadge(ctx, "did:echo:alice", BadgeKYCVerified, tt.level)

			if badge == nil || reward == nil {
				t.Fatal("badge or reward is nil")
			}

			if reward.Amount != tt.expectedReward {
				t.Errorf("reward = %d, want %d", reward.Amount, tt.expectedReward)
			}
		})
	}
}

func TestRevokeBadge(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	badge, _, _ := service.EarnBadge(ctx, "did:echo:alice", BadgePhoneVerified, LevelBronze)

	err := service.RevokeBadge(ctx, badge.BadgeID)
	if err != nil {
		t.Errorf("RevokeBadge failed: %v", err)
	}

	err = service.RevokeBadge(ctx, "")
	if err == nil {
		t.Error("expected error for empty badge ID")
	}
}

func TestGetBadgeStats(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	expiryDays := 30
	expiryTime := time.Now().AddDate(0, 0, expiryDays)

	badges := []*AchievementBadge{
		{
			BadgeID:   "badge_1",
			BadgeType: BadgePhoneVerified,
			EarnedAt:  time.Now(),
			IsActive:  true,
		},
		{
			BadgeID:   "badge_2",
			BadgeType: BadgeEmailVerified,
			EarnedAt:  time.Now(),
			IsActive:  true,
		},
		{
			BadgeID:   "badge_3",
			BadgeType: BadgeKYCVerified,
			EarnedAt:  time.Now(),
			IsActive:  true,
		},
		{
			BadgeID:   "badge_4",
			BadgeType: BadgeTrustedUser,
			EarnedAt:  time.Now(),
			ExpiresAt: &expiryTime,
			IsActive:  true,
		},
		{
			BadgeID:   "badge_5",
			BadgeType: BadgePhoneVerified,
			EarnedAt:  time.Now(),
			IsActive:  false,
		},
	}

	stats := service.GetBadgeStats(ctx, badges)

	if stats["total_badges"] != 5 {
		t.Errorf("total = %v, want 5", stats["total_badges"])
	}

	if stats["active_badges"] != 4 {
		t.Errorf("active = %v, want 4", stats["active_badges"])
	}
}

func TestCalculateAchievementLevel(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	tests := []struct {
		name          string
		badgeCount    int
		totalECHO     int64
		expectedLevel AchievementLevel
	}{
		{
			name:          "bronze",
			badgeCount:    2,
			totalECHO:     100,
			expectedLevel: LevelBronze,
		},
		{
			name:          "silver",
			badgeCount:    5,
			totalECHO:     2500,
			expectedLevel: LevelSilver,
		},
		{
			name:          "gold",
			badgeCount:    10,
			totalECHO:     5500,
			expectedLevel: LevelGold,
		},
		{
			name:          "platinum",
			badgeCount:    15,
			totalECHO:     10500,
			expectedLevel: LevelPlatinum,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badges := make([]*AchievementBadge, tt.badgeCount)
			for i := 0; i < tt.badgeCount; i++ {
				badges[i] = &AchievementBadge{
					IsActive:  true,
					ExpiresAt: nil,
				}
			}

			level := service.CalculateAchievementLevel(ctx, badges, tt.totalECHO)

			if level != tt.expectedLevel {
				t.Errorf("level = %s, want %s", level, tt.expectedLevel)
			}
		})
	}
}

func TestGetBadgesForCategory(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	verificationBadges := service.GetBadgesForCategory(ctx, "verification")
	if len(verificationBadges) == 0 {
		t.Error("no verification badges found")
	}

	trustBadges := service.GetBadgesForCategory(ctx, "trust")
	if len(trustBadges) == 0 {
		t.Error("no trust badges found")
	}

	activityBadges := service.GetBadgesForCategory(ctx, "activity")
	if len(activityBadges) == 0 {
		t.Error("no activity badges found")
	}
}

func TestEstimateNextBadges(t *testing.T) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	score := &TrustScoreSnapshot{
		Score: 50,
		Verification: VerificationPoints{
			EmailVerifiedBool: true,
			PhoneVerifiedBool: true,
			KYCFullVerified:   1,
		},
		Behavior: BehaviorPoints{
			ContactCount:     60,
			GroupsCreated:    6,
			DailyLoginStreak: 40,
		},
		OnChain: OnChainPoints{
			TransactionCount:    10,
			StakedAmount:        2000,
			GovernanceVoteCount: 5,
			ReferralCount:       8,
		},
	}

	nextBadges := service.EstimateNextBadges(ctx, score, nil, 5)

	if len(nextBadges) > 5 {
		t.Errorf("estimated badges = %d, exceeds limit of 5", len(nextBadges))
	}
}

func TestBadgeIDGeneration(t *testing.T) {
	badge1, _, _ := NewAchievementService(nil).EarnBadge(context.Background(), "did:echo:alice", BadgePhoneVerified, LevelBronze)
	badge2, _, _ := NewAchievementService(nil).EarnBadge(context.Background(), "did:echo:alice", BadgePhoneVerified, LevelBronze)

	if badge1.BadgeID == badge2.BadgeID {
		t.Error("badge IDs should be unique")
	}

	if badge1.BadgeID == "" {
		t.Error("badge ID should not be empty")
	}
}

// Benchmark badge earning
func BenchmarkEarnBadge(b *testing.B) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.EarnBadge(ctx, "did:echo:alice", BadgePhoneVerified, LevelBronze)
	}
}

// Benchmark check unlock
func BenchmarkCheckBadgeUnlock(b *testing.B) {
	service := NewAchievementService(nil)
	ctx := context.Background()

	score := &TrustScoreSnapshot{
		Score: 50,
		Verification: VerificationPoints{
			EmailVerifiedBool: true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CheckBadgeUnlock(ctx, BadgeEmailVerified, score, nil)
	}
}
