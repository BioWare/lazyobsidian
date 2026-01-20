// Package components provides reusable UI components.
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ProgressBarStyle defines the style of progress bar.
type ProgressBarStyle string

const (
	ProgressStyleBlocks  ProgressBarStyle = "blocks"
	ProgressStyleLine    ProgressBarStyle = "line"
	ProgressStyleDots    ProgressBarStyle = "dots"
	ProgressStylePercent ProgressBarStyle = "percent"
)

// ProgressBar renders a progress bar.
type ProgressBar struct {
	Width      int
	Progress   float64 // 0.0 - 1.0
	Style      ProgressBarStyle
	FillStyle  lipgloss.Style
	EmptyStyle lipgloss.Style
	ShowLabel  bool
}

// NewProgressBar creates a new progress bar with default styling.
func NewProgressBar(width int, progress float64) *ProgressBar {
	return &ProgressBar{
		Width:      width,
		Progress:   progress,
		Style:      ProgressStyleBlocks,
		FillStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("#2D5A7B")),
		EmptyStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#D4C9B5")),
		ShowLabel:  true,
	}
}

// Render renders the progress bar.
func (p *ProgressBar) Render() string {
	switch p.Style {
	case ProgressStyleBlocks:
		return p.renderBlocks()
	case ProgressStyleLine:
		return p.renderLine()
	case ProgressStyleDots:
		return p.renderDots()
	case ProgressStylePercent:
		return p.renderPercent()
	default:
		return p.renderBlocks()
	}
}

func (p *ProgressBar) renderBlocks() string {
	// Reserve space for percentage label if showing
	barWidth := p.Width
	if p.ShowLabel {
		barWidth -= 5 // " 100%"
	}

	if barWidth < 1 {
		barWidth = 1
	}

	filled := int(float64(barWidth) * p.Progress)
	if filled > barWidth {
		filled = barWidth
	}

	empty := barWidth - filled

	bar := p.FillStyle.Render(strings.Repeat("█", filled)) +
		p.EmptyStyle.Render(strings.Repeat("░", empty))

	if p.ShowLabel {
		percent := int(p.Progress * 100)
		bar += p.FillStyle.Render(" ") + lipgloss.NewStyle().Render(padLeft(itoa(percent)+"%", 4))
	}

	return bar
}

func (p *ProgressBar) renderLine() string {
	barWidth := p.Width
	if p.ShowLabel {
		barWidth -= 5
	}

	if barWidth < 1 {
		barWidth = 1
	}

	filled := int(float64(barWidth) * p.Progress)
	empty := barWidth - filled

	bar := p.FillStyle.Render(strings.Repeat("━", filled)) +
		p.EmptyStyle.Render(strings.Repeat("─", empty))

	if p.ShowLabel {
		percent := int(p.Progress * 100)
		bar += " " + padLeft(itoa(percent)+"%", 4)
	}

	return bar
}

func (p *ProgressBar) renderDots() string {
	barWidth := p.Width
	if p.ShowLabel {
		barWidth -= 5
	}

	if barWidth < 1 {
		barWidth = 1
	}

	filled := int(float64(barWidth) * p.Progress)
	empty := barWidth - filled

	bar := p.FillStyle.Render(strings.Repeat("●", filled)) +
		p.EmptyStyle.Render(strings.Repeat("○", empty))

	if p.ShowLabel {
		percent := int(p.Progress * 100)
		bar += " " + padLeft(itoa(percent)+"%", 4)
	}

	return bar
}

func (p *ProgressBar) renderPercent() string {
	percent := int(p.Progress * 100)
	return p.FillStyle.Render(itoa(percent) + "%")
}

// Helper functions

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}
