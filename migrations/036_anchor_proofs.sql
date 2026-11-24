-- Durable message-anchoring record (self-contained proofs).
--
-- The metagraph's merkle-root index lives in memory on the chain side, and the
-- gateway previously kept proofs only in Redis/memory. These tables let anchor
-- proofs and finality survive a gateway restart and stay checkable with the
-- chain switched off.

-- One row per Merkle root submitted to Data L1.
--
-- ml0_* / gl0_* describe the signed snapshots fetched when (or after) the root
-- was observed finalized. They are evidence captured at observation time:
--   * ml0_contains_root = TRUE means the root was found inside the archived
--     Metagraph L0 snapshot's data-application on-chain state.
--   * ml0_contains_root = FALSE/NULL means only "observed finalized at or
--     before this ordinal".
--   * The Global L0 snapshot is always only "observed at or before": nothing
--     here links it to the specific metagraph snapshot.
-- *_archive_sha256 is the SHA-256 of the archived JSON body, not the
-- Tessellation snapshot hash (*_snapshot_hash, as reported by the node).
CREATE TABLE IF NOT EXISTS anchor_roots (
    merkle_root          TEXT PRIMARY KEY,
    network              TEXT NOT NULL,
    leaf_count           INTEGER NOT NULL,
    submitted_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finalized            BOOLEAN NOT NULL DEFAULT FALSE,
    finalized_at         TIMESTAMPTZ,
    ml0_snapshot_ordinal BIGINT,
    ml0_snapshot_hash    TEXT,
    ml0_contains_root    BOOLEAN,
    ml0_signers          JSONB,
    ml0_archive_key      TEXT,
    ml0_archive_sha256   TEXT,
    gl0_snapshot_ordinal BIGINT,
    gl0_snapshot_hash    TEXT,
    gl0_signers          JSONB,
    gl0_archive_key      TEXT,
    gl0_archive_sha256   TEXT
);

-- The finality tracker scans for roots that still need work, newest first.
CREATE INDEX IF NOT EXISTS idx_anchor_roots_incomplete
    ON anchor_roots (submitted_at DESC)
    WHERE NOT finalized OR ml0_snapshot_ordinal IS NULL;

-- One row per anchored message: its leaf and Merkle path to merkle_root.
-- commitment is a hash; no message content or DID is stored here.
-- submit_tx_ref is the Data L1 submission reference returned at anchoring time
-- (served as "snapshotHash" by GET /v1/messages/{id}/merkle-proof).
CREATE TABLE IF NOT EXISTS anchor_message_proofs (
    message_id    TEXT PRIMARY KEY,
    commitment    TEXT NOT NULL,
    siblings      JSONB NOT NULL DEFAULT '[]'::jsonb,
    leaf_index    INTEGER NOT NULL,
    merkle_root   TEXT NOT NULL,
    submit_tx_ref TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_anchor_message_proofs_root ON anchor_message_proofs (merkle_root);
