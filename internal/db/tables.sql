DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS task_contexts;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS areas;
DROP TABLE IF EXISTS contexts;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id                  INT AUTO_INCREMENT NOT NULL,
    name                VARCHAR(255) NOT NULL,
    email               VARCHAR(255) NOT NULL,
    email_verified_at   DATETIME NULL,
    password_hash       VARCHAR(255) NOT NULL,
    password_changed_at DATETIME NULL,
    inbox               TEXT,
    PRIMARY KEY (id),
    UNIQUE KEY uniq_users_email (email),
    UNIQUE KEY uniq_users_name (name)
);

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

CREATE TABLE contexts (
    id      INT AUTO_INCREMENT NOT NULL,
    name    VARCHAR(255) NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE TABLE areas (
    id      INT AUTO_INCREMENT NOT NULL,
    name    VARCHAR(255) NOT NULL,
    user_id INT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE TABLE projects (
    id          INT AUTO_INCREMENT NOT NULL,
    name        VARCHAR(255) NOT NULL,
    status      VARCHAR(24) NOT NULL,
    notes       TEXT,
    area_id     INT,
    user_id     INT NOT NULL,
    FOREIGN KEY (area_id) REFERENCES areas(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE TABLE tasks (
    id              INT AUTO_INCREMENT NOT NULL,
    name            VARCHAR(255) NOT NULL,
    status          VARCHAR(24) NOT NULL,
    priority        INT,
    notes           TEXT,
    date_created    DATETIME NOT NULL,
    project_id      INT,
    area_id         INT,
    user_id         INT NOT NULL,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    FOREIGN KEY (area_id) REFERENCES areas(id) ON DELETE SET NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (id)
);

CREATE TABLE task_contexts (
    task_id     INT,
    context_id  INT,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    FOREIGN KEY (context_id) REFERENCES contexts(id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, context_id)
)