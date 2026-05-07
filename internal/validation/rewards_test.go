package validation

import (
	"errors"
	"testing"
)

func TestValidateRewardDecay(t *testing.T) {
	if err := ValidateRewardDecay(0.5, 0.5); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRewardDecay(0.4, 0.5); !errors.Is(err, ErrInvalidDecayFactor) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateTrustMultiplier(t *testing.T) {
	if err := ValidateTrustMultiplier(1.5, 1.5); err != nil {
		t.Fatal(err)
	}
	if err := ValidateTrustMultiplier(2.0, 1.0); !errors.Is(err, ErrInvalidTrustMultiplier) {
		t.Fatalf("got %v", err)
	}
}
