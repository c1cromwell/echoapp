-- migrations/024_wallet_accounts.sql
-- Links did:key identity to Constellation DAG wallet address (T7 public).

CREATE TABLE IF NOT EXISTS wallet_accounts (
    did          TEXT PRIMARY KEY,
    dag_address  TEXT NOT NULL,
    linked_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallet_accounts_dag ON wallet_accounts(dag_address);

CREATE TABLE IF NOT EXISTS staking_delegations (
    id            TEXT PRIMARY KEY,
    did           TEXT NOT NULL,
    stake_id      TEXT NOT NULL,
    validator_id  TEXT NOT NULL,
    amount        BIGINT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_staking_delegations_did ON staking_delegations(did);
