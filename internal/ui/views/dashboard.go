// Package views contains the different views/pages of the application.
package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/ui/components"
	"github.com/BioWare/lazyobsidian/internal/ui/icons"
	"github.com/BioWare/lazyobsidian/internal/ui/layout"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// Dashboard represents the dashboard view.
type Dashboard struct {
	Width  int
	Height int

	// Data
	Tasks         []types.Task
	ActiveCourses []types.Course
	CurrentBook   *types.Book
	PomodoroState PomodoroState
	WeeklyStats   WeeklyStats
	RecentNotes   []RecentNote

	// Focus state
	FocusedModule int
	SelectedTask  int
}

// PomodoroState holds current pomodoro state for display.
type PomodoroState struct {
	State     string // "ready", "running", "paused", "break"
	Remaining time.Duration
	DailyGoal int
	DailyDone int
	Context   string
}

// WeeklyStats holds weekly statistics.
type WeeklyStats struct {
	Pomodoros int
	FocusTime time.Duration
	Streak    int
	ByDay     [7]int // Mon-Sun
}

// RecentNote represents a recently modified note.
type RecentNote struct {
	Title    string
	Path     string
	Modified time.Time
}

// DashboardModule constants
const (
	ModuleTodayFocus = iota
	ModulePomodoro
	ModuleWeekStats
	ModuleCourses
	ModuleBook
	ModuleRecentNotes
)

// NewDashboard creates a new dashboard view.
func NewDashboard(width, height int) *Dashboard {
	return &Dashboard{
		Width:  width,
		Height: height,
	}
}

// Render renders the dashboard using the Frame-based layout system.
func (d *Dashboard) Render() string {
	if d.Width <= 0 || d.Height <= 0 {
		return ""
	}

	// Calculate section heights
	// Today's Focus: ~35% of height (min 8 lines)
	// Pomodoro/Week: ~25% of height (min 7 lines)
	// Courses/Book: ~20% of height (min 6 lines)
	// Recent Notes: remaining space

	todayHeight := max(8, d.Height*35/100)
	middleHeight := max(7, d.Height*25/100)
	bottomHeight := max(6, d.Height*20/100)
	recentHeight := d.Height - todayHeight - middleHeight - bottomHeight

	if recentHeight < 4 {
		// Redistribute if Recent Notes is too small
		recentHeight = 4
		todayHeight = d.Height - middleHeight - bottomHeight - recentHeight
	}

	var sections []string

	// Section 1: Today's Focus (full width)
	todayFocus := d.renderTodayFocus(d.Width, todayHeight)
	sections = append(sections, todayFocus)

	// Section 2: Pomodoro + This Week (side by side)
	leftWidth := d.Width * 40 / 100
	rightWidth := d.Width - leftWidth

	pomodoro := d.renderPomodoro(leftWidth, middleHeight)
	weekStats := d.renderWeekStats(rightWidth, middleHeight)
	middleRow := lipgloss.JoinHorizontal(lipgloss.Top, pomodoro, weekStats)
	sections = append(sections, middleRow)

	// Section 3: Active Courses + Current Book (side by side)
	courses := d.renderCourses(leftWidth, bottomHeight)
	book := d.renderBook(rightWidth, bottomHeight)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, courses, book)
	sections = append(sections, bottomRow)

	// Section 4: Recent Notes (full width)
	recent := d.renderRecentNotes(d.Width, recentHeight)
	sections = append(sections, recent)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// renderTodayFocus renders the Today's Focus panel.
