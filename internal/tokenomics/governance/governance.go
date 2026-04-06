package governance

import (
	"math/big"
	"time"
)

// ProposalType defines the kind of proposal
type ProposalType int

const (
	ParameterChange ProposalType = iota
	EcosystemGrant
	TreasurySpend
	ProtocolUpgrade
	Emergency
)

// ProposalStatus represents proposal state
type ProposalStatus int

const (
	Draft ProposalStatus = iota
	Active
	Queued
	Executed
	Defeated
	QuorumNotMet
)

// Proposal represents a governance proposal
type Proposal struct {
	ProposalID     string
	ProposerID     string
	ProposalType   ProposalType
	Title          string
	Description    string
	Status         ProposalStatus
	CreatedAt      time.Time
	VotingStartsAt time.Time
	VotingEndsAt   time.Time
	ExecutionTime  time.Time
	VotesFor       *big.Int
	VotesAgainst   *big.Int
	VotesAbstain   *big.Int
	TotalVotes     *big.Int
	QuorumRequired *big.Int
}

// Vote represents a single vote cast
type Vote struct {
	VoterID    string
	ProposalID string
	Option     int // 0=For, 1=Against, 2=Abstain
	Weight     *big.Int
	CastAt     time.Time
}

// GovernanceParams contains configuration
type GovernanceParams struct {
	ProposalThreshold  *big.Int
	QuorumPercent      float64
	VotingPeriodDays   int
	TimeLockDays       int
	MinProposalBalance *big.Int
	MaxProposalsActive int
}

// DefaultGovernanceParams returns standard configuration
func DefaultGovernanceParams() *GovernanceParams {
	threshold := new(big.Int)
	threshold.SetString("10000000000000", 10) // 100,000 ECHO

	minBalance := new(big.Int)
	minBalance.SetString("1000000000000", 10) // 10,000 ECHO

	return &GovernanceParams{
		ProposalThreshold:  threshold,
		QuorumPercent:      4.0,
		VotingPeriodDays:   7,
		TimeLockDays:       2,
		MinProposalBalance: minBalance,
		MaxProposalsActive: 50,
	}
}

// GovernanceEngine manages proposals and voting
type GovernanceEngine struct {
	Proposals         map[string]*Proposal
	Votes             map[string][]*Vote
	Params            *GovernanceParams
	CirculatingSupply *big.Int
}

// NewGovernanceEngine creates an engine
func NewGovernanceEngine() *GovernanceEngine {
	total := new(big.Int)
	total.SetString("100000000000000000", 10)

	return &GovernanceEngine{
		Proposals:         make(map[string]*Proposal),
		Votes:             make(map[string][]*Vote),
		Params:            DefaultGovernanceParams(),
		CirculatingSupply: total,
	}
}

// CreateProposal initiates a new proposal
func (ge *GovernanceEngine) CreateProposal(
	proposerID string,
	propType ProposalType,
	title string,
	description string,
) *Proposal {
	proposal := &Proposal{
		ProposalID:     generateProposalID(),
		ProposerID:     proposerID,
		ProposalType:   propType,
		Title:          title,
		Description:    description,
		Status:         Draft,
		CreatedAt:      time.Now(),
		VotingStartsAt: time.Now().Add(time.Hour),
		VotingEndsAt:   time.Now().Add(time.Duration(ge.Params.VotingPeriodDays*24) * time.Hour),
		VotesFor:       big.NewInt(0),
		VotesAgainst:   big.NewInt(0),
		VotesAbstain:   big.NewInt(0),
		TotalVotes:     big.NewInt(0),
		QuorumRequired: calculateQuorum(ge.CirculatingSupply, ge.Params.QuorumPercent),
	}

	ge.Proposals[proposal.ProposalID] = proposal
	ge.Votes[proposal.ProposalID] = make([]*Vote, 0)

	return proposal
}

// CastVote records a vote
func (ge *GovernanceEngine) CastVote(voterID string, proposalID string, option int, weight *big.Int) bool {
	proposal, exists := ge.Proposals[proposalID]
	if !exists || proposal.Status != Active {
		return false
	}

	vote := &Vote{
		VoterID:    voterID,
		ProposalID: proposalID,
		Option:     option,
		Weight:     weight,
		CastAt:     time.Now(),
	}

	ge.Votes[proposalID] = append(ge.Votes[proposalID], vote)

	switch option {
	case 0:
		proposal.VotesFor.Add(proposal.VotesFor, weight)
	case 1:
		proposal.VotesAgainst.Add(proposal.VotesAgainst, weight)
	case 2:
		proposal.VotesAbstain.Add(proposal.VotesAbstain, weight)
	}

	proposal.TotalVotes.Add(proposal.TotalVotes, weight)

	return true
}

// FinalizeProposal concludes voting and determines outcome
func (ge *GovernanceEngine) FinalizeProposal(proposalID string) bool {
	proposal, exists := ge.Proposals[proposalID]
	if !exists || proposal.Status != Active {
		return false
	}

	if time.Now().Before(proposal.VotingEndsAt) {
		return false
	}

	if proposal.TotalVotes.Cmp(proposal.QuorumRequired) < 0 {
		proposal.Status = QuorumNotMet
		return true
	}

	if proposal.VotesFor.Cmp(proposal.VotesAgainst) > 0 {
		proposal.Status = Queued
		proposal.ExecutionTime = time.Now().Add(
			time.Duration(ge.Params.TimeLockDays*24) * time.Hour,
		)
	} else {
		proposal.Status = Defeated
	}

	return true
}

// ExecuteProposal runs the queued proposal
func (ge *GovernanceEngine) ExecuteProposal(proposalID string) bool {
	proposal, exists := ge.Proposals[proposalID]
	if !exists || proposal.Status != Queued {
		return false
	}

	if time.Now().Before(proposal.ExecutionTime) {
		return false
	}

	proposal.Status = Executed

	return true
}

// GetProposalStats returns voting statistics
func (ge *GovernanceEngine) GetProposalStats(proposalID string) map[string]interface{} {
	proposal := ge.Proposals[proposalID]
	if proposal == nil {
		return nil
	}

	return map[string]interface{}{
		"proposal_id":     proposal.ProposalID,
		"title":           proposal.Title,
		"status":          proposal.Status,
		"votes_for":       proposal.VotesFor.String(),
		"votes_against":   proposal.VotesAgainst.String(),
		"votes_abstain":   proposal.VotesAbstain.String(),
		"total_votes":     proposal.TotalVotes.String(),
		"quorum_required": proposal.QuorumRequired.String(),
	}
}

// Helper functions
func generateProposalID() string {
	return "prop-" + time.Now().Format("20060102150405")
}

func calculateQuorum(supply *big.Int, percent float64) *big.Int {
	result := new(big.Int).Mul(supply, big.NewInt(int64(percent)))
	result.Div(result, big.NewInt(100))
	return result
}
