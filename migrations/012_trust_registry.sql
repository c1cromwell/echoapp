-- migrations/012_trust_registry.sql
-- WO-118: durable trust registry for verifiable credential issuer verification.

BEGIN;

CREATE TABLE IF NOT EXISTS trusted_issuers (
    issuer_id                    TEXT        PRIMARY KEY,
    name                         TEXT        NOT NULL,
    did                          TEXT        NOT NULL UNIQUE,
    issuer_type                  TEXT        NOT NULL DEFAULT 'identity_provider',
    jurisdiction                 TEXT        NOT NULL DEFAULT 'global',
    trust_level                  TEXT        NOT NULL DEFAULT 'basic',
    status                       TEXT        NOT NULL DEFAULT 'active',
    credential_types             TEXT[]      NOT NULL DEFAULT '{}',
    verification_public_key_b64  TEXT,
    public_key_url               TEXT,
    risk_score                   INT         NOT NULL DEFAULT 50,
    onboarding_weight            INT         NOT NULL DEFAULT 10,
    activation_threshold         DOUBLE PRECISION NOT NULL DEFAULT 0.5,
    contact_email                TEXT,
    documentation_url            TEXT,
    established_at               TIMESTAMPTZ,
    verified_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_verified_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    suspended_at                 TIMESTAMPTZ,
    revoked_at                   TIMESTAMPTZ,
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT trusted_issuers_status_chk
        CHECK (status IN ('active', 'suspended', 'revoked', 'pending'))
);

CREATE INDEX IF NOT EXISTS idx_trusted_issuers_did ON trusted_issuers (did);
CREATE INDEX IF NOT EXISTS idx_trusted_issuers_status ON trusted_issuers (status);

COMMIT;
