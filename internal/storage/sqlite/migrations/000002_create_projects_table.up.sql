CREATE TABLE IF NOT EXISTS projects (
	name TEXT PRIMARY KEY,
	description TEXT,
	created_at TEXT NOT NULL,
	prev_hash TEXT,
	hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_projects_created_at ON projects (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_projects_prev_hash ON projects (prev_hash);
