package governance

import (
	"context"
	"time"
)

// ProposalStore abstracts persistence for proposals and votes.
type ProposalStore interface {
	// CreateProposal inserts a new proposal.
	CreateProposal(ctx context.Context, proposal *Proposal) error

	// GetProposal returns a proposal by ID.
	GetProposal(ctx context.Context, id string) (*Proposal, error)

	// ListActiveProposals returns proposals with status "active" and endsAt > now.
	ListActiveProposals(ctx context.Context) ([]Proposal, error)

	// RecordVote persists a vote record and updates the cached tally on the proposal.
	RecordVote(ctx context.Context, vote *VoteRecord) error

	// HasVoted returns true if the DID has already voted on the proposal.
	HasVoted(ctx context.Context, did, proposalID string) (bool, error)

	// GetTally returns the current tally for a proposal.
	GetTally(ctx context.Context, proposalID string) (*ProposalTally, error)

	// UpdateProposalStatus sets a proposal's status (passed, failed, executed).
	UpdateProposalStatus(ctx context.Context, id, status string) error
}

// FinalizeExpiredProposals checks active proposals past their endsAt and marks them passed/failed.
func FinalizeExpiredProposals(ctx context.Context, store ProposalStore) (int, error) {
	proposals, err := store.ListActiveProposals(ctx)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	finalized := 0

	for _, p := range proposals {
		if now.Before(p.EndsAt) {
			continue
		}

		tally, err := store.GetTally(ctx, p.ID)
		if err != nil {
			continue
		}

		passed := CheckThresholdPassed(p.Threshold, tally.ForWeight, tally.AgainstWeight, tally.TotalWeight)

		newStatus := StatusFailed
		if passed {
			newStatus = StatusPassed
		}

		if err := store.UpdateProposalStatus(ctx, p.ID, newStatus); err != nil {
			continue
		}
		finalized++
	}

	return finalized, nil
}

// BuildTally computes a ProposalTally from the stored proposal data.
func BuildTally(proposalID string, forWeight, againstWeight, abstainWeight int64, voterCount int, threshold string) *ProposalTally {
	totalWeight := forWeight + againstWeight + abstainWeight

	var forPercent float64
	if totalWeight > 0 {
		forPercent = float64(forWeight*100) / float64(totalWeight)
	}

	passed := CheckThresholdPassed(threshold, forWeight, againstWeight, totalWeight)

	return &ProposalTally{
		ProposalID:    proposalID,
		ForWeight:     forWeight,
		AgainstWeight: againstWeight,
		AbstainWeight: abstainWeight,
		TotalWeight:   totalWeight,
		ForPercent:    forPercent,
		VoterCount:    voterCount,
		Passed:        passed,
	}
}
