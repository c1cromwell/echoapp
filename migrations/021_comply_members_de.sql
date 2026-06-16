-- migrations/021_comply_members_de.sql
-- Org membership RBAC + Digital Evidence fingerprint index (message IDs + hash refs only).

CREATE TABLE IF NOT EXISTS comply_org_members (
    org_did    TEXT NOT NULL,
    member_did TEXT NOT NULL,
    role       TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'moderator', 'member')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_did, member_did)
);
CREATE INDEX IF NOT EXISTS idx_comply_org_members_member ON comply_org_members(member_did);

CREATE TABLE IF NOT EXISTS comply_de_fingerprints (
    org_did         TEXT NOT NULL,
    message_id      TEXT NOT NULL,
    fingerprint_ref TEXT NOT NULL,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (org_did, message_id)
);
CREATE INDEX IF NOT EXISTS idx_comply_de_fp_org ON comply_de_fingerprints(org_did);
