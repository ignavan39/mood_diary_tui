package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/domain/repository"
)

type SQLiteMoodRepository struct {
	db *sql.DB
}

func NewSQLiteMoodRepository(db *sql.DB) *SQLiteMoodRepository {
	return &SQLiteMoodRepository{db: db}
}

func (r *SQLiteMoodRepository) Upsert(ctx context.Context, entry *entity.MoodEntry) error {
	query := `
		INSERT INTO mood_entries (id, date, level, note, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(date) DO UPDATE SET level=excluded.level, note=excluded.note, updated_at=excluded.updated_at
	`

	dateStr := entry.Date.Format("2006-01-02")

	_, err := r.db.ExecContext(ctx, query,
		entry.ID.String(),
		dateStr,
		entry.Level.Int(),
		entry.Note,
		entry.CreatedAt,
		entry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create mood entry: %w", err)
	}

	return nil
}

func (r *SQLiteMoodRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM mood_entries WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id.String())
	if err != nil {
		return fmt.Errorf("failed to delete mood entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrNotFound
	}

	return nil
}

func (r *SQLiteMoodRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.MoodEntry, error) {
	query := `
		SELECT id, date, level, note, created_at, updated_at
		FROM mood_entries
		WHERE id = ?
	`

	var (
		idStr     string
		dateStr   string
		level     int
		note      string
		createdAt time.Time
		updatedAt time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id.String()).Scan(
		&idStr, &dateStr, &level, &note, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find mood entry: %w", err)
	}

	return r.scanMoodEntry(idStr, dateStr, level, note, createdAt, updatedAt)
}

func (r *SQLiteMoodRepository) FindByDate(ctx context.Context, date time.Time) (*entity.MoodEntry, error) {
	query := `
		SELECT id, date, level, note, created_at, updated_at
		FROM mood_entries
		WHERE date = ?
	`

	dateStr := date.Format("2006-01-02")

	var (
		idStr        string
		foundDateStr string
		level        int
		note         string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := r.db.QueryRowContext(ctx, query, dateStr).Scan(
		&idStr, &foundDateStr, &level, &note, &createdAt, &updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find mood entry by date: %w", err)
	}

	return r.scanMoodEntry(idStr, foundDateStr, level, note, createdAt, updatedAt)
}

func (r *SQLiteMoodRepository) FindByDateRange(ctx context.Context, start, end time.Time) ([]*entity.MoodEntry, error) {
	query := `
		SELECT id, date, level, note, created_at, updated_at
		FROM mood_entries
		WHERE date BETWEEN ? AND ?
		ORDER BY date DESC
	`

	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	rows, err := r.db.QueryContext(ctx, query, startStr, endStr)
	if err != nil {
		return nil, fmt.Errorf("failed to find mood entries by date range: %w", err)
	}
	defer rows.Close()

	return r.scanMoodEntries(rows)
}

func (r *SQLiteMoodRepository) FindRecent(ctx context.Context, limit int) ([]*entity.MoodEntry, error) {
	query := `
		SELECT id, date, level, note, created_at, updated_at
		FROM mood_entries
		ORDER BY date DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find recent mood entries: %w", err)
	}
	defer rows.Close()

	return r.scanMoodEntries(rows)
}

func (r *SQLiteMoodRepository) FindAll(ctx context.Context) ([]*entity.MoodEntry, error) {
	query := `
		SELECT id, date, level, note, created_at, updated_at
		FROM mood_entries
		ORDER BY date DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find all mood entries: %w", err)
	}
	defer rows.Close()

	return r.scanMoodEntries(rows)
}

func (r *SQLiteMoodRepository) GetStatistics(ctx context.Context, start, end time.Time) (*repository.MoodStatistics, error) {
	stats := repository.NewMoodStatistics()
	stats.StartDate = start
	stats.EndDate = end

	query := `
		SELECT 
			COUNT(*) as total,
			AVG(level) as average,
			MIN(level) as min_level,
			MAX(level) as max_level
		FROM mood_entries
		WHERE date BETWEEN ? AND ?
	`
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	err := r.db.QueryRowContext(ctx, query, startStr, endStr).
		Scan(&stats.Count, &stats.Average, &stats.MinLevel, &stats.MaxLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	countQuery := `
		SELECT level, COUNT(*) as count
		FROM mood_entries
		WHERE date BETWEEN ? AND ?
		GROUP BY level
	`
	rows, err := r.db.QueryContext(ctx, countQuery, startStr, endStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get mood counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var level, count int
		if err := rows.Scan(&level, &count); err != nil {
			continue
		}
		stats.Distribution[level] = count
	}

	stats.TotalDays = int(end.Sub(start).Hours()/24) + 1

	stats.Trend = r.calculateTrend(ctx, start, end)

	return stats, nil
}

func (r *SQLiteMoodRepository) calculateTrend(ctx context.Context, start, end time.Time) float64 {

	query := `
		SELECT level, date
		FROM mood_entries
		WHERE date BETWEEN ? AND ?
		ORDER BY date ASC
	`

	rows, err := r.db.QueryContext(ctx, query, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return 0
	}
	defer rows.Close()

	var levels []int
	for rows.Next() {
		var level int
		var date string
		if err := rows.Scan(&level, &date); err != nil {
			continue
		}
		levels = append(levels, level)
	}

	if len(levels) < 2 {
		return 0
	}

	mid := len(levels) / 2
	firstHalfAvg := average(levels[:mid])
	secondHalfAvg := average(levels[mid:])

	return secondHalfAvg - firstHalfAvg
}

func (r *SQLiteMoodRepository) scanMoodEntry(idStr, dateStr string, level int, note string, createdAt, updatedAt time.Time) (*entity.MoodEntry, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse UUID: %w", err)
	}

	var date time.Time

	date, err = time.Parse("2006-01-02", dateStr)
	if err != nil {

		date, err = time.Parse(time.RFC3339, dateStr)
		if err != nil {

			if len(dateStr) >= 10 {
				date, err = time.Parse("2006-01-02", dateStr[:10])
			}
			if err != nil {
				return nil, fmt.Errorf("failed to parse date %q: %w", dateStr, err)
			}
		}
	}

	moodLevel, err := entity.NewMoodLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid mood level: %w", err)
	}

	return &entity.MoodEntry{
		ID:        id,
		Date:      date,
		Level:     moodLevel,
		Note:      note,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *SQLiteMoodRepository) scanMoodEntries(rows *sql.Rows) ([]*entity.MoodEntry, error) {
	var entries []*entity.MoodEntry

	for rows.Next() {
		var (
			idStr     string
			dateStr   string
			level     int
			note      string
			createdAt time.Time
			updatedAt time.Time
		)

		if err := rows.Scan(&idStr, &dateStr, &level, &note, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		entry, err := r.scanMoodEntry(idStr, dateStr, level, note, createdAt, updatedAt)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}

func average(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return float64(sum) / float64(len(nums))
}
