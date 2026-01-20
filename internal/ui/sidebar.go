package ui

import (
	"fmt"
	"strings"
)

// SidebarItem represents an item in the sidebar.
type SidebarItem struct {
	ID    View
	Label string
	Icon  string
}

// Sidebar manages the sidebar navigation.
type Sidebar struct {
	items    []SidebarItem
	selected int
}

// NewSidebar creates a new sidebar with default items.
func NewSidebar() *Sidebar {
	return &Sidebar{
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
		selected: 0,
	}
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

// Render renders the sidebar.
func (s *Sidebar) Render(width, height int, focused bool) string {
	var b strings.Builder

	// Calculate available height for items
	contentHeight := height - 2 // borders

	for i, item := range s.items {
		if i >= contentHeight {
			break
		}

		line := s.renderItem(item, i == s.selected, focused, width-2)
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Pad remaining height
	for i := len(s.items); i < contentHeight; i++ {
		b.WriteString(strings.Repeat(" ", width-2))
		b.WriteString("\n")
	}

	return b.String()
}

func (s *Sidebar) renderItem(item SidebarItem, selected, focused bool, width int) string {
	icon := item.Icon
	if icon == "" {
		icon = " "
	}

	label := item.Label
	indicator := " "

	if selected {
		indicator = "▶"
	}

	// Build the line
	content := fmt.Sprintf(" %s %s %s", icon, label, indicator)

	// Pad or truncate to fit width
	if len(content) > width {
		content = content[:width-1] + "…"
	} else {
		content = content + strings.Repeat(" ", width-len(content))
	}

	// Apply styling based on selection state
	if selected && focused {
		// Highlighted: reverse colors (would use lipgloss in full implementation)
		return ">" + content[1:]
	} else if selected {
		return "│" + content[1:]
	}

	return content
}
