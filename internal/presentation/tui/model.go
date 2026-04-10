package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ignavan39/mood-diary/internal/application/usecase"
	"github.com/ignavan39/mood-diary/internal/domain/repository"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/screens"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type Model struct {
	ctx          context.Context
	service      *usecase.MoodService
	translator   i18n.Translator
	settingsRepo repository.SettingsRepository

	current     state.Screen
	currentType state.ScreenType

	history []state.ScreenType

	width  int
	height int
}

func NewModel(
	ctx context.Context,
	service *usecase.MoodService,
	translator i18n.Translator,
	settingsRepo repository.SettingsRepository,
) *Model {
	m := &Model{
		ctx:          ctx,
		service:      service,
		translator:   translator,
		settingsRepo: settingsRepo,
		currentType:  state.ScreenMenu,
	}

	m.current = screens.NewMenuScreen(translator)

	return m
}

func (m *Model) Init() tea.Cmd {
	return m.current.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		var cmd tea.Cmd
		m.current, cmd = m.current.Update(msg)
		return m, cmd

	case tea.KeyMsg:

		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "q" || msg.String() == "esc" {
			if m.currentType == state.ScreenMenu {
				return m, tea.Quit
			}

		}

	case state.NavigateMsg:
		return m, m.navigate(msg.To, msg.Params)
	}

	var cmd tea.Cmd
	m.current, cmd = m.current.Update(msg)

	return m, cmd
}

func (m *Model) View() string {
	return m.current.View()
}

func (m *Model) navigate(to state.ScreenType, params interface{}) tea.Cmd {

	if to != m.currentType {
		m.history = append(m.history, m.currentType)
	}
	m.currentType = to

	switch to {
	case state.ScreenMenu:
		m.current = screens.NewMenuScreen(m.translator)

	case state.ScreenMoodForm:
		if p, ok := params.(state.MoodFormParams); ok {
			m.current = screens.NewMoodFormScreen(
				m.service,
				m.translator,
				p.Date,
				p.Entry,
			)
		}

	case state.ScreenCalendar:
		m.current = screens.NewCalendarScreen(m.service, m.translator)

	case state.ScreenHistory:
		m.current = screens.NewHistoryScreen(m.service, m.translator)

	case state.ScreenStats:
		m.current = screens.NewStatsScreen(m.service, m.translator)

	case state.ScreenSettings:
		m.current = screens.NewSettingsScreen(m.translator, m.settingsRepo)
	}

	return m.current.Init()
}

func (m *Model) navigateBack() tea.Cmd {
	if len(m.history) == 0 {
		return state.Navigate(state.ScreenMenu, nil)
	}

	previous := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]

	return state.Navigate(previous, nil)
}
