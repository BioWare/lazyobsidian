// Package config handles loading and managing application configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// Config represents the application configuration.
type Config struct {
	Vault       VaultConfig       `yaml:"vault"`
	Folders     FoldersConfig     `yaml:"folders"`
	Daily       DailyConfig       `yaml:"daily"`
	Tasks       TasksConfig       `yaml:"tasks"`
	Pomodoro    PomodoroConfig    `yaml:"pomodoro"`
	Sounds      SoundsConfig      `yaml:"sounds"`
	Icons       IconsConfig       `yaml:"icons"`
	Theme       ThemeConfig       `yaml:"theme"`
	Display     DisplayConfig     `yaml:"display"`
	Import      ImportConfig      `yaml:"import"`
	Language    string            `yaml:"language"`
	Keybindings map[string]string `yaml:"keybindings"`
}

// VaultConfig holds vault-related settings.
type VaultConfig struct {
	Path              string `yaml:"path"`
	AutoCreateFolders bool   `yaml:"auto_create_folders"`
}

// FoldersConfig holds folder path mappings.
type FoldersConfig struct {
	Daily     string `yaml:"daily"`
	Goals     string `yaml:"goals"`
	Courses   string `yaml:"courses"`
	Books     string `yaml:"books"`
	Notes     string `yaml:"notes"`
	Templates string `yaml:"templates"`
	Wishlist  string `yaml:"wishlist"`
}

// DailyConfig holds daily note settings.
type DailyConfig struct {
	Folder         string `yaml:"folder"`
	FilenameFormat string `yaml:"filename_format"`
}

// TaskStatusConfig holds a task status definition.
type TaskStatusConfig struct {
	Symbol string `yaml:"symbol"`
	Name   string `yaml:"name"`
	Icon   string `yaml:"icon"`
	Color  string `yaml:"color"`
}

// SubtasksConfig holds subtask settings.
type SubtasksConfig struct {
	Enabled           bool `yaml:"enabled"`
	AutoCloseChildren bool `yaml:"auto_close_children"`
}

// TaskNotesConfig holds task notes settings.
type TaskNotesConfig struct {
	Folder          string `yaml:"folder"`
	IncludeInSearch bool   `yaml:"include_in_search"`
	IncludeInGraph  bool   `yaml:"include_in_graph"`
	IncludeInStats  bool   `yaml:"include_in_stats"`
}

// TasksConfig holds task-related settings.
type TasksConfig struct {
	Statuses []TaskStatusConfig `yaml:"statuses"`
	Subtasks SubtasksConfig     `yaml:"subtasks"`
	Notes    TaskNotesConfig    `yaml:"notes"`
}

// PomodoroLoggingConfig holds pomodoro logging settings.
type PomodoroLoggingConfig struct {
	Mode       string `yaml:"mode"`
	Format     string `yaml:"format"`
	SingleFile string `yaml:"single_file"`
	LogBreaks  bool   `yaml:"log_breaks"`
}

// PomodoroConfig holds pomodoro timer settings.
type PomodoroConfig struct {
	WorkMinutes        int                   `yaml:"work_minutes"`
	ShortBreak         int                   `yaml:"short_break"`
	LongBreak          int                   `yaml:"long_break"`
	SessionsBeforeLong int                   `yaml:"sessions_before_long"`
	DailyGoal          int                   `yaml:"daily_goal"`
	AutoStartBreak     bool                  `yaml:"auto_start_break"`
	AutoStartWork      bool                  `yaml:"auto_start_work"`
	RequireContext     bool                  `yaml:"require_context"`
	RunInBackground    bool                  `yaml:"run_in_background"`
	ShowInStatusline   bool                  `yaml:"show_in_statusline"`
	Logging            PomodoroLoggingConfig `yaml:"logging"`
}

// SoundsConfig holds sound settings.
type SoundsConfig struct {
	Enabled bool              `yaml:"enabled"`
	Volume  float64           `yaml:"volume"`
	Preset  string            `yaml:"preset"`
	Custom  map[string]string `yaml:"custom"`
}

// IconsConfig holds icon settings.
type IconsConfig struct {
	Mode   string            `yaml:"mode"`
	Custom map[string]string `yaml:"custom"`
}

// ThemeConfig holds theme settings.
type ThemeConfig struct {
	Current          string `yaml:"current"`
	SyncWithNeovim   bool   `yaml:"sync_with_neovim"`
	NeovimColorscheme string `yaml:"neovim_colorscheme"`
}

// DynamicWindowsConfig holds dynamic window settings.
type DynamicWindowsConfig struct {
	Enabled          bool `yaml:"enabled"`
	SidebarNormal    int  `yaml:"sidebar_normal"`
	SidebarMinimized int  `yaml:"sidebar_minimized"`
}

