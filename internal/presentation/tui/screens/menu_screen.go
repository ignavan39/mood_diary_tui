package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

const moodDiaryBanner = `
 ███╗   ███╗ ██████╗  ██████╗ ██████╗     ██████╗ ██╗ █████╗ ██████╗ ██╗   ██╗
 ████╗ ████║██╔═══██╗██╔═══██╗██╔══██╗    ██╔══██╗██║██╔══██╗██╔══██╗╚██╗ ██╔╝
 ██╔████╔██║██║   ██║██║   ██║██║  ██║    ██║  ██║██║███████║██████╔╝ ╚████╔╝ 
 ██║╚██╔╝██║██║   ██║██║   ██║██║  ██║    ██║  ██║██║██╔══██║██╔══██╗  ╚██╔╝  
 ██║ ╚═╝ ██║╚██████╔╝╚██████╔╝██████╔╝    ██████╔╝██║██║  ██║██║  ██║   ██║   
 ╚═╝     ╚═╝ ╚═════╝  ╚═════╝ ╚═════╝     ╚═════╝ ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   
`

type menuChoice struct {
	label  string
	screen state.ScreenType
	params any
	key    string
	isExit bool
}

type MenuScreen struct {
	state.BaseState
	translator i18n.Translator
	choices    []menuChoice
	cursor     int
}

func NewMenuScreen(tr i18n.Translator) state.Screen {
	s := &MenuScreen{
		translator: tr,
		cursor:     0,
	}
	s.updateChoices()
	return s
}

func (s *MenuScreen) updateChoices() {
	s.choices = []menuChoice{
		{label: s.t("menu.item_record"), screen: state.ScreenMoodForm, params: nil, key: "r"},
		{label: s.t("menu.item_calendar"), screen: state.ScreenCalendar, params: nil, key: "c"},
		{label: s.t("menu.item_history"), screen: state.ScreenHistory, params: nil, key: "h"},
		{label: s.t("menu.item_stats"), screen: state.ScreenStats, params: nil, key: "s"},
		{label: s.t("menu.item_settings"), screen: state.ScreenSettings, params: nil, key: "o"},
		{label: s.t("menu.item_exit"), screen: state.ScreenMenu, params: nil, key: "q", isExit: true},
	}
}

func (s *MenuScreen) t(key string, args ...any) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *MenuScreen) Init() tea.Cmd {
	s.updateChoices()
	return nil
}

func (s *MenuScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.choices)-1 {
				s.cursor++
			}
		case "enter", " ":
			choice := s.choices[s.cursor]
			if choice.isExit {
				return s, tea.Quit
			}
			return s, state.Navigate(choice.screen, choice.params)
		case "r", "c", "h", "s", "o":
			for _, ch := range s.choices {
				if ch.key == msg.String() && !ch.isExit {
					return s, state.Navigate(ch.screen, ch.params)
				}
			}
		case "q", "esc":
			choice := s.choices[s.cursor]
			if choice.isExit {
				return s, tea.Quit
			}
		}
	}
	return s, nil
}

func (s *MenuScreen) View() string {
	w, h := s.Width, s.Height
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 40
	}

	var b strings.Builder

	trimmedBanner := strings.TrimSpace(moodDiaryBanner)

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.PastelLavender).
		Bold(true).
		Align(lipgloss.Left).
		Width(w)

	b.WriteString(titleStyle.Render(trimmedBanner))
	b.WriteString("\n")

	subtitle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Align(lipgloss.Left).
		Width(w).
		Render("✦ " + s.t("menu.subtitle") + " ✦")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	for i, choice := range s.choices {
		var item strings.Builder

		baseStyle := lipgloss.NewStyle().
			Padding(0, 2).
			Width(w).
			Align(lipgloss.Left)

		if i == s.cursor {

			baseStyle = baseStyle.
				Background(styles.PastelLavender).
				Foreground(styles.TextLight).
				Bold(true)
		} else {
			baseStyle = baseStyle.Foreground(styles.TextDark)
		}

		label := choice.label
		if choice.key != "" {
			if i == s.cursor {

				label += fmt.Sprintf("  [%s]", choice.key)
			} else {

				label += lipgloss.NewStyle().
					Foreground(styles.TextMuted).
					Render(fmt.Sprintf("  [%s]", choice.key))
			}
		}

		item.WriteString(baseStyle.Render(label))
		b.WriteString(item.String())
		b.WriteString("\n")
	}

	b.WriteString("\n")

	help := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Align(lipgloss.Left).
		Width(w).
		Render(s.t("help.navigation.menu"))
	b.WriteString(help)

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Padding(1, 0).
		Render(b.String())
}
