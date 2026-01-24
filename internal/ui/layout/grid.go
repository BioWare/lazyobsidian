package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Grid provides a grid-based layout system.
// Cells can span multiple rows or columns.
type Grid struct {
	// Rows is the number of rows.
	Rows int
	// Cols is the number of columns.
	Cols int
	// Width is the total width.
	Width int
	// Height is the total height.
	Height int
	// RowHeights specifies fixed heights for rows (0 means auto).
	RowHeights []int
	// ColWidths specifies fixed widths for columns (0 means auto).
	ColWidths []int
	// Gap is the spacing between cells.
	Gap int
	// Cells contains the grid cells.
	Cells []*GridCell
}

// GridCell represents a cell in the grid.
type GridCell struct {
	// Row is the starting row (0-indexed).
	Row int
	// Col is the starting column (0-indexed).
	Col int
	// RowSpan is the number of rows this cell spans.
	RowSpan int
	// ColSpan is the number of columns this cell spans.
	ColSpan int
	// Content is the rendered content string.
	Content string
	// Panel is an optional panel to render (overrides Content).
	Panel *Panel
}

// NewGrid creates a new grid.
func NewGrid(rows, cols int) *Grid {
	return &Grid{
		Rows:       rows,
		Cols:       cols,
		RowHeights: make([]int, rows),
		ColWidths:  make([]int, cols),
		Gap:        0,
	}
}

// SetSize sets the total grid size.
func (g *Grid) SetSize(width, height int) *Grid {
	g.Width = width
	g.Height = height
	return g
}

// SetGap sets the gap between cells.
func (g *Grid) SetGap(gap int) *Grid {
	g.Gap = gap
	return g
}

// SetRowHeight sets a fixed height for a row.
func (g *Grid) SetRowHeight(row, height int) *Grid {
	if row >= 0 && row < g.Rows {
		g.RowHeights[row] = height
	}
	return g
}

// SetColWidth sets a fixed width for a column.
func (g *Grid) SetColWidth(col, width int) *Grid {
	if col >= 0 && col < g.Cols {
		g.ColWidths[col] = width
	}
	return g
}

// AddCell adds a cell to the grid.
func (g *Grid) AddCell(row, col int, content string) *Grid {
	g.Cells = append(g.Cells, &GridCell{
		Row:     row,
		Col:     col,
		RowSpan: 1,
		ColSpan: 1,
		Content: content,
	})
	return g
}

// AddCellSpan adds a cell that spans multiple rows/columns.
func (g *Grid) AddCellSpan(row, col, rowSpan, colSpan int, content string) *Grid {
	g.Cells = append(g.Cells, &GridCell{
		Row:     row,
		Col:     col,
		RowSpan: rowSpan,
		ColSpan: colSpan,
		Content: content,
	})
	return g
}

// AddPanel adds a panel cell to the grid.
func (g *Grid) AddPanel(row, col int, panel *Panel) *Grid {
	g.Cells = append(g.Cells, &GridCell{
		Row:     row,
		Col:     col,
		RowSpan: 1,
		ColSpan: 1,
		Panel:   panel,
	})
	return g
}

// AddPanelSpan adds a panel that spans multiple rows/columns.
func (g *Grid) AddPanelSpan(row, col, rowSpan, colSpan int, panel *Panel) *Grid {
	g.Cells = append(g.Cells, &GridCell{
		Row:     row,
		Col:     col,
		RowSpan: rowSpan,
		ColSpan: colSpan,
		Panel:   panel,
	})
	return g
}

// Render renders the grid to a string.
func (g *Grid) Render() string {
	if g.Width <= 0 || g.Height <= 0 || g.Rows <= 0 || g.Cols <= 0 {
		return ""
	}

	// Calculate actual row heights and column widths
	rowHeights := g.calculateRowHeights()
	colWidths := g.calculateColWidths()

	// Calculate cell positions
	rowPositions := g.calculatePositions(rowHeights)
	colPositions := g.calculatePositions(colWidths)

	// Create output buffer
	buf := newBuffer(g.Width, g.Height)

	// Render each cell
	for _, cell := range g.Cells {
		if cell.Row >= g.Rows || cell.Col >= g.Cols {
			continue
		}

		// Calculate cell dimensions
		x := colPositions[cell.Col]
		y := rowPositions[cell.Row]

		// Calculate cell width (sum of spanned columns + gaps)
		cellWidth := 0
		for c := cell.Col; c < cell.Col+cell.ColSpan && c < g.Cols; c++ {
			cellWidth += colWidths[c]
			if c > cell.Col && g.Gap > 0 {
				cellWidth += g.Gap
			}
		}

		// Calculate cell height (sum of spanned rows + gaps)
		cellHeight := 0
		for r := cell.Row; r < cell.Row+cell.RowSpan && r < g.Rows; r++ {
			cellHeight += rowHeights[r]
			if r > cell.Row && g.Gap > 0 {
				cellHeight += g.Gap
			}
		}

		// Render cell content
		var content string
		if cell.Panel != nil {
			cell.Panel.Width = cellWidth
			cell.Panel.Height = cellHeight
			content = cell.Panel.Render()
		} else {
			content = fitContentToSize(cell.Content, cellWidth, cellHeight)
		}

		// Write to buffer
		buf.writeStyled(x, y, content)
	}

	return buf.String()
}

