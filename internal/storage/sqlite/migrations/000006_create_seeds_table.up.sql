CREATE TABLE IF NOT EXISTS seeds (
    id              TEXT PRIMARY KEY,
    artifact_id     TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    version_label   TEXT,
    seed_type       TEXT NOT NULL,
    role_access_bits INTEGER NOT NULL DEFAULT 1,
    description     TEXT,
    content         TEXT NOT NULL,
    token_estimate  INTEGER NOT NULL DEFAULT 0,
    tags            TEXT,
    author_role     TEXT,
    created_at      TEXT NOT NULL,
    FOREIGN KEY (artifact_id) REFERENCES intents(id)
);
CREATE INDEX IF NOT EXISTS idx_seeds_name ON seeds(name);
CREATE INDEX IF NOT EXISTS idx_seeds_type ON seeds(seed_type);
