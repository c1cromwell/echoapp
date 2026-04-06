package models

import (
	"math/big"
	"time"
)

// RewardType represents different reward categories
type RewardType int

const (
	TextReward RewardType = iota
	VoiceReward
	VideoReward
	ReferralReward
	GovernanceReward
	StakingReward
	BurnReward
	BridgeReward
)

// RewardEarning represents a single earned reward
type RewardEarning struct {
	UserID     string
	RewardType RewardType
	Amount     *big.Int
	Multiplier float64
	EarnedAt   time.Time
	Claimed    bool
	ClaimedAt  time.Time
}

// DailyRewardTracker tracks daily reward limits
type DailyRewardTracker struct {
	UserID           string
	Date             time.Time
	MessagesRewarded int
	EchoEarned       *big.Int
	TotalActions     int
}

// IsLimitReached checks if daily limits exceeded
func (drt *DailyRewardTracker) IsLimitReached() bool {
	return drt.MessagesRewarded >= 500
}

// ReferralInfo tracks referral bonuses
type ReferralInfo struct {
	ReferrerID     string
	RefereeID      string
	SignupBonus    *big.Int
	VerifyBonus    *big.Int
	MilestoneBonus *big.Int
	TotalBonus     *big.Int
	CreatedAt      time.Time
}

// TrustScore represents user trust level
type TrustScore struct {
	UserID     string
	Score      int // 0-100
	Level      string
	UpdatedAt  time.Time
	Components map[string]int
}

// GetMultiplier returns reward multiplier for trust score
func (ts *TrustScore) GetMultiplier() float64 {
	switch {
	case ts.Score < 20:
		return 0.5
	case ts.Score < 40:
		return 1.0
	case ts.Score < 60:
		return 1.5
	case ts.Score < 80:
		return 2.5
	default:
		return 5.0
	}
}
