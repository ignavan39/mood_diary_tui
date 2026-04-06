package tui

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
)

type LanguageOption struct {
	Code  string
	Label string
}

type SettingsModel struct {
	translator   i18n.Translator
	settingsRepo repository.SettingsRepository
	languages    []LanguageOption
	cursor       int
	step         int
	tempCursor   int
	errorMsg     string
}

func NewSettingsModel(
	translator i18n.Translator,
	settingsRepo repository.SettingsRepository,
) *SettingsModel {
	return &SettingsModel{
		translator:   translator,
		settingsRepo: settingsRepo,
		languages: []LanguageOption{
			{Code: "en", Label: "settings.language.en"},
			{Code: "ru", Label: "settings.language.ru"},
		},
		step:       0,
		tempCursor: 0,
	}
}

func (m *SettingsModel) t(key string, args ...any) string {
	if m.translator == nil {
		return key
	}
	return m.translator.T(key, args...)
}

func (m *SettingsModel) Init() tea.Cmd {
	return m.loadCurrentLanguage()
}

func (m *SettingsModel) loadCurrentLanguage() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		if lang, err := m.settingsRepo.Get(ctx, entity.SettingsKeyLanguage); err == nil {
			locale := i18n.NormalizeLocale(lang.Value)
			if locale == i18n.LocaleRU {
				m.cursor = 1
			} else {
				m.cursor = 0
			}
		}
		m.tempCursor = m.cursor
		return SettingsLoadedMsg{}
	}
}

func (m *SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.step {
		case 0:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
					m.tempCursor = m.cursor
				}
			case "down", "j":
				if m.cursor < len(m.languages)-1 {
					m.cursor++
					m.tempCursor = m.cursor
				}
			case "enter":

				m.step = 1
				return m, nil
			case "esc", "q":
				return m, Navigate(ScreenMenu)
			}

		case 1:
			switch msg.String() {
			case "y", "Y", "enter":

				m.step = 2
				return m, m.saveLanguage()
			case "n", "N", "esc":

				m.step = 0
				return m, nil
			}

		case 2:

		}

	case SettingsSavedMsg:

		return m, tea.Batch(
			tea.Tick(1200*time.Millisecond, func(t time.Time) tea.Msg {
				return NavigateMsg{Screen: ScreenMenu}
			}),
		)

	case ErrorMsg:
		m.errorMsg = msg.Error.Error()
		m.step = 0
	}

	return m, nil
}

func (m *SettingsModel) saveLanguage() tea.Cmd {
	return func() tea.Msg {
		selectedLang := m.languages[m.tempCursor].Code
		ctx := context.Background()

		err := m.settingsRepo.Upsert(ctx, &entity.UserSettings{
			Key:   entity.SettingsKeyLanguage,
			Value: selectedLang,
		})
		if err != nil {
			return ErrorMsg{Error: err}
		}

		locale := i18n.Locale(selectedLang)
		_ = m.translator.SetLocale(locale)

		m.cursor = m.tempCursor

		return SettingsSavedMsg{}
	}
}

func (m *SettingsModel) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render(m.t("settings.title"))
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.errorMsg != "" {
		b.WriteString(styles.ErrorStyle.Render(m.t("common.error_prefix") + m.errorMsg))
		b.WriteString("\n\n")
	}

	switch m.step {
	case 0:
		b.WriteString(m.renderLanguageSelection())
	case 1:
		b.WriteString(m.renderConfirmation())
	case 2:
		b.WriteString(m.renderSaving())
	}

	b.WriteString("\n")

	help := styles.HelpStyle.Render(m.t("help.navigation.settings"))
	b.WriteString(help)

	return lipgloss.NewStyle().
		Padding(2, 4).
		Render(b.String())
}

func (m *SettingsModel) renderLanguageSelection() string {
	var b strings.Builder

	b.WriteString(styles.SubtitleStyle.Render(m.t("settings.option_language")))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Render(m.t("settings.option_language_desc")))
	b.WriteString("\n\n")

	for i, lang := range m.languages {
		label := m.t(lang.Label)

		if i == m.cursor {

			selectedStyle := lipgloss.NewStyle().
				Foreground(styles.PastelLavender).
				Background(lipgloss.Color("#4A4A6A")).
				Bold(true).
				Padding(0, 2)

			b.WriteString("→ ")
			b.WriteString(selectedStyle.Render(fmt.Sprintf("● %s", label)))
			b.WriteString(" ←")
		} else {

			unselectedStyle := lipgloss.NewStyle().
				Foreground(styles.TextMuted)

			b.WriteString("  ")
			b.WriteString(unselectedStyle.Render(fmt.Sprintf("○ %s", label)))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	return b.String()
}

func (m *SettingsModel) renderConfirmation() string {
	var b strings.Builder

	selectedLang := m.t(m.languages[m.tempCursor].Label)

	b.WriteString(styles.SubtitleStyle.Render(m.t("settings.confirm_change")))
	b.WriteString("\n\n")

	boxContent := fmt.Sprintf(
		m.t("settings.confirm_text"),
		selectedLang,
	)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PastelLavender).
		Padding(1, 2).
		Render(boxContent)

	b.WriteString(box)
	b.WriteString("\n\n")

	return b.String()
}

func (m *SettingsModel) renderSaving() string {
	var b strings.Builder

	b.WriteString(styles.InfoStyle.Render(m.t("settings.success_edit")))
	b.WriteString("\n\n")

	return b.String()
}

type SettingsLoadedMsg struct{}
type SettingsSavedMsg struct{}
