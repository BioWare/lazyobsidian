package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// BorderStyle defines the visual style of panel borders.
type BorderStyle int

const (
	// BorderNone renders no border.
	BorderNone BorderStyle = iota
	// BorderSingle uses single-line characters: ─│┌┐└┘
	BorderSingle
	// BorderDouble uses double-line characters: ═║╔╗╚╝
	BorderDouble
	// BorderRounded uses rounded corners: ─│╭╮╰╯
	BorderRounded
	// BorderHidden uses spaces for invisible borders (maintains spacing).
	BorderHidden
)

// BorderChars holds the characters for drawing borders.
type BorderChars struct {
	Horizontal rune // ─
	Vertical   rune // │
	TopLeft    rune // ┌
	TopRight   rune // ┐
	BottomLeft rune // └
	BottomRight rune // ┘
}

// GetBorderChars returns the border characters for a given style.
func GetBorderChars(style BorderStyle) BorderChars {
	switch style {
	case BorderSingle:
		return BorderChars{'─', '│', '┌', '┐', '└', '┘'}
	case BorderDouble:
		return BorderChars{'═', '║', '╔', '╗', '╚', '╝'}
	case BorderRounded:
		return BorderChars{'─', '│', '╭', '╮', '╰', '╯'}
	case BorderHidden:
		return BorderChars{' ', ' ', ' ', ' ', ' ', ' '}
	default:
		return BorderChars{' ', ' ', ' ', ' ', ' ', ' '}
	}
}

// Panel represents a bordered panel with a title and content.
type Panel struct {
	// Title displayed in the top border.
	Title string
	// Content to render inside the panel.
	Content string
	// Width of the panel including borders.
	Width int
	// Height of the panel including borders.
	Height int
	// Border style.
	Border BorderStyle
	// Focused indicates if this panel has focus.
	Focused bool

	// Styling
	BorderColor       lipgloss.Color
	FocusedBorderColor lipgloss.Color
	TitleColor        lipgloss.Color
	ContentColor      lipgloss.Color
	BackgroundColor   lipgloss.Color
}

// NewPanel creates a new panel with default styling.
func NewPanel(title string, width, height int) *Panel {
	return &Panel{
		Title:             title,
		Width:             width,
		Height:            height,
		Border:            BorderRounded,
		Focused:           false,
		BorderColor:       lipgloss.Color("#D4C9B5"),
		FocusedBorderColor: lipgloss.Color("#2D5A7B"),
		TitleColor:        lipgloss.Color("#3D3428"),
		ContentColor:      lipgloss.Color("#3D3428"),
		BackgroundColor:   lipgloss.Color("#F5F0E6"),
	}
}

// SetContent sets the panel content.
func (p *Panel) SetContent(content string) *Panel {
	p.Content = content
	return p
}

// SetTitle sets the panel title.
func (p *Panel) SetTitle(title string) *Panel {
	p.Title = title
	return p
}

// SetBorder sets the border style.
func (p *Panel) SetBorder(style BorderStyle) *Panel {
	p.Border = style
	return p
}

// SetFocused sets the focused state.
func (p *Panel) SetFocused(focused bool) *Panel {
	p.Focused = focused
	return p
}

// ContentWidth returns the available width for content (excluding borders).
func (p *Panel) ContentWidth() int {
	if p.Border == BorderNone {
		return p.Width
	}
	return p.Width - 2 // 1 char border on each side
}

// ContentHeight returns the available height for content (excluding borders).
func (p *Panel) ContentHeight() int {
	if p.Border == BorderNone {
		return p.Height
	}
	return p.Height - 2 // top and bottom border
}

