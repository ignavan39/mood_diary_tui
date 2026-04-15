package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/application/usecase"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/forms"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type MoodFormScreen struct {
	state.BaseState

	service    *usecase.MoodService
	translator i18n.Translator

	date  time.Time
	entry *entity.MoodEntry

	wizard *forms.Wizard

	moodLevel int
	note      string

	saved  bool
	saving bool
}

func NewMoodFormScreen(
	service *usecase.MoodService,
	translator i18n.Translator,
	date time.Time,
	entry *entity.MoodEntry,
) *MoodFormScreen {
	screen := &MoodFormScreen{
		service:    service,
		translator: translator,
		date:       date,
		entry:      entry,
	}

	steps := []forms.Step{
		NewMoodLevelStep(screen),
		NewMoodNoteStep(screen),
		NewMoodConfirmationStep(screen),
	}

	screen.wizard = forms.NewWizard(steps, translator)

	return screen
}

func (s *MoodFormScreen) t(key string, args ...interface{}) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *MoodFormScreen) Init() tea.Cmd {
	return nil
}

func (s *MoodFormScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)
		s.wizard.SetSize(msg.Width, msg.Height)

	case state.MoodSavedMsg:
		s.saved = true
		s.saving = false

		return s, tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
			return state.NavigateMsg{To: state.ScreenMenu}
		})

	case state.ErrorMsg:
		s.SetError(msg.Error)
		s.saving = false
		return s, nil
	}

	cmd := s.wizard.Update(msg)

	if s.wizard.IsComplete() && !s.saving {
		s.saving = true
		return s, s.save()
	}

	if s.wizard.IsCancelled() {
		return s, state.Navigate(state.ScreenMenu, nil)
	}

	return s, cmd
}

func (s *MoodFormScreen) View() string {
	var b strings.Builder

	headerKey := i18n.RecordTitleNewKey
	if s.entry != nil {
		headerKey = i18n.RecordTitleEditKey
	}
	header := styles.HeaderStyle.Render(s.t(headerKey))
	b.WriteString(header)
	b.WriteString("\n\n")

	if s.saved {
		successKey := i18n.RecordSuccessNewKey
		if s.entry != nil {
			successKey = i18n.RecordSuccessEditKey
		}
		success := styles.SuccessStyle.Render(s.t(successKey))
		b.WriteString(success)
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render(s.t(i18n.CommonReturningKey)))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if s.saving {
		b.WriteString(styles.InfoStyle.Render(s.t(i18n.RecordSavingKey)))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	if s.Error != nil {
		errMsg := styles.ErrorStyle.Render(s.t(i18n.CommonErrorPrefixKey) + s.Error.Error())
		b.WriteString(errMsg)
		b.WriteString("\n\n")
	}

	wizardContent := s.wizard.View()
	b.WriteString(wizardContent)

	return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
}

func (s *MoodFormScreen) save() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var err error
		if s.entry != nil {

			err = s.service.RecordMood(ctx, s.moodLevel, s.note, &s.date)
		} else {

			err = s.service.RecordMood(ctx, s.moodLevel, s.note, &s.date)
		}

		if err != nil {
			return state.ErrorMsg{Error: err}
		}

		return state.MoodSavedMsg{}
	}
}

type MoodLevelStep struct {
	screen *MoodFormScreen
}

func NewMoodLevelStep(screen *MoodFormScreen) *MoodLevelStep {
	initial := 5
	if screen.entry != nil {
		initial = screen.entry.Level.Int()
	}

	screen.moodLevel = initial

	return &MoodLevelStep{
		screen: screen,
	}
}

func (s *MoodLevelStep) Render(width, height int) string {
	var b strings.Builder

	b.WriteString(styles.SubtitleStyle.Render(s.screen.t(i18n.RecordPromptFeelingKey)))
	b.WriteString("\n\n")

	b.WriteString(s.renderMoodScale())
	b.WriteString("\n\n")

	moodLevel := entity.MoodLevel(s.screen.moodLevel)
	desc := s.screen.t(moodLevel.StringKey())
	current := fmt.Sprintf("%s %s (%d/10)", moodLevel.Emoji(), desc, s.screen.moodLevel)
	currentStyle := styles.MoodStyle(s.screen.moodLevel)
	b.WriteString(currentStyle.Render(current))
	b.WriteString("\n\n")

	help := styles.HelpStyle.Render(s.screen.t(i18n.HelpNavigationRecordStep0Key))
	b.WriteString(help)

	return b.String()
}

