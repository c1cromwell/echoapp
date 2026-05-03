-- migrations/008_identity_device.sql
-- WO-273: multi-device did:key rows (same subject DID, multiple device keys).
-- Migrates legacy did_registry rows from migration 007 when present.

BEGIN;

CREATE TABLE IF NOT EXISTS identity_device (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          TEXT        NULL,
    did              TEXT        NOT NULL,
    public_key_hex   TEXT        NOT NULL,
    device_label     TEXT        NOT NULL DEFAULT 'primary',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT identity_device_did_chk
        CHECK (did LIKE 'did:key:z%'),
    CONSTRAINT identity_device_public_key_hex_chk
        CHECK (public_key_hex ~ '^04[0-9a-f]{128}$'),
    CONSTRAINT identity_device_did_pubkey_unique UNIQUE (did, public_key_hex)
);

CREATE INDEX IF NOT EXISTS idx_identity_device_did
    ON identity_device (did);

CREATE INDEX IF NOT EXISTS idx_identity_device_public_key_hex
    ON identity_device (public_key_hex);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'did_registry'
    ) THEN
        INSERT INTO identity_device (did, public_key_hex, device_label, created_at)
        SELECT did, public_key_hex, 'primary', registered_at
        FROM did_registry
        ON CONFLICT (did, public_key_hex) DO NOTHING;
        DROP TABLE did_registry;
    END IF;
END
$$;

COMMIT;

-- Rollback (manual):
-- BEGIN;
-- CREATE TABLE IF NOT EXISTS did_registry (...);
-- INSERT ... SELECT ... from identity_device WHERE device_label = 'primary';
-- DROP TABLE identity_device;
-- COMMIT;
