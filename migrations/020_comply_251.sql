-- migrations/020_comply_251.sql
-- WO-251 litigation holds and eDiscovery export metadata (zero message content).

CREATE TABLE IF NOT EXISTS comply_litigation_matters (
    matter_id        TEXT PRIMARY KEY,
    org_did          TEXT NOT NULL,
    scope_label      TEXT,
    status           TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'released')),
    custodian_count  INT NOT NULL DEFAULT 0,
    activated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    activated_by_did TEXT NOT NULL,
    released_at      TIMESTAMPTZ,
    released_by_did  TEXT,
    data_l1_ref      TEXT
);
CREATE INDEX IF NOT EXISTS idx_comply_matters_org ON comply_litigation_matters(org_did, status);

CREATE TABLE IF NOT EXISTS comply_litigation_custodians (
    matter_id        TEXT NOT NULL REFERENCES comply_litigation_matters(matter_id) ON DELETE CASCADE,
    custodian_did    TEXT NOT NULL,
    conversation_id  TEXT NOT NULL,
    PRIMARY KEY (matter_id, custodian_did, conversation_id)
);
CREATE INDEX IF NOT EXISTS idx_comply_custodian_did ON comply_litigation_custodians(custodian_did);

CREATE TABLE IF NOT EXISTS comply_ediscovery_exports (
    export_id        TEXT PRIMARY KEY,
    org_did          TEXT NOT NULL,
    matter_id        TEXT NOT NULL REFERENCES comply_litigation_matters(matter_id),
    status           TEXT NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending', 'processing', 'ready', 'delivered', 'failed')),
    query_hash       TEXT NOT NULL,
    message_count    INT NOT NULL DEFAULT 0,
    requester_did    TEXT NOT NULL,
    date_from        TIMESTAMPTZ,
    date_to          TIMESTAMPTZ,
    cover_sheet_ref  TEXT,
    data_l1_ref      TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ready_at         TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_comply_exports_org ON comply_ediscovery_exports(org_did, status);

CREATE TABLE IF NOT EXISTS comply_audit_events (
    id           TEXT PRIMARY KEY,
    org_did      TEXT NOT NULL,
    event_type   TEXT NOT NULL,
    ref_id       TEXT,
    data_l1_ref  TEXT,
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_comply_audit_org_time ON comply_audit_events(org_did, occurred_at DESC);
