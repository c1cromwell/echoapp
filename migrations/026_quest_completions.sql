-- WO-271 quest completion state (Data L1 mirror for API)
CREATE TABLE IF NOT EXISTS quest_completions (
    did         TEXT NOT NULL,
    quest_id    TEXT NOT NULL,
    completed_at TIMESTAMPTZ,
    reward_claimed BOOLEAN NOT NULL DEFAULT FALSE,
    tx_hash     TEXT,
    PRIMARY KEY (did, quest_id)
);

CREATE INDEX IF NOT EXISTS idx_quest_completions_did ON quest_completions (did);
