// Package rewards implements the 10-year declining emission curve and daily cap tracking.
// 400M ECHO community reward pool, declining annually per tokenomics spec.
package rewards

import "time"

// EmissionSchedule manages the 10-year declining emission curve for community rewards.
// 400M ECHO total, declining annually. Enforced by Currency L1 Scala validation.
// Go backend pre-validates to avoid rejected submissions.
type EmissionSchedule struct {
	GenesisDate   time.Time
	TotalPool     int64     // 400_000_000_00000000 (with 8 decimal places)
	YearlyPercent []float64 // percentage of pool emitted each year
}

// NewEmissionSchedule creates the emission schedule from genesis.
func NewEmissionSchedule(genesis time.Time) *EmissionSchedule {
	return &EmissionSchedule{
		GenesisDate:   genesis,
		TotalPool:     400_000_000_00000000,
		YearlyPercent: []float64{0.20, 0.16, 0.13, 0.11, 0.09, 0.07, 0.06, 0.06, 0.06, 0.06},
	}
}

// YearlyBudget returns the emission budget for a given year (1-indexed).
func (e *EmissionSchedule) YearlyBudget(year int) int64 {
	if year < 1 || year > 10 {
		return 0
	}
	return int64(float64(e.TotalPool) * e.YearlyPercent[year-1])
}

// DailyBudget returns the daily emission budget for the current year.
func (e *EmissionSchedule) DailyBudget() int64 {
	year := e.CurrentYear()
	if year > 10 {
		return 0
	}
	return e.YearlyBudget(year) / 365
}

// RemainingToday returns how much of today's budget remains after claimed rewards.
func (e *EmissionSchedule) RemainingToday(claimedToday int64) int64 {
	daily := e.DailyBudget()
	remaining := daily - claimedToday
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CurrentYear returns the current year (1-indexed) since genesis.
func (e *EmissionSchedule) CurrentYear() int {
	elapsed := time.Since(e.GenesisDate)
	return int(elapsed.Hours()/8760) + 1
}

// TotalEmittedByYear returns cumulative emissions through the given year.
func (e *EmissionSchedule) TotalEmittedByYear(year int) int64 {
	var total int64
	for y := 1; y <= year && y <= 10; y++ {
		total += e.YearlyBudget(y)
	}
	return total
}

// RemainingPool returns how much of the 400M pool has not yet been emitted.
func (e *EmissionSchedule) RemainingPool() int64 {
	emitted := e.TotalEmittedByYear(e.CurrentYear() - 1)
	// Add partial year
	dayOfYear := time.Since(e.GenesisDate).Hours() / 24
	yearStart := float64((e.CurrentYear() - 1) * 365)
	daysIntoYear := dayOfYear - yearStart
	if daysIntoYear > 0 {
		emitted += int64(float64(e.YearlyBudget(e.CurrentYear())) * daysIntoYear / 365)
	}
	remaining := e.TotalPool - emitted
	if remaining < 0 {
		return 0
	}
	return remaining
}
