ALTER TABLE users
    ADD COLUMN email VARCHAR(255) NULL AFTER name,
    ADD COLUMN email_verified_at DATETIME NULL AFTER email,
    ADD COLUMN password_changed_at DATETIME NULL AFTER password_hash,
    ADD UNIQUE KEY uniq_users_email (email),
    ADD UNIQUE KEY uniq_users_name (name);

CREATE TABLE auth_tokens (
    id          INT AUTO_INCREMENT NOT NULL,
    user_id     INT NOT NULL,
    token_hash  CHAR(64) NOT NULL,
    purpose     VARCHAR(16) NOT NULL,
    expires_at  DATETIME NOT NULL,
    used_at     DATETIME NULL,
    created_at  DATETIME NOT NULL,
    PRIMARY KEY (id),
    UNIQUE KEY uniq_auth_tokens_hash (token_hash),
    KEY idx_auth_tokens_user_purpose (user_id, purpose),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
