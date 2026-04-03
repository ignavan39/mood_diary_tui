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

type EditModel struct {
	service    *usecase.MoodService
	entry      *entity.MoodEntry
	moodLevel  int
	noteInput  textinput.Model
	step       int
	showDelete bool
	success    bool
	deleted    bool
	errorMsg   string
}

func NewEditModel(service *usecase.MoodService) *EditModel {
	ti := textinput.New()
	ti.Placeholder = "Добавьте заметку (необязательно)..."
	ti.CharLimit = 200
	ti.Width = 50

	return &EditModel{
		service:   service,
		noteInput: ti,
		step:      0,
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

	default:

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
			Render("Запись не выбрана. Нажмите Esc для возврата.")
	}

	var b strings.Builder

	header := styles.HeaderStyle.Render("✏️ Редактировать запись")
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.success {
		success := styles.SuccessStyle.Render("✓ Запись успешно обновлена!")
		b.WriteString(success)
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Возврат к истории..."))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if m.deleted {
		deleted := styles.SuccessStyle.Render("✓ Запись успешно удалена!")
		b.WriteString(deleted)
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Возврат к истории..."))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if m.errorMsg != "" {
		errMsg := styles.ErrorStyle.Render("✗ Ошибка: " + m.errorMsg)
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	dateInfo := fmt.Sprintf("Дата: %s", m.entry.Date.Format("02.01.2006"))
	b.WriteString(styles.InfoStyle.Render(dateInfo))
	b.WriteString("\n\n")

	if m.step == 0 {
		b.WriteString(styles.SubtitleStyle.Render("Выберите новое настроение:"))
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

		help := styles.HelpStyle.Render("←/→: Изменить • Enter: Далее • d: Удалить • Esc: Назад")
		b.WriteString(help)
	}

	if m.step == 1 {
		b.WriteString(styles.SubtitleStyle.Render("Измените заметку:"))
		b.WriteString("\n\n")

		b.WriteString(m.noteInput.View())
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render("Enter: Подтвердить • Esc: Назад")
		b.WriteString(help)
	}

	if m.step == 2 {
		b.WriteString(styles.SubtitleStyle.Render("Подтвердите изменения:"))
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
				m.entry.Date.Format("02.01.2006"),
			))

		b.WriteString(box)
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render("y/Enter: Сохранить • n/Esc: Отмена")
		b.WriteString(help)
	}

	if m.step == 3 {
		b.WriteString(styles.InfoStyle.Render("Сохранение..."))
	}

	if m.step == 4 && m.showDelete {
		b.WriteString(styles.SubtitleStyle.Render("⚠️  Удалить эту запись?"))
		b.WriteString("\n\n")

		warning := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.ErrorRed).
			Padding(1, 2).
			Render(fmt.Sprintf(
				"Запись от %s будет удалена без возможности восстановления.\n\nВы уверены?",
				m.entry.Date.Format("02.01.2006"),
			))

		b.WriteString(warning)
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render("y: Да, удалить • n/Esc: Отмена")
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
