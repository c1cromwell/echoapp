package emission

import (
	"testing"
	"time"
)

func TestTracker_Get(t *testing.T) {
	genesis := time.Now().AddDate(0, -1, 0)
	tr := NewTracker(genesis, func() int64 { return 1_000_000 })
	st := tr.Get()
	if st.CurrentYear != 1 {
		t.Errorf("expected year 1, got %d", st.CurrentYear)
	}
	if st.AnnualCap <= 0 {
		t.Error("expected positive annual cap")
	}
	if st.CachedAt == "" {
		t.Error("expected cached_at")
	}
	// Second call should hit cache (same CachedAt within TTL).
	st2 := tr.Get()
	if st2.CachedAt != st.CachedAt {
		t.Error("expected cache hit")
	}
}
