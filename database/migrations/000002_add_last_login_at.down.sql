DROP INDEX IF EXISTS idx_users_last_login_at;

ALTER TABLE users DROP COLUMN last_login_at;
