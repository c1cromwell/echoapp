package governance

import (
	"context"
	"fmt"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// TrustQuerier abstracts trust tier lookups.
type TrustQuerier interface {
	GetTrustTier(ctx context.Context, did string) (int, error)
}

// StakeQuerier abstracts staked position lookups for governance weight.
type StakeQuerier interface {
	GetTotalStaked(ctx context.Context, did string) (int64, error)
}

// GovernanceService orchestrates proposal CRUD, vote submission, and tally queries.
type GovernanceService struct {
	metagraph *metagraph.MetagraphClient
	trust     TrustQuerier
	stake     StakeQuerier
	store     ProposalStore
}

// NewGovernanceService creates a GovernanceService with the given dependencies.
func NewGovernanceService(
	metagraph *metagraph.MetagraphClient,
	trust TrustQuerier,
	stake StakeQuerier,
	store ProposalStore,
) *GovernanceService {
	return &GovernanceService{
		metagraph: metagraph,
		trust:     trust,
		stake:     stake,
		store:     store,
	}
}

// GetVotingPower returns the pre-validated governance weight for a DID.
// The Data L1 Scala validator performs the authoritative calculation.
func (s *GovernanceService) GetVotingPower(ctx context.Context, did string) (*VotingPower, error) {
	tier, err := s.trust.GetTrustTier(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("get trust tier: %w", err)
	}

	totalStaked, err := s.stake.GetTotalStaked(ctx, did)
	if err != nil {
		return nil, fmt.Errorf("get total staked: %w", err)
	}

	weight := CalculateWeight(totalStaked, tier)
	canVote := CanVote(tier, totalStaked)

	return &VotingPower{
		DID:         did,
		TrustTier:   tier,
		Multiplier:  TierMultiplierFloat(tier),
		TotalStaked: totalStaked,
		Weight:      weight,
		CanVote:     canVote,
	}, nil
}

// SubmitVote pre-validates and submits a governance vote to Data L1.
func (s *GovernanceService) SubmitVote(ctx context.Context, req VoteRequest) (*VoteResult, error) {
	if !ValidateVoteValue(req.Value) {
		return nil, ErrInvalidVoteValue
	}

	// Get voting power
	power, err := s.GetVotingPower(ctx, req.DID)
	if err != nil {
		return nil, err
	}
	if !power.CanVote {
		return nil, ErrCannotVote
	}

	// Check proposal is active
	proposal, err := s.store.GetProposal(ctx, req.ProposalID)
	if err != nil {
		return nil, ErrProposalNotFound
	}
	if proposal.Status != StatusActive {
		return nil, ErrProposalNotActive
	}
	if time.Now().After(proposal.EndsAt) {
		return nil, ErrProposalExpired
	}

	// Check one-vote-per-DID
	voted, err := s.store.HasVoted(ctx, req.DID, req.ProposalID)
	if err != nil {
		return nil, fmt.Errorf("check voted: %w", err)
	}
	if voted {
		return nil, ErrAlreadyVoted
	}

	// Submit to Data L1
	dataL1Tx := map[string]interface{}{
		"type":        "governance_vote",
		"proposalId":  req.ProposalID,
		"voterDid":    req.DID,
		"value":       req.Value,
		"stakeWeight": power.Weight,
		"trustTier":   power.TrustTier,
	}

	txHash, err := s.metagraph.SubmitDataL1(ctx, dataL1Tx)
	if err != nil {
		return nil, fmt.Errorf("submit vote to data l1: %w", err)
	}

	// Record vote locally for cache/audit
	record := &VoteRecord{
		DID:        req.DID,
		ProposalID: req.ProposalID,
		Value:      req.Value,
		Weight:     power.Weight,
		TrustTier:  power.TrustTier,
		Staked:     power.TotalStaked,
		TxHash:     txHash,
		CreatedAt:  time.Now(),
	}
	if err := s.store.RecordVote(ctx, record); err != nil {
		// Non-fatal: L1 is source of truth, local cache is best-effort
		_ = err
	}

	return &VoteResult{
		TxHash: txHash,
		Weight: power.Weight,
	}, nil
}

// CreateProposal creates a new governance proposal.
func (s *GovernanceService) CreateProposal(ctx context.Context, req CreateProposalRequest) (*Proposal, error) {
	if !ValidateProposalType(req.Type) {
		return nil, ErrInvalidProposalType
	}
	if !ValidateThreshold(req.Threshold) {
		return nil, ErrInvalidThreshold
	}

	proposal := &Proposal{
		ID:          fmt.Sprintf("prop_%d", time.Now().UnixNano()),
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Threshold:   req.Threshold,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		EndsAt:      req.EndsAt,
		Status:      StatusActive,
	}

	if err := s.store.CreateProposal(ctx, proposal); err != nil {
		return nil, fmt.Errorf("create proposal: %w", err)
	}

	return proposal, nil
}

// GetProposalTally returns the current vote tally for a proposal.
func (s *GovernanceService) GetProposalTally(ctx context.Context, proposalID string) (*ProposalTally, error) {
	tally, err := s.store.GetTally(ctx, proposalID)
	if err != nil {
		return nil, fmt.Errorf("get tally: %w", err)
	}
	return tally, nil
}

// ListActiveProposals returns all proposals currently open for voting.
func (s *GovernanceService) ListActiveProposals(ctx context.Context) ([]Proposal, error) {
	return s.store.ListActiveProposals(ctx)
}
