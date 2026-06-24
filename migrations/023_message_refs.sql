-- migrations/023_message_refs.sql
-- M1 reply/forward thread metadata (WO-59). Server stores opaque message-id refs only;
-- reply preview text travels inside client-encrypted payloads (T7).

CREATE TABLE IF NOT EXISTS message_refs (
    message_id                      TEXT PRIMARY KEY,
    conversation_id                 TEXT NOT NULL,
    author_did                      TEXT NOT NULL,
    reply_to_message_id             TEXT,
    forwarded_from_message_id       TEXT,
    forwarded_from_conversation_id  TEXT,
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_message_refs_conversation ON message_refs(conversation_id);
CREATE INDEX IF NOT EXISTS idx_message_refs_reply ON message_refs(reply_to_message_id)
    WHERE reply_to_message_id IS NOT NULL;
