package comply

import (
	"context"
	"fmt"
	"time"

	"github.com/thechadcromwell/echoapp/internal/database"
	"github.com/thechadcromwell/echoapp/internal/evidence"
)

// EnsureOrgAdmin upserts an admin member (service/bootstrap path).
func (s *Service) EnsureOrgAdmin(ctx context.Context, orgDID, memberDID string) error {
	if orgDID == "" || memberDID == "" {
		return nil
	}
	return s.store.UpsertOrgMember(ctx, &database.OrgMember{
		OrgDID:    orgDID,
		MemberDID: memberDID,
		Role:      database.OrgRoleAdmin,
	})
}

// AuthorizeOrgRead checks membership for user-authenticated read paths.
func (s *Service) AuthorizeOrgRead(ctx context.Context, orgDID, actorDID string) error {
	if actorDID == "" {
		return database.ErrComplyOrgForbidden
	}
	mem, err := s.store.GetOrgMember(ctx, orgDID, actorDID)
	if err != nil {
		return err
	}
	if !mem.Role.CanReadComply() {
		return database.ErrComplyOrgForbidden
	}
	return nil
}

// AuthorizeOrgWrite checks admin/owner for mutation paths.
func (s *Service) AuthorizeOrgWrite(ctx context.Context, orgDID, actorDID string) error {
	if actorDID == "" {
		return database.ErrComplyOrgForbidden
	}
	mem, err := s.store.GetOrgMember(ctx, orgDID, actorDID)
	if err != nil {
		return err
	}
	if !mem.Role.CanWriteComply() {
		return database.ErrComplyOrgForbidden
	}
	return nil
}

func (s *Service) deCoverageRate(ctx context.Context, orgDID string) string {
	total, err := s.store.CountOrgScopedMessages(ctx, orgDID)
	if err != nil || total == 0 {
		return "0%"
	}
	fp, err := s.store.CountDEFingerprints(ctx, orgDID)
	if err != nil {
		return "0%"
	}
	if fp > total {
		fp = total
	}
	pct := fp * 100 / total
	return fmt.Sprintf("%d%%", pct)
}

func (s *Service) fingerprintConversationMessages(ctx context.Context, orgDID, convID string) {
	if s.convIndex == nil {
		return
	}
	entries, err := s.convIndex.ListMessageManifest(ctx, []string{convID}, nil, nil)
	if err != nil {
		return
	}
	for _, e := range entries {
		ref := s.recordDEFingerprint(ctx, orgDID, e)
		if ref == "" {
			ref = policyAnchorRef(orgDID, "de_fingerprint", e.MessageID, e.ConversationID)
		}
		_ = s.store.RecordDEFingerprint(ctx, &database.DEFingerprintRecord{
			OrgDID:         orgDID,
			MessageID:      e.MessageID,
			FingerprintRef: ref,
		})
	}
}

func (s *Service) recordDEFingerprint(ctx context.Context, orgDID string, e *database.ExportManifestEntry) string {
	if s.evidence == nil || s.evidenceConfig == nil {
		return ""
	}
	contentHash := policyAnchorRef(orgDID, "de_fingerprint", e.MessageID, e.ConversationID)
	req, err := evidence.SmartCheckmarkFingerprint(
		s.evidenceConfig,
		contentHash,
		e.Timestamp.UTC().Format(time.RFC3339),
		e.SenderDID,
	)
	if err != nil {
		return ""
	}
	resp, err := s.evidence.SubmitFingerprint(ctx, req)
	if err != nil || resp == nil {
		return ""
	}
	if resp.EventID != "" {
		return resp.EventID
	}
	return resp.ExplorerURL
}
