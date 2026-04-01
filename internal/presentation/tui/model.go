package tui

import (
	"context"
	"mood-diary/internal/application/usecase"

	tea "github.com/charmbracelet/bubbletea"
)

type Screen int

const (
	ScreenMenu Screen = iota
	ScreenRecord
	ScreenStats
	ScreenHistory
	ScreenEdit
)

type Model struct {
	ctx           context.Context
	service       *usecase.MoodService
	currentScreen Screen

	menuModel    *MenuModel
	recordModel  *RecordModel
	statsModel   *StatsModel
	historyModel *HistoryModel
	editModel    *EditModel

	width  int
	height int
	err    error
}

func NewModel(ctx context.Context, service *usecase.MoodService) *Model {
	m := &Model{
		ctx:           ctx,
		service:       service,
		currentScreen: ScreenMenu,
	}

	m.menuModel = NewMenuModel()
	m.recordModel = NewRecordModel(service)
	m.statsModel = NewStatsModel(service)
	m.historyModel = NewHistoryModel(service)
	m.editModel = NewEditModel(service)

	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentScreen == ScreenMenu {
				return m, tea.Quit
			}
			m.currentScreen = ScreenMenu
			return m, nil
		}

	case NavigateMsg:
		m.currentScreen = msg.Screen
		return m, nil

	case ErrorMsg:
		m.err = msg.Error
		return m, nil
	}

	var cmd tea.Cmd
	switch m.currentScreen {
	case ScreenMenu:
		var menuModel tea.Model
		menuModel, cmd = m.menuModel.Update(msg)
		m.menuModel = menuModel.(*MenuModel)

	case ScreenRecord:
		var recordModel tea.Model
		recordModel, cmd = m.recordModel.Update(msg)
		m.recordModel = recordModel.(*RecordModel)

	case ScreenStats:
		var statsModel tea.Model
		statsModel, cmd = m.statsModel.Update(msg)
		m.statsModel = statsModel.(*StatsModel)

	case ScreenHistory:
		var historyModel tea.Model
		historyModel, cmd = m.historyModel.Update(msg)
		m.historyModel = historyModel.(*HistoryModel)

	case ScreenEdit:
		var editModel tea.Model
		editModel, cmd = m.editModel.Update(msg)
		m.editModel = editModel.(*EditModel)
	}

	return m, cmd
}

func (m *Model) View() string {
	switch m.currentScreen {
	case ScreenMenu:
		return m.menuModel.View()
	case ScreenRecord:
		return m.recordModel.View()
	case ScreenStats:
		return m.statsModel.View()
	case ScreenHistory:
		return m.historyModel.View()
	case ScreenEdit:
		return m.editModel.View()
	default:
		return "Unknown screen"
	}
}

type NavigateMsg struct {
	Screen Screen
}

type ErrorMsg struct {
	Error error
}

func Navigate(screen Screen) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{Screen: screen}
	}
}
