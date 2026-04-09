package screens

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/application/usecase"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/components"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/forms"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type MoodFormScreen struct {
	state.BaseState

	service    *usecase.MoodService
	translator i18n.Translator

	// Данные формы
	date  time.Time
	entry *entity.MoodEntry // nil = create mode

	// UI компоненты
	wizard *forms.Wizard

	// Данные из шагов
	moodLevel int
	note      string

	// Состояние
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

	// Создаем шаги визарда
	steps := []forms.Step{
		NewMoodLevelStep(screen),
		NewMoodNoteStep(screen),
		NewMoodConfirmationStep(screen),
	}

	screen.wizard = forms.NewWizard(steps)

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
		// Автоматический возврат через 1.5 секунды
		return s, tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
			return state.NavigateMsg{To: state.ScreenMenu}
		})

	case state.ErrorMsg:
		s.SetError(msg.Error)
		s.saving = false
		return s, nil
	}

	// Делегируем визарду
	cmd := s.wizard.Update(msg)

	// Проверяем завершение
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
	// Успешное сохранение
	if s.saved {
		return components.NewSuccess(s.t("record.success")).View()
	}

	// Процесс сохранения
	if s.saving {
		loading := components.NewLoading(s.t("common.saving"))
		return loading.View()
	}

	// Заголовок
	var title string
	if s.entry != nil {
		title = "✏️  " + s.t("edit.title")
	} else {
		title = "📝 " + s.t("record.title")
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#D4BAFF")).
		Align(lipgloss.Center).
		Width(s.Width)

	header := headerStyle.Render(title)

	// Дата
	dateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		Align(lipgloss.Center).
		Width(s.Width)

	dateStr := dateStyle.Render(s.date.Format("02.01.2006"))

	// Контент визарда
	wizardContent := s.wizard.View()

	// Ошибка
	var errorView string
	if s.Error != nil {
		errorView = components.NewError(s.Error.Error()).View()
	}

	// Общая справка
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		Align(lipgloss.Center).
		Width(s.Width)

	help := helpStyle.Render(s.t("help.wizard"))

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		dateStr,
		"",
		wizardContent,
		errorView,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Width(s.Width).
		Height(s.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(content)
}

func (s *MoodFormScreen) save() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		var err error
		if s.entry != nil {
			// Update mode
			err = s.service.UpdateMood(ctx, s.date, s.moodLevel, s.note)
		} else {
			// Create mode
			err = s.service.RecordMood(ctx, s.moodLevel, s.note, &s.date)
		}

		if err != nil {
			return state.ErrorMsg{Error: err}
		}

		return state.MoodSavedMsg{}
	}
}

// ============================================================================
// WIZARD STEPS - Шаги для формы настроения
// ============================================================================

// Шаг 1: Выбор уровня настроения
type MoodLevelStep struct {
	screen   *MoodFormScreen
	selector *components.MoodSelector
}

func NewMoodLevelStep(screen *MoodFormScreen) *MoodLevelStep {
	initial := 5
	if screen.entry != nil {
		initial = screen.entry.Level.Int()
	}

	step := &MoodLevelStep{
		screen:   screen,
		selector: components.NewMoodSelector(initial),
	}

	// Коллбэк при изменении значения
	step.selector.OnChange(func(value int) {
		screen.moodLevel = value
	})

	// Устанавливаем начальное значение
	screen.moodLevel = initial

	return step
}

func (s *MoodLevelStep) Render(width, height int) string {
	prompt := s.screen.t("record.prompt_level")

	promptStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFB3BA")).
		Align(lipgloss.Center).
		Width(width)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		promptStyle.Render(prompt),
		"",
		s.selector.View(),
	)
}

func (s *MoodLevelStep) Update(msg tea.Msg) (forms.Step, tea.Cmd) {
	// Автоматический переход на Tab или Enter
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if keyMsg.String() == "enter" {
			return s, nil // Wizard обработает переход
		}
	}

	cmd := s.selector.Update(msg)
	return s, cmd
}

