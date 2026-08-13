-- Roaming per-user app preferences (notifications, appearance, privacy toggles).
-- Stored verbatim as an opaque JSON object keyed by DID (last-write-wins),
-- mirroring chat_folders (029).
CREATE TABLE IF NOT EXISTS user_prefs (
    did        TEXT PRIMARY KEY,
    prefs      JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
