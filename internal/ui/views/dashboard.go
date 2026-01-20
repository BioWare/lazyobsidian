// Package views contains the different views/pages of the application.
package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/ui/components"
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// Dashboard represents the dashboard view.
type Dashboard struct {
	Width  int
	Height int

	// Data
	Tasks          []types.Task
	ActiveCourses  []types.Course
	CurrentBook    *types.Book
	PomodoroState  PomodoroState
	WeeklyStats    WeeklyStats
	RecentNotes    []RecentNote

	// Focus state
	FocusedModule int
}

// PomodoroState holds current pomodoro state for display.
type PomodoroState struct {
	State       string // "ready", "running", "paused", "break"
	Remaining   time.Duration
	DailyGoal   int
	DailyDone   int
	Context     string
}

// WeeklyStats holds weekly statistics.
type WeeklyStats struct {
	Pomodoros    int
	FocusTime    time.Duration
	Streak       int
	ByDay        [7]int // Mon-Sun
}

// RecentNote represents a recently modified note.
type RecentNote struct {
	Title    string
	Path     string
	Modified time.Time
}

// NewDashboard creates a new dashboard view.
func NewDashboard(width, height int) *Dashboard {
	return &Dashboard{
		Width:  width,
		Height: height,
	}
}

// Render renders the dashboard.
func (d *Dashboard) Render() string {
	// Calculate module dimensions
	leftWidth := d.Width * 55 / 100
	rightWidth := d.Width - leftWidth - 3

	// Top section: Today's Focus
	todayFocus := d.renderTodayFocus(d.Width-4, 10)

	// Middle section: Pomodoro + This Week
	pomodoroWidth := leftWidth
	weekWidth := rightWidth
	pomodoro := d.renderPomodoro(pomodoroWidth, 6)
	week := d.renderWeekStats(weekWidth, 6)
	middleRow := lipgloss.JoinHorizontal(lipgloss.Top, pomodoro, "  ", week)

	// Bottom section: Active Courses + Current Book
	courses := d.renderActiveCourses(leftWidth, 5)
	book := d.renderCurrentBook(rightWidth, 5)
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, courses, "  ", book)

	// Recent Notes
	recentNotes := d.renderRecentNotes(d.Width-4, 4)

	// Join all sections
	content := lipgloss.JoinVertical(lipgloss.Left,
		todayFocus,
		"",
		middleRow,
		"",
		bottomRow,
		"",
		recentNotes,
	)

	return content
}

func (d *Dashboard) renderTodayFocus(width, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3D3428"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D4C9B5"))

	var lines []string
	lines = append(lines, borderStyle.Render("┌─ ")+titleStyle.Render("Today's Focus")+" "+borderStyle.Render(strings.Repeat("─", width-18)+"┐"))

	if len(d.Tasks) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9C8B75")).
			Render("No daily note. Press Enter to create.")
		lines = append(lines, borderStyle.Render("│ ")+emptyMsg+strings.Repeat(" ", width-len(emptyMsg)-4)+borderStyle.Render(" │"))
	} else {
		taskList := components.NewTaskList(d.Tasks, width-4, height-4)
		taskContent := taskList.Render()
		for _, line := range strings.Split(taskContent, "\n") {
			paddedLine := line + strings.Repeat(" ", max(0, width-len(line)-4))
			lines = append(lines, borderStyle.Render("│ ")+paddedLine+borderStyle.Render(" │"))
		}

		// Progress bar
		completed, total := taskList.Progress()
		progress := float64(0)
		if total > 0 {
			progress = float64(completed) / float64(total)
		}
		progressBar := components.NewProgressBar(width-20, progress)
		progressLine := fmt.Sprintf("Progress: %s %d/%d", progressBar.Render(), completed, total)
		lines = append(lines, borderStyle.Render("│ ")+progressLine+strings.Repeat(" ", max(0, width-len(progressLine)-4))+borderStyle.Render(" │"))
	}

	// Pad remaining height
	for len(lines) < height-1 {
		lines = append(lines, borderStyle.Render("│")+strings.Repeat(" ", width-2)+borderStyle.Render("│"))
	}

	lines = append(lines, borderStyle.Render("└"+strings.Repeat("─", width-2)+"┘"))

	return strings.Join(lines, "\n")
}

