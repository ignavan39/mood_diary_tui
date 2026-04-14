package screens

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ignavan39/mood-diary/internal/application/usecase"
	"github.com/ignavan39/mood-diary/internal/domain/entity"
	"github.com/ignavan39/mood-diary/internal/infrastructure/i18n"
	"github.com/ignavan39/mood-diary/internal/presentation/styles"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/formatters"
	"github.com/ignavan39/mood-diary/internal/presentation/tui/state"
)

type CalendarScreen struct {
	state.BaseState

	service    *usecase.MoodService
	translator i18n.Translator

	currentMonth time.Time
	selectedDate time.Time
	moodData     map[time.Time]*entity.MoodEntry

	cursorRow int
	cursorCol int
}

func NewCalendarScreen(service *usecase.MoodService, translator i18n.Translator) *CalendarScreen {
	now := time.Now()
	return &CalendarScreen{
		service:      service,
		translator:   translator,
		currentMonth: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		selectedDate: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		moodData:     make(map[time.Time]*entity.MoodEntry),
	}
}

func (s *CalendarScreen) t(key string, args ...interface{}) string {
	if s.translator == nil {
		return key
	}
	return s.translator.T(key, args...)
}

func (s *CalendarScreen) Init() tea.Cmd {
	s.SetLoading(true)
	return s.loadMonthData()
}

func (s *CalendarScreen) Update(msg tea.Msg) (state.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.SetSize(msg.Width, msg.Height)

	case tea.KeyMsg:
		return s.handleKeyMsg(msg)

	case state.CalendarDataLoadedMsg:
		s.moodData = msg.Data
		s.SetLoading(false)
		s.ClearError()
		s.updateCursorPosition()
		return s, nil

	case state.ErrorMsg:
		s.SetError(msg.Error)
		return s, nil

	case state.MoodDeletedMsg:

		s.SetLoading(true)
		return s, s.loadMonthData()
	}

	return s, nil
}

func (s *CalendarScreen) handleKeyMsg(msg tea.KeyMsg) (state.Screen, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		return s, s.moveCursor(-1)
	case "right", "l":
		return s, s.moveCursor(1)
	case "up", "k":
		return s, s.moveCursor(-7)
	case "down", "j":
		return s, s.moveCursor(7)

	case "n":

		return s, state.NavigateToMoodForm(s.selectedDate, nil)

	case "e", "enter":

		entry := s.moodData[s.selectedDate]
		return s, state.NavigateToMoodForm(s.selectedDate, entry)

	case "d":

		entry := s.moodData[s.selectedDate]
		if entry != nil {
			return s, s.deleteMood(entry)
		}

	case "esc", "q":
		return s, state.NavigateToMenu()
	}

	return s, nil
}

func (s *CalendarScreen) moveCursor(deltaDays int) tea.Cmd {
	newDate := s.selectedDate.AddDate(0, 0, deltaDays)

	if newDate.Year() != s.currentMonth.Year() || newDate.Month() != s.currentMonth.Month() {
		s.currentMonth = time.Date(newDate.Year(), newDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		s.selectedDate = newDate
		s.SetLoading(true)
		return s.loadMonthData()
	}

	s.selectedDate = newDate
	s.updateCursorPosition()
	return nil
}

func (s *CalendarScreen) updateCursorPosition() {
	year, month := s.currentMonth.Year(), s.currentMonth.Month()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	weekday := int(firstDay.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	offset := weekday - 1

	if s.selectedDate.Year() != year || s.selectedDate.Month() != month {
		s.cursorRow = 0
		s.cursorCol = offset
		return
	}

	daysDiff := s.selectedDate.Day() - 1
	totalPos := offset + daysDiff

	s.cursorRow = totalPos / 7
	s.cursorCol = totalPos % 7

	if s.cursorRow > 5 {
		s.cursorRow = 5
	}
	if s.cursorCol > 6 {
		s.cursorCol = 6
	}
}

func (s *CalendarScreen) loadMonthData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		entries, err := s.service.GetAllMoods(ctx)

		data := make(map[time.Time]*entity.MoodEntry)
		if err == nil {
			for _, e := range entries {
				day := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, time.UTC)
				if day.Year() == s.currentMonth.Year() && day.Month() == s.currentMonth.Month() {
					data[day] = e
				}
			}
		} else {
			return state.ErrorMsg{Error: err}
		}

		return state.CalendarDataLoadedMsg{Data: data}
	}
}

