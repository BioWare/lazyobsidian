// Package views contains the different views/pages of the application.
package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/ui/icons"
	"github.com/BioWare/lazyobsidian/internal/ui/layout"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// CalendarViewMode represents the calendar display mode.
type CalendarViewMode int

const (
	CalendarModeMonth CalendarViewMode = iota
	CalendarModeYear
	CalendarModeDay
)

// CalendarLayer represents what data to display on the calendar.
type CalendarLayer int

const (
	LayerTasks CalendarLayer = 1 << iota
	LayerGoals
	LayerJournal
	LayerActivity
	LayerAll = LayerTasks | LayerGoals | LayerJournal | LayerActivity
)

// CalendarEvent represents an event/task on a specific day.
type CalendarEvent struct {
	Date  time.Time
	Title string
	Type  string // "task", "goal", "journal", "activity"
	Done  bool
}

// Calendar represents the calendar view.
type Calendar struct {
	Width  int
	Height int

	Mode         CalendarViewMode
	CurrentDate  time.Time // The date to display (month/year determines view center)
	SelectedDate time.Time // The currently selected date
	Today        time.Time

	// Data
	Events       []CalendarEvent
	ActivityData map[string]int // Date string -> activity count (pomodoros)

	// Settings
	FirstDayOfWeek int // 0 = Sunday, 1 = Monday
	ActiveLayers   CalendarLayer

	// Navigation state
	focused bool
}

// NewCalendar creates a new calendar view.
func NewCalendar(width, height int) *Calendar {
	now := time.Now()
	return &Calendar{
		Width:          width,
		Height:         height,
		Mode:           CalendarModeMonth,
		CurrentDate:    now,
		SelectedDate:   now,
		Today:          now,
		FirstDayOfWeek: 1, // Monday
		ActiveLayers:   LayerAll,
		ActivityData:   make(map[string]int),
	}
}

// SetFocused sets the focus state.
func (c *Calendar) SetFocused(focused bool) {
	c.focused = focused
}

// SetEvents sets the calendar events.
func (c *Calendar) SetEvents(events []CalendarEvent) {
	c.Events = events
}

// SetActivityData sets the activity data for the heatmap.
func (c *Calendar) SetActivityData(data map[string]int) {
	c.ActivityData = data
}

// Render renders the calendar based on the current mode.
func (c *Calendar) Render() string {
	switch c.Mode {
	case CalendarModeYear:
		return c.renderYearView()
	case CalendarModeDay:
		return c.renderDayView()
	default:
		return c.renderMonthView()
	}
}

// renderMonthView renders the month calendar view.
func (c *Calendar) renderMonthView() string {
	if c.Width <= 0 || c.Height <= 0 {
		return ""
	}

	// Split layout: Calendar grid on left, day panel on right
	calendarWidth := c.Width * 60 / 100
	panelWidth := c.Width - calendarWidth

	calendar := c.renderMonthCalendar(calendarWidth, c.Height)
	dayPanel := c.renderDayPanel(panelWidth, c.Height)

	return lipgloss.JoinHorizontal(lipgloss.Top, calendar, dayPanel)
}

// renderMonthCalendar renders the month calendar grid.
func (c *Calendar) renderMonthCalendar(width, height int) string {
	frame := layout.NewFrame(width, height)

	monthName := c.CurrentDate.Format("January 2006")
	frame.SetTitle(monthName)
	frame.SetBorder(layout.BorderRounded)
	frame.SetFocused(c.focused)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	contentWidth := frame.ContentWidth()
	contentHeight := frame.ContentHeight()

	var lines []string

	// Day header
	dayWidth := contentWidth / 7
	var dayHeaders []string
	weekdays := c.getWeekdayNames()
	for _, day := range weekdays {
		dayHeaders = append(dayHeaders, layout.PadCenter(day, dayWidth))
	}
	headerStyle := theme.S.CalendarHeader
	lines = append(lines, headerStyle.Render(strings.Join(dayHeaders, "")))

	// Get first day of month
	firstOfMonth := time.Date(c.CurrentDate.Year(), c.CurrentDate.Month(), 1, 0, 0, 0, 0, c.CurrentDate.Location())
	firstWeekday := int(firstOfMonth.Weekday())

	// Adjust for first day of week
	if c.FirstDayOfWeek == 1 { // Monday
		firstWeekday = (firstWeekday + 6) % 7
	}

	// Days in month
	daysInMonth := daysInMonth(c.CurrentDate.Year(), c.CurrentDate.Month())

	// Available rows for days
	weeksAvailable := contentHeight - 1 // -1 for header

	// Build calendar grid
	dayNum := 1
	for week := 0; week < weeksAvailable && dayNum <= daysInMonth; week++ {
		var weekLine []string

		for day := 0; day < 7; day++ {
			if week == 0 && day < firstWeekday {
				// Empty cell before month starts
				weekLine = append(weekLine, strings.Repeat(" ", dayWidth))
			} else if dayNum > daysInMonth {
				// Empty cell after month ends
				weekLine = append(weekLine, strings.Repeat(" ", dayWidth))
			} else {
				// Render day cell
				cell := c.renderDayCell(dayNum, dayWidth)
				weekLine = append(weekLine, cell)
				dayNum++
			}
		}

		lines = append(lines, strings.Join(weekLine, ""))
	}

	frame.SetContentLines(lines)
	return frame.Render()
}