func (d *Dashboard) renderPomodoro(width, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D4C9B5"))

	var lines []string
	lines = append(lines, borderStyle.Render("┌─ ")+titleStyle.Render("Pomodoro")+" "+borderStyle.Render(strings.Repeat("─", width-14)+"┐"))

	// Timer display
	timerStyle := lipgloss.NewStyle().Bold(true)
	if d.PomodoroState.State == "running" {
		timerStyle = timerStyle.Foreground(lipgloss.Color("#8B3A3A"))
	} else if d.PomodoroState.State == "break" {
		timerStyle = timerStyle.Foreground(lipgloss.Color("#4A7C59"))
	}

	minutes := int(d.PomodoroState.Remaining.Minutes())
	seconds := int(d.PomodoroState.Remaining.Seconds()) % 60
	timerText := fmt.Sprintf("%02d:%02d", minutes, seconds)

	stateText := "[" + strings.Title(d.PomodoroState.State) + "]"
	timerLine := fmt.Sprintf("      %s  %s", timerStyle.Render(timerText), stateText)
	lines = append(lines, borderStyle.Render("│ ")+timerLine+strings.Repeat(" ", max(0, width-len(timerLine)-4))+borderStyle.Render(" │"))

	// Daily goal
	goalText := fmt.Sprintf("Daily: %d/%d 🍅", d.PomodoroState.DailyDone, d.PomodoroState.DailyGoal)
	lines = append(lines, borderStyle.Render("│ ")+goalText+strings.Repeat(" ", max(0, width-len(goalText)-4))+borderStyle.Render(" │"))

	// Pad and close
	for len(lines) < height-1 {
		lines = append(lines, borderStyle.Render("│")+strings.Repeat(" ", width-2)+borderStyle.Render("│"))
	}
	lines = append(lines, borderStyle.Render("└"+strings.Repeat("─", width-2)+"┘"))

	return strings.Join(lines, "\n")
}

func (d *Dashboard) renderWeekStats(width, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D4C9B5"))

	var lines []string
	lines = append(lines, borderStyle.Render("┌─ ")+titleStyle.Render("This Week")+" "+borderStyle.Render(strings.Repeat("─", width-15)+"┐"))

	// Day labels
	dayLabels := "Mo Tu We Th Fr Sa Su"
	lines = append(lines, borderStyle.Render("│ ")+dayLabels+strings.Repeat(" ", max(0, width-len(dayLabels)-4))+borderStyle.Render(" │"))

	// Activity bars
	var bars []string
	for _, count := range d.WeeklyStats.ByDay {
		bar := "░░"
		if count >= 5 {
			bar = "██"
		} else if count >= 3 {
			bar = "▓▓"
		} else if count >= 1 {
			bar = "▒▒"
		}
		bars = append(bars, bar)
	}
	barsLine := strings.Join(bars, " ")
	lines = append(lines, borderStyle.Render("│ ")+barsLine+strings.Repeat(" ", max(0, width-len(barsLine)-4))+borderStyle.Render(" │"))

	// Summary
	focusHours := d.WeeklyStats.FocusTime.Hours()
	summaryText := fmt.Sprintf("%d🍅 • %.1fh • 🔥%d", d.WeeklyStats.Pomodoros, focusHours, d.WeeklyStats.Streak)
	lines = append(lines, borderStyle.Render("│ ")+summaryText+strings.Repeat(" ", max(0, width-len(summaryText)-4))+borderStyle.Render(" │"))

	// Pad and close
	for len(lines) < height-1 {
		lines = append(lines, borderStyle.Render("│")+strings.Repeat(" ", width-2)+borderStyle.Render("│"))
	}
	lines = append(lines, borderStyle.Render("└"+strings.Repeat("─", width-2)+"┘"))

	return strings.Join(lines, "\n")
}