func (s *CalendarScreen) deleteMood(entry *entity.MoodEntry) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		err := s.service.DeleteMood(ctx, entry.Date)
		if err != nil {
			return state.ErrorMsg{Error: err}
		}
		return state.MoodDeletedMsg{Date: entry.Date}
	}
}

func (s *CalendarScreen) View() string {
	var b strings.Builder

	monthName := s.t(fmt.Sprintf("date.month.%d", s.currentMonth.Month()))
	header := styles.HeaderStyle.Render(fmt.Sprintf("📅 %s %d", monthName, s.currentMonth.Year()))
	b.WriteString(header)
	b.WriteString("\n\n")

	if s.Error != nil {
		b.WriteString(styles.ErrorStyle.Render(s.t(i18n.CommonErrorPrefixKey) + s.Error.Error()))
		b.WriteString("\n\n")
	}

	if s.Loading {
		b.WriteString(styles.InfoStyle.Render(s.t(i18n.CommonLoaderMessageKey)))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	weekdaysOrder := []time.Weekday{
		time.Monday, time.Tuesday, time.Wednesday, time.Thursday,
		time.Friday, time.Saturday, time.Sunday,
	}
	weekdaysStr := ""
	for _, wd := range weekdaysOrder {
		weekdayName := s.t(fmt.Sprintf("date.weekday.%d", wd))
		weekdaysStr += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render(weekdayName)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(styles.PastelLavender).Bold(true).Render(weekdaysStr))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 42) + "\n")

	rows := s.buildCalendarGrid()
	for i, row := range rows {
		var line strings.Builder
		for j, cell := range row {
			style := lipgloss.NewStyle().Width(6).Align(lipgloss.Center)

			if i == s.cursorRow && j == s.cursorCol {
				style = style.Background(styles.PastelLavender).Foreground(styles.TextLight).Bold(true)
			} else if entry, ok := s.moodData[cell.date]; ok {

				style = style.Foreground(styles.GetMoodColor(int(entry.Level)))
			} else if cell.date.Month() != s.currentMonth.Month() {

				style = style.Foreground(styles.TextMuted)
			}

			line.WriteString(style.Render(cell.text))
		}
		b.WriteString(line.String() + "\n")
	}

	b.WriteString("\n" + styles.FooterStyle.Render(s.renderFooter()))

	b.WriteString("\n" + styles.HelpStyle.Render(s.t("help.navigation.calendar")))

	return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
}

type calendarCell struct {
	date time.Time
	text string
}

func (s *CalendarScreen) buildCalendarGrid() [][]calendarCell {
	year, month := s.currentMonth.Year(), s.currentMonth.Month()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	weekday := int(firstDay.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	offset := weekday - 1
	gridStart := firstDay.AddDate(0, 0, -offset)

	var grid [][]calendarCell
	for week := 0; week < 6; week++ {
		var row []calendarCell
		for day := 0; day < 7; day++ {
			current := gridStart.AddDate(0, 0, week*7+day)
			cell := calendarCell{date: current}

			if current.Month() != month {
				cell.text = lipgloss.NewStyle().Foreground(styles.TextMuted).Render("  ")
			} else {
				dayNum := fmt.Sprintf("%2d", current.Day())
				if entry, ok := s.moodData[current]; ok {
					cell.text = dayNum + entry.Level.Emoji()
				} else {
					cell.text = dayNum
				}
			}
			row = append(row, cell)
		}
		grid = append(grid, row)
	}
	return grid
}

func (s *CalendarScreen) renderFooter() string {
	dateStr := formatters.FormatDate(s.selectedDate)
	if entry, ok := s.moodData[s.selectedDate]; ok {
		note := formatters.TruncateNote(entry.Note, 25)
		return fmt.Sprintf("📍 %s | %s %d/10 | %s", dateStr, entry.Level.Emoji(), entry.Level.Int(), note)
	}
	return fmt.Sprintf("📍 %s | %s", dateStr, s.t(i18n.CalendarNoEntryKey))
}
