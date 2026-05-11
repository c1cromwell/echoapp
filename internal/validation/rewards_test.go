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

func TestComputeVolumeDecay(t *testing.T) {
	cases := []struct {
		count int
		want  float64
	}{
		{0, 1.0},
		{-5, 1.0},    // negative treated as zero
		{9, 0.7071},  // 1/sqrt(1+log10(10)) = 1/sqrt(2) ≈ 0.7071
		{99, 0.5774}, // 1/sqrt(1+log10(100)) = 1/sqrt(3) ≈ 0.5774
		{999, 0.5},   // 1/sqrt(1+log10(1000)) = 1/sqrt(4) = 0.5
	}
	for _, tc := range cases {
		got := ComputeVolumeDecay(tc.count)
		if got != tc.want {
			t.Errorf("ComputeVolumeDecay(%d) = %v, want %v", tc.count, got, tc.want)
		}
	}
}

func TestComputeVolumeDecay_Floor(t *testing.T) {
	// Very high message count should never go below minFloor (0.10)
	got := ComputeVolumeDecay(10_000_000)
	if got < 0.10 {
		t.Errorf("decay below floor: %v", got)
	}
}

func TestValidateRewardClaimPre(t *testing.T) {
	mults := map[int]float64{2: 0.5, 3: 1.0, 4: 1.5}

	t.Run("ok — no optional fields submitted", func(t *testing.T) {
		err := ValidateRewardClaimPre(RewardClaimPreValidation{
			CachedTrustTier: 3,
			TierMultipliers: mults,
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("ok — correct decay and multiplier", func(t *testing.T) {
		decay := ComputeVolumeDecay(9) // 0.7071
		mult := 1.0
		err := ValidateRewardClaimPre(RewardClaimPreValidation{
			SubmittedDecayFactor:     &decay,
			SubmittedTrustMultiplier: &mult,
			MessageCount:             9,
			CachedTrustTier:          3,
			TierMultipliers:          mults,
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("wrong decay factor", func(t *testing.T) {
		wrong := 0.99
		err := ValidateRewardClaimPre(RewardClaimPreValidation{
			SubmittedDecayFactor: &wrong,
			MessageCount:         9,
			CachedTrustTier:      3,
			TierMultipliers:      mults,
		})
		if !errors.Is(err, ErrInvalidDecayFactor) {
			t.Fatalf("want ErrInvalidDecayFactor, got %v", err)
		}
	})

	t.Run("wrong trust multiplier", func(t *testing.T) {
		bad := 99.0
		err := ValidateRewardClaimPre(RewardClaimPreValidation{
			SubmittedTrustMultiplier: &bad,
			CachedTrustTier:          3,
			TierMultipliers:          mults,
		})
		if !errors.Is(err, ErrInvalidTrustMultiplier) {
			t.Fatalf("want ErrInvalidTrustMultiplier, got %v", err)
		}
	})

	t.Run("unknown trust tier", func(t *testing.T) {
		mult := 1.0
		err := ValidateRewardClaimPre(RewardClaimPreValidation{
			SubmittedTrustMultiplier: &mult,
			CachedTrustTier:          99,
			TierMultipliers:          mults,
		})
		if !errors.Is(err, ErrInvalidTrustMultiplier) {
			t.Fatalf("want ErrInvalidTrustMultiplier, got %v", err)
		}
	})
}
