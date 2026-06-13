-- migrations/019_device_sync.sql
-- WO-CA3: content-blind, per-device-addressed message-history sync log.
-- The server stores opaque ciphertext wrapped to a target device's key (pairwise
-- ECDH) and a monotonic seq per (controller_did, target_device_id). It never reads
-- content. Revoking a device closes and purges its stream ("revoke stops sync").

CREATE TABLE IF NOT EXISTS device_sync_log (
    controller_did    TEXT NOT NULL,
    target_device_id  TEXT NOT NULL,
    seq               BIGINT NOT NULL,
    entry_type        TEXT NOT NULL DEFAULT '',  -- opaque hint: message | history | tombstone
    ciphertext        BYTEA NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (controller_did, target_device_id, seq)
);
CREATE INDEX IF NOT EXISTS idx_device_sync_log_pull
    ON device_sync_log(controller_did, target_device_id, seq);

CREATE TABLE IF NOT EXISTS device_sync_revoked (
    controller_did    TEXT NOT NULL,
    target_device_id  TEXT NOT NULL,
    revoked_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (controller_did, target_device_id)
);
