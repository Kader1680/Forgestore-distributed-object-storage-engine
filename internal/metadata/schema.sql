
CREATE TABLE IF NOT EXISTS files (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    size BIGINT NOT NULL CHECK (size >= 0),
    content_type TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chunks (
    id TEXT PRIMARY KEY,

    file_id TEXT NOT NULL
        REFERENCES files(id)
        ON DELETE CASCADE,

    chunk_index INTEGER NOT NULL CHECK (chunk_index >= 0),

    node_id TEXT NOT NULL,

    size BIGINT NOT NULL CHECK (size >= 0),

    checksum TEXT NOT NULL DEFAULT '',

    UNIQUE (file_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS chunks_node_id_idx
ON chunks(node_id);
