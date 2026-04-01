package database

const CreateTableSQL = `
CREATE TABLE IF NOT EXISTS mood_entries (
    id TEXT PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    note TEXT,
    level INTEGER NOT NULL CHECK(level >= 0 AND level <= 10),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_mood_entries_date ON mood_entries(date DESC);

CREATE INDEX IF NOT EXISTS idx_mood_entries_date_range ON mood_entries(date);

CREATE TRIGGER IF NOT EXISTS update_mood_entries_updated_at
AFTER UPDATE ON mood_entries
FOR EACH ROW
BEGIN
    UPDATE mood_entries SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
`

const DropTableSQL = `
DROP TRIGGER IF EXISTS update_mood_entries_updated_at;
DROP INDEX IF EXISTS idx_mood_entries_date_range;
DROP INDEX IF EXISTS idx_mood_entries_date;
DROP TABLE IF EXISTS mood_entries;
`
