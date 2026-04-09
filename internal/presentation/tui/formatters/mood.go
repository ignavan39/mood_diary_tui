package formatters

import (
	"fmt"

	"github.com/ignavan39/mood-diary/internal/domain/entity"
)

// FormatMoodLevel форматирует уровень настроения для отображения
func FormatMoodLevel(level entity.MoodLevel) string {
	return fmt.Sprintf("%s %s (%d/10)", level.Emoji(), level.String(), level.Int())
}

// FormatMoodEntry форматирует полную запись настроения
func FormatMoodEntry(entry *entity.MoodEntry) string {
	dateStr := entry.Date.Format("02.01.2006")
	moodStr := FormatMoodLevel(entry.Level)
	
	note := entry.Note
	if note == "" {
		note = "без заметки"
	}
	
	return fmt.Sprintf("%s | %s | %s", dateStr, moodStr, note)
}

// TruncateNote обрезает заметку до указанной длины
func TruncateNote(note string, maxLength int) string {
	if len(note) <= maxLength {
		return note
	}
	return note[:maxLength-3] + "..."
}

// GetMoodEmoji возвращает эмоджи для уровня настроения
func GetMoodEmoji(level int) string {
	if level < 0 || level > 10 {
		return "❓"
	}
	return entity.MoodLevel(level).Emoji()
}
