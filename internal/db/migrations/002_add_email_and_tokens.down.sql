DROP TABLE IF EXISTS auth_tokens;

ALTER TABLE users
    DROP INDEX uniq_users_name,
    DROP INDEX uniq_users_email,
    DROP COLUMN password_changed_at,
    DROP COLUMN email_verified_at,
    DROP COLUMN email;
