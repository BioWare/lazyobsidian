// Package layout provides a constraint-based layout system for TUI components.
package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Screen represents a terminal screen buffer that properly handles ANSI styling.
// Unlike a character buffer, it stores styled lines and composes them correctly.
type Screen struct {
	width  int
	height int
	lines  []string
}

// NewScreen creates a new screen with the given dimensions.
func NewScreen(width, height int) *Screen {
	lines := make([]string, height)
	emptyLine := strings.Repeat(" ", width)
	for i := range lines {
		lines[i] = emptyLine
	}
	return &Screen{
		width:  width,
		height: height,
		lines:  lines,
	}
}

// Width returns the screen width.
func (s *Screen) Width() int {
	return s.width
}

// Height returns the screen height.
func (s *Screen) Height() int {
	return s.height
}

// Clear resets the screen to empty spaces.
func (s *Screen) Clear() {
	emptyLine := strings.Repeat(" ", s.width)
	for i := range s.lines {
		s.lines[i] = emptyLine
	}
}

// SetLine sets a line at the given row, ensuring it fits exactly within width.
func (s *Screen) SetLine(row int, content string) {
	if row < 0 || row >= s.height {
		return
	}
	s.lines[row] = FitToWidth(content, s.width)
}

// DrawBlock draws a multi-line string at the given position.
// The content should already be properly styled.
func (s *Screen) DrawBlock(x, y int, content string) {
	contentLines := strings.Split(content, "\n")

	for i, line := range contentLines {
		row := y + i
		if row < 0 || row >= s.height {
			continue
		}

		if x == 0 && lipgloss.Width(line) == s.width {
			// Fast path: line already fits perfectly
			s.lines[row] = line
		} else {
			// Need to compose with existing content
			s.lines[row] = composeLine(s.lines[row], x, line, s.width)
		}
	}
}

// DrawPanel draws a Panel at the specified position.
func (s *Screen) DrawPanel(x, y int, panel *Panel) {
	rendered := panel.Render()
	s.DrawBlock(x, y, rendered)
}

// String returns the screen as a single string.
func (s *Screen) String() string {
	return strings.Join(s.lines, "\n")
}

// composeLine overlays content onto an existing line at position x.
func composeLine(existing string, x int, content string, totalWidth int) string {
	contentWidth := lipgloss.Width(content)

	// Simple case: content starts at 0 and spans full width
	if x == 0 && contentWidth >= totalWidth {
		return FitToWidth(content, totalWidth)
	}

	// Need to preserve existing content before and after
	var result strings.Builder

	// Part before content
	if x > 0 {
		existingWidth := lipgloss.Width(existing)
		if existingWidth >= x {
			result.WriteString(TruncateToWidth(existing, x))
		} else {
			result.WriteString(existing)
			result.WriteString(strings.Repeat(" ", x-existingWidth))
		}
	}

	// The content itself
	result.WriteString(content)

	// Part after content (if any)
	afterX := x + contentWidth
	if afterX < totalWidth {
		// Extract remaining part from existing line
		remaining := totalWidth - afterX
		existingAfter := extractFromPosition(existing, afterX, remaining)
		result.WriteString(existingAfter)
	}

	return FitToWidth(result.String(), totalWidth)
}

// extractFromPosition extracts content from a styled string starting at visual position.
func extractFromPosition(s string, start, maxWidth int) string {
	if start < 0 || maxWidth <= 0 {
		return ""
	}

	totalWidth := lipgloss.Width(s)
	if start >= totalWidth {
		return strings.Repeat(" ", maxWidth)
	}

	// Skip to start position
	var result strings.Builder
	pos := 0
	inEscape := false
	collecting := false
	collectedWidth := 0

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			if collecting {
				result.WriteRune(r)
			}
			continue
		}

		if inEscape {
			if collecting {
				result.WriteRune(r)
			}
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		// Regular character
		charWidth := RuneWidth(r)

		if pos >= start && !collecting {
			collecting = true
		}

		if collecting {
			if collectedWidth+charWidth > maxWidth {
				break
			}
			result.WriteRune(r)
			collectedWidth += charWidth
		}

		pos += charWidth
	}

	// Pad if needed
	if collectedWidth < maxWidth {
		result.WriteString(strings.Repeat(" ", maxWidth-collectedWidth))
	}

	return result.String()
}

