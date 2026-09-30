CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    path TEXT NOT NULL UNIQUE,
    last_opened_at DATETIME NOT NULL
);

-- AI provider config. The API key itself is never stored here - only a
-- reference resolved against OS keychain / encrypted-file storage.
CREATE TABLE IF NOT EXISTS ai_providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    kind TEXT NOT NULL,           -- 'hosted' or 'cli'
    model TEXT NOT NULL DEFAULT '',
    endpoint TEXT NOT NULL DEFAULT '',
    cli_path TEXT NOT NULL DEFAULT '',
    cli_args TEXT NOT NULL DEFAULT '', -- space-separated
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL
);

-- A Specflow: an ordered sequence of changes in one project, scheduled to be
-- implemented automatically via a configured AI provider.
CREATE TABLE IF NOT EXISTS specflows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    provider_id INTEGER NOT NULL,
    scheduled_at DATETIME NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending|running|succeeded|failed|cancelled
    missed INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS specflow_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    flow_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    change_name TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending', -- pending|running|succeeded|failed
    log TEXT NOT NULL DEFAULT '',
    started_at DATETIME,
    finished_at DATETIME,
    FOREIGN KEY (flow_id) REFERENCES specflows(id)
);
