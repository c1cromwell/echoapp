package models

import (
	"math/big"
	"testing"
	"time"
)

func TestTokenConfiguration(t *testing.T) {
	config := NewTokenConfig()

	if config.Name != "ECHO" {
		t.Errorf("Token name: got %v, want ECHO", config.Name)
	}

	if config.Symbol != "ECHO" {
		t.Errorf("Token symbol: got %v, want ECHO", config.Symbol)
	}

	if config.Decimals != 8 {
		t.Errorf("Decimals: got %d, want 8", config.Decimals)
	}

	if !config.HardCapped {
		t.Errorf("Hard capped: got false, want true")
	}
}

func TestTotalSupply(t *testing.T) {
	config := NewTokenConfig()
	expected := new(big.Int)
	expected.SetString("100000000000000000", 10)

	if config.TotalSupply.Cmp(expected) != 0 {
		t.Errorf("TotalSupply mismatch")
	}
}

func TestAllocationBreakdown(t *testing.T) {
	allocation := NewAllocationBreakdown()

	total := new(big.Int)
	total.Add(allocation.UserRewards, allocation.ValidatorRewards)
	total.Add(total, allocation.Ecosystem)
	total.Add(total, allocation.Team)
	total.Add(total, allocation.Treasury)
	total.Add(total, allocation.Liquidity)

	expected := new(big.Int)
	expected.SetString("100000000000000000", 10)

	if total.Cmp(expected) != 0 {
		t.Errorf("Allocation total mismatch")
	}
}

func TestVestingSchedule(t *testing.T) {
	schedule := &VestingSchedule{
		TotalAmount:   big.NewInt(1000),
		ReleasedAt:    time.Now(),
		CliffMonths:   12,
		VestMonths:    24,
		ReleasedSoFar: big.NewInt(0),
	}

	releasable := schedule.CalculateReleasable()
	if releasable.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("Before cliff should return 0")
	}
}

func BenchmarkTokenConfiguration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewTokenConfig()
	}
}

func BenchmarkAllocationBreakdown(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewAllocationBreakdown()
	}
}