func (d *Dashboard) renderTodayFocus(width, height int) string {
	frame := layout.NewFrame(width, height)
	frame.SetTitle("Today's Focus")
	frame.SetBorder(layout.BorderRounded)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	if d.FocusedModule == ModuleTodayFocus {
		frame.SetFocused(true)
	}

	contentWidth := frame.ContentWidth()
	contentHeight := frame.ContentHeight()

	var lines []string

	if len(d.Tasks) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_muted"))
		emptyMsg := emptyStyle.Render("No daily note. Press Enter to create.")
		lines = append(lines, layout.PadCenter(emptyMsg, contentWidth))
	} else {
		// Render tasks
		maxTasks := contentHeight - 2 // Leave room for progress bar
		for i, task := range d.Tasks {
			if i >= maxTasks {
				break
			}

			line := d.renderTaskLine(task, i == d.SelectedTask && d.FocusedModule == ModuleTodayFocus, contentWidth)
			lines = append(lines, line)
		}

		// Pad remaining lines
		for len(lines) < maxTasks {
			lines = append(lines, strings.Repeat(" ", contentWidth))
		}

		// Progress bar
		completed, total := countTaskProgress(d.Tasks)
		progress := float64(0)
		if total > 0 {
			progress = float64(completed) / float64(total)
		}
		progressBar := components.NewProgressBar(contentWidth-20, progress)
		progressLine := fmt.Sprintf("Progress: %s %d/%d", progressBar.Render(), completed, total)
		lines = append(lines, layout.FitToWidth(progressLine, contentWidth))
	}

	frame.SetContentLines(lines)
	return frame.Render()
}

// renderTaskLine renders a single task line.
func (d *Dashboard) renderTaskLine(task types.Task, selected bool, width int) string {
	// Get icon for status
	var icon string
	var style lipgloss.Style

	switch task.Status {
	case "done":
		icon = icons.Get("task_done")
		style = theme.S.TaskDone
	case "cancelled":
		icon = icons.Get("task_cancelled")
		style = theme.S.TaskCancelled
	case "in_progress":
		icon = icons.Get("task_in_progress")
		style = theme.S.TaskInProgress
	case "deferred":
		icon = icons.Get("task_deferred")
		style = theme.S.TaskDeferred
	case "question":
		icon = icons.Get("task_question")
		style = theme.S.TaskQuestion
	default:
		icon = icons.Get("task_open")
		style = theme.S.TaskOpen
	}

	// Build task text
	text := task.Text
	iconWidth := lipgloss.Width(icon) + 1 // icon + space
	maxTextWidth := width - iconWidth - 2  // padding

	if lipgloss.Width(text) > maxTextWidth {
		text = layout.TruncateWithEllipsis(text, maxTextWidth)
	}

	content := icon + " " + style.Render(text)

	if selected {
		// Highlight selected task
		bgStyle := lipgloss.NewStyle().
			Background(theme.Current.Color("bg_active")).
			Bold(true)
		content = bgStyle.Render(layout.FitToWidth(content, width))
	}

	return layout.FitToWidth(content, width)
}

// renderPomodoro renders the Pomodoro panel.
func (d *Dashboard) renderPomodoro(width, height int) string {
	frame := layout.NewFrame(width, height)
	frame.SetTitle("Pomodoro")
	frame.SetBorder(layout.BorderRounded)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	if d.FocusedModule == ModulePomodoro {
		frame.SetFocused(true)
	}

	contentWidth := frame.ContentWidth()
	var lines []string

	// Timer display
	timerStyle := lipgloss.NewStyle().Bold(true)
	switch d.PomodoroState.State {
	case "running":
		timerStyle = timerStyle.Foreground(theme.Current.Color("pomodoro_work"))
	case "break":
		timerStyle = timerStyle.Foreground(theme.Current.Color("pomodoro_break"))
	case "paused":
		timerStyle = timerStyle.Foreground(theme.Current.Color("pomodoro_paused"))
	default:
		timerStyle = timerStyle.Foreground(theme.Current.Color("text_primary"))
	}

	minutes := int(d.PomodoroState.Remaining.Minutes())
	seconds := int(d.PomodoroState.Remaining.Seconds()) % 60
	timerText := fmt.Sprintf("%02d:%02d", minutes, seconds)

	stateStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_secondary"))
	stateText := "[" + capitalizeFirst(d.PomodoroState.State) + "]"

	pomIcon := icons.Get("pomodoro_work")
	timerLine := fmt.Sprintf(" %s  %s  %s", pomIcon, timerStyle.Render(timerText), stateStyle.Render(stateText))
	lines = append(lines, layout.FitToWidth(timerLine, contentWidth))

	// Daily goal progress
	goalStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_secondary"))
	goalText := fmt.Sprintf(" Daily: %d/%d", d.PomodoroState.DailyDone, d.PomodoroState.DailyGoal)
	lines = append(lines, layout.FitToWidth(goalStyle.Render(goalText), contentWidth))

	// Context if set
	if d.PomodoroState.Context != "" {
		ctxStyle := lipgloss.NewStyle().
			Foreground(theme.Current.Color("text_muted")).
			Italic(true)
		ctxText := " " + d.PomodoroState.Context
		lines = append(lines, layout.FitToWidth(ctxStyle.Render(ctxText), contentWidth))
	}

	frame.SetContentLines(lines)
	return frame.Render()
}

