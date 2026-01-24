// Package layout provides a constraint-based layout system for TUI components.
package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Frame renders content within a bordered frame with exact dimensions.
// This is the recommended way to create bordered panels that don't "float".
type Frame struct {
	// Dimensions
	Width  int
	Height int

	// Title in the top border
	Title string

	// Border style
	Border BorderStyle

	// Focus state - affects border style
	Focused bool

	// Colors (as hex strings or lipgloss.Color)
	BorderColor        lipgloss.Color
	FocusedBorderColor lipgloss.Color
	TitleColor         lipgloss.Color
	BackgroundColor    lipgloss.Color

	// Content to render inside the frame
	content []string
}

// NewFrame creates a new frame with the given dimensions.
func NewFrame(width, height int) *Frame {
	return &Frame{
		Width:              width,
		Height:             height,
		Border:             BorderRounded,
		BorderColor:        lipgloss.Color("#D4C9B5"),
		FocusedBorderColor: lipgloss.Color("#2D5A7B"),
		TitleColor:         lipgloss.Color("#3D3428"),
		BackgroundColor:    lipgloss.Color("#F5F0E6"),
	}
}

// SetTitle sets the frame title.
func (f *Frame) SetTitle(title string) *Frame {
	f.Title = title
	return f
}

// SetBorder sets the border style.
func (f *Frame) SetBorder(style BorderStyle) *Frame {
	f.Border = style
	return f
}

// SetFocused sets the focused state.
func (f *Frame) SetFocused(focused bool) *Frame {
	f.Focused = focused
	return f
}

// SetColors sets all colors at once.
func (f *Frame) SetColors(border, focusedBorder, title, bg lipgloss.Color) *Frame {
	f.BorderColor = border
	f.FocusedBorderColor = focusedBorder
	f.TitleColor = title
	f.BackgroundColor = bg
	return f
}

// ContentWidth returns the available width for content (excluding borders).
func (f *Frame) ContentWidth() int {
	if f.Border == BorderNone {
		return f.Width
	}
	return max(0, f.Width-2)
}

// ContentHeight returns the available height for content (excluding borders).
func (f *Frame) ContentHeight() int {
	if f.Border == BorderNone {
		return f.Height
	}
	return max(0, f.Height-2)
}

// SetContent sets the content from a multi-line string.
func (f *Frame) SetContent(content string) *Frame {
	f.content = SplitIntoLines(content)
	return f
}

// SetContentLines sets the content from a slice of lines.
func (f *Frame) SetContentLines(lines []string) *Frame {
	f.content = lines
	return f
}

// Render renders the frame to a string.
// Every line is guaranteed to be exactly Width characters (visual width).
func (f *Frame) Render() string {
	if f.Width <= 0 || f.Height <= 0 {
		return ""
	}

	if f.Border == BorderNone {
		return f.renderNoBorder()
	}

	// Minimum size for border: 2x2
	if f.Width < 2 || f.Height < 2 {
		return f.renderNoBorder()
	}

	lines := make([]string, 0, f.Height)

	// Get border characters (use double for focused)
	borderStyle := f.Border
	if f.Focused && f.Border != BorderHidden {
		borderStyle = BorderDouble
	}
	chars := GetBorderChars(borderStyle)

	// Get border color
	borderColor := f.BorderColor
	if f.Focused {
		borderColor = f.FocusedBorderColor
	}
	borderStyle_ := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(f.TitleColor).Bold(true)

	// Top border with title
	topBorder := f.renderTopBorder(chars, borderStyle_, titleStyle)
	lines = append(lines, topBorder)

	// Content lines
	contentLines := f.renderContentLines(chars, borderStyle_)
	lines = append(lines, contentLines...)

	// Bottom border
	bottomBorder := f.renderBottomBorder(chars, borderStyle_)
	lines = append(lines, bottomBorder)

	return strings.Join(lines, "\n")
}

