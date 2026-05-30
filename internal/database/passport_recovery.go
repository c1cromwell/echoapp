package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/thechadcromwell/echoapp/pkg/passport/recovery"
)

func (p *PostgresDB) UpsertPolicy(ctx context.Context, policy recovery.Policy) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO passport_recovery_policy (holder_did, threshold_m, total_n, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (holder_did) DO UPDATE SET
			threshold_m = EXCLUDED.threshold_m,
			total_n = EXCLUDED.total_n,
			updated_at = EXCLUDED.updated_at`,
		policy.HolderDID, policy.Threshold, policy.Total, policy.CreatedAt, policy.UpdatedAt,
	)
	return err
}

func (p *PostgresDB) GetPolicy(ctx context.Context, holderDID string) (*recovery.Policy, error) {
	var policy recovery.Policy
	err := p.pool.QueryRow(ctx, `
		SELECT holder_did, threshold_m, total_n, created_at, updated_at
		FROM passport_recovery_policy WHERE holder_did = $1`, holderDID).Scan(
		&policy.HolderDID, &policy.Threshold, &policy.Total, &policy.CreatedAt, &policy.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

func (p *PostgresDB) ReplaceShareholders(ctx context.Context, holderDID string, shareholders []recovery.Shareholder) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM passport_recovery_share WHERE holder_did = $1`, holderDID); err != nil {
		return err
	}
	for _, sh := range shareholders {
		_, err := tx.Exec(ctx, `
			INSERT INTO passport_recovery_share (
				share_id, holder_did, share_index, guardian_did, guardian_role,
				status, guardian_vc_id, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9)`,
			sh.ShareID, sh.HolderDID, sh.ShareIndex, sh.GuardianDID, sh.Role,
			sh.Status, sh.GuardianVCID, sh.CreatedAt, sh.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (p *PostgresDB) ListShareholders(ctx context.Context, holderDID string) ([]recovery.Shareholder, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT share_id, holder_did, share_index, guardian_did, guardian_role,
		       status, COALESCE(guardian_vc_id, ''), created_at, updated_at
		FROM passport_recovery_share
		WHERE holder_did = $1
		ORDER BY share_index`, holderDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []recovery.Shareholder
	for rows.Next() {
		var sh recovery.Shareholder
		if err := rows.Scan(
			&sh.ShareID, &sh.HolderDID, &sh.ShareIndex, &sh.GuardianDID, &sh.Role,
			&sh.Status, &sh.GuardianVCID, &sh.CreatedAt, &sh.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, sh)
	}
	return out, rows.Err()
}

func (p *PostgresDB) InsertSession(ctx context.Context, session recovery.Session) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO passport_recovery_session (
			session_id, holder_did, status, required_shares, expires_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		session.SessionID, session.HolderDID, session.Status,
		session.RequiredShares, session.ExpiresAt, session.CreatedAt,
	)
	return err
}

func (p *PostgresDB) GetSession(ctx context.Context, holderDID, sessionID string) (*recovery.Session, error) {
	var session recovery.Session
	var completedAt *time.Time
	err := p.pool.QueryRow(ctx, `
		SELECT session_id, holder_did, status, required_shares,
		       COALESCE(root_key_commitment, ''), expires_at, completed_at, created_at
		FROM passport_recovery_session
		WHERE holder_did = $1 AND session_id = $2`, holderDID, sessionID).Scan(
		&session.SessionID, &session.HolderDID, &session.Status, &session.RequiredShares,
		&session.RootKeyCommitment, &session.ExpiresAt, &completedAt, &session.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	session.CompletedAt = completedAt
	return &session, nil
}

func (p *PostgresDB) CompleteSession(ctx context.Context, holderDID, sessionID, commitment string, completedAt time.Time) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE passport_recovery_session
		SET status = 'completed', root_key_commitment = $3, completed_at = $4
		WHERE holder_did = $1 AND session_id = $2 AND status = 'initiated'`,
		holderDID, sessionID, commitment, completedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return recovery.ErrSessionNotFound
	}
	return nil
}
