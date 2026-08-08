-- Telegram-style custom chat folders, synced per user (DID).
-- The folder list is stored verbatim as a JSON array of
-- {id, name, order, conversation_ids} objects, keyed by DID (last-write-wins).
CREATE TABLE IF NOT EXISTS chat_folders (
    did        TEXT PRIMARY KEY,
    folders    JSONB NOT NULL DEFAULT '[]'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
