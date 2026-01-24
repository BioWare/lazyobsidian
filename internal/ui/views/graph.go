// Package views implements the UI views for LazyObsidian.
package views

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/i18n"
	"github.com/BioWare/lazyobsidian/internal/ui/icons"
	"github.com/BioWare/lazyobsidian/internal/ui/layout"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
)

// GraphNode represents a node in the graph.
type GraphNode struct {
	ID       string
	Label    string
	Type     string // "note", "daily", "goal", "course", "book"
	X, Y     float64
	Links    []string // IDs of connected nodes
	Selected bool
}

// GraphFilter represents filtering options for the graph.
type GraphFilter struct {
	ShowNotes   bool
	ShowDaily   bool
	ShowGoals   bool
	ShowCourses bool
	ShowBooks   bool
	SearchQuery string
}

// GraphView represents the graph visualization view.
type GraphView struct {
	Width  int
	Height int

	// Data
	Nodes       []GraphNode
	NodeIndex   map[string]*GraphNode
	VisibleNodes []*GraphNode

	// View state
	ViewX, ViewY float64 // View center
	Zoom         float64
	SelectedIdx  int

	// Filter
	Filter GraphFilter

	// UI state
	Focused        bool
	ShowFilterMenu bool
}

// NewGraphView creates a new graph view.
func NewGraphView(width, height int) *GraphView {
	return &GraphView{
		Width:     width,
		Height:    height,
		NodeIndex: make(map[string]*GraphNode),
		Zoom:      1.0,
		Filter: GraphFilter{
			ShowNotes:   true,
			ShowDaily:   true,
			ShowGoals:   true,
			ShowCourses: true,
			ShowBooks:   true,
		},
	}
}

// SetSize updates the view dimensions.
func (g *GraphView) SetSize(width, height int) {
	g.Width = width
	g.Height = height
}

// SetNodes sets the graph nodes.
func (g *GraphView) SetNodes(nodes []GraphNode) {
	g.Nodes = nodes
	g.NodeIndex = make(map[string]*GraphNode)
	for i := range g.Nodes {
		g.NodeIndex[g.Nodes[i].ID] = &g.Nodes[i]
	}
	g.applyLayout()
	g.filterNodes()
}

// SetFocused sets the focus state.
func (g *GraphView) SetFocused(focused bool) {
	g.Focused = focused
}

// ToggleFilterMenu toggles the filter menu.
func (g *GraphView) ToggleFilterMenu() {
	g.ShowFilterMenu = !g.ShowFilterMenu
}

// ToggleFilter toggles a specific filter.
func (g *GraphView) ToggleFilter(filter string) {
	switch filter {
	case "notes":
		g.Filter.ShowNotes = !g.Filter.ShowNotes
	case "daily":
		g.Filter.ShowDaily = !g.Filter.ShowDaily
	case "goals":
		g.Filter.ShowGoals = !g.Filter.ShowGoals
	case "courses":
		g.Filter.ShowCourses = !g.Filter.ShowCourses
	case "books":
		g.Filter.ShowBooks = !g.Filter.ShowBooks
	}
	g.filterNodes()
}

// SetSearchQuery sets the search filter.
func (g *GraphView) SetSearchQuery(query string) {
	g.Filter.SearchQuery = query
	g.filterNodes()
}

// Pan moves the view.
func (g *GraphView) Pan(dx, dy float64) {
	g.ViewX += dx / g.Zoom
	g.ViewY += dy / g.Zoom
}

// ZoomIn increases zoom level.
func (g *GraphView) ZoomIn() {
	g.Zoom *= 1.2
	if g.Zoom > 5.0 {
		g.Zoom = 5.0
	}
}

// ZoomOut decreases zoom level.
func (g *GraphView) ZoomOut() {
	g.Zoom /= 1.2
	if g.Zoom < 0.2 {
		g.Zoom = 0.2
	}
}

// ResetView resets the view to default.
func (g *GraphView) ResetView() {
	g.ViewX = 0
	g.ViewY = 0
	g.Zoom = 1.0
}

// SelectNext selects the next node.
func (g *GraphView) SelectNext() {
	if len(g.VisibleNodes) == 0 {
		return
	}
	g.SelectedIdx = (g.SelectedIdx + 1) % len(g.VisibleNodes)
	g.centerOnSelected()
}

// SelectPrev selects the previous node.
func (g *GraphView) SelectPrev() {
	if len(g.VisibleNodes) == 0 {
		return
	}
	g.SelectedIdx = (g.SelectedIdx - 1 + len(g.VisibleNodes)) % len(g.VisibleNodes)
	g.centerOnSelected()
}

