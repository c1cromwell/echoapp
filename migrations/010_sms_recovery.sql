-- Wave 12: SMS recovery phone-hash registration.
-- Stores H(E.164 phone) alongside each DID for account recovery.
-- The raw phone number is NEVER stored here — only the SHA-256 commitment.

CREATE TABLE IF NOT EXISTS sms_recovery (
    did        TEXT PRIMARY KEY REFERENCES users(did) ON DELETE CASCADE,
    phone_hash TEXT NOT NULL,       -- "sha256:" + lowercase hex
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sms_recovery_phone_hash ON sms_recovery (phone_hash);
