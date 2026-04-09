package components

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
)

type MoodSelector struct {
	value    int
	min      int
	max      int
	onChange func(int)
	focused  bool
}

func NewMoodSelector(initial int) *MoodSelector {
	return &MoodSelector{
		value: initial,
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
			// Прямой ввод числа
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

	// Цветовая палитра от красного к зеленому
	colors := []lipgloss.Color{
		lipgloss.Color("#FF6B6B"), // 0 - очень плохо
		lipgloss.Color("#FF8E8E"),
		lipgloss.Color("#FFB3B3"),
		lipgloss.Color("#FFD4A3"),
		lipgloss.Color("#FFEB9C"),
		lipgloss.Color("#FFFFBA"), // 5 - нейтрально
		lipgloss.Color("#E8FFC4"),
		lipgloss.Color("#C4FFCF"),
		lipgloss.Color("#9BFFAB"),
		lipgloss.Color("#6BFFB8"),
		lipgloss.Color("#51CF66"), // 10 - отлично
	}

	// Визуальная шкала
	var scale string
	for i := m.min; i <= m.max; i++ {
		style := lipgloss.NewStyle().
			Width(5).
			Align(lipgloss.Center).
			Padding(0, 1)

		if i == m.value {
			// Активный элемент
			style = style.
				Background(colors[i]).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)
		} else {
			// Неактивный элемент
			style = style.
				Foreground(colors[i])
		}

		scale += style.Render(fmt.Sprintf("%d", i))
	}

	// Эмоджи и описание
	emoji := level.Emoji()
	description := level.String()

	currentStyle := lipgloss.NewStyle().
		Foreground(colors[m.value]).
		Bold(true).
		Align(lipgloss.Center).
		Width(55)

	current := currentStyle.Render(
		fmt.Sprintf("%s  %s  (%d/10)", emoji, description, m.value),
	)

	// Хинт
	hint := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
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
