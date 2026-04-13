package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/domain/repository"
	"github.com/ignavan39/mood-diary/internal/infrastructure/persistence"
)

var (
	ErrMoodAlreadyRecorded = errors.New("mood for this date is already recorded")
	ErrMoodNotFound        = errors.New("mood entry not found")
)

type MoodService struct {
	repo repository.MoodRepository
}

func NewMoodService(repo repository.MoodRepository) *MoodService {
	return &MoodService{repo: repo}
}

func (s *MoodService) RecordMood(ctx context.Context, level int, note string, date *time.Time) error {

	moodLevel, err := entity.NewMoodLevel(level)
	if err != nil {
		return err
	}

	targetDate := time.Now()
	if date != nil {
		targetDate = *date
	}

	existing, err := s.repo.FindByDate(ctx, targetDate)
	if err != nil && err != persistence.ErrNotFound {
		return err
	}

	if existing != nil {
		return ErrMoodAlreadyRecorded
	}

	entry, err := entity.NewMoodEntry(moodLevel, note, targetDate)
	if err != nil {
		return err
	}

	return s.repo.Create(ctx, entry)
}

func (s *MoodService) UpdateMood(ctx context.Context, date time.Time, level int, note string) error {

	moodLevel, err := entity.NewMoodLevel(level)
	if err != nil {
		return err
	}

	entry, err := s.repo.FindByDate(ctx, date)
	if err != nil {
		if err == persistence.ErrNotFound {
			return ErrMoodNotFound
		}
		return err
	}

	if err := entry.Update(moodLevel, note); err != nil {
		return err
	}

	return s.repo.Update(ctx, entry)
}

func (s *MoodService) DeleteMood(ctx context.Context, date time.Time) error {
	entry, err := s.repo.FindByDate(ctx, date)
	if err != nil {
		if err == persistence.ErrNotFound {
			return ErrMoodNotFound
		}
		return err
	}

	return s.repo.Delete(ctx, entry.ID)
}

func (s *MoodService) GetMoodForDate(ctx context.Context, date time.Time) (*entity.MoodEntry, error) {
	entry, err := s.repo.FindByDate(ctx, date)
	if err != nil {
		if err == persistence.ErrNotFound {
			return nil, ErrMoodNotFound
		}
		return nil, err
	}
	return entry, nil
}

func (s *MoodService) GetTodayMood(ctx context.Context) (*entity.MoodEntry, error) {
	return s.GetMoodForDate(ctx, time.Now())
}

func (s *MoodService) GetRecentMoods(ctx context.Context, limit int) ([]*entity.MoodEntry, error) {
	if limit <= 0 {
		limit = 30
	}
	return s.repo.FindRecent(ctx, limit)
}

func (s *MoodService) GetMoodsForPeriod(ctx context.Context, period Period) ([]*entity.MoodEntry, error) {
	start, end := period.DateRange()
	return s.repo.FindByDateRange(ctx, start, end)
}

func (s *MoodService) GetMoodsByDateRange(ctx context.Context, start, end time.Time) ([]*entity.MoodEntry, error) {
	return s.repo.FindByDateRange(ctx, start, end)
}

func (s *MoodService) GetStatistics(ctx context.Context, period Period) (*repository.MoodStatistics, error) {
	start, end := period.DateRange()
	return s.repo.GetStatistics(ctx, start, end)
}

func (s *MoodService) GetAllMoods(ctx context.Context) ([]*entity.MoodEntry, error) {
	return s.repo.FindAll(ctx)
}

type Period int

const (
	PeriodWeek Period = iota
	PeriodMonth
	PeriodQuarter
	PeriodYear
	PeriodAll
)

func (p Period) DateRange() (start, end time.Time) {
	now := time.Now()
	end = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	switch p {
	case PeriodWeek:
		start = end.AddDate(0, 0, -7)
	case PeriodMonth:
		start = end.AddDate(0, -1, 0)
	case PeriodQuarter:
		start = end.AddDate(0, -3, 0)
	case PeriodYear:
		start = end.AddDate(-1, 0, 0)
	case PeriodAll:
		start = time.Date(2020, 1, 1, 0, 0, 0, 0, now.Location())
	}

	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	return start, end
}
