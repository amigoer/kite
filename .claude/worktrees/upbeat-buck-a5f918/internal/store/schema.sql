CREATE TABLE IF NOT EXISTS rooms (
    id         TEXT PRIMARY KEY,
    name       TEXT,
    created_at INTEGER NOT NULL,
    closed_at  INTEGER,
    status     TEXT NOT NULL,
    mode       TEXT NOT NULL DEFAULT 'scripted',
    cwd        TEXT,
    shell      TEXT,
    metadata   TEXT
);

-- Add mode column for databases created by an older daemon. Errors (e.g.
-- "duplicate column name") are ignored by the driver via the ALTER… IF NOT
-- EXISTS-style guard below.
CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY);
INSERT OR IGNORE INTO schema_migrations(version) VALUES (1);

CREATE TABLE IF NOT EXISTS events (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id    TEXT NOT NULL,
    timestamp  INTEGER NOT NULL,
    type       TEXT NOT NULL,
    payload    BLOB NOT NULL,
    FOREIGN KEY (room_id) REFERENCES rooms(id)
);

CREATE INDEX IF NOT EXISTS idx_events_room_id ON events(room_id, id);
CREATE INDEX IF NOT EXISTS idx_rooms_status ON rooms(status);
