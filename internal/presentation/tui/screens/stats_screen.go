package screens

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
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type StatsScreen struct {
	state.BaseState

	service    *usecase.MoodService
	translator i18n.Translator

	period  usecase.Period
	stats   *repository.MoodStatistics
	entries []*entity.MoodEntry
}

func NewStatsScreen(service *usecase.MoodService, translator i18n.Translator) *StatsScreen {
	return &StatsScreen{
		service:    service,
		translator: translator,
		period:     usecase.PeriodMonth,
	}
}

func (s *StatsScreen) t(key string, args ...interface{}) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *StatsScreen) Init() tea.Cmd {
	s.SetLoading(true)
	return s.loadStats()
}

func (s *StatsScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		return s.handleKeyMsg(msg)

	case state.StatsLoadedMsg:
		s.stats = msg.Stats
		s.entries = msg.Entries
		s.SetLoading(false)
		s.ClearError()
		return s, nil

	case state.ErrorMsg:
		s.SetError(msg.Error)
		return s, nil
	}

	return s, nil
}

func (s *StatsScreen) handleKeyMsg(msg tea.KeyMsg) (state.Screen, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		s.changePeriod(-1)
		s.SetLoading(true)
		return s, s.loadStats()

	case "right", "l":
		s.changePeriod(1)
		s.SetLoading(true)
		return s, s.loadStats()

	case "esc", "q":
		return s, state.NavigateToMenu()
	}

	return s, nil
}

func (s *StatsScreen) changePeriod(delta int) {
	periods := []usecase.Period{
		usecase.PeriodWeek,
		usecase.PeriodMonth,
		usecase.PeriodQuarter,
		usecase.PeriodYear,
		usecase.PeriodAll,
	}

	for i, p := range periods {
		if p == s.period {
			newIndex := i + delta
			if newIndex >= 0 && newIndex < len(periods) {
				s.period = periods[newIndex]
			}
			break
		}
	}
}

func (s *StatsScreen) loadStats() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		stats, err := s.service.GetStatistics(ctx, s.period)
		if err != nil {
			return state.ErrorMsg{Error: err}
		}

		start, end := s.period.DateRange()
		entries, err := s.service.GetMoodsByDateRange(ctx, start, end)
		if err != nil {
			return state.ErrorMsg{Error: err}
		}

		return state.StatsLoadedMsg{
			Stats:   stats,
			Entries: entries,
		}
	}
}

func (s *StatsScreen) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render(s.t("stats.title"))
	b.WriteString(header)
	b.WriteString("\n\n")

	if s.Error != nil {
		errMsg := styles.ErrorStyle.Render(s.t("common.error_prefix") + s.Error.Error())
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	b.WriteString(s.renderPeriodSelector())
	b.WriteString("\n\n")

	if s.Loading {
		b.WriteString(styles.InfoStyle.Render(s.t("common.loading")))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if s.stats == nil || s.stats.Count == 0 {
		b.WriteString(styles.InfoStyle.Render(s.t("stats.no_data")))
		b.WriteString("\n\n")
		help := styles.HelpStyle.Render(s.t("help.navigation.stats"))
		b.WriteString(help)
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	b.WriteString(s.renderStatsCards())
	b.WriteString("\n\n")

	b.WriteString(s.renderDistribution())
	b.WriteString("\n\n")

	help := styles.HelpStyle.Render(s.t("help.navigation.stats"))
	b.WriteString(help)

	return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
}

func (s *StatsScreen) renderPeriodSelector() string {
	var items []string

	periods := []usecase.Period{
		usecase.PeriodWeek,
		usecase.PeriodMonth,
		usecase.PeriodQuarter,
		usecase.PeriodYear,
		usecase.PeriodAll,
	}

	for _, p := range periods {
		text := s.PeriodLabel(p)
		if p == s.period {
			items = append(items, styles.SelectedButtonStyle.Render(text))
		} else {
			items = append(items, styles.ButtonStyle.Render(text))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, items...)
}

func (s *StatsScreen) renderStatsCards() string {
	if s.stats == nil {
		return ""
	}

	totalCard := s.createStatCard(
		s.t("stats.total_entries"),
		fmt.Sprintf("%d", s.stats.Count),
		styles.PastelSky,
	)

	avgCard := s.createStatCard(
		s.t("stats.average"),
		fmt.Sprintf("%.1f", s.stats.Average),
		styles.GetMoodColor(int(s.stats.Average+0.5)),
	)

	rangeCard := s.createStatCard(
		s.t("stats.range"),
		fmt.Sprintf("%d - %d", s.stats.MinLevel, s.stats.MaxLevel),
		styles.PastelMint,
	)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		totalCard, "  ",
		avgCard, "  ",
		rangeCard,
	)
}

func (s *StatsScreen) createStatCard(title, value string, color lipgloss.Color) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Bold(true)

	valueStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)

	content := titleStyle.Render(title) + "\n" + valueStyle.Render(value)

	return styles.BoxStyle.Render(content)
}

func (s *StatsScreen) PeriodLabel(p usecase.Period) string {
	switch p {
	case usecase.PeriodWeek:
		return s.translator.T("stats.week")
	case usecase.PeriodMonth:
		return s.translator.T("stats.month")
	case usecase.PeriodQuarter:
		return s.translator.T("stats.quarter")
	case usecase.PeriodYear:
		return s.translator.T("stats.year")
	case usecase.PeriodAll:
		return s.translator.T("stats.all")
	default:
		return "Unknown"
	}
}

func (s *StatsScreen) renderDistribution() string {
	if s.stats == nil || len(s.stats.Distribution) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(styles.SubtitleStyle.Render(s.t("stats.distribution")))
	b.WriteString("\n\n")

	maxCount := 0
	for _, count := range s.stats.Distribution {
		if count > maxCount {
			maxCount = count
		}
	}

	for level := 0; level <= 10; level++ {
		count := s.stats.Distribution[level]
		if maxCount == 0 {
			maxCount = 1
		}

		barWidth := int(float64(count) / float64(maxCount) * 20)
		if count > 0 && barWidth == 0 {
			barWidth = 1
		}

		emoji := entity.MoodLevel(level).Emoji()
		bar := strings.Repeat("█", barWidth)

		color := styles.GetMoodColor(level)
		barStyle := lipgloss.NewStyle().Foreground(color)

		b.WriteString(fmt.Sprintf("%s %2d: %s %d\n", emoji, level, barStyle.Render(bar), count))
	}

	return b.String()
}
