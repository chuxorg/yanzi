CREATE TABLE IF NOT EXISTS packs (
    id              TEXT PRIMARY KEY,
    artifact_id     TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    version_label   TEXT,
    extends_id      TEXT,
    role_bits       INTEGER NOT NULL DEFAULT 1,
    role_label      TEXT,
    description     TEXT,
    pack_context    TEXT,
    seeds           TEXT NOT NULL DEFAULT '[]',
    token_estimate  TEXT NOT NULL DEFAULT '{}',
    tags            TEXT,
    author_role     TEXT,
    created_at      TEXT NOT NULL,
    FOREIGN KEY (artifact_id) REFERENCES intents(id)
);
CREATE INDEX IF NOT EXISTS idx_packs_name ON packs(name);
CREATE INDEX IF NOT EXISTS idx_packs_extends ON packs(extends_id);