// renderDayCell renders a single day cell in the calendar.
func (c *Calendar) renderDayCell(day, width int) string {
	date := time.Date(c.CurrentDate.Year(), c.CurrentDate.Month(), day, 0, 0, 0, 0, c.CurrentDate.Location())
	dateStr := date.Format("2006-01-02")

	var style lipgloss.Style

	// Determine style based on date
	isToday := c.isSameDay(date, c.Today)
	isSelected := c.isSameDay(date, c.SelectedDate)
	isWeekend := date.Weekday() == time.Saturday || date.Weekday() == time.Sunday

	if isToday {
		style = theme.S.CalendarDayToday
	} else if isSelected {
		style = theme.S.CalendarDaySelected
	} else if isWeekend {
		style = theme.S.CalendarWeekend
	} else {
		style = theme.S.CalendarDay
	}

	// Day number
	dayStr := fmt.Sprintf("%2d", day)

	// Activity indicator
	indicator := " "
	if activity, ok := c.ActivityData[dateStr]; ok && activity > 0 {
		if activity >= 5 {
			indicator = "●"
		} else if activity >= 3 {
			indicator = "◐"
		} else {
			indicator = "○"
		}
	}

	// Check for events
	hasEvent := false
	for _, event := range c.Events {
		if c.isSameDay(event.Date, date) {
			hasEvent = true
			break
		}
	}
	if hasEvent {
		indicator = "◆"
	}

	content := fmt.Sprintf("%s%s", dayStr, indicator)
	return layout.PadCenter(style.Render(content), width)
}

// renderDayPanel renders the right panel showing selected day details.
func (c *Calendar) renderDayPanel(width, height int) string {
	frame := layout.NewFrame(width, height)

	dateStr := c.SelectedDate.Format("Mon, Jan 2")
	frame.SetTitle(dateStr)
	frame.SetBorder(layout.BorderRounded)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	contentWidth := frame.ContentWidth()
	contentHeight := frame.ContentHeight()

	var lines []string

	// Filter events for selected date
	var dayEvents []CalendarEvent
	for _, event := range c.Events {
		if c.isSameDay(event.Date, c.SelectedDate) {
			dayEvents = append(dayEvents, event)
		}
	}

	if len(dayEvents) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_muted"))
		lines = append(lines, layout.FitToWidth(emptyStyle.Render(" No events"), contentWidth))
	} else {
		for i, event := range dayEvents {
			if i >= contentHeight-1 {
				more := fmt.Sprintf(" ... and %d more", len(dayEvents)-i)
				lines = append(lines, layout.FitToWidth(more, contentWidth))
				break
			}

			var icon string
			var style lipgloss.Style

			switch event.Type {
			case "task":
				if event.Done {
					icon = icons.Get("task_done")
					style = theme.S.TaskDone
				} else {
					icon = icons.Get("task_open")
					style = theme.S.TaskOpen
				}
			case "goal":
				icon = icons.Get("goals")
				style = theme.S.GoalTitle
			case "journal":
				icon = icons.Get("note")
				style = theme.S.TextPrimary
			default:
				icon = icons.Get("bullet")
				style = theme.S.TextSecondary
			}

			title := event.Title
			maxLen := contentWidth - 4
			if lipgloss.Width(title) > maxLen {
				title = layout.TruncateWithEllipsis(title, maxLen)
			}

			line := fmt.Sprintf(" %s %s", icon, style.Render(title))
			lines = append(lines, layout.FitToWidth(line, contentWidth))
		}
	}

	// Activity summary
	dateKey := c.SelectedDate.Format("2006-01-02")
	if activity, ok := c.ActivityData[dateKey]; ok && activity > 0 {
		activityStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("primary"))
		activityLine := fmt.Sprintf(" %s %d pomodoros", icons.Get("pomodoro_work"), activity)
		lines = append(lines, "")
		lines = append(lines, layout.FitToWidth(activityStyle.Render(activityLine), contentWidth))
	}

	frame.SetContentLines(lines)
	return frame.Render()
}

