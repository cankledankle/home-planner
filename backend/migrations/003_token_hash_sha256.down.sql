-- Cannot restore bcrypt hashes after truncation.
-- Truncate to ensure clean state on rollback.
TRUNCATE refresh_tokens;
