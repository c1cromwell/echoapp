package trustnet

import (
	"testing"
	"time"
)

func TestFileDispute(t *testing.T) {
	t.Run("valid dispute", func(t *testing.T) {
		svc := NewDisputeService()
		d, err := svc.FileDispute("user1", "user2", DisputeFalseReport, "Evidence here", 10.0)
		if err != nil {
			t.Fatalf("FileDispute failed: %v", err)
		}
		if d.Status != DisputeOpen {
			t.Errorf("status = %s, want open", d.Status)
		}
		if d.FiledBy != "user1" {
			t.Error("incorrect filer")
		}
		if d.StakeAmount != 10.0 {
			t.Errorf("stake = %f, want 10", d.StakeAmount)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		svc := NewDisputeService()
		_, err := svc.FileDispute("user1", "user2", DisputeType("fake"), "evidence", 10)
		if err != ErrDisputeInvalidType {
			t.Errorf("expected ErrDisputeInvalidType, got %v", err)
		}
	})

	t.Run("90 day cooldown", func(t *testing.T) {
		svc := NewDisputeService()
		svc.FileDispute("user1", "user2", DisputeFalseReport, "first", 10)
		_, err := svc.FileDispute("user1", "user3", DisputeSystemError, "second", 10)
		if err != ErrDisputeRateLimited {
			t.Errorf("expected ErrDisputeRateLimited, got %v", err)
		}
	})

	t.Run("cooldown expires", func(t *testing.T) {
		svc := NewDisputeService()
		svc.FileDispute("user1", "user2", DisputeFalseReport, "first", 10)

		// Expire cooldown
		svc.mu.Lock()
		for _, d := range svc.disputes {
			d.CreatedAt = time.Now().Add(-91 * 24 * time.Hour)
		}
		svc.mu.Unlock()

		_, err := svc.FileDispute("user1", "user3", DisputeSystemError, "second", 10)
		if err != nil {
			t.Fatalf("dispute after cooldown should work: %v", err)
		}
	})
}

func TestAssignJurors(t *testing.T) {
	t.Run("assign valid jurors", func(t *testing.T) {
		svc := NewDisputeService()
		d, _ := svc.FileDispute("user1", "user2", DisputeFalseReport, "evidence", 10)

		jurors := []string{"juror1", "juror2", "juror3", "juror4", "juror5"}
		err := svc.AssignJurors(d.ID, jurors)
		if err != nil {
			t.Fatalf("AssignJurors failed: %v", err)
		}

		got, _ := svc.GetDispute(d.ID)
		if got.Status != DisputeReview {
			t.Errorf("status = %s, want in_review", got.Status)
		}
		if len(got.Jurors) != 5 {
			t.Errorf("jurors = %d, want 5", len(got.Jurors))
		}
	})

	t.Run("juror conflict with filer", func(t *testing.T) {
		svc := NewDisputeService()
		d, _ := svc.FileDispute("user1", "user2", DisputeFalseReport, "evidence", 10)

		err := svc.AssignJurors(d.ID, []string{"user1", "j2", "j3", "j4", "j5"})
		if err != ErrJurorConflict {
			t.Errorf("expected ErrJurorConflict, got %v", err)
		}
	})

	t.Run("juror conflict with target", func(t *testing.T) {
		svc := NewDisputeService()
		d, _ := svc.FileDispute("user1", "user2", DisputeFalseReport, "evidence", 10)

		err := svc.AssignJurors(d.ID, []string{"j1", "user2", "j3", "j4", "j5"})
		if err != ErrJurorConflict {
			t.Errorf("expected ErrJurorConflict, got %v", err)
		}
	})
}

func TestCastVote(t *testing.T) {
	svc := NewDisputeService()
	d, _ := svc.FileDispute("user1", "user2", DisputeFalseReport, "evidence", 10)
	jurors := []string{"j1", "j2", "j3", "j4", "j5"}
	svc.AssignJurors(d.ID, jurors)

	t.Run("valid vote", func(t *testing.T) {
		err := svc.CastVote(d.ID, "j1", VoteUphold, "I agree")
		if err != nil {
			t.Fatalf("CastVote failed: %v", err)
		}
	})

	t.Run("non-juror vote", func(t *testing.T) {
		err := svc.CastVote(d.ID, "random", VoteUphold, "")
		if err != ErrJurorIneligible {
			t.Errorf("expected ErrJurorIneligible, got %v", err)
		}
	})

	t.Run("duplicate vote", func(t *testing.T) {
		err := svc.CastVote(d.ID, "j1", VoteReject, "changed mind")
		if err != ErrJurorAlreadyVoted {
			t.Errorf("expected ErrJurorAlreadyVoted, got %v", err)
		}
	})
}

func TestDisputeResolution(t *testing.T) {
	t.Run("upheld by majority", func(t *testing.T) {
		svc := NewDisputeService()
		d, _ := svc.FileDispute("user1", "user2", DisputeFalseReport, "evidence", 10)
		jurors := []string{"j1", "j2", "j3", "j4", "j5"}
		svc.AssignJurors(d.ID, jurors)

		svc.CastVote(d.ID, "j1", VoteUphold, "")
		svc.CastVote(d.ID, "j2", VoteUphold, "")
		svc.CastVote(d.ID, "j3", VoteUphold, "")
		svc.CastVote(d.ID, "j4", VoteReject, "")
		svc.CastVote(d.ID, "j5", VoteReject, "")

		got, _ := svc.GetDispute(d.ID)
		if got.Status != DisputeUpheld {
			t.Errorf("status = %s, want upheld", got.Status)
		}
		if !got.StakeRefunded {
			t.Error("stake should be refunded when upheld")
		}
	})

	t.Run("rejected by majority", func(t *testing.T) {
		svc := NewDisputeService()
		d, _ := svc.FileDispute("user1", "user2", DisputeCoordinatedAttack, "evidence", 10)
		jurors := []string{"j1", "j2", "j3", "j4", "j5"}
		svc.AssignJurors(d.ID, jurors)

		svc.CastVote(d.ID, "j1", VoteReject, "")
		svc.CastVote(d.ID, "j2", VoteReject, "")
		svc.CastVote(d.ID, "j3", VoteReject, "")
		svc.CastVote(d.ID, "j4", VoteUphold, "")
		svc.CastVote(d.ID, "j5", VoteUphold, "")

		got, _ := svc.GetDispute(d.ID)
		if got.Status != DisputeRejected {
			t.Errorf("status = %s, want rejected", got.Status)
		}
		if got.StakeRefunded {
			t.Error("stake should not be refunded when rejected")
		}
	})
}

func TestResolveExpired(t *testing.T) {
	t.Run("expire with no votes", func(t *testing.T) {
		svc := NewDisputeService()
		d, _ := svc.FileDispute("user1", "user2", DisputeSystemError, "evidence", 10)
		jurors := []string{"j1", "j2", "j3", "j4", "j5"}
		svc.AssignJurors(d.ID, jurors)

		// Force expire
		svc.mu.Lock()
		svc.disputes[d.ID].ExpiresAt = time.Now().Add(-1 * time.Hour)
		svc.mu.Unlock()

		expired := svc.ResolveExpired()
		if len(expired) != 1 {
			t.Fatalf("expired = %d, want 1", len(expired))
		}
		if expired[0].Status != DisputeExpired {
			t.Errorf("status = %s, want expired", expired[0].Status)
		}
		if !expired[0].StakeRefunded {
			t.Error("stake should be refunded when expired with no votes")
		}
	})

	t.Run("expire with partial votes resolves", func(t *testing.T) {
		svc := NewDisputeService()
		d, _ := svc.FileDispute("user1", "user2", DisputeFalseReport, "evidence", 10)
		jurors := []string{"j1", "j2", "j3", "j4", "j5"}
		svc.AssignJurors(d.ID, jurors)

		svc.CastVote(d.ID, "j1", VoteUphold, "")
		svc.CastVote(d.ID, "j2", VoteUphold, "")

		// Force expire
		svc.mu.Lock()
		svc.disputes[d.ID].ExpiresAt = time.Now().Add(-1 * time.Hour)
		svc.mu.Unlock()

		expired := svc.ResolveExpired()
		if len(expired) != 1 {
			t.Fatalf("expired = %d, want 1", len(expired))
		}
		// With 2 upholds and 0 rejects, should be upheld
		if expired[0].Status != DisputeUpheld {
			t.Errorf("status = %s, want upheld", expired[0].Status)
		}
	})
}

func TestVoteCount(t *testing.T) {
	svc := NewDisputeService()
	d, _ := svc.FileDispute("user1", "user2", DisputeFalseReport, "evidence", 10)
	jurors := []string{"j1", "j2", "j3", "j4", "j5"}
	svc.AssignJurors(d.ID, jurors)

	svc.CastVote(d.ID, "j1", VoteUphold, "")
	svc.CastVote(d.ID, "j2", VoteReject, "")
	svc.CastVote(d.ID, "j3", VoteAbstain, "")

	u, r, a, err := svc.VoteCount(d.ID)
	if err != nil {
		t.Fatalf("VoteCount failed: %v", err)
	}
	if u != 1 || r != 1 || a != 1 {
		t.Errorf("votes = uphold:%d reject:%d abstain:%d, want 1:1:1", u, r, a)
	}
}

func TestGetUserDisputes(t *testing.T) {
	svc := NewDisputeService()
	svc.FileDispute("user1", "user2", DisputeFalseReport, "evidence", 10)

	disputes := svc.GetUserDisputes("user1")
	if len(disputes) != 1 {
		t.Errorf("disputes = %d, want 1", len(disputes))
	}

	disputes = svc.GetUserDisputes("nobody")
	if len(disputes) != 0 {
		t.Errorf("disputes = %d, want 0", len(disputes))
	}
}

func TestGetJurorDisputes(t *testing.T) {
	svc := NewDisputeService()
	d, _ := svc.FileDispute("user1", "user2", DisputeFalseReport, "evidence", 10)
	svc.AssignJurors(d.ID, []string{"j1", "j2", "j3", "j4", "j5"})

	disputes := svc.GetJurorDisputes("j1")
	if len(disputes) != 1 {
		t.Errorf("juror disputes = %d, want 1", len(disputes))
	}
}