func (s *MoodLevelStep) Validate() error {
	// Уровень всегда валиден т.к. selector контролирует диапазон
	return nil
}

func (s *MoodLevelStep) OnEnter() tea.Cmd {
	s.selector.Focus()
	return nil
}

func (s *MoodLevelStep) OnExit() tea.Cmd {
	s.selector.Blur()
	return nil
}

// Шаг 2: Ввод заметки
type MoodNoteStep struct {
	screen *MoodFormScreen
	input  textinput.Model
}

func NewMoodNoteStep(screen *MoodFormScreen) *MoodNoteStep {
	ti := textinput.New()
	ti.Placeholder = screen.t("record.prompt_note")
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
	prompt := s.screen.t("record.prompt_note_description")

	promptStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#BAFFC9")).
		Align(lipgloss.Center).
		Width(width)

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#BAFFC9")).
		Padding(1, 2).
		Width(width - 4).
		Align(lipgloss.Center)

	counterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9B9B9B")).
		Align(lipgloss.Center)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		promptStyle.Render(prompt),
		"",
		inputStyle.Render(s.input.View()),
		"",
		counterStyle.Render(fmt.Sprintf("%d/%d символов", len(s.input.Value()), s.input.CharLimit)),
	)
}

func (s *MoodNoteStep) Update(msg tea.Msg) (forms.Step, tea.Cmd) {
	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	s.screen.note = s.input.Value()
	return s, cmd
}

func (s *MoodNoteStep) Validate() error {
	// Заметка опциональна, всегда валидна
	return nil
}

func (s *MoodNoteStep) OnEnter() tea.Cmd {
	return s.input.Focus()
}

func (s *MoodNoteStep) OnExit() tea.Cmd {
	s.input.Blur()
	return nil
}

// Шаг 3: Подтверждение
type MoodConfirmationStep struct {
	screen *MoodFormScreen
}

func NewMoodConfirmationStep(screen *MoodFormScreen) *MoodConfirmationStep {
	return &MoodConfirmationStep{
		screen: screen,
	}
}

func (s *MoodConfirmationStep) Render(width, height int) string {
	level := entity.MoodLevel(s.screen.moodLevel)

	// Заголовок
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#D4BAFF")).
		Align(lipgloss.Center).
		Width(width)

	title := titleStyle.Render(s.screen.t("record.confirm_title"))

	// Данные
	dataStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#D4BAFF")).
		Padding(1, 2).
		Width(width - 4)

	dateStr := s.screen.date.Format("02.01.2006")
	moodStr := fmt.Sprintf("%s %s (%d/10)", level.Emoji(), level.String(), s.screen.moodLevel)

	noteStr := s.screen.note
	if noteStr == "" {
		noteStr = s.screen.t("record.no_note")
	}

	data := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render(s.screen.t("record.date")+": ")+dateStr,
		lipgloss.NewStyle().Bold(true).Render(s.screen.t("record.mood")+": ")+moodStr,
		lipgloss.NewStyle().Bold(true).Render(s.screen.t("record.note")+": ")+noteStr,
	)

	// Кнопки
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1)

	saveBtn := buttonStyle.Copy().
		Background(lipgloss.Color("#51CF66")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Render("✓ " + s.screen.t("common.save") + " (Tab)")

	cancelBtn := buttonStyle.Copy().
		Background(lipgloss.Color("#FF6B6B")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Render("✗ " + s.screen.t("common.cancel") + " (Esc)")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, saveBtn, cancelBtn)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		dataStyle.Render(data),
		"",
		buttons,
	)
}

func (s *MoodConfirmationStep) Update(msg tea.Msg) (forms.Step, tea.Cmd) {
	// Wizard обработает Tab и Esc
	return s, nil
}

func (s *MoodConfirmationStep) Validate() error {
	// Всегда валидно, просто показываем подтверждение
	return nil
}

func (s *MoodConfirmationStep) OnEnter() tea.Cmd {
	return nil
}

func (s *MoodConfirmationStep) OnExit() tea.Cmd {
	return nil
}
