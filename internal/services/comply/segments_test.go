package comply

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/database"
)

func TestSegmentFromTier(t *testing.T) {
	cases := map[string]string{
		"healthcare": "hipaa",
		"hipaa":      "hipaa",
		"government": "foia",
		"legal":      "law_firm",
		"starter":    "general",
	}
	for tier, want := range cases {
		if got := segmentFromTier(tier); got != want {
			t.Fatalf("tier %q -> %q, want %q", tier, got, want)
		}
	}
}

func TestSegmentDashboard_MemoryDB(t *testing.T) {
	db := database.NewMemoryDB()
	svc := NewService(db, Deps{})
	resp, err := svc.SegmentDashboard(context.Background(), "did:org:acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(resp.Segments))
	}
	for _, seg := range resp.Segments {
		for _, m := range seg.Metrics {
			if m.Key == "" || m.Label == "" {
				t.Fatalf("empty metric in segment %s", seg.Segment)
			}
		}
	}
}
