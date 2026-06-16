package comply

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// ActivateLitigationHoldInput is the WO-251 hold activation request.
type ActivateLitigationHoldInput struct {
	OrgDID         string
	MatterID       string
	CustodianDIDs  []string
	ScopeLabel     string
	ActivatedByDID string
}

// ActivateLitigationHold disables disappearing messages, binds litigation_hold policies,
// and notifies custodians within the 5-second SLA target.
func (s *Service) ActivateLitigationHold(ctx context.Context, in ActivateLitigationHoldInput) (*database.LitigationMatter, error) {
	if in.OrgDID == "" || in.MatterID == "" || in.ActivatedByDID == "" || len(in.CustodianDIDs) == 0 {
		return nil, ErrInvalidPolicy
	}
	if s.convIndex == nil {
		return nil, fmt.Errorf("conversation index not configured")
	}

	now := time.Now().UTC()
	anchorRef := s.anchorRef(in.OrgDID, "litigation_hold_active", in.MatterID, now)

	matter := &database.LitigationMatter{
		MatterID:       in.MatterID,
		OrgDID:         in.OrgDID,
		ScopeLabel:     in.ScopeLabel,
		Status:         database.MatterActive,
		ActivatedAt:    now,
		ActivatedByDID: in.ActivatedByDID,
		DataL1Ref:      anchorRef,
	}

	conversationSet := make(map[string]struct{})
	for _, custodian := range in.CustodianDIDs {
		convIDs, err := s.convIndex.ListConversationIDsForParticipant(ctx, custodian)
		if err != nil {
			return nil, err
		}
		for _, convID := range convIDs {
			conversationSet[convID] = struct{}{}
			pol, err := s.CreateRetentionPolicy(ctx, CreatePolicyInput{
				OrgDID:         in.OrgDID,
				PolicyType:     database.PolicyLitigationHold,
				ConversationID: convID,
				ScopeLabel:     in.ScopeLabel,
				CreatedByDID:   in.ActivatedByDID,
			})
			if err != nil {
				return nil, err
			}
			_ = s.store.AddLitigationCustodian(ctx, &database.LitigationCustodianBinding{
				MatterID:       in.MatterID,
				CustodianDID:   custodian,
				ConversationID: convID,
			})
			s.fingerprintConversationMessages(ctx, in.OrgDID, convID)
			_ = pol
		}
		if s.notifier != nil {
			_ = s.notifier.NotifyLitigationHold(ctx, custodian, in.MatterID)
		}
	}

	matter.CustodianCount = len(conversationSet)
	_ = s.EnsureOrgAdmin(ctx, in.OrgDID, in.ActivatedByDID)
	if err := s.store.CreateLitigationMatter(ctx, matter); err != nil {
		return nil, err
	}
	_ = s.store.AppendAuditEvent(ctx, &database.AuditEvent{
		ID:         uuid.NewString(),
		OrgDID:     in.OrgDID,
		EventType:  "litigation_hold_activated",
		RefID:      in.MatterID,
		DataL1Ref:  anchorRef,
		OccurredAt: now,
	})
	return matter, nil
}

// ReleaseLitigationHold releases all conversations bound under a matter.
func (s *Service) ReleaseLitigationHold(ctx context.Context, orgDID, matterID, releasedByDID string) (*database.LitigationMatter, error) {
	matter, err := s.store.GetLitigationMatter(ctx, matterID)
	if err != nil {
		return nil, err
	}
	if matter.OrgDID != orgDID {
		return nil, database.ErrComplyMatterNotFound
	}
	bindings, err := s.store.ListLitigationCustodians(ctx, matterID)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	for _, b := range bindings {
		if _, ok := seen[b.ConversationID]; ok {
			continue
		}
		seen[b.ConversationID] = struct{}{}
		_ = s.ReleaseConversation(ctx, b.ConversationID)
	}
	now := time.Now().UTC()
	matter.Status = database.MatterReleased
	matter.ReleasedAt = &now
	matter.ReleasedByDID = releasedByDID
	matter.DataL1Ref = s.anchorRef(orgDID, "litigation_hold_released", matterID, now)
	if err := s.store.UpdateLitigationMatter(ctx, matter); err != nil {
		return nil, err
	}
	_ = s.store.AppendAuditEvent(ctx, &database.AuditEvent{
		ID:         uuid.NewString(),
		OrgDID:     orgDID,
		EventType:  "litigation_hold_released",
		RefID:      matterID,
		DataL1Ref:  matter.DataL1Ref,
		OccurredAt: now,
	})
	return matter, nil
}

// GetLitigationHold returns hold status for a matter.
func (s *Service) GetLitigationHold(ctx context.Context, orgDID, matterID string) (*database.LitigationMatter, error) {
	matter, err := s.store.GetLitigationMatter(ctx, matterID)
	if err != nil {
		return nil, err
	}
	if matter.OrgDID != orgDID {
		return nil, database.ErrComplyMatterNotFound
	}
	return matter, nil
}

func (s *Service) anchorRef(orgDID, eventType, refID string, at time.Time) string {
	return policyAnchorRef(orgDID, database.RetentionPolicyType(eventType), refID, at.Format(time.RFC3339))
}
