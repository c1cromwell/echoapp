-- Safety reports: say what was reported, and keep reports through account deletion.
--
-- message_id / conversation_id / channel_id are optional pointers to the thing
-- that was reported (context 'message' or 'channel'). details is optional
-- free text from the reporter; it is user text, capped, and only ever read
-- through the operator moderation endpoint.
ALTER TABLE safety_reports ADD COLUMN IF NOT EXISTS message_id      TEXT;
ALTER TABLE safety_reports ADD COLUMN IF NOT EXISTS conversation_id TEXT;
ALTER TABLE safety_reports ADD COLUMN IF NOT EXISTS channel_id      TEXT;
ALTER TABLE safety_reports ADD COLUMN IF NOT EXISTS details         TEXT CHECK (details IS NULL OR char_length(details) <= 1000);

-- A deleted account must not take the reports it filed with it, and must not
-- be able to erase reports filed against it. Account deletion now keeps the
-- row and clears reporter_did (see internal/database/account_purge.go).
ALTER TABLE safety_reports ALTER COLUMN reporter_did DROP NOT NULL;

-- Moderation reads newest-first, optionally for one subject.
CREATE INDEX IF NOT EXISTS idx_safety_reports_created ON safety_reports (created_at DESC, id);
