package tui

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
)

type CalendarModel struct {
	service    *usecase.MoodService
	translator i18n.Translator

	currentMonth time.Time
	selectedDate time.Time
	moodData     map[time.Time]*entity.MoodEntry

	cursorRow int
	cursorCol int

	loading  bool
	errorMsg string
}

func NewCalendarModel(service *usecase.MoodService, translator i18n.Translator) *CalendarModel {
	now := time.Now()
	return &CalendarModel{
		service:      service,
		translator:   translator,
		currentMonth: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		selectedDate: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		moodData:     make(map[time.Time]*entity.MoodEntry),
		loading:      true,
	}
}

func (m *CalendarModel) t(key string, args ...any) string {
	if m.translator == nil {
		return key
	}
	return m.translator.T(key, args...)
}

func (m *CalendarModel) Init() tea.Cmd {
	return m.loadMonthData()
}

func (m *CalendarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case CalendarDataLoadedMsg:
		m.moodData = msg.Data
		m.loading = false
		m.errorMsg = ""
		m.updateCursorPosition()
		return m, nil

	case ErrorMsg:
		m.errorMsg = msg.Error.Error()
		m.loading = false
		return m, nil
	}
	return m, nil
}

func (m *CalendarModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "h":
		return m, m.moveCursor(-1)
	case "right", "l":
		return m, m.moveCursor(1)
	case "up", "k":
		return m, m.moveCursor(-7)
	case "down", "j":
		return m, m.moveCursor(7)

	case "enter", " ":
		return m, m.selectDate()
	case "n":
		return m, NavigateToNewEntry(m.selectedDate)
	case "e":
		entry := m.moodData[m.selectedDate]
		if entry != nil {
			return m, NavigateToEdit(entry)
		}
		return m, NavigateToNewEntry(m.selectedDate)
	case "d":
		entry := m.moodData[m.selectedDate]
		if entry != nil {
			return m, NavigateToDelete(entry)
		}
	case "esc", "q":
		return m, Navigate(ScreenMenu)
	}
	return m, nil
}

func (m *CalendarModel) moveCursor(deltaDays int) tea.Cmd {
	newDate := m.selectedDate.AddDate(0, 0, deltaDays)

	if newDate.Year() != m.currentMonth.Year() || newDate.Month() != m.currentMonth.Month() {
		m.currentMonth = time.Date(newDate.Year(), newDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		m.selectedDate = newDate
		m.loading = true
		return m.loadMonthData()
	}

	m.selectedDate = newDate
	m.updateCursorPosition()
	return nil
}

func (m *CalendarModel) updateCursorPosition() {
	year, month := m.currentMonth.Year(), m.currentMonth.Month()
	firstDay := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)

	weekday := int(firstDay.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	offset := weekday - 1

	if m.selectedDate.Year() != year || m.selectedDate.Month() != month {
		m.cursorRow = 0
		m.cursorCol = offset
		return
	}

	daysDiff := m.selectedDate.Day() - 1
	totalPos := offset + daysDiff

	m.cursorRow = totalPos / 7
	m.cursorCol = totalPos % 7

	if m.cursorRow > 5 {
		m.cursorRow = 5
	}
	if m.cursorCol > 6 {
		m.cursorCol = 6
	}
}

func (m *CalendarModel) selectDate() tea.Cmd {
	return func() tea.Msg {
		return CalendarDateSelectedMsg{Date: m.selectedDate}
	}
}

func (m *CalendarModel) loadMonthData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		entries, err := m.service.GetAllMoods(ctx)

		data := make(map[time.Time]*entity.MoodEntry)
		if err == nil {
			for _, e := range entries {
				day := time.Date(e.Date.Year(), e.Date.Month(), e.Date.Day(), 0, 0, 0, 0, time.UTC)
				if day.Year() == m.currentMonth.Year() && day.Month() == m.currentMonth.Month() {
					data[day] = e
				}
			}
		} else {
			return ErrorMsg{Error: err}
		}

		return CalendarDataLoadedMsg{Data: data}
	}
}

