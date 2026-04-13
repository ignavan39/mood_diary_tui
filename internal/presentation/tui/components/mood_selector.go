package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

type MoodSelector struct {
	value    int
	min      int
	max      int
	onChange func(int)
	focused  bool
	tr       i18n.Translator
}

func NewMoodSelector(initial int, tr i18n.Translator) *MoodSelector {
	return &MoodSelector{
		value: initial,
		tr:    tr,
		min:   0,
		max:   10,
	}
}

func (m *MoodSelector) Value() int {
	return m.value
}

func (m *MoodSelector) SetValue(v int) {
	if v >= m.min && v <= m.max {
		m.value = v
		if m.onChange != nil {
			m.onChange(v)
		}
	}
}

func (m *MoodSelector) OnChange(fn func(int)) {
	m.onChange = fn
}

func (m *MoodSelector) Focus() {
	m.focused = true
}

func (m *MoodSelector) Blur() {
	m.focused = false
}

func (m *MoodSelector) Update(msg tea.Msg) tea.Cmd {
	if !m.focused {
		return nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left", "h":
			if m.value > m.min {
				m.SetValue(m.value - 1)
			}
		case "right", "l":
			if m.value < m.max {
				m.SetValue(m.value + 1)
			}
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":

			if len(keyMsg.Runes) > 0 {
				digit := int(keyMsg.Runes[0] - '0')
				if digit >= m.min && digit <= m.max {
					m.SetValue(digit)
				}
			}
		}
	}
	return nil
}

func (m *MoodSelector) View() string {
	level := entity.MoodLevel(m.value)

	var scale string
	for i := m.min; i <= m.max; i++ {
		style := lipgloss.NewStyle().
			Width(5).
			Align(lipgloss.Center).
			Padding(0, 1)

		if i == m.value {

			style = style.
				Background(styles.MoodColors[i]).
				Foreground(styles.TextLight).
				Bold(true)
		} else {

			style = style.
				Foreground(styles.MoodColors[i])
		}

		scale += style.Render(fmt.Sprintf("%d", i))
	}

	emoji := level.Emoji()
	description := m.tr.T(level.StringKey())

	currentStyle := lipgloss.NewStyle().
		Foreground(styles.MoodColors[m.value]).
		Bold(true).
		Align(lipgloss.Center).
		Width(55)

	current := currentStyle.Render(
		fmt.Sprintf("%s  %s  (%d/10)", emoji, description, m.value),
	)

	hint := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Align(lipgloss.Center).
		Width(55).
		Render("← / → для изменения, 0-9 для прямого ввода")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		scale,
		"",
		current,
		"",
		hint,
	)
}
