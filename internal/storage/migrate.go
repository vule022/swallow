package storage

import (
	"database/sql"
	"fmt"
)

const schema = `
CREATE TABLE IF NOT EXISTS projects (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    root_path   TEXT NOT NULL DEFAULT '',
    summary     TEXT NOT NULL DEFAULT '',
    tags        TEXT NOT NULL DEFAULT '[]',
    active_goals TEXT NOT NULL DEFAULT '[]',
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS project_settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS documents (
    id            TEXT PRIMARY KEY,
    project_id    TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    path          TEXT NOT NULL,
    relative_path TEXT NOT NULL,
    kind          TEXT NOT NULL DEFAULT 'other',
    hash          TEXT NOT NULL,
    size          INTEGER NOT NULL DEFAULT 0,
    modified_at   DATETIME NOT NULL,
    summary       TEXT NOT NULL DEFAULT '',
    raw_stored    INTEGER NOT NULL DEFAULT 0,
    created_at    DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_documents_project_path ON documents(project_id, relative_path);
CREATE INDEX IF NOT EXISTS idx_documents_hash ON documents(hash);

CREATE TABLE IF NOT EXISTS coding_outputs (
    id                     TEXT PRIMARY KEY,
    project_id             TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source                 TEXT NOT NULL DEFAULT '',
    raw_text               TEXT NOT NULL DEFAULT '',
    goal                   TEXT NOT NULL DEFAULT '',
    actions                TEXT NOT NULL DEFAULT '[]',
    files                  TEXT NOT NULL DEFAULT '[]',
    decisions              TEXT NOT NULL DEFAULT '[]',
    blockers               TEXT NOT NULL DEFAULT '[]',
    next_actions           TEXT NOT NULL DEFAULT '[]',
    validation_notes       TEXT NOT NULL DEFAULT '[]',
    commit_recommendations TEXT NOT NULL DEFAULT '[]',
    created_at             DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_coding_outputs_project ON coding_outputs(project_id, created_at DESC);

CREATE TABLE IF NOT EXISTS session_entries (
    id             TEXT PRIMARY KEY,
    project_id     TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type           TEXT NOT NULL,
    summary        TEXT NOT NULL DEFAULT '',
    related_files  TEXT NOT NULL DEFAULT '[]',
    decisions      TEXT NOT NULL DEFAULT '[]',
    open_questions TEXT NOT NULL DEFAULT '[]',
    next_action    TEXT NOT NULL DEFAULT '',
    created_at     DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_session_entries_project ON session_entries(project_id, created_at DESC);
`

func migrate(db *sql.DB) error {
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
