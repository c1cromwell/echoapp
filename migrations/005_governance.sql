-- 005_governance.sql
-- Trust-tier weighted governance: proposals and votes

CREATE TABLE proposals (
    id              TEXT PRIMARY KEY,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL,
    type            TEXT NOT NULL CHECK (type IN ('protocol_upgrade', 'treasury_allocation', 'parameter_change', 'board_election')),
    threshold       TEXT NOT NULL DEFAULT 'simple_majority' CHECK (threshold IN ('simple_majority', 'supermajority_67', 'supermajority_75')),
    created_by      TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ends_at         TIMESTAMPTZ NOT NULL,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'passed', 'failed', 'executed')),

    -- Cached tally (updated on each vote for fast reads)
    for_weight      BIGINT NOT NULL DEFAULT 0,
    against_weight  BIGINT NOT NULL DEFAULT 0,
    abstain_weight  BIGINT NOT NULL DEFAULT 0,
    voter_count     INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_proposals_status ON proposals(status);
CREATE INDEX idx_proposals_ends_at ON proposals(ends_at);
CREATE INDEX idx_proposals_created_by ON proposals(created_by);

CREATE TABLE votes (
    did             TEXT NOT NULL,
    proposal_id     TEXT NOT NULL REFERENCES proposals(id),
    value           TEXT NOT NULL CHECK (value IN ('for', 'against', 'abstain')),
    weight          BIGINT NOT NULL,
    trust_tier      INT NOT NULL,
    staked          BIGINT NOT NULL,
    tx_hash         TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (did, proposal_id)
);

CREATE INDEX idx_votes_proposal_id ON votes(proposal_id);
CREATE INDEX idx_votes_tx_hash ON votes(tx_hash);
