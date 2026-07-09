package leaderboard

import (
	"context"
	"fmt"
	"time"
)

// DefaultLimit caps how many ranked entries an endpoint returns by default.
const DefaultLimit = 100

// Service records earnings against rolling windows and serves ranked snapshots.
type Service struct {
	store Store
}

// NewService constructs a leaderboard service. A nil store defaults to in-memory.
func NewService(store Store) *Service {
	if store == nil {
		store = newMemoryStore()
	}
	return &Service{store: store}
}

// BucketKey returns the storage key for a window at time t.
//   - weekly:  "weekly:<ISOyear>-W<ISOweek>"
//   - monthly: "monthly:<year>-<month>"
func BucketKey(w Window, t time.Time) string {
	t = t.UTC()
	switch w {
	case WindowMonthly:
		return fmt.Sprintf("monthly:%04d-%02d", t.Year(), int(t.Month()))
	default: // weekly
		isoYear, isoWeek := t.ISOWeek()
		return fmt.Sprintf("weekly:%04d-W%02d", isoYear, isoWeek)
	}
}

// RecordEarning attributes an ECHO earning to both the weekly and monthly
// leaderboards for the current time. Users below MinTrustTier are still recorded
// (so a later tier bump surfaces them) but filtered at read time. Zero/negative
// amounts are ignored.
func (s *Service) RecordEarning(ctx context.Context, did string, trustTier int, amountDatum int64, at time.Time) error {
	if did == "" || amountDatum <= 0 {
		return nil
	}
	for _, w := range []Window{WindowWeekly, WindowMonthly} {
		if err := s.store.Add(ctx, BucketKey(w, at), did, trustTier, amountDatum); err != nil {
			return err
		}
	}
	return nil
}

// Top returns the current ranked snapshot for a window. minTier defaults to
// MinTrustTier; limit defaults to DefaultLimit.
func (s *Service) Top(ctx context.Context, w Window, limit int) (*Snapshot, error) {
	if !w.Valid() {
		w = WindowWeekly
	}
	if limit <= 0 || limit > DefaultLimit {
		limit = DefaultLimit
	}
	key := BucketKey(w, time.Now())
	entries, err := s.store.Top(ctx, key, limit, MinTrustTier)
	if err != nil {
		return nil, err
	}
	return &Snapshot{
		Window:    w,
		BucketKey: key,
		Entries:   entries,
		UpdatedAt: time.Now().UTC(),
	}, nil
}
