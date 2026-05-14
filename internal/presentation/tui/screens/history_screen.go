package screens

import (
	"context"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/application/usecase"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/components"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/formatters"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

const historyPageSize = 20

type HistoryScreen struct {
	state.BaseState

	ctx        context.Context
	service    *usecase.MoodService
	translator i18n.Translator

	entries    []*entity.MoodEntry
	cursor     int
	page       int
	hasMore    bool
	confirmDlg *components.ConfirmationDialog
}

func NewHistoryScreen(ctx context.Context, service *usecase.MoodService, translator i18n.Translator) *HistoryScreen {
	return &HistoryScreen{
		ctx:        ctx,
		service:    service,
		translator: translator,
		entries:    []*entity.MoodEntry{},
		page:       0,
	}
}

func (s *HistoryScreen) t(key string, args ...any) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *HistoryScreen) Init() tea.Cmd {
	s.SetLoading(true)
	return s.loadEntries()
}

func (s *HistoryScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	if s.confirmDlg != nil && !s.confirmDlg.IsDone() {
		cmd := s.confirmDlg.Update(msg)
		if s.confirmDlg.IsDone() {
			s.confirmDlg = nil
		}
		return s, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		return s.handleKeyMsg(msg)

	case state.ErrorMsg:
		s.SetError(msg.Error)
		return s, nil

	case state.HistoryLoadedMsg:
		s.entries = msg.Entries
		s.SetLoading(false)
		s.ClearError()
		return s, nil

	case state.MoodDeletedMsg:

		s.SetLoading(true)
		return s, s.loadEntries()
	}

	return s, nil
}

func (s *HistoryScreen) handleKeyMsg(msg tea.KeyMsg) (state.Screen, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}

	case "down", "j":
		if s.cursor < len(s.entries)-1 {
			s.cursor++
		}

	case "enter", "e":
		if len(s.entries) > 0 && s.cursor < len(s.entries) {
			entry := s.entries[s.cursor]
			return s, state.NavigateToMoodForm(entry.Date, entry)
		}

	case "d":
		if len(s.entries) > 0 && s.cursor < len(s.entries) {
			entry := s.entries[s.cursor]
			msg := fmt.Sprintf(s.t(i18n.EditDeleteWarningKey), formatters.FormatDate(entry.Date))
			s.confirmDlg = components.NewConfirmation(msg, s.translator, s.deleteMood(entry), nil)
			return s, nil
		}

	case "n":
		if s.hasMore {
			s.page++
			s.cursor = 0
			s.SetLoading(true)
			return s, s.loadEntries()
		}

	case "p":
		if s.page > 0 {
			s.page--
			s.cursor = 0
			s.SetLoading(true)
			return s, s.loadEntries()
		}

	case "esc", "q":
		return s, state.NavigateBack()
	}

	return s, nil
}

func (s *HistoryScreen) loadEntries() tea.Cmd {
	return func() tea.Msg {
		limit := historyPageSize + 1
		offset := s.page * historyPageSize

		entries, err := s.service.GetRecentMoodsWithOffset(s.ctx, limit, offset)

		if err != nil {
			return state.ErrorMsg{Error: err}
		}

		s.hasMore = len(entries) > historyPageSize

		if s.hasMore {
			entries = entries[:historyPageSize]
		}

		return state.HistoryLoadedMsg{Entries: entries}
	}
}

func (s *HistoryScreen) deleteMood(entry *entity.MoodEntry) tea.Cmd {
	return func() tea.Msg {
		err := s.service.DeleteMood(s.ctx, entry.Date)
		if err != nil {
			return state.ErrorMsg{Error: err}
		}
		return state.MoodDeletedMsg{Date: entry.Date}
	}
}

func (s *HistoryScreen) View() string {

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.PastelLavender).
		Align(lipgloss.Center).
		Width(s.Width)

	header := headerStyle.Render("📜 " + s.t(i18n.HistoryTitleKey))

	if s.confirmDlg != nil {
		return lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			"",
			s.confirmDlg.View(),
		)
	}

	if s.Loading {
		loading := components.NewLoading(s.t(i18n.CommonLoaderMessageKey))
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

	if len(s.entries) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9B9B9B")).
			Align(lipgloss.Center).
			Width(s.Width).
			Padding(2, 0)

		return lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			"",
			emptyStyle.Render(s.t(i18n.HistoryEmptyKey)),
		)
	}

	var listContent string
	for i, entry := range s.entries {
		itemStyle := styles.ListItemStyle

		if i == s.cursor {
			itemStyle = styles.SelectedListItemStyle
		}

		dateStr := formatters.FormatRelativeDate(entry.Date, s.translator)
		moodStr := formatters.FormatMoodLevel(entry.Level, s.translator)
		note := formatters.TruncateNote(entry.Note, 40)

		line := dateStr + " | " + moodStr + " | " + note
		listContent += itemStyle.Render(line) + "\n"
	}

	pageInfo := fmt.Sprintf("  %s %d", s.t(i18n.StatsPageKey), s.page+1)
	if !s.hasMore {
		pageInfo += " (last)"
	}
	pageStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Padding(0, 2)
	listContent += pageStyle.Render(pageInfo) + "\n"

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		Align(lipgloss.Center).
		Width(s.Width).
		Padding(1, 0)

	help := helpStyle.Render(s.t(i18n.HelpNavigationHistoryKey))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		listContent,
		help,
	)

	return lipgloss.NewStyle().
		Padding(2, 4).
		Width(s.Width).
		Render(content)
}