// FitToWidth ensures a string is exactly the given visual width.
// Truncates if too long, pads with spaces if too short.
func FitToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}

	currentWidth := lipgloss.Width(s)

	if currentWidth == width {
		return s
	}

	if currentWidth > width {
		return TruncateToWidth(s, width)
	}

	// Pad with spaces
	return s + strings.Repeat(" ", width-currentWidth)
}

// TruncateToWidth truncates a string to fit within maxWidth.
// Handles ANSI escape sequences properly.
func TruncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	currentWidth := lipgloss.Width(s)
	if currentWidth <= maxWidth {
		return s
	}

	var result strings.Builder
	width := 0
	inEscape := false
	escapeSeq := ""

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
			escapeSeq = string(r)
			continue
		}

		if inEscape {
			escapeSeq += string(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
				result.WriteString(escapeSeq)
				escapeSeq = ""
			}
			continue
		}

		// Regular character
		charWidth := RuneWidth(r)
		if width+charWidth > maxWidth {
			break
		}

		result.WriteRune(r)
		width += charWidth
	}

	// Add reset if we were in styled text
	resultStr := result.String()
	if strings.Contains(resultStr, "\x1b[") && !strings.HasSuffix(resultStr, "\x1b[0m") {
		resultStr += "\x1b[0m"
	}

	return resultStr
}

// TruncateWithEllipsis truncates a string and adds ellipsis if needed.
func TruncateWithEllipsis(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	currentWidth := lipgloss.Width(s)
	if currentWidth <= maxWidth {
		return s
	}

	if maxWidth <= 1 {
		return "…"
	}

	return TruncateToWidth(s, maxWidth-1) + "…"
}

// RuneWidth returns the display width of a rune.
func RuneWidth(r rune) int {
	// Wide characters (CJK, etc.)
	if r >= 0x1100 &&
		(r <= 0x115f || r == 0x2329 || r == 0x232a ||
			(r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
			(r >= 0xac00 && r <= 0xd7a3) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xfe10 && r <= 0xfe19) ||
			(r >= 0xfe30 && r <= 0xfe6f) ||
			(r >= 0xff00 && r <= 0xff60) ||
			(r >= 0xffe0 && r <= 0xffe6) ||
			(r >= 0x20000 && r <= 0x2fffd) ||
			(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}

	// Some emojis take 2 cells
	if r >= 0x1F300 && r <= 0x1F9FF {
		return 2
	}

	return 1
}

// PadCenter centers text within the given width.
func PadCenter(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		return TruncateToWidth(s, width)
	}

	padding := width - currentWidth
	leftPad := padding / 2
	rightPad := padding - leftPad

	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// PadRight pads text on the right to fill width.
func PadRight(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		return TruncateToWidth(s, width)
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

// PadLeft pads text on the left to fill width.
func PadLeft(s string, width int) string {
	currentWidth := lipgloss.Width(s)
	if currentWidth >= width {
		return TruncateToWidth(s, width)
	}
	return strings.Repeat(" ", width-currentWidth) + s
}

// JoinLines joins lines vertically, ensuring each has the specified width.
func JoinLines(lines []string, width int) string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = FitToWidth(line, width)
	}
	return strings.Join(result, "\n")
}

// SplitIntoLines splits a string by newlines.
func SplitIntoLines(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, "\n")
}

// WrapText wraps text to fit within the specified width.
func WrapText(s string, width int) []string {
	if width <= 0 {
		return nil
	}

	var lines []string
	var currentLine strings.Builder
	currentWidth := 0

	words := strings.Fields(s)
	for i, word := range words {
		wordWidth := lipgloss.Width(word)

		if currentWidth == 0 {
			// First word on line
			if wordWidth > width {
				// Word too long, truncate
				lines = append(lines, TruncateWithEllipsis(word, width))
			} else {
				currentLine.WriteString(word)
				currentWidth = wordWidth
			}
		} else if currentWidth+1+wordWidth <= width {
			// Word fits with space
			currentLine.WriteString(" ")
			currentLine.WriteString(word)
			currentWidth += 1 + wordWidth
		} else {
			// Word doesn't fit, start new line
			lines = append(lines, currentLine.String())
			currentLine.Reset()

			if wordWidth > width {
				lines = append(lines, TruncateWithEllipsis(word, width))
				currentWidth = 0
			} else {
				currentLine.WriteString(word)
				currentWidth = wordWidth
			}
		}

		_ = i // Silence unused variable warning
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}
