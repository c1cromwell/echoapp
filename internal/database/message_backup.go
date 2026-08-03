package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/thechadcromwell/echoapp/pkg/passport"
)

// UpsertMessageBackupBlob stores encrypted message-backup metadata separately
// from passport credential sync data.
func (p *PostgresDB) UpsertMessageBackupBlob(ctx context.Context, record passport.SyncBlobRecord) error {
	_, err := p.pool.Exec(ctx, `
		INSERT INTO message_backup_blob (
			holder_did, storage_uri, content_hash, byte_size, version, ciphertext, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (holder_did) DO UPDATE SET
			storage_uri = EXCLUDED.storage_uri,
			content_hash = EXCLUDED.content_hash,
			byte_size = EXCLUDED.byte_size,
			version = EXCLUDED.version,
			ciphertext = EXCLUDED.ciphertext,
			updated_at = EXCLUDED.updated_at`,
		record.HolderDID, record.StorageURI, record.ContentHash, record.ByteSize,
		record.Version, record.Ciphertext, record.UpdatedAt,
	)
	return err
}

// GetMessageBackupBlob returns the encrypted message-backup record for a holder.
func (p *PostgresDB) GetMessageBackupBlob(ctx context.Context, holderDID string) (*passport.SyncBlobRecord, error) {
	var row passport.SyncBlobRecord
	err := p.pool.QueryRow(ctx, `
		SELECT holder_did, storage_uri, content_hash, byte_size, version, ciphertext, updated_at
		FROM message_backup_blob
		WHERE holder_did = $1`, holderDID).Scan(
		&row.HolderDID, &row.StorageURI, &row.ContentHash, &row.ByteSize,
		&row.Version, &row.Ciphertext, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}
