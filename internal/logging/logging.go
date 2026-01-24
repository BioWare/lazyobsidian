// Package logging provides a file-based logging system for debugging.
// Logs are written to ~/.lazyobsidian/debug.log
package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents log severity levels.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger is the main logging instance.
type Logger struct {
	file    *os.File
	mu      sync.Mutex
	level   Level
	enabled bool
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Init initializes the default logger.
func Init(enabled bool) error {
	var initErr error
	once.Do(func() {
		defaultLogger = &Logger{
			level:   LevelDebug,
			enabled: enabled,
		}

		if !enabled {
			return
		}

		// Create log directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			initErr = fmt.Errorf("failed to get home directory: %w", err)
			return
		}

		logDir := filepath.Join(homeDir, ".lazyobsidian")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			initErr = fmt.Errorf("failed to create log directory: %w", err)
			return
		}

		logPath := filepath.Join(logDir, "debug.log")
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}

		defaultLogger.file = file

		// Write startup marker
		defaultLogger.Info("=== LazyObsidian started ===")
	})

	return initErr
}

// Close closes the log file.
func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.Info("=== LazyObsidian stopped ===")
		defaultLogger.file.Close()
	}
}

// SetLevel sets the minimum log level.
func SetLevel(level Level) {
	if defaultLogger != nil {
		defaultLogger.level = level
	}
}

// log writes a log entry.
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if !l.enabled || level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	entry := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level.String(), message)

	l.file.WriteString(entry)
}

// Debug logs a debug message.
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelDebug, format, args...)
	}
}

// Info logs an info message.
func Info(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelInfo, format, args...)
	}
}

// Warn logs a warning message.
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelWarn, format, args...)
	}
}

// Error logs an error message.
func Error(format string, args ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(LevelError, format, args...)
	}
}

// Logger methods for instance-based logging.
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}
