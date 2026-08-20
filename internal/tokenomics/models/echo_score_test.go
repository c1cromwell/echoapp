package models

import "testing"

func TestSnapshotFromScoreUnlockLadder(t *testing.T) {
	cases := []struct {
		score      int
		tier       int
		level      string
		mult       float64
		unlockTier int
		needed     int
		hasUnlock  bool
	}{
		{10, 1, "unverified", 1.0, 2, 10, true},
		{20, 2, "newcomer", 1.2, 3, 20, true},
		{50, 3, "member", 1.5, 4, 10, true},
		{70, 4, "trusted", 2.0, 5, 10, true},
		{90, 5, "verified", 3.0, 0, 0, false},
		{100, 5, "verified", 3.0, 0, 0, false},
	}
	for _, c := range cases {
		s := SnapshotFromScore(c.score)
		if s.Tier != c.tier || s.Level != c.level || s.Multiplier != c.mult {
			t.Errorf("score %d: got tier=%d level=%s mult=%.1f", c.score, s.Tier, s.Level, s.Multiplier)
		}
		if c.hasUnlock {
			if s.NextUnlock == nil || s.NextUnlock.Tier != c.unlockTier || s.NextUnlock.PointsNeeded != c.needed {
				t.Errorf("score %d unlock: %+v want tier %d needed %d", c.score, s.NextUnlock, c.unlockTier, c.needed)
			}
		} else if s.NextUnlock != nil {
			t.Errorf("score %d: expected no next unlock, got %+v", c.score, s.NextUnlock)
		}
	}
}

func TestSnapshotFromTierMidpoints(t *testing.T) {
	s := SnapshotFromTier(1)
	if s.Score != 10 || s.Tier != 1 {
		t.Errorf("tier 1 midpoint: %+v", s)
	}
	s = SnapshotFromTier(5)
	if s.Score != 90 || s.Tier != 5 {
		t.Errorf("tier 5 midpoint: %+v", s)
	}
}

func TestSnapshotClampsScore(t *testing.T) {
	if SnapshotFromScore(-4).Score != 0 {
		t.Error("negative score should clamp to 0")
	}
	if SnapshotFromScore(140).Score != 100 {
		t.Error("oversize score should clamp to 100")
	}
}