func (d *Dashboard) renderActiveCourses(width, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D4C9B5"))

	var lines []string
	lines = append(lines, borderStyle.Render("┌─ ")+titleStyle.Render("Active Courses")+" "+borderStyle.Render(strings.Repeat("─", width-19)+"┐"))

	if len(d.ActiveCourses) == 0 {
		lines = append(lines, borderStyle.Render("│ ")+"No active courses"+strings.Repeat(" ", max(0, width-21))+borderStyle.Render(" │"))
	} else {
		for i, course := range d.ActiveCourses {
			if i >= height-3 {
				break
			}
			progress := float64(course.Completed) / float64(course.TotalLessons)
			progressBar := components.NewProgressBar(10, progress)
			progressBar.ShowLabel = false

			title := course.Title
			if len(title) > width-20 {
				title = title[:width-23] + "..."
			}

			line := fmt.Sprintf("%s %s %d%%", title, progressBar.Render(), int(progress*100))
			lines = append(lines, borderStyle.Render("│ ")+line+strings.Repeat(" ", max(0, width-len(line)-4))+borderStyle.Render(" │"))
		}
	}

	// Pad and close
	for len(lines) < height-1 {
		lines = append(lines, borderStyle.Render("│")+strings.Repeat(" ", width-2)+borderStyle.Render("│"))
	}
	lines = append(lines, borderStyle.Render("└"+strings.Repeat("─", width-2)+"┘"))

	return strings.Join(lines, "\n")
}

func (d *Dashboard) renderCurrentBook(width, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D4C9B5"))

	var lines []string
	lines = append(lines, borderStyle.Render("┌─ ")+titleStyle.Render("Current Book")+" "+borderStyle.Render(strings.Repeat("─", width-17)+"┐"))

	if d.CurrentBook == nil {
		lines = append(lines, borderStyle.Render("│ ")+"No book in progress"+strings.Repeat(" ", max(0, width-23))+borderStyle.Render(" │"))
	} else {
		progress := float64(d.CurrentBook.CurrentPage) / float64(d.CurrentBook.TotalPages)
		progressBar := components.NewProgressBar(10, progress)
		progressBar.ShowLabel = false

		title := d.CurrentBook.Title
		if len(title) > width-6 {
			title = title[:width-9] + "..."
		}

		line1 := fmt.Sprintf("%s %s %d%%", title, progressBar.Render(), int(progress*100))
		lines = append(lines, borderStyle.Render("│ ")+line1+strings.Repeat(" ", max(0, width-len(line1)-4))+borderStyle.Render(" │"))

		line2 := fmt.Sprintf("p.%d/%d", d.CurrentBook.CurrentPage, d.CurrentBook.TotalPages)
		lines = append(lines, borderStyle.Render("│ ")+line2+strings.Repeat(" ", max(0, width-len(line2)-4))+borderStyle.Render(" │"))
	}

	// Pad and close
	for len(lines) < height-1 {
		lines = append(lines, borderStyle.Render("│")+strings.Repeat(" ", width-2)+borderStyle.Render("│"))
	}
	lines = append(lines, borderStyle.Render("└"+strings.Repeat("─", width-2)+"┘"))

	return strings.Join(lines, "\n")
}

func (d *Dashboard) renderRecentNotes(width, height int) string {
	titleStyle := lipgloss.NewStyle().Bold(true)
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D4C9B5"))

	var lines []string
	lines = append(lines, borderStyle.Render("┌─ ")+titleStyle.Render("Recent Notes")+" "+borderStyle.Render(strings.Repeat("─", width-17)+"┐"))

	if len(d.RecentNotes) == 0 {
		lines = append(lines, borderStyle.Render("│ ")+"No recent notes"+strings.Repeat(" ", max(0, width-19))+borderStyle.Render(" │"))
	} else {
		for i, note := range d.RecentNotes {
			if i >= height-2 {
				break
			}
			timeAgo := formatTimeAgo(note.Modified)
			title := note.Title
			maxTitleLen := width - len(timeAgo) - 8
			if len(title) > maxTitleLen {
				title = title[:maxTitleLen-3] + "..."
			}
			line := title + strings.Repeat(" ", max(0, maxTitleLen-len(title))) + "  " + timeAgo
			lines = append(lines, borderStyle.Render("│ ")+line+borderStyle.Render(" │"))
		}
	}

	// Pad and close
	for len(lines) < height-1 {
		lines = append(lines, borderStyle.Render("│")+strings.Repeat(" ", width-2)+borderStyle.Render("│"))
	}
	lines = append(lines, borderStyle.Render("└"+strings.Repeat("─", width-2)+"┘"))

	return strings.Join(lines, "\n")
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
