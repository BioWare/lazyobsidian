package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Box renders a bordered box with title.
type Box struct {
	Title       string
	Content     string
	Width       int
	Height      int
	BorderStyle lipgloss.Border
	BorderColor lipgloss.Color
	TitleStyle  lipgloss.Style
}

// NewBox creates a new box component.
func NewBox(title string, width, height int) *Box {
	return &Box{
		Title:       title,
		Width:       width,
		Height:      height,
		BorderStyle: lipgloss.RoundedBorder(),
		BorderColor: lipgloss.Color("#D4C9B5"),
		TitleStyle:  lipgloss.NewStyle().Bold(true),
	}
}

// SetContent sets the box content.
func (b *Box) SetContent(content string) *Box {
	b.Content = content
	return b
}

// Render renders the box.
func (b *Box) Render() string {
	style := lipgloss.NewStyle().
		Border(b.BorderStyle).
		BorderForeground(b.BorderColor).
		Width(b.Width - 2).
		Height(b.Height - 2)

	// Build title line
	titleLine := ""
	if b.Title != "" {
		titleLine = b.TitleStyle.Render(b.Title)
	}

	// Wrap content
	content := b.Content
	if content == "" {
		content = strings.Repeat("\n", b.Height-3)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		titleLine,
		style.Render(content),
	)
}

// Panel renders a panel with title bar.
type Panel struct {
	Title      string
	Content    string
	Width      int
	Height     int
	Focused    bool
	TitleStyle lipgloss.Style
	BodyStyle  lipgloss.Style
}

// NewPanel creates a new panel.
func NewPanel(title string, width, height int) *Panel {
	return &Panel{
		Title:  title,
		Width:  width,
		Height: height,
		TitleStyle: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1),
		BodyStyle: lipgloss.NewStyle().
			Padding(0, 1),
	}
}

// SetContent sets panel content.
func (p *Panel) SetContent(content string) *Panel {
	p.Content = content
	return p
}

// SetFocused sets focus state.
func (p *Panel) SetFocused(focused bool) *Panel {
	p.Focused = focused
	return p
}

// Render renders the panel.
func (p *Panel) Render() string {
	// Title bar
	borderChar := "─"
	titleText := " " + p.Title + " "
	remainingWidth := p.Width - len(titleText) - 2

	leftBorder := "┌"
	rightBorder := "┐"
	if p.Focused {
		leftBorder = "╔"
		rightBorder = "╗"
		borderChar = "═"
	}

	titleBar := leftBorder + borderChar + p.TitleStyle.Render(p.Title) + " " +
		strings.Repeat(borderChar, max(0, remainingWidth-len(p.Title)-3)) + rightBorder

	// Content lines
	lines := strings.Split(p.Content, "\n")
	contentHeight := p.Height - 3 // title + bottom border

	var contentLines []string
	for i := 0; i < contentHeight; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}

		// Pad or truncate line
		if len(line) > p.Width-4 {
			line = line[:p.Width-5] + "…"
		}
		line = line + strings.Repeat(" ", max(0, p.Width-4-len(line)))

		sideChar := "│"
		if p.Focused {
			sideChar = "║"
		}
		contentLines = append(contentLines, sideChar+" "+line+" "+sideChar)
	}

	// Bottom border
	bottomLeft := "└"
	bottomRight := "┘"
	if p.Focused {
		bottomLeft = "╚"
		bottomRight = "╝"
	}
	bottomBar := bottomLeft + strings.Repeat(borderChar, p.Width-2) + bottomRight

	return titleBar + "\n" + strings.Join(contentLines, "\n") + "\n" + bottomBar
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
