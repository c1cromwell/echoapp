package packs

import (
	"context"
	"testing"
	"time"
)

func TestPreviewDeterministicByStreak(t *testing.T) {
	s := NewService()
	at := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	p := s.Preview("did:a", 3, at)
	if p.Opened || p.Label != "Weekly pack" || len(p.Items) != 1 {
		t.Fatalf("starter pack: %+v", p)
	}
	week := p.WeekKey
	p7 := s.Preview("did:a", 7, at)
	if p7.WeekKey != week || p7.Label != "Week One pack" || len(p7.Items) != 2 {
		t.Fatalf("week-one pack: %+v", p7)
	}
	p30 := s.Preview("did:a", 30, at)
	if p30.Label != "Committed pack" {
		t.Fatalf("committed pack: %+v", p30)
	}
}

func TestOpenIdempotent(t *testing.T) {
	s := NewService()
	at := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	ctx := context.Background()
	first := s.Open(ctx, "did:a", 8, at)
	if !first.Opened {
		t.Fatal("expected opened")
	}
	second := s.Open(ctx, "did:a", 8, at)
	if second.OpenedAt == nil || first.OpenedAt == nil || !second.OpenedAt.Equal(*first.OpenedAt) {
		t.Fatalf("idempotent open should keep first timestamp: %+v %+v", first, second)
	}
	other := s.Preview("did:b", 8, at)
	if other.Opened {
		t.Fatal("other DID should not inherit opened state")
	}
}
