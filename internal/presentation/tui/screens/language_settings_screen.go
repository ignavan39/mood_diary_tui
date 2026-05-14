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
	"github.com/ignavan39/mood-diary/internal/presentation/tui/constants"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type LanguageSettingsScreen struct {
	state.BaseState

	ctx          context.Context
	translator   i18n.Translator
	settingsRepo repository.SettingsRepository

	cursor         int
	locales        []i18n.Locale
	currentLocale  i18n.Locale
	saved          bool
}

func NewLanguageSettingsScreen(
	ctx context.Context,
	translator i18n.Translator,
	settingsRepo repository.SettingsRepository,
) *LanguageSettingsScreen {
	s := &LanguageSettingsScreen{
		ctx:          ctx,
		translator:   translator,
		settingsRepo: settingsRepo,
		locales:      i18n.SupportedLocales(),
		currentLocale: translator.Locale(),
	}
	return s
}

func (s *LanguageSettingsScreen) t(key string, args ...any) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *LanguageSettingsScreen) Init() tea.Cmd {
	return nil
}

func (s *LanguageSettingsScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		return s.handleKeyMsg(msg)

	case savedMsg:
		s.saved = true
		return s, tea.Tick(1200*time.Millisecond, func(t time.Time) tea.Msg {
			return state.NavigateMsg{To: state.ScreenSettings}
		})

	case state.ErrorMsg:
		s.SetError(msg.Error)
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
		selected := s.locales[s.cursor]
		if selected != s.currentLocale {
			return s, s.saveLocale(selected)
		}

	case "esc", "q":
		return s, state.Navigate(state.ScreenSettings, nil)
	}
	return s, nil
}

type savedMsg struct{}

func (s *LanguageSettingsScreen) saveLocale(locale i18n.Locale) tea.Cmd {
	return func() tea.Msg {
		settings := &entity.UserSettings{
			Key:   entity.SettingsKeyLanguage,
			Value: string(locale),
		}
		if err := s.settingsRepo.Upsert(s.ctx, settings); err != nil {
			return state.ErrorMsg{Error: err}
		}
		_ = s.translator.SetLocale(locale)
		s.currentLocale = locale
		return savedMsg{}
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
		label := s.localeName(locale)
		isCurrent := locale == s.currentLocale

		if i == s.cursor {
			b.WriteString(constants.ArrowRight + " ")
			if isCurrent {
				b.WriteString(styles.SelectedListItemStyle.Render(
					fmt.Sprintf("%s %s %s", constants.FilledDot, label, constants.Checkmark)))
			} else {
				b.WriteString(styles.SelectedListItemStyle.Render(
					fmt.Sprintf("%s %s", constants.FilledDot, label)))
			}
		} else {
			b.WriteString("  ")
			if isCurrent {
				b.WriteString(styles.ListItemStyle.Render(
					fmt.Sprintf("%s %s %s", constants.EmptyDot, label, constants.Checkmark)))
			} else {
				b.WriteString(styles.ListItemStyle.Render(
					fmt.Sprintf("%s %s", constants.EmptyDot, label)))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	return b.String()
}

func (s *LanguageSettingsScreen) localeName(locale i18n.Locale) string {
	key := fmt.Sprintf("settings.language.%s", locale)
	name := s.t(key)
	if name == key {
		return string(locale)
	}
	return fmt.Sprintf("[%s] %s", strings.ToUpper(string(locale)), name)
}