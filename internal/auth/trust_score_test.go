package auth

import (
	"context"
	"testing"
)

func TestGetTrustLevel(t *testing.T) {
	tests := []struct {
		score    int
		expected TrustScoreLevel
	}{
		{0, TrustLevelNewcomer},
		{20, TrustLevelNewcomer},
		{21, TrustLevelBasic},
		{40, TrustLevelBasic},
		{41, TrustLevelTrusted},
		{60, TrustLevelTrusted},
		{61, TrustLevelVerified},
		{80, TrustLevelVerified},
		{81, TrustLevelElite},
		{100, TrustLevelElite},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			result := getTrustLevel(tt.score)
			if result != tt.expected {
				t.Errorf("getTrustLevel(%d) = %v, want %v", tt.score, result, tt.expected)
			}
		})
	}
}

func TestCalculateVerificationScore(t *testing.T) {
	config := getDefaultConfig()

	tests := []struct {
		name     string
		vp       *VerificationPoints
		expected int
	}{
		{
			name:     "no verification",
			vp:       &VerificationPoints{},
			expected: 0,
		},
		{
			name: "passkey only",
			vp: &VerificationPoints{
				PasskeyCreated: 1,
			},
			expected: 5,
		},
		{
			name: "multiple verifications",
			vp: &VerificationPoints{
				PasskeyCreated:  1,
				PhoneVerified:   1,
				EmailVerified:   1,
				KYCLiteVerified: 1,
			},
			expected: 25,
		},
		{
			name: "full verification",
			vp: &VerificationPoints{
				PasskeyCreated:  1,
				PhoneVerified:   1,
				EmailVerified:   1,
				KYCLiteVerified: 1,
				KYCFullVerified: 1,
				AppleDigitalID:  1,
				OrgVerified:     1,
			},
			expected: 70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateVerificationScore(tt.vp, config)
			if result != tt.expected {
				t.Errorf("calculateVerificationScore() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCalculateBehaviorScore(t *testing.T) {
	config := getDefaultConfig()

	tests := []struct {
		name     string
		bp       *BehaviorPoints
		expected int
	}{
		{
			name:     "new user",
			bp:       &BehaviorPoints{},
			expected: 0,
		},
		{
			name: "1000 messages",
			bp: &BehaviorPoints{
				MessageCount: 1000,
			},
			expected: 10,
		},
		{
			name: "6 months old",
			bp: &BehaviorPoints{
				AccountAgeMonths: 6,
			},
			expected: 6,
		},
		{
			name: "daily login streak",
			bp: &BehaviorPoints{
				DailyLoginStreak: 5,
			},
			expected: 5,
		},
		{
			name: "full behavior profile",
			bp: &BehaviorPoints{
				AccountAgeMonths: 12,
				MessageCount:     2000,
				VerifiedContacts: 100,
				ActiveGroups:     20,
				DailyLoginStreak: 10,
			},
			expected: 76,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateBehaviorScore(tt.bp, config)
			if result != tt.expected {
				t.Errorf("calculateBehaviorScore() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCalculateOnChainScore(t *testing.T) {
	config := getDefaultConfig()

	tests := []struct {
		name     string
		ocp      *OnChainPoints
		expected int
	}{
		{
			name:     "no activity",
			ocp:      &OnChainPoints{},
			expected: 0,
		},
		{
			name: "10 transactions",
			ocp: &OnChainPoints{
				TransactionCount: 10,
			},
			expected: 10,
		},
		{
			name: "5000 ECHO staked",
			ocp: &OnChainPoints{
				StakedAmount: 5000,
			},
			expected: 50,
		},
		{
			name: "5 governance votes",
			ocp: &OnChainPoints{
				GovernanceVotes: 5,
			},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateOnChainScore(tt.ocp, config)
			if result != tt.expected {
				t.Errorf("calculateOnChainScore() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCalculatePenaltyScore(t *testing.T) {
	config := getDefaultConfig()

	tests := []struct {
		name     string
		pp       *PenaltyPoints
		expected int
	}{
		{
			name:     "no penalties",
			pp:       &PenaltyPoints{},
			expected: 0,
		},
		{
			name: "one spam report",
			pp: &PenaltyPoints{
				SpamReports: 1,
			},
			expected: -2,
		},
		{
			name: "multiple penalties",
			pp: &PenaltyPoints{
				SpamReports:    2,
				FraudReports:   1,
				InactiveMonths: 3,
			},
			expected: -12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePenaltyScore(tt.pp, config)
			if result != tt.expected {
				t.Errorf("calculatePenaltyScore() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCalculateTrustScore(t *testing.T) {
	config := getDefaultConfig()
	service := NewTrustScoreService(config)
	ctx := context.Background()

	tests := []struct {
		name               string
		verification       *VerificationPoints
		behavior           *BehaviorPoints
		onChain            *OnChainPoints
		penalties          *PenaltyPoints
		expectedLevel      TrustScoreLevel
		expectedMultiplier float64
	}{
		{
			name: "new user",
			verification: &VerificationPoints{
				PasskeyCreated: 1,
			},
			behavior:           &BehaviorPoints{},
			onChain:            &OnChainPoints{},
			penalties:          &PenaltyPoints{},
			expectedLevel:      TrustLevelNewcomer,
			expectedMultiplier: 1.0,
		},
		{
			name: "basic user",
			verification: &VerificationPoints{
				PasskeyCreated: 1,
				PhoneVerified:  1,
				EmailVerified:  1,
			},
			behavior: &BehaviorPoints{
				AccountAgeMonths: 2,
				MessageCount:     500,
			},
			onChain:            &OnChainPoints{},
			penalties:          &PenaltyPoints{},
			expectedLevel:      TrustLevelBasic,
			expectedMultiplier: 1.1,
		},
		{
			name: "trusted user",
			verification: &VerificationPoints{
				PasskeyCreated:  1,
				PhoneVerified:   1,
				EmailVerified:   1,
				KYCLiteVerified: 1,
			},
			behavior: &BehaviorPoints{
				AccountAgeMonths: 3,
				MessageCount:     500,
				VerifiedContacts: 20,
				ActiveGroups:     5,
				DailyLoginStreak: 2,
			},
			onChain: &OnChainPoints{
				TransactionCount: 5,
				StakedAmount:     500,
			},
			penalties:          &PenaltyPoints{},
			expectedLevel:      TrustLevelTrusted,
			expectedMultiplier: 1.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.CalculateTrustScore(ctx, tt.verification, tt.behavior, tt.onChain, tt.penalties)

			if err != nil {
				t.Fatalf("CalculateTrustScore() error = %v", err)
			}

			if result.Level != tt.expectedLevel {
				t.Errorf("level = %v, want %v", result.Level, tt.expectedLevel)
			}

			if result.EarningMultiplier != tt.expectedMultiplier {
				t.Errorf("multiplier = %v, want %v", result.EarningMultiplier, tt.expectedMultiplier)
			}

			if result.CurrentScore < 0 || result.CurrentScore > 100 {
				t.Errorf("score = %d, expected 0-100", result.CurrentScore)
			}
		})
	}
}

func TestRecordVerification(t *testing.T) {
	config := getDefaultConfig()
	service := NewTrustScoreService(config)
	ctx := context.Background()

	tests := []struct {
		name               string
		verificationType   string
		currentScore       int
		expectedMultiplier float64
	}{
		{
			name:               "passkey by newcomer",
			verificationType:   "passkey",
			currentScore:       10,
			expectedMultiplier: 1.0,
		},
		{
			name:               "kyc by trusted user",
			verificationType:   "kyc_full",
			currentScore:       50,
			expectedMultiplier: 1.25,
		},
		{
			name:               "phone by elite",
			verificationType:   "phone",
			currentScore:       90,
			expectedMultiplier: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward, err := service.RecordVerification(ctx, "did:echo:test", tt.verificationType, tt.currentScore)

			if err != nil {
				t.Fatalf("RecordVerification() error = %v", err)
			}

			if reward.Multiplier != tt.expectedMultiplier {
				t.Errorf("multiplier = %v, want %v", reward.Multiplier, tt.expectedMultiplier)
			}

			if reward.Amount <= 0 {
				t.Errorf("reward amount = %d, expected > 0", reward.Amount)
			}

			if reward.RewardType != "verification" {
				t.Errorf("reward type = %s, want 'verification'", reward.RewardType)
			}

			if reward.EarnedAt.IsZero() {
				t.Error("earned at is zero")
			}
		})
	}
}

func TestRecordBehavior(t *testing.T) {
	config := getDefaultConfig()
	service := NewTrustScoreService(config)
	ctx := context.Background()

	tests := []struct {
		name               string
		behaviorType       string
		currentScore       int
		count              int
		expectedMultiplier float64
	}{
		{
			name:               "daily login by basic user",
			behaviorType:       "daily_login",
			currentScore:       30,
			count:              1,
			expectedMultiplier: 1.1,
		},
		{
			name:               "multiple messages by verified user",
			behaviorType:       "message",
			currentScore:       70,
			count:              10,
			expectedMultiplier: 1.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reward, err := service.RecordBehavior(ctx, "did:echo:test", tt.behaviorType, tt.currentScore, tt.count)

			if err != nil {
				t.Fatalf("RecordBehavior() error = %v", err)
			}

			if reward.Multiplier != tt.expectedMultiplier {
				t.Errorf("multiplier = %v, want %v", reward.Multiplier, tt.expectedMultiplier)
			}

			if reward.RewardType != "behavior" {
				t.Errorf("reward type = %s, want 'behavior'", reward.RewardType)
			}
		})
	}
}

func TestApplyPenalty(t *testing.T) {
	config := getDefaultConfig()
	service := NewTrustScoreService(config)
	ctx := context.Background()

	penalty, err := service.ApplyPenalty(ctx, "did:echo:test", "spam")

	if err != nil {
		t.Fatalf("ApplyPenalty() error = %v", err)
	}

	if penalty.Amount >= 0 {
		t.Errorf("penalty amount = %d, expected < 0", penalty.Amount)
	}

	if penalty.RewardType != "penalty" {
		t.Errorf("reward type = %s, want 'penalty'", penalty.RewardType)
	}
}

func TestGetEarningMultiplier(t *testing.T) {
	config := getDefaultConfig()
	service := NewTrustScoreService(config)

	tests := []struct {
		score    int
		expected float64
	}{
		{10, 1.0},
		{30, 1.1},
		{50, 1.25},
		{70, 1.5},
		{90, 2.0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			multiplier := service.GetEarningMultiplier(tt.score)
			if multiplier != tt.expected {
				t.Errorf("GetEarningMultiplier(%d) = %v, want %v", tt.score, multiplier, tt.expected)
			}
		})
	}
}

// Benchmark trust score calculation
func BenchmarkCalculateTrustScore(b *testing.B) {
	config := getDefaultConfig()
	service := NewTrustScoreService(config)
	ctx := context.Background()

	verification := &VerificationPoints{PasskeyCreated: 1, PhoneVerified: 1}
	behavior := &BehaviorPoints{AccountAgeMonths: 6, MessageCount: 1000}
	onChain := &OnChainPoints{TransactionCount: 20, StakedAmount: 5000}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CalculateTrustScore(ctx, verification, behavior, onChain, &PenaltyPoints{})
	}
}

// Benchmark reward calculation
func BenchmarkRecordVerification(b *testing.B) {
	config := getDefaultConfig()
	service := NewTrustScoreService(config)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.RecordVerification(ctx, "did:echo:test", "kyc_full", 60)
	}
}

func getDefaultConfig() TrustScoreConfig {
	config := TrustScoreConfig{}

	// Verification weights
	config.VerificationWeights.PasskeyCreated = 5
	config.VerificationWeights.PhoneVerified = 5
	config.VerificationWeights.EmailVerified = 5
	config.VerificationWeights.KYCLite = 10
	config.VerificationWeights.FullKYC = 15
	config.VerificationWeights.AppleDigitalID = 15
	config.VerificationWeights.OrgVerified = 20
	config.VerificationWeights.MaxVerification = 70

	// Behavior weights
	config.BehaviorWeights.AccountAgePerMonth = 1
	config.BehaviorWeights.MessagesPerHundred = 1
	config.BehaviorWeights.ContactsPerTen = 3
	config.BehaviorWeights.GroupsPerFive = 1
	config.BehaviorWeights.DailyLoginStreak = 1
	config.BehaviorWeights.MaxBehavior = 80

	// On-chain weights
	config.OnChainWeights.TransactionsPerTx = 1
	config.OnChainWeights.StakingPer100ECHO = 1
	config.OnChainWeights.GovernancePerVote = 1
	config.OnChainWeights.MaxOnChain = 70

	// Penalty weights
	config.PenaltyWeights.SpamReport = 2
	config.PenaltyWeights.FraudReport = 5
	config.PenaltyWeights.BlockedByUser = 1
	config.PenaltyWeights.InactivityPerMonth = 1
	config.PenaltyWeights.MaxPenalty = 20

	// ECHO rewards
	config.ECHORewards.VerificationRewards = map[string]int64{
		"passkey":  10,
		"phone":    15,
		"email":    10,
		"kyc_lite": 50,
		"kyc_full": 100,
		"apple_id": 100,
		"org":      200,
	}

	config.ECHORewards.BehaviorRewards = map[string]int64{
		"daily_login": 1,
		"contact":     2,
		"group":       5,
	}

	config.ECHORewards.OnChainRewards = map[string]int64{
		"transaction": 1,
		"staking":     1,
		"governance":  10,
	}

	config.ECHORewards.PenaltyReductions = map[string]int64{
		"spam":  10,
		"fraud": 50,
	}

	// Multiplier tiers
	config.MultiplierTiers = map[TrustScoreLevel]float64{
		TrustLevelNewcomer: 1.0,
		TrustLevelBasic:    1.1,
		TrustLevelTrusted:  1.25,
		TrustLevelVerified: 1.5,
		TrustLevelElite:    2.0,
	}

	return config
}
