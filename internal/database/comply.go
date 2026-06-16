package database

import (
	"context"
	"errors"
	"time"
)

// RetentionPolicyType classifies a Comply retention policy (WO-250).
type RetentionPolicyType string

const (
	PolicyPermanent      RetentionPolicyType = "permanent"
	PolicyTimeLimited    RetentionPolicyType = "time_limited"
	PolicyLitigationHold RetentionPolicyType = "litigation_hold"
)

func (t RetentionPolicyType) Valid() bool {
	switch t {
	case PolicyPermanent, PolicyTimeLimited, PolicyLitigationHold:
		return true
	default:
		return false
	}
}

// RetentionPolicy is org-scoped compliance metadata (no message content).
type RetentionPolicy struct {
	ID             string              `json:"id"`
	OrgDID         string              `json:"orgDid"`
	PolicyType     RetentionPolicyType `json:"policyType"`
	ConversationID string              `json:"conversationId,omitempty"`
	ScopeLabel     string              `json:"scopeLabel,omitempty"`
	EffectiveAt    time.Time           `json:"effectiveAt"`
	ExpiresAt      *time.Time          `json:"expiresAt,omitempty"`
	DataL1Ref      string              `json:"dataL1Ref,omitempty"`
	Active         bool                `json:"active"`
	CreatedByDID   string              `json:"createdByDid"`
	CreatedAt      time.Time           `json:"createdAt"`
}

// ConversationPolicyBinding links a conversation to an active retention policy.
type ConversationPolicyBinding struct {
	ConversationID string    `json:"conversationId"`
	PolicyID       string    `json:"policyId"`
	OrgDID         string    `json:"orgDid"`
	BoundAt        time.Time `json:"boundAt"`
}

// ComplyOrgProfile holds org tier/seat metadata (WO-250 pricing).
type ComplyOrgProfile struct {
	OrgDID    string    `json:"orgDid"`
	Tier      string    `json:"tier"`
	Seats     int       `json:"seats"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ComplyDashboardStats are aggregate zero-PII metrics (WO-252).
type ComplyDashboardStats struct {
	ActiveRetentionPolicies int
	LitigationHolds         int
	PendingExports          int
	AnchorHealth            string
	LastAnchorAt            *time.Time
}

var ErrComplyPolicyNotFound = errors.New("retention policy not found")

// ComplyStore persists Comply policy metadata (WO-250).
type ComplyStore interface {
	CreateRetentionPolicy(ctx context.Context, p *RetentionPolicy) error
	ListRetentionPolicies(ctx context.Context, orgDID string, activeOnly bool) ([]*RetentionPolicy, error)
	GetRetentionPolicy(ctx context.Context, policyID string) (*RetentionPolicy, error)
	BindConversationPolicy(ctx context.Context, binding *ConversationPolicyBinding) error
	GetConversationBinding(ctx context.Context, conversationID string) (*ConversationPolicyBinding, error)
	UnbindConversation(ctx context.Context, conversationID string) error
	UpsertOrgProfile(ctx context.Context, profile *ComplyOrgProfile) error
	GetOrgProfile(ctx context.Context, orgDID string) (*ComplyOrgProfile, error)
	CountActivePolicies(ctx context.Context, orgDID string, policyType RetentionPolicyType) (int, error)
}
