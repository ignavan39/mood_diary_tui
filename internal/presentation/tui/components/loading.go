package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/constants"
)

type LoadingIndicator struct {
	message string
	spinner int
}

func NewLoading(message string) *LoadingIndicator {
	return &LoadingIndicator{
		message: message,
		spinner: 0,
	}
}

type TickMsg time.Time

func (l *LoadingIndicator) Init() tea.Cmd {
	return l.tick()
}

func (l *LoadingIndicator) Update(msg tea.Msg) (LoadingIndicator, tea.Cmd) {
	if _, ok := msg.(TickMsg); ok {
		l.spinner = (l.spinner + 1) % len(constants.LoadingFrames)
		return *l, l.tick()
	}
	return *l, nil
}

func (l *LoadingIndicator) tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (l *LoadingIndicator) View() string {
	spinner := constants.LoadingFrames[l.spinner]

	style := lipgloss.NewStyle().
		Foreground(styles.PastelPink).
		Padding(2, 0)

	return style.Render(string(spinner) + " " + l.message)
}

type SuccessMessage struct {
	message string
}

func NewSuccess(message string) *SuccessMessage {
	return &SuccessMessage{message: message}
}

func (s *SuccessMessage) View() string {
	style := lipgloss.NewStyle().
		Foreground(styles.TextDark).
		Background(styles.SuccessGreen).
		Bold(true).
		Padding(2, 0)

	return style.Render(constants.Checkmark + " " + s.message)
}

type ErrorMessage struct {
	message string
}

func NewError(message string) *ErrorMessage {
	return &ErrorMessage{message: message}
}

func (e *ErrorMessage) View() string {
	style := lipgloss.NewStyle().
		Foreground(styles.TextDark).
		Background(styles.ErrorRed).
		Bold(true).
		Padding(1, 0)

	return style.Render(constants.WarningSign + " " + e.message)
}
