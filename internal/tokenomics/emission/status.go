// Package emission exposes WO-206 emission analytics from the community rewards curve.
package emission

import (
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/rewards"
	"github.com/thechadcromwell/echoapp/internal/tokenomics/genesis"
)

// Status is returned by GET /tokens/emission/status.
type Status struct {
	CurrentYear       int     `json:"current_year"`
	AnnualCap         int64   `json:"annual_cap"`
	DistributedToDate int64   `json:"distributed_to_date"`
	RemainingBudget   int64   `json:"remaining_budget"`
	PercentConsumed   float64 `json:"percent_consumed"`
	GenesisDate       string  `json:"genesis_date"`
	CachedAt          string  `json:"cached_at"`
}

// Tracker caches emission status with a short TTL (WO-206: 5s).
type Tracker struct {
	mu       sync.RWMutex
	schedule *rewards.EmissionSchedule
	claimed  func() int64
	ttl      time.Duration
	cached   *Status
	cachedAt time.Time
}

// NewTracker creates an emission status tracker.
func NewTracker(genesisDate time.Time, claimedToday func() int64) *Tracker {
	return &Tracker{
		schedule: rewards.NewEmissionSchedule(genesisDate),
		claimed:  claimedToday,
		ttl:      5 * time.Second,
	}
}

// Get returns cached emission status.
func (t *Tracker) Get() Status {
	t.mu.RLock()
	if t.cached != nil && time.Since(t.cachedAt) < t.ttl {
		st := *t.cached
		t.mu.RUnlock()
		return st
	}
	t.mu.RUnlock()

	st := t.compute()
	t.mu.Lock()
	t.cached = &st
	t.cachedAt = time.Now()
	t.mu.Unlock()
	return st
}

func (t *Tracker) compute() Status {
	year := t.schedule.CurrentYear()
	if year > 10 {
		year = 10
	}
	var annualCap int64
	if year >= 1 && year <= len(genesis.YearlyEmissionCaps) {
		annualCap = genesis.YearlyEmissionCaps[year-1]
	}

	distributedYear := t.distributedThisYear()
	remaining := annualCap - distributedYear
	if remaining < 0 {
		remaining = 0
	}
	var pct float64
	if annualCap > 0 {
		pct = float64(distributedYear) / float64(annualCap) * 100
	}

	var distributedTotal int64
	for y := 1; y < year; y++ {
		distributedTotal += genesis.YearlyEmissionCaps[y-1]
	}
	distributedTotal += distributedYear

	return Status{
		CurrentYear:       year,
		AnnualCap:         annualCap,
		DistributedToDate: distributedTotal,
		RemainingBudget:   remaining,
		PercentConsumed:   pct,
		GenesisDate:       t.schedule.GenesisDate.Format(time.RFC3339),
		CachedAt:          time.Now().UTC().Format(time.RFC3339),
	}
}

func (t *Tracker) distributedThisYear() int64 {
	if t.claimed != nil {
		return t.claimed()
	}
	return 0
}
