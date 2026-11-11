-- Echo Bar allow-list. Discovery and a normal contact add leave this false.
ALTER TABLE contacts ADD COLUMN IF NOT EXISTS bar_allowed BOOLEAN NOT NULL DEFAULT FALSE;
