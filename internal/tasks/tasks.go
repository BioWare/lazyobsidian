// Package tasks handles task management functionality.
package tasks

import (
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// Manager handles task operations.
type Manager struct {
	statusMap map[string]types.TaskStatus
}

// NewManager creates a new task manager.
func NewManager(statuses []types.TaskStatus) *Manager {
	statusMap := make(map[string]types.TaskStatus)
	for _, s := range statuses {
		statusMap[s.Symbol] = s
	}

	return &Manager{
		statusMap: statusMap,
	}
}

// GetStatus returns the status definition for a symbol.
func (m *Manager) GetStatus(symbol string) (types.TaskStatus, bool) {
	status, ok := m.statusMap[symbol]
	return status, ok
}

// GetIcon returns the icon for a task status.
func (m *Manager) GetIcon(symbol string) string {
	if status, ok := m.statusMap[symbol]; ok {
		return status.Icon
	}
	return "?"
}

// NextStatus returns the next status in the cycle.
func (m *Manager) NextStatus(current string) string {
	// Cycle: open -> in_progress -> done -> open
	switch current {
	case " ":
		return "/"
	case "/":
		return "x"
	case "x":
		return " "
	default:
		return " "
	}
}

// IsComplete returns true if the task is completed.
func (m *Manager) IsComplete(symbol string) bool {
	return symbol == "x"
}

// IsOpen returns true if the task is open.
func (m *Manager) IsOpen(symbol string) bool {
	return symbol == " "
}

// IsCancelled returns true if the task is cancelled.
func (m *Manager) IsCancelled(symbol string) bool {
	return symbol == "-"
}

// CalculateProgress calculates progress for a list of tasks.
func CalculateProgress(tasks []types.Task) (completed, total int) {
	for _, task := range tasks {
		total++
		if task.Status == "x" {
			completed++
		}
		// Recursively count subtasks
		c, t := CalculateProgress(task.Subtasks)
		completed += c
		total += t
	}
	return completed, total
}
