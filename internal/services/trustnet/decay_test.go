package trustnet

import (
	"testing"
	"time"
)

func TestSetAndGetScore(t *testing.T) {
	svc := NewTrustDecayService()
	svc.SetScore("user1", 75.0)

	if s := svc.GetScore("user1"); s != 75.0 {
		t.Errorf("score = %f, want 75.0", s)
	}

	// Clamp to 0-100
	svc.SetScore("user1", 150)
	if s := svc.GetScore("user1"); s != 100 {
		t.Errorf("score = %f, want 100 (clamped)", s)
	}

	svc.SetScore("user1", -10)
	if s := svc.GetScore("user1"); s != 0 {
		t.Errorf("score = %f, want 0 (clamped)", s)
	}
}

func TestPeakScore(t *testing.T) {
	svc := NewTrustDecayService()
	svc.SetScore("user1", 50)
	svc.SetScore("user1", 80)
	svc.SetScore("user1", 60) // lower, peak stays at 80

	if p := svc.GetPeakScore("user1"); p != 80 {
		t.Errorf("peak = %f, want 80", p)
	}
}

func TestApplyDecay(t *testing.T) {
	t.Run("decay with no activity", func(t *testing.T) {
		svc := NewTrustDecayService()
		svc.SetScore("user1", 80)

		decay := svc.ApplyDecay("user1")
		if decay <= 0 {
			t.Error("decay should be positive without activity")
		}

		score := svc.GetScore("user1")
		if score >= 80 {
			t.Errorf("score should have decreased from 80, got %f", score)
		}
	})

	t.Run("no decay with recent activity", func(t *testing.T) {
		svc := NewTrustDecayService()
		svc.SetScore("user1", 80)
		svc.RecordActivity("user1")

		decay := svc.ApplyDecay("user1")
		if decay != 0 {
			t.Errorf("decay = %f, want 0 with recent activity", decay)
		}
	})

	t.Run("decay respects floor", func(t *testing.T) {
		svc := NewTrustDecayService()
		svc.SetScore("user1", 80) // peak = 80, floor = 48

		// Simulate extended inactivity
		svc.mu.Lock()
		svc.lastActivity["user1"] = time.Now().Add(-365 * 24 * time.Hour)
		svc.mu.Unlock()

		svc.ApplyDecay("user1")
		score := svc.GetScore("user1")
		floor := svc.CalculateDecayFloor("user1")

		if score < floor {
			t.Errorf("score %f fell below floor %f", score, floor)
		}
	})

	t.Run("zero score no decay", func(t *testing.T) {
		svc := NewTrustDecayService()
		decay := svc.ApplyDecay("unknown")
		if decay != 0 {
			t.Errorf("decay = %f, want 0 for zero score", decay)
		}
	})
}

func TestApplyDecayAll(t *testing.T) {
	svc := NewTrustDecayService()
	svc.SetScore("user1", 80)
	svc.SetScore("user2", 60)
	svc.RecordActivity("user2") // user2 is active, should not decay

	results := svc.ApplyDecayAll()
	if _, ok := results["user1"]; !ok {
		t.Error("user1 should have decayed")
	}
	if _, ok := results["user2"]; ok {
		t.Error("user2 should not have decayed (active)")
	}
}

func TestRecordSnapshot(t *testing.T) {
	svc := NewTrustDecayService()
	svc.SetScore("user1", 75)

	components := ScoreComponents{
		Verification: 20,
		Network:      15,
		Behavior:     22,
		Transactions: 18,
	}
	svc.RecordSnapshot("user1", components)

	history := svc.GetHistory("user1")
	if len(history) != 1 {
		t.Fatalf("history len = %d, want 1", len(history))
	}
	if history[0].Score != 75 {
		t.Errorf("snapshot score = %f, want 75", history[0].Score)
	}
	if history[0].Components.Total() != 75 {
		t.Errorf("components total = %f, want 75", history[0].Components.Total())
	}
}

func TestGetHistorySince(t *testing.T) {
	svc := NewTrustDecayService()
	svc.SetScore("user1", 50)

	// Record snapshots at different times
	svc.RecordSnapshot("user1", ScoreComponents{Verification: 50})

	svc.SetScore("user1", 60)
	svc.RecordSnapshot("user1", ScoreComponents{Verification: 60})

	since := time.Now().Add(-1 * time.Second)
	history := svc.GetHistorySince("user1", since)
	if len(history) != 2 {
		t.Errorf("history since = %d, want 2", len(history))
	}

	// Future date should return nothing
	future := time.Now().Add(1 * time.Hour)
	history = svc.GetHistorySince("user1", future)
	if len(history) != 0 {
		t.Errorf("future history = %d, want 0", len(history))
	}
}

func TestGetScoreChange(t *testing.T) {
	svc := NewTrustDecayService()
	svc.SetScore("user1", 50)
	svc.RecordSnapshot("user1", ScoreComponents{Verification: 50})

	svc.SetScore("user1", 70)

	change := svc.GetScoreChange("user1", time.Now().Add(-1*time.Minute))
	if change != 20 {
		t.Errorf("change = %f, want 20", change)
	}
}

func TestCalculateDecayFloor(t *testing.T) {
	svc := NewTrustDecayService()
	svc.SetScore("user1", 100)

	floor := svc.CalculateDecayFloor("user1")
	if floor != 60 {
		t.Errorf("floor = %f, want 60 (60%% of peak 100)", floor)
	}
}

func TestHistoryPruning(t *testing.T) {
	svc := NewTrustDecayService()
	svc.SetScore("user1", 50)

	// Add an old snapshot manually
	svc.mu.Lock()
	svc.history["user1"] = append(svc.history["user1"], TrustScoreSnapshot{
		UserDID:    "user1",
		Score:      30,
		RecordedAt: time.Now().Add(-100 * 24 * time.Hour), // 100 days ago
	})
	svc.mu.Unlock()

	// Recording a new snapshot should prune the old one
	svc.RecordSnapshot("user1", ScoreComponents{Verification: 50})

	history := svc.GetHistory("user1")
	for _, h := range history {
		age := time.Since(h.RecordedAt)
		if age > time.Duration(HistoryRetentionDays)*24*time.Hour {
			t.Error("old snapshot should have been pruned")
		}
	}
}

func BenchmarkApplyDecay(b *testing.B) {
	svc := NewTrustDecayService()
	for i := 0; i < 1000; i++ {
		svc.SetScore(contactDID(i), 80)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.ApplyDecayAll()
	}
}
