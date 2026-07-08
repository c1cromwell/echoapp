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

// DailyRewardTracker tracks daily reward activity.
//
// Per PRD v2.5.1 the auto-scaling model replaced the old hard daily cap: every
// message always earns, and the per-message rate scales with network activity
// (rate = daily budget / total daily activity weight). There is no cutoff after
// which a user stops earning — mirrors iOS `NetworkActivityTracker`.
type DailyRewardTracker struct {
	UserID           string
	Date             time.Time
	MessagesRewarded int
	EchoEarned       *big.Int
	TotalActions     int
}

// IsLimitReached is retained for compatibility but always reports false under the
// auto-scaling model (PRD v2.5.1) — there is no hard daily cap. Kept so existing
// callers compile; earning throttling is handled by the network-activity rate,
// not a per-user wall.
func (drt *DailyRewardTracker) IsLimitReached() bool {
	return false
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

// GetMultiplier returns the reward multiplier for a trust score.
//
// Canonical table (Tokenomics v2.0) — kept in lockstep with iOS
// `RewardsTrustScore.getMultiplier()` (ios/Echo/Sources/Models/Rewards.swift):
// Tier 1 (0–19) 1.0×, Tier 2 (20–39) 1.2×, Tier 3 (40–59) 1.5×,
// Tier 4 (60–79) 2.0×, Tier 5 (80–100) 3.0×.
func (ts *TrustScore) GetMultiplier() float64 {
	switch {
	case ts.Score < 20:
		return 1.0
	case ts.Score < 40:
		return 1.2
	case ts.Score < 60:
		return 1.5
	case ts.Score < 80:
		return 2.0
	default:
		return 3.0
	}
}
