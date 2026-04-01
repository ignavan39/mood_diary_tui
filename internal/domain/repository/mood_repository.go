package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
)

type MoodRepository interface {
	Create(ctx context.Context, entry *entity.MoodEntry) error

	Update(ctx context.Context, entry *entity.MoodEntry) error

	Delete(ctx context.Context, id uuid.UUID) error

	FindByID(ctx context.Context, id uuid.UUID) (*entity.MoodEntry, error)

	FindByDate(ctx context.Context, date time.Time) (*entity.MoodEntry, error)

	FindByDateRange(ctx context.Context, start, end time.Time) ([]*entity.MoodEntry, error)

	FindRecent(ctx context.Context, limit int) ([]*entity.MoodEntry, error)

	FindAll(ctx context.Context) ([]*entity.MoodEntry, error)

	GetStatistics(ctx context.Context, start, end time.Time) (*MoodStatistics, error)
}

type MoodStatistics struct {
	TotalEntries int
	AverageMood  float64
	MinMood      entity.MoodLevel
	MaxMood      entity.MoodLevel
	MoodCounts   map[entity.MoodLevel]int
	Trend        float64
}
