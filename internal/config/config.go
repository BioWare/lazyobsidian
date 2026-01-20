// Package config handles loading and managing application configuration.
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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
