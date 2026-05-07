-- WO-274: durable StatusList2021 slot tracking + L1 publish outbox (Postgres).
BEGIN;

CREATE TABLE IF NOT EXISTS credential_vc_status (
    credential_id      TEXT        NOT NULL PRIMARY KEY,
    issuer_did         TEXT        NOT NULL,
    status_list_index  INT         NOT NULL CHECK (status_list_index >= 0 AND status_list_index < 131072),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_credential_vc_status_issuer
    ON credential_vc_status (issuer_did);

CREATE UNIQUE INDEX IF NOT EXISTS uq_credential_vc_status_issuer_slot
    ON credential_vc_status (issuer_did, status_list_index);

CREATE TABLE IF NOT EXISTS status_list_l1_outbox (
    issuer_did              TEXT        NOT NULL PRIMARY KEY,
    pending_publish         BOOLEAN     NOT NULL DEFAULT FALSE,
    last_published_sequence BIGINT      NOT NULL DEFAULT 0,
    last_publish_at         TIMESTAMPTZ,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
