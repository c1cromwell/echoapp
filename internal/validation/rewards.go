package validation

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidDecayFactor   = errors.New("validation: decay_factor does not match server model")
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
