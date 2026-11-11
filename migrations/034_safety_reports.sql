-- Abuse reports filed from the app (contact, message, channel).
-- Reason is a short code. The report text is not a message body.
CREATE TABLE IF NOT EXISTS safety_reports (
    id           TEXT PRIMARY KEY,
    reporter_did TEXT NOT NULL,
    subject_did  TEXT NOT NULL,
    reason       TEXT NOT NULL,
    context      TEXT NOT NULL DEFAULT 'contact',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_safety_reports_reporter ON safety_reports (reporter_did);
CREATE INDEX IF NOT EXISTS idx_safety_reports_subject ON safety_reports (subject_did);
