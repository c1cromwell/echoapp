package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (p *PostgresDB) UpsertOrgMember(ctx context.Context, member *OrgMember) error {
	created := member.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO comply_org_members (org_did, member_did, role, created_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (org_did, member_did) DO UPDATE SET role = EXCLUDED.role`,
		member.OrgDID, member.MemberDID, member.Role, created)
	return err
}

func (p *PostgresDB) GetOrgMember(ctx context.Context, orgDID, memberDID string) (*OrgMember, error) {
	var m OrgMember
	err := p.pool.QueryRow(ctx, `
		SELECT org_did, member_did, role, created_at FROM comply_org_members
		WHERE org_did = $1 AND member_did = $2`, orgDID, memberDID).
		Scan(&m.OrgDID, &m.MemberDID, &m.Role, &m.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrComplyOrgForbidden
		}
		return nil, err
	}
	return &m, nil
}

func (p *PostgresDB) RecordDEFingerprint(ctx context.Context, rec *DEFingerprintRecord) error {
	recorded := rec.RecordedAt
	if recorded.IsZero() {
		recorded = time.Now().UTC()
	}
	_, err := p.pool.Exec(ctx, `
		INSERT INTO comply_de_fingerprints (org_did, message_id, fingerprint_ref, recorded_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (org_did, message_id) DO UPDATE SET fingerprint_ref = EXCLUDED.fingerprint_ref`,
		rec.OrgDID, rec.MessageID, rec.FingerprintRef, recorded)
	return err
}

func (p *PostgresDB) CountDEFingerprints(ctx context.Context, orgDID string) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM comply_de_fingerprints WHERE org_did = $1`, orgDID).Scan(&n)
	return n, err
}

func (p *PostgresDB) CountOrgScopedMessages(ctx context.Context, orgDID string) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT mq.message_id)
		FROM message_queue mq
		INNER JOIN comply_conversation_bindings b ON b.conversation_id = mq.conversation_id
		WHERE b.org_did = $1 AND mq.conversation_id <> ''`, orgDID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count org messages: %w", err)
	}
	return n, nil
}