// Render renders the panel to a string.
func (p *Panel) Render() string {
	if p.Width <= 0 || p.Height <= 0 {
		return ""
	}

	// No border - just render content
	if p.Border == BorderNone {
		return p.renderContent(p.Width, p.Height)
	}

	// Get border characters (use double for focused)
	borderStyle := p.Border
	if p.Focused && p.Border != BorderNone && p.Border != BorderHidden {
		borderStyle = BorderDouble
	}
	chars := GetBorderChars(borderStyle)

	// Get border color
	borderColor := p.BorderColor
	if p.Focused {
		borderColor = p.FocusedBorderColor
	}
	borderStyle_ := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(p.TitleColor).Bold(true)

	var lines []string

	// Top border with title
	topBorder := p.renderTopBorder(chars, borderStyle_, titleStyle)
	lines = append(lines, topBorder)

	// Content lines
	contentLines := p.renderContentLines(chars, borderStyle_)
	lines = append(lines, contentLines...)

	// Bottom border
	bottomBorder := p.renderBottomBorder(chars, borderStyle_)
	lines = append(lines, bottomBorder)

	return strings.Join(lines, "\n")
}

// renderTopBorder renders the top border with title.
func (p *Panel) renderTopBorder(chars BorderChars, borderStyle, titleStyle lipgloss.Style) string {
	if p.Width < 2 {
		return ""
	}

	innerWidth := p.Width - 2 // Space between corners

	// Calculate title portion
	title := p.Title
	if title != "" {
		title = " " + title + " "
	}
	titleWidth := lipgloss.Width(title)

	// Format: ╭─ Title ────────────────────────────╮
	leftPad := 1 // one dash after corner

	// Calculate available space for title
	maxTitleWidth := innerWidth - leftPad - 1 // at least 1 dash on right
	if titleWidth > maxTitleWidth && maxTitleWidth > 3 {
		// Truncate title
		title = " " + TruncateWithEllipsis(p.Title, maxTitleWidth-3) + " "
		titleWidth = lipgloss.Width(title)
	} else if titleWidth > maxTitleWidth {
		title = ""
		titleWidth = 0
	}

	rightPad := innerWidth - leftPad - titleWidth
	if rightPad < 0 {
		rightPad = 0
	}

	var line strings.Builder
	line.WriteString(borderStyle.Render(string(chars.TopLeft)))
	line.WriteString(borderStyle.Render(strings.Repeat(string(chars.Horizontal), leftPad)))
	if titleWidth > 0 {
		line.WriteString(titleStyle.Render(title))
	}
	line.WriteString(borderStyle.Render(strings.Repeat(string(chars.Horizontal), rightPad)))
	line.WriteString(borderStyle.Render(string(chars.TopRight)))

	return FitToWidth(line.String(), p.Width)
}

// renderContentLines renders the content area with side borders.
func (p *Panel) renderContentLines(chars BorderChars, borderStyle lipgloss.Style) []string {
	contentHeight := p.ContentHeight()
	contentWidth := p.ContentWidth()

	if contentHeight <= 0 || contentWidth <= 0 {
		return nil
	}

	// Split content into lines
	content := p.Content
	if content == "" {
		content = strings.Repeat("\n", contentHeight-1)
	}

	rawLines := strings.Split(content, "\n")

	// Create content style with background
	contentStyle := lipgloss.NewStyle().Background(p.BackgroundColor)

	lines := make([]string, contentHeight)
	for i := 0; i < contentHeight; i++ {
		var lineContent string
		if i < len(rawLines) {
			lineContent = rawLines[i]
		}

		// Ensure proper width using FitToWidth
		lineContent = FitToWidth(lineContent, contentWidth)

		// Apply background to content
		styledContent := contentStyle.Render(lineContent)

		// Build line with borders
		var line strings.Builder
		line.WriteString(borderStyle.Render(string(chars.Vertical)))
		line.WriteString(styledContent)
		line.WriteString(borderStyle.Render(string(chars.Vertical)))

		// Ensure the complete line is exactly the right width
		lines[i] = FitToWidth(line.String(), p.Width)
	}

	return lines
}

// renderBottomBorder renders the bottom border.
func (p *Panel) renderBottomBorder(chars BorderChars, borderStyle lipgloss.Style) string {
	if p.Width < 2 {
		return ""
	}

	innerWidth := p.Width - 2
	var line strings.Builder
	line.WriteString(borderStyle.Render(string(chars.BottomLeft)))
	line.WriteString(borderStyle.Render(strings.Repeat(string(chars.Horizontal), innerWidth)))
	line.WriteString(borderStyle.Render(string(chars.BottomRight)))

	return FitToWidth(line.String(), p.Width)
}

