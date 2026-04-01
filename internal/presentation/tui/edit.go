package tui

import (
	"mood-diary/internal/application/usecase"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type EditModel struct {
	service   *usecase.MoodService
	noteInput textinput.Model
}

func NewEditModel(service *usecase.MoodService) *EditModel {
	ti := textinput.New()

	return &EditModel{
		service:   service,
		noteInput: ti,
	}
}

func (m *EditModel) Init() tea.Cmd {
	return nil
}

func (m *EditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			return m, Navigate(ScreenHistory)
		}
	}
	return m, nil
}

func (m *EditModel) View() string {
	return "Edit Screen (TODO)"
}
