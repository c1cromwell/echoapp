-- migrations/002_messaging.sql
-- Message queue, Merkle batches, contacts, media metadata

-- Offline encrypted message queue
CREATE TABLE IF NOT EXISTS message_queue (
    message_id      TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    sender_did      TEXT NOT NULL,
    recipient_did   TEXT NOT NULL,
    encrypted_payload BYTEA NOT NULL,
    commitment      BYTEA,
    content_type    TEXT NOT NULL DEFAULT 'text',
    queued_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '30 days'),
    delivered_at    TIMESTAMPTZ,
    status          TEXT NOT NULL DEFAULT 'queued' CHECK (status IN (
        'queued', 'delivered', 'expired', 'failed'
    ))
);

CREATE INDEX idx_message_queue_recipient ON message_queue(recipient_did, status);
CREATE INDEX idx_message_queue_conversation ON message_queue(conversation_id);
CREATE INDEX idx_message_queue_expires ON message_queue(expires_at);
CREATE INDEX idx_message_queue_status ON message_queue(status);

-- Merkle commitment batches for message anchoring
CREATE TABLE IF NOT EXISTS merkle_batches (
    batch_id            TEXT PRIMARY KEY,
    merkle_root         BYTEA NOT NULL,
    commitment_count    INT NOT NULL CHECK (commitment_count > 0),
    time_range_from     TIMESTAMPTZ NOT NULL,
    time_range_to       TIMESTAMPTZ NOT NULL,
    data_l1_tx_hash     TEXT,
    snapshot_hash       TEXT,
    snapshot_height     BIGINT,
    schema_version      INT NOT NULL DEFAULT 1,
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending', 'submitted', 'finalized', 'failed'
    )),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finalized_at        TIMESTAMPTZ
);

CREATE INDEX idx_merkle_batches_status ON merkle_batches(status);
CREATE INDEX idx_merkle_batches_time ON merkle_batches(time_range_from, time_range_to);

-- Contacts and block lists
CREATE TABLE IF NOT EXISTS contacts (
    id              TEXT PRIMARY KEY,
    owner_did       TEXT NOT NULL,
    contact_did     TEXT NOT NULL,
    nickname        TEXT,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN (
        'active', 'blocked', 'pending', 'removed'
    )),
    trust_tier      INT,
    phone_hash      TEXT,
    added_via       TEXT NOT NULL DEFAULT 'manual' CHECK (added_via IN (
        'manual', 'phone_discovery', 'qr_code', 'username_search', 'invite_link'
    )),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(owner_did, contact_did)
);

CREATE INDEX idx_contacts_owner ON contacts(owner_did, status);
CREATE INDEX idx_contacts_contact ON contacts(contact_did);
CREATE INDEX idx_contacts_phone_hash ON contacts(phone_hash);

-- Invite/referral tracking
CREATE TABLE IF NOT EXISTS invite_links (
    invite_id       TEXT PRIMARY KEY,
    inviter_did     TEXT NOT NULL,
    invite_code     TEXT UNIQUE NOT NULL,
    invitee_did     TEXT,
    status          TEXT NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending', 'accepted', 'expired', 'revoked'
    )),
    reward_claimed  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accepted_at     TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '30 days')
);

CREATE INDEX idx_invite_links_inviter ON invite_links(inviter_did);
CREATE INDEX idx_invite_links_code ON invite_links(invite_code);

-- Encrypted media metadata
CREATE TABLE IF NOT EXISTS media_files (
    file_id         TEXT PRIMARY KEY,
    uploader_did    TEXT NOT NULL,
    encrypted_size  BIGINT NOT NULL,
    content_type    TEXT NOT NULL,
    storage_backend TEXT NOT NULL DEFAULT 'storj' CHECK (storage_backend IN ('storj', 's3', 'ipfs')),
    storage_key     TEXT NOT NULL,
    chunk_count     INT NOT NULL DEFAULT 1,
    scan_status     TEXT NOT NULL DEFAULT 'pending' CHECK (scan_status IN (
        'pending', 'clean', 'flagged', 'error'
    )),
    ipfs_cid        TEXT,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_media_files_uploader ON media_files(uploader_did);
CREATE INDEX idx_media_files_scan ON media_files(scan_status);

-- Media file chunks for IPFS distribution
CREATE TABLE IF NOT EXISTS media_chunks (
    chunk_id    TEXT PRIMARY KEY,
    file_id     TEXT NOT NULL REFERENCES media_files(file_id),
    chunk_index INT NOT NULL,
    ipfs_cid    TEXT NOT NULL,
    size_bytes  INT NOT NULL,
    UNIQUE(file_id, chunk_index)
);

CREATE INDEX idx_media_chunks_file ON media_chunks(file_id);

-- Delivery receipts
CREATE TABLE IF NOT EXISTS delivery_receipts (
    message_id      TEXT NOT NULL,
    recipient_did   TEXT NOT NULL,
    receipt_type    TEXT NOT NULL CHECK (receipt_type IN ('delivered', 'read')),
    received_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (message_id, recipient_did, receipt_type)
);

CREATE INDEX idx_delivery_receipts_message ON delivery_receipts(message_id);
