-- migrations/004_evidence_audit.sql
-- Digital Evidence fingerprint records

-- Digital Evidence fingerprint records (Org tier)
CREATE TABLE evidence_fingerprints (
    event_id         TEXT PRIMARY KEY,
    content_hash     TEXT NOT NULL,
    source_type      TEXT NOT NULL, -- media, audit_batch, message, retention_proof
    message_id       TEXT,
    sender_did       TEXT,
    verification_url TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_evidence_message ON evidence_fingerprints(message_id);
CREATE INDEX idx_evidence_sender ON evidence_fingerprints(sender_did);
CREATE INDEX idx_evidence_type ON evidence_fingerprints(source_type);
