package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ignavan39/mood-diary/internal/application/usecase"
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
	var cmds []tea.Cmd

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
		oldScreen := m.currentScreen
		m.currentScreen = msg.Screen

		if oldScreen != msg.Screen {
			cmd := m.initCurrentScreen()
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case NavigateToEditMsg:

		m.currentScreen = ScreenEdit
		m.editModel.SetEntry(msg.Entry)

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

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m *Model) initCurrentScreen() tea.Cmd {
	switch m.currentScreen {
	case ScreenMenu:
		return m.menuModel.Init()
	case ScreenRecord:

		m.recordModel = NewRecordModel(m.service)
		return m.recordModel.Init()
	case ScreenStats:

		m.statsModel = NewStatsModel(m.service)
		return m.statsModel.Init()
	case ScreenHistory:

		m.historyModel = NewHistoryModel(m.service)
		return m.historyModel.Init()
	case ScreenEdit:
		return m.editModel.Init()
	}
	return nil
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