// SelectedNode returns the currently selected node.
func (g *GraphView) SelectedNode() *GraphNode {
	if g.SelectedIdx >= 0 && g.SelectedIdx < len(g.VisibleNodes) {
		return g.VisibleNodes[g.SelectedIdx]
	}
	return nil
}

func (g *GraphView) centerOnSelected() {
	node := g.SelectedNode()
	if node != nil {
		g.ViewX = node.X
		g.ViewY = node.Y
	}
}

func (g *GraphView) applyLayout() {
	// Simple force-directed layout simulation
	if len(g.Nodes) == 0 {
		return
	}

	// Initialize positions in a circle
	for i := range g.Nodes {
		angle := 2 * math.Pi * float64(i) / float64(len(g.Nodes))
		radius := float64(20)
		g.Nodes[i].X = radius * math.Cos(angle)
		g.Nodes[i].Y = radius * math.Sin(angle)
	}

	// Run a few iterations of force-directed layout
	for iter := 0; iter < 50; iter++ {
		// Repulsion between all nodes
		for i := range g.Nodes {
			for j := range g.Nodes {
				if i == j {
					continue
				}
				dx := g.Nodes[i].X - g.Nodes[j].X
				dy := g.Nodes[i].Y - g.Nodes[j].Y
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist < 0.1 {
					dist = 0.1
				}
				force := 100.0 / (dist * dist)
				g.Nodes[i].X += dx / dist * force * 0.1
				g.Nodes[i].Y += dy / dist * force * 0.1
			}
		}

		// Attraction along edges
		for i := range g.Nodes {
			for _, linkID := range g.Nodes[i].Links {
				linked := g.NodeIndex[linkID]
				if linked == nil {
					continue
				}
				dx := linked.X - g.Nodes[i].X
				dy := linked.Y - g.Nodes[i].Y
				dist := math.Sqrt(dx*dx + dy*dy)
				if dist > 5 {
					force := (dist - 5) * 0.05
					g.Nodes[i].X += dx / dist * force
					g.Nodes[i].Y += dy / dist * force
				}
			}
		}
	}
}

func (g *GraphView) filterNodes() {
	g.VisibleNodes = make([]*GraphNode, 0)

	for i := range g.Nodes {
		node := &g.Nodes[i]

		// Type filter
		show := false
		switch node.Type {
		case "note":
			show = g.Filter.ShowNotes
		case "daily":
			show = g.Filter.ShowDaily
		case "goal":
			show = g.Filter.ShowGoals
		case "course":
			show = g.Filter.ShowCourses
		case "book":
			show = g.Filter.ShowBooks
		default:
			show = g.Filter.ShowNotes
		}

		if !show {
			continue
		}

		// Search filter
		if g.Filter.SearchQuery != "" {
			if !strings.Contains(strings.ToLower(node.Label), strings.ToLower(g.Filter.SearchQuery)) {
				continue
			}
		}

		g.VisibleNodes = append(g.VisibleNodes, node)
	}

	// Reset selection if needed
	if g.SelectedIdx >= len(g.VisibleNodes) {
		g.SelectedIdx = len(g.VisibleNodes) - 1
	}
	if g.SelectedIdx < 0 && len(g.VisibleNodes) > 0 {
		g.SelectedIdx = 0
	}
}

// Render renders the graph view.
func (g *GraphView) Render() string {
	if g.ShowFilterMenu {
		return g.renderWithFilterMenu()
	}
	return g.renderGraph()
}

func (g *GraphView) renderGraph() string {
	th := theme.Current
	t := i18n.T()

	// Split: graph area + info panel
	graphWidth := g.Width * 70 / 100
	infoWidth := g.Width - graphWidth

	// Graph panel
	graphFrame := layout.NewFrame(graphWidth, g.Height)
	graphFrame.SetTitle(icons.Get("graph") + " " + t.Nav.Graph)
	graphFrame.SetBorder(layout.BorderSingle)
	graphFrame.SetFocused(g.Focused)
	graphFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	graphFrame.SetContentLines(g.renderGraphContent(graphFrame.ContentWidth(), graphFrame.ContentHeight(), th))

	// Info panel
	infoFrame := layout.NewFrame(infoWidth, g.Height)
	infoFrame.SetTitle("Node Info")
	infoFrame.SetBorder(layout.BorderSingle)
	infoFrame.SetFocused(false)
	infoFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	infoFrame.SetContentLines(g.renderNodeInfo(infoFrame.ContentWidth(), th))

	return combineHorizontal(graphFrame.Render(), infoFrame.Render())
}

