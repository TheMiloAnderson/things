INSERT INTO users (
    name, 
    password_hash, 
    inbox
) VALUES (
    "Milo Anderson", 
    "fake hash for now", 
    "inbox text has a 1:1 relationship with user!"
);

INSERT INTO contexts (
    name, 
    user_id
) SELECT 
    "Business Hours", 
    id
FROM users
WHERE name = "Milo Anderson";

INSERT INTO contexts (
    name, 
    user_id
) SELECT 
    "Phone", 
    id
FROM users
WHERE name = "Milo Anderson";

INSERT INTO areas (
    name, 
    user_id
) SELECT 
    "Music", 
    id
FROM users
WHERE name = "Milo Anderson";

INSERT INTO projects (
    name,
    status,
    notes,
    area_id,
    user_id
) SELECT 
    "Get amp repaired",
    "active",
    "Aviator Audio: 555-123-4567",
    areas.id,
    users.id
FROM users 
INNER JOIN areas ON users.id = areas.user_id
WHERE users.name = "Milo Anderson"
AND areas.name = "Music";

INSERT INTO tasks (
    name,
    status,
    priority,
    date_created,
    project_id,
    area_id,
    user_id
) SELECT
    "Call Aviator",
    "active",
    0,
    "2025-10-17 11:15:49",
    projects.id,
    areas.id,
    users.id
FROM users
INNER JOIN projects ON users.id = projects.user_id
INNER JOIN areas ON users.id = areas.user_id
WHERE users.name = "Milo Anderson"
AND areas.name = "Music"
AND projects.name = "Get amp repaired";

INSERT INTO task_contexts (
    task_id, 
    context_id
) SELECT 
    tasks.id,
    contexts.id
FROM users
INNER JOIN tasks ON users.id = tasks.user_id
INNER JOIN contexts ON users.id = contexts.user_id
WHERE tasks.name = "Call Aviator"
AND contexts.name = "Phone";

INSERT INTO task_contexts (
    task_id, 
    context_id
) SELECT 
    tasks.id,
    contexts.id
FROM users
INNER JOIN tasks ON users.id = tasks.user_id
INNER JOIN contexts ON users.id = contexts.user_id
WHERE tasks.name = "Call Aviator"
AND contexts.name = "Business Hours";