// Package layout provides a constraint-based layout system for TUI components.
// Inspired by lazygit's boxlayout system, adapted for BubbleTea/lipgloss.
package layout

import (
	"github.com/charmbracelet/lipgloss"
)

// Direction specifies the layout direction.
type Direction int

const (
	// Row stacks children vertically (column-wise layout).
	Row Direction = iota
	// Column stacks children horizontally (row-wise layout).
	Column
)

// Dimensions represents the calculated position and size of a component.
type Dimensions struct {
	X0, Y0 int // Top-left corner (inclusive)
	X1, Y1 int // Bottom-right corner (inclusive)
}

// Width returns the width of the dimensions.
func (d Dimensions) Width() int {
	return d.X1 - d.X0 + 1
}

// Height returns the height of the dimensions.
func (d Dimensions) Height() int {
	return d.Y1 - d.Y0 + 1
}

// Box represents a layout box that can contain children or a window.
type Box struct {
	// Direction specifies how children are arranged.
	Direction Direction

	// Children are nested boxes.
	Children []*Box

	// ConditionalChildren allows dynamic children based on available size.
	ConditionalChildren func(width, height int) []*Box

	// ConditionalDirection allows dynamic direction based on available size.
	ConditionalDirection func(width, height int) Direction

	// Window is the name of the window/panel to render in this box.
	// If set, this is a leaf node.
	Window string

	// Size is a fixed size (height for Row, width for Column).
	// If > 0, this size is used instead of Weight.
	Size int

	// Weight is the proportional weight for distributing remaining space.
	// Only used if Size == 0. Default is 1.
	Weight int
}

// ArrangeWindows calculates dimensions for all windows in the box tree.
// Returns a map from window name to its dimensions.
func ArrangeWindows(root *Box, x0, y0, width, height int) map[string]Dimensions {
	if root == nil {
		return nil
	}

	result := make(map[string]Dimensions)

	// If this is a leaf node (has a window), return its dimensions
	if root.Window != "" {
		result[root.Window] = Dimensions{
			X0: x0,
			Y0: y0,
			X1: x0 + width - 1,
			Y1: y0 + height - 1,
		}
		return result
	}

	// Get children (possibly dynamic)
	children := root.getChildren(width, height)
	if len(children) == 0 {
		return result
	}

	// Get direction (possibly dynamic)
	direction := root.getDirection(width, height)

	// Calculate available size based on direction
	availableSize := height
	if direction == Column {
		availableSize = width
	}

	// Calculate sizes for each child
	sizes := calcSizes(children, availableSize)

	// Arrange each child
	offset := 0
	for i, child := range children {
		boxSize := sizes[i]
		if boxSize <= 0 {
			continue
		}

		var childResult map[string]Dimensions
		if direction == Column {
			childResult = ArrangeWindows(child, x0+offset, y0, boxSize, height)
		} else {
			childResult = ArrangeWindows(child, x0, y0+offset, width, boxSize)
		}

		// Merge results
		for k, v := range childResult {
			result[k] = v
		}

		offset += boxSize
	}

	return result
}

// getChildren returns the children, using ConditionalChildren if set.
func (b *Box) getChildren(width, height int) []*Box {
	if b.ConditionalChildren != nil {
		return b.ConditionalChildren(width, height)
	}
	return b.Children
}

// getDirection returns the direction, using ConditionalDirection if set.
func (b *Box) getDirection(width, height int) Direction {
	if b.ConditionalDirection != nil {
		return b.ConditionalDirection(width, height)
	}
	return b.Direction
}

// calcSizes calculates the size for each child box.
func calcSizes(boxes []*Box, availableSpace int) []int {
	if len(boxes) == 0 {
		return nil
	}

	sizes := make([]int, len(boxes))

	// First pass: allocate fixed sizes
	reservedSpace := 0
	totalWeight := 0

	for i, box := range boxes {
		if box.Size > 0 {
			// Fixed size
			size := min(box.Size, availableSpace-reservedSpace)
			sizes[i] = size
			reservedSpace += size
		} else {
			// Dynamic size - count weight
			weight := box.Weight
			if weight <= 0 {
				weight = 1 // default weight
			}
			totalWeight += weight
		}
	}

	// Second pass: distribute remaining space by weight
	remainingSpace := availableSpace - reservedSpace
	if totalWeight > 0 && remainingSpace > 0 {
		// Calculate base unit size
		unitSize := remainingSpace / totalWeight
		extraSpace := remainingSpace % totalWeight

		for i, box := range boxes {
			if box.Size <= 0 {
				weight := box.Weight
				if weight <= 0 {
					weight = 1
				}
				sizes[i] = unitSize * weight

				// Distribute extra pixels to first boxes
				if extraSpace > 0 {
					extra := min(weight, extraSpace)
					sizes[i] += extra
					extraSpace -= extra
				}
			}
		}
	}

	return sizes
}

