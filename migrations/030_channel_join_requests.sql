-- Pending membership requests for approval-gated broadcast channels
-- (channel.require_approval). Approving a request creates a subscriber.
CREATE TABLE IF NOT EXISTS broadcast_channel_join_requests (
    channel_id    TEXT NOT NULL REFERENCES broadcast_channels(id) ON DELETE CASCADE,
    subscriber_id TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'pending',
    requested_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (channel_id, subscriber_id)
);

CREATE INDEX IF NOT EXISTS idx_channel_join_requests_pending
    ON broadcast_channel_join_requests (channel_id, status);
