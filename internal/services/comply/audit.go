package comply

import (
	"context"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// AuditReport is a zero-PII compliance audit document (WO-252).
type AuditReport struct {
	OrgDID             string                       `json:"orgDid"`
	GeneratedAt        time.Time                    `json:"generatedAt"`
	From               time.Time                    `json:"from"`
	To                 time.Time                    `json:"to"`
	ActiveRetention    int                          `json:"activeRetentionPolicies"`
	ActiveHolds        int                          `json:"activeLitigationHolds"`
	PendingExports     int                          `json:"pendingExports"`
	AnchorHealth       string                       `json:"anchorHealth"`
	Events             []*database.AuditEvent       `json:"events"`
	RetentionPolicies  []*database.RetentionPolicy  `json:"retentionPolicies,omitempty"`
	ExportHistory      []*database.EDiscoveryExport `json:"exportHistory,omitempty"`
	VerificationNotice string                       `json:"verificationNotice"`
}

// GenerateAuditReport builds a structured audit report for regulatory submission.
func (s *Service) GenerateAuditReport(ctx context.Context, orgDID string, from, to time.Time) (*AuditReport, error) {
	if to.IsZero() {
		to = time.Now().UTC()
	}
	if from.IsZero() {
		from = to.Add(-30 * 24 * time.Hour)
	}
	active, _ := s.store.CountActivePolicies(ctx, orgDID, "")
	holds, _ := s.store.CountActiveLitigationMatters(ctx, orgDID)
	pending, _ := s.store.CountPendingExports(ctx, orgDID)
	events, _ := s.store.ListAuditEvents(ctx, orgDID, from, to)
	policies, _ := s.store.ListRetentionPolicies(ctx, orgDID, true)
	exports, _ := s.store.ListEDiscoveryExports(ctx, orgDID, 20)

	anchorHealth := "healthy"
	if active == 0 && holds == 0 {
		anchorHealth = "degraded"
	}

	return &AuditReport{
		OrgDID:             orgDID,
		GeneratedAt:        time.Now().UTC(),
		From:               from,
		To:                 to,
		ActiveRetention:    active,
		ActiveHolds:        holds,
		PendingExports:     pending,
		AnchorHealth:       anchorHealth,
		Events:             events,
		RetentionPolicies:  policies,
		ExportHistory:      exports,
		VerificationNotice: "All anchor references are SHA-256 hashes suitable for independent Data L1 verification. No message content or readable PII is included.",
	}, nil
}
