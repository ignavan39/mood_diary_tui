package state

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/domain/repository"
)

// ScreenType определяет тип экрана
type ScreenType int

const (
	ScreenMenu ScreenType = iota
	ScreenMoodForm
	ScreenCalendar
	ScreenHistory
	ScreenStats
	ScreenSettings
)

// NavigateMsg - унифицированное сообщение для навигации
type NavigateMsg struct {
	To     ScreenType
	Params interface{}
}

func Navigate(to ScreenType, params interface{}) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{To: to, Params: params}
	}
}

// Параметры для различных экранов

type MoodFormParams struct {
	Date  time.Time
	Entry *entity.MoodEntry // nil = create, filled = edit
}

type CalendarParams struct {
	InitialDate time.Time
}

type StatsParams struct {
	Period string
}

// Сообщения результатов операций

type MoodSavedMsg struct {
	Entry *entity.MoodEntry
}

type MoodDeletedMsg struct {
	Date time.Time
}

type DataLoadedMsg struct {
	Data  interface{}
	Error error
}

func (m DataLoadedMsg) IsSuccess() bool {
	return m.Error == nil
}

func (m DataLoadedMsg) GetError() error {
	return m.Error
}

// Сообщения для календаря
type CalendarDataLoadedMsg struct {
	Data map[time.Time]*entity.MoodEntry
}

type CalendarDateSelectedMsg struct {
	Date time.Time
}

// Сообщения для истории
type HistoryLoadedMsg struct {
	Entries []*entity.MoodEntry
}

// Сообщения для статистики
type StatsLoadedMsg struct {
	Stats   *repository.MoodStatistics
	Entries []*entity.MoodEntry
}

// Общие сообщения ошибок
type ErrorMsg struct {
	Error error
}

// Сообщение для подтверждения действия
type ConfirmActionMsg struct {
	Action    string
	OnConfirm tea.Cmd
	OnCancel  tea.Cmd
}

// Вспомогательные функции для создания сообщений

func NavigateToMenu() tea.Cmd {
	return Navigate(ScreenMenu, nil)
}

func NavigateToMoodForm(date time.Time, entry *entity.MoodEntry) tea.Cmd {
	return Navigate(ScreenMoodForm, MoodFormParams{
		Date:  date,
		Entry: entry,
	})
}

func NavigateToCalendar(date time.Time) tea.Cmd {
	return Navigate(ScreenCalendar, CalendarParams{
		InitialDate: date,
	})
}

func NavigateToHistory() tea.Cmd {
	return Navigate(ScreenHistory, nil)
}

func NavigateToStats(period string) tea.Cmd {
	return Navigate(ScreenStats, StatsParams{
		Period: period,
	})
}

func NavigateToSettings() tea.Cmd {
	return Navigate(ScreenSettings, nil)
}
