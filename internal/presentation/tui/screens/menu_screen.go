package screens

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
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
	params interface{}
	key    string
}

func NewMenuScreen(translator i18n.Translator) *MenuScreen {
	s := &MenuScreen{
		translator: translator,
	}

	s.choices = []menuChoice{
		{
			label:  s.t("menu.record"),
			screen: state.ScreenMoodForm,
			params: state.MoodFormParams{Date: time.Now(), Entry: nil},
			key:    "r",
		},
		{
			label:  s.t("menu.calendar"),
			screen: state.ScreenCalendar,
			params: state.CalendarParams{InitialDate: time.Now()},
			key:    "c",
		},
		{
			label:  s.t("menu.history"),
			screen: state.ScreenHistory,
			params: nil,
			key:    "h",
		},
		{
			label:  s.t("menu.stats"),
			screen: state.ScreenStats,
			params: state.StatsParams{Period: "month"},
			key:    "s",
		},
		{
			label:  s.t("menu.settings"),
			screen: state.ScreenSettings,
			params: nil,
			key:    "o",
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

		// Горячие клавиши
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
	// Заголовок
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.PastelLavender).
		Align(lipgloss.Center).
		Width(s.Width)

	title := titleStyle.Render("📔 " + s.t("menu.title"))

	// Пункты меню
	var menuItems string
	for i, choice := range s.choices {
		itemStyle := styles.ListItemStyle.Copy()

		if i == s.cursor {
			itemStyle = styles.SelectedListItemStyle.Copy()
		}

		label := choice.label
		if choice.key != "" {
			label = label + " (" + choice.key + ")"
		}

		menuItems += itemStyle.Render(label) + "\n"
	}

	// Подсказка
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		Align(lipgloss.Center).
		Width(s.Width).
		Padding(1, 0)

	help := helpStyle.Render(s.t("help.navigation.menu"))

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		menuItems,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Width(s.Width).
		Height(s.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}
