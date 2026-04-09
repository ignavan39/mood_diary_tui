package formatters

import (
	"testing"
	"time"

	"github.com/ignavan39/mood-diary/internal/domain/entity"
)

func TestFormatMoodLevel(t *testing.T) {
	level := entity.MoodLevel(5)
	result := FormatMoodLevel(level)

	if result == "" {
		t.Error("FormatMoodLevel should not return empty string")
	}

	// Проверяем что результат содержит основные элементы
	if len(result) < 5 {
		t.Error("FormatMoodLevel should return meaningful string")
	}
}

func TestTruncateNote(t *testing.T) {
	tests := []struct {
		name      string
		note      string
		maxLength int
		expected  string
	}{
		{
			name:      "Short note",
			note:      "Hello",
			maxLength: 10,
			expected:  "Hello",
		},
		{
			name:      "Long note",
			note:      "This is a very long note that should be truncated",
			maxLength: 20,
			expected:  "This is a very lo...",
		},
		{
			name:      "Exact length",
			note:      "Exactly",
			maxLength: 7,
			expected:  "Exactly",
		},
		{
			name:      "Empty note",
			note:      "",
			maxLength: 10,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateNote(tt.note, tt.maxLength)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetMoodEmoji(t *testing.T) {
	tests := []struct {
		level    int
		hasEmoji bool
	}{
		{0, true},
		{5, true},
		{10, true},
		{-1, true}, // Возвращает "❓"
		{11, true}, // Возвращает "❓"
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			emoji := GetMoodEmoji(tt.level)
			if tt.hasEmoji && emoji == "" {
				t.Errorf("Level %d should have an emoji", tt.level)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	date := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	result := FormatDate(date)

	expected := "15.03.2024"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFormatDateTime(t *testing.T) {
	dateTime := time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC)
	result := FormatDateTime(dateTime)

	expected := "15.03.2024 14:30"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFormatDateRange(t *testing.T) {
	start := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)

	result := FormatDateRange(start, end)

	expected := "01.03.2024 — 31.03.2024"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestFormatRelativeDate(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		date     time.Time
		expected string
	}{
		{
			name:     "Today",
			date:     today,
			expected: "Сегодня",
		},
		{
			name:     "Yesterday",
			date:     today.AddDate(0, 0, -1),
			expected: "Вчера",
		},
		{
			name:     "Tomorrow",
			date:     today.AddDate(0, 0, 1),
			expected: "Завтра",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRelativeDate(tt.date)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDaysAgo(t *testing.T) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		date     time.Time
		expected int
	}{
		{
			name:     "Today",
			date:     today,
			expected: 0,
		},
		{
			name:     "Yesterday",
			date:     today.AddDate(0, 0, -1),
			expected: 1,
		},
		{
			name:     "Week ago",
			date:     today.AddDate(0, 0, -7),
			expected: 7,
		},
		{
			name:     "Tomorrow",
			date:     today.AddDate(0, 0, 1),
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DaysAgo(tt.date)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}
