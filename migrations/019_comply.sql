-- migrations/019_comply.sql
-- WO-250 ECHO Comply: org-scoped retention policies and conversation bindings.
-- Server stores policy metadata only — no message plaintext (T0).

CREATE TABLE IF NOT EXISTS comply_retention_policies (
    id               TEXT PRIMARY KEY,
    org_did          TEXT NOT NULL,
    policy_type      TEXT NOT NULL CHECK (policy_type IN ('permanent', 'time_limited', 'litigation_hold')),
    conversation_id  TEXT,
    scope_label      TEXT,
    effective_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at       TIMESTAMPTZ,
    data_l1_ref      TEXT,
    active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_by_did   TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_comply_policies_org ON comply_retention_policies(org_did, active);
CREATE INDEX IF NOT EXISTS idx_comply_policies_conv ON comply_retention_policies(conversation_id) WHERE conversation_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS comply_conversation_bindings (
    conversation_id TEXT PRIMARY KEY,
    policy_id       TEXT NOT NULL REFERENCES comply_retention_policies(id),
    org_did         TEXT NOT NULL,
    bound_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS comply_org_profiles (
    org_did    TEXT PRIMARY KEY,
    tier       TEXT NOT NULL DEFAULT 'starter',
    seats      INT NOT NULL DEFAULT 10,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