// DisplayConfig holds display settings.
type DisplayConfig struct {
	DateFormat       string               `yaml:"date_format"`
	TimeFormat       string               `yaml:"time_format"`
	FirstDayOfWeek   string               `yaml:"first_day_of_week"`
	DynamicWindows   DynamicWindowsConfig `yaml:"dynamic_windows"`
	ProgressBarStyle string               `yaml:"progress_bar_style"`
	PomodoroDisplay  string               `yaml:"pomodoro_display"`
	GoalsProgress    string               `yaml:"goals_progress"`
}

// ImportConfig holds import settings.
type ImportConfig struct {
	CreateMode         string `yaml:"create_mode"`
	Folder             string `yaml:"folder"`
	ImportCompleted    bool   `yaml:"import_completed"`
	AutoCloseCompleted bool   `yaml:"auto_close_completed"`
	MergeExisting      bool   `yaml:"merge_existing"`
}

// DefaultConfig returns a configuration with default values.
func DefaultConfig() *Config {
	return &Config{
		Vault: VaultConfig{
			Path:              "",
			AutoCreateFolders: true,
		},
		Folders: FoldersConfig{
			Daily:     "Journal",
			Goals:     "Plan",
			Courses:   "Input/Courses",
			Books:     "Input/Books",
			Notes:     "Zettlekasten",
			Templates: "templates",
			Wishlist:  "Wishlist",
		},
		Daily: DailyConfig{
			Folder:         "Journal",
			FilenameFormat: "2006-01-02",
		},
		Tasks: TasksConfig{
			Statuses: []TaskStatusConfig{
				{Symbol: " ", Name: "open", Icon: "○", Color: "text_secondary"},
				{Symbol: "x", Name: "done", Icon: "✓", Color: "success"},
				{Symbol: "-", Name: "cancelled", Icon: "⊘", Color: "text_muted"},
				{Symbol: "/", Name: "in_progress", Icon: "◐", Color: "primary"},
				{Symbol: ">", Name: "deferred", Icon: "→", Color: "warning"},
				{Symbol: "?", Name: "question", Icon: "?", Color: "info"},
			},
			Subtasks: SubtasksConfig{
				Enabled:           true,
				AutoCloseChildren: false,
			},
			Notes: TaskNotesConfig{
				Folder:          ".task-notes",
				IncludeInSearch: false,
				IncludeInGraph:  false,
				IncludeInStats:  false,
			},
		},
		Pomodoro: PomodoroConfig{
			WorkMinutes:        25,
			ShortBreak:         5,
			LongBreak:          15,
			SessionsBeforeLong: 4,
			DailyGoal:          5,
			AutoStartBreak:     true,
			AutoStartWork:      false,
			RequireContext:     true,
			RunInBackground:    true,
			ShowInStatusline:   true,
			Logging: PomodoroLoggingConfig{
				Mode:       "context",
				Format:     "section",
				SingleFile: "pomodoro-log.md",
				LogBreaks:  true,
			},
		},
		Sounds: SoundsConfig{
			Enabled: true,
			Volume:  0.8,
			Preset:  "corsair",
			Custom:  make(map[string]string),
		},
		Icons: IconsConfig{
			Mode:   "nerd_minimal",
			Custom: make(map[string]string),
		},
		Theme: ThemeConfig{
			Current:          "corsair-light",
			SyncWithNeovim:   false,
			NeovimColorscheme: "",
		},
		Display: DisplayConfig{
			DateFormat:       "Jan 2, 2006",
			TimeFormat:       "15:04",
			FirstDayOfWeek:   "monday",
			DynamicWindows: DynamicWindowsConfig{
				Enabled:          true,
				SidebarNormal:    20,
				SidebarMinimized: 12,
			},
			ProgressBarStyle: "blocks",
			PomodoroDisplay:  "adaptive",
			GoalsProgress:    "bar",
		},
		Import: ImportConfig{
			CreateMode:         "separate_files",
			Folder:             "imports",
			ImportCompleted:    true,
			AutoCloseCompleted: true,
			MergeExisting:      false,
		},
		Language:    "en",
		Keybindings: make(map[string]string),
	}
}

// Load loads configuration from the default location.
func Load() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return DefaultConfig(), nil
	}

	configPath := filepath.Join(configDir, "config.yaml")
	return LoadFromFile(configPath)
}

// LoadFromFile loads configuration from a specific file.
func LoadFromFile(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves the configuration to the default location.
func (c *Config) Save() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.yaml")
	return c.SaveToFile(configPath)
}

// SaveToFile saves the configuration to a specific file.
func (c *Config) SaveToFile(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func getConfigDir() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "lazyobsidian"), nil
}

