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
	"github.com/ignavan39/mood-diary/internal/presentation/tui/components"
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

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.PastelLavender).
		Align(lipgloss.Center).
		Width(s.Width)

	header := headerStyle.Render("📊 " + s.t("stats.title"))

	if s.Loading {
		loading := components.NewLoading(s.t("common.loading"))
		return lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			"",
			loading.View(),
		)
	}

	if s.Error != nil {
		errorMsg := components.NewError(s.Error.Error())
		return lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			"",
			errorMsg.View(),
		)
	}

	if s.stats == nil {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			"",
			s.t("stats.no_data"),
		)
	}

	periodStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		Align(lipgloss.Center).
		Width(s.Width)

	period := periodStyle.Render("◀ " + s.period.String() + " ▶")

	var stats strings.Builder
	stats.WriteString(fmt.Sprintf("📝 %s: %d\n", s.t("stats.total_entries"), s.stats.Count))
	stats.WriteString(fmt.Sprintf("📈 %s: %.1f\n", s.t("stats.average"), s.stats.Average))
	stats.WriteString(fmt.Sprintf("🔽 %s: %d\n", s.t("stats.min"), s.stats.MinLevel))
	stats.WriteString(fmt.Sprintf("🔼 %s: %d\n", s.t("stats.max"), s.stats.MaxLevel))

	if s.stats.TotalDays > 0 {
		completion := float64(s.stats.Count) / float64(s.stats.TotalDays) * 100
		stats.WriteString(fmt.Sprintf("✓ %s: %.0f%%\n", s.t("stats.completion"), completion))
	}

	statsStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PastelLavender).
		Padding(1, 2).
		Margin(1, 0)

	statsContent := statsStyle.Render(stats.String())

	distribution := s.renderDistribution()

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		Align(lipgloss.Center).
		Width(s.Width).
		Padding(1, 0)

	help := helpStyle.Render(s.t("help.navigation.stats"))

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		"",
		period,
		"",
		statsContent,
		"",
		distribution,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Padding(2, 4).
		Width(s.Width).
		Render(content)
}

func (s *StatsScreen) renderDistribution() string {
	if s.stats == nil || len(s.stats.Distribution) == 0 {
		return ""
	}

	var bars strings.Builder
	bars.WriteString(s.t("stats.distribution") + ":\n\n")

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

		bars.WriteString(fmt.Sprintf("%s %2d: %s %d\n", emoji, level, barStyle.Render(bar), count))
	}

	return bars.String()
}