// renderWeekStats renders the This Week stats panel.
func (d *Dashboard) renderWeekStats(width, height int) string {
	frame := layout.NewFrame(width, height)
	frame.SetTitle("This Week")
	frame.SetBorder(layout.BorderRounded)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	if d.FocusedModule == ModuleWeekStats {
		frame.SetFocused(true)
	}

	contentWidth := frame.ContentWidth()
	var lines []string

	// Day labels
	labelStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_secondary"))
	lines = append(lines, layout.FitToWidth(labelStyle.Render(" Mo Tu We Th Fr Sa Su"), contentWidth))

	// Activity heatmap
	var bars []string
	for _, count := range d.WeeklyStats.ByDay {
		bar := "░░"
		barStyle := theme.S.HeatmapLevel0
		if count >= 5 {
			bar = "██"
			barStyle = theme.S.HeatmapLevel4
		} else if count >= 3 {
			bar = "▓▓"
			barStyle = theme.S.HeatmapLevel3
		} else if count >= 1 {
			bar = "▒▒"
			barStyle = theme.S.HeatmapLevel1
		}
		bars = append(bars, barStyle.Render(bar))
	}
	heatmapLine := " " + strings.Join(bars, " ")
	lines = append(lines, layout.FitToWidth(heatmapLine, contentWidth))

	// Summary
	focusHours := d.WeeklyStats.FocusTime.Hours()
	summaryStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_primary"))
	summaryText := fmt.Sprintf(" %d pom • %.1fh", d.WeeklyStats.Pomodoros, focusHours)
	if d.WeeklyStats.Streak > 0 {
		streakIcon := icons.Get("fire")
		summaryText += fmt.Sprintf(" • %s%d", streakIcon, d.WeeklyStats.Streak)
	}
	lines = append(lines, layout.FitToWidth(summaryStyle.Render(summaryText), contentWidth))

	frame.SetContentLines(lines)
	return frame.Render()
}

// renderCourses renders the Active Courses panel.
func (d *Dashboard) renderCourses(width, height int) string {
	frame := layout.NewFrame(width, height)
	frame.SetTitle("Active Courses")
	frame.SetBorder(layout.BorderRounded)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	if d.FocusedModule == ModuleCourses {
		frame.SetFocused(true)
	}

	contentWidth := frame.ContentWidth()
	contentHeight := frame.ContentHeight()
	var lines []string

	if len(d.ActiveCourses) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_muted"))
		lines = append(lines, layout.FitToWidth(emptyStyle.Render(" No active courses"), contentWidth))
	} else {
		for i, course := range d.ActiveCourses {
			if i >= contentHeight {
				break
			}

			progress := float64(0)
			if course.TotalLessons > 0 {
				progress = float64(course.Completed) / float64(course.TotalLessons)
			}
			progressBar := components.NewProgressBar(10, progress)
			progressBar.ShowLabel = false

			title := course.Title
			progressWidth := lipgloss.Width(progressBar.Render()) + 5 // " XX%"
			maxTitleLen := contentWidth - progressWidth - 2

			if lipgloss.Width(title) > maxTitleLen {
				title = layout.TruncateWithEllipsis(title, maxTitleLen)
			}

			courseIcon := icons.Get("course")
			line := fmt.Sprintf(" %s %s %s %d%%",
				courseIcon,
				layout.PadRight(title, maxTitleLen),
				progressBar.Render(),
				int(progress*100))
			lines = append(lines, layout.FitToWidth(line, contentWidth))
		}
	}

	frame.SetContentLines(lines)
	return frame.Render()
}

