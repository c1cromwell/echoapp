package auth

import (
	"context"
	"fmt"
	"time"
)

// TrustScoreLevel represents user's trust tier
type TrustScoreLevel string

const (
	TrustLevelNewcomer TrustScoreLevel = "newcomer"
	TrustLevelBasic    TrustScoreLevel = "basic"
	TrustLevelTrusted  TrustScoreLevel = "trusted"
	TrustLevelVerified TrustScoreLevel = "verified"
	TrustLevelElite    TrustScoreLevel = "elite"
)

// VerificationPoints tracks verification progress
type VerificationPoints struct {
	PasskeyCreated    int
	PhoneVerified     int
	EmailVerified     int
	KYCLiteVerified   int
	KYCFullVerified   int
	AppleDigitalID    int
	OrgVerified       int
	PhoneVerifiedBool bool
	EmailVerifiedBool bool
	AppleIDVerified   bool
}

// BehaviorPoints tracks user activity
type BehaviorPoints struct {
	AccountAgeMonths int
	MessageCount     int
	VerifiedContacts int
	ContactCount     int
	ActiveGroups     int
	GroupsCreated    int
	DailyLoginStreak int
}

// OnChainPoints tracks blockchain activity
type OnChainPoints struct {
	TransactionCount    int64
	StakedAmount        float64
	GovernanceVotes     int64
	GovernanceVoteCount int64
	ReferralCount       int64
}

// PenaltyPoints tracks negative behaviors
type PenaltyPoints struct {
	SpamReports    int
	FraudReports   int
	BlockedByUser  int
	InactiveMonths int
}

// TrustScoreSnapshot represents a complete trust score at a point in time
type TrustScoreSnapshot struct {
	UserDID           string
	CurrentScore      int
	Score             int
	Level             TrustScoreLevel
	EarningMultiplier float64
	Verification      VerificationPoints
	Behavior          BehaviorPoints
	OnChain           OnChainPoints
	Penalties         PenaltyPoints
	SnapshotTimestamp time.Time
}

// ECHOReward represents ECHO token rewards
type ECHOReward struct {
	Amount     int64
	Source     string
	Multiplier float64
	RewardType string
	EarnedAt   time.Time
}

// TrustScoreWeights configuration
type TrustScoreWeights struct {
	PasskeyCreated     int
	PhoneVerified      int
	EmailVerified      int
	KYCLite            int
	FullKYC            int
	AppleDigitalID     int
	OrgVerified        int
	MaxVerification    int
	AccountAgePerMonth int
	MessagesPerHundred int
	ContactsPerTen     int
	GroupsPerFive      int
	DailyLoginStreak   int
	MaxBehavior        int
	TransactionsPerTx  int
	StakingPer100ECHO  int
	GovernancePerVote  int
	MaxOnChain         int
	SpamReport         int
	FraudReport        int
	BlockedByUser      int
	InactivityPerMonth int
	MaxPenalty         int
}

// TrustScoreConfig holds all configuration
type TrustScoreConfig struct {
	VerificationWeights TrustScoreWeights
	BehaviorWeights     TrustScoreWeights
	OnChainWeights      TrustScoreWeights
	PenaltyWeights      TrustScoreWeights
	ECHORewards         struct {
		VerificationRewards map[string]int64
		BehaviorRewards     map[string]int64
		OnChainRewards      map[string]int64
		PenaltyReductions   map[string]int64
	}
	MultiplierTiers map[TrustScoreLevel]float64
}

// TrustScoreService manages trust scoring
type TrustScoreService struct {
	config TrustScoreConfig
}

// NewTrustScoreService creates a new service
func NewTrustScoreService(config TrustScoreConfig) *TrustScoreService {
	return &TrustScoreService{config: config}
}

// CalculateTrustScore aggregates all components
func (s *TrustScoreService) CalculateTrustScore(
	ctx context.Context,
	verification *VerificationPoints,
	behavior *BehaviorPoints,
	onChain *OnChainPoints,
	penalties *PenaltyPoints,
) (*TrustScoreSnapshot, error) {

	verScore := calculateVerificationScore(verification, s.config)
	behScore := calculateBehaviorScore(behavior, s.config)
	onChainScore := calculateOnChainScore(onChain, s.config)
	penaltyScore := calculatePenaltyScore(penalties, s.config)

	totalScore := verScore + behScore + onChainScore + penaltyScore
	if totalScore < 0 {
		totalScore = 0
	}
	if totalScore > 100 {
		totalScore = 100
	}

	level := getTrustLevel(totalScore)
	multiplier := s.config.MultiplierTiers[level]

	snapshot := &TrustScoreSnapshot{
		CurrentScore:      totalScore,
		Score:             totalScore,
		Level:             level,
		EarningMultiplier: multiplier,
		Verification:      *verification,
		Behavior:          *behavior,
		OnChain:           *onChain,
		Penalties:         *penalties,
		SnapshotTimestamp: time.Now(),
	}

	return snapshot, nil
}

