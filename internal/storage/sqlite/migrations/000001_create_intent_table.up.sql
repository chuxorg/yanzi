CREATE TABLE IF NOT EXISTS intents (
	id TEXT PRIMARY KEY,
	created_at TEXT NOT NULL,
	author TEXT NOT NULL,
	source_type TEXT NOT NULL,
	title TEXT,
	prompt TEXT NOT NULL,
	response TEXT NOT NULL,
	meta TEXT,
	prev_hash TEXT,
	hash TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_intents_created_at ON intents (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_intents_prev_hash ON intents (prev_hash);
