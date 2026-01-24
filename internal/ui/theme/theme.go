// Package theme handles color themes and styling for the TUI.
package theme

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

// Theme represents a color theme.
type Theme struct {
	Name   string            `yaml:"name"`
	Type   string            `yaml:"type"` // "light" or "dark"
	Colors map[string]string `yaml:"colors"`

	// Cached lipgloss colors
	colors map[string]lipgloss.Color
}

// Current is the currently active theme.
var Current *Theme

// Styles contains all the styled components.
type Styles struct {
	// Base
	App    lipgloss.Style
	Header lipgloss.Style
	Footer lipgloss.Style

	// Sidebar
	Sidebar         lipgloss.Style
	SidebarItem     lipgloss.Style
	SidebarSelected lipgloss.Style
	SidebarFocused  lipgloss.Style
	SidebarSection  lipgloss.Style

	// Main panel
	MainPanel   lipgloss.Style
	PanelTitle  lipgloss.Style
	PanelBorder lipgloss.Style

	// Tasks
	TaskOpen       lipgloss.Style
	TaskDone       lipgloss.Style
	TaskCancelled  lipgloss.Style
	TaskInProgress lipgloss.Style
	TaskDeferred   lipgloss.Style
	TaskQuestion   lipgloss.Style
	TaskSelected   lipgloss.Style

	// Progress
	ProgressFill  lipgloss.Style
	ProgressEmpty lipgloss.Style

	// Pomodoro
	PomodoroWork    lipgloss.Style
	PomodoroBreak   lipgloss.Style
	PomorodoPaused  lipgloss.Style
	PomodoroTimer   lipgloss.Style
	PomodoroContext lipgloss.Style

	// Calendar
	CalendarDay         lipgloss.Style
	CalendarDayToday    lipgloss.Style
	CalendarDaySelected lipgloss.Style
	CalendarDayMuted    lipgloss.Style
	CalendarWeekend     lipgloss.Style
	CalendarHeader      lipgloss.Style
	CalendarEvent       lipgloss.Style

	// Goals
	GoalTitle       lipgloss.Style
	GoalProgress    lipgloss.Style
	GoalCompleted   lipgloss.Style
	GoalInProgress  lipgloss.Style
	GoalPending     lipgloss.Style
	GoalTreeBranch  lipgloss.Style

	// Courses
	CourseTitle    lipgloss.Style
	CourseProgress lipgloss.Style
	CourseSection  lipgloss.Style
	CourseLesson   lipgloss.Style

	// Books
	BookTitle    lipgloss.Style
	BookAuthor   lipgloss.Style
	BookProgress lipgloss.Style
	BookChapter  lipgloss.Style

	// Stats
	StatsValue     lipgloss.Style
	StatsLabel     lipgloss.Style
	StatsHighlight lipgloss.Style
	HeatmapLevel0  lipgloss.Style
	HeatmapLevel1  lipgloss.Style
	HeatmapLevel2  lipgloss.Style
	HeatmapLevel3  lipgloss.Style
	HeatmapLevel4  lipgloss.Style

	// Graph
	GraphNode       lipgloss.Style
	GraphNodeActive lipgloss.Style
	GraphEdge       lipgloss.Style

	// Settings
	SettingsSection lipgloss.Style
	SettingsKey     lipgloss.Style
	SettingsValue   lipgloss.Style
	SettingsChanged lipgloss.Style

	// Text
	TextPrimary   lipgloss.Style
	TextSecondary lipgloss.Style
	TextMuted     lipgloss.Style
	TextHighlight lipgloss.Style
	TextLink      lipgloss.Style

	// Semantic
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Info    lipgloss.Style

	// Help
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style

	// Modal
	ModalBackground lipgloss.Style
	ModalBorder     lipgloss.Style
	ModalTitle      lipgloss.Style
	ModalButton     lipgloss.Style
	ModalButtonActive lipgloss.Style

	// Input
	InputNormal  lipgloss.Style
	InputFocused lipgloss.Style
	InputError   lipgloss.Style
	InputLabel   lipgloss.Style
}

// S is the global styles instance.
var S Styles

// Load loads a theme from a file.
func Load(path string) (*Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var theme Theme
	if err := yaml.Unmarshal(data, &theme); err != nil {
		return nil, err
	}

	theme.colors = make(map[string]lipgloss.Color)
	for name, hex := range theme.Colors {
		theme.colors[name] = lipgloss.Color(hex)
	}

	return &theme, nil
}

// LoadBuiltin loads a built-in theme by name.
func LoadBuiltin(name string) (*Theme, error) {
	// Try to find in themes directory relative to executable
	execPath, err := os.Executable()
	if err == nil {
		themePath := filepath.Join(filepath.Dir(execPath), "themes", name+".yaml")
		if theme, err := Load(themePath); err == nil {
			return theme, nil
		}
	}

	// Fallback to hardcoded default
	if name == "corsair-light" || name == "" {
		return defaultLightTheme(), nil
	}
	if name == "corsair-dark" {
		return defaultDarkTheme(), nil
	}

	return defaultLightTheme(), nil
}

