-- migrations/001_core_tables.sql
-- Core tables: users, credentials, trust_scores, relay_registry, logs_index

-- Users table: account metadata (no secrets stored)
CREATE TABLE IF NOT EXISTS users (
    user_id     TEXT PRIMARY KEY,
    did         TEXT UNIQUE NOT NULL,
    username    TEXT UNIQUE NOT NULL,
    trust_tier  INT NOT NULL DEFAULT 1 CHECK (trust_tier BETWEEN 1 AND 5),
    status      TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deactivated')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_did ON users(did);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_trust_tier ON users(trust_tier);

-- Cached trust scores (TTL: 60s, refreshed from Cardano + metagraph)
CREATE TABLE IF NOT EXISTS trust_scores (
    did         TEXT PRIMARY KEY,
    score       INT NOT NULL DEFAULT 0 CHECK (score BETWEEN 0 AND 100),
    tier        INT NOT NULL DEFAULT 1 CHECK (tier BETWEEN 1 AND 5),
    verification_score INT NOT NULL DEFAULT 0,
    interaction_score  INT NOT NULL DEFAULT 0,
    behavior_score     INT NOT NULL DEFAULT 0,
    report_penalty     INT NOT NULL DEFAULT 0,
    issued_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '60 seconds')
);

CREATE INDEX idx_trust_scores_expires ON trust_scores(expires_at);

-- Cached credential status from Cardano
CREATE TABLE IF NOT EXISTS credentials (
    credential_id   TEXT PRIMARY KEY,
    did             TEXT NOT NULL,
    credential_type TEXT NOT NULL CHECK (credential_type IN (
        'proof_of_humanity', 'kyc_lite', 'high_assurance', 'professional'
    )),
    issuer_did      TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'expired', 'suspended')),
    issued_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    metadata        JSONB
);

CREATE INDEX idx_credentials_did ON credentials(did);
CREATE INDEX idx_credentials_type ON credentials(credential_type);
CREATE INDEX idx_credentials_status ON credentials(status);
CREATE INDEX idx_credentials_expires ON credentials(expires_at);

-- Community relay node registry (Phase 4)
CREATE TABLE IF NOT EXISTS relay_registry (
    node_did        TEXT PRIMARY KEY,
    endpoint_url    TEXT NOT NULL,
    stake_amount    BIGINT NOT NULL DEFAULT 0,
    cloud_provider  TEXT,
    uptime_pct      REAL NOT NULL DEFAULT 0 CHECK (uptime_pct BETWEEN 0 AND 100),
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'slashed')),
    registered_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Encrypted log CID index
CREATE TABLE IF NOT EXISTS logs_index (
    batch_id        TEXT PRIMARY KEY,
    cid             TEXT NOT NULL,
    time_range_from TIMESTAMPTZ NOT NULL,
    time_range_to   TIMESTAMPTZ NOT NULL,
    digest          TEXT NOT NULL,
    event_count     INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_logs_index_time ON logs_index(time_range_from, time_range_to);
CREATE INDEX idx_logs_index_cid ON logs_index(cid);

-- User devices (multi-device support)
CREATE TABLE IF NOT EXISTS user_devices (
    device_id       TEXT PRIMARY KEY,
    user_id         TEXT NOT NULL REFERENCES users(user_id),
    did             TEXT NOT NULL,
    device_label    TEXT NOT NULL,
    public_key      TEXT NOT NULL,
    apns_token      TEXT,
    last_seen_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_devices_user ON user_devices(user_id);
CREATE INDEX idx_user_devices_did ON user_devices(did);

-- Notification preferences per user
CREATE TABLE IF NOT EXISTS notification_preferences (
    did                 TEXT PRIMARY KEY,
    push_enabled        BOOLEAN NOT NULL DEFAULT TRUE,
    message_preview     BOOLEAN NOT NULL DEFAULT FALSE,
    group_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    channel_notifications BOOLEAN NOT NULL DEFAULT TRUE,
    quiet_hours_start   TIME,
    quiet_hours_end     TIME,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
