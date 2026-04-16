package rewards

import (
	"context"
	"errors"
	"sync"
	"time"
)

// AutoScaleState tracks the auto-scaling reward rate for the current day.
// Rate = Daily Budget ÷ Total Daily Activity Weight
// Activity weight per message = 1.0 × sender's trust tier multiplier
//
// As network activity grows, the per-message rate declines — but every
// message always earns something. Unused budget from low-activity days
// rolls forward within the same calendar year.
//
// Replaces DailyCapTracker (removed per PRD v2.5.1).
type AutoScaleState struct {
	Date                 time.Time `json:"date"`
	TotalActivityWeight  float64   `json:"totalActivityWeight"`  // Sum of all trust-tier-weighted messages today
	BudgetUsedToday      int64     `json:"budgetUsedToday"`      // Total ECHO distributed today
	CurrentRate          int64     `json:"currentRate"`          // Current per-message base rate (8 decimal places)
	DailyBudget          int64     `json:"dailyBudget"`          // Today's emission budget from annual curve
	EffectiveDailyBudget int64     `json:"effectiveDailyBudget"` // Including rollover from low-activity days
	RemainingToday       int64     `json:"remainingToday"`       // Budget remaining today
	RolloverBudget       int64     `json:"rolloverBudget"`       // Unused budget rolled forward from previous days this year
	LastUpdated          time.Time `json:"lastUpdated"`
}

// AutoScaleStore abstracts the persistence layer for auto-scale state.
type AutoScaleStore interface {
	GetState(ctx context.Context) (*AutoScaleState, error)
	SetState(ctx context.Context, state *AutoScaleState) error
}

// AutoScaleRateTracker manages the auto-scaling reward rate.
type AutoScaleRateTracker struct {
	store    AutoScaleStore
	emission *EmissionSchedule
	mu       sync.RWMutex
}

// NewAutoScaleRateTracker creates a tracker with the given store and emission schedule.
func NewAutoScaleRateTracker(store AutoScaleStore, emission *EmissionSchedule) *AutoScaleRateTracker {
	return &AutoScaleRateTracker{
		store:    store,
		emission: emission,
	}
}

// CurrentRate returns the current auto-scaled rate for a message claim.
func (t *AutoScaleRateTracker) CurrentRate(ctx context.Context) (int64, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state, err := t.getOrInitState(ctx)
	if err != nil {
		return 0, err
	}

	rolledOver := t.maybeRolloverDay(state)
	if rolledOver {
		if err := t.store.SetState(ctx, state); err != nil {
			return 0, err
		}
	}
	return state.CurrentRate, nil
}

// RecordClaim records a reward claim and recalculates the auto-scale rate.
func (t *AutoScaleRateTracker) RecordClaim(ctx context.Context, amount int64, trustTierMultiplier float64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	state, err := t.getOrInitState(ctx)
	if err != nil {
		return err
	}

	t.maybeRolloverDay(state)

	state.TotalActivityWeight += trustTierMultiplier
	state.BudgetUsedToday += amount
	state.CurrentRate = t.recalculateRate(state.TotalActivityWeight, state)
	state.LastUpdated = time.Now()

	return t.store.SetState(ctx, state)
}

// GetAutoScaleState returns the current auto-scale state for API responses.
func (t *AutoScaleRateTracker) GetAutoScaleState(ctx context.Context) (*AutoScaleState, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state, err := t.getOrInitState(ctx)
	if err != nil {
		return nil, err
	}

	rolledOver := t.maybeRolloverDay(state)
	if rolledOver {
		if err := t.store.SetState(ctx, state); err != nil {
			return nil, err
		}
	}
	effectiveBudget := t.effectiveDailyBudget(state)
	remaining := effectiveBudget - state.BudgetUsedToday
	if remaining < 0 {
		remaining = 0
	}

	// Return a copy for thread safety
	return &AutoScaleState{
		Date:                 state.Date,
		TotalActivityWeight:  state.TotalActivityWeight,
		BudgetUsedToday:      state.BudgetUsedToday,
		CurrentRate:          state.CurrentRate,
		DailyBudget:          t.emission.DailyBudget(),
		EffectiveDailyBudget: effectiveBudget,
		RemainingToday:       remaining,
		RolloverBudget:       state.RolloverBudget,
		LastUpdated:          state.LastUpdated,
	}, nil
}

// ValidateClaim validates a reward claim against the auto-scaling rate model.
func (t *AutoScaleRateTracker) ValidateClaim(ctx context.Context, amount int64, trustTierMultiplier float64, rewardType RewardType) error {
	currentRate, err := t.CurrentRate(ctx)
	if err != nil {
		return err
	}

	// Referral rewards are fixed (50 ECHO each), exempt from auto-scaling
	if rewardType == RewardReferral {
		fixedAmount := int64(50_00000000) // 50 ECHO
		if amount > fixedAmount {
			return ErrAutoScaleExceeded
		}
		return nil
	}

	// For messaging rewards, validate against auto-scaled rate
	maxExpected := int64(float64(currentRate) * trustTierMultiplier * 1.01) // 1% rounding tolerance
	if amount > maxExpected {
		return ErrAutoScaleExceeded
	}

	return nil
}

// EffectiveDailyBudget returns the effective daily budget including rollover.
func (t *AutoScaleRateTracker) effectiveDailyBudget(state *AutoScaleState) int64 {
	return t.emission.DailyBudget() + state.RolloverBudget
}

// RemainingToday returns the remaining budget for today.
func (t *AutoScaleRateTracker) RemainingToday(ctx context.Context) (int64, error) {
	state, err := t.GetAutoScaleState(ctx)
	if err != nil {
		return 0, err
	}

	effective := t.effectiveDailyBudget(state)
	remaining := effective - state.BudgetUsedToday
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}

func (t *AutoScaleRateTracker) getOrInitState(ctx context.Context) (*AutoScaleState, error) {
	state, err := t.store.GetState(ctx)
	if err != nil {
		// Initialize with default state
		now := time.Now().UTC()
		state = &AutoScaleState{
			Date:                time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
			TotalActivityWeight: 0.0,
			BudgetUsedToday:     0,
			CurrentRate:         t.calculateInitialRate(),
			RolloverBudget:      0,
			LastUpdated:         now,
		}
		if err := t.store.SetState(ctx, state); err != nil {
			return nil, err
		}
	}
	return state, nil
}

func (t *AutoScaleRateTracker) recalculateRate(activityWeight float64, state *AutoScaleState) int64 {
	if activityWeight <= 0 {
		return t.calculateInitialRate()
	}
	effectiveBudget := float64(t.effectiveDailyBudget(state))
	return int64(effectiveBudget / activityWeight)
}

// Initial rate when no activity has occurred (target: 0.1 ECHO/message).
func (t *AutoScaleRateTracker) calculateInitialRate() int64 {
	return 10_000_000 // 0.1 ECHO with 8 decimal places
}

func (t *AutoScaleRateTracker) maybeRolloverDay(state *AutoScaleState) bool {
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	if today.After(state.Date) {
		// Calculate unused budget rolling forward within the same year.
		effectiveBudget := t.effectiveDailyBudget(state)
		unused := effectiveBudget - state.BudgetUsedToday
		if unused < 0 {
			unused = 0
		}

		state.Date = today
		state.TotalActivityWeight = 0.0
		state.BudgetUsedToday = 0
		state.CurrentRate = t.calculateInitialRate()
		state.RolloverBudget = unused
		state.LastUpdated = now
		return true
	}
	return false
}

var ErrAutoScaleExceeded = errors.New("amount exceeds auto-scaled rate")
