// Package importer handles importing data from other applications.
package importer

import (
	"encoding/json"
	"fmt"
	"os"
)

// Source represents an import source type.
type Source string

const (
	SourceFocusToDo Source = "focus_todo"
	SourceTodoist   Source = "todoist"
	SourceTickTick  Source = "ticktick"
	SourceThings3   Source = "things3"
	SourceNotion    Source = "notion"
	SourceGenericCSV Source = "generic_csv"
)

// ImportResult contains the result of an import operation.
type ImportResult struct {
	TasksImported     int
	ProjectsImported  int
	PomodorosImported int
	NotesImported     int
	Errors            []string
}

// Importer handles importing from various sources.
type Importer struct {
	source Source
}

// New creates a new importer for the given source.
func New(source Source) *Importer {
	return &Importer{source: source}
}

// Import imports data from a file.
func (i *Importer) Import(filePath string) (*ImportResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	switch i.source {
	case SourceFocusToDo:
		return i.importFocusToDo(data)
	case SourceTodoist:
		return i.importTodoist(data)
	default:
		return nil, fmt.Errorf("unsupported import source: %s", i.source)
	}
}

func (i *Importer) importFocusToDo(data []byte) (*ImportResult, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result := &ImportResult{}

	// TODO: Parse Focus To-Do JSON structure
	// This is a placeholder implementation

	return result, nil
}

func (i *Importer) importTodoist(data []byte) (*ImportResult, error) {
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	result := &ImportResult{}

	// TODO: Parse Todoist JSON structure
	// This is a placeholder implementation

	return result, nil
}
