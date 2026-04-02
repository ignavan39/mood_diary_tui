package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/application/usecase"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/domain/repository"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

type StatsModel struct {
	service  *usecase.MoodService
	period   usecase.Period
	periods  []usecase.Period
	cursor   int
	stats    *repository.MoodStatistics
	entries  []*entity.MoodEntry
	loading  bool
	errorMsg string
}

func NewStatsModel(service *usecase.MoodService) *StatsModel {
	m := &StatsModel{
		service: service,
		period:  usecase.PeriodWeek,
		periods: []usecase.Period{
			usecase.PeriodWeek,
			usecase.PeriodMonth,
			usecase.PeriodQuarter,
			usecase.PeriodYear,
			usecase.PeriodAll,
		},
		cursor:  0,
		loading: true,
	}
	return m
}

func (m *StatsModel) Init() tea.Cmd {
	return m.loadStats()
}

func (m *StatsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
				m.period = m.periods[m.cursor]
				return m, m.loadStats()
			}
		case "right", "l":
			if m.cursor < len(m.periods)-1 {
				m.cursor++
				m.period = m.periods[m.cursor]
				return m, m.loadStats()
			}
		case "r":
			return m, m.loadStats()
		case "esc", "q":
			return m, Navigate(ScreenMenu)
		}

	case StatsLoadedMsg:
		m.stats = msg.Stats
		m.entries = msg.Entries
		m.loading = false
		m.errorMsg = ""

	case ErrorMsg:
		m.errorMsg = msg.Error.Error()
		m.loading = false

	default:

	}

	return m, nil
}

func (m *StatsModel) loadStats() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx := context.Background()

		stats, err := m.service.GetStatistics(ctx, m.period)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		entries, err := m.service.GetMoodsForPeriod(ctx, m.period)
		if err != nil {
			return ErrorMsg{Error: err}
		}

		return StatsLoadedMsg{
			Stats:   stats,
			Entries: entries,
		}
	}
}

func (m *StatsModel) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render("📊 Статистика настроений")
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.errorMsg != "" {
		errMsg := styles.ErrorStyle.Render("✗ Ошибка: " + m.errorMsg)
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	b.WriteString(m.renderPeriodSelector())
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(styles.InfoStyle.Render("Загрузка..."))
		return b.String()
	}

	if m.stats == nil || m.stats.TotalEntries == 0 {
		b.WriteString(styles.InfoStyle.Render("Нет данных за выбранный период"))
		b.WriteString("\n\n")
		help := styles.HelpStyle.Render("←/→: Изменить период • r: Обновить • Esc: Назад")
		b.WriteString(help)
		return b.String()
	}

	b.WriteString(m.renderStatsCards())
	b.WriteString("\n\n")

	b.WriteString(m.renderMoodDistribution())
	b.WriteString("\n\n")

	b.WriteString(m.renderRecentTrend())
	b.WriteString("\n\n")

	help := styles.HelpStyle.Render("←/→: Изменить период • r: Обновить • Esc: Назад")
	b.WriteString(help)

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

func (m *StatsModel) renderPeriodSelector() string {
	var items []string

	for i, period := range m.periods {
		text := period.String()
		if i == m.cursor {
			items = append(items, styles.SelectedButtonStyle.Render(text))
		} else {
			items = append(items, styles.ButtonStyle.Render(text))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, items...)
}

func (m *StatsModel) renderStatsCards() string {
	if m.stats == nil {
		return ""
	}

	totalCard := m.createStatCard(
		"Всего записей",
		fmt.Sprintf("%d", m.stats.TotalEntries),
		styles.PastelSky,
	)

	avgMood := m.stats.AverageMood
	avgEmoji := ""
	if m.stats.TotalEntries > 0 {
		avgLevel, _ := entity.NewMoodLevel(int(avgMood + 0.5))
		avgEmoji = avgLevel.Emoji()
	}
	avgCard := m.createStatCard(
		"Средний уровень",
		fmt.Sprintf("%.1f %s", avgMood, avgEmoji),
		styles.GetMoodColor(int(avgMood+0.5)),
	)

	trendText := "─"
	trendColor := styles.PastelGray
	if m.stats.Trend > 0.5 {
		trendText = "↑ Улучшается"
		trendColor = styles.SuccessGreen
	} else if m.stats.Trend < -0.5 {
		trendText = "↓ Ухудшается"
		trendColor = styles.ErrorRed
	} else {
		trendText = "─ Стабильно"
	}
	trendCard := m.createStatCard(
		"Тренд",
		trendText,
		trendColor,
	)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		totalCard, "  ",
		avgCard, "  ",
		trendCard,
	)
}

func (m *StatsModel) createStatCard(title, value string, color lipgloss.Color) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)

	content := titleStyle.Render(title) + "\n" + valueStyle.Render(value)

	return styles.BoxStyle.Render(content)
}

func (m *StatsModel) renderMoodDistribution() string {
	if m.stats == nil || len(m.stats.MoodCounts) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(styles.SubtitleStyle.Render("Распределение настроений"))
	b.WriteString("\n\n")

	maxCount := 0
	for _, count := range m.stats.MoodCounts {
		if count > maxCount {
			maxCount = count
		}
	}

	for level := 10; level >= 0; level-- {
		moodLevel, _ := entity.NewMoodLevel(level)
		count := m.stats.MoodCounts[moodLevel]

		barWidth := 0
		if maxCount > 0 {
			barWidth = (count * 30) / maxCount
		}

		bar := strings.Repeat("█", barWidth)
		barStyle := lipgloss.NewStyle().
			Foreground(styles.GetMoodColor(level))

		line := fmt.Sprintf("%2d %s  %s %s (%d)",
			level,
			moodLevel.Emoji(),
			barStyle.Render(bar),
			lipgloss.NewStyle().Foreground(styles.TextMuted).Render(moodLevel.String()),
			count,
		)

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m *StatsModel) renderRecentTrend() string {
	if len(m.entries) == 0 {
		return ""
	}

	var b strings.Builder

	b.WriteString(styles.SubtitleStyle.Render("Динамика настроения"))
	b.WriteString("\n\n")

	limit := 14
	if len(m.entries) < limit {
		limit = len(m.entries)
	}

	entries := make([]*entity.MoodEntry, limit)
	for i := 0; i < limit; i++ {
		entries[i] = m.entries[limit-1-i]
	}

	values := make([]float64, len(entries))
	for i, entry := range entries {
		values[i] = float64(entry.Level.Int())
	}

	var sum float64
	minVal, maxVal := values[0], values[0]
	for _, v := range values {
		sum += v
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	avg := sum / float64(len(values))

	scaleStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)
	b.WriteString(scaleStyle.Render("10 │"))
	b.WriteString("\n")
	b.WriteString(scaleStyle.Render(" 5 │"))
	b.WriteString("\n")
	b.WriteString(scaleStyle.Render(" 0 │"))
	b.WriteString("\n   └")

	sparkline := styles.Sparkline(values, len(values))
	b.WriteString(sparkline)
	b.WriteString("\n")

	if len(entries) > 0 {
		info := fmt.Sprintf("%s — %s  │  Мин: %.0f  Сред: %.1f  Макс: %.0f",
			entries[0].Date.Format("02.01"),
			entries[len(entries)-1].Date.Format("02.01"),
			minVal,
			avg,
			maxVal,
		)
		b.WriteString(styles.HelpStyle.Render(info))
	}

	return b.String()
}

type StatsLoadedMsg struct {
	Stats   *repository.MoodStatistics
	Entries []*entity.MoodEntry
}
