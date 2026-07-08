package leaderboard

import (
	"context"
	"testing"
	"time"
)

func TestBucketKeyWeeklyMonthly(t *testing.T) {
	ts := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	if got := BucketKey(WindowMonthly, ts); got != "monthly:2026-07" {
		t.Errorf("monthly bucket: got %q", got)
	}
	wk := BucketKey(WindowWeekly, ts)
	iy, iw := ts.ISOWeek()
	want := "weekly:2026-W28"
	if wk != want {
		t.Errorf("weekly bucket: got %q want %q (ISO %d-W%d)", wk, want, iy, iw)
	}
}

func TestRankingOrderAndTiebreak(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()
	at := time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)
	_ = svc.RecordEarning(ctx, "did:b", 3, 300, at)
	_ = svc.RecordEarning(ctx, "did:a", 3, 500, at)
	_ = svc.RecordEarning(ctx, "did:a", 3, 100, at) // did:a total 600
	_ = svc.RecordEarning(ctx, "did:c", 3, 300, at) // ties did:b at 300

	snap, err := svc.Top(ctx, WindowWeekly, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Entries) != 3 {
		t.Fatalf("entries: got %d want 3", len(snap.Entries))
	}
	if snap.Entries[0].DID != "did:a" || snap.Entries[0].Score != 600 || snap.Entries[0].Rank != 1 {
		t.Errorf("rank1: got %+v", snap.Entries[0])
	}
	// did:b and did:c tie at 300; DID asc tiebreak → did:b before did:c.
	if snap.Entries[1].DID != "did:b" || snap.Entries[2].DID != "did:c" {
		t.Errorf("tiebreak: got %q then %q", snap.Entries[1].DID, snap.Entries[2].DID)
	}
}

func TestMinTrustTierFilter(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()
	at := time.Now()
	_ = svc.RecordEarning(ctx, "did:low", 1, 1000, at) // below MinTrustTier
	_ = svc.RecordEarning(ctx, "did:ok", 2, 10, at)

	snap, _ := svc.Top(ctx, WindowWeekly, 10)
	if len(snap.Entries) != 1 || snap.Entries[0].DID != "did:ok" {
		t.Errorf("min-tier filter failed: %+v", snap.Entries)
	}
}

func TestLimitCapsResults(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()
	at := time.Now()
	for i := 0; i < 5; i++ {
		_ = svc.RecordEarning(ctx, string(rune('a'+i))+":did", 2, int64(100-i), at)
	}
	snap, _ := svc.Top(ctx, WindowWeekly, 3)
	if len(snap.Entries) != 3 {
		t.Errorf("limit: got %d want 3", len(snap.Entries))
	}
}

func TestZeroAmountIgnored(t *testing.T) {
	svc := NewService(nil)
	ctx := context.Background()
	_ = svc.RecordEarning(ctx, "did:x", 3, 0, time.Now())
	snap, _ := svc.Top(ctx, WindowWeekly, 10)
	if len(snap.Entries) != 0 {
		t.Errorf("zero amount should not record: %+v", snap.Entries)
	}
}
