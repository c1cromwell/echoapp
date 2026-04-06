package trustnet

import (
	"testing"
	"time"
)

func TestEndorse(t *testing.T) {
	t.Run("valid endorsement", func(t *testing.T) {
		svc := NewEndorsementService()
		e, err := svc.Endorse("endorser1", "endorsee1", EndorseReliable, 75, "Great person")
		if err != nil {
			t.Fatalf("Endorse failed: %v", err)
		}
		if e.EndorserDID != "endorser1" || e.EndorseeDID != "endorsee1" {
			t.Error("incorrect DIDs")
		}
		if e.Category != EndorseReliable {
			t.Errorf("category = %s, want reliable", e.Category)
		}
		if !e.Active {
			t.Error("should be active")
		}
		if e.Weight <= 0 || e.Weight > 1.0 {
			t.Errorf("weight = %f, should be between 0 and 1", e.Weight)
		}
	})

	t.Run("cannot self endorse", func(t *testing.T) {
		svc := NewEndorsementService()
		_, err := svc.Endorse("user1", "user1", EndorseReliable, 80, "")
		if err != ErrEndorsementSelfEndorse {
			t.Errorf("expected ErrEndorsementSelfEndorse, got %v", err)
		}
	})

	t.Run("insufficient trust", func(t *testing.T) {
		svc := NewEndorsementService()
		_, err := svc.Endorse("user1", "user2", EndorseReliable, 50, "")
		if err != ErrEndorsementInsufficientTrust {
			t.Errorf("expected ErrEndorsementInsufficientTrust, got %v", err)
		}
	})

	t.Run("duplicate endorsement", func(t *testing.T) {
		svc := NewEndorsementService()
		svc.Endorse("user1", "user2", EndorseReliable, 80, "")
		_, err := svc.Endorse("user1", "user2", EndorseReliable, 80, "")
		if err != ErrEndorsementDuplicate {
			t.Errorf("expected ErrEndorsementDuplicate, got %v", err)
		}
	})

	t.Run("different categories allowed", func(t *testing.T) {
		svc := NewEndorsementService()
		svc.Endorse("user1", "user2", EndorseReliable, 80, "")
		_, err := svc.Endorse("user1", "user2", EndorseHelpful, 80, "")
		if err != nil {
			t.Fatalf("different category should be allowed: %v", err)
		}
	})

	t.Run("daily rate limit", func(t *testing.T) {
		svc := NewEndorsementService()
		for i := 0; i < MaxEndorsementsPerDay; i++ {
			_, err := svc.Endorse("user1", contactDID(i), EndorseReliable, 80, "")
			if err != nil {
				t.Fatalf("endorsement %d failed: %v", i, err)
			}
		}
		_, err := svc.Endorse("user1", "overflow", EndorseReliable, 80, "")
		if err != ErrEndorsementRateLimited {
			t.Errorf("expected ErrEndorsementRateLimited, got %v", err)
		}
	})

	t.Run("weight scales with trust score", func(t *testing.T) {
		svc := NewEndorsementService()
		e60, _ := svc.Endorse("low", "user2", EndorseReliable, 60, "")
		e100, _ := svc.Endorse("high", "user3", EndorseReliable, 100, "")

		if e100.Weight <= e60.Weight {
			t.Errorf("higher trust should give higher weight: 100=%f, 60=%f", e100.Weight, e60.Weight)
		}
		if e60.Weight < 0.6 {
			t.Errorf("min weight should be 0.6, got %f", e60.Weight)
		}
		if e100.Weight > 1.0 {
			t.Errorf("max weight should be 1.0, got %f", e100.Weight)
		}
	})
}

