// T0–T7: Governance pre-validation receives T7 public chain data (proposal IDs,
// trust tiers) and T6 vote values. No message content, DIDs, or PII should be
// passed to these functions — the proposal and voter DIDs live in the transaction
// envelope (T7), not the governance data payload.
package validation

// Governance pre-validation (WO-35).
//
// These are pure, stateless checks that run before the handler hits the
// database. The GovernanceService itself performs authoritative validation
// (one-vote-per-DID, tally computation); this layer catches obvious invalid
// requests early to save DB round-trips.

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidVoteValue      = errors.New("validation: vote value must be 'for', 'against', or 'abstain'")
	ErrProposalNotActive     = errors.New("validation: proposal is not active")
	ErrProposalExpired       = errors.New("validation: proposal voting period has ended")
	ErrInsufficientTrustTier = errors.New("validation: trust tier too low for governance (minimum tier 2)")
	ErrNoStake               = errors.New("validation: user has no staked tokens (required for governance)")
)

// ValidVoteValues is the set of accepted vote strings.
var ValidVoteValues = map[string]struct{}{
	"for":     {},
	"against": {},
	"abstain": {},
}

// GovernanceVotePreValidation holds the fields for WO-35 governance pre-validation.
type GovernanceVotePreValidation struct {
	// VoteValue is the user's submitted vote string.
	VoteValue string
	// ProposalStatus is the current status of the proposal (e.g. "active").
	ProposalStatus string
	// ProposalEndsAt is the proposal's voting deadline.
	ProposalEndsAt time.Time
	// TrustTier is the voter's server-cached trust tier.
	TrustTier int
	// TotalStaked is the voter's current staked token balance.
	TotalStaked int64
}

// ValidateGovernanceVotePre performs server-side pre-checks for a governance vote.
// Returns the first error found; the GovernanceService performs additional
// authoritative validation (duplicate-vote check, weight calculation).
func ValidateGovernanceVotePre(v GovernanceVotePreValidation) error {
	if _, ok := ValidVoteValues[v.VoteValue]; !ok {
		return fmt.Errorf("%w: got %q", ErrInvalidVoteValue, v.VoteValue)
	}
	if v.ProposalStatus != "active" {
		return fmt.Errorf("%w: status=%q", ErrProposalNotActive, v.ProposalStatus)
	}
	if !v.ProposalEndsAt.IsZero() && time.Now().After(v.ProposalEndsAt) {
		return fmt.Errorf("%w: ended at %s", ErrProposalExpired, v.ProposalEndsAt.UTC().Format(time.RFC3339))
	}
	if v.TrustTier < 2 {
		return fmt.Errorf("%w: tier=%d", ErrInsufficientTrustTier, v.TrustTier)
	}
	if v.TotalStaked <= 0 {
		return ErrNoStake
	}
	return nil
}
