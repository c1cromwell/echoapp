-- migrations/013_phone_discovery_opt_in.sql
-- WO-220: tier-based PSI discoverability with explicit opt-in for Tier 1–2.

BEGIN;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS phone_discovery_opt_in BOOLEAN;

COMMENT ON COLUMN users.phone_discovery_opt_in IS
    'NULL = tier default (Tier 3+ discoverable); TRUE/FALSE = explicit user override';

COMMIT;