// renderYearView renders the year overview with 12 mini-months.
func (c *Calendar) renderYearView() string {
	frame := layout.NewFrame(c.Width, c.Height)

	yearStr := fmt.Sprintf("%d", c.CurrentDate.Year())
	frame.SetTitle(yearStr)
	frame.SetBorder(layout.BorderRounded)
	frame.SetFocused(c.focused)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	contentWidth := frame.ContentWidth()
	contentHeight := frame.ContentHeight()

	// Layout: 4 columns x 3 rows of mini-months
	monthWidth := contentWidth / 4
	monthHeight := contentHeight / 3

	var rows []string
	for row := 0; row < 3; row++ {
		var monthsInRow []string
		for col := 0; col < 4; col++ {
			monthNum := row*4 + col + 1
			miniMonth := c.renderMiniMonth(monthNum, monthWidth, monthHeight)
			monthsInRow = append(monthsInRow, miniMonth)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, monthsInRow...))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	frame.SetContent(content)
	return frame.Render()
}

// renderMiniMonth renders a miniature month view.
func (c *Calendar) renderMiniMonth(month int, width, height int) string {
	var lines []string

	monthDate := time.Date(c.CurrentDate.Year(), time.Month(month), 1, 0, 0, 0, 0, c.CurrentDate.Location())
	monthName := monthDate.Format("Jan")

	// Header
	headerStyle := theme.S.CalendarHeader
	header := layout.PadCenter(headerStyle.Render(monthName), width)
	lines = append(lines, header)

	// Days (simplified - just show activity dots)
	daysInMonth := daysInMonth(c.CurrentDate.Year(), time.Month(month))

	dayStyle := theme.S.CalendarDayMuted
	todayStyle := theme.S.CalendarDayToday

	var dayLine strings.Builder
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(c.CurrentDate.Year(), time.Month(month), day, 0, 0, 0, 0, c.CurrentDate.Location())
		dateKey := date.Format("2006-01-02")

		char := "·"
		style := dayStyle

		if c.isSameDay(date, c.Today) {
			char = "●"
			style = todayStyle
		} else if activity, ok := c.ActivityData[dateKey]; ok && activity > 0 {
			if activity >= 3 {
				char = "●"
			} else {
				char = "○"
			}
		}

		dayLine.WriteString(style.Render(char))

		// Wrap at week end
		if day%7 == 0 && day < daysInMonth {
			lines = append(lines, layout.FitToWidth(dayLine.String(), width))
			dayLine.Reset()
		}
	}

	if dayLine.Len() > 0 {
		lines = append(lines, layout.FitToWidth(dayLine.String(), width))
	}

	// Pad to height
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}

	return strings.Join(lines[:height], "\n")
}

// renderDayView renders the detailed day view.
func (c *Calendar) renderDayView() string {
	frame := layout.NewFrame(c.Width, c.Height)

	dateStr := c.SelectedDate.Format("Monday, January 2, 2006")
	frame.SetTitle(dateStr)
	frame.SetBorder(layout.BorderRounded)
	frame.SetFocused(c.focused)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	contentWidth := frame.ContentWidth()
	contentHeight := frame.ContentHeight()

	var lines []string

	// Layer sections
	layers := []struct {
		name  string
		layer CalendarLayer
		icon  string
	}{
		{"Tasks", LayerTasks, icons.Get("task_open")},
		{"Goals", LayerGoals, icons.Get("goals")},
		{"Journal", LayerJournal, icons.Get("note")},
		{"Activity", LayerActivity, icons.Get("pomodoro_work")},
	}

	for _, l := range layers {
		if c.ActiveLayers&l.layer == 0 {
			continue
		}

		// Section header
		headerStyle := lipgloss.NewStyle().
			Foreground(theme.Current.Color("text_primary")).
			Bold(true)
		header := fmt.Sprintf(" %s %s", l.icon, l.name)
		lines = append(lines, headerStyle.Render(header))

		// Section content
		events := c.filterEventsByType(l.name)
		if len(events) == 0 {
			emptyStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_muted"))
			lines = append(lines, layout.FitToWidth(emptyStyle.Render("   No items"), contentWidth))
		} else {
			for _, event := range events {
				if len(lines) >= contentHeight-2 {
					break
				}
				line := c.renderEventLine(event, contentWidth-4)
				lines = append(lines, "   "+line)
			}
		}

		lines = append(lines, "")
	}

	frame.SetContentLines(lines)
	return frame.Render()
}

