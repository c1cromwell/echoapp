package comply

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/thechadcromwell/echoapp/internal/database"
)

// CreateExportInput starts an eDiscovery export job (WO-251).
type CreateExportInput struct {
	OrgDID        string
	MatterID      string
	RequesterDID  string
	DateFrom      *time.Time
	DateTo        *time.Time
	CustodianDIDs []string
}

// CreateEDiscoveryExport enqueues export processing and returns immediately.
func (s *Service) CreateEDiscoveryExport(ctx context.Context, in CreateExportInput) (*database.EDiscoveryExport, error) {
	if in.OrgDID == "" || in.MatterID == "" || in.RequesterDID == "" {
		return nil, ErrInvalidPolicy
	}
	matter, err := s.store.GetLitigationMatter(ctx, in.MatterID)
	if err != nil || matter.OrgDID != in.OrgDID {
		return nil, database.ErrComplyMatterNotFound
	}
	queryPayload, _ := json.Marshal(map[string]interface{}{
		"matterId":   in.MatterID,
		"from":       in.DateFrom,
		"to":         in.DateTo,
		"custodians": in.CustodianDIDs,
	})
	export := &database.EDiscoveryExport{
		ExportID:     uuid.NewString(),
		OrgDID:       in.OrgDID,
		MatterID:     in.MatterID,
		Status:       database.ExportPending,
		QueryHash:    policyAnchorRef(in.OrgDID, "export_query", in.MatterID, string(queryPayload)),
		RequesterDID: in.RequesterDID,
		DateFrom:     in.DateFrom,
		DateTo:       in.DateTo,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.store.CreateEDiscoveryExport(ctx, export); err != nil {
		return nil, err
	}
	go s.processExport(export.ExportID, in)
	return export, nil
}

func (s *Service) processExport(exportID string, in CreateExportInput) {
	ctx := context.Background()
	export, err := s.store.GetEDiscoveryExport(ctx, exportID)
	if err != nil {
		return
	}
	export.Status = database.ExportProcessing
	_ = s.store.UpdateEDiscoveryExport(ctx, export)

	bindings, _ := s.store.ListLitigationCustodians(ctx, in.MatterID)
	convSet := make(map[string]struct{})
	for _, b := range bindings {
		if len(in.CustodianDIDs) > 0 {
			found := false
			for _, c := range in.CustodianDIDs {
				if c == b.CustodianDID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		convSet[b.ConversationID] = struct{}{}
	}
	convIDs := make([]string, 0, len(convSet))
	for id := range convSet {
		convIDs = append(convIDs, id)
	}

	var manifest []*database.ExportManifestEntry
	if s.convIndex != nil {
		manifest, _ = s.convIndex.ListMessageManifest(ctx, convIDs, in.DateFrom, in.DateTo)
	}
	for _, row := range manifest {
		row.MerkleRef = policyAnchorRef(in.OrgDID, "merkle", row.MessageID, row.ConversationID)
		row.EvidenceEventID = policyAnchorRef(in.OrgDID, "de_event", row.MessageID, row.Timestamp.Format(time.RFC3339))
		_ = s.store.RecordDEFingerprint(ctx, &database.DEFingerprintRecord{
			OrgDID:         in.OrgDID,
			MessageID:      row.MessageID,
			FingerprintRef: row.EvidenceEventID,
		})
	}

	now := time.Now().UTC()
	coverSheet, _ := json.Marshal(map[string]interface{}{
		"exportId":     exportID,
		"matterId":     in.MatterID,
		"messageCount": len(manifest),
		"queryHash":    export.QueryHash,
		"verification": "Verify export checksum against Data L1 anchor reference.",
		"manifestRef":  policyAnchorRef(in.OrgDID, "manifest", exportID, fmt.Sprintf("%d", len(manifest))),
	})
	export.MessageCount = len(manifest)
	export.Status = database.ExportReady
	export.ReadyAt = &now
	export.CoverSheetRef = policyAnchorRef(in.OrgDID, "cover_sheet", exportID, string(coverSheet))
	export.DataL1Ref = policyAnchorRef(in.OrgDID, "export_checksum", exportID, export.CoverSheetRef)
	_ = s.store.UpdateEDiscoveryExport(ctx, export)
	_ = s.store.AppendAuditEvent(ctx, &database.AuditEvent{
		ID:         uuid.NewString(),
		OrgDID:     in.OrgDID,
		EventType:  "ediscovery_export_ready",
		RefID:      exportID,
		DataL1Ref:  export.DataL1Ref,
		OccurredAt: now,
	})
}

// GetEDiscoveryExport returns export job status.
func (s *Service) GetEDiscoveryExport(ctx context.Context, orgDID, exportID string) (*database.EDiscoveryExport, error) {
	export, err := s.store.GetEDiscoveryExport(ctx, exportID)
	if err != nil {
		return nil, err
	}
	if export.OrgDID != orgDID {
		return nil, database.ErrComplyExportNotFound
	}
	return export, nil
}

// ListEDiscoveryExports lists recent exports for an org.
func (s *Service) ListEDiscoveryExports(ctx context.Context, orgDID string) ([]*database.EDiscoveryExport, error) {
	return s.store.ListEDiscoveryExports(ctx, orgDID, 50)
}
