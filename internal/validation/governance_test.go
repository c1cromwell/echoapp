package validation

import (
	"errors"
	"testing"
	"time"
)

func TestValidateGovernanceVotePre(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-time.Minute)

	valid := GovernanceVotePreValidation{
		VoteValue:      "for",
		ProposalStatus: "active",
		ProposalEndsAt: future,
		TrustTier:      3,
		TotalStaked:    1000,
	}

	t.Run("ok", func(t *testing.T) {
		if err := ValidateGovernanceVotePre(valid); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("all valid vote values", func(t *testing.T) {
		for _, v := range []string{"for", "against", "abstain"} {
			req := valid
			req.VoteValue = v
			if err := ValidateGovernanceVotePre(req); err != nil {
				t.Errorf("value %q: %v", v, err)
			}
		}
	})

	t.Run("invalid vote value", func(t *testing.T) {
		req := valid
		req.VoteValue = "maybe"
		if err := ValidateGovernanceVotePre(req); !errors.Is(err, ErrInvalidVoteValue) {
			t.Fatalf("want ErrInvalidVoteValue, got %v", err)
		}
	})

	t.Run("proposal not active", func(t *testing.T) {
		req := valid
		req.ProposalStatus = "passed"
		if err := ValidateGovernanceVotePre(req); !errors.Is(err, ErrProposalNotActive) {
			t.Fatalf("want ErrProposalNotActive, got %v", err)
		}
	})

	t.Run("proposal expired", func(t *testing.T) {
		req := valid
		req.ProposalEndsAt = past
		if err := ValidateGovernanceVotePre(req); !errors.Is(err, ErrProposalExpired) {
			t.Fatalf("want ErrProposalExpired, got %v", err)
		}
	})

	t.Run("zero endsAt skips expiry check", func(t *testing.T) {
		req := valid
		req.ProposalEndsAt = time.Time{}
		if err := ValidateGovernanceVotePre(req); err != nil {
			t.Fatalf("unexpected: %v", err)
		}
	})

	t.Run("trust tier too low", func(t *testing.T) {
		req := valid
		req.TrustTier = 1
		if err := ValidateGovernanceVotePre(req); !errors.Is(err, ErrInsufficientTrustTier) {
			t.Fatalf("want ErrInsufficientTrustTier, got %v", err)
		}
	})

	t.Run("no stake", func(t *testing.T) {
		req := valid
		req.TotalStaked = 0
		if err := ValidateGovernanceVotePre(req); !errors.Is(err, ErrNoStake) {
			t.Fatalf("want ErrNoStake, got %v", err)
		}
	})
}
