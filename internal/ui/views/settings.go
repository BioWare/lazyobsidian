// Package views implements the UI views for LazyObsidian.
package views

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/i18n"
	"github.com/BioWare/lazyobsidian/internal/ui/icons"
	"github.com/BioWare/lazyobsidian/internal/ui/layout"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
)

// SettingsPage represents a settings category.
type SettingsPage int

const (
	SettingsPageVault SettingsPage = iota
	SettingsPagePomodoro
	SettingsPageTasks
	SettingsPageDisplay
	SettingsPageKeybindings
)

// SettingType represents the type of a setting.
type SettingType int

const (
	SettingTypeString SettingType = iota
	SettingTypeInt
	SettingTypeBool
	SettingTypeSelect
	SettingTypePath
)

// Setting represents a single setting item.
type Setting struct {
	Key         string
	Label       string
	Description string
	Type        SettingType
	Value       interface{}
	Options     []string // for select type
	Min         int      // for int type
	Max         int      // for int type
}

// SettingsCategory represents a group of settings.
type SettingsCategory struct {
	Name     string
	Icon     string
	Settings []Setting
}

// SettingsView represents the settings view.
type SettingsView struct {
	Width  int
	Height int

	// Data
	Categories []SettingsCategory

	// UI state
	SelectedCategory int
	SelectedSetting  int
	ScrollOffset     int
	Focused          bool
	EditMode         bool
	EditValue        string
}

// NewSettingsView creates a new settings view.
func NewSettingsView(width, height int) *SettingsView {
	v := &SettingsView{
		Width:  width,
		Height: height,
	}
	v.initCategories()
	return v
}

