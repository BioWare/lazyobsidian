package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/pkg/types"
)

// TaskList renders a list of tasks.
type TaskList struct {
	Tasks    []types.Task
	Selected int
	Width    int
	Height   int
	Icons    map[string]string
	Styles   TaskStyles
}

// TaskStyles holds styles for different task states.
type TaskStyles struct {
	Open       lipgloss.Style
	Done       lipgloss.Style
	Cancelled  lipgloss.Style
	InProgress lipgloss.Style
	Deferred   lipgloss.Style
	Selected   lipgloss.Style
	HasNote    lipgloss.Style
}

// DefaultTaskStyles returns default task styling.
func DefaultTaskStyles() TaskStyles {
	return TaskStyles{
		Open:       lipgloss.NewStyle().Foreground(lipgloss.Color("#6B5D4D")),
		Done:       lipgloss.NewStyle().Foreground(lipgloss.Color("#4A7C59")).Strikethrough(true),
		Cancelled:  lipgloss.NewStyle().Foreground(lipgloss.Color("#9C8B75")).Strikethrough(true),
		InProgress: lipgloss.NewStyle().Foreground(lipgloss.Color("#2D5A7B")).Bold(true),
		Deferred:   lipgloss.NewStyle().Foreground(lipgloss.Color("#B8860B")),
		Selected:   lipgloss.NewStyle().Background(lipgloss.Color("#DDD7C9")),
		HasNote:    lipgloss.NewStyle().Foreground(lipgloss.Color("#8B6914")),
	}
}

// DefaultTaskIcons returns default task icons.
func DefaultTaskIcons() map[string]string {
	return map[string]string{
		" ": "○",
		"x": "✓",
		"-": "⊘",
		"/": "◐",
		">": "→",
		"?": "?",
	}
}

// NewTaskList creates a new task list component.
func NewTaskList(tasks []types.Task, width, height int) *TaskList {
	return &TaskList{
		Tasks:    tasks,
		Width:    width,
		Height:   height,
		Icons:    DefaultTaskIcons(),
		Styles:   DefaultTaskStyles(),
		Selected: -1,
	}
}

// SetSelected sets the selected task index.
func (t *TaskList) SetSelected(index int) *TaskList {
	t.Selected = index
	return t
}

// Render renders the task list.
func (t *TaskList) Render() string {
	if len(t.Tasks) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9C8B75")).
			Render("No tasks")
	}

	var lines []string
	for i, task := range t.Tasks {
		if i >= t.Height {
			break
		}

		line := t.renderTask(task, i == t.Selected, 0)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (t *TaskList) renderTask(task types.Task, selected bool, indent int) string {
	icon := t.Icons[task.Status]
	if icon == "" {
		icon = "?"
	}

	// Get style based on status
	style := t.getStyleForStatus(task.Status)

	// Build task line
	indentStr := strings.Repeat("  ", indent)
	text := task.Text

	// Truncate if too long
	maxLen := t.Width - len(indentStr) - 4
	if task.HasNote {
		maxLen -= 2 // space for note icon
	}
	if len(text) > maxLen {
		text = text[:maxLen-1] + "…"
	}

	line := fmt.Sprintf("%s%s %s", indentStr, icon, text)

	// Add note indicator
	if task.HasNote {
		line += " " + t.Styles.HasNote.Render("📎")
	}

	// Apply selection styling
	if selected {
		line = t.Styles.Selected.Render(line)
	} else {
		line = style.Render(line)
	}

	return line
}

func (t *TaskList) getStyleForStatus(status string) lipgloss.Style {
	switch status {
	case " ":
		return t.Styles.Open
	case "x":
		return t.Styles.Done
	case "-":
		return t.Styles.Cancelled
	case "/":
		return t.Styles.InProgress
	case ">":
		return t.Styles.Deferred
	default:
		return t.Styles.Open
	}
}

// RenderWithSubtasks renders tasks with their subtasks.
func (t *TaskList) RenderWithSubtasks() string {
	if len(t.Tasks) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9C8B75")).
			Render("No tasks")
	}

	var lines []string
	lineCount := 0

	for i, task := range t.Tasks {
		if lineCount >= t.Height {
			break
		}

		line := t.renderTask(task, i == t.Selected, 0)
		lines = append(lines, line)
		lineCount++

		// Render subtasks
		for _, subtask := range task.Subtasks {
			if lineCount >= t.Height {
				break
			}
			subline := t.renderTask(subtask, false, 1)
			lines = append(lines, subline)
			lineCount++
		}
	}

	return strings.Join(lines, "\n")
}

// Progress calculates the progress of the task list.
func (t *TaskList) Progress() (completed, total int) {
	for _, task := range t.Tasks {
		total++
		if task.Status == "x" {
			completed++
		}
	}
	return completed, total
}