// calculateRowHeights calculates the height of each row.
func (g *Grid) calculateRowHeights() []int {
	heights := make([]int, g.Rows)
	totalGap := g.Gap * (g.Rows - 1)
	availableHeight := g.Height - totalGap

	// Count fixed and auto rows
	fixedHeight := 0
	autoCount := 0

	for i := 0; i < g.Rows; i++ {
		if g.RowHeights[i] > 0 {
			fixedHeight += g.RowHeights[i]
			heights[i] = g.RowHeights[i]
		} else {
			autoCount++
		}
	}

	// Distribute remaining height to auto rows
	remaining := availableHeight - fixedHeight
	if autoCount > 0 && remaining > 0 {
		autoHeight := remaining / autoCount
		extraHeight := remaining % autoCount

		for i := 0; i < g.Rows; i++ {
			if g.RowHeights[i] == 0 {
				heights[i] = autoHeight
				if extraHeight > 0 {
					heights[i]++
					extraHeight--
				}
			}
		}
	}

	return heights
}

// calculateColWidths calculates the width of each column.
func (g *Grid) calculateColWidths() []int {
	widths := make([]int, g.Cols)
	totalGap := g.Gap * (g.Cols - 1)
	availableWidth := g.Width - totalGap

	// Count fixed and auto columns
	fixedWidth := 0
	autoCount := 0

	for i := 0; i < g.Cols; i++ {
		if g.ColWidths[i] > 0 {
			fixedWidth += g.ColWidths[i]
			widths[i] = g.ColWidths[i]
		} else {
			autoCount++
		}
	}

	// Distribute remaining width to auto columns
	remaining := availableWidth - fixedWidth
	if autoCount > 0 && remaining > 0 {
		autoWidth := remaining / autoCount
		extraWidth := remaining % autoCount

		for i := 0; i < g.Cols; i++ {
			if g.ColWidths[i] == 0 {
				widths[i] = autoWidth
				if extraWidth > 0 {
					widths[i]++
					extraWidth--
				}
			}
		}
	}

	return widths
}

// calculatePositions calculates the starting position of each row/column.
func (g *Grid) calculatePositions(sizes []int) []int {
	positions := make([]int, len(sizes))
	pos := 0
	for i := range sizes {
		positions[i] = pos
		pos += sizes[i]
		if i < len(sizes)-1 {
			pos += g.Gap
		}
	}
	return positions
}

// fitContentToSize fits content to a specific size.
func fitContentToSize(content string, width, height int) string {
	lines := strings.Split(content, "\n")

	var result []string
	for i := 0; i < height; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}
		line = fitToWidth(line, width)
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// writeStyled writes a multi-line string preserving ANSI styling.
func (b *buffer) writeStyled(x, y int, content string) {
	lines := strings.Split(content, "\n")
	for row, line := range lines {
		if y+row >= b.height || y+row < 0 {
			continue
		}

		// We need to write character by character, tracking position
		col := x
		inEscape := false

		for _, r := range line {
			if col >= b.width {
				break
			}

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

			if col >= 0 {
				b.cells[y+row][col] = r
			}
			col++
		}
	}
}

// FlexRow creates a flexible horizontal layout.
type FlexRow struct {
	Items   []FlexItem
	Width   int
	Height  int
	Gap     int
	Padding int
}

// FlexItem represents an item in a flex layout.
type FlexItem struct {
	Content string
	Panel   *Panel
	Width   int    // Fixed width (0 = auto/flex)
	Flex    int    // Flex grow factor (default 1)
	Align   string // "top", "center", "bottom"
}

// NewFlexRow creates a new flex row.
func NewFlexRow(width, height int) *FlexRow {
	return &FlexRow{
		Width:  width,
		Height: height,
	}
}

// SetGap sets the gap between items.
func (f *FlexRow) SetGap(gap int) *FlexRow {
	f.Gap = gap
	return f
}

// Add adds an item with a flex factor.
func (f *FlexRow) Add(content string, flex int) *FlexRow {
	f.Items = append(f.Items, FlexItem{
		Content: content,
		Flex:    flex,
	})
	return f
}

// AddFixed adds a fixed-width item.
func (f *FlexRow) AddFixed(content string, width int) *FlexRow {
	f.Items = append(f.Items, FlexItem{
		Content: content,
		Width:   width,
	})
	return f
}

// AddPanel adds a panel with a flex factor.
func (f *FlexRow) AddPanel(panel *Panel, flex int) *FlexRow {
	f.Items = append(f.Items, FlexItem{
		Panel: panel,
		Flex:  flex,
	})
	return f
}

// AddPanelFixed adds a fixed-width panel.
func (f *FlexRow) AddPanelFixed(panel *Panel, width int) *FlexRow {
	f.Items = append(f.Items, FlexItem{
		Panel: panel,
		Width: width,
	})
	return f
}

