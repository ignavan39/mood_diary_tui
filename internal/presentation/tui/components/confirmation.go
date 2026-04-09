package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
)

type ConfirmationDialog struct {
	message    string
	yesLabel   string
	noLabel    string
	onYes      tea.Cmd
	onNo       tea.Cmd
	result     *bool
	focused    bool
	yesPressed bool
}

func NewConfirmation(message string, onYes, onNo tea.Cmd) *ConfirmationDialog {
	return &ConfirmationDialog{
		message:  message,
		yesLabel: "Да",
		noLabel:  "Нет",
		onYes:    onYes,
		onNo:     onNo,
		focused:  true,
	}
}

func (c *ConfirmationDialog) SetLabels(yes, no string) {
	c.yesLabel = yes
	c.noLabel = no
}

func (c *ConfirmationDialog) Focus() {
	c.focused = true
}

func (c *ConfirmationDialog) Blur() {
	c.focused = false
}

func (c *ConfirmationDialog) Update(msg tea.Msg) tea.Cmd {
	if !c.focused || c.result != nil {
		return nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "y", "Y", "enter":
			result := true
			c.result = &result
			c.yesPressed = true
			return c.onYes

		case "n", "N", "esc":
			result := false
			c.result = &result
			c.yesPressed = false
			return c.onNo
		}
	}
	return nil
}

func (c *ConfirmationDialog) View() string {
	if c.result != nil {
		return ""
	}

	dialog := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PastelLavender).
		Padding(1, 2).
		Width(50)

	yesStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Background(lipgloss.Color("#51CF66")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	noStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Margin(0, 1).
		Background(lipgloss.Color("#FF6B6B")).
		Foreground(lipgloss.Color("#FFFFFF"))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		c.message,
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			yesStyle.Render(c.yesLabel+" (y)"),
			"  ",
			noStyle.Render(c.noLabel+" (n)"),
		),
	)

	return dialog.Render(content)
}

func (c *ConfirmationDialog) IsDone() bool {
	return c.result != nil
}

func (c *ConfirmationDialog) WasConfirmed() bool {
	return c.result != nil && *c.result
}

func (c *ConfirmationDialog) Reset() {
	c.result = nil
	c.yesPressed = false
}
