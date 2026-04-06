package governance

import (
	"errors"
	"time"
)

// Proposal types.
const (
	ProposalTypeProtocolUpgrade    = "protocol_upgrade"
	ProposalTypeTreasuryAllocation = "treasury_allocation"
	ProposalTypeParameterChange    = "parameter_change"
	ProposalTypeBoardElection      = "board_election"
)

// Threshold types.
const (
	ThresholdSimpleMajority = "simple_majority"
	ThresholdSupermajority67 = "supermajority_67"
	ThresholdSupermajority75 = "supermajority_75"
)

// Vote values.
const (
	VoteFor     = "for"
	VoteAgainst = "against"
	VoteAbstain = "abstain"
)

// Proposal status values.
const (
	StatusActive   = "active"
	StatusPassed   = "passed"
	StatusFailed   = "failed"
	StatusExecuted = "executed"
)

// Errors.
var (
	ErrCannotVote         = errors.New("user does not meet voting requirements")
	ErrProposalNotFound   = errors.New("proposal not found")
	ErrProposalNotActive  = errors.New("proposal is not active")
	ErrAlreadyVoted       = errors.New("already voted on this proposal")
	ErrInvalidVoteValue   = errors.New("vote value must be for, against, or abstain")
	ErrInvalidProposalType = errors.New("invalid proposal type")
	ErrInvalidThreshold   = errors.New("invalid threshold type")
	ErrProposalExpired    = errors.New("proposal voting period has ended")
)

// VotingPower represents a user's pre-validated governance weight.
type VotingPower struct {
	DID         string  `json:"did"`
	TrustTier   int     `json:"trustTier"`
	Multiplier  float64 `json:"multiplier"`
	TotalStaked int64   `json:"totalStaked"`
	Weight      int64   `json:"weight"`
	CanVote     bool    `json:"canVote"`
}

// VoteRequest is the input to SubmitVote.
type VoteRequest struct {
	DID        string `json:"did"`
	ProposalID string `json:"proposalId"`
	Value      string `json:"value"`
}

// VoteResult is the output from SubmitVote.
type VoteResult struct {
	TxHash string `json:"txHash"`
	Weight int64  `json:"weight"`
}

// Proposal represents a governance proposal.
type Proposal struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Type        string         `json:"type"`
	Threshold   string         `json:"threshold"`
	CreatedBy   string         `json:"createdBy"`
	CreatedAt   time.Time      `json:"createdAt"`
	EndsAt      time.Time      `json:"endsAt"`
	Status      string         `json:"status"`
	Tally       *ProposalTally `json:"tally,omitempty"`
}

// ProposalTally aggregates vote weights for a proposal.
type ProposalTally struct {
	ProposalID    string  `json:"proposalId"`
	ForWeight     int64   `json:"forWeight"`
	AgainstWeight int64   `json:"againstWeight"`
	AbstainWeight int64   `json:"abstainWeight"`
	TotalWeight   int64   `json:"totalWeight"`
	ForPercent    float64 `json:"forPercent"`
	VoterCount    int     `json:"voterCount"`
	Passed        bool    `json:"passed"`
}

// VoteRecord persists an individual vote for audit.
type VoteRecord struct {
	DID        string    `json:"did"`
	ProposalID string    `json:"proposalId"`
	Value      string    `json:"value"`
	Weight     int64     `json:"weight"`
	TrustTier  int       `json:"trustTier"`
	Staked     int64     `json:"staked"`
	TxHash     string    `json:"txHash"`
	CreatedAt  time.Time `json:"createdAt"`
}

// CreateProposalRequest is the input for creating a new proposal.
type CreateProposalRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Threshold   string    `json:"threshold"`
	CreatedBy   string    `json:"createdBy"`
	EndsAt      time.Time `json:"endsAt"`
}

// ValidateVoteValue returns true if the vote value is valid.
func ValidateVoteValue(value string) bool {
	return value == VoteFor || value == VoteAgainst || value == VoteAbstain
}

// ValidateProposalType returns true if the proposal type is valid.
func ValidateProposalType(t string) bool {
	switch t {
	case ProposalTypeProtocolUpgrade, ProposalTypeTreasuryAllocation,
		ProposalTypeParameterChange, ProposalTypeBoardElection:
		return true
	}
	return false
}

// ValidateThreshold returns true if the threshold type is valid.
func ValidateThreshold(t string) bool {
	switch t {
	case ThresholdSimpleMajority, ThresholdSupermajority67, ThresholdSupermajority75:
		return true
	}
	return false
}