func (m *CalendarModel) getMonthName(month time.Month) string {
	return m.t(fmt.Sprintf("date.month.%d", month))
}

func (m *CalendarModel) getWeekdayName(day time.Weekday) string {
	return m.t(fmt.Sprintf("date.weekday.%d", day))
}

func (m *CalendarModel) View() string {
	var b strings.Builder

	header := styles.HeaderStyle.Render(fmt.Sprintf("📅 %s %d", m.getMonthName(m.currentMonth.Month()), m.currentMonth.Year()))
	b.WriteString(header)
	b.WriteString("\n\n")

	if m.errorMsg != "" {
		b.WriteString(styles.ErrorStyle.Render(m.t("common.error_prefix") + m.errorMsg))
		b.WriteString("\n\n")
	}

	if m.loading {
		b.WriteString(styles.InfoStyle.Render(m.t("common.loading")))
		return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
	}

	weekdaysOrder := []time.Weekday{
		time.Monday, time.Tuesday, time.Wednesday, time.Thursday,
		time.Friday, time.Saturday, time.Sunday,
	}
	weekdaysStr := ""
	for _, wd := range weekdaysOrder {
		weekdaysStr += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render(m.getWeekdayName(wd))
	}
	b.WriteString(lipgloss.NewStyle().Foreground(styles.PastelLavender).Bold(true).Render(weekdaysStr))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 42) + "\n")

	rows := m.buildCalendarGrid()
	for i, row := range rows {
		line := ""
		for j, cell := range row {
			style := lipgloss.NewStyle().Width(6).Align(lipgloss.Center)

			if i == m.cursorRow && j == m.cursorCol {
				style = style.Background(styles.PastelLavender).Foreground(styles.TextLight).Bold(true)
			} else if entry, ok := m.moodData[cell.date]; ok {
				style = style.Foreground(styles.GetMoodColor(int(entry.Level)))
			} else if cell.date.Month() != m.currentMonth.Month() {
				style = style.Foreground(styles.TextMuted)
			}

			line += style.Render(cell.text)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n" + styles.FooterStyle.Render(m.renderFooter()))

	b.WriteString("\n" + styles.HelpStyle.Render(m.t("help.navigation.calendar")))

	return lipgloss.NewStyle().Padding(2, 4).Render(b.String())
}

type calendarCell struct {
	date time.Time
	text string
}

func (m *CalendarModel) buildCalendarGrid() [][]calendarCell {
	year, month := m.currentMonth.Year(), m.currentMonth.Month()
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
				if entry, ok := m.moodData[current]; ok {
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

func (m *CalendarModel) renderFooter() string {
	dateStr := m.selectedDate.Format("02.01.2006")
	if entry, ok := m.moodData[m.selectedDate]; ok {
		note := truncateNote(entry.Note, 25)
		return fmt.Sprintf("📍 %s | %s %d/10 | %s", dateStr, entry.Level.Emoji(), entry.Level.Int(), note)
	}
	return fmt.Sprintf("📍 %s | %s", dateStr, m.t("calendar.no_entry"))
}

func truncateNote(note string, max int) string {
	if len(note) <= max {
		return note
	}
	return note[:max-3] + "..."
}

type CalendarDataLoadedMsg struct {
	Data map[time.Time]*entity.MoodEntry
}

type CalendarDateSelectedMsg struct {
	Date time.Time
}

type NavigateToNewEntryMsg struct {
	Date time.Time
}

type NavigateToDeleteMsg struct {
	Entry *entity.MoodEntry
}

func NavigateToNewEntry(date time.Time) tea.Cmd {
	return func() tea.Msg {
		return NavigateToNewEntryMsg{Date: date}
	}
}

func NavigateToDelete(entry *entity.MoodEntry) tea.Cmd {
	return func() tea.Msg {
		return NavigateToDeleteMsg{Entry: entry}
	}
}
