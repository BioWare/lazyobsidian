// Package views implements the UI views for LazyObsidian.
package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/i18n"
	"github.com/BioWare/lazyobsidian/internal/ui/icons"
	"github.com/BioWare/lazyobsidian/internal/ui/layout"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// GoalNode represents a goal in the tree with UI state.
type GoalNode struct {
	Goal      types.Goal
	Expanded  bool
	Level     int
	Parent    *GoalNode
	Children  []*GoalNode
	Index     int // flat index for selection
}

// GoalsView represents the goals tree view.
type GoalsView struct {
	Width  int
	Height int

	// Data
	Goals     []types.Goal
	nodes     []*GoalNode
	flatNodes []*GoalNode // flattened visible nodes

	// UI state
	SelectedIndex int
	ScrollOffset  int
	Focused       bool
	ShowDetails   bool
}

// NewGoalsView creates a new goals tree view.
func NewGoalsView(width, height int) *GoalsView {
	return &GoalsView{
		Width:       width,
		Height:      height,
		ShowDetails: true,
	}
}

// SetSize updates the view dimensions.
func (g *GoalsView) SetSize(width, height int) {
	g.Width = width
	g.Height = height
}

// SetGoals sets the goals data.
func (g *GoalsView) SetGoals(goals []types.Goal) {
	g.Goals = goals
	g.buildTree()
	g.flattenVisible()
}

// SetFocused sets the focus state.
func (g *GoalsView) SetFocused(focused bool) {
	g.Focused = focused
}

// ToggleDetails toggles the details panel visibility.
func (g *GoalsView) ToggleDetails() {
	g.ShowDetails = !g.ShowDetails
}

// SelectNext moves selection down.
func (g *GoalsView) SelectNext() {
	if g.SelectedIndex < len(g.flatNodes)-1 {
		g.SelectedIndex++
		g.ensureVisible()
	}
}

// SelectPrev moves selection up.
func (g *GoalsView) SelectPrev() {
	if g.SelectedIndex > 0 {
		g.SelectedIndex--
		g.ensureVisible()
	}
}

// ToggleExpand expands or collapses the selected goal.
func (g *GoalsView) ToggleExpand() {
	if g.SelectedIndex >= 0 && g.SelectedIndex < len(g.flatNodes) {
		node := g.flatNodes[g.SelectedIndex]
		if len(node.Children) > 0 {
			node.Expanded = !node.Expanded
			g.flattenVisible()
		}
	}
}

// ExpandAll expands all nodes.
func (g *GoalsView) ExpandAll() {
	g.setExpandAll(g.nodes, true)
	g.flattenVisible()
}

// CollapseAll collapses all nodes.
func (g *GoalsView) CollapseAll() {
	g.setExpandAll(g.nodes, false)
	g.flattenVisible()
}

func (g *GoalsView) setExpandAll(nodes []*GoalNode, expanded bool) {
	for _, node := range nodes {
		node.Expanded = expanded
		g.setExpandAll(node.Children, expanded)
	}
}

// SelectedGoal returns the currently selected goal.
func (g *GoalsView) SelectedGoal() *types.Goal {
	if g.SelectedIndex >= 0 && g.SelectedIndex < len(g.flatNodes) {
		return &g.flatNodes[g.SelectedIndex].Goal
	}
	return nil
}

func (g *GoalsView) buildTree() {
	g.nodes = make([]*GoalNode, 0, len(g.Goals))
	for i := range g.Goals {
		node := g.buildNode(&g.Goals[i], nil, 0)
		g.nodes = append(g.nodes, node)
	}
}

func (g *GoalsView) buildNode(goal *types.Goal, parent *GoalNode, level int) *GoalNode {
	node := &GoalNode{
		Goal:     *goal,
		Level:    level,
		Parent:   parent,
		Expanded: level < 2, // auto-expand first two levels
	}
	for i := range goal.Children {
		child := g.buildNode(&goal.Children[i], node, level+1)
		node.Children = append(node.Children, child)
	}
	return node
}

func (g *GoalsView) flattenVisible() {
	g.flatNodes = make([]*GoalNode, 0)
	g.flattenNodes(g.nodes)

	// Update indices
	for i, node := range g.flatNodes {
		node.Index = i
	}

	// Adjust selection if needed
	if g.SelectedIndex >= len(g.flatNodes) {
		g.SelectedIndex = len(g.flatNodes) - 1
	}
	if g.SelectedIndex < 0 && len(g.flatNodes) > 0 {
		g.SelectedIndex = 0
	}
}

func (g *GoalsView) flattenNodes(nodes []*GoalNode) {
	for _, node := range nodes {
		g.flatNodes = append(g.flatNodes, node)
		if node.Expanded {
			g.flattenNodes(node.Children)
		}
	}
}

func (g *GoalsView) ensureVisible() {
	visibleLines := g.Height - 4 // account for title and borders
	if visibleLines < 1 {
		visibleLines = 1
	}

	if g.SelectedIndex < g.ScrollOffset {
		g.ScrollOffset = g.SelectedIndex
	}
	if g.SelectedIndex >= g.ScrollOffset+visibleLines {
		g.ScrollOffset = g.SelectedIndex - visibleLines + 1
	}
}