// Validate validates the configuration and returns any errors.
func (c *Config) Validate() ValidationErrors {
	var errs ValidationErrors

	// Vault validation
	if c.Vault.Path == "" {
		errs = append(errs, ValidationError{
			Field:   "vault.path",
			Message: "vault path is required",
		})
	} else if !filepath.IsAbs(c.Vault.Path) {
		errs = append(errs, ValidationError{
			Field:   "vault.path",
			Message: "vault path must be absolute",
		})
	}

	// Pomodoro validation
	if c.Pomodoro.WorkMinutes <= 0 {
		errs = append(errs, ValidationError{
			Field:   "pomodoro.work_minutes",
			Message: "work minutes must be positive",
		})
	}
	if c.Pomodoro.ShortBreak <= 0 {
		errs = append(errs, ValidationError{
			Field:   "pomodoro.short_break",
			Message: "short break must be positive",
		})
	}
	if c.Pomodoro.LongBreak <= 0 {
		errs = append(errs, ValidationError{
			Field:   "pomodoro.long_break",
			Message: "long break must be positive",
		})
	}
	if c.Pomodoro.SessionsBeforeLong <= 0 {
		errs = append(errs, ValidationError{
			Field:   "pomodoro.sessions_before_long",
			Message: "sessions before long break must be positive",
		})
	}
	if c.Pomodoro.DailyGoal < 0 {
		errs = append(errs, ValidationError{
			Field:   "pomodoro.daily_goal",
			Message: "daily goal cannot be negative",
		})
	}

	// Logging mode validation
	validLoggingModes := map[string]bool{"context": true, "daily": true, "single_file": true}
	if c.Pomodoro.Logging.Mode != "" && !validLoggingModes[c.Pomodoro.Logging.Mode] {
		errs = append(errs, ValidationError{
			Field:   "pomodoro.logging.mode",
			Message: fmt.Sprintf("invalid logging mode: %s (valid: context, daily, single_file)", c.Pomodoro.Logging.Mode),
		})
	}

	// Logging format validation
	validLoggingFormats := map[string]bool{"inline": true, "section": true, "table": true}
	if c.Pomodoro.Logging.Format != "" && !validLoggingFormats[c.Pomodoro.Logging.Format] {
		errs = append(errs, ValidationError{
			Field:   "pomodoro.logging.format",
			Message: fmt.Sprintf("invalid logging format: %s (valid: inline, section, table)", c.Pomodoro.Logging.Format),
		})
	}

	// Sound validation
	if c.Sounds.Volume < 0 || c.Sounds.Volume > 1 {
		errs = append(errs, ValidationError{
			Field:   "sounds.volume",
			Message: "volume must be between 0 and 1",
		})
	}

	// Icons mode validation
	validIconModes := map[string]bool{"emoji": true, "nerd": true, "nerd_minimal": true, "ascii": true}
	if c.Icons.Mode != "" && !validIconModes[c.Icons.Mode] {
		errs = append(errs, ValidationError{
			Field:   "icons.mode",
			Message: fmt.Sprintf("invalid icon mode: %s (valid: emoji, nerd, nerd_minimal, ascii)", c.Icons.Mode),
		})
	}

	// Theme validation
	validThemes := map[string]bool{"corsair-light": true, "corsair-dark": true}
	if c.Theme.Current != "" && !validThemes[c.Theme.Current] {
		// Allow custom themes, just warn if not a known built-in
		// This is not an error, just a note
	}

	// Display validation
	if c.Display.DynamicWindows.SidebarNormal < 10 || c.Display.DynamicWindows.SidebarNormal > 50 {
		errs = append(errs, ValidationError{
			Field:   "display.dynamic_windows.sidebar_normal",
			Message: "sidebar normal width must be between 10% and 50%",
		})
	}
	if c.Display.DynamicWindows.SidebarMinimized < 5 || c.Display.DynamicWindows.SidebarMinimized > 30 {
		errs = append(errs, ValidationError{
			Field:   "display.dynamic_windows.sidebar_minimized",
			Message: "sidebar minimized width must be between 5% and 30%",
		})
	}

	// Progress bar style validation
	validProgressStyles := map[string]bool{"blocks": true, "line": true, "dots": true, "percent": true}
	if c.Display.ProgressBarStyle != "" && !validProgressStyles[c.Display.ProgressBarStyle] {
		errs = append(errs, ValidationError{
			Field:   "display.progress_bar_style",
			Message: fmt.Sprintf("invalid progress bar style: %s (valid: blocks, line, dots, percent)", c.Display.ProgressBarStyle),
		})
	}

	// First day of week validation
	validFirstDays := map[string]bool{"monday": true, "sunday": true}
	if c.Display.FirstDayOfWeek != "" && !validFirstDays[c.Display.FirstDayOfWeek] {
		errs = append(errs, ValidationError{
			Field:   "display.first_day_of_week",
			Message: fmt.Sprintf("invalid first day of week: %s (valid: monday, sunday)", c.Display.FirstDayOfWeek),
		})
	}

	// Language validation
	validLanguages := map[string]bool{"en": true, "ru": true}
	if c.Language != "" && !validLanguages[c.Language] {
		errs = append(errs, ValidationError{
			Field:   "language",
			Message: fmt.Sprintf("unsupported language: %s (valid: en, ru)", c.Language),
		})
	}

	// Task statuses validation
	if len(c.Tasks.Statuses) == 0 {
		errs = append(errs, ValidationError{
			Field:   "tasks.statuses",
			Message: "at least one task status must be defined",
		})
	} else {
		symbolsSeen := make(map[string]bool)
		namesSeen := make(map[string]bool)
		for i, status := range c.Tasks.Statuses {
			if status.Symbol == "" {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("tasks.statuses[%d].symbol", i),
					Message: "status symbol is required",
				})
			} else if symbolsSeen[status.Symbol] {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("tasks.statuses[%d].symbol", i),
					Message: fmt.Sprintf("duplicate symbol: %s", status.Symbol),
				})
			} else {
				symbolsSeen[status.Symbol] = true
			}

			if status.Name == "" {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("tasks.statuses[%d].name", i),
					Message: "status name is required",
				})
			} else if namesSeen[status.Name] {
				errs = append(errs, ValidationError{
					Field:   fmt.Sprintf("tasks.statuses[%d].name", i),
					Message: fmt.Sprintf("duplicate name: %s", status.Name),
				})
			} else {
				namesSeen[status.Name] = true
			}
		}
	}

	return errs
}

