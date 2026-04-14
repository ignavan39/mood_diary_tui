package forms

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/constants"
)

type Step interface {
	Render(width, height int) string

	Update(tea.Msg) (Step, tea.Cmd)

	Validate() error

	OnEnter() tea.Cmd

	OnExit() tea.Cmd
}

type Wizard struct {
	steps        []Step
	currentStep  int
	complete     bool
	cancelled    bool
	width        int
	height       int
	errorMessage string
	translator   i18n.Translator
}

func NewWizard(steps []Step, translator i18n.Translator) *Wizard {
	return &Wizard{
		steps:       steps,
		currentStep: 0,
		translator:  translator,
	}
}

func (w *Wizard) SetSize(width, height int) {
	w.width = width
	w.height = height
}

func (w *Wizard) Update(msg tea.Msg) tea.Cmd {

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

	var cmd tea.Cmd
	w.steps[w.currentStep], cmd = w.steps[w.currentStep].Update(msg)
	return cmd
}

func (w *Wizard) nextStep() tea.Cmd {

	if err := w.CurrentStep().Validate(); err != nil {
		w.errorMessage = err.Error()
		return nil
	}

	exitCmd := w.CurrentStep().OnExit()

	if w.currentStep < len(w.steps)-1 {
		w.currentStep++
		enterCmd := w.CurrentStep().OnEnter()
		return tea.Batch(exitCmd, enterCmd)
	}

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

	progress := w.renderProgress()

	var errorView string
	if w.errorMessage != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(styles.TextDark).
			Background(styles.ErrorRed).
			Bold(true).
			Padding(1, 0)
		errorView = errorStyle.Render(constants.WarningSign + " " + w.errorMessage)
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

	var dots string
	for i := 0; i < total; i++ {
		if i < current {
			dots += constants.FilledDot + " "
		} else {
			dots += constants.EmptyDot + " "
		}
	}

	text := fmt.Sprintf("%s %d / %d", w.translator.T(i18n.CommonStepLabelKey), current, total)

	style := lipgloss.NewStyle().
		Foreground(styles.TextMuted).
		Align(lipgloss.Left)

	return style.Render(dots + "\n\n" + text)
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
