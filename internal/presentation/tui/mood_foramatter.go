package tui

import (
	"fmt"

	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
)

func FormatMoodLevel(level entity.MoodLevel, t i18n.Translator) string {
	if t == nil {

		return level.String()
	}
	return t.T(level.StringKey())
}

func FormatMoodWithEmoji(level entity.MoodLevel, t i18n.Translator) string {
	desc := FormatMoodLevel(level, t)
	return fmt.Sprintf("%s %s", level.Emoji(), desc)
}
