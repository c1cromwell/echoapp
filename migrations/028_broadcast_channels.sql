-- Durable broadcast channels / communities (replaces in-memory ChannelService maps).
BEGIN;

CREATE TABLE IF NOT EXISTS broadcast_channels (
    id                   TEXT PRIMARY KEY,
    creator_id           TEXT NOT NULL,
    name                 TEXT NOT NULL,
    topic                TEXT NOT NULL DEFAULT '',
    description          TEXT NOT NULL DEFAULT '',
    visibility_mode      TEXT NOT NULL DEFAULT 'public',
    channel_type         TEXT NOT NULL DEFAULT 'community',
    cover_image_url      TEXT NOT NULL DEFAULT '',
    tags                 JSONB NOT NULL DEFAULT '[]'::jsonb,
    is_active            BOOLEAN NOT NULL DEFAULT TRUE,
    is_muted             BOOLEAN NOT NULL DEFAULT FALSE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_post_at         TIMESTAMPTZ,
    subscriber_count     BIGINT NOT NULL DEFAULT 0,
    total_post_count     BIGINT NOT NULL DEFAULT 0,
    language             TEXT NOT NULL DEFAULT 'en',
    website              TEXT NOT NULL DEFAULT '',
    trust_score          DOUBLE PRECISION NOT NULL DEFAULT 0,
    verification_status  TEXT NOT NULL DEFAULT 'unverified',
    allow_comments       BOOLEAN NOT NULL DEFAULT TRUE,
    allow_polls          BOOLEAN NOT NULL DEFAULT TRUE,
    allow_links          BOOLEAN NOT NULL DEFAULT TRUE,
    allow_media          BOOLEAN NOT NULL DEFAULT TRUE,
    require_approval     BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_broadcast_channels_visibility
    ON broadcast_channels (visibility_mode, is_active);
CREATE INDEX IF NOT EXISTS idx_broadcast_channels_creator
    ON broadcast_channels (creator_id);

CREATE TABLE IF NOT EXISTS broadcast_channel_posts (
    id                TEXT PRIMARY KEY,
    channel_id        TEXT NOT NULL REFERENCES broadcast_channels(id) ON DELETE CASCADE,
    creator_id        TEXT NOT NULL,
    content           TEXT NOT NULL DEFAULT '',
    content_type      TEXT NOT NULL DEFAULT 'text',
    encrypted_content BYTEA,
    like_count        BIGINT NOT NULL DEFAULT 0,
    comment_count     BIGINT NOT NULL DEFAULT 0,
    share_count       BIGINT NOT NULL DEFAULT 0,
    published_at      TIMESTAMPTZ,
    scheduled_for     TIMESTAMPTZ,
    publish_status    TEXT NOT NULL DEFAULT 'published',
    is_pinned         BOOLEAN NOT NULL DEFAULT FALSE,
    is_featured       BOOLEAN NOT NULL DEFAULT FALSE,
    is_sponsored      BOOLEAN NOT NULL DEFAULT FALSE,
    allow_replies     BOOLEAN NOT NULL DEFAULT TRUE,
    flag_count        INT NOT NULL DEFAULT 0,
    mod_status        TEXT NOT NULL DEFAULT 'approved',
    mod_notes         TEXT NOT NULL DEFAULT '',
    edit_count        INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_broadcast_posts_channel_created
    ON broadcast_channel_posts (channel_id, publish_status, created_at DESC);

CREATE TABLE IF NOT EXISTS broadcast_channel_subscribers (
    id                       TEXT NOT NULL,
    channel_id               TEXT NOT NULL REFERENCES broadcast_channels(id) ON DELETE CASCADE,
    subscriber_id            TEXT NOT NULL,
    joined_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    subscription_tier        TEXT NOT NULL DEFAULT 'free',
    subscription_expires_at  TIMESTAMPTZ,
    auto_renew               BOOLEAN NOT NULL DEFAULT FALSE,
    role                     TEXT NOT NULL DEFAULT 'subscriber',
    last_seen_at             TIMESTAMPTZ,
    notification_mode        TEXT NOT NULL DEFAULT 'all',
    is_muted                 BOOLEAN NOT NULL DEFAULT FALSE,
    is_blocked               BOOLEAN NOT NULL DEFAULT FALSE,
    post_count               BIGINT NOT NULL DEFAULT 0,
    comment_count            BIGINT NOT NULL DEFAULT 0,
    like_count               BIGINT NOT NULL DEFAULT 0,
    trust_score              DOUBLE PRECISION NOT NULL DEFAULT 0,
    moderation_flags         INT NOT NULL DEFAULT 0,
    PRIMARY KEY (channel_id, subscriber_id)
);

CREATE INDEX IF NOT EXISTS idx_broadcast_subs_subscriber
    ON broadcast_channel_subscribers (subscriber_id);

COMMIT;
