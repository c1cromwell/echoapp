package messaging

import (
	"testing"
	"time"
)

func TestGetTrustLevelFromScore(t *testing.T) {
	tests := []struct {
		score    int
		expected TrustLevel
	}{
		{0, TrustUnverified},
		{10, TrustUnverified},
		{19, TrustUnverified},
		{20, TrustNewcomer},
		{39, TrustNewcomer},
		{40, TrustMember},
		{59, TrustMember},
		{60, TrustTrusted},
		{79, TrustTrusted},
		{80, TrustVerified},
		{100, TrustVerified},
	}

	for _, tt := range tests {
		result := GetTrustLevelFromScore(tt.score)
		if result != tt.expected {
			t.Errorf("GetTrustLevelFromScore(%d) = %s, want %s", tt.score, result, tt.expected)
		}
	}
}

func TestCheckSilentLimit(t *testing.T) {
	t.Run("within limits", func(t *testing.T) {
		rl := NewRateLimiter()
		err := rl.CheckSilentLimit("user1", "recipient1", TrustMember)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("exceeds daily limit", func(t *testing.T) {
		rl := NewRateLimiter()
		// Unverified users can send 5 silent messages per day
		for i := 0; i < 5; i++ {
			rl.RecordSilent("user1", "recipient1")
		}
		err := rl.CheckSilentLimit("user1", "recipient2", TrustUnverified)
		if err != ErrSilentRateLimitExceeded {
			t.Errorf("expected ErrSilentRateLimitExceeded, got %v", err)
		}
	})

	t.Run("per-recipient hard cap", func(t *testing.T) {
		rl := NewRateLimiter()
		// Even verified users hit the per-recipient cap
		for i := 0; i < PerRecipientDailyLimit; i++ {
			rl.RecordSilent("user1", "recipient1")
		}
		err := rl.CheckSilentLimit("user1", "recipient1", TrustVerified)
		if err != ErrPerRecipientLimitExceeded {
			t.Errorf("expected ErrPerRecipientLimitExceeded, got %v", err)
		}
	})

	t.Run("unlimited for verified", func(t *testing.T) {
		rl := NewRateLimiter()
		// Send many messages to different recipients
		for i := 0; i < 200; i++ {
			rl.RecordSilent("user1", "recipient-other")
		}
		// Should still be allowed to different recipient
		err := rl.CheckSilentLimit("user1", "recipient-new", TrustVerified)
		if err != nil {
			t.Errorf("verified user should have unlimited silent messages, got %v", err)
		}
	})
}

func TestCheckScheduledLimit(t *testing.T) {
	t.Run("within limits", func(t *testing.T) {
		rl := NewRateLimiter()
		scheduledAt := time.Now().Add(12 * time.Hour)
		err := rl.CheckScheduledLimit("user1", TrustMember, scheduledAt, 0)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("exceeds daily limit", func(t *testing.T) {
		rl := NewRateLimiter()
		scheduledAt := time.Now().Add(12 * time.Hour)
		// Unverified can schedule 3 per day
		for i := 0; i < 3; i++ {
			rl.RecordScheduled("user1")
		}
		err := rl.CheckScheduledLimit("user1", TrustUnverified, scheduledAt, 0)
		if err != ErrScheduledLimitExceeded {
			t.Errorf("expected ErrScheduledLimitExceeded, got %v", err)
		}
	})

	t.Run("schedule too far ahead", func(t *testing.T) {
		rl := NewRateLimiter()
		// Unverified can only schedule 1 day ahead
		tooFar := time.Now().Add(48 * time.Hour)
		err := rl.CheckScheduledLimit("user1", TrustUnverified, tooFar, 0)
		if err != ErrScheduledTimeTooFar {
			t.Errorf("expected ErrScheduledTimeTooFar, got %v", err)
		}
	})

	t.Run("total pending limit", func(t *testing.T) {
		rl := NewRateLimiter()
		scheduledAt := time.Now().Add(12 * time.Hour)
		// Unverified max total is 3
		err := rl.CheckScheduledLimit("user1", TrustUnverified, scheduledAt, 3)
		if err != ErrScheduledLimitExceeded {
			t.Errorf("expected ErrScheduledLimitExceeded, got %v", err)
		}
	})

	t.Run("trusted user higher limits", func(t *testing.T) {
		rl := NewRateLimiter()
		// Trusted can schedule 30 days ahead
		farAhead := time.Now().Add(25 * 24 * time.Hour)
		err := rl.CheckScheduledLimit("user1", TrustTrusted, farAhead, 0)
		if err != nil {
			t.Errorf("trusted user should allow 25-day schedule, got %v", err)
		}
	})
}

func TestGetUsage(t *testing.T) {
	rl := NewRateLimiter()

	rl.RecordSilent("user1", "r1")
	rl.RecordSilent("user1", "r2")
	rl.RecordScheduled("user1")

	silent, scheduled := rl.GetUsage("user1")
	if silent != 2 {
		t.Errorf("silent = %d, want 2", silent)
	}
	if scheduled != 1 {
		t.Errorf("scheduled = %d, want 1", scheduled)
	}
}

func TestGetLimits(t *testing.T) {
	rl := NewRateLimiter()

	tests := []struct {
		level           TrustLevel
		expectedSilent  int
		expectedSched   int
		expectedDays    int
	}{
		{TrustUnverified, 5, 3, 1},
		{TrustNewcomer, 20, 10, 7},
		{TrustMember, 50, 25, 14},
		{TrustTrusted, 100, 50, 30},
		{TrustVerified, -1, -1, 30},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			limits := rl.GetLimits(tt.level)
			if limits.SilentMessagesPerDay != tt.expectedSilent {
				t.Errorf("silent/day = %d, want %d", limits.SilentMessagesPerDay, tt.expectedSilent)
			}
			if limits.ScheduledMessagesPerDay != tt.expectedSched {
				t.Errorf("scheduled/day = %d, want %d", limits.ScheduledMessagesPerDay, tt.expectedSched)
			}
			if limits.MaxScheduleAheadDays != tt.expectedDays {
				t.Errorf("max days = %d, want %d", limits.MaxScheduleAheadDays, tt.expectedDays)
			}
		})
	}
}

func TestUnknownTrustLevel(t *testing.T) {
	rl := NewRateLimiter()

	// Unknown level should fall back to unverified limits
	limits := rl.GetLimits(TrustLevel("unknown"))
	unverifiedLimits := rl.GetLimits(TrustUnverified)

	if limits.SilentMessagesPerDay != unverifiedLimits.SilentMessagesPerDay {
		t.Error("unknown trust level should fall back to unverified")
	}
}

func TestCustomLimits(t *testing.T) {
	custom := map[TrustLevel]TrustLimits{
		TrustUnverified: {SilentMessagesPerDay: 1, ScheduledMessagesPerDay: 1, MaxScheduleAheadDays: 1, MaxTotalScheduled: 1},
	}
	rl := NewRateLimiterWithLimits(custom)

	limits := rl.GetLimits(TrustUnverified)
	if limits.SilentMessagesPerDay != 1 {
		t.Errorf("custom limit = %d, want 1", limits.SilentMessagesPerDay)
	}
}

func TestRateLimiterDailyReset(t *testing.T) {
	rl := NewRateLimiter()

	// Record some usage
	rl.RecordSilent("user1", "r1")
	silent, _ := rl.GetUsage("user1")
	if silent != 1 {
		t.Fatalf("silent = %d, want 1", silent)
	}

	// Simulate date change by manipulating the record
	rl.mu.Lock()
	rl.usage["user1"].Date = "2020-01-01" // old date
	rl.mu.Unlock()

	// Should reset on next access
	silent, _ = rl.GetUsage("user1")
	if silent != 0 {
		t.Errorf("after date change, silent = %d, want 0", silent)
	}
}

func BenchmarkCheckSilentLimit(b *testing.B) {
	rl := NewRateLimiter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.CheckSilentLimit("user1", "recipient1", TrustMember)
	}
}