// Apply applies the theme and generates styles.
func (t *Theme) Apply() {
	Current = t
	S = t.generateStyles()
}

// Color returns a lipgloss color by name.
func (t *Theme) Color(name string) lipgloss.Color {
	if c, ok := t.colors[name]; ok {
		return c
	}
	return lipgloss.Color("#888888")
}

func (t *Theme) generateStyles() Styles {
	return Styles{
		// Base
		App: lipgloss.NewStyle().
			Background(t.Color("bg_primary")).
			Foreground(t.Color("text_primary")),

		Header: lipgloss.NewStyle().
			Background(t.Color("bg_secondary")).
			Foreground(t.Color("text_primary")).
			Bold(true).
			Padding(0, 1),

		Footer: lipgloss.NewStyle().
			Background(t.Color("bg_secondary")).
			Foreground(t.Color("text_secondary")).
			Padding(0, 1),

		// Sidebar
		Sidebar: lipgloss.NewStyle().
			Background(t.Color("bg_secondary")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(t.Color("border_default")).
			BorderRight(true),

		SidebarItem: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")).
			Padding(0, 1),

		SidebarSelected: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Background(t.Color("bg_tertiary")).
			Bold(true).
			Padding(0, 1),

		SidebarFocused: lipgloss.NewStyle().
			Foreground(t.Color("text_inverse")).
			Background(t.Color("primary")).
			Bold(true).
			Padding(0, 1),

		SidebarSection: lipgloss.NewStyle().
			Foreground(t.Color("border_muted")),

		// Main panel
		MainPanel: lipgloss.NewStyle().
			Background(t.Color("bg_primary")).
			Padding(0, 1),

		PanelTitle: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Bold(true).
			MarginBottom(1),

		PanelBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Color("border_default")),

		// Tasks
		TaskOpen: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),

		TaskDone: lipgloss.NewStyle().
			Foreground(t.Color("success")).
			Strikethrough(true),

		TaskCancelled: lipgloss.NewStyle().
			Foreground(t.Color("text_muted")).
			Strikethrough(true),

		TaskInProgress: lipgloss.NewStyle().
			Foreground(t.Color("primary")).
			Bold(true),

		TaskDeferred: lipgloss.NewStyle().
			Foreground(t.Color("warning")),

		TaskQuestion: lipgloss.NewStyle().
			Foreground(t.Color("info")),

		TaskSelected: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Background(t.Color("bg_active")).
			Bold(true),

		// Progress
		ProgressFill: lipgloss.NewStyle().
			Foreground(t.Color("progress_fill")),

		ProgressEmpty: lipgloss.NewStyle().
			Foreground(t.Color("progress_empty")),

		// Pomodoro
		PomodoroWork: lipgloss.NewStyle().
			Foreground(t.Color("pomodoro_work")).
			Bold(true),

		PomodoroBreak: lipgloss.NewStyle().
			Foreground(t.Color("pomodoro_break")),

		PomorodoPaused: lipgloss.NewStyle().
			Foreground(t.Color("pomodoro_paused")),

		PomodoroTimer: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Bold(true),

		PomodoroContext: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")).
			Italic(true),

		// Calendar
		CalendarDay: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")),

		CalendarDayToday: lipgloss.NewStyle().
			Foreground(t.Color("text_inverse")).
			Background(t.Color("primary")).
			Bold(true),

		CalendarDaySelected: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Background(t.Color("bg_active")),

		CalendarDayMuted: lipgloss.NewStyle().
			Foreground(t.Color("text_muted")),

		CalendarWeekend: lipgloss.NewStyle().
			Foreground(t.Color("secondary")),

		CalendarHeader: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")).
			Bold(true),

		CalendarEvent: lipgloss.NewStyle().
			Foreground(t.Color("accent")),

		// Goals
		GoalTitle: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Bold(true),

		GoalProgress: lipgloss.NewStyle().
			Foreground(t.Color("primary")),

		GoalCompleted: lipgloss.NewStyle().
			Foreground(t.Color("success")),

		GoalInProgress: lipgloss.NewStyle().
			Foreground(t.Color("warning")),

		GoalPending: lipgloss.NewStyle().
			Foreground(t.Color("text_muted")),

		GoalTreeBranch: lipgloss.NewStyle().
			Foreground(t.Color("border_default")),

		// Courses
		CourseTitle: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Bold(true),

		CourseProgress: lipgloss.NewStyle().
			Foreground(t.Color("primary")),

		CourseSection: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")).
			Bold(true),

		CourseLesson: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),

		// Books
		BookTitle: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Bold(true),

		BookAuthor: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")).
			Italic(true),

		BookProgress: lipgloss.NewStyle().
			Foreground(t.Color("primary")),

		BookChapter: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),

		// Stats
		StatsValue: lipgloss.NewStyle().
			Foreground(t.Color("primary")).
			Bold(true),

		StatsLabel: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),

		StatsHighlight: lipgloss.NewStyle().
			Foreground(t.Color("accent")).
			Bold(true),

		HeatmapLevel0: lipgloss.NewStyle().
			Foreground(t.Color("heatmap_0")),

		HeatmapLevel1: lipgloss.NewStyle().
			Foreground(t.Color("heatmap_1")),

		HeatmapLevel2: lipgloss.NewStyle().
			Foreground(t.Color("heatmap_2")),

		HeatmapLevel3: lipgloss.NewStyle().
			Foreground(t.Color("heatmap_3")),

		HeatmapLevel4: lipgloss.NewStyle().
			Foreground(t.Color("heatmap_4")),

		// Graph
		GraphNode: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),

		GraphNodeActive: lipgloss.NewStyle().
			Foreground(t.Color("primary")).
			Bold(true),

		GraphEdge: lipgloss.NewStyle().
			Foreground(t.Color("border_default")),

		// Settings
		SettingsSection: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Bold(true),

		SettingsKey: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),

		SettingsValue: lipgloss.NewStyle().
			Foreground(t.Color("primary")),

		SettingsChanged: lipgloss.NewStyle().
			Foreground(t.Color("warning")).
			Bold(true),

		// Text
		TextPrimary: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")),

		TextSecondary: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),

		TextMuted: lipgloss.NewStyle().
			Foreground(t.Color("text_muted")),

		TextHighlight: lipgloss.NewStyle().
			Foreground(t.Color("accent")).
			Bold(true),

		TextLink: lipgloss.NewStyle().
			Foreground(t.Color("info")).
			Underline(true),

		// Semantic
		Success: lipgloss.NewStyle().
			Foreground(t.Color("success")),

		Warning: lipgloss.NewStyle().
			Foreground(t.Color("warning")),

		Error: lipgloss.NewStyle().
			Foreground(t.Color("error")),

		Info: lipgloss.NewStyle().
			Foreground(t.Color("info")),

		// Help
		HelpKey: lipgloss.NewStyle().
			Foreground(t.Color("primary")).
			Bold(true),

		HelpDesc: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),

		// Modal
		ModalBackground: lipgloss.NewStyle().
			Background(t.Color("bg_secondary")),

		ModalBorder: lipgloss.NewStyle().
			BorderForeground(t.Color("border_active")),

		ModalTitle: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Bold(true),

		ModalButton: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")).
			Background(t.Color("bg_tertiary")).
			Padding(0, 2),

		ModalButtonActive: lipgloss.NewStyle().
			Foreground(t.Color("text_inverse")).
			Background(t.Color("primary")).
			Bold(true).
			Padding(0, 2),

		// Input
		InputNormal: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Background(t.Color("bg_tertiary")).
			BorderForeground(t.Color("border_default")),

		InputFocused: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")).
			Background(t.Color("bg_tertiary")).
			BorderForeground(t.Color("border_active")),

		InputError: lipgloss.NewStyle().
			Foreground(t.Color("error")).
			Background(t.Color("bg_tertiary")).
			BorderForeground(t.Color("error")),

		InputLabel: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),
	}
}

