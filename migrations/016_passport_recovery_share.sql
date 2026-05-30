-- WO-296: Echo Passport social-threshold recovery metadata (no share bytes, no secrets).
BEGIN;

CREATE TABLE IF NOT EXISTS passport_recovery_policy (
    holder_did      TEXT        PRIMARY KEY,
    threshold_m     INT         NOT NULL,
    total_n         INT         NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT passport_recovery_policy_threshold
        CHECK (threshold_m >= 2 AND total_n >= threshold_m AND total_n <= 255)
);

CREATE TABLE IF NOT EXISTS passport_recovery_share (
    share_id        TEXT        PRIMARY KEY,
    holder_did      TEXT        NOT NULL,
    share_index     INT         NOT NULL,
    guardian_did    TEXT        NOT NULL,
    guardian_role   TEXT        NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'pending',
    guardian_vc_id  TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT passport_recovery_share_role
        CHECK (guardian_role IN ('device', 'contact', 'org')),
    CONSTRAINT passport_recovery_share_status
        CHECK (status IN ('pending', 'active', 'revoked')),
    CONSTRAINT passport_recovery_share_index
        CHECK (share_index >= 1 AND share_index <= 255),
    UNIQUE (holder_did, share_index)
);

CREATE INDEX IF NOT EXISTS idx_passport_recovery_share_holder
    ON passport_recovery_share (holder_did);

CREATE TABLE IF NOT EXISTS passport_recovery_session (
    session_id            TEXT        PRIMARY KEY,
    holder_did            TEXT        NOT NULL,
    status                TEXT        NOT NULL DEFAULT 'initiated',
    required_shares       INT         NOT NULL,
    root_key_commitment   TEXT,
    expires_at            TIMESTAMPTZ NOT NULL,
    completed_at          TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT passport_recovery_session_status
        CHECK (status IN ('initiated', 'completed', 'expired', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_passport_recovery_session_holder
    ON passport_recovery_session (holder_did);

COMMIT;
