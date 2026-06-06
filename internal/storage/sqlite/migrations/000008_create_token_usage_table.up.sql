CREATE TABLE IF NOT EXISTS token_usage (
    id          TEXT PRIMARY KEY,
    project     TEXT NOT NULL,
    phase       TEXT,
    task        TEXT,
    artifact_id TEXT,
    pack_id     TEXT,
    token_count INTEGER NOT NULL DEFAULT 0,
    approximate INTEGER NOT NULL DEFAULT 1,
    model_hint  TEXT,
    recorded_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_token_usage_project ON token_usage(project);
CREATE INDEX IF NOT EXISTS idx_token_usage_recorded ON token_usage(recorded_at DESC);