func defaultLightTheme() *Theme {
	colors := map[string]string{
		// Background colors
		"bg_primary":      "#F5F0E6",
		"bg_secondary":    "#EDE8DC",
		"bg_tertiary":     "#E5DFD3",
		"bg_active":       "#DDD7C9",

		// Border colors
		"border_default":  "#D4C9B5",
		"border_muted":    "#E0D8CA",
		"border_active":   "#2D5A7B",

		// Text colors
		"text_primary":    "#3D3428",
		"text_secondary":  "#6B5D4D",
		"text_muted":      "#9C8B75",
		"text_inverse":    "#F5F0E6",

		// Brand colors
		"primary":         "#2D5A7B",
		"secondary":       "#7B4D2D",
		"accent":          "#8B6914",

		// Semantic colors
		"success":         "#4A7C59",
		"warning":         "#B8860B",
		"error":           "#8B3A3A",
		"info":            "#4A708B",

		// Pomodoro colors
		"pomodoro_work":   "#8B3A3A",
		"pomodoro_break":  "#4A7C59",
		"pomodoro_paused": "#B8860B",

		// Progress bar colors
		"progress_fill":   "#2D5A7B",
		"progress_empty":  "#D4C9B5",

		// Heatmap colors (5 levels: 0=none, 4=max)
		"heatmap_0":       "#E5DFD3",
		"heatmap_1":       "#B8D4B8",
		"heatmap_2":       "#7AB87A",
		"heatmap_3":       "#4A9C4A",
		"heatmap_4":       "#2D7C2D",

		// Calendar colors
		"calendar_today":    "#2D5A7B",
		"calendar_selected": "#DDD7C9",
		"calendar_event":    "#8B6914",

		// Task priority colors
		"priority_high":   "#8B3A3A",
		"priority_medium": "#B8860B",
		"priority_low":    "#4A7C59",

		// Graph colors
		"graph_node":      "#6B5D4D",
		"graph_edge":      "#D4C9B5",
		"graph_highlight": "#2D5A7B",
	}

	theme := &Theme{
		Name:   "corsair-light",
		Type:   "light",
		Colors: colors,
		colors: make(map[string]lipgloss.Color),
	}

	for name, hex := range colors {
		theme.colors[name] = lipgloss.Color(hex)
	}

	return theme
}

