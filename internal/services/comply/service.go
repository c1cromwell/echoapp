package comply

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/database"
)

var (
	ErrRetentionPolicyActive = errors.New("retention_policy_active")
	ErrInvalidPolicy         = errors.New("invalid retention policy")
)

// MessageOps applies conversation retention flags for M1 hybrid edit history.
type MessageOps interface {
	SetConversationRetention(ctx context.Context, conversationID string, retained bool) error
	SetDisappearingTTL(ctx context.Context, conversationID string, ttlSeconds int) error
}

// Service implements WO-250 Comply retention policy management and enforcement hooks.
type Service struct {
	store        database.ComplyExtendedStore
	messageOps   MessageOps
	convIndex    database.ConversationIndex
	serviceToken string
	notifier     PushNotifier
}

// Deps wires optional integrations for the Comply service.
type Deps struct {
	MessageOps        MessageOps
	ConversationIndex database.ConversationIndex
	ServiceToken      string
	Notifier          PushNotifier
}

func NewService(store database.ComplyExtendedStore, deps Deps) *Service {
	return &Service{
		store:        store,
		messageOps:   deps.MessageOps,
		convIndex:    deps.ConversationIndex,
		serviceToken: deps.ServiceToken,
		notifier:     deps.Notifier,
	}
}

// NewServiceLegacy preserves the gateway constructor (store implements both interfaces).
func NewServiceLegacy(store database.ComplyExtendedStore, messageOps MessageOps, serviceToken string) *Service {
	var idx database.ConversationIndex
	if ci, ok := store.(database.ConversationIndex); ok {
		idx = ci
	}
	return NewService(store, Deps{
		MessageOps:        messageOps,
		ConversationIndex: idx,
		ServiceToken:      serviceToken,
	})
}

func (s *Service) ServiceToken() string { return s.serviceToken }

type CreatePolicyInput struct {
	OrgDID         string
	PolicyType     database.RetentionPolicyType
	ConversationID string
	ScopeLabel     string
	ExpiresAt      *time.Time
	CreatedByDID   string
}

// CreateRetentionPolicy registers a policy and optionally binds a conversation (WO-250).
func (s *Service) CreateRetentionPolicy(ctx context.Context, in CreatePolicyInput) (*database.RetentionPolicy, error) {
	if in.OrgDID == "" || !in.PolicyType.Valid() || in.CreatedByDID == "" {
		return nil, ErrInvalidPolicy
	}
	pol := &database.RetentionPolicy{
		ID:             uuid.NewString(),
		OrgDID:         in.OrgDID,
		PolicyType:     in.PolicyType,
		ConversationID: in.ConversationID,
		ScopeLabel:     in.ScopeLabel,
		EffectiveAt:    time.Now().UTC(),
		ExpiresAt:      in.ExpiresAt,
		DataL1Ref:      policyAnchorRef(in.OrgDID, in.PolicyType, in.ConversationID, in.ScopeLabel),
		Active:         true,
		CreatedByDID:   in.CreatedByDID,
		CreatedAt:      time.Now().UTC(),
	}
	if err := s.store.CreateRetentionPolicy(ctx, pol); err != nil {
		return nil, err
	}
	_ = s.store.UpsertOrgProfile(ctx, &database.ComplyOrgProfile{OrgDID: in.OrgDID})
	_ = s.EnsureOrgAdmin(ctx, in.OrgDID, in.CreatedByDID)
	if in.ConversationID != "" {
		if err := s.applyPolicyToConversation(ctx, pol, in.ConversationID); err != nil {
			return nil, err
		}
	}
	return pol, nil
}

// ApplyPolicyToConversation binds an existing policy to a conversation.
func (s *Service) ApplyPolicyToConversation(ctx context.Context, policyID, conversationID string) error {
	pol, err := s.store.GetRetentionPolicy(ctx, policyID)
	if err != nil {
		return err
	}
	return s.applyPolicyToConversation(ctx, pol, conversationID)
}

