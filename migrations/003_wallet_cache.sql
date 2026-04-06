-- migrations/003_wallet_cache.sql
-- Wallet balance cache and staking positions

-- Wallet balance cache (refreshed from metagraph every 5s)
CREATE TABLE wallet_balance_cache (
    did            TEXT PRIMARY KEY,
    total_balance  BIGINT NOT NULL DEFAULT 0,
    available      BIGINT NOT NULL DEFAULT 0,
    staked         BIGINT NOT NULL DEFAULT 0,
    pending_rewards BIGINT NOT NULL DEFAULT 0,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Staking positions (mirror of on-chain TokenLock state)
CREATE TABLE staking_positions (
    id            TEXT PRIMARY KEY,
    did           TEXT NOT NULL,
    amount        BIGINT NOT NULL,
    tier          TEXT NOT NULL,
    locked_until  TIMESTAMPTZ NOT NULL,
    vesting_type  TEXT, -- 'founder' or NULL
    delegated_to  TEXT, -- validator ID or NULL
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_staking_positions_did ON staking_positions(did);

-- Daily reward caps (reset at UTC midnight)
CREATE TABLE daily_reward_caps (
    did          TEXT NOT NULL,
    reward_type  TEXT NOT NULL, -- messaging, referral, staking, payment_rail
    earned_today BIGINT NOT NULL DEFAULT 0,
    cap          BIGINT NOT NULL,
    reset_at     TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (did, reward_type)
);

-- Validator directory (synced from metagraph)
CREATE TABLE validators (
    id              TEXT PRIMARY KEY,
    address         TEXT NOT NULL,
    layer           TEXT NOT NULL, -- currency_l1, data_l1
    uptime_percent  REAL NOT NULL DEFAULT 0,
    commission_pct  REAL NOT NULL DEFAULT 0,
    total_delegated BIGINT NOT NULL DEFAULT 0,
    delegator_count INT NOT NULL DEFAULT 0,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
