package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/ui/layout"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
)

// SidebarSection represents a section divider in the sidebar.
type SidebarSection struct {
	Label string
}

// SidebarItem represents an item in the sidebar.
type SidebarItem struct {
	ID    View
	Label string
	Icon  string
}

// Sidebar manages the sidebar navigation.
type Sidebar struct {
	items       []SidebarItem
	sections    map[int]SidebarSection // index -> section header before this item
	selected    int
	pomodoroWidget string // Pomodoro status widget content
}

// NewSidebar creates a new sidebar with default items.
func NewSidebar() *Sidebar {
	s := &Sidebar{
		items: []SidebarItem{
			{ID: ViewDashboard, Label: "Dashboard", Icon: ""},
			{ID: ViewCalendar, Label: "Calendar", Icon: ""},
			{ID: ViewGoals, Label: "Goals", Icon: ""},
			{ID: ViewCourses, Label: "Courses", Icon: ""},
			{ID: ViewBooks, Label: "Books", Icon: ""},
			{ID: ViewWishlist, Label: "Wishlist", Icon: ""},
			{ID: ViewGraph, Label: "Graph", Icon: ""},
			{ID: ViewStats, Label: "Stats", Icon: ""},
			{ID: ViewSettings, Label: "Settings", Icon: ""},
		},
		sections: make(map[int]SidebarSection),
		selected: 0,
	}

	// Add section divider before Settings
	s.sections[8] = SidebarSection{Label: ""}

	return s
}

// MoveDown moves selection down.
func (s *Sidebar) MoveDown() {
	if s.selected < len(s.items)-1 {
		s.selected++
	}
}

// MoveUp moves selection up.
func (s *Sidebar) MoveUp() {
	if s.selected > 0 {
		s.selected--
	}
}

// MoveToTop moves selection to the first item.
func (s *Sidebar) MoveToTop() {
	s.selected = 0
}

// MoveToBottom moves selection to the last item.
func (s *Sidebar) MoveToBottom() {
	s.selected = len(s.items) - 1
}

// SetIndex sets the selected index directly.
func (s *Sidebar) SetIndex(index int) {
	if index >= 0 && index < len(s.items) {
		s.selected = index
	}
}

// CurrentView returns the currently selected view.
func (s *Sidebar) CurrentView() View {
	return s.items[s.selected].ID
}

// Selected returns the currently selected index.
func (s *Sidebar) Selected() int {
	return s.selected
}

// SetPomodoroWidget sets the pomodoro widget content.
func (s *Sidebar) SetPomodoroWidget(content string) {
	s.pomodoroWidget = content
}

// Render renders the sidebar with proper borders using the new layout system.
func (s *Sidebar) Render(width, height int, focused bool) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	// Create panel for the sidebar
	panel := layout.NewPanel("LazyObsidian", width, height)
	panel.Border = layout.BorderRounded
	panel.Focused = focused

	// Apply theme colors
	if theme.Current != nil {
		panel.BorderColor = theme.Current.Color("border_default")
		panel.FocusedBorderColor = theme.Current.Color("border_active")
		panel.TitleColor = theme.Current.Color("text_primary")
		panel.BackgroundColor = theme.Current.Color("bg_secondary")
	}

	// Render content
	content := s.renderContent(panel.ContentWidth(), panel.ContentHeight(), focused)
	panel.SetContent(content)

	return panel.Render()
}

// renderContent renders the sidebar items without borders.
func (s *Sidebar) renderContent(width, height int, focused bool) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	var lines []string

	// Calculate space for items and widgets
	pomodoroHeight := 0
	if s.pomodoroWidget != "" {
		pomodoroHeight = 4 // Timer widget height
	}
	separatorCount := len(s.sections)
	itemsHeight := height - pomodoroHeight - separatorCount

	// Render navigation items
	for i, item := range s.items {
		if len(lines) >= itemsHeight {
			break
		}

		// Add section separator if needed
		if section, ok := s.sections[i]; ok {
			sep := s.renderSeparator(width, section.Label)
			lines = append(lines, sep)
		}

		line := s.renderItem(item, i == s.selected, focused, width)
		lines = append(lines, line)
	}

	// Pad remaining space before pomodoro widget
	for len(lines) < height-pomodoroHeight {
		lines = append(lines, strings.Repeat(" ", width))
	}

	// Add pomodoro widget if present
	if s.pomodoroWidget != "" {
		// Separator before pomodoro
		lines = append(lines, s.renderSeparator(width, ""))

		// Render pomodoro widget
		pomLines := strings.Split(s.pomodoroWidget, "\n")
		for _, pl := range pomLines {
			if len(lines) >= height {
				break
			}
			lines = append(lines, fitToWidth(pl, width))
		}
	}

	// Ensure we have exactly height lines
	for len(lines) < height {
		lines = append(lines, strings.Repeat(" ", width))
	}
	if len(lines) > height {
		lines = lines[:height]
	}

	return strings.Join(lines, "\n")
}

// renderItem renders a single sidebar item.
func (s *Sidebar) renderItem(item SidebarItem, selected, focused bool, width int) string {
	// Get styles from theme
	var style lipgloss.Style
	if selected && focused {
		style = theme.S.SidebarFocused
	} else if selected {
		style = theme.S.SidebarSelected
	} else {
		style = theme.S.SidebarItem
	}

	// Build content
	icon := item.Icon
	if icon == "" {
		icon = " " // placeholder for alignment
	}

	indicator := "  "
	if selected {
		indicator = " ▶"
	}

	// Calculate available width for label
	// Format: " icon  label  indicator "
	padding := 4 // spaces around content
	iconWidth := lipgloss.Width(icon)
	indicatorWidth := lipgloss.Width(indicator)
	labelWidth := width - padding - iconWidth - indicatorWidth

	// Truncate label if needed
	label := item.Label
	if lipgloss.Width(label) > labelWidth {
		label = truncateToWidth(label, labelWidth-1) + "…"
	}

	// Pad label to fill space
	labelPadded := label + strings.Repeat(" ", max(0, labelWidth-lipgloss.Width(label)))

	// Build the line
	content := " " + icon + " " + labelPadded + indicator + " "

	// Apply style and ensure exact width
	styled := style.Render(content)

	// Ensure the line is exactly the right width
	return fitToWidth(styled, width)
}

// renderSeparator renders a section separator.
func (s *Sidebar) renderSeparator(width int, label string) string {
	if width <= 0 {
		return ""
	}

	separatorChar := "─"
	style := lipgloss.NewStyle().Foreground(theme.Current.Color("border_muted"))

	if label == "" {
		return style.Render(strings.Repeat(separatorChar, width))
	}

	// Label in the middle
	label = " " + label + " "
	labelWidth := lipgloss.Width(label)
	sideWidth := (width - labelWidth) / 2

	left := strings.Repeat(separatorChar, sideWidth)
	right := strings.Repeat(separatorChar, width-sideWidth-labelWidth)

	return style.Render(left) + label + style.Render(right)
}

// fitToWidth ensures a string is exactly the given visual width.
// Delegates to layout.FitToWidth for consistent handling.
func fitToWidth(s string, width int) string {
	return layout.FitToWidth(s, width)
}

// Note: truncateToWidth is defined in app.go