// filterEventsByType filters events for the selected date and type.
func (c *Calendar) filterEventsByType(layerName string) []CalendarEvent {
	var result []CalendarEvent
	typeMap := map[string]string{
		"Tasks":    "task",
		"Goals":    "goal",
		"Journal":  "journal",
		"Activity": "activity",
	}

	targetType := typeMap[layerName]

	for _, event := range c.Events {
		if c.isSameDay(event.Date, c.SelectedDate) && event.Type == targetType {
			result = append(result, event)
		}
	}

	return result
}

// renderEventLine renders a single event line.
func (c *Calendar) renderEventLine(event CalendarEvent, width int) string {
	var icon string
	var style lipgloss.Style

	switch event.Type {
	case "task":
		if event.Done {
			icon = icons.Get("task_done")
			style = theme.S.TaskDone
		} else {
			icon = icons.Get("task_open")
			style = theme.S.TaskOpen
		}
	case "goal":
		icon = icons.Get("goals")
		style = theme.S.GoalTitle
	default:
		icon = icons.Get("bullet")
		style = theme.S.TextSecondary
	}

	title := event.Title
	maxLen := width - 3
	if lipgloss.Width(title) > maxLen {
		title = layout.TruncateWithEllipsis(title, maxLen)
	}

	return fmt.Sprintf("%s %s", icon, style.Render(title))
}

// Navigation methods

// NextMonth moves to the next month.
func (c *Calendar) NextMonth() {
	c.CurrentDate = c.CurrentDate.AddDate(0, 1, 0)
}

// PrevMonth moves to the previous month.
func (c *Calendar) PrevMonth() {
	c.CurrentDate = c.CurrentDate.AddDate(0, -1, 0)
}

// NextYear moves to the next year.
func (c *Calendar) NextYear() {
	c.CurrentDate = c.CurrentDate.AddDate(1, 0, 0)
}

// PrevYear moves to the previous year.
func (c *Calendar) PrevYear() {
	c.CurrentDate = c.CurrentDate.AddDate(-1, 0, 0)
}

// SelectNext moves selection to the next day.
func (c *Calendar) SelectNext() {
	c.SelectedDate = c.SelectedDate.AddDate(0, 0, 1)
	// Auto-advance month if needed
	if c.SelectedDate.Month() != c.CurrentDate.Month() {
		c.CurrentDate = c.SelectedDate
	}
}

// SelectPrev moves selection to the previous day.
func (c *Calendar) SelectPrev() {
	c.SelectedDate = c.SelectedDate.AddDate(0, 0, -1)
	if c.SelectedDate.Month() != c.CurrentDate.Month() {
		c.CurrentDate = c.SelectedDate
	}
}

// SelectNextWeek moves selection to the next week.
func (c *Calendar) SelectNextWeek() {
	c.SelectedDate = c.SelectedDate.AddDate(0, 0, 7)
	if c.SelectedDate.Month() != c.CurrentDate.Month() {
		c.CurrentDate = c.SelectedDate
	}
}

// SelectPrevWeek moves selection to the previous week.
func (c *Calendar) SelectPrevWeek() {
	c.SelectedDate = c.SelectedDate.AddDate(0, 0, -7)
	if c.SelectedDate.Month() != c.CurrentDate.Month() {
		c.CurrentDate = c.SelectedDate
	}
}

// GoToToday jumps to today's date.
func (c *Calendar) GoToToday() {
	c.CurrentDate = c.Today
	c.SelectedDate = c.Today
}

// SetMode changes the calendar view mode.
func (c *Calendar) SetMode(mode CalendarViewMode) {
	c.Mode = mode
}

// ToggleLayer toggles a calendar layer.
func (c *Calendar) ToggleLayer(layer CalendarLayer) {
	c.ActiveLayers ^= layer
}

// Helper methods

func (c *Calendar) isSameDay(a, b time.Time) bool {
	return a.Year() == b.Year() && a.YearDay() == b.YearDay()
}

func (c *Calendar) getWeekdayNames() []string {
	if c.FirstDayOfWeek == 1 {
		return []string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
	}
	return []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// ConvertTasksToEvents converts types.Task slice to CalendarEvent slice.
func ConvertTasksToEvents(tasks []types.Task) []CalendarEvent {
	var events []CalendarEvent
	for _, task := range tasks {
		events = append(events, CalendarEvent{
			Title: task.Text,
			Type:  "task",
			Done:  task.Status == "done",
		})
	}
	return events
}
