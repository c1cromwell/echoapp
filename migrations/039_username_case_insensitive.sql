-- Usernames are unique regardless of case: "Alice" and "alice" are the same name.
-- Without this, someone could register a look-alike of an existing username.
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username_lower ON users (lower(username));