func defaultDarkTheme() *Theme {
	colors := map[string]string{
		// Background colors
		"bg_primary":      "#1C1915",
		"bg_secondary":    "#2A2520",
		"bg_tertiary":     "#3D3630",
		"bg_active":       "#4A433B",

		// Border colors
		"border_default":  "#4A433B",
		"border_muted":    "#3D3630",
		"border_active":   "#5B8FB9",

		// Text colors
		"text_primary":    "#E8E0D4",
		"text_secondary":  "#B8A898",
		"text_muted":      "#786858",
		"text_inverse":    "#1C1915",

		// Brand colors
		"primary":         "#5B8FB9",
		"secondary":       "#C17F59",
		"accent":          "#D4A934",

		// Semantic colors
		"success":         "#6B9B6B",
		"warning":         "#D4A934",
		"error":           "#B85B5B",
		"info":            "#5B8FB9",

		// Pomodoro colors
		"pomodoro_work":   "#B85B5B",
		"pomodoro_break":  "#6B9B6B",
		"pomodoro_paused": "#D4A934",

		// Progress bar colors
		"progress_fill":   "#5B8FB9",
		"progress_empty":  "#3D3630",

		// Heatmap colors (5 levels: 0=none, 4=max)
		"heatmap_0":       "#3D3630",
		"heatmap_1":       "#4A6B4A",
		"heatmap_2":       "#5B8B5B",
		"heatmap_3":       "#6BAB6B",
		"heatmap_4":       "#7BCB7B",

		// Calendar colors
		"calendar_today":    "#5B8FB9",
		"calendar_selected": "#4A433B",
		"calendar_event":    "#D4A934",

		// Task priority colors
		"priority_high":   "#B85B5B",
		"priority_medium": "#D4A934",
		"priority_low":    "#6B9B6B",

		// Graph colors
		"graph_node":      "#B8A898",
		"graph_edge":      "#4A433B",
		"graph_highlight": "#5B8FB9",
	}

	theme := &Theme{
		Name:   "corsair-dark",
		Type:   "dark",
		Colors: colors,
		colors: make(map[string]lipgloss.Color),
	}

	for name, hex := range colors {
		theme.colors[name] = lipgloss.Color(hex)
	}

	return theme
}

// ListBuiltinThemes returns the names of all built-in themes.
func ListBuiltinThemes() []string {
	return []string{"corsair-light", "corsair-dark"}
}

// IsDark returns true if this is a dark theme.
func (t *Theme) IsDark() bool {
	return t.Type == "dark" || strings.Contains(strings.ToLower(t.Name), "dark")
}

// GetThemesDir returns the themes directory path.
func GetThemesDir() (string, error) {
	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		themesDir := filepath.Join(filepath.Dir(execPath), "themes")
		if info, err := os.Stat(themesDir); err == nil && info.IsDir() {
			return themesDir, nil
		}
	}

	// Try XDG config directory
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}

	themesDir := filepath.Join(configDir, "lazyobsidian", "themes")
	return themesDir, nil
}

// ListCustomThemes returns the names of all custom themes in the themes directory.
func ListCustomThemes() ([]string, error) {
	themesDir, err := GetThemesDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(themesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var themes []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml") {
			themeName := strings.TrimSuffix(strings.TrimSuffix(name, ".yaml"), ".yml")
			themes = append(themes, themeName)
		}
	}

	return themes, nil
}

// LoadCustomTheme loads a custom theme by name from the themes directory.
func LoadCustomTheme(name string) (*Theme, error) {
	themesDir, err := GetThemesDir()
	if err != nil {
		return nil, err
	}

	// Try .yaml first, then .yml
	for _, ext := range []string{".yaml", ".yml"} {
		path := filepath.Join(themesDir, name+ext)
		if theme, err := Load(path); err == nil {
			return theme, nil
		}
	}

	return nil, os.ErrNotExist
}