// renderBook renders the Current Book panel.
func (d *Dashboard) renderBook(width, height int) string {
	frame := layout.NewFrame(width, height)
	frame.SetTitle("Current Book")
	frame.SetBorder(layout.BorderRounded)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	if d.FocusedModule == ModuleBook {
		frame.SetFocused(true)
	}

	contentWidth := frame.ContentWidth()
	var lines []string

	if d.CurrentBook == nil {
		emptyStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_muted"))
		lines = append(lines, layout.FitToWidth(emptyStyle.Render(" No book in progress"), contentWidth))
	} else {
		progress := float64(0)
		if d.CurrentBook.TotalPages > 0 {
			progress = float64(d.CurrentBook.CurrentPage) / float64(d.CurrentBook.TotalPages)
		}
		progressBar := components.NewProgressBar(12, progress)
		progressBar.ShowLabel = false

		// Book title
		bookIcon := icons.Get("book")
		title := d.CurrentBook.Title
		maxTitleLen := contentWidth - 4
		if lipgloss.Width(title) > maxTitleLen {
			title = layout.TruncateWithEllipsis(title, maxTitleLen)
		}
		titleStyle := theme.S.BookTitle
		lines = append(lines, layout.FitToWidth(fmt.Sprintf(" %s %s", bookIcon, titleStyle.Render(title)), contentWidth))

		// Author if present
		if d.CurrentBook.Author != "" {
			authorStyle := theme.S.BookAuthor
			author := "   by " + d.CurrentBook.Author
			if lipgloss.Width(author) > contentWidth {
				author = layout.TruncateWithEllipsis(author, contentWidth)
			}
			lines = append(lines, layout.FitToWidth(authorStyle.Render(author), contentWidth))
		}

		// Progress
		progressLine := fmt.Sprintf(" %s %d%% • p.%d/%d",
			progressBar.Render(),
			int(progress*100),
			d.CurrentBook.CurrentPage,
			d.CurrentBook.TotalPages)
		lines = append(lines, layout.FitToWidth(progressLine, contentWidth))
	}

	frame.SetContentLines(lines)
	return frame.Render()
}

// renderRecentNotes renders the Recent Notes panel.
func (d *Dashboard) renderRecentNotes(width, height int) string {
	frame := layout.NewFrame(width, height)
	frame.SetTitle("Recent Notes")
	frame.SetBorder(layout.BorderRounded)

	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}

	if d.FocusedModule == ModuleRecentNotes {
		frame.SetFocused(true)
	}

	contentWidth := frame.ContentWidth()
	contentHeight := frame.ContentHeight()
	var lines []string

	if len(d.RecentNotes) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_muted"))
		lines = append(lines, layout.FitToWidth(emptyStyle.Render(" No recent notes"), contentWidth))
	} else {
		for i, note := range d.RecentNotes {
			if i >= contentHeight {
				break
			}

			timeAgo := formatTimeAgo(note.Modified)
			timeStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_muted"))

			noteIcon := icons.Get("note")
			title := note.Title
			timeWidth := lipgloss.Width(timeAgo) + 3
			maxTitleLen := contentWidth - timeWidth - 4

			if lipgloss.Width(title) > maxTitleLen {
				title = layout.TruncateWithEllipsis(title, maxTitleLen)
			}

			// Pad title to align time
			titlePadded := layout.PadRight(title, maxTitleLen)
			line := fmt.Sprintf(" %s %s  %s", noteIcon, titlePadded, timeStyle.Render(timeAgo))
			lines = append(lines, layout.FitToWidth(line, contentWidth))
		}
	}

	frame.SetContentLines(lines)
	return frame.Render()
}

// Helper functions

func countTaskProgress(tasks []types.Task) (completed, total int) {
	for _, t := range tasks {
		total++
		if t.Status == "done" {
			completed++
		}
	}
	return
}

func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)
	if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
}

func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// truncateToWidth truncates a string to fit within maxWidth.
// Delegates to layout.TruncateToWidth for consistent handling.
func truncateToWidth(s string, maxWidth int) string {
	return layout.TruncateToWidth(s, maxWidth)
}
