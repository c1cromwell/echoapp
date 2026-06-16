package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// --- PostgresDB conversation index (WO-251 eDiscovery) ---

func (p *PostgresDB) ListConversationIDsForParticipant(ctx context.Context, did string) ([]string, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT DISTINCT conversation_id FROM message_queue
		WHERE conversation_id <> '' AND (sender_did = $1 OR recipient_did = $1)`,
		did)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (p *PostgresDB) ListMessageManifest(ctx context.Context, conversationIDs []string, from, to *time.Time) ([]*ExportManifestEntry, error) {
	if len(conversationIDs) == 0 {
		return nil, nil
	}
	q := `
		SELECT message_id, conversation_id, sender_did, recipient_did, queued_at
		FROM message_queue
		WHERE conversation_id = ANY($1)`
	args := []any{conversationIDs}
	argN := 2
	if from != nil {
		q += fmt.Sprintf(" AND queued_at >= $%d", argN)
		args = append(args, *from)
		argN++
	}
	if to != nil {
		q += fmt.Sprintf(" AND queued_at <= $%d", argN)
		args = append(args, *to)
	}
	rows, err := p.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list manifest: %w", err)
	}
	defer rows.Close()
	var out []*ExportManifestEntry
	for rows.Next() {
		var e ExportManifestEntry
		if err := rows.Scan(&e.MessageID, &e.ConversationID, &e.SenderDID, &e.RecipientDID, &e.Timestamp); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

// --- PostgresDB ComplyExtendedStore (WO-251) ---

func (p *PostgresDB) CreateLitigationMatter(ctx context.Context, m *LitigationMatter) error {
	if m == nil {
		return fmt.Errorf("invalid matter")
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO comply_litigation_matters
		  (matter_id, org_did, scope_label, status, custodian_count, activated_at, activated_by_did, data_l1_ref)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		m.MatterID, m.OrgDID, m.ScopeLabel, m.Status, m.CustodianCount, m.ActivatedAt, m.ActivatedByDID, m.DataL1Ref)
	if err != nil {
		return fmt.Errorf("create litigation matter: %w", err)
	}
	return nil
}

func (p *PostgresDB) GetLitigationMatter(ctx context.Context, matterID string) (*LitigationMatter, error) {
	var m LitigationMatter
	err := p.pool.QueryRow(ctx, `
		SELECT matter_id, org_did, COALESCE(scope_label,''), status, custodian_count,
		       activated_at, activated_by_did, released_at, COALESCE(released_by_did,''), COALESCE(data_l1_ref,'')
		FROM comply_litigation_matters WHERE matter_id = $1`, matterID).
		Scan(&m.MatterID, &m.OrgDID, &m.ScopeLabel, &m.Status, &m.CustodianCount,
			&m.ActivatedAt, &m.ActivatedByDID, &m.ReleasedAt, &m.ReleasedByDID, &m.DataL1Ref)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrComplyMatterNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (p *PostgresDB) UpdateLitigationMatter(ctx context.Context, m *LitigationMatter) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE comply_litigation_matters
		SET status = $2, custodian_count = $3, released_at = $4, released_by_did = $5, data_l1_ref = $6
		WHERE matter_id = $1`,
		m.MatterID, m.Status, m.CustodianCount, m.ReleasedAt, m.ReleasedByDID, m.DataL1Ref)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrComplyMatterNotFound
	}
	return nil
}

func (p *PostgresDB) AddLitigationCustodian(ctx context.Context, b *LitigationCustodianBinding) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO comply_litigation_custodians (matter_id, custodian_did, conversation_id)
		VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
		b.MatterID, b.CustodianDID, b.ConversationID)
	return err
}

func (p *PostgresDB) ListLitigationCustodians(ctx context.Context, matterID string) ([]*LitigationCustodianBinding, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT matter_id, custodian_did, conversation_id FROM comply_litigation_custodians WHERE matter_id = $1`,
		matterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*LitigationCustodianBinding
	for rows.Next() {
		var b LitigationCustodianBinding
		if err := rows.Scan(&b.MatterID, &b.CustodianDID, &b.ConversationID); err != nil {
			return nil, err
		}
		out = append(out, &b)
	}
	return out, rows.Err()
}

func (p *PostgresDB) CountActiveLitigationMatters(ctx context.Context, orgDID string) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM comply_litigation_matters WHERE org_did = $1 AND status = 'active'`, orgDID).Scan(&n)
	return n, err
}

