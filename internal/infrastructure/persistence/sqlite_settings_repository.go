package persistence

import (
    "context"
    "database/sql"
    "errors"
    "github.com/ignavan39/mood-diary/internal/domain/entity"
    "github.com/ignavan39/mood-diary/internal/domain/repository"
)

type SQLiteSettingsRepository struct {
    db *sql.DB
}

func NewSQLiteSettingsRepository(db *sql.DB) repository.SettingsRepository {
    return &SQLiteSettingsRepository{db: db}
}

func (r *SQLiteSettingsRepository) Get(
    ctx context.Context, 
    key entity.SettingsKey,
) (*entity.UserSettings, error) {
    query := `SELECT key, value, created_at, updated_at 
              FROM user_settings WHERE key = ? LIMIT 1`
    
    var setting entity.UserSettings
    err := r.db.QueryRowContext(ctx, query, key).Scan(
        &setting.Key,
        &setting.Value,
        &setting.CreatedAt,
        &setting.UpdatedAt,
    )
    
    if errors.Is(err, sql.ErrNoRows) {
        return nil, entity.ErrSettingsNotFound
    }
    if err != nil {
        return nil, err
    }
    
    return &setting, nil
}

func (r *SQLiteSettingsRepository) Upsert(
    ctx context.Context, 
    setting *entity.UserSettings,
) error {
    if err := setting.Validate(); err != nil {
        return err
    }
    
    query := `INSERT INTO user_settings (key, value, created_at, updated_at)
              VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
              ON CONFLICT(key) DO UPDATE 
              SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`
    
    _, err := r.db.ExecContext(ctx, query, setting.Key, setting.Value)
    return err
}

func (r *SQLiteSettingsRepository) GetAll(
    ctx context.Context,
) ([]*entity.UserSettings, error) {
    query := `SELECT key, value, created_at, updated_at FROM user_settings`
    
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var settings []*entity.UserSettings
    for rows.Next() {
        var s entity.UserSettings
        if err := rows.Scan(&s.Key, &s.Value, &s.CreatedAt, &s.UpdatedAt); err != nil {
            return nil, err
        }
        settings = append(settings, &s)
    }
    
    return settings, rows.Err()
}