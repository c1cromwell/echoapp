package trust

import (
	"context"
	"testing"

	"github.com/thechadcromwell/echoapp/internal/auth"
)

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.scorer == nil {
		t.Fatal("expected non-nil scorer")
	}
	if svc.cache == nil {
		t.Fatal("expected non-nil cache")
	}
}

func TestGetScore(t *testing.T) {
	svc := NewService()
	record, err := svc.GetScore(context.Background(), "did:echo:testuser1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record == nil {
		t.Fatal("expected non-nil record")
	}
	if record.DID != "did:echo:testuser1" {
		t.Errorf("expected DID did:echo:testuser1, got %s", record.DID)
	}
	if record.Score < 0 || record.Score > 100 {
		t.Errorf("score %d out of range [0, 100]", record.Score)
	}
	if record.Tier < 1 || record.Tier > 5 {
		t.Errorf("tier %d out of range [1, 5]", record.Tier)
	}
	if record.Multiplier < 1.0 {
		t.Errorf("multiplier %f should be >= 1.0", record.Multiplier)
	}
}

func TestGetScoreCaching(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	did := "did:echo:cachetest"

	record1, err := svc.GetScore(ctx, did)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	record2, err := svc.GetScore(ctx, did)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record1.UpdatedAt != record2.UpdatedAt {
		t.Error("expected cached result with same UpdatedAt")
	}
}

func TestGetScoreBatch(t *testing.T) {
	svc := NewService()
	dids := []string{"did:echo:a", "did:echo:b", "did:echo:c"}
	records, err := svc.GetScoreBatch(context.Background(), dids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}
	for i, r := range records {
		if r.DID != dids[i] {
			t.Errorf("record %d: expected DID %s, got %s", i, dids[i], r.DID)
		}
	}
}

func TestSubmitReport(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	did := "did:echo:target"

	// Prime the cache
	_, err := svc.GetScore(ctx, did)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	report := Report{
		ReporterDID: "did:echo:reporter",
		TargetDID:   did,
		ReportType:  "spam",
		Reason:      "unsolicited messages",
	}
	if err := svc.SubmitReport(report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Cache should be invalidated — fetch fresh
	record, err := svc.GetScore(ctx, did)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if record == nil {
		t.Fatal("expected non-nil record after report")
	}
}

func TestGetReports(t *testing.T) {
	svc := NewService()
	target := "did:echo:target2"

	svc.SubmitReport(Report{ReporterDID: "did:echo:r1", TargetDID: target, ReportType: "spam"})
	svc.SubmitReport(Report{ReporterDID: "did:echo:r2", TargetDID: target, ReportType: "fraud"})
	svc.SubmitReport(Report{ReporterDID: "did:echo:r3", TargetDID: "did:echo:other", ReportType: "spam"})

	reports := svc.GetReports(target)
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
}

func TestInvalidateCache(t *testing.T) {
	svc := NewService()
	ctx := context.Background()
	did := "did:echo:invalidate"

	record1, _ := svc.GetScore(ctx, did)
	svc.InvalidateCache(did)
	record2, _ := svc.GetScore(ctx, did)

	if record1.UpdatedAt == record2.UpdatedAt {
		t.Error("expected different UpdatedAt after cache invalidation")
	}
}

func TestTierFromLevel(t *testing.T) {
	tests := []struct {
		level auth.TrustScoreLevel
		tier  int
	}{
		{auth.TrustLevelNewcomer, 1},
		{auth.TrustLevelBasic, 2},
		{auth.TrustLevelTrusted, 3},
		{auth.TrustLevelVerified, 4},
		{auth.TrustLevelElite, 5},
	}
	for _, tt := range tests {
		if got := tierFromLevel(tt.level); got != tt.tier {
			t.Errorf("tierFromLevel(%v) = %d, want %d", tt.level, got, tt.tier)
		}
	}
}
