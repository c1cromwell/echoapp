package metagraph

import "time"

// GroupMetadataUpdate is anchored on Data L1 for group creation / membership hash (WO-156).
type GroupMetadataUpdate struct {
	Type            string `json:"type"` // group_metadata
	GroupID         string `json:"groupId"`
	AdminDID        string `json:"adminDID"`
	MemberCountHash string `json:"memberCountHash"`
	CreatedAt       int64  `json:"createdAt"` // epoch millis
}

// GovernanceVoteAnchor records closed vote outcomes on Data L1 (WO-156).
type GovernanceVoteAnchor struct {
	Type              string  `json:"type"` // governance_vote
	GroupID           string  `json:"groupId"`
	ProposalID        string  `json:"proposalId"`
	Result            string  `json:"result"`
	ParticipationPct  float64 `json:"participationPct"`
	ClosedAt          int64   `json:"closedAt"`
}

// GroupBlockchainProof is returned by GET /v1/groups/{id}/blockchain-proof.
type GroupBlockchainProof struct {
	GroupID         string    `json:"groupId"`
	AdminDID        string    `json:"adminDID"`
	MemberCountHash string    `json:"memberCountHash"`
	TxHash          string    `json:"txHash,omitempty"`
	SnapshotHash    string    `json:"snapshotHash,omitempty"`
	AnchoredAt      time.Time `json:"anchoredAt"`
	EventType       string    `json:"eventType"`
}