func TestRevoke(t *testing.T) {
	t.Run("revoke own endorsement", func(t *testing.T) {
		svc := NewEndorsementService()
		e, _ := svc.Endorse("user1", "user2", EndorseReliable, 80, "")

		err := svc.Revoke(e.ID, "user1")
		if err != nil {
			t.Fatalf("revoke failed: %v", err)
		}

		got, _ := svc.GetEndorsement(e.ID)
		if got.Active {
			t.Error("should be inactive after revoke")
		}
		if got.RevokedAt == nil {
			t.Error("RevokedAt should be set")
		}
	})

	t.Run("cannot revoke others endorsement", func(t *testing.T) {
		svc := NewEndorsementService()
		e, _ := svc.Endorse("user1", "user2", EndorseReliable, 80, "")

		err := svc.Revoke(e.ID, "user3")
		if err != ErrEndorsementNotOwner {
			t.Errorf("expected ErrEndorsementNotOwner, got %v", err)
		}
	})

	t.Run("revocation cooldown", func(t *testing.T) {
		svc := NewEndorsementService()
		e, _ := svc.Endorse("user1", "user2", EndorseReliable, 80, "")
		svc.Revoke(e.ID, "user1")

		// Try to re-endorse immediately — should be blocked by cooldown
		_, err := svc.Endorse("user1", "user2", EndorseReliable, 80, "")
		if err != ErrEndorsementCooldown {
			t.Errorf("expected ErrEndorsementCooldown, got %v", err)
		}
	})

	t.Run("revocation cooldown expires", func(t *testing.T) {
		svc := NewEndorsementService()
		e, _ := svc.Endorse("user1", "user2", EndorseReliable, 80, "")
		svc.Revoke(e.ID, "user1")

		// Manually expire the cooldown
		svc.mu.Lock()
		svc.revocations[0].RevokedAt = time.Now().Add(-RevocationCooldown - time.Hour)
		svc.mu.Unlock()

		_, err := svc.Endorse("user1", "user2", EndorseReliable, 80, "")
		if err != nil {
			t.Fatalf("re-endorse after cooldown should work: %v", err)
		}
	})
}

func TestGetEndorsements(t *testing.T) {
	svc := NewEndorsementService()
	svc.Endorse("a", "target", EndorseReliable, 80, "")
	svc.Endorse("b", "target", EndorseHelpful, 80, "")
	svc.Endorse("c", "target", EndorseProfessional, 80, "")

	// Revoke one
	by := svc.GetEndorsementsBy("a")
	svc.Revoke(by[0].ID, "a")

	forTarget := svc.GetEndorsementsFor("target")
	if len(forTarget) != 2 {
		t.Errorf("active endorsements for target = %d, want 2", len(forTarget))
	}

	byCategory := svc.GetEndorsementsByCategory("target", EndorseHelpful)
	if len(byCategory) != 1 {
		t.Errorf("helpful endorsements = %d, want 1", len(byCategory))
	}
}

func TestCountActiveEndorsements(t *testing.T) {
	svc := NewEndorsementService()
	svc.Endorse("a", "target", EndorseReliable, 80, "")
	svc.Endorse("b", "target", EndorseHelpful, 80, "")

	if n := svc.CountActiveEndorsements("target"); n != 2 {
		t.Errorf("count = %d, want 2", n)
	}
}

func TestCalculateNetworkScore(t *testing.T) {
	svc := NewEndorsementService()

	// No endorsements
	if s := svc.CalculateNetworkScore("nobody"); s != 0 {
		t.Errorf("score = %f, want 0", s)
	}

	// Single endorsement
	svc.Endorse("a", "target", EndorseReliable, 80, "")
	s := svc.CalculateNetworkScore("target")
	if s <= 0 {
		t.Error("score should be positive with endorsement")
	}
	if s > 25 {
		t.Errorf("score = %f, should not exceed 25", s)
	}
}

func TestDailyEndorsementsRemaining(t *testing.T) {
	svc := NewEndorsementService()

	if r := svc.DailyEndorsementsRemaining("user1"); r != MaxEndorsementsPerDay {
		t.Errorf("remaining = %d, want %d", r, MaxEndorsementsPerDay)
	}

	svc.Endorse("user1", "target", EndorseReliable, 80, "")
	if r := svc.DailyEndorsementsRemaining("user1"); r != MaxEndorsementsPerDay-1 {
		t.Errorf("remaining = %d, want %d", r, MaxEndorsementsPerDay-1)
	}
}

func BenchmarkEndorse(b *testing.B) {
	svc := NewEndorsementService()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.Endorse(contactDID(i), "target", EndorseReliable, 80, "")
	}
}
