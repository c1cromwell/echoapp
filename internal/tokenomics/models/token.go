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

// AllocationBreakdown defines token distribution percentages
type AllocationBreakdown struct {
	UserRewards      *big.Int // 40% - 400M
	ValidatorRewards *big.Int // 25% - 250M
	Ecosystem        *big.Int // 20% - 200M
	Team             *big.Int // 8% - 80M
	Treasury         *big.Int // 5% - 50M
	Liquidity        *big.Int // 2% - 20M
}

// NewAllocationBreakdown creates the allocation breakdown
func NewAllocationBreakdown() *AllocationBreakdown {
	total := new(big.Int)
	total.SetString("100000000000000000", 10)

	userRewards := new(big.Int).Mul(total, big.NewInt(40))
	userRewards.Div(userRewards, big.NewInt(100))

	validatorRewards := new(big.Int).Mul(total, big.NewInt(25))
	validatorRewards.Div(validatorRewards, big.NewInt(100))

	ecosystem := new(big.Int).Mul(total, big.NewInt(20))
	ecosystem.Div(ecosystem, big.NewInt(100))

	team := new(big.Int).Mul(total, big.NewInt(8))
	team.Div(team, big.NewInt(100))

	treasury := new(big.Int).Mul(total, big.NewInt(5))
	treasury.Div(treasury, big.NewInt(100))

	liquidity := new(big.Int).Mul(total, big.NewInt(2))
	liquidity.Div(liquidity, big.NewInt(100))

	return &AllocationBreakdown{
		UserRewards:      userRewards,
		ValidatorRewards: validatorRewards,
		Ecosystem:        ecosystem,
		Team:             team,
		Treasury:         treasury,
		Liquidity:        liquidity,
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
