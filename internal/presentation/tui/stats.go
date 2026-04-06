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
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

type StatsModel struct {
	service    *usecase.MoodService
	translator i18n.Translator
	period     usecase.Period
	periods    []usecase.Period
	cursor     int
	stats      *repository.MoodStatistics
	entries    []*entity.MoodEntry
	loading    bool
	errorMsg   string
}

func (m *StatsModel) t(key string, args ...any) string {
	if m.translator == nil {
		return key
	}
	return m.translator.T(key, args...)
}

func NewStatsModel(service *usecase.MoodService, translator i18n.Translator) *StatsModel {
	m := &StatsModel{
		service:    service,
		translator: translator,
		period:     usecase.PeriodWeek,
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

	header := styles.HeaderStyle.Render(m.t("stats.title"))
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.errorMsg != "" {

		errMsg := styles.ErrorStyle.Render(m.t("common.error_prefix") + m.errorMsg)
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	b.WriteString(m.renderPeriodSelector())
	b.WriteString("\n\n")

	if m.loading {

		b.WriteString(styles.InfoStyle.Render(m.t("common.loading")))
		return b.String()
	}

	if m.stats == nil || m.stats.TotalEntries == 0 {

		b.WriteString(styles.InfoStyle.Render(m.t("common.no_data")))
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render(m.t("help.navigation.stats"))
		b.WriteString(help)
		return b.String()
	}

	b.WriteString(m.renderStatsCards())
	b.WriteString("\n\n")

	b.WriteString(m.renderMoodDistribution())
	b.WriteString("\n\n")

	b.WriteString(m.renderRecentTrend())
	b.WriteString("\n\n")

	help := styles.HelpStyle.Render(m.t("help.navigation.stats"))
	b.WriteString(help)

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

func (m *StatsModel) renderPeriodSelector() string {
	var items []string

	periodKeys := []string{
		"stats.period_week",
		"stats.period_month",
		"stats.period_quarter",
		"stats.period_year",
		"stats.period_all",
	}

	for i, period := range m.periods {
		_ = period
		text := m.t(periodKeys[i])
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
		m.t("stats.card_total"),
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
		m.t("stats.card_average"),
		fmt.Sprintf("%.1f %s", avgMood, avgEmoji),
		styles.GetMoodColor(int(avgMood+0.5)),
	)

	trendText := m.t("stats.trend_stable")
	trendColor := styles.PastelGray
	if m.stats.Trend > 0.5 {
		trendText = m.t("stats.trend_improving")
		trendColor = styles.SuccessGreen
	} else if m.stats.Trend < -0.5 {
		trendText = m.t("stats.trend_worsening")
		trendColor = styles.ErrorRed
	}
	trendCard := m.createStatCard(
		m.t("stats.card_trend"),
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

	b.WriteString(styles.SubtitleStyle.Render(m.t("stats.section_distribution")))
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
		desc := m.t(moodLevel.StringKey())
		line := fmt.Sprintf("%2d %s  %s %s (%d)",
			level,
			moodLevel.Emoji(),
			barStyle.Render(bar),
			lipgloss.NewStyle().Foreground(styles.TextMuted).Render(desc),
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

	b.WriteString(styles.SubtitleStyle.Render(m.t("stats.section_trend")))
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
	var sum float64
	minVal, maxVal := 10.0, 0.0

	for i, entry := range entries {
		val := float64(entry.Level.Int())
		values[i] = val
		sum += val
		if val < minVal {
			minVal = val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	avg := sum / float64(len(values))

	sparkline := styles.Sparkline(values)
	b.WriteString(sparkline)
	b.WriteString("\n\n")

	if len(entries) > 0 {

		info := fmt.Sprintf(m.t("stats.trend_info"),
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