// renderContent renders just the content without borders.
func (p *Panel) renderContent(width, height int) string {
	rawLines := strings.Split(p.Content, "\n")

	var lines []string
	for i := 0; i < height; i++ {
		var lineContent string
		if i < len(rawLines) {
			lineContent = rawLines[i]
		}
		lineContent = fitToWidth(lineContent, width)
		lines = append(lines, lineContent)
	}

	return strings.Join(lines, "\n")
}

// fitToWidth ensures a string is exactly the given width.
// Delegates to the screen.go FitToWidth function.
func fitToWidth(s string, width int) string {
	return FitToWidth(s, width)
}

// truncateWithWidth truncates a string to fit within maxWidth.
// Delegates to the screen.go TruncateToWidth function.
func truncateWithWidth(s string, maxWidth int) string {
	return TruncateToWidth(s, maxWidth)
}

// runeWidth returns the display width of a rune.
// Delegates to the screen.go RuneWidth function.
func runeWidth(r rune) int {
	return RuneWidth(r)
}

// PanelGroup renders multiple panels in a row or column.
type PanelGroup struct {
	Direction Direction
	Panels    []*Panel
	Weights   []int // Weight for each panel, nil means equal weights
	Spacing   int   // Spacing between panels
}

// NewPanelGroup creates a new panel group.
func NewPanelGroup(direction Direction) *PanelGroup {
	return &PanelGroup{
		Direction: direction,
		Spacing:   0,
	}
}

// Add adds a panel with optional weight.
func (g *PanelGroup) Add(panel *Panel, weight int) *PanelGroup {
	g.Panels = append(g.Panels, panel)
	g.Weights = append(g.Weights, weight)
	return g
}

// Render renders the panel group to a string.
func (g *PanelGroup) Render(width, height int) string {
	if len(g.Panels) == 0 {
		return ""
	}

	// Calculate sizes
	sizes := g.calculateSizes(width, height)

	if g.Direction == Column {
		return g.renderHorizontal(sizes, height)
	}
	return g.renderVertical(sizes, width)
}

// calculateSizes calculates the size for each panel.
func (g *PanelGroup) calculateSizes(width, height int) []int {
	n := len(g.Panels)
	if n == 0 {
		return nil
	}

	totalSpace := width
	if g.Direction == Row {
		totalSpace = height
	}

	// Subtract spacing
	totalSpace -= g.Spacing * (n - 1)
	if totalSpace < 0 {
		totalSpace = 0
	}

	// Calculate total weight
	totalWeight := 0
	for i := 0; i < n; i++ {
		w := 1
		if i < len(g.Weights) && g.Weights[i] > 0 {
			w = g.Weights[i]
		}
		totalWeight += w
	}

	// Distribute space
	sizes := make([]int, n)
	remaining := totalSpace

	for i := 0; i < n; i++ {
		w := 1
		if i < len(g.Weights) && g.Weights[i] > 0 {
			w = g.Weights[i]
		}

		if i == n-1 {
			// Last panel gets remaining space
			sizes[i] = remaining
		} else {
			size := totalSpace * w / totalWeight
			sizes[i] = size
			remaining -= size
		}
	}

	return sizes
}

// renderHorizontal renders panels side by side.
func (g *PanelGroup) renderHorizontal(sizes []int, height int) string {
	var rendered []string

	for i, panel := range g.Panels {
		panel.Width = sizes[i]
		panel.Height = height
		rendered = append(rendered, panel.Render())
	}

	// Join horizontally with lipgloss
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// renderVertical renders panels stacked vertically.
func (g *PanelGroup) renderVertical(sizes []int, width int) string {
	var rendered []string

	for i, panel := range g.Panels {
		panel.Width = width
		panel.Height = sizes[i]
		rendered = append(rendered, panel.Render())
	}

	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}
