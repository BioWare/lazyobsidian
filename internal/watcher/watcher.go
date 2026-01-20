// Package watcher monitors the vault for file changes.
package watcher

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// EventType represents the type of file event.
type EventType int

const (
	EventCreate EventType = iota
	EventModify
	EventDelete
	EventRename
)

// Event represents a file change event.
type Event struct {
	Type EventType
	Path string
}

// Watcher monitors the vault for file changes.
type Watcher struct {
	watcher   *fsnotify.Watcher
	vaultPath string
	Events    chan Event
	Errors    chan error
	done      chan struct{}
}

// New creates a new file watcher for the vault.
func New(vaultPath string) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher:   fsWatcher,
		vaultPath: vaultPath,
		Events:    make(chan Event, 100),
		Errors:    make(chan error, 10),
		done:      make(chan struct{}),
	}

	return w, nil
}

// Start starts watching the vault.
func (w *Watcher) Start() error {
	// Add vault path recursively
	if err := w.addRecursive(w.vaultPath); err != nil {
		return err
	}

	go w.watch()
	return nil
}

// Stop stops the watcher.
func (w *Watcher) Stop() error {
	close(w.done)
	return w.watcher.Close()
}

func (w *Watcher) watch() {
	for {
		select {
		case <-w.done:
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only care about markdown files
			if !strings.HasSuffix(event.Name, ".md") {
				continue
			}

			var eventType EventType
			switch {
			case event.Op&fsnotify.Create == fsnotify.Create:
				eventType = EventCreate
			case event.Op&fsnotify.Write == fsnotify.Write:
				eventType = EventModify
			case event.Op&fsnotify.Remove == fsnotify.Remove:
				eventType = EventDelete
			case event.Op&fsnotify.Rename == fsnotify.Rename:
				eventType = EventRename
			default:
				continue
			}

			w.Events <- Event{
				Type: eventType,
				Path: event.Name,
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			w.Errors <- err
		}
	}
}

func (w *Watcher) addRecursive(path string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return w.watcher.Add(path)
		}

		return nil
	})
}
