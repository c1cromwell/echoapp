-- migrations/006_trust_tier_default_zero.sql
-- Per PRD v3.1 AC-INFRA-004.5:
--   New users default to Tier 0 (provisioning in progress).
--   Tier 1 is set by the passkey webhook — last stage of silent provisioning.
--   Existing rows with a registered passkey are backfilled to stay at Tier 1.

BEGIN;

-- Widen the CHECK constraint to allow Tier 0.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_trust_tier_check;
ALTER TABLE users ADD CONSTRAINT users_trust_tier_check CHECK (trust_tier BETWEEN 0 AND 5);

-- Change default from 1 to 0.
ALTER TABLE users ALTER COLUMN trust_tier SET DEFAULT 0;

-- Add display_name and did_minted_at for first-run flow (v3.1 §3.4).
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS display_name      VARCHAR(64),
    ADD COLUMN IF NOT EXISTS did_minted_at     TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS passkey_registered_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_users_did_minted_at ON users(did_minted_at);

-- Backfill: existing users who completed passkey registration stay at Tier 1.
-- For users without the passkey_registered_at column yet, promote any trust_tier = 0
-- row that was created before this migration (they went through the old phone flow).
UPDATE users
   SET trust_tier = 1
 WHERE passkey_registered_at IS NOT NULL
   AND trust_tier = 0;

COMMIT;

-- Rollback (run manually if needed):
-- BEGIN;
-- ALTER TABLE users ALTER COLUMN trust_tier SET DEFAULT 1;
-- ALTER TABLE users DROP CONSTRAINT IF EXISTS users_trust_tier_check;
-- ALTER TABLE users ADD CONSTRAINT users_trust_tier_check CHECK (trust_tier BETWEEN 1 AND 5);
-- COMMIT;
-- Note: display_name, did_minted_at, passkey_registered_at columns left in place.