func (s *SettingsView) initCategories() {
	s.Categories = []SettingsCategory{
		{
			Name: "Vault",
			Icon: "folder",
			Settings: []Setting{
				{Key: "vault_path", Label: "Vault Path", Description: "Path to your Obsidian vault", Type: SettingTypePath, Value: ""},
				{Key: "daily_folder", Label: "Daily Notes Folder", Description: "Folder for daily notes", Type: SettingTypeString, Value: "Daily"},
				{Key: "goals_folder", Label: "Goals Folder", Description: "Folder for goal notes", Type: SettingTypeString, Value: "Goals"},
				{Key: "courses_folder", Label: "Courses Folder", Description: "Folder for course notes", Type: SettingTypeString, Value: "Courses"},
				{Key: "books_folder", Label: "Books Folder", Description: "Folder for book notes", Type: SettingTypeString, Value: "Books"},
			},
		},
		{
			Name: "Pomodoro",
			Icon: "pomodoro",
			Settings: []Setting{
				{Key: "work_duration", Label: "Work Duration", Description: "Minutes per work session", Type: SettingTypeInt, Value: 25, Min: 1, Max: 120},
				{Key: "short_break", Label: "Short Break", Description: "Minutes for short break", Type: SettingTypeInt, Value: 5, Min: 1, Max: 30},
				{Key: "long_break", Label: "Long Break", Description: "Minutes for long break", Type: SettingTypeInt, Value: 15, Min: 1, Max: 60},
				{Key: "sessions_before_long", Label: "Sessions Before Long Break", Description: "Work sessions before long break", Type: SettingTypeInt, Value: 4, Min: 1, Max: 10},
				{Key: "daily_goal", Label: "Daily Goal", Description: "Target pomodoros per day", Type: SettingTypeInt, Value: 8, Min: 1, Max: 20},
				{Key: "auto_start_breaks", Label: "Auto-start Breaks", Description: "Automatically start break timer", Type: SettingTypeBool, Value: false},
				{Key: "sound_enabled", Label: "Sound Notifications", Description: "Play sound when timer ends", Type: SettingTypeBool, Value: true},
			},
		},
		{
			Name: "Tasks",
			Icon: "tasks",
			Settings: []Setting{
				{Key: "show_completed", Label: "Show Completed Tasks", Description: "Display completed tasks in lists", Type: SettingTypeBool, Value: false},
				{Key: "show_cancelled", Label: "Show Cancelled Tasks", Description: "Display cancelled tasks", Type: SettingTypeBool, Value: false},
				{Key: "sort_by", Label: "Sort Tasks By", Description: "Default task sorting", Type: SettingTypeSelect, Value: "priority", Options: []string{"priority", "date", "status", "name"}},
				{Key: "confirm_complete", Label: "Confirm Complete", Description: "Ask before marking task done", Type: SettingTypeBool, Value: false},
			},
		},
		{
			Name: "Display",
			Icon: "display",
			Settings: []Setting{
				{Key: "theme", Label: "Theme", Description: "Color theme", Type: SettingTypeSelect, Value: "corsair-light", Options: []string{"corsair-light", "corsair-dark", "dracula", "nord"}},
				{Key: "icon_mode", Label: "Icon Mode", Description: "Icon style to use", Type: SettingTypeSelect, Value: "emoji", Options: []string{"emoji", "nerd", "nerd_minimal", "ascii"}},
				{Key: "language", Label: "Language", Description: "Interface language", Type: SettingTypeSelect, Value: "en", Options: []string{"en", "ru", "es", "de", "fr"}},
				{Key: "show_borders", Label: "Show Borders", Description: "Display panel borders", Type: SettingTypeBool, Value: true},
				{Key: "compact_mode", Label: "Compact Mode", Description: "Reduce spacing for small screens", Type: SettingTypeBool, Value: false},
			},
		},
		{
			Name: "Keybindings",
			Icon: "keyboard",
			Settings: []Setting{
				{Key: "vim_mode", Label: "Vim Mode", Description: "Enable vim-style navigation", Type: SettingTypeBool, Value: true},
				{Key: "key_quit", Label: "Quit Key", Description: "Key to quit application", Type: SettingTypeString, Value: "q"},
				{Key: "key_help", Label: "Help Key", Description: "Key to show help", Type: SettingTypeString, Value: "?"},
				{Key: "key_focus_sidebar", Label: "Focus Sidebar", Description: "Key to focus sidebar", Type: SettingTypeString, Value: "1"},
				{Key: "key_focus_main", Label: "Focus Main Panel", Description: "Key to focus main panel", Type: SettingTypeString, Value: "2"},
			},
		},
	}
}

// SetSize updates the view dimensions.
func (s *SettingsView) SetSize(width, height int) {
	s.Width = width
	s.Height = height
}

// SetFocused sets the focus state.
func (s *SettingsView) SetFocused(focused bool) {
	s.Focused = focused
}

// SelectNextCategory moves to the next category.
func (s *SettingsView) SelectNextCategory() {
	if s.SelectedCategory < len(s.Categories)-1 {
		s.SelectedCategory++
		s.SelectedSetting = 0
		s.ScrollOffset = 0
	}
}

// SelectPrevCategory moves to the previous category.
func (s *SettingsView) SelectPrevCategory() {
	if s.SelectedCategory > 0 {
		s.SelectedCategory--
		s.SelectedSetting = 0
		s.ScrollOffset = 0
	}
}

// SelectNextSetting moves to the next setting.
func (s *SettingsView) SelectNextSetting() {
	if s.SelectedCategory >= 0 && s.SelectedCategory < len(s.Categories) {
		cat := s.Categories[s.SelectedCategory]
		if s.SelectedSetting < len(cat.Settings)-1 {
			s.SelectedSetting++
			s.ensureVisible()
		}
	}
}

// SelectPrevSetting moves to the previous setting.
func (s *SettingsView) SelectPrevSetting() {
	if s.SelectedSetting > 0 {
		s.SelectedSetting--
		s.ensureVisible()
	}
}

