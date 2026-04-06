package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/application/usecase"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

type EditModel struct {
	service    *usecase.MoodService
	translator i18n.Translator
	entry      *entity.MoodEntry
	moodLevel  int
	noteInput  textinput.Model
	step       int
	showDelete bool
	success    bool
	deleted    bool
	errorMsg   string
}

func (m *EditModel) t(key string, args ...any) string {
	if m.translator == nil {
		return key
	}
	return m.translator.T(key, args...)
}

func NewEditModel(service *usecase.MoodService, translator i18n.Translator) *EditModel {
	ti := textinput.New()

	ti.Placeholder = translator.T("record.prompt_note")
	ti.CharLimit = 200
	ti.Width = 50

	return &EditModel{
		service:    service,
		translator: translator,
		noteInput:  ti,
		step:       0,
	}
}

func (m *EditModel) SetEntry(entry *entity.MoodEntry) {
	m.entry = entry
	m.moodLevel = entry.Level.Int()
	m.noteInput.SetValue(entry.Note)
	m.step = 0
	m.success = false
	m.deleted = false
	m.errorMsg = ""
	m.showDelete = false
}

func (m *EditModel) Init() tea.Cmd {
	return nil
}

func (m *EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case 0:
			switch msg.String() {
			case "left", "h":
				if m.moodLevel > 0 {
					m.moodLevel--
				}
			case "right", "l":
				if m.moodLevel < 10 {
					m.moodLevel++
				}
			case "enter":
				m.step = 1
				m.noteInput.Focus()
				return m, textinput.Blink
			case "d":
				m.showDelete = true
				m.step = 4
			case "esc", "q":
				return m, Navigate(ScreenHistory)
			}

		case 1:
			switch msg.String() {
			case "enter":
				m.step = 2
				m.noteInput.Blur()
				return m, nil
			case "esc":
				m.step = 0
				m.noteInput.Blur()
				return m, nil
			default:
				m.noteInput, cmd = m.noteInput.Update(msg)
				return m, cmd
			}

		case 2:
			switch msg.String() {
			case "y", "Y", "enter":
				m.step = 3
				return m, m.saveEdit()
			case "n", "N", "esc":
				m.step = 0
				return m, nil
			}

		case 4:
			switch msg.String() {
			case "y", "Y":
				return m, m.deleteEntry()
			case "n", "N", "esc":
				m.showDelete = false
				m.step = 0
				return m, nil
			}
		}

	case SavedMsg:
		m.success = true
		m.errorMsg = ""
		return m, tea.Batch(
			tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
				return NavigateMsg{Screen: ScreenHistory}
			}),
		)

	case DeletedMsg:
		m.deleted = true
		m.errorMsg = ""
		return m, tea.Batch(
			tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
				return NavigateMsg{Screen: ScreenHistory}
			}),
		)

	case ErrorMsg:
		m.errorMsg = msg.Error.Error()
		m.success = false
		m.step = 0
		return m, nil
	}

	return m, nil
}

func (m *EditModel) saveEdit() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.service.UpdateMood(ctx, m.entry.Date, m.moodLevel, m.noteInput.Value())
		if err != nil {
			return ErrorMsg{Error: err}
		}
		return SavedMsg{}
	}
}

func (m *EditModel) deleteEntry() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.service.DeleteMood(ctx, m.entry.Date)
		if err != nil {
			return ErrorMsg{Error: err}
		}
		return DeletedMsg{}
	}
}

func (m *EditModel) View() string {

	if m.entry == nil {
		return lipgloss.NewStyle().
			Padding(2, 4).
			Render(m.t("edit.no_entry_selected"))
	}

	var b strings.Builder

	header := styles.HeaderStyle.Render(m.t("edit.title"))
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.success {

		success := styles.SuccessStyle.Render(m.t("edit.success_updated"))
		b.WriteString(success)
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render(m.t("edit.returning_to_history")))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if m.deleted {

		deleted := styles.SuccessStyle.Render(m.t("edit.success_deleted"))
		b.WriteString(deleted)
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render(m.t("edit.returning_to_history")))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if m.errorMsg != "" {

		errMsg := styles.ErrorStyle.Render(m.t("common.error_prefix") + m.errorMsg)
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	dateInfo := fmt.Sprintf("%s %s",
		m.t("common.date_label"),
		m.entry.Date.Format("02.01.2006"))
	b.WriteString(styles.InfoStyle.Render(dateInfo))
	b.WriteString("\n\n")

	if m.step == 0 {

		b.WriteString(styles.SubtitleStyle.Render(m.t("edit.prompt_select_mood")))
		b.WriteString("\n\n")

		moodLevel, _ := entity.NewMoodLevel(m.moodLevel)

		b.WriteString(m.renderMoodScale())
		b.WriteString("\n\n")

		current := fmt.Sprintf(m.t("record.box_mood"),
			moodLevel.Emoji(),
			m.t(moodLevel.StringKey()),
			m.moodLevel)
		currentStyle := styles.MoodStyle(m.moodLevel)
		b.WriteString(currentStyle.Render(current))
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render(m.t("help.navigation.edit_step0"))
		b.WriteString(help)
	}

	if m.step == 1 {

		b.WriteString(styles.SubtitleStyle.Render(m.t("edit.prompt_edit_note")))
		b.WriteString("\n\n")

		b.WriteString(m.noteInput.View())
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render(m.t("help.navigation.edit_step1"))
		b.WriteString(help)
	}

	if m.step == 2 {

		b.WriteString(styles.SubtitleStyle.Render(m.t("edit.prompt_confirm_changes")))
		b.WriteString("\n\n")

		moodLevel, _ := entity.NewMoodLevel(m.moodLevel)

		note := m.noteInput.Value()
		if note == "" {

			note = m.t("common.no_note")
		}

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.GetMoodColor(m.moodLevel)).
			Padding(1, 2).
			Render(fmt.Sprintf(
				m.t("record.box_mood")+"\n"+
					m.t("record.box_note")+"\n"+
					m.t("record.box_date"),
				moodLevel.Emoji(),
				m.t(moodLevel.StringKey()),
				m.moodLevel,
				note,
				m.entry.Date.Format("02.01.2006"),
			))

		b.WriteString(box)
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render(m.t("help.navigation.edit_step2"))
		b.WriteString(help)
	}

	if m.step == 3 {

		b.WriteString(styles.InfoStyle.Render(m.t("common.loading")))
	}

	if m.step == 4 && m.showDelete {

		b.WriteString(styles.SubtitleStyle.Render(m.t("edit.delete_title")))
		b.WriteString("\n\n")

		warning := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ErrorRed).
			Padding(1, 2).
			Render(fmt.Sprintf(
				m.t("edit.delete_warning"),
				m.entry.Date.Format("02.01.2006"),
			))

		b.WriteString(warning)
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render(m.t("help.navigation.delete_confirm"))
		b.WriteString(help)
	}

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

func (m *EditModel) renderMoodScale() string {
	var parts []string

	for i := 0; i <= 10; i++ {
		moodLevel, _ := entity.NewMoodLevel(i)
		emoji := moodLevel.Emoji()

		if i == m.moodLevel {
			selected := lipgloss.NewStyle().
				Foreground(styles.TextLight).
				Background(styles.GetMoodColor(i)).
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#000000")).
				Bold(true).
				Padding(0, 1).
				Render(emoji)
			parts = append(parts, selected)
		} else {
			unselected := lipgloss.NewStyle().
				Foreground(styles.GetMoodColor(i)).
				Faint(true).
				Padding(0, 1).
				Render(emoji)
			parts = append(parts, unselected)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, parts...)
}

type DeletedMsg struct{}
