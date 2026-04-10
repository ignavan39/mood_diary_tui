package screens

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/domain/repository"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type SettingsScreen struct {
	state.BaseState

	translator   i18n.Translator
	settingsRepo repository.SettingsRepository

	cursor        int
	locales       []string
	currentLocale string
}

func NewSettingsScreen(translator i18n.Translator, settingsRepo repository.SettingsRepository) *SettingsScreen {
	return &SettingsScreen{
		translator:   translator,
		settingsRepo: settingsRepo,
		locales:      []string{"en", "ru"},
	}
}

func (s *SettingsScreen) t(key string, args ...interface{}) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *SettingsScreen) Init() tea.Cmd {
	s.SetLoading(true)
	return s.loadSettings()
}

func (s *SettingsScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		return s.handleKeyMsg(msg)

	case settingsLoadedMsg:
		s.currentLocale = msg.locale
		s.SetLoading(false)
		return s, nil

	case settingsSavedMsg:

		s.SetLoading(false)
		return s, nil

	case state.ErrorMsg:
		s.SetError(msg.Error)
		return s, nil
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
		if s.cursor < len(s.locales)-1 {
			s.cursor++
		}

	case "enter", "Enter", " ":

		selectedLocale := s.locales[s.cursor]
		if selectedLocale != s.currentLocale {
			s.SetLoading(true)
			return s, s.saveLocale(selectedLocale)
		}

	case "esc", "q":
		return s, state.NavigateToMenu()
	}

	return s, nil
}

type settingsLoadedMsg struct {
	locale string
}

type settingsSavedMsg struct{}

func (s *SettingsScreen) loadSettings() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		settings, err := s.settingsRepo.Get(ctx, entity.SettingsKeyLanguage)
		if err != nil {
			return state.ErrorMsg{Error: err}
		}

		return settingsLoadedMsg{
			locale: settings.Value,
		}
	}
}

func (s *SettingsScreen) saveLocale(locale string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		settings, err := s.settingsRepo.Get(ctx, entity.SettingsKeyLanguage)
		if err != nil {
			return state.ErrorMsg{Error: err}
		}

		settings.Value = locale
		err = s.settingsRepo.Upsert(ctx, settings)
		if err != nil {
			return state.ErrorMsg{Error: err}
		}

		return settingsSavedMsg{}
	}
}

func (s *SettingsScreen) View() string {

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.PastelLavender).
		Align(lipgloss.Center).
		Width(s.Width)

	header := headerStyle.Render("⚙️  " + s.t("settings.title"))

	if s.Loading {
		loadingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB3BA")).
			Align(lipgloss.Center).
			Width(s.Width).
			Padding(2, 0)

		return lipgloss.JoinVertical(
			lipgloss.Center,
			header,
			"",
			loadingStyle.Render(s.t("common.loading")),
		)
	}

	languageLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#9B9B9B")).
		Render(s.t("settings.language") + ":")

	var localeItems string
	for i, locale := range s.locales {
		itemStyle := styles.ListItemStyle.Copy()

		if i == s.cursor {
			itemStyle = styles.SelectedListItemStyle.Copy()
		}

		localeLabel := s.getLocaleLabel(locale)
		if locale == s.currentLocale {
			localeLabel = "✓ " + localeLabel
		}

		localeItems += itemStyle.Render(localeLabel) + "\n"
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		Align(lipgloss.Center).
		Width(s.Width).
		Padding(1, 0)

	help := helpStyle.Render(s.t("help.navigation.settings"))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		languageLabel,
		"",
		localeItems,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Padding(2, 4).
		Width(s.Width).
		Render(content)
}

func (s *SettingsScreen) getLocaleLabel(locale string) string {
	labels := map[string]string{
		"en": "English",
		"ru": "Русский",
	}

	if label, ok := labels[locale]; ok {
		return label
	}
	return locale
}
