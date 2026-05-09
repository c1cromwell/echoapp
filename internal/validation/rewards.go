// T0–T7: Reward claim pre-validation operates on T5/T6 data only.
// DecayFactor and TrustMultiplier are numeric parameters — no PII.
// The calling handler must authenticate the DID separately (T7 public data)
// and must NOT pass message content or key material into these functions.
package validation

import (
	"errors"
	"fmt"
	"math"
)

var (
	ErrInvalidDecayFactor     = errors.New("validation: decay_factor does not match server model")
	ErrInvalidTrustMultiplier = errors.New("validation: trust_multiplier does not match cached tier")
)

// ValidateRewardDecay checks the client-submitted decay against the server-computed value (WO-35).
func ValidateRewardDecay(submittedDecay, expectedDecay float64) error {
	if submittedDecay != expectedDecay {
		return fmt.Errorf("%w: expected %v, got %v", ErrInvalidDecayFactor, expectedDecay, submittedDecay)
	}
	return nil
}

// ValidateTrustMultiplier checks the submitted multiplier against the tier cache (WO-35).
func ValidateTrustMultiplier(submittedMultiplier, expectedMultiplier float64) error {
	if submittedMultiplier != expectedMultiplier {
		return fmt.Errorf("%w: expected %v, got %v", ErrInvalidTrustMultiplier, expectedMultiplier, submittedMultiplier)
	}
	return nil
}

// ComputeVolumeDecay returns the expected per-user messaging decay factor for a given
// message count in the current reward period (WO-35 / WO-213 model).
//
// Formula: decay = 1 / (1 + log10(count+1))^0.5, floored at 0.10, rounded to 4 d.p.
// This gives decay values that iOS and the backend can compute identically:
//
//	count=0   → 1.0000
//	count=9   → 0.7071
//	count=99  → 0.5774
//	count=999 → 0.5000
func ComputeVolumeDecay(messagesCount int) float64 {
	const minFloor = 0.10
	if messagesCount <= 0 {
		return 1.0
	}
	raw := 1.0 / math.Sqrt(1.0+math.Log10(float64(messagesCount+1)))
	rounded := math.Round(raw*10000) / 10000
	if rounded < minFloor {
		return minFloor
	}
	return rounded
}

// RewardClaimPreValidation holds the fields needed for WO-35 server-side pre-validation.
type RewardClaimPreValidation struct {
	// SubmittedDecayFactor is the decay factor the client sent with the claim.
	// Nil means the client did not submit one (validation is skipped for that field).
	SubmittedDecayFactor *float64
	// SubmittedTrustMultiplier is the multiplier the client sent.
	// Nil means the client did not submit one.
	SubmittedTrustMultiplier *float64
	// MessageCount is the per-user message count for decay computation.
	MessageCount int
	// CachedTrustTier is the server's current cached trust tier for the DID.
	CachedTrustTier int
	// TierMultipliers maps trust tiers to expected multipliers.
	TierMultipliers map[int]float64
}

// ValidateRewardClaimPre runs all WO-35 reward pre-validation checks.
// L1 remains authoritative; this is a fast rejection layer only.
func ValidateRewardClaimPre(v RewardClaimPreValidation) error {
	if v.SubmittedDecayFactor != nil {
		expected := ComputeVolumeDecay(v.MessageCount)
		if err := ValidateRewardDecay(*v.SubmittedDecayFactor, expected); err != nil {
			return err
		}
	}
	if v.SubmittedTrustMultiplier != nil {
		expectedMult, ok := v.TierMultipliers[v.CachedTrustTier]
		if !ok {
			return fmt.Errorf("%w: unknown trust tier %d", ErrInvalidTrustMultiplier, v.CachedTrustTier)
		}
		if err := ValidateTrustMultiplier(*v.SubmittedTrustMultiplier, expectedMult); err != nil {
			return err
		}
	}
	return nil
}