// ValidateAndFix validates the configuration and applies default fixes where possible.
// Returns true if any fixes were applied.
func (c *Config) ValidateAndFix() (bool, ValidationErrors) {
	fixed := false

	// Apply defaults for missing values
	if c.Pomodoro.WorkMinutes <= 0 {
		c.Pomodoro.WorkMinutes = 25
		fixed = true
	}
	if c.Pomodoro.ShortBreak <= 0 {
		c.Pomodoro.ShortBreak = 5
		fixed = true
	}
	if c.Pomodoro.LongBreak <= 0 {
		c.Pomodoro.LongBreak = 15
		fixed = true
	}
	if c.Pomodoro.SessionsBeforeLong <= 0 {
		c.Pomodoro.SessionsBeforeLong = 4
		fixed = true
	}
	if c.Pomodoro.DailyGoal < 0 {
		c.Pomodoro.DailyGoal = 5
		fixed = true
	}

	// Fix sound volume
	if c.Sounds.Volume < 0 {
		c.Sounds.Volume = 0
		fixed = true
	} else if c.Sounds.Volume > 1 {
		c.Sounds.Volume = 1
		fixed = true
	}

	// Fix sidebar widths
	if c.Display.DynamicWindows.SidebarNormal < 10 {
		c.Display.DynamicWindows.SidebarNormal = 10
		fixed = true
	} else if c.Display.DynamicWindows.SidebarNormal > 50 {
		c.Display.DynamicWindows.SidebarNormal = 50
		fixed = true
	}
	if c.Display.DynamicWindows.SidebarMinimized < 5 {
		c.Display.DynamicWindows.SidebarMinimized = 5
		fixed = true
	} else if c.Display.DynamicWindows.SidebarMinimized > 30 {
		c.Display.DynamicWindows.SidebarMinimized = 30
		fixed = true
	}

	// Default task statuses if missing
	if len(c.Tasks.Statuses) == 0 {
		c.Tasks.Statuses = DefaultConfig().Tasks.Statuses
		fixed = true
	}

	// Validate remaining issues that can't be auto-fixed
	errs := c.Validate()

	return fixed, errs
}

// EnsureVaultPath checks if the vault path exists and optionally creates required folders.
func (c *Config) EnsureVaultPath() error {
	if c.Vault.Path == "" {
		return errors.New("vault path is not configured")
	}

	// Check if vault exists
	info, err := os.Stat(c.Vault.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vault path does not exist: %s", c.Vault.Path)
		}
		return fmt.Errorf("cannot access vault path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("vault path is not a directory: %s", c.Vault.Path)
	}

	// Create required folders if enabled
	if c.Vault.AutoCreateFolders {
		folders := []string{
			c.Folders.Daily,
			c.Folders.Goals,
			c.Folders.Courses,
			c.Folders.Books,
			c.Folders.Notes,
			c.Folders.Templates,
			c.Folders.Wishlist,
		}

		for _, folder := range folders {
			if folder == "" {
				continue
			}
			fullPath := filepath.Join(c.Vault.Path, folder)
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				return fmt.Errorf("failed to create folder %s: %w", folder, err)
			}
		}
	}

	return nil
}