func (p *PostgresDB) ListLitigationMatters(ctx context.Context, orgDID string, activeOnly bool) ([]*LitigationMatter, error) {
	q := `
		SELECT matter_id, org_did, COALESCE(scope_label,''), status, custodian_count,
		       activated_at, activated_by_did, released_at, COALESCE(released_by_did,''), COALESCE(data_l1_ref,'')
		FROM comply_litigation_matters WHERE org_did = $1`
	if activeOnly {
		q += ` AND status = 'active'`
	}
	q += ` ORDER BY activated_at DESC`
	rows, err := p.pool.Query(ctx, q, orgDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*LitigationMatter
	for rows.Next() {
		var m LitigationMatter
		if err := rows.Scan(&m.MatterID, &m.OrgDID, &m.ScopeLabel, &m.Status, &m.CustodianCount,
			&m.ActivatedAt, &m.ActivatedByDID, &m.ReleasedAt, &m.ReleasedByDID, &m.DataL1Ref); err != nil {
			return nil, err
		}
		out = append(out, &m)
	}
	return out, rows.Err()
}

func (p *PostgresDB) CreateEDiscoveryExport(ctx context.Context, e *EDiscoveryExport) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO comply_ediscovery_exports
		  (export_id, org_did, matter_id, status, query_hash, message_count, requester_did,
		   date_from, date_to, cover_sheet_ref, data_l1_ref, created_at, ready_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		e.ExportID, e.OrgDID, e.MatterID, e.Status, e.QueryHash, e.MessageCount, e.RequesterDID,
		e.DateFrom, e.DateTo, e.CoverSheetRef, e.DataL1Ref, e.CreatedAt, e.ReadyAt)
	return err
}

func (p *PostgresDB) GetEDiscoveryExport(ctx context.Context, exportID string) (*EDiscoveryExport, error) {
	var e EDiscoveryExport
	err := p.pool.QueryRow(ctx, `
		SELECT export_id, org_did, matter_id, status, query_hash, message_count, requester_did,
		       date_from, date_to, COALESCE(cover_sheet_ref,''), COALESCE(data_l1_ref,''), created_at, ready_at
		FROM comply_ediscovery_exports WHERE export_id = $1`, exportID).
		Scan(&e.ExportID, &e.OrgDID, &e.MatterID, &e.Status, &e.QueryHash, &e.MessageCount, &e.RequesterDID,
			&e.DateFrom, &e.DateTo, &e.CoverSheetRef, &e.DataL1Ref, &e.CreatedAt, &e.ReadyAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrComplyExportNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (p *PostgresDB) UpdateEDiscoveryExport(ctx context.Context, e *EDiscoveryExport) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE comply_ediscovery_exports
		SET status = $2, message_count = $3, cover_sheet_ref = $4, data_l1_ref = $5, ready_at = $6
		WHERE export_id = $1`,
		e.ExportID, e.Status, e.MessageCount, e.CoverSheetRef, e.DataL1Ref, e.ReadyAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrComplyExportNotFound
	}
	return nil
}

func (p *PostgresDB) CountPendingExports(ctx context.Context, orgDID string) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM comply_ediscovery_exports
		WHERE org_did = $1 AND status IN ('pending','processing')`, orgDID).Scan(&n)
	return n, err
}

func (p *PostgresDB) ListEDiscoveryExports(ctx context.Context, orgDID string, limit int) ([]*EDiscoveryExport, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := p.pool.Query(ctx, `
		SELECT export_id, org_did, matter_id, status, query_hash, message_count, requester_did,
		       date_from, date_to, COALESCE(cover_sheet_ref,''), COALESCE(data_l1_ref,''), created_at, ready_at
		FROM comply_ediscovery_exports WHERE org_did = $1 ORDER BY created_at DESC LIMIT $2`, orgDID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanExports(rows)
}

func (p *PostgresDB) AppendAuditEvent(ctx context.Context, e *AuditEvent) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO comply_audit_events (id, org_did, event_type, ref_id, data_l1_ref, occurred_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		e.ID, e.OrgDID, e.EventType, e.RefID, e.DataL1Ref, e.OccurredAt)
	return err
}

func (p *PostgresDB) ListAuditEvents(ctx context.Context, orgDID string, from, to time.Time) ([]*AuditEvent, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, org_did, event_type, COALESCE(ref_id,''), COALESCE(data_l1_ref,''), occurred_at
		FROM comply_audit_events
		WHERE org_did = $1 AND occurred_at >= $2 AND occurred_at <= $3
		ORDER BY occurred_at ASC`, orgDID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*AuditEvent
	for rows.Next() {
		var e AuditEvent
		if err := rows.Scan(&e.ID, &e.OrgDID, &e.EventType, &e.RefID, &e.DataL1Ref, &e.OccurredAt); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

func scanExports(rows pgx.Rows) ([]*EDiscoveryExport, error) {
	defer rows.Close()
	var out []*EDiscoveryExport
	for rows.Next() {
		var e EDiscoveryExport
		if err := rows.Scan(&e.ExportID, &e.OrgDID, &e.MatterID, &e.Status, &e.QueryHash, &e.MessageCount, &e.RequesterDID,
			&e.DateFrom, &e.DateTo, &e.CoverSheetRef, &e.DataL1Ref, &e.CreatedAt, &e.ReadyAt); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}
