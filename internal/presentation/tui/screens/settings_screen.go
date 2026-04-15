package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/constants"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type settingsChoice struct {
	label  string
	screen state.ScreenType
}

type SettingsScreen struct {
	state.BaseState

	translator i18n.Translator
	choices    []settingsChoice
	cursor     int
}

func NewSettingsScreen(translator i18n.Translator) *SettingsScreen {
	s := &SettingsScreen{
		translator: translator,
	}
	s.choices = []settingsChoice{
		{
			label:  s.t(i18n.SettingsOptionLanguageKey),
			screen: state.ScreenLanguageSettings,
		},
	}
	return s
}

func (s *SettingsScreen) t(key string, args ...interface{}) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *SettingsScreen) Init() tea.Cmd {
	return nil
}

func (s *SettingsScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		return s.handleKeyMsg(msg)
	}

	return s, nil
}

func (s *SettingsScreen) handleKeyMsg(msg tea.KeyMsg) (state.Screen, tea.Cmd) {
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
		return s, state.Navigate(s.choices[s.cursor].screen, nil)

	case "esc", "q":
		return s, state.Navigate(state.ScreenMenu, nil)
	}

	return s, nil
}

func (s *SettingsScreen) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render(s.t(i18n.SettingsTitleKey))
	b.WriteString(header)
	b.WriteString("\n\n")

	for i, choice := range s.choices {
		if i == s.cursor {
			b.WriteString(fmt.Sprintf("%s ", constants.ArrowRight))
			b.WriteString(styles.SelectedListItemStyle.Render(choice.label))
		} else {
			b.WriteString("  ")
			b.WriteString(styles.ListItemStyle.Render(choice.label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	help := styles.HelpStyle.Render(s.t(i18n.HelpNavigationSettingsKey))
	b.WriteString(help)

	return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
}
