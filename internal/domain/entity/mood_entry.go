package entity

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidMoodLevel = errors.New("mood level must be between 0 and 10")
	ErrEmptyID          = errors.New("mood entry ID cannot be empty")
)

type MoodLevel int

const (
	MinMoodLevel MoodLevel = 0
	MaxMoodLevel MoodLevel = 10
)

func NewMoodLevel(level int) (MoodLevel, error) {
	ml := MoodLevel(level)
	if ml < MinMoodLevel || ml > MaxMoodLevel {
		return 0, ErrInvalidMoodLevel
	}
	return ml, nil
}

func (m MoodLevel) Int() int {
	return int(m)
}

func (m MoodLevel) StringKey() string {
	return fmt.Sprintf("mood.level.%d", m)
}

func (m MoodLevel) String() string {

	return fmt.Sprintf("level_%d", m)
}

func (m MoodLevel) Emoji() string {
	emojis := map[MoodLevel]string{
		0: "😢", 1: "😞", 2: "😔", 3: "😕", 4: "😐",
		5: "😶", 6: "🙂", 7: "😊", 8: "😄", 9: "😁", 10: "🤩",
	}
	return emojis[m]
}

type MoodEntry struct {
	ID        uuid.UUID
	Date      time.Time
	Level     MoodLevel
	Note      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewMoodEntry(level MoodLevel, note string, date time.Time) (*MoodEntry, error) {
	if level < MinMoodLevel || level > MaxMoodLevel {
		return nil, ErrInvalidMoodLevel
	}

	now := time.Now()

	normalizedDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	return &MoodEntry{
		ID:        uuid.New(),
		Date:      normalizedDate,
		Level:     level,
		Note:      note,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (m *MoodEntry) Update(level MoodLevel, note string) error {
	if level < MinMoodLevel || level > MaxMoodLevel {
		return ErrInvalidMoodLevel
	}

	m.Level = level
	m.Note = note
	m.UpdatedAt = time.Now()

	return nil
}

func (m *MoodEntry) IsToday() bool {
	now := time.Now()
	return m.Date.Year() == now.Year() &&
		m.Date.Month() == now.Month() &&
		m.Date.Day() == now.Day()
}

func (m *MoodEntry) DaysAgo() int {
	now := time.Now()
	normalizedNow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	duration := normalizedNow.Sub(m.Date)
	return int(duration.Hours() / 24)
}
