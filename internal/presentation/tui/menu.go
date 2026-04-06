package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

type MenuModel struct {
	choices    []string
	cursor     int
	selected   int
	translator i18n.Translator
}

func NewMenuModel(translator i18n.Translator) *MenuModel {
	return &MenuModel{

		choices:    nil,
		cursor:     0,
		selected:   -1,
		translator: translator,
	}
}

func (m *MenuModel) getChoices() []string {
	return []string{
		m.translator.T("menu.item_record"),
		m.translator.T("menu.item_stats"),
		m.translator.T("menu.item_history"),
		m.translator.T("menu.item_settings"),
		m.translator.T("menu.item_exit"),
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

			choices := m.getChoices()
			if m.cursor < len(choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = m.cursor
			return m, m.handleSelection()
		case "q", "ctrl+c":
			return m, tea.Quit
		}
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
		return Navigate(ScreenSettings)
	case 4:
		return tea.Quit
	}
	return nil
}

func (m *MenuModel) View() string {
	var b strings.Builder

	choices := m.getChoices()

	header := styles.HeaderStyle.Render(m.renderHeader())
	b.WriteString(header)
	b.WriteString("\n\n")

	for i, choice := range choices {
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

	help := styles.HelpStyle.Render(m.translator.T("help.navigation.menu"))
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

	subtitle := m.translator.T("menu.subtitle")

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.PastelLavender).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Italic(true)

	return titleStyle.Render(title) + "\n" + subtitleStyle.Render(subtitle)
}
