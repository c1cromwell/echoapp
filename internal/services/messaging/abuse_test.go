package messaging

import (
	"fmt"
	"testing"
)

func TestAbuseReport(t *testing.T) {
	tracker := NewAbuseTracker()

	t.Run("single report no suspension", func(t *testing.T) {
		report := &AbuseReport{
			ID:         "report-1",
			ReporterID: "reporter1",
			ReportedID: "baduser",
			MessageID:  "msg-1",
			Type:       AbuseTypeSilentSpam,
		}
		suspension, err := tracker.Report(report)
		if err != nil {
			t.Fatalf("Report failed: %v", err)
		}
		if suspension != nil {
			t.Error("single report should not cause suspension")
		}
	})

	t.Run("three unique reporters triggers suspension", func(t *testing.T) {
		tracker := NewAbuseTracker()
		var lastSuspension *Suspension

		for i := 1; i <= 3; i++ {
			report := &AbuseReport{
				ID:         fmt.Sprintf("report-%d", i),
				ReporterID: fmt.Sprintf("reporter%d", i),
				ReportedID: "spammer",
				MessageID:  fmt.Sprintf("msg-%d", i),
				Type:       AbuseTypeSilentSpam,
			}
			s, err := tracker.Report(report)
			if err != nil {
				t.Fatalf("Report %d failed: %v", i, err)
			}
			if s != nil {
				lastSuspension = s
			}
		}

		if lastSuspension == nil {
			t.Fatal("3 unique reporters should trigger suspension")
		}
		if lastSuspension.Feature != "silent" {
			t.Errorf("feature = %s, want silent", lastSuspension.Feature)
		}
		if lastSuspension.UserID != "spammer" {
			t.Errorf("userID = %s, want spammer", lastSuspension.UserID)
		}
	})

	t.Run("same reporter multiple times doesn't count as unique", func(t *testing.T) {
		tracker := NewAbuseTracker()
		for i := 0; i < 5; i++ {
			report := &AbuseReport{
				ID:         fmt.Sprintf("report-%d", i),
				ReporterID: "same-reporter",
				ReportedID: "user1",
				MessageID:  fmt.Sprintf("msg-%d", i),
				Type:       AbuseTypeSilentSpam,
			}
			suspension, _ := tracker.Report(report)
			if suspension != nil {
				t.Error("same reporter should not trigger suspension")
			}
		}
	})

	t.Run("scheduled abuse tracked separately", func(t *testing.T) {
		tracker := NewAbuseTracker()
		for i := 1; i <= 3; i++ {
			report := &AbuseReport{
				ID:         fmt.Sprintf("report-%d", i),
				ReporterID: fmt.Sprintf("reporter%d", i),
				ReportedID: "scheduler-spammer",
				MessageID:  fmt.Sprintf("msg-%d", i),
				Type:       AbuseTypeScheduledSpam,
			}
			tracker.Report(report)
		}

		if !tracker.IsSuspended("scheduler-spammer", "scheduled") {
			t.Error("should be suspended for scheduled")
		}
		if tracker.IsSuspended("scheduler-spammer", "silent") {
			t.Error("should not be suspended for silent")
		}
	})
}

func TestIsSuspended(t *testing.T) {
	tracker := NewAbuseTracker()

	t.Run("not suspended", func(t *testing.T) {
		if tracker.IsSuspended("cleanuser", "silent") {
			t.Error("clean user should not be suspended")
		}
	})

	t.Run("suspended after reports", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			tracker.Report(&AbuseReport{
				ID:         fmt.Sprintf("r-%d", i),
				ReporterID: fmt.Sprintf("reporter%d", i),
				ReportedID: "baduser",
				Type:       AbuseTypeSilentSpam,
			})
		}
		if !tracker.IsSuspended("baduser", "silent") {
			t.Error("baduser should be suspended")
		}
	})
}

func TestGetActiveSuspension(t *testing.T) {
	tracker := NewAbuseTracker()

	t.Run("no suspension", func(t *testing.T) {
		s := tracker.GetActiveSuspension("user1", "silent")
		if s != nil {
			t.Error("should return nil for no suspension")
		}
	})

	t.Run("active suspension returned", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			tracker.Report(&AbuseReport{
				ID:         fmt.Sprintf("r-%d", i),
				ReporterID: fmt.Sprintf("reporter%d", i),
				ReportedID: "suspended-user",
				Type:       AbuseTypeSilentSpam,
			})
		}
		s := tracker.GetActiveSuspension("suspended-user", "silent")
		if s == nil {
			t.Fatal("should return active suspension")
		}
		if s.UserID != "suspended-user" {
			t.Errorf("userID = %s, want suspended-user", s.UserID)
		}
	})
}

