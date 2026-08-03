package database

import (
	"context"

	"github.com/thechadcromwell/echoapp/pkg/passport"
)

// MessageBackupSyncStore isolates message-history backup metadata from passport
// credential sync blobs while retaining the shared encrypted SyncStore contract.
type MessageBackupSyncStore struct {
	DB *PostgresDB
}

var _ passport.SyncStore = (*MessageBackupSyncStore)(nil)

func (m *MessageBackupSyncStore) UpsertSyncBlob(ctx context.Context, record passport.SyncBlobRecord) error {
	return m.DB.UpsertMessageBackupBlob(ctx, record)
}

func (m *MessageBackupSyncStore) GetSyncBlob(ctx context.Context, holderDID string) (*passport.SyncBlobRecord, error) {
	return m.DB.GetMessageBackupBlob(ctx, holderDID)
}