// ToggleCurrentSetting toggles a boolean setting or cycles select options.
func (s *SettingsView) ToggleCurrentSetting() {
	setting := s.currentSetting()
	if setting == nil {
		return
	}

	switch setting.Type {
	case SettingTypeBool:
		if v, ok := setting.Value.(bool); ok {
			setting.Value = !v
		}
	case SettingTypeSelect:
		if v, ok := setting.Value.(string); ok {
			for i, opt := range setting.Options {
				if opt == v {
					nextIdx := (i + 1) % len(setting.Options)
					setting.Value = setting.Options[nextIdx]
					break
				}
			}
		}
	case SettingTypeInt:
		if v, ok := setting.Value.(int); ok {
			newVal := v + 1
			if newVal > setting.Max {
				newVal = setting.Min
			}
			setting.Value = newVal
		}
	}
}

// EnterEditMode enters edit mode for the current setting.
func (s *SettingsView) EnterEditMode() {
	setting := s.currentSetting()
	if setting == nil {
		return
	}

	if setting.Type == SettingTypeString || setting.Type == SettingTypePath {
		s.EditMode = true
		s.EditValue = fmt.Sprintf("%v", setting.Value)
	}
}

// ExitEditMode exits edit mode, optionally saving the value.
func (s *SettingsView) ExitEditMode(save bool) {
	if save {
		setting := s.currentSetting()
		if setting != nil {
			setting.Value = s.EditValue
		}
	}
	s.EditMode = false
	s.EditValue = ""
}

// UpdateEditValue updates the value being edited.
func (s *SettingsView) UpdateEditValue(value string) {
	s.EditValue = value
}

func (s *SettingsView) currentSetting() *Setting {
	if s.SelectedCategory < 0 || s.SelectedCategory >= len(s.Categories) {
		return nil
	}
	cat := &s.Categories[s.SelectedCategory]
	if s.SelectedSetting < 0 || s.SelectedSetting >= len(cat.Settings) {
		return nil
	}
	return &cat.Settings[s.SelectedSetting]
}

func (s *SettingsView) ensureVisible() {
	visibleLines := s.Height - 8
	if visibleLines < 1 {
		visibleLines = 1
	}

	// Each setting takes ~2 lines
	settingLine := s.SelectedSetting * 2

	if settingLine < s.ScrollOffset {
		s.ScrollOffset = settingLine
	}
	if settingLine >= s.ScrollOffset+visibleLines {
		s.ScrollOffset = settingLine - visibleLines + 2
	}
}

// Render renders the settings view.
func (s *SettingsView) Render() string {
	th := theme.Current
	t := i18n.T()

	// Split: 25% categories, 75% settings
	catWidth := s.Width * 25 / 100
	if catWidth < 20 {
		catWidth = 20
	}
	settingsWidth := s.Width - catWidth

	// Categories panel
	catFrame := layout.NewFrame(catWidth, s.Height)
	catFrame.SetTitle(icons.Get("settings") + " " + t.Nav.Settings)
	catFrame.SetBorder(layout.BorderSingle)
	catFrame.SetFocused(s.Focused && !s.EditMode)
	catFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	catFrame.SetContentLines(s.renderCategoryList(catFrame.ContentWidth(), th))

	// Settings panel
	settingsFrame := layout.NewFrame(settingsWidth, s.Height)
	if s.SelectedCategory >= 0 && s.SelectedCategory < len(s.Categories) {
		settingsFrame.SetTitle(s.Categories[s.SelectedCategory].Name)
	}
	settingsFrame.SetBorder(layout.BorderSingle)
	settingsFrame.SetFocused(s.Focused && s.EditMode)
	settingsFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	settingsFrame.SetContentLines(s.renderSettingsList(settingsFrame.ContentWidth(), th))

	return combineHorizontal(catFrame.Render(), settingsFrame.Render())
}

