package groups

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/thechadcromwell/echoapp/internal/metagraph"
)

// GroupAnchoringService submits group metadata to Data L1 (WO-156).
type GroupAnchoringService struct {
	dataL1 *metagraph.MetagraphClient

	mu            sync.Mutex
	proofs        map[string]metagraph.GroupBlockchainProof
	pending       map[string]pendingMembership
	lastSubmitted map[string]time.Time
}

type pendingMembership struct {
	groupID     string
	adminDID    string
	memberCount int
	dirtyAt     time.Time
}

// NewGroupAnchoringService creates a group metadata anchoring pipeline.
func NewGroupAnchoringService(dataL1 *metagraph.MetagraphClient) *GroupAnchoringService {
	return &GroupAnchoringService{
		dataL1:        dataL1,
		proofs:        make(map[string]metagraph.GroupBlockchainProof),
		pending:       make(map[string]pendingMembership),
		lastSubmitted: make(map[string]time.Time),
	}
}

// MemberCountHash returns H(memberCount|salt) without revealing the raw count on-chain.
func MemberCountHash(memberCount int, salt string) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%s", memberCount, salt)))
	return hex.EncodeToString(sum[:])
}

// AnchorGroupCreated submits immediate group creation metadata.
func (s *GroupAnchoringService) AnchorGroupCreated(ctx context.Context, groupID, adminDID string, memberCount int) (string, error) {
	if s == nil || s.dataL1 == nil {
		return "", nil
	}
	update := metagraph.GroupMetadataUpdate{
		Type:            "group_metadata",
		GroupID:         groupID,
		AdminDID:        adminDID,
		MemberCountHash: MemberCountHash(memberCount, groupID),
		CreatedAt:       time.Now().UnixMilli(),
	}
	txHash, err := s.dataL1.SubmitDataL1(ctx, update)
	if err != nil {
		return "", err
	}
	s.cacheProof(groupID, adminDID, update.MemberCountHash, txHash, "group_metadata")
	return txHash, nil
}

// QueueMembershipUpdate marks a group for the next 24h metadata batch.
func (s *GroupAnchoringService) QueueMembershipUpdate(groupID, adminDID string, memberCount int) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[groupID] = pendingMembership{
		groupID:     groupID,
		adminDID:    adminDID,
		memberCount: memberCount,
		dirtyAt:     time.Now().UTC(),
	}
}

// FlushPending submits membership hash updates older than maxAge.
func (s *GroupAnchoringService) FlushPending(ctx context.Context, maxAge time.Duration) error {
	if s == nil || s.dataL1 == nil {
		return nil
	}
	now := time.Now().UTC()
	var batch []pendingMembership
	s.mu.Lock()
	for id, p := range s.pending {
		last := s.lastSubmitted[id]
		if now.Sub(p.dirtyAt) >= maxAge || now.Sub(last) >= maxAge {
			batch = append(batch, p)
			delete(s.pending, id)
		}
	}
	s.mu.Unlock()

	for _, p := range batch {
		if _, err := s.AnchorGroupCreated(ctx, p.groupID, p.adminDID, p.memberCount); err != nil {
			return err
		}
	}
	return nil
}

// AnchorGovernanceVote submits a closed governance vote outcome.
func (s *GroupAnchoringService) AnchorGovernanceVote(ctx context.Context, groupID, proposalID, result string, participationPct float64) (string, error) {
	if s == nil || s.dataL1 == nil {
		return "", nil
	}
	anchor := metagraph.GovernanceVoteAnchor{
		Type:             "governance_vote",
		GroupID:          groupID,
		ProposalID:       proposalID,
		Result:           result,
		ParticipationPct: participationPct,
		ClosedAt:         time.Now().UnixMilli(),
	}
	txHash, err := s.dataL1.SubmitDataL1(ctx, anchor)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	s.proofs[groupID] = metagraph.GroupBlockchainProof{
		GroupID:    groupID,
		TxHash:     txHash,
		AnchoredAt: time.Now().UTC(),
		EventType:  "governance_vote",
	}
	s.mu.Unlock()
	return txHash, nil
}

// GetBlockchainProof returns the latest cached anchored metadata for a group.
func (s *GroupAnchoringService) GetBlockchainProof(groupID string) (metagraph.GroupBlockchainProof, bool) {
	if s == nil {
		return metagraph.GroupBlockchainProof{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.proofs[groupID]
	return p, ok
}

func (s *GroupAnchoringService) cacheProof(groupID, adminDID, memberCountHash, txHash, eventType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proofs[groupID] = metagraph.GroupBlockchainProof{
		GroupID:         groupID,
		AdminDID:        adminDID,
		MemberCountHash: memberCountHash,
		TxHash:          txHash,
		AnchoredAt:      time.Now().UTC(),
		EventType:       eventType,
	}
	s.lastSubmitted[groupID] = time.Now().UTC()
}

// Run flushes queued membership updates every 24 hours.
func (s *GroupAnchoringService) Run(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.FlushPending(ctx, 24*time.Hour)
		}
	}
}
