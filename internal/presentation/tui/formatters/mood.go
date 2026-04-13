package formatters

import (
	"fmt"

	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
)

func FormatMoodLevel(level entity.MoodLevel, translator i18n.Translator) string {
	return fmt.Sprintf("%s %s (%d/10)", level.Emoji(), translator.T(level.StringKey()), level.Int())
}

func FormatMoodEntry(entry *entity.MoodEntry, translator i18n.Translator) string {
	dateStr := entry.Date.Format("02.01.2006")
	moodStr := FormatMoodLevel(entry.Level, translator)

	note := entry.Note
	if note == "" {
		note = translator.T(i18n.RecordWithoutNoteKey)
	}

	return fmt.Sprintf("%s | %s | %s", dateStr, moodStr, note)
}

func TruncateNote(note string, maxLength int) string {
	if len(note) <= maxLength {
		return note
	}
	return note[:maxLength-3] + "..."
}

func GetMoodEmoji(level int) string {
	if level < 0 || level > 10 {
		return "❓"
	}
	return entity.MoodLevel(level).Emoji()
}
