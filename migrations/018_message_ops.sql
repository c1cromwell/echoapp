-- migrations/018_message_ops.sql
-- M1 message operations (WO-25 edit history, WO-84 delete, WO-59 pins).
-- Hybrid model: edit history is only persisted for conversations flagged for
-- retention (Comply / litigation hold); content is opaque ciphertext, never plaintext.

-- Per-conversation retention / litigation-hold flag (Comply sets this).
CREATE TABLE IF NOT EXISTS conversation_retention (
    conversation_id          TEXT PRIMARY KEY,
    retained                 BOOLEAN NOT NULL DEFAULT FALSE,
    disappearing_ttl_seconds INT NOT NULL DEFAULT 0,  -- 0 = disappearing off
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Immutable edit versions (only written for retained conversations).
CREATE TABLE IF NOT EXISTS message_edits (
    message_id      TEXT NOT NULL,
    version         INT NOT NULL,
    conversation_id TEXT NOT NULL,
    editor_did      TEXT NOT NULL,
    ciphertext      BYTEA NOT NULL,
    edited_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, version)
);
CREATE INDEX IF NOT EXISTS idx_message_edits_message ON message_edits(message_id);
CREATE INDEX IF NOT EXISTS idx_message_edits_conversation ON message_edits(conversation_id);

-- Synchronized-delete tombstones (WO-84). `retained` preserves history for eDiscovery.
CREATE TABLE IF NOT EXISTS message_tombstones (
    message_id  TEXT PRIMARY KEY,
    deleted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    retained    BOOLEAN NOT NULL DEFAULT FALSE
);

-- Pinned messages (WO-59), max 5 per conversation (enforced in the store).
CREATE TABLE IF NOT EXISTS message_pins (
    conversation_id TEXT NOT NULL,
    message_id      TEXT NOT NULL,
    pinner_did      TEXT NOT NULL,
    pinned_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (conversation_id, message_id)
);
CREATE INDEX IF NOT EXISTS idx_message_pins_conversation ON message_pins(conversation_id);
