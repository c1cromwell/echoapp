package models

import (
	"math/big"
	"time"
)

// TokenConfig represents the core ECHO token specification
type TokenConfig struct {
	Name        string
	Symbol      string
	TotalSupply *big.Int // 1 billion * 10^8
	Decimals    int
	HardCapped  bool
}

// NewTokenConfig creates the ECHO token configuration
func NewTokenConfig() *TokenConfig {
	totalSupply := new(big.Int)
	totalSupply.SetString("100000000000000000", 10) // 1 billion * 10^8

	return &TokenConfig{
		Name:        "ECHO",
		Symbol:      "ECHO",
		TotalSupply: totalSupply,
		Decimals:    8,
		HardCapped:  true,
	}
}

// AllocationBreakdown defines WO-214 genesis pool distribution (1B ECHO fixed supply).
type AllocationBreakdown struct {
	CommunityRewards *big.Int // 400M — 10-year emission account
	Treasury         *big.Int // 220M — 3-of-5 multi-sig
	Founders         *big.Int // 180M — TokenLock positions
	FutureTeam       *big.Int // 100M — multi-sig controlled
	Ecosystem        *big.Int // 100M — governance-controlled
}

// NewAllocationBreakdown creates the WO-214 genesis allocation breakdown.
func NewAllocationBreakdown() *AllocationBreakdown {
	echo := func(millions int64) *big.Int {
		// millions * 1e6 ECHO * 1e8 datum
		return new(big.Int).Mul(big.NewInt(millions), big.NewInt(100_000_000_000000))
	}
	return &AllocationBreakdown{
		CommunityRewards: echo(400),
		Treasury:         echo(220),
		Founders:         echo(180),
		FutureTeam:       echo(100),
		Ecosystem:        echo(100),
	}
}

// TokenBalance tracks user balance and vesting
type TokenBalance struct {
	Address          string
	AvailableBalance *big.Int
	VestingSchedule  *VestingSchedule
}

// VestingSchedule defines time-locked token releases
type VestingSchedule struct {
	TotalAmount   *big.Int
	ReleasedAt    time.Time
	CliffMonths   int
	VestMonths    int
	ReleasedSoFar *big.Int
}

// CalculateReleasable calculates currently releasable amount
func (vs *VestingSchedule) CalculateReleasable() *big.Int {
	elapsed := time.Since(vs.ReleasedAt)
	monthsElapsed := int(elapsed.Hours() / 730)

	if monthsElapsed < vs.CliffMonths {
		return big.NewInt(0)
	}

	if monthsElapsed >= vs.CliffMonths+vs.VestMonths {
		return new(big.Int).Sub(vs.TotalAmount, vs.ReleasedSoFar)
	}

	vestingMonths := monthsElapsed - vs.CliffMonths
	monthlyRelease := new(big.Int).Div(vs.TotalAmount, big.NewInt(int64(vs.VestMonths)))
	releasable := new(big.Int).Mul(monthlyRelease, big.NewInt(int64(vestingMonths)))

	releasable.Sub(releasable, vs.ReleasedSoFar)
	return releasable
}