// renderTopBorder renders the top border with title.
func (f *Frame) renderTopBorder(chars BorderChars, borderStyle, titleStyle lipgloss.Style) string {
	innerWidth := f.Width - 2 // Space between corners

	// Format: ╭─ Title ────────────────────────────╮
	title := f.Title
	if title != "" {
		title = " " + title + " "
	}

	titleWidth := lipgloss.Width(title)

	// Calculate padding
	leftPad := 1 // One dash after left corner

	// Check if title fits
	maxTitleWidth := innerWidth - leftPad - 1 // Leave at least 1 dash on right
	if titleWidth > maxTitleWidth && maxTitleWidth > 3 {
		// Truncate title
		title = " " + TruncateWithEllipsis(f.Title, maxTitleWidth-3) + " "
		titleWidth = lipgloss.Width(title)
	} else if titleWidth > maxTitleWidth {
		// No room for title
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

	return FitToWidth(line.String(), f.Width)
}

// renderContentLines renders the content area with side borders.
func (f *Frame) renderContentLines(chars BorderChars, borderStyle lipgloss.Style) []string {
	contentHeight := f.ContentHeight()
	contentWidth := f.ContentWidth()

	if contentHeight <= 0 || contentWidth <= 0 {
		return nil
	}

	// Create background style
	bgStyle := lipgloss.NewStyle().Background(f.BackgroundColor)

	lines := make([]string, contentHeight)

	for i := 0; i < contentHeight; i++ {
		var lineContent string
		if i < len(f.content) {
			lineContent = f.content[i]
		}

		// Fit content to exact width
		lineContent = FitToWidth(lineContent, contentWidth)

		// Apply background
		styledContent := bgStyle.Render(lineContent)

		// Build complete line with borders
		var line strings.Builder
		line.WriteString(borderStyle.Render(string(chars.Vertical)))
		line.WriteString(styledContent)
		line.WriteString(borderStyle.Render(string(chars.Vertical)))

		lines[i] = FitToWidth(line.String(), f.Width)
	}

	return lines
}

// renderBottomBorder renders the bottom border.
func (f *Frame) renderBottomBorder(chars BorderChars, borderStyle lipgloss.Style) string {
	innerWidth := f.Width - 2

	var line strings.Builder
	line.WriteString(borderStyle.Render(string(chars.BottomLeft)))
	line.WriteString(borderStyle.Render(strings.Repeat(string(chars.Horizontal), innerWidth)))
	line.WriteString(borderStyle.Render(string(chars.BottomRight)))

	return FitToWidth(line.String(), f.Width)
}

// renderNoBorder renders content without a border.
func (f *Frame) renderNoBorder() string {
	lines := make([]string, f.Height)
	bgStyle := lipgloss.NewStyle().Background(f.BackgroundColor)

	for i := 0; i < f.Height; i++ {
		var lineContent string
		if i < len(f.content) {
			lineContent = f.content[i]
		}
		lineContent = FitToWidth(lineContent, f.Width)
		lines[i] = bgStyle.Render(lineContent)
	}

	return strings.Join(lines, "\n")
}

// Layout renders multiple frames side by side or stacked.
type Layout struct {
	Direction Direction
	Gap       int
	frames    []*LayoutItem
}

// LayoutItem represents a frame in a layout with its size specification.
type LayoutItem struct {
	Frame  *Frame
	Size   int // Fixed size (0 = flexible)
	Weight int // Flex weight (only used if Size == 0)
}

// NewLayout creates a new layout.
func NewLayout(direction Direction) *Layout {
	return &Layout{
		Direction: direction,
	}
}

// AddFrame adds a frame with a flex weight.
func (l *Layout) AddFrame(frame *Frame, weight int) *Layout {
	l.frames = append(l.frames, &LayoutItem{
		Frame:  frame,
		Weight: weight,
	})
	return l
}

// AddFrameFixed adds a frame with a fixed size.
func (l *Layout) AddFrameFixed(frame *Frame, size int) *Layout {
	l.frames = append(l.frames, &LayoutItem{
		Frame: frame,
		Size:  size,
	})
	return l
}

// Render renders the layout to a string.
func (l *Layout) Render(width, height int) string {
	if len(l.frames) == 0 || width <= 0 || height <= 0 {
		return ""
	}

	// Calculate sizes for each item
	sizes := l.calculateSizes(width, height)

	// Render each frame
	var rendered []string
	for i, item := range l.frames {
		if l.Direction == Column {
			item.Frame.Width = sizes[i]
			item.Frame.Height = height
		} else {
			item.Frame.Width = width
			item.Frame.Height = sizes[i]
		}
		rendered = append(rendered, item.Frame.Render())
	}

	// Join frames
	if l.Direction == Column {
		return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

// calculateSizes calculates the size for each item.
func (l *Layout) calculateSizes(width, height int) []int {
	n := len(l.frames)
	if n == 0 {
		return nil
	}

	totalSpace := height
	if l.Direction == Column {
		totalSpace = width
	}

	// Account for gaps
	totalSpace -= l.Gap * (n - 1)
	if totalSpace < 0 {
		totalSpace = 0
	}

	sizes := make([]int, n)

	// First pass: allocate fixed sizes
	fixedTotal := 0
	flexTotal := 0

	for i, item := range l.frames {
		if item.Size > 0 {
			sizes[i] = item.Size
			fixedTotal += item.Size
		} else {
			weight := item.Weight
			if weight <= 0 {
				weight = 1
			}
			flexTotal += weight
		}
	}

	// Second pass: distribute remaining space
	remaining := totalSpace - fixedTotal
	if remaining < 0 {
		remaining = 0
	}

	for i, item := range l.frames {
		if item.Size == 0 {
			weight := item.Weight
			if weight <= 0 {
				weight = 1
			}
			if flexTotal > 0 {
				sizes[i] = remaining * weight / flexTotal
			}
		}
	}

	// Adjust last flex item to use any remaining pixels
	if flexTotal > 0 {
		usedByFlex := 0
		lastFlexIdx := -1
		for i, item := range l.frames {
			if item.Size == 0 {
				usedByFlex += sizes[i]
				lastFlexIdx = i
			}
		}
		if lastFlexIdx >= 0 && usedByFlex < remaining {
			sizes[lastFlexIdx] += remaining - usedByFlex
		}
	}

	return sizes
}

// Note: max function is defined in layout.go
