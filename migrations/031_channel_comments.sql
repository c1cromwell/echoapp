-- Discussion comments on broadcast channel posts.
CREATE TABLE IF NOT EXISTS broadcast_channel_comments (
    id         TEXT PRIMARY KEY,
    channel_id TEXT NOT NULL REFERENCES broadcast_channels(id) ON DELETE CASCADE,
    post_id    TEXT NOT NULL REFERENCES broadcast_channel_posts(id) ON DELETE CASCADE,
    author_id  TEXT NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_channel_comments_post
    ON broadcast_channel_comments (post_id, created_at);
