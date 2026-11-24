-- Anchor evidence moves from Tessellation snapshots to Besu chain proofs
-- (docs/BESU_MIGRATION_PLAN.md). A root is now proven by one self-contained
-- proof: a block header signed by the validators plus Merkle proofs of the
-- anchor contract's storage. The snapshot columns from 036 are dropped; no
-- deployed environment ever held data in them.

ALTER TABLE anchor_roots
    ADD COLUMN IF NOT EXISTS block_number BIGINT,
    ADD COLUMN IF NOT EXISTS tx_hash      TEXT,
    ADD COLUMN IF NOT EXISTS contract     TEXT,
    ADD COLUMN IF NOT EXISTS chain_proof  JSONB;

DROP INDEX IF EXISTS idx_anchor_roots_incomplete;

ALTER TABLE anchor_roots
    DROP COLUMN IF EXISTS ml0_snapshot_ordinal,
    DROP COLUMN IF EXISTS ml0_snapshot_hash,
    DROP COLUMN IF EXISTS ml0_contains_root,
    DROP COLUMN IF EXISTS ml0_signers,
    DROP COLUMN IF EXISTS ml0_archive_key,
    DROP COLUMN IF EXISTS ml0_archive_sha256,
    DROP COLUMN IF EXISTS gl0_snapshot_ordinal,
    DROP COLUMN IF EXISTS gl0_snapshot_hash,
    DROP COLUMN IF EXISTS gl0_signers,
    DROP COLUMN IF EXISTS gl0_archive_key,
    DROP COLUMN IF EXISTS gl0_archive_sha256;

CREATE INDEX IF NOT EXISTS idx_anchor_roots_incomplete
    ON anchor_roots (submitted_at DESC)
    WHERE NOT finalized OR chain_proof IS NULL;
