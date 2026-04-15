package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/domain/repository"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/components"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/constants"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type LanguageSettingsScreen struct {
	state.BaseState

	translator   i18n.Translator
	settingsRepo repository.SettingsRepository

	cursor        int
	locales       []string
	currentLocale string
	saved         bool
}

func NewLanguageSettingsScreen(translator i18n.Translator, settingsRepo repository.SettingsRepository) *LanguageSettingsScreen {
	return &LanguageSettingsScreen{
		translator:   translator,
		settingsRepo: settingsRepo,
		locales:      []string{"en", "ru", "ja"},
	}
}

func (s *LanguageSettingsScreen) t(key string, args ...interface{}) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *LanguageSettingsScreen) Init() tea.Cmd {
	s.SetLoading(true)
	return s.loadSettings()
}

func (s *LanguageSettingsScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		return s.handleKeyMsg(msg)

	case languageSettingsLoadedMsg:
		s.currentLocale = msg.locale
		s.SetLoading(false)
		return s, nil

	case languageSettingsSavedMsg:

		s.saved = true
		s.SetLoading(false)

		return s, tea.Tick(1200*time.Millisecond, func(t time.Time) tea.Msg {
			return state.NavigateMsg{To: state.ScreenSettings}
		})

	case state.ErrorMsg:
		s.SetError(msg.Error)
		s.SetLoading(false)
		return s, nil
	}

	return s, nil
}

func (s *LanguageSettingsScreen) handleKeyMsg(msg tea.KeyMsg) (state.Screen, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}

	case "down", "j":
		if s.cursor < len(s.locales)-1 {
			s.cursor++
		}

	case "enter", " ":

		selectedLocale := s.locales[s.cursor]
		if selectedLocale != s.currentLocale {
			s.SetLoading(true)
			return s, s.saveLocale(selectedLocale)
		}

	case "esc", "q":
		return s, state.Navigate(state.ScreenSettings, nil)
	}

	return s, nil
}

type languageSettingsLoadedMsg struct {
	locale string
}

type languageSettingsSavedMsg struct{}

func (s *LanguageSettingsScreen) loadSettings() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		settings, err := s.settingsRepo.Get(ctx, entity.SettingsKeyLanguage)
		if err != nil {
			return state.ErrorMsg{Error: err}
		}

		return languageSettingsLoadedMsg{
			locale: settings.Value,
		}
	}
}

func (s *LanguageSettingsScreen) saveLocale(locale string) tea.Cmd {
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

		newLocale := i18n.Locale(locale)
		_ = s.translator.SetLocale(newLocale)

		s.currentLocale = locale

		return languageSettingsSavedMsg{}
	}
}

func (s *LanguageSettingsScreen) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render(s.t(i18n.SettingsOptionLanguageKey))
	b.WriteString(header)
	b.WriteString("\n\n")

	if s.Error != nil {
		b.WriteString(styles.ErrorStyle.Render(s.t(i18n.CommonErrorPrefixKey) + s.Error.Error()))
		b.WriteString("\n\n")
	}

	if s.saved {
		b.WriteString(styles.SuccessStyle.Render(s.t(i18n.SettingsSuccessEditKey)))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render(s.t(i18n.CommonReturningKey)))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if s.Loading {
		loading := components.NewLoading(s.t(i18n.CommonLoaderMessageKey))
		b.WriteString(styles.InfoStyle.Render(s.t(i18n.CommonLoaderMessageKey)))
		b.WriteString("\n")
		b.WriteString(loading.View())
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	b.WriteString(s.renderLanguageSelection())
	b.WriteString("\n")

	help := styles.HelpStyle.Render(s.t(i18n.HelpNavigationSettingsKey))
	b.WriteString(help)

	return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
}

func (s *LanguageSettingsScreen) renderLanguageSelection() string {
	var b strings.Builder

	b.WriteString(styles.SubtitleStyle.Render(s.t(i18n.SettingsOptionLanguageKey)))
	b.WriteString("\n\n")

	for i, locale := range s.locales {
		label := s.getLocaleLabel(locale)

		if i == s.cursor {
			b.WriteString(constants.ArrowRight + " ")
			b.WriteString(styles.SelectedListItemStyle.Render(fmt.Sprintf("%s %s", constants.FilledDot, label)))
			if locale == s.currentLocale {
				b.WriteString(constants.Checkmark)
			}
		} else {
			b.WriteString("  ")
			style := styles.ListItemStyle
			if locale == s.currentLocale {
				label = label + "  " + constants.Checkmark
			}
			b.WriteString(style.Render(fmt.Sprintf("%s %s", constants.EmptyDot, label)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	return b.String()
}

func (s *LanguageSettingsScreen) getLocaleLabel(locale string) string {
	labels := map[string]string{
		"en": "[En] - English",
		"ru": "[Ru] - Русский",
		"ja": "[Ja] - 日本語",
	}

	if label, ok := labels[locale]; ok {
		return label
	}
	return locale
}
