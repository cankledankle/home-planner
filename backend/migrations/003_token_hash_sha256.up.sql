-- Switch token storage from bcrypt to SHA-256 hashes.
-- SHA-256 is appropriate here because tokens are already high-entropy random UUIDs,
-- so bcrypt's slow-hash property adds no security value and causes O(N) scan on validation.
-- This invalidates all existing sessions (forces re-login).
TRUNCATE refresh_tokens;