// RecordVerification tracks verification events
func (s *TrustScoreService) RecordVerification(
	ctx context.Context,
	userDID string,
	verificationType string,
	currentScore int,
) (*ECHOReward, error) {

	baseReward := s.config.ECHORewards.VerificationRewards[verificationType]
	multiplier := s.GetEarningMultiplier(currentScore)
	finalReward := int64(float64(baseReward) * multiplier)

	return &ECHOReward{
		Amount:     finalReward,
		Source:     fmt.Sprintf("verification:%s", verificationType),
		Multiplier: multiplier,
		RewardType: "verification",
		EarnedAt:   time.Now(),
	}, nil
}

// RecordBehavior tracks behavior events
func (s *TrustScoreService) RecordBehavior(
	ctx context.Context,
	userDID string,
	behaviorType string,
	currentScore int,
	count int,
) (*ECHOReward, error) {

	baseReward := s.config.ECHORewards.BehaviorRewards[behaviorType]
	multiplier := s.GetEarningMultiplier(currentScore)
	finalReward := int64(float64(baseReward*int64(count)) * multiplier)

	return &ECHOReward{
		Amount:     finalReward,
		Source:     fmt.Sprintf("behavior:%s", behaviorType),
		Multiplier: multiplier,
		RewardType: "behavior",
		EarnedAt:   time.Now(),
	}, nil
}

// ApplyPenalty records penalty
func (s *TrustScoreService) ApplyPenalty(
	ctx context.Context,
	userDID string,
	penaltyType string,
) (*ECHOReward, error) {

	reduction := s.config.ECHORewards.PenaltyReductions[penaltyType]

	return &ECHOReward{
		Amount:     -reduction,
		Source:     fmt.Sprintf("penalty:%s", penaltyType),
		Multiplier: 1.0,
		RewardType: "penalty",
		EarnedAt:   time.Now(),
	}, nil
}

// GetEarningMultiplier returns current multiplier
func (s *TrustScoreService) GetEarningMultiplier(score int) float64 {
	level := getTrustLevel(score)
	return s.config.MultiplierTiers[level]
}

// Helper functions
func getTrustLevel(score int) TrustScoreLevel {
	switch {
	case score >= 81:
		return TrustLevelElite
	case score >= 61:
		return TrustLevelVerified
	case score >= 41:
		return TrustLevelTrusted
	case score >= 21:
		return TrustLevelBasic
	default:
		return TrustLevelNewcomer
	}
}

func calculateVerificationScore(vp *VerificationPoints, config TrustScoreConfig) int {
	if vp == nil {
		return 0
	}

	score := 0
	score += vp.PasskeyCreated * config.VerificationWeights.PasskeyCreated
	score += vp.PhoneVerified * config.VerificationWeights.PhoneVerified
	score += vp.EmailVerified * config.VerificationWeights.EmailVerified
	score += vp.KYCLiteVerified * config.VerificationWeights.KYCLite
	score += vp.KYCFullVerified * config.VerificationWeights.FullKYC
	score += vp.AppleDigitalID * config.VerificationWeights.AppleDigitalID
	score += vp.OrgVerified * config.VerificationWeights.OrgVerified

	if score > config.VerificationWeights.MaxVerification {
		score = config.VerificationWeights.MaxVerification
	}

	return score
}

func calculateBehaviorScore(bp *BehaviorPoints, config TrustScoreConfig) int {
	if bp == nil {
		return 0
	}

	score := 0
	score += bp.AccountAgeMonths * config.BehaviorWeights.AccountAgePerMonth
	score += (bp.MessageCount / 100) * config.BehaviorWeights.MessagesPerHundred
	score += (bp.VerifiedContacts / 10) * config.BehaviorWeights.ContactsPerTen
	score += (bp.ActiveGroups / 5) * config.BehaviorWeights.GroupsPerFive
	score += bp.DailyLoginStreak * config.BehaviorWeights.DailyLoginStreak

	if score > config.BehaviorWeights.MaxBehavior {
		score = config.BehaviorWeights.MaxBehavior
	}

	return score
}

func calculateOnChainScore(ocp *OnChainPoints, config TrustScoreConfig) int {
	if ocp == nil {
		return 0
	}

	score := 0
	score += int(ocp.TransactionCount) * config.OnChainWeights.TransactionsPerTx
	score += int(ocp.StakedAmount/100) * config.OnChainWeights.StakingPer100ECHO
	score += int(ocp.GovernanceVoteCount) * config.OnChainWeights.GovernancePerVote

	if score > config.OnChainWeights.MaxOnChain {
		score = config.OnChainWeights.MaxOnChain
	}

	return score
}

func calculatePenaltyScore(pp *PenaltyPoints, config TrustScoreConfig) int {
	if pp == nil {
		return 0
	}

	score := 0
	score -= pp.SpamReports * config.PenaltyWeights.SpamReport
	score -= pp.FraudReports * config.PenaltyWeights.FraudReport
	score -= pp.BlockedByUser * config.PenaltyWeights.BlockedByUser
	score -= pp.InactiveMonths * config.PenaltyWeights.InactivityPerMonth

	if score < -config.PenaltyWeights.MaxPenalty {
		score = -config.PenaltyWeights.MaxPenalty
	}

	return score
}