// Renderer handles rendering of the layout.
type Renderer struct {
	// Width of the terminal.
	Width int
	// Height of the terminal.
	Height int
	// Root is the root box of the layout tree.
	Root *Box
	// Panels maps window names to their Panel components.
	Panels map[string]*Panel
	// FocusedWindow is the name of the currently focused window.
	FocusedWindow string
}

// NewRenderer creates a new layout renderer.
func NewRenderer(width, height int) *Renderer {
	return &Renderer{
		Width:  width,
		Height: height,
		Panels: make(map[string]*Panel),
	}
}

// SetLayout sets the root box for the layout.
func (r *Renderer) SetLayout(root *Box) {
	r.Root = root
}

// AddPanel adds a panel to the renderer.
func (r *Renderer) AddPanel(name string, panel *Panel) {
	r.Panels[name] = panel
}

// SetFocus sets the focused window.
func (r *Renderer) SetFocus(name string) {
	r.FocusedWindow = name
}

// Resize updates the renderer dimensions.
func (r *Renderer) Resize(width, height int) {
	r.Width = width
	r.Height = height
}

// Render renders the entire layout to a string.
func (r *Renderer) Render() string {
	if r.Root == nil || r.Width == 0 || r.Height == 0 {
		return ""
	}

	// Calculate dimensions for all windows
	dimensions := ArrangeWindows(r.Root, 0, 0, r.Width, r.Height)

	// Create a 2D buffer for compositing
	buffer := newBuffer(r.Width, r.Height)

	// Render each panel to the buffer
	for name, dim := range dimensions {
		panel, ok := r.Panels[name]
		if !ok {
			continue
		}

		// Update panel dimensions
		panel.Width = dim.Width()
		panel.Height = dim.Height()
		panel.Focused = (name == r.FocusedWindow)

		// Render panel to string
		rendered := panel.Render()

		// Write to buffer at correct position
		buffer.write(dim.X0, dim.Y0, rendered)
	}

	return buffer.String()
}

// buffer is a 2D character buffer for compositing.
type buffer struct {
	cells  [][]rune
	width  int
	height int
}

// newBuffer creates a new buffer filled with spaces.
func newBuffer(width, height int) *buffer {
	cells := make([][]rune, height)
	for i := range cells {
		cells[i] = make([]rune, width)
		for j := range cells[i] {
			cells[i][j] = ' '
		}
	}
	return &buffer{
		cells:  cells,
		width:  width,
		height: height,
	}
}

// write writes a multi-line string to the buffer at position (x, y).
func (b *buffer) write(x, y int, content string) {
	// Use lipgloss to properly handle ANSI sequences
	lines := splitLines(content)
	for row, line := range lines {
		if y+row >= b.height {
			break
		}
		if y+row < 0 {
			continue
		}

		col := x
		// We need to render character by character, accounting for ANSI
		// For simplicity, we'll strip ANSI for width calculation
		lineWidth := lipgloss.Width(line)

		// Write the line as-is (we'll handle ANSI in final output)
		runes := []rune(stripAnsi(line))
		for i, r := range runes {
			if col+i >= b.width {
				break
			}
			if col+i < 0 {
				continue
			}
			b.cells[y+row][col+i] = r
		}

		// Pad remaining space if line is shorter than expected
		for i := len(runes); i < lineWidth && col+i < b.width; i++ {
			// Already spaces from init
		}
	}
}

// String converts the buffer to a string.
func (b *buffer) String() string {
	var result string
	for i, row := range b.cells {
		result += string(row)
		if i < len(b.cells)-1 {
			result += "\n"
		}
	}
	return result
}

// splitLines splits a string by newlines.
func splitLines(s string) []string {
	var lines []string
	var current string
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" || len(s) == 0 {
		lines = append(lines, current)
	}
	return lines
}

// stripAnsi removes ANSI escape sequences from a string.
// This is a simplified version - for production use lipgloss's internal functions.
func stripAnsi(s string) string {
	var result []rune
	inEscape := false
	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result = append(result, r)
	}
	return string(result)
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
