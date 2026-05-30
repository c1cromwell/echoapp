package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/thechadcromwell/echoapp/pkg/passport"
)

// PassportCredentialRefRow is the Postgres row for passport_credential_ref (WO-293).
type PassportCredentialRefRow struct {
	RefID           string
	HolderDID       string
	IssuerDID       string
	CredentialType  string
	CredentialHash  string
	StatusListIndex *int
	StatusListCred  string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func rowToCredentialRef(row PassportCredentialRefRow) passport.CredentialRef {
	return passport.CredentialRef{
		RefID:            row.RefID,
		HolderDID:        row.HolderDID,
		IssuerDID:        row.IssuerDID,
		CredentialType:   row.CredentialType,
		CredentialHash:   row.CredentialHash,
		StatusListIndex:  row.StatusListIndex,
		StatusListCred:   row.StatusListCred,
		RevocationStatus: "unknown",
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
}

func (p *PostgresDB) ListCredentialRefs(ctx context.Context, holderDID string) ([]passport.CredentialRef, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT ref_id, holder_did, issuer_did, credential_type, credential_hash,
		       status_list_index, COALESCE(status_list_cred, ''), created_at, updated_at
		FROM passport_credential_ref
		WHERE holder_did = $1
		ORDER BY created_at DESC`, holderDID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []passport.CredentialRef
	for rows.Next() {
		var row PassportCredentialRefRow
		if err := rows.Scan(
			&row.RefID, &row.HolderDID, &row.IssuerDID, &row.CredentialType,
			&row.CredentialHash, &row.StatusListIndex, &row.StatusListCred,
			&row.CreatedAt, &row.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, rowToCredentialRef(row))
	}
	return out, rows.Err()
}

func (p *PostgresDB) GetCredentialRef(ctx context.Context, holderDID, refID string) (*passport.CredentialRef, error) {
	var row PassportCredentialRefRow
	err := p.pool.QueryRow(ctx, `
		SELECT ref_id, holder_did, issuer_did, credential_type, credential_hash,
		       status_list_index, COALESCE(status_list_cred, ''), created_at, updated_at
		FROM passport_credential_ref
		WHERE holder_did = $1 AND ref_id = $2`, holderDID, refID).Scan(
		&row.RefID, &row.HolderDID, &row.IssuerDID, &row.CredentialType,
		&row.CredentialHash, &row.StatusListIndex, &row.StatusListCred,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ref := rowToCredentialRef(row)
	return &ref, nil
}

func (p *PostgresDB) InsertCredentialRef(ctx context.Context, ref passport.CredentialRef) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO passport_credential_ref (
			ref_id, holder_did, issuer_did, credential_type, credential_hash,
			status_list_index, status_list_cred, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9)`,
		ref.RefID, ref.HolderDID, ref.IssuerDID, ref.CredentialType, ref.CredentialHash,
		ref.StatusListIndex, ref.StatusListCred, ref.CreatedAt, ref.UpdatedAt,
	)
	return err
}

// IsStatusListSlotRevoked checks credential_vc_status for a revoked slot (WO-293 helper).
func (p *PostgresDB) IsStatusListSlotRevoked(ctx context.Context, issuerDID string, statusListIndex int) (bool, error) {
	var revokedAt *time.Time
	err := p.pool.QueryRow(ctx, `
		SELECT revoked_at FROM credential_vc_status
		WHERE issuer_did = $1 AND status_list_index = $2`, issuerDID, statusListIndex).Scan(&revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return revokedAt != nil, nil
}
