-- migrations/017_message_read_receipts.sql
-- WO-192: durable read receipts on the content-blind relay.
-- Adds a read_at timestamp and the 'read' delivery state to message_queue.
-- This is metadata only (delivery state + timestamps) — no message content.

ALTER TABLE message_queue
    ADD COLUMN IF NOT EXISTS read_at TIMESTAMPTZ;

-- Extend the status CHECK to allow 'read' (read implies delivered).
ALTER TABLE message_queue
    DROP CONSTRAINT IF EXISTS message_queue_status_check;

ALTER TABLE message_queue
    ADD CONSTRAINT message_queue_status_check
    CHECK (status IN ('queued', 'delivered', 'read', 'expired', 'failed'));