func (g *GraphView) renderGraphContent(width, height int, th *theme.Theme) []string {
	// Create a character buffer for the graph
	buffer := make([][]rune, height)
	for i := range buffer {
		buffer[i] = make([]rune, width)
		for j := range buffer[i] {
			buffer[i][j] = ' '
		}
	}

	// Calculate visible area
	halfW := float64(width) / 2 / g.Zoom
	halfH := float64(height) / 2 / g.Zoom

	// Draw edges first
	edgeStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
	for _, node := range g.VisibleNodes {
		for _, linkID := range node.Links {
			linked := g.NodeIndex[linkID]
			if linked == nil {
				continue
			}

			// Check if linked node is visible
			linkVisible := false
			for _, vn := range g.VisibleNodes {
				if vn.ID == linkID {
					linkVisible = true
					break
				}
			}
			if !linkVisible {
				continue
			}

			// Draw line between nodes
			g.drawLine(buffer, width, height, halfW, halfH,
				node.X, node.Y, linked.X, linked.Y, '·')
		}
	}

	// Draw nodes
	for _, node := range g.VisibleNodes {
		// Convert world coords to screen coords
		screenX := int((node.X-g.ViewX)*g.Zoom + float64(width)/2)
		screenY := int((node.Y-g.ViewY)*g.Zoom + float64(height)/2)

		// Check if in bounds
		if screenX < 0 || screenX >= width || screenY < 0 || screenY >= height {
			continue
		}

		// Node character based on type
		var nodeChar rune
		switch node.Type {
		case "daily":
			nodeChar = '◆'
		case "goal":
			nodeChar = '★'
		case "course":
			nodeChar = '◈'
		case "book":
			nodeChar = '◉'
		default:
			nodeChar = '●'
		}

		buffer[screenY][screenX] = nodeChar

		// Draw label if space allows
		label := layout.TruncateWithEllipsis(node.Label, 10)
		labelX := screenX + 2
		if labelX+len(label) < width {
			for i, r := range label {
				if labelX+i < width {
					buffer[screenY][labelX+i] = r
				}
			}
		}
	}

	// Highlight selected node
	selectedNode := g.SelectedNode()
	if selectedNode != nil {
		screenX := int((selectedNode.X-g.ViewX)*g.Zoom + float64(width)/2)
		screenY := int((selectedNode.Y-g.ViewY)*g.Zoom + float64(height)/2)

		if screenX > 0 && screenX < width-1 && screenY > 0 && screenY < height-1 {
			buffer[screenY][screenX-1] = '['
			buffer[screenY][screenX+1] = ']'
		}
	}

	// Convert buffer to styled lines
	var lines []string
	_ = edgeStyle
	nodeColors := map[string]lipgloss.Color{
		"note":   th.Color("text_primary"),
		"daily":  th.Color("info"),
		"goal":   th.Color("success"),
		"course": th.Color("warning"),
		"book":   th.Color("accent"),
	}

	for y := 0; y < height; y++ {
		line := string(buffer[y])

		// Apply colors to node characters (simplified - in real implementation would track positions)
		styledLine := lipgloss.NewStyle().Foreground(th.Color("text_primary")).Render(line)
		_ = nodeColors

		lines = append(lines, styledLine)
	}

	// Add status line
	statusStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
	statusLine := statusStyle.Render(fmt.Sprintf("Zoom: %.0f%%  Nodes: %d/%d  [+/-] Zoom  [hjkl] Pan  [n/p] Select  [f] Filter",
		g.Zoom*100, len(g.VisibleNodes), len(g.Nodes)))
	if len(lines) > 0 {
		lines[len(lines)-1] = layout.FitToWidth(statusLine, width)
	}

	return lines
}

