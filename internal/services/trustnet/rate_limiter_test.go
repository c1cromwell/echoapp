package trustnet

import (
	"testing"
	"time"
)

func TestRateLimiterCheck(t *testing.T) {
	t.Run("allows within limit", func(t *testing.T) {
		rl := NewTrustRateLimiter()
		err := rl.Check("user1", OpEndorsement)
		if err != nil {
			t.Fatalf("should allow: %v", err)
		}
	})

	t.Run("blocks when limit reached", func(t *testing.T) {
		rl := NewTrustRateLimiter()
		for i := 0; i < TrustOpLimits[OpEndorsement].Daily; i++ {
			rl.Record("user1", OpEndorsement)
		}
		err := rl.Check("user1", OpEndorsement)
		if err != ErrTrustOpRateLimited {
			t.Errorf("expected ErrTrustOpRateLimited, got %v", err)
		}
	})

	t.Run("unknown operation allowed", func(t *testing.T) {
		rl := NewTrustRateLimiter()
		err := rl.Check("user1", TrustOperation("unknown"))
		if err != nil {
			t.Errorf("unknown op should be allowed: %v", err)
		}
	})
}

func TestRateLimiterCooldown(t *testing.T) {
	t.Run("dispute cooldown", func(t *testing.T) {
		rl := NewTrustRateLimiter()
		rl.Record("user1", OpDispute)

		err := rl.Check("user1", OpDispute)
		if err != ErrTrustOpRateLimited {
			t.Errorf("expected rate limited during cooldown, got %v", err)
		}
	})

	t.Run("cooldown expires", func(t *testing.T) {
		rl := NewTrustRateLimiter()
		rl.Record("user1", OpDispute)

		// Manually expire cooldown
		rl.mu.Lock()
		rl.usage["user1"][OpDispute].LastUsedAt = time.Now().Add(-91 * 24 * time.Hour)
		rl.usage["user1"][OpDispute].Date = time.Now().Add(-91 * 24 * time.Hour).Format("2006-01-02")
		rl.mu.Unlock()

		err := rl.Check("user1", OpDispute)
		if err != nil {
			t.Errorf("should allow after cooldown: %v", err)
		}
	})
}

func TestRateLimiterWithTarget(t *testing.T) {
	t.Run("per target limit", func(t *testing.T) {
		rl := NewTrustRateLimiter()
		rl.RecordWithTarget("user1", OpReport, "target1")

		// Can still report other targets
		err := rl.CheckWithTarget("user1", OpReport, "target2")
		if err != nil {
			t.Errorf("should allow different target: %v", err)
		}

		// Cannot re-report same target
		err = rl.CheckWithTarget("user1", OpReport, "target1")
		if err != ErrTrustOpRateLimited {
			t.Errorf("expected rate limited for same target, got %v", err)
		}
	})

	t.Run("daily limit with targets", func(t *testing.T) {
		rl := NewTrustRateLimiter()
		for i := 0; i < TrustOpLimits[OpReport].Daily; i++ {
			rl.RecordWithTarget("user1", OpReport, contactDID(i))
		}

		err := rl.CheckWithTarget("user1", OpReport, "newTarget")
		if err != ErrTrustOpRateLimited {
			t.Errorf("expected rate limited at daily cap, got %v", err)
		}
	})
}

func TestGetRemaining(t *testing.T) {
	rl := NewTrustRateLimiter()

	if r := rl.GetRemaining("user1", OpEndorsement); r != TrustOpLimits[OpEndorsement].Daily {
		t.Errorf("remaining = %d, want %d", r, TrustOpLimits[OpEndorsement].Daily)
	}

	rl.Record("user1", OpEndorsement)
	rl.Record("user1", OpEndorsement)

	if r := rl.GetRemaining("user1", OpEndorsement); r != TrustOpLimits[OpEndorsement].Daily-2 {
		t.Errorf("remaining = %d, want %d", r, TrustOpLimits[OpEndorsement].Daily-2)
	}
}

func TestGetCooldownRemaining(t *testing.T) {
	rl := NewTrustRateLimiter()

	// No cooldown for endorsements
	if d := rl.GetCooldownRemaining("user1", OpEndorsement); d != 0 {
		t.Errorf("cooldown = %v, want 0", d)
	}

	// Dispute has 90-day cooldown
	rl.Record("user1", OpDispute)
	cd := rl.GetCooldownRemaining("user1", OpDispute)
	if cd <= 0 {
		t.Error("dispute should have cooldown remaining")
	}
}

func TestHasTargeted(t *testing.T) {
	rl := NewTrustRateLimiter()
	rl.RecordWithTarget("user1", OpReport, "target1")

	if !rl.HasTargeted("user1", OpReport, "target1") {
		t.Error("should have targeted target1")
	}
	if rl.HasTargeted("user1", OpReport, "target2") {
		t.Error("should not have targeted target2")
	}
}

func TestDifferentOperationsIndependent(t *testing.T) {
	rl := NewTrustRateLimiter()

	// Use up all endorsements
	for i := 0; i < TrustOpLimits[OpEndorsement].Daily; i++ {
		rl.Record("user1", OpEndorsement)
	}

	// Promotions should still be available
	err := rl.Check("user1", OpPromotion)
	if err != nil {
		t.Errorf("promotions should be independent: %v", err)
	}
}

func TestDifferentUsersIndependent(t *testing.T) {
	rl := NewTrustRateLimiter()

	for i := 0; i < TrustOpLimits[OpEndorsement].Daily; i++ {
		rl.Record("user1", OpEndorsement)
	}

	err := rl.Check("user2", OpEndorsement)
	if err != nil {
		t.Errorf("user2 should be independent: %v", err)
	}
}

func BenchmarkRateLimiterCheck(b *testing.B) {
	rl := NewTrustRateLimiter()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Check("user1", OpEndorsement)
	}
}
