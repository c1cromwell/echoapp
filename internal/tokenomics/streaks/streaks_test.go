package streaks

import (
	"context"
	"testing"
	"time"
)

func day(y, m, d int) time.Time { return time.Date(y, time.Month(m), d, 12, 0, 0, 0, time.UTC) }

func TestConsecutiveDaysIncrement(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()
	_, _ = svc.RecordActivity(ctx, "did:a", day(2026, 7, 1))
	_, _ = svc.RecordActivity(ctx, "did:a", day(2026, 7, 2))
	st, _ := svc.RecordActivity(ctx, "did:a", day(2026, 7, 3))
	if st.CurrentDays != 3 {
		t.Errorf("current: got %d want 3", st.CurrentDays)
	}
}

func TestSameDayIdempotent(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()
	_, _ = svc.RecordActivity(ctx, "did:a", day(2026, 7, 1))
	st, _ := svc.RecordActivity(ctx, "did:a", time.Date(2026, 7, 1, 23, 0, 0, 0, time.UTC))
	if st.CurrentDays != 1 {
		t.Errorf("same-day should not increment: got %d", st.CurrentDays)
	}
}

func TestGapResets(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()
	_, _ = svc.RecordActivity(ctx, "did:a", day(2026, 7, 1))
	_, _ = svc.RecordActivity(ctx, "did:a", day(2026, 7, 2))
	st, _ := svc.RecordActivity(ctx, "did:a", day(2026, 7, 5)) // 2-day gap
	if st.CurrentDays != 1 {
		t.Errorf("gap should reset: got %d want 1", st.CurrentDays)
	}
	if st.LongestDays != 2 {
		t.Errorf("longest should persist: got %d want 2", st.LongestDays)
	}
}

func TestMilestoneMultipliers(t *testing.T) {
	cases := []struct {
		days int
		mult float64
		name string
	}{
		{1, 1.00, ""},
		{7, 1.10, "Week One"},
		{14, 1.25, "Fortnight"},
		{30, 1.50, "Committed"},
		{60, 1.75, "Diehard"},
		{100, 2.00, "Century Club"},
	}
	for _, c := range cases {
		m, label := multiplierFor(c.days)
		if m != c.mult || label != c.name {
			t.Errorf("day %d: got (%.2f,%q) want (%.2f,%q)", c.days, m, label, c.mult, c.name)
		}
	}
}

func TestGetUnknownReturnsBaseline(t *testing.T) {
	svc := NewService(nil)
	st, _ := svc.Get(context.Background(), "did:none")
	if st.CurrentDays != 0 || st.Multiplier != 1.0 {
		t.Errorf("unknown baseline: got %+v", st)
	}
}
