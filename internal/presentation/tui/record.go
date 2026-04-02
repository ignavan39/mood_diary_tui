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
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

type RecordModel struct {
	service       *usecase.MoodService
	moodLevel     int
	noteInput     textinput.Model
	step          int
	existingEntry *entity.MoodEntry
	isUpdate      bool
	success       bool
	errorMsg      string
}

func NewRecordModel(service *usecase.MoodService) *RecordModel {
	ti := textinput.New()
	ti.Placeholder = "Добавьте заметку (необязательно)..."
	ti.CharLimit = 200
	ti.Width = 50

	return &RecordModel{
		service:   service,
		moodLevel: 5,
		noteInput: ti,
		step:      0,
	}
}

func (m *RecordModel) Init() tea.Cmd {
	return m.checkExistingEntry()
}

func (m *RecordModel) checkExistingEntry() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		entry, err := m.service.GetTodayMood(ctx)
		if err == nil && entry != nil {
			return ExistingEntryMsg{Entry: entry}
		}
		return nil
	}
}

func (m *RecordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case ExistingEntryMsg:

		m.existingEntry = msg.Entry
		m.isUpdate = true
		m.moodLevel = msg.Entry.Level.Int()
		m.noteInput.SetValue(msg.Entry.Note)
		return m, nil

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
			case "esc", "q":
				return m, Navigate(ScreenMenu)
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
				return m, m.saveMood()
			case "n", "N", "esc":
				m.step = 0
				return m, nil
			}
		}

	case SavedMsg:
		m.success = true
		m.errorMsg = ""

		return m, tea.Batch(
			tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
				return NavigateMsg{Screen: ScreenMenu}
			}),
		)

	case ErrorMsg:
		m.errorMsg = msg.Error.Error()
		m.success = false
		m.step = 0
		return m, nil

	default:

	}

	return m, nil
}

func (m *RecordModel) saveMood() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var err error
		if m.isUpdate {

			err = m.service.UpdateMood(ctx, time.Now(), m.moodLevel, m.noteInput.Value())
		} else {

			err = m.service.RecordMood(ctx, m.moodLevel, m.noteInput.Value(), nil)
		}

		if err != nil {
			return ErrorMsg{Error: err}
		}
		return SavedMsg{}
	}
}

func (m *RecordModel) View() string {
	var b strings.Builder

	headerText := "📝 Записать настроение"
	if m.isUpdate {
		headerText = "✏️ Изменить настроение"
	}
	header := styles.HeaderStyle.Render(headerText)
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.isUpdate && m.step == 0 {
		info := styles.InfoStyle.Render("ℹ️  Запись за сегодня уже существует. Вы можете её изменить.")
		b.WriteString(info)
		b.WriteString("\n\n")
	}

	if m.success {
		successText := "✓ Настроение успешно записано!"
		if m.isUpdate {
			successText = "✓ Настроение успешно обновлено!"
		}
		success := styles.SuccessStyle.Render(successText)
		b.WriteString(success)
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Возврат в меню..."))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if m.errorMsg != "" {
		errMsg := styles.ErrorStyle.Render("✗ Ошибка: " + m.errorMsg)
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	if m.step == 0 {
		b.WriteString(styles.SubtitleStyle.Render("Как вы себя чувствуете сегодня?"))
		b.WriteString("\n\n")

		moodLevel, _ := entity.NewMoodLevel(m.moodLevel)

		b.WriteString(m.renderMoodScale())
		b.WriteString("\n\n")

		current := fmt.Sprintf("%s %s (%d/10)",
			moodLevel.Emoji(),
			moodLevel.String(),
			m.moodLevel)
		currentStyle := styles.MoodStyle(m.moodLevel)
		b.WriteString(currentStyle.Render(current))
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render("←/→: Изменить • Enter: Далее • Esc: Назад")
		b.WriteString(help)
	}

	if m.step == 1 {
		b.WriteString(styles.SubtitleStyle.Render("Добавьте заметку (необязательно):"))
		b.WriteString("\n\n")

		b.WriteString(m.noteInput.View())
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render("Enter: Подтвердить • Esc: Назад")
		b.WriteString(help)
	}

	if m.step == 2 {
		confirmText := "Подтвердите запись:"
		if m.isUpdate {
			confirmText = "Подтвердите изменение:"
		}
		b.WriteString(styles.SubtitleStyle.Render(confirmText))
		b.WriteString("\n\n")

		moodLevel, _ := entity.NewMoodLevel(m.moodLevel)

		note := m.noteInput.Value()
		if note == "" {
			note = "(без заметки)"
		}

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.GetMoodColor(m.moodLevel)).
			Padding(1, 2).
			Render(fmt.Sprintf(
				"Настроение: %s %s (%d/10)\nЗаметка: %s\nДата: %s",
				moodLevel.Emoji(),
				moodLevel.String(),
				m.moodLevel,
				note,
				time.Now().Format("02.01.2006"),
			))

		b.WriteString(box)
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render("y/Enter: Сохранить • n/Esc: Отмена")
		b.WriteString(help)
	}

	if m.step == 3 {
		b.WriteString(styles.InfoStyle.Render("Сохранение..."))
	}

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

func (m *RecordModel) renderMoodScale() string {
	var parts []string

	for i := 0; i <= 10; i++ {
		moodLevel, _ := entity.NewMoodLevel(i)
		emoji := moodLevel.Emoji()

		if i == m.moodLevel {

			selected := lipgloss.NewStyle().
				Foreground(styles.TextLight).
				Background(styles.GetMoodColor(i)).
				Border(lipgloss.ThickBorder()).
				BorderForeground(lipgloss.Color("#4A4A4A")).
				Padding(0, 1).
				Bold(false).
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

type ExistingEntryMsg struct {
	Entry *entity.MoodEntry
}

type SavedMsg struct{}
