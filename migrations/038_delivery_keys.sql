-- Delivery keys: sealed-sender delivery without the server learning the sender
-- (docs/RELAY_METADATA_REVIEW.md §5).
--
-- An account gives a random delivery key to its contacts inside encrypted
-- messages. The server stores only SHA-256(key) and the account it delivers
-- to. A sender presents the key on an unauthenticated request; nothing here,
-- or anywhere else, records who sent what.

CREATE TABLE IF NOT EXISTS delivery_keys (
    key_hash   TEXT PRIMARY KEY,
    did        TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_delivery_keys_did ON delivery_keys (did, created_at DESC);
