package database

import (
	"context"
	"errors"
	"time"
)

// LitigationMatterStatus tracks hold lifecycle (WO-251).
type LitigationMatterStatus string

const (
	MatterActive   LitigationMatterStatus = "active"
	MatterReleased LitigationMatterStatus = "released"
)

// LitigationMatter is org-scoped hold metadata (no message content).
type LitigationMatter struct {
	MatterID       string                 `json:"matterId"`
	OrgDID         string                 `json:"orgDid"`
	ScopeLabel     string                 `json:"scopeLabel,omitempty"`
	Status         LitigationMatterStatus `json:"status"`
	CustodianCount int                    `json:"custodianCount"`
	ActivatedAt    time.Time              `json:"activatedAt"`
	ActivatedByDID string                 `json:"activatedByDid"`
	ReleasedAt     *time.Time             `json:"releasedAt,omitempty"`
	ReleasedByDID  string                 `json:"releasedByDid,omitempty"`
	DataL1Ref      string                 `json:"dataL1Ref,omitempty"`
}

// LitigationCustodianBinding links a custodian conversation under a matter.
type LitigationCustodianBinding struct {
	MatterID       string `json:"matterId"`
	CustodianDID   string `json:"custodianDid"`
	ConversationID string `json:"conversationId"`
}

// ExportStatus tracks eDiscovery export pipeline state (WO-251).
type ExportStatus string

const (
	ExportPending    ExportStatus = "pending"
	ExportProcessing ExportStatus = "processing"
	ExportReady      ExportStatus = "ready"
	ExportDelivered  ExportStatus = "delivered"
	ExportFailed     ExportStatus = "failed"
)

// EDiscoveryExport is export job metadata (encrypted blobs referenced by ID only).
type EDiscoveryExport struct {
	ExportID      string       `json:"exportId"`
	OrgDID        string       `json:"orgDid"`
	MatterID      string       `json:"matterId"`
	Status        ExportStatus `json:"status"`
	QueryHash     string       `json:"queryHash"`
	MessageCount  int          `json:"messageCount"`
	RequesterDID  string       `json:"requesterDid"`
	DateFrom      *time.Time   `json:"dateFrom,omitempty"`
	DateTo        *time.Time   `json:"dateTo,omitempty"`
	CoverSheetRef string       `json:"coverSheetRef,omitempty"`
	DataL1Ref     string       `json:"dataL1Ref,omitempty"`
	CreatedAt     time.Time    `json:"createdAt"`
	ReadyAt       *time.Time   `json:"readyAt,omitempty"`
}

// ExportManifestEntry is a zero-PII row in an export package.
type ExportManifestEntry struct {
	MessageID       string    `json:"messageId"`
	ConversationID  string    `json:"conversationId"`
	SenderDID       string    `json:"senderDid"`
	RecipientDID    string    `json:"recipientDid"`
	Timestamp       time.Time `json:"timestamp"`
	MerkleRef       string    `json:"merkleRef,omitempty"`
	EvidenceEventID string    `json:"evidenceEventId,omitempty"`
}

// AuditEvent is a compliance audit log entry (metadata only).
type AuditEvent struct {
	ID         string    `json:"id"`
	OrgDID     string    `json:"orgDid"`
	EventType  string    `json:"eventType"`
	RefID      string    `json:"refId,omitempty"`
	DataL1Ref  string    `json:"dataL1Ref,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
}

var (
	ErrComplyMatterNotFound = errors.New("litigation matter not found")
	ErrComplyExportNotFound = errors.New("ediscovery export not found")
)

// ComplyExtendedStore adds WO-251 persistence to ComplyStore.
type ComplyExtendedStore interface {
	ComplyStore
	ComplyRBACStore
	CreateLitigationMatter(ctx context.Context, m *LitigationMatter) error
	GetLitigationMatter(ctx context.Context, matterID string) (*LitigationMatter, error)
	UpdateLitigationMatter(ctx context.Context, m *LitigationMatter) error
	AddLitigationCustodian(ctx context.Context, b *LitigationCustodianBinding) error
	ListLitigationCustodians(ctx context.Context, matterID string) ([]*LitigationCustodianBinding, error)
	ListLitigationMatters(ctx context.Context, orgDID string, activeOnly bool) ([]*LitigationMatter, error)
	CountActiveLitigationMatters(ctx context.Context, orgDID string) (int, error)
	CreateEDiscoveryExport(ctx context.Context, e *EDiscoveryExport) error
	GetEDiscoveryExport(ctx context.Context, exportID string) (*EDiscoveryExport, error)
	UpdateEDiscoveryExport(ctx context.Context, e *EDiscoveryExport) error
	CountPendingExports(ctx context.Context, orgDID string) (int, error)
	ListEDiscoveryExports(ctx context.Context, orgDID string, limit int) ([]*EDiscoveryExport, error)
	AppendAuditEvent(ctx context.Context, e *AuditEvent) error
	ListAuditEvents(ctx context.Context, orgDID string, from, to time.Time) ([]*AuditEvent, error)
}

// ConversationIndex lists conversation IDs for a participant DID (relay metadata only).
type ConversationIndex interface {
	ListConversationIDsForParticipant(ctx context.Context, did string) ([]string, error)
	ListMessageManifest(ctx context.Context, conversationIDs []string, from, to *time.Time) ([]*ExportManifestEntry, error)
}
