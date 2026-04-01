package repository

import (
	"context"
	"mood-diary/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type MoodRepository interface {
	// Create saves a new mood entry
	Create(ctx context.Context, entry *entity.MoodEntry) error

	// Update updates an existing mood entry
	Update(ctx context.Context, entry *entity.MoodEntry) error

	// Delete removes a mood entry
	Delete(ctx context.Context, id uuid.UUID) error

	// FindByID retrieves a mood entry by its ID
	FindByID(ctx context.Context, id uuid.UUID) (*entity.MoodEntry, error)

	// FindByDate retrieves a mood entry for a specific date
	FindByDate(ctx context.Context, date time.Time) (*entity.MoodEntry, error)

	// FindByDateRange retrieves mood entries within a date range
	FindByDateRange(ctx context.Context, start, end time.Time) ([]*entity.MoodEntry, error)

	// FindRecent retrieves the most recent N mood entries
	FindRecent(ctx context.Context, limit int) ([]*entity.MoodEntry, error)

	// FindAll retrieves all mood entries ordered by date
	FindAll(ctx context.Context) ([]*entity.MoodEntry, error)

	// GetStatistics retrieves aggregated statistics
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