func (s *MoodLevelStep) renderMoodScale() string {
	var parts []string

	for i := 0; i <= 10; i++ {
		moodLevel, _ := entity.NewMoodLevel(i)
		emoji := moodLevel.Emoji()

		if i == s.screen.moodLevel {
			selectedStyle := lipgloss.NewStyle().
				Foreground(styles.TextLight).
				Background(styles.GetMoodColor(i)).
				Bold(true)
			selected := selectedStyle.Render("[ " + emoji + " ]")
			parts = append(parts, selected)
		} else {
			unselectedStyle := lipgloss.NewStyle().
				Foreground(styles.GetMoodColor(i)).
				Faint(true).
				Padding(0, 1).
				Render(emoji)
			parts = append(parts, unselectedStyle)
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}

func (s *MoodLevelStep) Update(msg tea.Msg) (forms.Step, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left", "h":
			if s.screen.moodLevel > 0 {
				s.screen.moodLevel--
			}
		case "right", "l":
			if s.screen.moodLevel < 10 {
				s.screen.moodLevel++
			}
		}
	}
	return s, nil
}

func (s *MoodLevelStep) Validate() error {
	return nil
}

func (s *MoodLevelStep) OnEnter() tea.Cmd {
	return nil
}

func (s *MoodLevelStep) OnExit() tea.Cmd {
	return nil
}

type MoodNoteStep struct {
	screen *MoodFormScreen
	input  textinput.Model
}

func NewMoodNoteStep(screen *MoodFormScreen) *MoodNoteStep {
	ti := textinput.New()
	ti.Placeholder = screen.t(i18n.RecordPromptNoteKey)
	ti.CharLimit = 200
	ti.Width = 50

	if screen.entry != nil {
		ti.SetValue(screen.entry.Note)
		screen.note = screen.entry.Note
	}

	return &MoodNoteStep{
		screen: screen,
		input:  ti,
	}
}

func (s *MoodNoteStep) Render(width, height int) string {
	var b strings.Builder

	b.WriteString("\n\n")

	b.WriteString(s.input.View())
	b.WriteString("\n\n")

	help := styles.HelpStyle.Render(s.screen.t(i18n.HelpNavigationRecordStep1Key))
	b.WriteString(help)

	return b.String()
}

func (s *MoodNoteStep) Update(msg tea.Msg) (forms.Step, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	s.screen.note = s.input.Value()
	return s, cmd
}

func (s *MoodNoteStep) Validate() error {

	return nil
}

func (s *MoodNoteStep) OnEnter() tea.Cmd {
	return s.input.Focus()
}

func (s *MoodNoteStep) OnExit() tea.Cmd {
	s.input.Blur()
	return nil
}

type MoodConfirmationStep struct {
	screen *MoodFormScreen
}

func NewMoodConfirmationStep(screen *MoodFormScreen) *MoodConfirmationStep {
	return &MoodConfirmationStep{
		screen: screen,
	}
}

func (s *MoodConfirmationStep) Render(width, height int) string {
	var b strings.Builder

	confirmKey := i18n.RecordPromptConfirmNewKey
	if s.screen.entry != nil {
		confirmKey = i18n.RecordPromptConfirmEditKey
	}
	b.WriteString(styles.SubtitleStyle.Render(s.screen.t(confirmKey)))
	b.WriteString("\n\n")

	moodLevel := entity.MoodLevel(s.screen.moodLevel)

	note := s.screen.note
	if note == "" {
		note = s.screen.t(i18n.CommonNoNoteKey)
	}

	desc := s.screen.t(moodLevel.StringKey())

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PastelDarkSlateBlue).
		Padding(1, 2).
		Render(fmt.Sprintf(
			s.screen.t(i18n.RecordBoxMoodKey)+"\n"+
				s.screen.t(i18n.RecordBoxNoteKey)+"\n"+
				s.screen.t(i18n.RecordBoxDateKey),
			moodLevel.Emoji(),
			desc,
			s.screen.moodLevel,
			note,
			s.screen.date.Format("02.01.2006"),
		))

	b.WriteString(box)
	b.WriteString("\n\n")

	help := styles.HelpStyle.Render(s.screen.t(i18n.HelpNavigationRecordStep2Key))
	b.WriteString(help)

	return b.String()
}

func (s *MoodConfirmationStep) Update(msg tea.Msg) (forms.Step, tea.Cmd) {

	return s, nil
}

func (s *MoodConfirmationStep) Validate() error {

	return nil
}

func (s *MoodConfirmationStep) OnEnter() tea.Cmd {
	return nil
}

func (s *MoodConfirmationStep) OnExit() tea.Cmd {
	return nil
}
