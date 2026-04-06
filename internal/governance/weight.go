package governance

// TierMultipliers maps trust tier (1-5) to basis points (10000 = 1.0x).
// Tier 1 (Unverified) gets zero governance power.
// This prevents plutocratic capture — unverified whales have no vote weight.
var TierMultipliers = map[int]int64{
	1: 0,     // Unverified: no governance
	2: 5000,  // 0.5x
	3: 10000, // 1.0x
	4: 15000, // 1.5x
	5: 20000, // 2.0x
}

// TierMultiplierFloat returns the human-readable multiplier for display.
func TierMultiplierFloat(tier int) float64 {
	bps, ok := TierMultipliers[tier]
	if !ok {
		return 0
	}
	return float64(bps) / 10000.0
}

// CalculateWeight computes GovernanceWeight = (StakedECHO × MultiplierBps) / 10000.
// Returns 0 if the tier is invalid or has a zero multiplier.
func CalculateWeight(totalStaked int64, trustTier int) int64 {
	multiplier, ok := TierMultipliers[trustTier]
	if !ok || multiplier == 0 {
		return 0
	}
	return (totalStaked * multiplier) / 10000
}

// CanVote returns true if the user meets minimum governance requirements:
// Trust Tier 2+ and at least one staked position (totalStaked > 0).
func CanVote(trustTier int, totalStaked int64) bool {
	return trustTier >= 2 && totalStaked > 0
}

// CheckThresholdPassed evaluates whether a proposal passes its threshold.
func CheckThresholdPassed(threshold string, forWeight, againstWeight, totalWeight int64) bool {
	if totalWeight == 0 {
		return false
	}

	forPercent := (forWeight * 100) / totalWeight

	switch threshold {
	case ThresholdSimpleMajority:
		return forWeight > againstWeight
	case ThresholdSupermajority67:
		return forPercent >= 67
	case ThresholdSupermajority75:
		return forPercent >= 75
	default:
		return false
	}
}
