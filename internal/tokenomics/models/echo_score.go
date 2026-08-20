package models

// EchoScoreSnapshot is the Rewards-home living score (Cero-style number,
// Echo-native inputs). Score is 0–100; tier 1–5 matches GetMultiplier.
type EchoScoreSnapshot struct {
	Score      int            `json:"score"`
	Tier       int            `json:"tier"`
	Level      string         `json:"level"`
	Multiplier float64        `json:"multiplier"`
	NextUnlock *UnlockFeature `json:"next_unlock,omitempty"`
}

// UnlockFeature is the next trust-gated capability. Copy is standing, not cash.
type UnlockFeature struct {
	Tier         int    `json:"tier"`
	MinScore     int    `json:"min_score"`
	Feature      string `json:"feature"`
	PointsNeeded int    `json:"points_needed"`
}

var unlockLadder = []UnlockFeature{
	{Tier: 2, MinScore: 20, Feature: "Appear on the weekly leaderboard"},
	{Tier: 3, MinScore: 40, Feature: "Create broadcast channels"},
	{Tier: 4, MinScore: 60, Feature: "Governance voting"},
	{Tier: 5, MinScore: 80, Feature: "Highest earn multiplier"},
}

// SnapshotFromScore builds the Rewards-home score card from a 0–100 trust score.
func SnapshotFromScore(score int) EchoScoreSnapshot {
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	ts := &TrustScore{Score: score}
	snap := EchoScoreSnapshot{
		Score:      score,
		Tier:       TierFromScore(score),
		Level:      LevelFromScore(score),
		Multiplier: ts.GetMultiplier(),
	}
	for i := range unlockLadder {
		if score < unlockLadder[i].MinScore {
			u := unlockLadder[i]
			u.PointsNeeded = u.MinScore - score
			snap.NextUnlock = &u
			break
		}
	}
	return snap
}

// SnapshotFromTier uses a 1–5 trust tier when a 0–100 score is unavailable
// (JWT claim only). Midpoints match RewardsTrustScore buckets.
func SnapshotFromTier(tier int) EchoScoreSnapshot {
	return SnapshotFromScore(ScoreMidpointForTier(tier))
}

// TierFromScore maps 0–100 onto trust tiers 1–5.
func TierFromScore(score int) int {
	switch {
	case score < 20:
		return 1
	case score < 40:
		return 2
	case score < 60:
		return 3
	case score < 80:
		return 4
	default:
		return 5
	}
}

// LevelFromScore returns the canonical tier label.
func LevelFromScore(score int) string {
	switch {
	case score < 20:
		return "unverified"
	case score < 40:
		return "newcomer"
	case score < 60:
		return "member"
	case score < 80:
		return "trusted"
	default:
		return "verified"
	}
}

// ScoreMidpointForTier returns a representative 0–100 score for a 1–5 tier.
func ScoreMidpointForTier(tier int) int {
	switch {
	case tier <= 1:
		return 10
	case tier == 2:
		return 30
	case tier == 3:
		return 50
	case tier == 4:
		return 70
	default:
		return 90
	}
}
