CREATE TABLE IF NOT EXISTS checkpoints (
	hash TEXT PRIMARY KEY,
	project TEXT NOT NULL,
	summary TEXT NOT NULL,
	created_at TEXT NOT NULL,
	artifact_ids TEXT NOT NULL,
	previous_checkpoint_id TEXT
);

CREATE INDEX IF NOT EXISTS idx_checkpoints_project_created_at ON checkpoints (project, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_checkpoints_previous_id ON checkpoints (previous_checkpoint_id);