func TestSilentBlocking(t *testing.T) {
	tracker := NewAbuseTracker()

	t.Run("block and check", func(t *testing.T) {
		tracker.BlockSilent("recipient1", "sender1")
		if !tracker.IsSilentBlocked("recipient1", "sender1") {
			t.Error("sender1 should be blocked")
		}
		if tracker.IsSilentBlocked("recipient1", "sender2") {
			t.Error("sender2 should not be blocked")
		}
	})

	t.Run("unblock", func(t *testing.T) {
		tracker.BlockSilent("recipient1", "sender1")
		tracker.UnblockSilent("recipient1", "sender1")
		if tracker.IsSilentBlocked("recipient1", "sender1") {
			t.Error("sender1 should be unblocked after UnblockSilent")
		}
	})

	t.Run("unblock nonexistent", func(t *testing.T) {
		// Should not panic
		tracker.UnblockSilent("nobody", "nobody")
	})

	t.Run("check nonexistent user", func(t *testing.T) {
		if tracker.IsSilentBlocked("nonexistent", "sender1") {
			t.Error("nonexistent user should have no blocks")
		}
	})
}

func TestGetReportCount(t *testing.T) {
	tracker := NewAbuseTracker()

	if count := tracker.GetReportCount("user1"); count != 0 {
		t.Errorf("initial count = %d, want 0", count)
	}

	tracker.Report(&AbuseReport{
		ID: "r1", ReporterID: "a", ReportedID: "user1", Type: AbuseTypeSilentSpam,
	})
	tracker.Report(&AbuseReport{
		ID: "r2", ReporterID: "b", ReportedID: "user1", Type: AbuseTypeHarassment,
	})

	if count := tracker.GetReportCount("user1"); count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestGetReports(t *testing.T) {
	tracker := NewAbuseTracker()

	tracker.Report(&AbuseReport{
		ID: "r1", ReporterID: "a", ReportedID: "user1", Type: AbuseTypeSilentSpam,
	})

	reports := tracker.GetReports("user1")
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].ID != "r1" {
		t.Errorf("report ID = %s, want r1", reports[0].ID)
	}

	// Verify it's a copy (modifying shouldn't affect tracker)
	reports[0].ID = "modified"
	original := tracker.GetReports("user1")
	if original[0].ID != "r1" {
		t.Error("GetReports should return a copy")
	}
}

func TestCalculateTrustPenalty(t *testing.T) {
	tests := []struct {
		name     string
		reports  []AbuseReportType
		expected int
	}{
		{
			name:     "no reports",
			reports:  nil,
			expected: 0,
		},
		{
			name:     "silent spam",
			reports:  []AbuseReportType{AbuseTypeSilentSpam},
			expected: TrustPenaltyPerReport,
		},
		{
			name:     "scheduled spam",
			reports:  []AbuseReportType{AbuseTypeScheduledSpam},
			expected: TrustPenaltyScheduledReport,
		},
		{
			name:     "mixed reports",
			reports:  []AbuseReportType{AbuseTypeSilentSpam, AbuseTypeScheduledSpam, AbuseTypeHarassment},
			expected: TrustPenaltyPerReport + TrustPenaltyScheduledReport + TrustPenaltyPerReport,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewAbuseTracker()
			for i, reportType := range tt.reports {
				tr.Report(&AbuseReport{
					ID:         fmt.Sprintf("r-%d", i),
					ReporterID: fmt.Sprintf("reporter-%d", i),
					ReportedID: "target",
					Type:       reportType,
				})
			}
			penalty := tr.CalculateTrustPenalty("target")
			if penalty != tt.expected {
				t.Errorf("penalty = %d, want %d", penalty, tt.expected)
			}
		})
	}
}

func TestReportTimestamp(t *testing.T) {
	tracker := NewAbuseTracker()
	report := &AbuseReport{
		ID:         "r1",
		ReporterID: "reporter",
		ReportedID: "user1",
		Type:       AbuseTypeSilentSpam,
	}
	tracker.Report(report)

	reports := tracker.GetReports("user1")
	if reports[0].CreatedAt.IsZero() {
		t.Error("report timestamp should be set")
	}
}

func BenchmarkAbuseReport(b *testing.B) {
	tracker := NewAbuseTracker()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.Report(&AbuseReport{
			ID:         fmt.Sprintf("r-%d", i),
			ReporterID: fmt.Sprintf("reporter-%d", i),
			ReportedID: "target",
			Type:       AbuseTypeSilentSpam,
		})
	}
}

func BenchmarkIsSuspended(b *testing.B) {
	tracker := NewAbuseTracker()
	// Pre-populate some data
	for i := 0; i < 100; i++ {
		tracker.Report(&AbuseReport{
			ID:         fmt.Sprintf("r-%d", i),
			ReporterID: fmt.Sprintf("reporter-%d", i%50),
			ReportedID: fmt.Sprintf("user-%d", i%10),
			Type:       AbuseTypeSilentSpam,
		})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.IsSuspended("user-5", "silent")
	}
}
