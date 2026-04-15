package screens

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/constants"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type MenuScreen struct {
	state.BaseState

	translator i18n.Translator
	choices    []menuChoice
	cursor     int
}

type menuChoice struct {
	label  string
	screen state.ScreenType
	params any
	key    string
	icon   string
}

func NewMenuScreen(translator i18n.Translator) *MenuScreen {
	s := &MenuScreen{
		translator: translator,
	}

	s.choices = []menuChoice{
		{
			label:  s.t(i18n.MenuRecordKey),
			screen: state.ScreenMoodForm,
			params: state.MoodFormParams{Date: time.Now(), Entry: nil},
			key:    "r",
			icon:   constants.EditIcon,
		},
		{
			label:  s.t(i18n.MenuCalendarKey),
			screen: state.ScreenCalendar,
			params: state.CalendarParams{InitialDate: time.Now()},
			key:    "c",
			icon:   constants.CalendarIcon,
		},
		{
			label:  s.t(i18n.MenuHistoryKey),
			screen: state.ScreenHistory,
			params: nil,
			key:    "h",
			icon:   constants.HistoryIcon,
		},
		{
			label:  s.t(i18n.MenuStatsKey),
			screen: state.ScreenStats,
			params: state.StatsParams{Period: "month"},
			key:    "s",
			icon:   constants.StatsIcon,
		},
		{
			label:  s.t(i18n.MenuSettingsKey),
			screen: state.ScreenSettings,
			params: nil,
			key:    "o",
			icon:   constants.SettingsIcon,
		},
	}

	return s
}

func (s *MenuScreen) t(key string, args ...interface{}) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *MenuScreen) Init() tea.Cmd {
	return nil
}

func (s *MenuScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)

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
			return s, state.Navigate(choice.screen, choice.params)

		case "r":
			return s, state.NavigateToMoodForm(time.Now(), nil)
		case "c":
			return s, state.NavigateToCalendar(time.Now())
		case "h":
			return s, state.NavigateToHistory()
		case "s":
			return s, state.NavigateToStats("month")
		case "o":
			return s, state.NavigateToSettings()
		}
	}

	return s, nil
}

func (s *MenuScreen) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render(s.renderHeader())
	b.WriteString(header)
	b.WriteString("\n\n")

	for i, choice := range s.choices {
		cursor := "  "
		choiceText := fmt.Sprintf("%s %s", choice.icon, choice.label)
		if choice.key != "" {
			choiceText = choiceText + " (" + choice.key + ")"
		}

		if i == s.cursor {
			cursor = fmt.Sprintf("%s ", constants.ArrowRight)
			choiceText = styles.SelectedListItemStyle.Render(choiceText)
		} else {
			choiceText = styles.ListItemStyle.Render(choiceText)
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, choiceText))
	}

	b.WriteString("\n")
	help := styles.HelpStyle.Render(s.t(i18n.HelpNavigationMenuKey))
	b.WriteString(help)

	return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
}

func (s *MenuScreen) renderHeader() string {
	title := `
 ███╗   ███╗ ██████╗  ██████╗ ██████╗     ██████╗ ██╗ █████╗ ██████╗ ██╗   ██╗
 ████╗ ████║██╔═══██╗██╔═══██╗██╔══██╗    ██╔══██╗██║██╔══██╗██╔══██╗╚██╗ ██╔╝
 ██╔████╔██║██║   ██║██║   ██║██║  ██║    ██║  ██║██║███████║██████╔╝ ╚████╔╝ 
 ██║╚██╔╝██║██║   ██║██║   ██║██║  ██║    ██║  ██║██║██╔══██║██╔══██╗  ╚██╔╝  
 ██║ ╚═╝ ██║╚██████╔╝╚██████╔╝██████╔╝    ██████╔╝██║██║  ██║██║  ██║   ██║   
 ╚═╝     ╚═╝ ╚═════╝  ╚═════╝ ╚═════╝     ╚═════╝ ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝   
`

	subtitle := s.t(i18n.MenuTitleKey)

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.PastelLavender).
		Bold(true)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Italic(true)

	return titleStyle.Render(title) + "\n" + subtitleStyle.Render(subtitle)
}
