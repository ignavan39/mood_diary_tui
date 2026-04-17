package state

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/domain/repository"
)

type ScreenType int

const (
	ScreenMenu ScreenType = iota
	ScreenMoodForm
	ScreenCalendar
	ScreenHistory
	ScreenStats
	ScreenSettings
	ScreenLanguageSettings
)

type NavigateMsg struct {
	To     ScreenType
	Params interface{}
}

func Navigate(to ScreenType, params interface{}) tea.Cmd {
	return func() tea.Msg {
		return NavigateMsg{To: to, Params: params}
	}
}

type MoodFormParams struct {
	Date  time.Time
	Entry *entity.MoodEntry
}

type CalendarParams struct {
	InitialDate time.Time
}

type StatsParams struct {
	Period string
}

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

type CalendarDataLoadedMsg struct {
	Data map[time.Time]*entity.MoodEntry
}

type CalendarDateSelectedMsg struct {
	Date time.Time
}

type HistoryLoadedMsg struct {
	Entries []*entity.MoodEntry
}

type StatsLoadedMsg struct {
	Stats   *repository.MoodStatistics
	Entries []*entity.MoodEntry
}

type ErrorMsg struct {
	Error error
}

type ConfirmActionMsg struct {
	Action    string
	OnConfirm tea.Cmd
	OnCancel  tea.Cmd
}

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

func NavigateToLanguageSettings() tea.Cmd {
	return Navigate(ScreenLanguageSettings, nil)
}