func (g *GraphView) drawLine(buffer [][]rune, width, height int, halfW, halfH float64, x1, y1, x2, y2 float64, char rune) {
	// Bresenham's line algorithm
	sx1 := int((x1-g.ViewX)*g.Zoom + float64(width)/2)
	sy1 := int((y1-g.ViewY)*g.Zoom + float64(height)/2)
	sx2 := int((x2-g.ViewX)*g.Zoom + float64(width)/2)
	sy2 := int((y2-g.ViewY)*g.Zoom + float64(height)/2)

	dx := abs(sx2 - sx1)
	dy := -abs(sy2 - sy1)
	sx := 1
	if sx1 > sx2 {
		sx = -1
	}
	sy := 1
	if sy1 > sy2 {
		sy = -1
	}
	err := dx + dy

	for {
		if sx1 >= 0 && sx1 < width && sy1 >= 0 && sy1 < height {
			// Don't overwrite node characters
			if buffer[sy1][sx1] == ' ' {
				buffer[sy1][sx1] = char
			}
		}

		if sx1 == sx2 && sy1 == sy2 {
			break
		}

		e2 := 2 * err
		if e2 >= dy {
			if sx1 == sx2 {
				break
			}
			err += dy
			sx1 += sx
		}
		if e2 <= dx {
			if sy1 == sy2 {
				break
			}
			err += dx
			sy1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *GraphView) renderNodeInfo(width int, th *theme.Theme) []string {
	var lines []string

	node := g.SelectedNode()
	if node == nil {
		mutedStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted")).Italic(true)
		lines = append(lines, mutedStyle.Render("Select a node to see details"))
		return lines
	}

	labelStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(th.Color("text_primary"))
	mutedStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))

	// Title
	lines = append(lines, labelStyle.Render("Title:"))
	lines = append(lines, "  "+valueStyle.Render(layout.TruncateWithEllipsis(node.Label, width-2)))
	lines = append(lines, "")

	// Type
	typeIcon := icons.Get(node.Type)
	typeLine := labelStyle.Render("Type: ") + valueStyle.Render(typeIcon+" "+node.Type)
	lines = append(lines, typeLine)
	lines = append(lines, "")

	// Connections
	lines = append(lines, labelStyle.Render("Connections:"))
	if len(node.Links) == 0 {
		lines = append(lines, "  "+mutedStyle.Render("No connections"))
	} else {
		for i, linkID := range node.Links {
			if i >= 5 { // Show max 5 links
				lines = append(lines, "  "+mutedStyle.Render(fmt.Sprintf("...and %d more", len(node.Links)-5)))
				break
			}
			linked := g.NodeIndex[linkID]
			if linked != nil {
				linkLine := "  → " + layout.TruncateWithEllipsis(linked.Label, width-6)
				lines = append(lines, mutedStyle.Render(linkLine))
			}
		}
	}
	lines = append(lines, "")

	// Position (for debugging)
	posLine := fmt.Sprintf("Position: (%.1f, %.1f)", node.X, node.Y)
	lines = append(lines, mutedStyle.Render(posLine))

	return lines
}

func (g *GraphView) renderWithFilterMenu() string {
	th := theme.Current
	t := i18n.T()
	screen := layout.NewScreen(g.Width, g.Height)

	// Render graph behind (dimmed)
	graphContent := g.renderGraph()
	screen.DrawBlock(0, 0, graphContent)

	// Filter menu overlay
	menuWidth := 30
	menuHeight := 12
	menuX := (g.Width - menuWidth) / 2
	menuY := (g.Height - menuHeight) / 2

	// Menu frame
	menuFrame := layout.NewFrame(menuWidth, menuHeight)
	menuFrame.SetTitle(icons.Get("filter") + " Filters")
	menuFrame.SetBorder(layout.BorderDouble)
	menuFrame.SetFocused(true)
	menuFrame.SetColors(
		th.Color("border_active"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)

	var menuLines []string
	_ = t

	filters := []struct {
		key     string
		label   string
		enabled bool
	}{
		{"notes", "Notes", g.Filter.ShowNotes},
		{"daily", "Daily Notes", g.Filter.ShowDaily},
		{"goals", "Goals", g.Filter.ShowGoals},
		{"courses", "Courses", g.Filter.ShowCourses},
		{"books", "Books", g.Filter.ShowBooks},
	}

	for _, f := range filters {
		var icon string
		if f.enabled {
			icon = icons.Get("check")
		} else {
			icon = icons.Get("cross")
		}

		style := lipgloss.NewStyle().Foreground(th.Color("text_primary"))
		if f.enabled {
			style = style.Foreground(th.Color("success"))
		}

		line := style.Render(fmt.Sprintf("[%s] %s %s", f.key[0:1], icon, f.label))
		menuLines = append(menuLines, line)
	}

	menuLines = append(menuLines, "")
	helpStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
	menuLines = append(menuLines, helpStyle.Render("[1-5] Toggle  [Esc] Close"))

	menuFrame.SetContentLines(menuLines)
	screen.DrawBlock(menuX, menuY, menuFrame.Render())

	return screen.String()
}
