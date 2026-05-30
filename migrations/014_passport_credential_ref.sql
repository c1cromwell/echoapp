-- WO-293: Echo Passport holder credential references (metadata only — no PII, no VC body).
BEGIN;

CREATE TABLE IF NOT EXISTS passport_credential_ref (
    ref_id              TEXT        PRIMARY KEY,
    holder_did          TEXT        NOT NULL,
    issuer_did          TEXT        NOT NULL,
    credential_type     TEXT        NOT NULL,
    credential_hash     TEXT        NOT NULL,
    status_list_index   INT,
    status_list_cred    TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT passport_credential_ref_hash_len
        CHECK (char_length(credential_hash) = 64)
);

CREATE INDEX IF NOT EXISTS idx_passport_credential_ref_holder
    ON passport_credential_ref (holder_did);

CREATE UNIQUE INDEX IF NOT EXISTS uq_passport_credential_ref_holder_hash
    ON passport_credential_ref (holder_did, credential_hash);

COMMIT;
