// Package theme handles color themes and styling for the TUI.
package theme

import (
	"os"
	"path/filepath"

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

	// Progress
	ProgressFill  lipgloss.Style
	ProgressEmpty lipgloss.Style

	// Pomodoro
	PomodoroWork   lipgloss.Style
	PomodoroBreak  lipgloss.Style
	PomorodoPaused lipgloss.Style

	// Text
	TextPrimary   lipgloss.Style
	TextSecondary lipgloss.Style
	TextMuted     lipgloss.Style

	// Semantic
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Info    lipgloss.Style

	// Help
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
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

		// Text
		TextPrimary: lipgloss.NewStyle().
			Foreground(t.Color("text_primary")),

		TextSecondary: lipgloss.NewStyle().
			Foreground(t.Color("text_secondary")),

		TextMuted: lipgloss.NewStyle().
			Foreground(t.Color("text_muted")),

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
	}
}

func defaultLightTheme() *Theme {
	colors := map[string]string{
		"bg_primary":      "#F5F0E6",
		"bg_secondary":    "#EDE8DC",
		"bg_tertiary":     "#E5DFD3",
		"bg_active":       "#DDD7C9",
		"border_default":  "#D4C9B5",
		"border_muted":    "#E0D8CA",
		"border_active":   "#2D5A7B",
		"text_primary":    "#3D3428",
		"text_secondary":  "#6B5D4D",
		"text_muted":      "#9C8B75",
		"text_inverse":    "#F5F0E6",
		"primary":         "#2D5A7B",
		"secondary":       "#7B4D2D",
		"accent":          "#8B6914",
		"success":         "#4A7C59",
		"warning":         "#B8860B",
		"error":           "#8B3A3A",
		"info":            "#4A708B",
		"pomodoro_work":   "#8B3A3A",
		"pomodoro_break":  "#4A7C59",
		"pomodoro_paused": "#B8860B",
		"progress_fill":   "#2D5A7B",
		"progress_empty":  "#D4C9B5",
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
		"bg_primary":      "#1C1915",
		"bg_secondary":    "#2A2520",
		"bg_tertiary":     "#3D3630",
		"bg_active":       "#4A433B",
		"border_default":  "#4A433B",
		"border_muted":    "#3D3630",
		"border_active":   "#5B8FB9",
		"text_primary":    "#E8E0D4",
		"text_secondary":  "#B8A898",
		"text_muted":      "#786858",
		"text_inverse":    "#1C1915",
		"primary":         "#5B8FB9",
		"secondary":       "#C17F59",
		"accent":          "#D4A934",
		"success":         "#6B9B6B",
		"warning":         "#D4A934",
		"error":           "#B85B5B",
		"info":            "#5B8FB9",
		"pomodoro_work":   "#B85B5B",
		"pomodoro_break":  "#6B9B6B",
		"pomodoro_paused": "#D4A934",
		"progress_fill":   "#5B8FB9",
		"progress_empty":  "#3D3630",
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
