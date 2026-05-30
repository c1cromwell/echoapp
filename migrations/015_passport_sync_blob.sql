-- WO-294: Echo Passport client-encrypted credential sync blob (ciphertext is T2 — server holds no keys).
BEGIN;

CREATE TABLE IF NOT EXISTS passport_sync_blob (
    holder_did      TEXT        PRIMARY KEY,
    storage_uri     TEXT        NOT NULL,
    content_hash    TEXT        NOT NULL,
    byte_size       INT         NOT NULL,
    version         INT         NOT NULL DEFAULT 1,
    ciphertext      BYTEA       NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT passport_sync_blob_hash_len
        CHECK (char_length(content_hash) = 64)
);

COMMIT;
