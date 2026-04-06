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

CREATE TABLE IF NOT EXISTS user_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_settings_key ON user_settings(key);


CREATE INDEX IF NOT EXISTS idx_mood_entries_date ON mood_entries(date DESC);

CREATE INDEX IF NOT EXISTS idx_mood_entries_date_range ON mood_entries(date);

CREATE TRIGGER IF NOT EXISTS update_mood_entries_updated_at
AFTER UPDATE ON mood_entries
FOR EACH ROW
BEGIN
    UPDATE mood_entries SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_settings_timestamp
AFTER UPDATE ON user_settings
BEGIN
    UPDATE user_settings 
    SET updated_at = CURRENT_TIMESTAMP 
    WHERE id = NEW.id;
END;

INSERT OR IGNORE INTO user_settings (key, value) VALUES ('language', 'en');

`

const DropTableSQL = `
DROP TRIGGER IF EXISTS update_mood_entries_updated_at;
DROP INDEX IF EXISTS idx_mood_entries_date_range;
DROP INDEX IF EXISTS idx_mood_entries_date;
DROP TABLE IF EXISTS mood_entries;
DROP TRIGGER IF EXISTS update_settings_timestamp;
DROP INDEX IF EXISTS idx_settings_key;
DROP TABLE IF EXISTS user_settings;
`
