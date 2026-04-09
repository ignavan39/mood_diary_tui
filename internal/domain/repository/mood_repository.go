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
	Count        int
	Average      float64
	MinLevel     int
	MaxLevel     int
	TotalDays    int
	Distribution map[int]int
	Trend        float64
	StartDate    time.Time
	EndDate      time.Time
}

func NewMoodStatistics() *MoodStatistics {
	return &MoodStatistics{
		Distribution: make(map[int]int),
	}
}
