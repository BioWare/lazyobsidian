// Package vault handles reading and writing to the Obsidian vault.
package vault

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/BioWare/lazyobsidian/internal/config"
	"github.com/BioWare/lazyobsidian/internal/logging"
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// Writer handles writing changes to vault files.
type Writer struct {
	vaultPath string
	config    *config.Config
	parser    *Parser
}

// NewWriter creates a new vault writer.
func NewWriter(vaultPath string, cfg *config.Config, parser *Parser) *Writer {
	return &Writer{
		vaultPath: vaultPath,
		config:    cfg,
		parser:    parser,
	}
}

// ToggleTask toggles a task's status in the specified file.
// It reads the file, modifies the task at the given line, and writes it back.
func (w *Writer) ToggleTask(filePath string, task *types.Task) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}

	logging.Debug("Toggling task at line %d in %s", task.Line, filePath)

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find and modify the task line
	if task.Line < 1 || task.Line > len(lines) {
		return fmt.Errorf("invalid line number: %d (file has %d lines)", task.Line, len(lines))
	}

	lineIdx := task.Line - 1 // Convert to 0-indexed
	oldLine := lines[lineIdx]

	// Determine the new status
	newStatus := "done"
	if task.Status == "done" {
		newStatus = "open"
	}

	// Get the symbol for the new status
	newSymbol := w.parser.statusToSymbol(newStatus)

	// Replace the status in the line
	newLine := w.replaceTaskStatus(oldLine, newSymbol)
	if newLine == oldLine {
		logging.Warn("Task line unchanged, pattern might not match: %s", oldLine)
		return fmt.Errorf("could not find task pattern in line")
	}

	lines[lineIdx] = newLine
	logging.Debug("Changed line from '%s' to '%s'", oldLine, newLine)

	// Write the file back
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update the task status in memory
	task.Status = newStatus

	logging.Info("Task toggled: %s -> %s", task.Text, newStatus)
	return nil
}

// replaceTaskStatus replaces the status symbol in a task line.
func (w *Writer) replaceTaskStatus(line, newSymbol string) string {
	// Find task pattern: - [X] text
	// Use regex to find and replace
	matches := taskPattern.FindStringSubmatch(line)
	if matches == nil {
		return line
	}

	// matches[0] = full match
	// matches[1] = indentation
	// matches[2] = old status symbol
	// matches[3] = task text

	indent := matches[1]
	oldSymbol := matches[2]
	text := matches[3]

	// Build new line
	newLine := fmt.Sprintf("%s- [%s] %s", indent, newSymbol, text)

	logging.Debug("Replacing task status: [%s] -> [%s]", oldSymbol, newSymbol)
	return newLine
}

// UpdateTaskStatus updates a task's status to a specific value.
func (w *Writer) UpdateTaskStatus(filePath string, task *types.Task, newStatus string) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}

	logging.Debug("Updating task at line %d to status %s in %s", task.Line, newStatus, filePath)

	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find and modify the task line
	if task.Line < 1 || task.Line > len(lines) {
		return fmt.Errorf("invalid line number: %d (file has %d lines)", task.Line, len(lines))
	}

	lineIdx := task.Line - 1 // Convert to 0-indexed
	oldLine := lines[lineIdx]

	// Get the symbol for the new status
	newSymbol := w.parser.statusToSymbol(newStatus)

	// Replace the status in the line
	newLine := w.replaceTaskStatus(oldLine, newSymbol)
	if newLine == oldLine {
		logging.Warn("Task line unchanged, pattern might not match: %s", oldLine)
		return fmt.Errorf("could not find task pattern in line")
	}

	lines[lineIdx] = newLine

	// Write the file back
	newContent := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update the task status in memory
	task.Status = newStatus

	logging.Info("Task status updated: %s -> %s", task.Text, newStatus)
	return nil
}

// AppendToFile appends content to a file.
func (w *Writer) AppendToFile(filePath string, content string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// InsertAtSection inserts content at a specific section in a file.
// If the section doesn't exist, it appends to the end.
func (w *Writer) InsertAtSection(filePath string, sectionHeading string, content string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string
	inserted := false

	for i, line := range lines {
		newLines = append(newLines, line)

		// Check if this is the target section
		if !inserted && strings.HasPrefix(line, "#") && strings.Contains(line, sectionHeading) {
			// Find the end of the section (next heading or end of file)
			insertIdx := i + 1
			for insertIdx < len(lines) {
				if strings.HasPrefix(lines[insertIdx], "#") {
					break
				}
				insertIdx++
			}

			// Insert content at the end of the section
			if insertIdx <= i+1 {
				// Empty section, insert right after heading
				newLines = append(newLines, content)
			} else {
				// Insert before the last empty line in the section
				for insertIdx > i+1 && strings.TrimSpace(lines[insertIdx-1]) == "" {
					insertIdx--
				}
				// Rebuild with insertion
				newLines = lines[:insertIdx]
				newLines = append(newLines, content)
				newLines = append(newLines, lines[insertIdx:]...)
			}
			inserted = true
			break
		}
	}

	if !inserted {
		// Section not found, append to end
		newLines = append(newLines, "", content)
	}

	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(filePath, []byte(newContent), 0644)
}

// CreateDailyNote creates a new daily note from template.
func (w *Writer) CreateDailyNote(filePath string) error {
	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return nil // File exists, don't overwrite
	}

	// Create default content
	content := `---
type: daily
---

## Today's Focus

- [ ]

## Notes

## Pomodoros

`

	// Ensure directory exists
	dir := filePath[:strings.LastIndex(filePath, "/")]
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

// ReadFileLines reads a file and returns its lines.
func (w *Writer) ReadFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}
