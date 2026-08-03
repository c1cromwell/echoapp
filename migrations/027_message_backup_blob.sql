-- Wave S1 / WO-CA2: dedicated message-history backup blob (isolates from passport_sync_blob).
-- Ciphertext is client-encrypted; server holds no keys (T2).
BEGIN;

CREATE TABLE IF NOT EXISTS message_backup_blob (
    holder_did      TEXT        PRIMARY KEY,
    storage_uri     TEXT        NOT NULL,
    content_hash    TEXT        NOT NULL,
    byte_size       INT         NOT NULL,
    version         INT         NOT NULL DEFAULT 1,
    ciphertext      BYTEA       NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT message_backup_blob_hash_len
        CHECK (char_length(content_hash) = 64)
);

COMMIT;
