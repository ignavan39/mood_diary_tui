package repository

import (
    "context"
    "github.com/ignavan39/mood-diary/internal/domain/entity"
)

type SettingsRepository interface {
    Get(ctx context.Context, key entity.SettingsKey) (*entity.UserSettings, error)
    Upsert(ctx context.Context, setting *entity.UserSettings) error
    GetAll(ctx context.Context) ([]*entity.UserSettings, error)
}