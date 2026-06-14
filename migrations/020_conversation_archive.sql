-- migrations/020_conversation_archive.sql
-- M6 conversation archive flags (WO-198).

CREATE TABLE IF NOT EXISTS conversation_archive (
    conversation_id TEXT PRIMARY KEY,
    archived          BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
