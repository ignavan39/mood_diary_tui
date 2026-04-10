package formatters

import (
	"fmt"
	"time"
)

func FormatDate(t time.Time) string {
	return t.Format("02.01.2006")
}

func FormatDateTime(t time.Time) string {
	return t.Format("02.01.2006 15:04")
}

func FormatDateRange(start, end time.Time) string {
	return fmt.Sprintf("%s — %s", FormatDate(start), FormatDate(end))
}

func FormatRelativeDate(t time.Time) string {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	diff := today.Sub(date).Hours() / 24

	switch {
	case diff == 0:
		return "Сегодня"
	case diff == 1:
		return "Вчера"
	case diff == -1:
		return "Завтра"
	case diff > 1 && diff < 7:
		return fmt.Sprintf("%d дня назад", int(diff))
	case diff < -1 && diff > -7:
		return fmt.Sprintf("Через %d дня", int(-diff))
	default:
		return FormatDate(t)
	}
}

func DaysAgo(t time.Time) int {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)

	return int(today.Sub(date).Hours() / 24)
}