// Render renders the flex row.
func (f *FlexRow) Render() string {
	if len(f.Items) == 0 {
		return ""
	}

	// Calculate widths
	widths := f.calculateWidths()

	// Render each item
	var rendered []string
	for i, item := range f.Items {
		w := widths[i]

		var content string
		if item.Panel != nil {
			item.Panel.Width = w
			item.Panel.Height = f.Height
			content = item.Panel.Render()
		} else {
			content = fitContentToSize(item.Content, w, f.Height)
		}

		rendered = append(rendered, content)
	}

	// Join with gap
	if f.Gap > 0 {
		gap := strings.Repeat(" ", f.Gap)
		var result []string
		for i, r := range rendered {
			result = append(result, r)
			if i < len(rendered)-1 {
				result = append(result, fitContentToSize(gap, f.Gap, f.Height))
			}
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, result...)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// calculateWidths calculates the width for each item.
func (f *FlexRow) calculateWidths() []int {
	widths := make([]int, len(f.Items))
	totalGap := f.Gap * (len(f.Items) - 1)
	available := f.Width - totalGap

	// First pass: allocate fixed widths
	fixedTotal := 0
	flexTotal := 0

	for i, item := range f.Items {
		if item.Width > 0 {
			widths[i] = item.Width
			fixedTotal += item.Width
		} else {
			flex := item.Flex
			if flex <= 0 {
				flex = 1
			}
			flexTotal += flex
		}
	}

	// Second pass: distribute remaining space
	remaining := available - fixedTotal
	if remaining < 0 {
		remaining = 0
	}

	for i, item := range f.Items {
		if item.Width == 0 {
			flex := item.Flex
			if flex <= 0 {
				flex = 1
			}
			if flexTotal > 0 {
				widths[i] = remaining * flex / flexTotal
			}
		}
	}

	return widths
}

// FlexColumn creates a flexible vertical layout.
type FlexColumn struct {
	Items   []FlexItem
	Width   int
	Height  int
	Gap     int
	Padding int
}

// NewFlexColumn creates a new flex column.
func NewFlexColumn(width, height int) *FlexColumn {
	return &FlexColumn{
		Width:  width,
		Height: height,
	}
}

// SetGap sets the gap between items.
func (f *FlexColumn) SetGap(gap int) *FlexColumn {
	f.Gap = gap
	return f
}

// Add adds an item with a flex factor.
func (f *FlexColumn) Add(content string, flex int) *FlexColumn {
	f.Items = append(f.Items, FlexItem{
		Content: content,
		Flex:    flex,
	})
	return f
}

// AddFixed adds a fixed-height item.
func (f *FlexColumn) AddFixed(content string, height int) *FlexColumn {
	f.Items = append(f.Items, FlexItem{
		Content: content,
		Width:   height, // Using Width field for height in column layout
	})
	return f
}

// AddPanel adds a panel with a flex factor.
func (f *FlexColumn) AddPanel(panel *Panel, flex int) *FlexColumn {
	f.Items = append(f.Items, FlexItem{
		Panel: panel,
		Flex:  flex,
	})
	return f
}

// AddPanelFixed adds a fixed-height panel.
func (f *FlexColumn) AddPanelFixed(panel *Panel, height int) *FlexColumn {
	f.Items = append(f.Items, FlexItem{
		Panel: panel,
		Width: height, // Using Width field for height in column layout
	})
	return f
}

// Render renders the flex column.
func (f *FlexColumn) Render() string {
	if len(f.Items) == 0 {
		return ""
	}

	// Calculate heights
	heights := f.calculateHeights()

	// Render each item
	var rendered []string
	for i, item := range f.Items {
		h := heights[i]

		var content string
		if item.Panel != nil {
			item.Panel.Width = f.Width
			item.Panel.Height = h
			content = item.Panel.Render()
		} else {
			content = fitContentToSize(item.Content, f.Width, h)
		}

		rendered = append(rendered, content)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rendered...)
}

// calculateHeights calculates the height for each item.
func (f *FlexColumn) calculateHeights() []int {
	heights := make([]int, len(f.Items))
	totalGap := f.Gap * (len(f.Items) - 1)
	available := f.Height - totalGap

	// First pass: allocate fixed heights
	fixedTotal := 0
	flexTotal := 0

	for i, item := range f.Items {
		if item.Width > 0 { // Width is used for height in column layout
			heights[i] = item.Width
			fixedTotal += item.Width
		} else {
			flex := item.Flex
			if flex <= 0 {
				flex = 1
			}
			flexTotal += flex
		}
	}

	// Second pass: distribute remaining space
	remaining := available - fixedTotal
	if remaining < 0 {
		remaining = 0
	}

	for i, item := range f.Items {
		if item.Width == 0 {
			flex := item.Flex
			if flex <= 0 {
				flex = 1
			}
			if flexTotal > 0 {
				heights[i] = remaining * flex / flexTotal
			}
		}
	}

	return heights
}