// Render renders the goals view.
func (g *GoalsView) Render() string {
	if len(g.flatNodes) == 0 {
		return g.renderEmpty()
	}

	if g.ShowDetails {
		return g.renderWithDetails()
	}
	return g.renderTreeOnly()
}

func (g *GoalsView) renderEmpty() string {
	th := theme.Current
	t := i18n.T()
	screen := layout.NewScreen(g.Width, g.Height)

	msg := lipgloss.NewStyle().
		Foreground(th.Color("text_muted")).
		Render(icons.Get("goals") + " " + t.Goals.NoGoals)

	x := (g.Width - lipgloss.Width(msg)) / 2
	y := g.Height / 2
	screen.DrawBlock(x, y, msg)

	return screen.String()
}

func (g *GoalsView) renderTreeOnly() string {
	th := theme.Current
	t := i18n.T()

	frame := layout.NewFrame(g.Width, g.Height)
	frame.SetTitle(icons.Get("goals") + " " + t.Nav.Goals)
	frame.SetBorder(layout.BorderSingle)
	frame.SetFocused(g.Focused)
	frame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)

	lines := g.renderTreeLines(frame.ContentWidth())
	frame.SetContentLines(lines)

	return frame.Render()
}

func (g *GoalsView) renderWithDetails() string {
	th := theme.Current
	t := i18n.T()

	// Split: 60% tree, 40% details
	treeWidth := g.Width * 60 / 100
	detailsWidth := g.Width - treeWidth

	// Tree panel
	treeFrame := layout.NewFrame(treeWidth, g.Height)
	treeFrame.SetTitle(icons.Get("goals") + " " + t.Nav.Goals)
	treeFrame.SetBorder(layout.BorderSingle)
	treeFrame.SetFocused(g.Focused)
	treeFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	treeFrame.SetContentLines(g.renderTreeLines(treeFrame.ContentWidth()))

	// Details panel
	detailsFrame := layout.NewFrame(detailsWidth, g.Height)
	detailsFrame.SetTitle("Details")
	detailsFrame.SetBorder(layout.BorderSingle)
	detailsFrame.SetFocused(false)
	detailsFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	detailsFrame.SetContentLines(g.renderDetailsLines(detailsFrame.ContentWidth()))

	// Combine horizontally
	treeRendered := treeFrame.Render()
	detailsRendered := detailsFrame.Render()

	treeLines := strings.Split(treeRendered, "\n")
	detailsLines := strings.Split(detailsRendered, "\n")

	var result strings.Builder
	maxLines := len(treeLines)
	if len(detailsLines) > maxLines {
		maxLines = len(detailsLines)
	}

	for i := 0; i < maxLines; i++ {
		var treeLine, detailsLine string
		if i < len(treeLines) {
			treeLine = treeLines[i]
		}
		if i < len(detailsLines) {
			detailsLine = detailsLines[i]
		}
		result.WriteString(treeLine)
		result.WriteString(detailsLine)
		if i < maxLines-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (g *GoalsView) renderTreeLines(width int) []string {
	th := theme.Current
	var lines []string

	visibleLines := g.Height - 4
	if visibleLines < 1 {
		visibleLines = 1
	}

	endIdx := g.ScrollOffset + visibleLines
	if endIdx > len(g.flatNodes) {
		endIdx = len(g.flatNodes)
	}

	for i := g.ScrollOffset; i < endIdx; i++ {
		node := g.flatNodes[i]
		line := g.renderGoalLine(node, width, i == g.SelectedIndex, th)
		lines = append(lines, line)
	}

	return lines
}

func (g *GoalsView) renderGoalLine(node *GoalNode, width int, selected bool, th *theme.Theme) string {
	var prefix strings.Builder

	// Indentation
	indent := strings.Repeat("  ", node.Level)
	prefix.WriteString(indent)

	// Expand/collapse indicator
	if len(node.Children) > 0 {
		if node.Expanded {
			prefix.WriteString(icons.Get("expanded") + " ")
		} else {
			prefix.WriteString(icons.Get("collapsed") + " ")
		}
	} else {
		prefix.WriteString("  ")
	}

	// Goal icon based on progress
	var icon string
	if node.Goal.Progress >= 1.0 {
		icon = icons.Get("check")
	} else if node.Goal.Progress > 0 {
		icon = icons.Get("progress")
	} else {
		icon = icons.Get("goal")
	}
	prefix.WriteString(icon + " ")

	prefixStr := prefix.String()
	prefixWidth := lipgloss.Width(prefixStr)

	// Progress indicator
	progressStr := fmt.Sprintf(" %d%%", int(node.Goal.Progress*100))
	progressWidth := lipgloss.Width(progressStr)

	// Title
	titleWidth := width - prefixWidth - progressWidth - 2
	if titleWidth < 10 {
		titleWidth = 10
	}
	title := layout.TruncateWithEllipsis(node.Goal.Title, titleWidth)

	// Style
	var titleStyle, progressStyle lipgloss.Style
	if selected && g.Focused {
		titleStyle = lipgloss.NewStyle().
			Foreground(th.Color("bg_primary")).
			Background(th.Color("accent")).
			Bold(true)
		progressStyle = titleStyle
	} else if selected {
		titleStyle = lipgloss.NewStyle().
			Foreground(th.Color("text_primary")).
			Background(th.Color("bg_secondary"))
		progressStyle = titleStyle
	} else {
		titleStyle = lipgloss.NewStyle().Foreground(th.Color("text_primary"))
		if node.Goal.Progress >= 1.0 {
			progressStyle = lipgloss.NewStyle().Foreground(th.Color("success"))
		} else if node.Goal.Progress >= 0.5 {
			progressStyle = lipgloss.NewStyle().Foreground(th.Color("warning"))
		} else {
			progressStyle = lipgloss.NewStyle().Foreground(th.Color("text_muted"))
		}
	}

	styledTitle := titleStyle.Render(layout.FitToWidth(title, titleWidth))
	styledProgress := progressStyle.Render(progressStr)

	line := prefixStr + styledTitle + styledProgress
	return layout.FitToWidth(line, width)
}

func (g *GoalsView) renderDetailsLines(width int) []string {
	th := theme.Current
	t := i18n.T()
	var lines []string

	goal := g.SelectedGoal()
	if goal == nil {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(th.Color("text_muted")).
			Italic(true).
			Render("Select a goal to see details"))
		return lines
	}

	labelStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(th.Color("text_primary"))
	mutedStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))

	// Title
	titleLine := labelStyle.Render("Title: ") + valueStyle.Render(goal.Title)
	lines = append(lines, layout.TruncateWithEllipsis(titleLine, width))
	lines = append(lines, "")

	// Description
	if goal.Description != "" {
		lines = append(lines, labelStyle.Render("Description:"))
		descLines := wrapText(goal.Description, width-2)
		for _, dl := range descLines {
			lines = append(lines, "  "+mutedStyle.Render(dl))
		}
		lines = append(lines, "")
	}

	// Progress bar
	lines = append(lines, labelStyle.Render(t.Goals.Progress))
	progressBar := g.renderProgressBar(goal.Progress, width-4, th)
	lines = append(lines, "  "+progressBar)
	lines = append(lines, "")

	// Due date
	if goal.DueDate != nil {
		dueLabel := labelStyle.Render(t.Goals.Due + " ")
		dueValue := valueStyle.Render(goal.DueDate.Format("2006-01-02"))

		daysLeft := int(time.Until(*goal.DueDate).Hours() / 24)
		var daysStyle lipgloss.Style
		if daysLeft < 0 {
			daysStyle = lipgloss.NewStyle().Foreground(th.Color("error"))
		} else if daysLeft <= 7 {
			daysStyle = lipgloss.NewStyle().Foreground(th.Color("warning"))
		} else {
			daysStyle = lipgloss.NewStyle().Foreground(th.Color("text_muted"))
		}
		daysStr := fmt.Sprintf(" (%d days left)", daysLeft)
		if daysLeft < 0 {
			daysStr = fmt.Sprintf(" (%d days overdue)", -daysLeft)
		}

		lines = append(lines, dueLabel+dueValue+daysStyle.Render(daysStr))
		lines = append(lines, "")
	}

	// Pomodoros
	lines = append(lines, labelStyle.Render("Pomodoros:"))
	ownStr := fmt.Sprintf("  Own: %d", goal.OwnPomodoros)
	totalStr := fmt.Sprintf("  Total: %d", goal.Pomodoros)
	lines = append(lines, valueStyle.Render(ownStr))
	lines = append(lines, valueStyle.Render(totalStr))
	lines = append(lines, "")

	// Pace calculation
	if goal.DueDate != nil && goal.Progress < 1.0 {
		daysLeft := int(time.Until(*goal.DueDate).Hours() / 24)
		if daysLeft > 0 {
			remaining := 1.0 - goal.Progress
			pacePerDay := remaining / float64(daysLeft) * 100
			paceStr := fmt.Sprintf("%.1f%%/day needed", pacePerDay)
			lines = append(lines, labelStyle.Render(t.Goals.Pace+" ")+mutedStyle.Render(paceStr))
		}
	}

	return lines
}

func (g *GoalsView) renderProgressBar(progress float64, width int, th *theme.Theme) string {
	if width < 10 {
		width = 10
	}

	filledWidth := int(float64(width) * progress)
	emptyWidth := width - filledWidth

	var filledColor lipgloss.Color
	if progress >= 1.0 {
		filledColor = th.Color("success")
	} else if progress >= 0.5 {
		filledColor = th.Color("warning")
	} else {
		filledColor = th.Color("info")
	}

	filledStyle := lipgloss.NewStyle().Foreground(filledColor)
	emptyStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))

	filled := filledStyle.Render(strings.Repeat("█", filledWidth))
	empty := emptyStyle.Render(strings.Repeat("░", emptyWidth))

	percent := int(progress * 100)
	percentStr := fmt.Sprintf(" %d%%", percent)

	return filled + empty + percentStr
}

// Helper function to wrap text
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return lines
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if lipgloss.Width(currentLine+" "+word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
