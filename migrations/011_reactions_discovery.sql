-- Phase 3 durable message reactions + Phase 2 (D2) OPRF contact-discovery index.
-- Replaces the previous in-memory stores so reactions and discovery survive
-- restarts and are shared across instances.

-- One emoji per (message, reactor); re-reacting replaces via the PK upsert.
CREATE TABLE IF NOT EXISTS message_reactions (
    message_id  TEXT NOT NULL,
    reactor_did TEXT NOT NULL,
    emoji       TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, reactor_did)
);

CREATE INDEX IF NOT EXISTS idx_message_reactions_message ON message_reactions (message_id);

-- OPRF contact-discovery index: hex(OPRF_k(phone)) -> DID.
-- Raw phone numbers are NEVER stored here — only the OPRF output.
-- Cascades on account deletion so a deleted user drops out of discovery.
CREATE TABLE IF NOT EXISTS contact_discovery_index (
    oprf_key   TEXT PRIMARY KEY,
    did        TEXT NOT NULL REFERENCES users(did) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_contact_discovery_did ON contact_discovery_index (did);
