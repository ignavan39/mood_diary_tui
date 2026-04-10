package forms

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

// Step представляет один шаг в визарде
type Step interface {
	// Render отрисовывает содержимое шага
	Render(width, height int) string

	// Update обрабатывает события для этого шага
	Update(tea.Msg) (Step, tea.Cmd)

	// Validate проверяет можно ли перейти к следующему шагу
	Validate() error

	// OnEnter вызывается при входе на шаг
	OnEnter() tea.Cmd

	// OnExit вызывается при выходе с шага
	OnExit() tea.Cmd
}

// Wizard управляет переходами между шагами
type Wizard struct {
	steps        []Step
	currentStep  int
	complete     bool
	cancelled    bool
	width        int
	height       int
	errorMessage string
}

func NewWizard(steps []Step) *Wizard {
	return &Wizard{
		steps:       steps,
		currentStep: 0,
	}
}

func (w *Wizard) SetSize(width, height int) {
	w.width = width
	w.height = height
}

func (w *Wizard) Update(msg tea.Msg) tea.Cmd {
	// Сброс ошибки при новом вводе
	if _, ok := msg.(tea.KeyMsg); ok {
		w.errorMessage = ""
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "esc":
			w.cancelled = true
			return nil

		case "enter":
			return w.nextStep()
		}
	}

	// Делегирование текущему шагу
	var cmd tea.Cmd
	w.steps[w.currentStep], cmd = w.steps[w.currentStep].Update(msg)
	return cmd
}

func (w *Wizard) nextStep() tea.Cmd {
	// Валидация текущего шага
	if err := w.CurrentStep().Validate(); err != nil {
		w.errorMessage = err.Error()
		return nil
	}

	// Выход из текущего шага
	exitCmd := w.CurrentStep().OnExit()

	// Переход к следующему
	if w.currentStep < len(w.steps)-1 {
		w.currentStep++
		enterCmd := w.CurrentStep().OnEnter()
		return tea.Batch(exitCmd, enterCmd)
	}

	// Последний шаг - завершение
	w.complete = true
	return exitCmd
}

func (w *Wizard) previousStep() tea.Cmd {
	if w.currentStep > 0 {
		exitCmd := w.CurrentStep().OnExit()
		w.currentStep--
		enterCmd := w.CurrentStep().OnEnter()
		return tea.Batch(exitCmd, enterCmd)
	}
	return nil
}

func (w *Wizard) View() string {
	if w.currentStep >= len(w.steps) {
		return ""
	}

	content := w.CurrentStep().Render(w.width, w.height)

	// Добавляем индикатор прогресса
	progress := w.renderProgress()

	// Добавляем сообщение об ошибке если есть
	var errorView string
	if w.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(styles.TextDark).
			Background(styles.ErrorRed).
			Bold(true).
			Padding(1, 0)
		errorView = errorStyle.Render("⚠ " + w.errorMessage)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		progress,
		"",
		content,
		errorView,
	)
}

func (w *Wizard) renderProgress() string {
	current := w.currentStep + 1
	total := len(w.steps)

	// Визуальный прогресс-бар
	filled := "●"
	empty := "○"

	var dots string
	for i := 0; i < total; i++ {
		if i < current {
			dots += filled + " "
		} else {
			dots += empty + " "
		}
	}

	text := fmt.Sprintf("Шаг %d из %d", current, total)

	style := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Align(lipgloss.Center)

	return style.Render(dots + "\n" + text)
}

func (w *Wizard) CurrentStep() Step {
	return w.steps[w.currentStep]
}

func (w *Wizard) IsComplete() bool {
	return w.complete
}

func (w *Wizard) IsCancelled() bool {
	return w.cancelled
}

func (w *Wizard) Progress() (current int, total int) {
	return w.currentStep + 1, len(w.steps)
}

func (w *Wizard) Reset() {
	w.currentStep = 0
	w.complete = false
	w.cancelled = false
	w.errorMessage = ""
}