func (s *SettingsView) renderCategoryList(width int, th *theme.Theme) []string {
	var lines []string

	for i, cat := range s.Categories {
		icon := icons.Get(cat.Icon)
		name := cat.Name

		var style lipgloss.Style
		if i == s.SelectedCategory && s.Focused {
			style = lipgloss.NewStyle().
				Foreground(th.Color("bg_primary")).
				Background(th.Color("accent")).
				Bold(true)
		} else if i == s.SelectedCategory {
			style = lipgloss.NewStyle().
				Foreground(th.Color("text_primary")).
				Background(th.Color("bg_secondary"))
		} else {
			style = lipgloss.NewStyle().Foreground(th.Color("text_primary"))
		}

		line := style.Render(layout.FitToWidth(icon+" "+name, width))
		lines = append(lines, line)
	}

	return lines
}

func (s *SettingsView) renderSettingsList(width int, th *theme.Theme) []string {
	var lines []string

	if s.SelectedCategory < 0 || s.SelectedCategory >= len(s.Categories) {
		return lines
	}

	cat := s.Categories[s.SelectedCategory]
	labelStyle := lipgloss.NewStyle().Foreground(th.Color("text_primary"))
	descStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted")).Italic(true)
	valueStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary"))
	selectedStyle := lipgloss.NewStyle().
		Foreground(th.Color("bg_primary")).
		Background(th.Color("accent")).
		Bold(true)

	for i, setting := range cat.Settings {
		// Skip if scrolled out
		settingLine := i * 2
		if settingLine < s.ScrollOffset {
			continue
		}
		if settingLine >= s.ScrollOffset+s.Height-8 {
			break
		}

		isSelected := i == s.SelectedSetting

		// Label line
		label := setting.Label

		// Value representation
		var valueStr string
		switch setting.Type {
		case SettingTypeBool:
			if v, ok := setting.Value.(bool); ok {
				if v {
					valueStr = icons.Get("check") + " On"
				} else {
					valueStr = icons.Get("cross") + " Off"
				}
			}
		case SettingTypeSelect:
			if v, ok := setting.Value.(string); ok {
				valueStr = "[" + v + "]"
			}
		case SettingTypeInt:
			valueStr = fmt.Sprintf("%v", setting.Value)
		case SettingTypeString, SettingTypePath:
			if s.EditMode && isSelected {
				valueStr = s.EditValue + "▏"
			} else {
				v := fmt.Sprintf("%v", setting.Value)
				if v == "" {
					valueStr = "(not set)"
				} else {
					valueStr = v
				}
			}
		}

		// Render label + value on same line
		valueWidth := lipgloss.Width(valueStr)
		labelWidth := width - valueWidth - 3
		if labelWidth < 10 {
			labelWidth = 10
		}

		var line string
		if isSelected && s.Focused {
			labelText := layout.TruncateWithEllipsis(label, labelWidth)
			line = selectedStyle.Render(layout.FitToWidth(labelText+"  "+valueStr, width))
		} else {
			labelText := labelStyle.Render(layout.TruncateWithEllipsis(label, labelWidth))
			line = labelText + "  " + valueStyle.Render(valueStr)
		}
		lines = append(lines, layout.FitToWidth(line, width))

		// Description line
		desc := layout.TruncateWithEllipsis(setting.Description, width-2)
		lines = append(lines, "  "+descStyle.Render(desc))
	}

	// Help text at bottom
	lines = append(lines, "")
	helpStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
	var helpText string
	if s.EditMode {
		helpText = "[Enter] Save  [Esc] Cancel"
	} else {
		setting := s.currentSetting()
		if setting != nil {
			switch setting.Type {
			case SettingTypeBool:
				helpText = "[Enter/Space] Toggle  [j/k] Navigate"
			case SettingTypeSelect:
				helpText = "[Enter/Space] Cycle options  [j/k] Navigate"
			case SettingTypeInt:
				helpText = "[Enter/Space] Increase  [j/k] Navigate"
			default:
				helpText = "[Enter] Edit  [j/k] Navigate"
			}
		}
	}
	lines = append(lines, helpStyle.Render(helpText))

	return lines
}
