// Package streaks implements the daily-activity streak mechanic
// (requirements §Mechanic 4): consecutive active days grant an escalating reward
// multiplier, from 7 days (+10%) up to a 100-day "Century Club" (+100%).
//
// Like the leaderboard, this is a value-free launch (Wave R0) mechanic: the
// multiplier scales ECHO earned into the interim, non-redeemable wallet.
package streaks

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Streak is a user's current daily-activity streak state.
type Streak struct {
	DID         string    `json:"did"`
	CurrentDays int       `json:"current_days"`
	LongestDays int       `json:"longest_days"`
	LastActive  string    `json:"last_active"` // YYYY-MM-DD (UTC), empty if never
	Multiplier  float64   `json:"multiplier"`
	Milestone   string    `json:"milestone"` // label, empty below the first tier
	UpdatedAt   time.Time `json:"updated_at"`
}

// multiplierFor maps a current streak length to its reward multiplier and label.
// Endpoints match the spec (7d → +10%, 100d → +100%); intermediate tiers are a
// design choice consistent with those anchors.
func multiplierFor(days int) (float64, string) {
	switch {
	case days >= 100:
		return 2.00, "Century Club"
	case days >= 60:
		return 1.75, "Diehard"
	case days >= 30:
		return 1.50, "Committed"
	case days >= 14:
		return 1.25, "Fortnight"
	case days >= 7:
		return 1.10, "Week One"
	default:
		return 1.00, ""
	}
}

const dayLayout = "2006-01-02"

func dayKey(t time.Time) string { return t.UTC().Format(dayLayout) }

// isConsecutive reports whether today is exactly one day after prev.
func isConsecutive(prev, today string) bool {
	p, err := time.Parse(dayLayout, prev)
	if err != nil {
		return false
	}
	return dayKey(p.AddDate(0, 0, 1)) == today
}

// Store persists per-user streak state. Safe for concurrent use.
// R0 ships in-memory; a Postgres-backed Store is the persistence follow-up.
type Store interface {
	Get(ctx context.Context, did string) (*Streak, error)
	Put(ctx context.Context, s *Streak) error
}

// ErrNotFound is returned by Store.Get when a user has no streak yet.
var ErrNotFound = errors.New("streak not found")

type memoryStore struct {
	mu   sync.RWMutex
	rows map[string]*Streak
}

func newMemoryStore() *memoryStore { return &memoryStore{rows: make(map[string]*Streak)} }

func (m *memoryStore) Get(_ context.Context, did string) (*Streak, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.rows[did]
	if !ok {
		return nil, ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *memoryStore) Put(_ context.Context, s *Streak) error {
	if s == nil || s.DID == "" {
		return errors.New("did required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *s
	m.rows[s.DID] = &cp
	return nil
}

// Service records daily activity and serves streak state.
type Service struct {
	store Store
}

// NewService constructs a streak service. A nil store defaults to in-memory.
func NewService(store Store) *Service {
	if store == nil {
		store = newMemoryStore()
	}
	return &Service{store: store}
}

// RecordActivity registers activity for did at time at and returns the updated
// streak. Same-day activity is idempotent; a consecutive day increments; any gap
// resets to 1.
func (s *Service) RecordActivity(ctx context.Context, did string, at time.Time) (*Streak, error) {
	if did == "" {
		return nil, errors.New("did required")
	}
	today := dayKey(at)
	st, err := s.store.Get(ctx, did)
	if err != nil || st == nil {
		st = &Streak{DID: did, CurrentDays: 1, LongestDays: 1, LastActive: today}
	} else {
		switch {
		case st.LastActive == today:
			// idempotent within the same UTC day
		case isConsecutive(st.LastActive, today):
			st.CurrentDays++
			st.LastActive = today
		default:
			st.CurrentDays = 1
			st.LastActive = today
		}
	}
	if st.CurrentDays > st.LongestDays {
		st.LongestDays = st.CurrentDays
	}
	st.Multiplier, st.Milestone = multiplierFor(st.CurrentDays)
	st.UpdatedAt = at.UTC()
	if err := s.store.Put(ctx, st); err != nil {
		return nil, err
	}
	cp := *st
	return &cp, nil
}

// Get returns the current streak for did, or a zeroed streak (multiplier 1.0) if
// the user has none yet.
func (s *Service) Get(ctx context.Context, did string) (*Streak, error) {
	st, err := s.store.Get(ctx, did)
	if err != nil || st == nil {
		return &Streak{DID: did, Multiplier: 1.0}, nil
	}
	return st, nil
}
