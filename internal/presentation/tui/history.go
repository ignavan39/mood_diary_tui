package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/application/usecase"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

type HistoryModel struct {
	service    *usecase.MoodService
	translator i18n.Translator
	entries    []*entity.MoodEntry
	cursor     int
	loading    bool
	errorMsg   string
}

func (m *HistoryModel) t(key string, args ...any) string {
	if m.translator == nil {
		return key
	}
	return m.translator.T(key, args...)
}

func NewHistoryModel(service *usecase.MoodService, translator i18n.Translator) *HistoryModel {
	return &HistoryModel{
		service:    service,
		translator: translator,
		cursor:     0,
		loading:    true,
	}
}

func (m *HistoryModel) Init() tea.Cmd {
	return m.loadHistory()
}

func (m *HistoryModel) GetSelectedEntry() *entity.MoodEntry {
	if m.cursor >= 0 && m.cursor < len(m.entries) {
		return m.entries[m.cursor]
	}
	return nil
}

func (m *HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "enter", " ", "e":
			if len(m.entries) > 0 && m.cursor < len(m.entries) {
				return m, NavigateToEdit(m.entries[m.cursor])
			}
		case "r":
			return m, m.loadHistory()
		case "esc", "q":
			return m, Navigate(ScreenMenu)
		}

	case HistoryLoadedMsg:
		m.entries = msg.Entries
		m.loading = false
		m.errorMsg = ""

		if m.cursor >= len(m.entries) {
			m.cursor = len(m.entries) - 1
		}
		if m.cursor < 0 && len(m.entries) > 0 {
			m.cursor = 0
		}

	case ErrorMsg:
		m.errorMsg = msg.Error.Error()
		m.loading = false
	}

	return m, nil
}

func (m *HistoryModel) loadHistory() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		ctx := context.Background()
		entries, err := m.service.GetRecentMoods(ctx, 30)
		if err != nil {
			return ErrorMsg{Error: err}
		}
		return HistoryLoadedMsg{Entries: entries}
	}
}

func (m *HistoryModel) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render(m.t("history.title"))
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.errorMsg != "" {

		errMsg := styles.ErrorStyle.Render(m.t("common.error_prefix") + m.errorMsg)
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	if m.loading {

		b.WriteString(styles.InfoStyle.Render(m.t("common.loading")))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if len(m.entries) == 0 {

		b.WriteString(styles.InfoStyle.Render(m.t("common.no_entries")))
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render(m.t("help.navigation.history"))
		b.WriteString(help)
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	b.WriteString(m.renderEntries())
	b.WriteString("\n\n")

	help := styles.HelpStyle.Render(m.t("help.navigation.history"))
	b.WriteString(help)

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

func (m *HistoryModel) renderEntries() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(styles.PastelLavender).
		Bold(true)

	header := fmt.Sprintf("%-12s  %-4s  %-15s  %-30s",
		m.t("history.header_date"),
		"",
		m.t("history.header_mood"),
		m.t("history.header_note"),
	)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	separator := strings.Repeat("─", 70)
	b.WriteString(lipgloss.NewStyle().
		Foreground(styles.PastelGray).
		Render(separator))
	b.WriteString("\n")

	for i, entry := range m.entries {
		line := m.renderEntry(entry, i == m.cursor)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m *HistoryModel) renderEntry(entry *entity.MoodEntry, selected bool) string {
	dateStr := entry.Date.Format("02.01.2006")

	emoji := entry.Level.Emoji()

	moodDesc := m.t(entry.Level.StringKey())
	moodText := fmt.Sprintf("%s %2d/10", moodDesc, entry.Level.Int())

	note := entry.Note
	maxNoteLen := 30
	if len(note) > maxNoteLen {
		note = note[:maxNoteLen-3] + "..."
	}
	if note == "" {

		note = lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Italic(true).
			Render(m.t("common.no_note"))
	}

	line := fmt.Sprintf("%-12s  %-4s  %-15s  %-30s",
		dateStr,
		emoji,
		moodText,
		note,
	)

	if selected {
		return styles.SelectedListItemStyle.Render("→ " + line)
	}

	color := styles.GetMoodColor(entry.Level.Int())
	return lipgloss.NewStyle().
		Foreground(color).
		Render("  " + line)
}

type HistoryLoadedMsg struct {
	Entries []*entity.MoodEntry
}

type NavigateToEditMsg struct {
	Entry *entity.MoodEntry
}

func NavigateToEdit(entry *entity.MoodEntry) tea.Cmd {
	return func() tea.Msg {
		return NavigateToEditMsg{Entry: entry}
	}
}
