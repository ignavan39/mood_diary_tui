package state

import tea "github.com/charmbracelet/bubbletea"

// Screen представляет любой экран в приложении
type Screen interface {
	Init() tea.Cmd
	Update(tea.Msg) (Screen, tea.Cmd)
	View() string
}

// BaseState содержит общее состояние для всех экранов
type BaseState struct {
	Loading bool
	Error   error
	Width   int
	Height  int
}

func (s *BaseState) SetSize(w, h int) {
	s.Width = w
	s.Height = h
}

func (s *BaseState) SetError(err error) {
	s.Error = err
	s.Loading = false
}

func (s *BaseState) SetLoading(loading bool) {
	s.Loading = loading
}

func (s *BaseState) ClearError() {
	s.Error = nil
}
