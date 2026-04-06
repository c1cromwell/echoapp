package emissions

import (
	"math/big"
	"time"
)

// EmissionSchedule manages token emission rates
type EmissionSchedule struct {
	StartTime    time.Time
	HalvingYears int
}

// NewEmissionSchedule creates a new emission schedule
func NewEmissionSchedule(startTime time.Time) *EmissionSchedule {
	return &EmissionSchedule{
		StartTime:    startTime,
		HalvingYears: 2,
	}
}

// UserRewardEmissionDaily calculates daily user reward emissions
func (es *EmissionSchedule) UserRewardEmissionDaily(atTime time.Time) *big.Int {
	elapsed := atTime.Sub(es.StartTime)
	days := int(elapsed.Hours() / 24)
	years := days / 365

	// Halving every 2 years
	halvings := years / 2

	initialEmission := new(big.Int)
	initialEmission.SetString("27397260000000", 10) // 273,972.60 ECHO with decimals

	for i := 0; i < halvings; i++ {
		initialEmission.Div(initialEmission, big.NewInt(2))
	}

	minimum := new(big.Int)
	minimum.SetString("2739726000000", 10) // 27,397.26 ECHO floor

	if initialEmission.Cmp(minimum) < 0 {
		return minimum
	}

	return initialEmission
}

// ValidatorRewardEmissionAnnual calculates annual validator emissions
func (es *EmissionSchedule) ValidatorRewardEmissionAnnual(atTime time.Time) *big.Int {
	elapsed := atTime.Sub(es.StartTime)
	years := int(elapsed.Hours() / (24 * 365))

	result := new(big.Int)

	switch {
	case years < 2:
		result.SetString("5000000000000000", 10) // 50M ECHO
	case years < 5:
		result.SetString("3000000000000000", 10) // 30M ECHO
	case years < 10:
		result.SetString("1000000000000000", 10) // 10M ECHO
	default:
		result.SetInt64(0) // Fee-based only
	}

	return result
}

// InflationRate calculates current inflation rate
func (es *EmissionSchedule) InflationRate(atTime time.Time) float64 {
	dailyEmission := es.UserRewardEmissionDaily(atTime)
	annualEmission := new(big.Int).Mul(dailyEmission, big.NewInt(365))

	totalSupply := new(big.Int)
	totalSupply.SetString("100000000000000000", 10)

	rate := new(big.Float).Quo(
		new(big.Float).SetInt(annualEmission),
		new(big.Float).SetInt(totalSupply),
	)

	rateFloat, _ := rate.Float64()
	return rateFloat * 100
}

// EmissionPhase represents an emission phase
type EmissionPhase struct {
	StartYear int
	EndYear   int
	Rate      *big.Int
	Name      string
}

// GetCurrentPhase gets the current emission phase
func (es *EmissionSchedule) GetCurrentPhase(atTime time.Time) *EmissionPhase {
	elapsed := atTime.Sub(es.StartTime)
	years := int(elapsed.Hours() / (24 * 365))

	if years < 2 {
		return &EmissionPhase{
			StartYear: 0,
			EndYear:   2,
			Name:      "Bootstrap",
		}
	} else if years < 5 {
		return &EmissionPhase{
			StartYear: 2,
			EndYear:   5,
			Name:      "Growth",
		}
	} else if years < 10 {
		return &EmissionPhase{
			StartYear: 5,
			EndYear:   10,
			Name:      "Mature",
		}
	}

	return &EmissionPhase{
		StartYear: 10,
		EndYear:   -1,
		Name:      "Stable",
	}
}
