package rewards

import (
	"testing"
	"time"
)

func TestYearlyBudget(t *testing.T) {
	genesis := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	es := NewEmissionSchedule(genesis)

	// Year 1: 20% of 400M = 80M ECHO (with 8 decimals)
	year1 := es.YearlyBudget(1)
	expected1 := int64(float64(400_000_000_00000000) * 0.20)
	if year1 != expected1 {
		t.Errorf("year 1 budget: expected %d, got %d", expected1, year1)
	}

	// Year 2: 16% = 64M
	year2 := es.YearlyBudget(2)
	expected2 := int64(float64(400_000_000_00000000) * 0.16)
	if year2 != expected2 {
		t.Errorf("year 2 budget: expected %d, got %d", expected2, year2)
	}

	// Year 10: 6% = 24M
	year10 := es.YearlyBudget(10)
	expected10 := int64(float64(400_000_000_00000000) * 0.06)
	if year10 != expected10 {
		t.Errorf("year 10 budget: expected %d, got %d", expected10, year10)
	}

	// Year 0 and 11: 0
	if es.YearlyBudget(0) != 0 {
		t.Error("year 0 should have 0 budget")
	}
	if es.YearlyBudget(11) != 0 {
		t.Error("year 11 should have 0 budget")
	}
}

func TestDailyBudget(t *testing.T) {
	// Genesis was just now, so we're in year 1
	genesis := time.Now().Add(-24 * time.Hour)
	es := NewEmissionSchedule(genesis)

	daily := es.DailyBudget()
	yearlyBudget := es.YearlyBudget(1)
	expected := yearlyBudget / 365

	if daily != expected {
		t.Errorf("daily budget: expected %d, got %d", expected, daily)
	}
}

func TestRemainingToday(t *testing.T) {
	genesis := time.Now().Add(-24 * time.Hour)
	es := NewEmissionSchedule(genesis)

	daily := es.DailyBudget()

	// No claims yet
	remaining := es.RemainingToday(0)
	if remaining != daily {
		t.Errorf("expected full daily budget %d, got %d", daily, remaining)
	}

	// Half claimed
	half := daily / 2
	remaining = es.RemainingToday(half)
	if remaining != daily-half {
		t.Errorf("expected %d remaining, got %d", daily-half, remaining)
	}

	// Over-claimed (shouldn't happen but handled)
	remaining = es.RemainingToday(daily + 1000)
	if remaining != 0 {
		t.Errorf("expected 0 when over-claimed, got %d", remaining)
	}
}

func TestCurrentYear(t *testing.T) {
	genesis := time.Now().Add(-24 * time.Hour)
	es := NewEmissionSchedule(genesis)

	if es.CurrentYear() != 1 {
		t.Errorf("expected year 1, got %d", es.CurrentYear())
	}

	// Genesis was 2 years ago
	es2 := NewEmissionSchedule(time.Now().Add(-2 * 365 * 24 * time.Hour))
	if es2.CurrentYear() != 3 {
		t.Errorf("expected year 3, got %d", es2.CurrentYear())
	}
}

func TestTotalEmittedByYear(t *testing.T) {
	genesis := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	es := NewEmissionSchedule(genesis)

	// Through year 1: 80M
	total1 := es.TotalEmittedByYear(1)
	if total1 != es.YearlyBudget(1) {
		t.Errorf("total through year 1: expected %d, got %d", es.YearlyBudget(1), total1)
	}

	// Through year 2: 80M + 64M = 144M
	total2 := es.TotalEmittedByYear(2)
	expected := es.YearlyBudget(1) + es.YearlyBudget(2)
	if total2 != expected {
		t.Errorf("total through year 2: expected %d, got %d", expected, total2)
	}

	// Through year 10: sum of all percentages (20+16+13+11+9+7+6+6+6+6 = 100%)
	total10 := es.TotalEmittedByYear(10)
	if total10 != es.TotalPool {
		t.Errorf("total through year 10 should equal total pool: expected %d, got %d", es.TotalPool, total10)
	}
}

func TestNoEmissionAfterYear10(t *testing.T) {
	// Genesis was 11 years ago
	genesis := time.Now().Add(-11 * 365 * 24 * time.Hour)
	es := NewEmissionSchedule(genesis)

	daily := es.DailyBudget()
	if daily != 0 {
		t.Errorf("expected 0 daily budget after year 10, got %d", daily)
	}
}
