-- migrations/007_did_registry.sql
-- WO-278: persistent storage for did:key registrations served by
-- POST /identity/register (see internal/api/identity_register_handlers.go).
--
-- Phase-1 testnet runs against the in-memory MemoryDIDRegistry. This
-- migration defines the canonical Postgres schema so Phase-2 (and any
-- non-testnet deploy) can flip the wiring without a schema redesign.
--
-- Decision context: docs/adr/0001-phase1-identity-method.md
--
-- Schema notes:
--   * `did` is the W3C did:key identifier (e.g. "did:key:z2dm…"). Maximum
--     observed length for P-256 multibase encodings is ~58 chars; we keep
--     a generous TEXT to leave room for did:key revisions and for
--     forward-compat with did:web / did:plc style longer identifiers
--     should we ever extend the registry.
--   * `public_key_hex` is the SEC1 uncompressed P-256 point (65 bytes,
--     `04` prefix), lowercase hex => 130 chars. Stored as TEXT both for
--     debug ergonomics and so future curves with different point sizes
--     don't require a migration.
--   * `(did, public_key_hex)` is unique together: a re-POST with the
--     identical key is idempotent (handled in the handler), but binding
--     the same DID to a different key MUST be rejected as 409.
--   * `registered_at` is the canonical ordering key for "first time we
--     saw this DID" — not the most-recent-touch timestamp.

BEGIN;

CREATE TABLE IF NOT EXISTS did_registry (
    did            TEXT        PRIMARY KEY,
    public_key_hex TEXT        NOT NULL,
    registered_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT did_registry_did_format_chk
        CHECK (did LIKE 'did:key:z%'),
    CONSTRAINT did_registry_public_key_hex_chk
        CHECK (public_key_hex ~ '^04[0-9a-f]{128}$')
);

-- Lookup helper for the inverse direction (key -> DID), useful for the
-- passkey enrollment flow that knows the public key first and wants to
-- check whether a binding already exists.
CREATE INDEX IF NOT EXISTS idx_did_registry_public_key_hex
    ON did_registry (public_key_hex);

COMMIT;

-- Rollback (run manually if needed):
-- BEGIN;
-- DROP INDEX IF EXISTS idx_did_registry_public_key_hex;
-- DROP TABLE IF EXISTS did_registry;
-- COMMIT;
--
-- Note: migrations/008_identity_device.sql (WO-273) migrates from this table
-- then drops it when upgrading an existing database.
