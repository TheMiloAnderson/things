DROP TABLE IF EXISTS task_contexts;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS areas;
DROP TABLE IF EXISTS contexts;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id              INT AUTO_INCREMENT NOT NULL,
    name            VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL,
    inbox           TEXT,
    PRIMARY KEY (id)
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