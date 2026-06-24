-- migrations/022_groups.sql
-- Persistent group metadata + membership (M2 / WO-207). Message content stays E2EE client-side.

CREATE TABLE IF NOT EXISTS echo_groups (
    group_id            TEXT PRIMARY KEY,
    owner_id            TEXT NOT NULL,
    group_type          TEXT NOT NULL CHECK (group_type IN ('public', 'private', 'secret')),
    name                TEXT NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    avatar              TEXT NOT NULL DEFAULT '',
    category            TEXT NOT NULL DEFAULT 'other',
    tags                JSONB NOT NULL DEFAULT '[]',
    rules               TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    requirements        JSONB NOT NULL DEFAULT '{}',
    max_members         INT NOT NULL DEFAULT 100,
    current_members     INT NOT NULL DEFAULT 0,
    max_admins          INT NOT NULL DEFAULT 5,
    max_moderators      INT NOT NULL DEFAULT 10,
    settings            JSONB NOT NULL DEFAULT '{}',
    permissions         JSONB NOT NULL DEFAULT '{}',
    governance          JSONB NOT NULL DEFAULT '{}',
    stats               JSONB NOT NULL DEFAULT '{}',
    creation_tx_hash    TEXT NOT NULL DEFAULT '',
    last_update_tx_hash TEXT NOT NULL DEFAULT '',
    snapshot_id         TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_echo_groups_owner ON echo_groups(owner_id);

CREATE TABLE IF NOT EXISTS echo_group_members (
    group_id            TEXT NOT NULL REFERENCES echo_groups(group_id) ON DELETE CASCADE,
    member_id           TEXT NOT NULL,
    display_name        TEXT NOT NULL DEFAULT '',
    avatar              TEXT NOT NULL DEFAULT '',
    persona_id          TEXT NOT NULL DEFAULT '',
    role                TEXT NOT NULL,
    permissions         JSONB NOT NULL DEFAULT '[]',
    trust_score         INT NOT NULL DEFAULT 0,
    trust_level         TEXT NOT NULL DEFAULT 'newcomer',
    badges              JSONB NOT NULL DEFAULT '[]',
    verified_at         TIMESTAMPTZ,
    joined_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    message_count       INT NOT NULL DEFAULT 0,
    warning_count       INT NOT NULL DEFAULT 0,
    is_muted            BOOLEAN NOT NULL DEFAULT FALSE,
    muted_until         TIMESTAMPTZ,
    is_banned           BOOLEAN NOT NULL DEFAULT FALSE,
    notification_level  TEXT NOT NULL DEFAULT 'all',
    nickname            TEXT NOT NULL DEFAULT '',
    show_trust_score     BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (group_id, member_id)
);

CREATE INDEX IF NOT EXISTS idx_echo_group_members_member ON echo_group_members(member_id);
