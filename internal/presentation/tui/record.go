package tui

import (
	"context"
	"fmt"
	"mood-diary/internal/application/usecase"
	"mood-diary/internal/domain/entity"
	"mood-diary/internal/presentation/styles"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RecordModel struct {
	service   *usecase.MoodService
	moodLevel int
	noteInput textinput.Model
	step      int
	success   bool
	errorMsg  string
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
	return nil
}

func (m *RecordModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			case "esc", "q":
				return m, Navigate(ScreenMenu)
			}

		case 1:
			switch msg.String() {
			case "enter":
				m.step = 2
				m.noteInput.Blur()
			case "esc":
				m.step = 0
				m.noteInput.Blur()
			default:
				m.noteInput, cmd = m.noteInput.Update(msg)
				return m, cmd
			}

		case 2:
			switch msg.String() {
			case "y", "enter":
				return m, m.saveMood()
			case "n", "esc":
				m.step = 0
			}
		}

	case SavedMsg:
		m.success = true
		m.errorMsg = ""
		return m, tea.Sequence(
			tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return Navigate(ScreenMenu)
			}),
		)

	case ErrorMsg:
		m.errorMsg = msg.Error.Error()
		m.step = 0
	}

	return m, nil
}

func (m *RecordModel) saveMood() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := m.service.RecordMood(ctx, m.moodLevel, m.noteInput.Value(), nil)
		if err != nil {
			return ErrorMsg{Error: err}
		}
		return SavedMsg{}
	}
}

func (m *RecordModel) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render("📝 Записать настроение")
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.success {
		success := styles.SuccessStyle.Render("✓ Настроение успешно записано!")
		b.WriteString(success)
		return b.String()
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
		b.WriteString(styles.SubtitleStyle.Render("Подтвердите запись:"))
		b.WriteString("\n\n")

		moodLevel, _ := entity.NewMoodLevel(m.moodLevel)

		box := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.GetMoodColor(m.moodLevel)).
			Padding(1, 2).
			Render(fmt.Sprintf(
				"Настроение: %s %s (%d/10)\nЗаметка: %s\nДата: %s",
				moodLevel.Emoji(),
				moodLevel.String(),
				m.moodLevel,
				m.noteInput.Value(),
				time.Now().Format("02.01.2006"),
			))

		b.WriteString(box)
		b.WriteString("\n\n")

		help := styles.HelpStyle.Render("y/Enter: Сохранить • n/Esc: Отмена")
		b.WriteString(help)
	}

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

func (m *RecordModel) renderMoodScale() string {
	scale := ""

	for i := 0; i <= 10; i++ {
		moodLevel, _ := entity.NewMoodLevel(i)
		emoji := moodLevel.Emoji()

		if i == m.moodLevel {
			scale += lipgloss.NewStyle().
				Foreground(styles.TextLight).
				Background(styles.GetMoodColor(i)).
				Bold(true).
				Padding(0, 1).
				Render(emoji)
		} else {
			scale += lipgloss.NewStyle().
				Foreground(styles.GetMoodColor(i)).
				Padding(0, 1).
				Render(emoji)
		}
	}

	return scale
}

type SavedMsg struct{}