func (s *Service) applyPolicyToConversation(ctx context.Context, pol *database.RetentionPolicy, conversationID string) error {
	if s.messageOps == nil {
		return fmt.Errorf("message ops not configured")
	}
	if err := s.store.BindConversationPolicy(ctx, &database.ConversationPolicyBinding{
		ConversationID: conversationID,
		PolicyID:       pol.ID,
		OrgDID:         pol.OrgDID,
		BoundAt:        time.Now().UTC(),
	}); err != nil {
		return err
	}
	if err := s.messageOps.SetConversationRetention(ctx, conversationID, true); err != nil {
		return err
	}
	if pol.PolicyType == database.PolicyLitigationHold || pol.PolicyType == database.PolicyPermanent {
		_ = s.messageOps.SetDisappearingTTL(ctx, conversationID, 0)
	}
	return nil
}

// ReleaseConversation removes Comply binding and clears the M1 retention flag.
func (s *Service) ReleaseConversation(ctx context.Context, conversationID string) error {
	if s.messageOps == nil {
		return fmt.Errorf("message ops not configured")
	}
	_ = s.store.UnbindConversation(ctx, conversationID)
	return s.messageOps.SetConversationRetention(ctx, conversationID, false)
}

// ActivePolicyForConversation returns the bound policy if any.
func (s *Service) ActivePolicyForConversation(ctx context.Context, conversationID string) (*database.RetentionPolicy, error) {
	binding, err := s.store.GetConversationBinding(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	pol, err := s.store.GetRetentionPolicy(ctx, binding.PolicyID)
	if err != nil || !pol.Active {
		return nil, database.ErrComplyPolicyNotFound
	}
	if pol.ExpiresAt != nil && time.Now().UTC().After(*pol.ExpiresAt) {
		return nil, database.ErrComplyPolicyNotFound
	}
	return pol, nil
}

// BlocksDeletion reports whether Comply enforcement should reject a delete (WO-250).
func (s *Service) BlocksDeletion(ctx context.Context, conversationID string) bool {
	pol, err := s.ActivePolicyForConversation(ctx, conversationID)
	if err != nil {
		return false
	}
	switch pol.PolicyType {
	case database.PolicyPermanent, database.PolicyLitigationHold, database.PolicyTimeLimited:
		return true
	default:
		return false
	}
}

// BlocksDisappearing reports whether disappearing messages are forbidden (litigation hold / permanent).
func (s *Service) BlocksDisappearing(ctx context.Context, conversationID string) bool {
	pol, err := s.ActivePolicyForConversation(ctx, conversationID)
	if err != nil {
		return false
	}
	return pol.PolicyType == database.PolicyLitigationHold || pol.PolicyType == database.PolicyPermanent
}

// ListPolicies returns active retention policies for an org.
func (s *Service) ListPolicies(ctx context.Context, orgDID string) ([]*database.RetentionPolicy, error) {
	return s.store.ListRetentionPolicies(ctx, orgDID, true)
}

// Dashboard returns zero-PII aggregate stats for WO-252 / web portal.
func (s *Service) Dashboard(ctx context.Context, orgDID string) (map[string]interface{}, error) {
	active, err := s.store.CountActivePolicies(ctx, orgDID, "")
	if err != nil {
		return nil, err
	}
	holds, err := s.store.CountActiveLitigationMatters(ctx, orgDID)
	if err != nil {
		return nil, err
	}
	pending, err := s.store.CountPendingExports(ctx, orgDID)
	if err != nil {
		return nil, err
	}
	anchorHealth := "healthy"
	if active == 0 && holds == 0 {
		anchorHealth = "degraded"
	}
	return map[string]interface{}{
		"deCoverageRate":          s.deCoverageRate(ctx, orgDID),
		"activeRetentionPolicies": active,
		"litigationHolds":         holds,
		"pendingExports":          pending,
		"anchorHealth":            anchorHealth,
	}, nil
}

func policyAnchorRef(orgDID string, policyType database.RetentionPolicyType, conversationID, scopeLabel string) string {
	payload := orgDID + "|" + string(policyType) + "|" + conversationID + "|" + scopeLabel + "|" + time.Now().UTC().Format(time.RFC3339)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}
