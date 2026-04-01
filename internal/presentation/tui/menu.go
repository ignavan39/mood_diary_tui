package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

type MenuModel struct {
	choices  []string
	cursor   int
	selected int
}

func NewMenuModel() *MenuModel {
	return &MenuModel{
		choices: []string{
			"📝 Записать настроение",
			"📊 Посмотреть статистику",
			"📅 История записей",
			"❌ Выход",
		},
		cursor:   0,
		selected: -1,
	}
}

func (m *MenuModel) Init() tea.Cmd {
	return nil
}

func (m *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
			return m, m.handleSelection()
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	default:

	}
	return m, nil
}

func (m *MenuModel) handleSelection() tea.Cmd {
	switch m.selected {
	case 0:
		return Navigate(ScreenRecord)
	case 1:
		return Navigate(ScreenStats)
	case 2:
		return Navigate(ScreenHistory)
	case 3:
		return tea.Quit
	}
	return nil
}

func (m *MenuModel) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render(m.renderHeader())
	b.WriteString(header)
	b.WriteString("\n\n")

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "→ "
			choice = styles.SelectedListItemStyle.Render(choice)
		} else {
			choice = styles.ListItemStyle.Render(choice)
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, choice))
	}

	b.WriteString("\n")
	help := styles.HelpStyle.Render("↑/↓: Навигация • Enter: Выбрать • q: Выход")
	b.WriteString(help)

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

func (m *MenuModel) renderHeader() string {
	title := `
 ███╗   ███╗ ██████╗  ██████╗ ██████╗     ██████╗ ██╗ █████╗ ██████╗ ██╗   ██╗
 ████╗ ████║██╔═══██╗██╔═══██╗██╔══██╗    ██╔══██╗██║██╔══██╗██╔══██╗╚██╗ ██╔╝
 ██╔████╔██║██║   ██║██║   ██║██║  ██║    ██║  ██║██║███████║██████╔╝ ╚████╔╝ 
 ██║╚██╔╝██║██║   ██║██║   ██║██║  ██║    ██║  ██║██║██╔══██║██╔══██╗  ╚██╔╝  
 ██║ ╚═╝ ██║╚██████╔╝╚██████╔╝██████╔╝    ██████╔╝██║██║  ██║██║  ██║   ██║   
 ╚═╝     ╚═╝ ╚═════╝  ╚═════╝ ╚═════╝     ╚═════╝ ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   
`

	subtitle := "Дневник настроения • Отслеживайте свои эмоции 🌸"

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.PastelLavender).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Italic(true)

	return titleStyle.Render(title) + "\n" + subtitleStyle.Render(subtitle)
}
