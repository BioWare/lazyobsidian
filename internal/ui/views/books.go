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

// BookViewMode represents the viewing mode for books.
type BookViewMode int

const (
	BookViewList BookViewMode = iota
	BookViewChapters
)

// BookNode represents a node in the book tree (book or chapter).
type BookNode struct {
	Type     BookNodeType
	Book     *types.Book
	Chapter  *types.BookChapter
	Level    int
	Expanded bool
	Index    int
	Children []*BookNode
}

// BookNodeType represents the type of node.
type BookNodeType int

const (
	NodeTypeBook BookNodeType = iota
	NodeTypeChapter
)

// BooksView represents the books list view.
type BooksView struct {
	Width  int
	Height int

	// Data
	Books     []types.Book
	nodes     []*BookNode
	flatNodes []*BookNode

	// UI state
	Mode          BookViewMode
	SelectedIndex int
	ScrollOffset  int
	Focused       bool
	ShowCompleted bool
}

// NewBooksView creates a new books view.
func NewBooksView(width, height int) *BooksView {
	return &BooksView{
		Width:         width,
		Height:        height,
		Mode:          BookViewList,
		ShowCompleted: false,
	}
}

// SetSize updates the view dimensions.
func (b *BooksView) SetSize(width, height int) {
	b.Width = width
	b.Height = height
}

// SetBooks sets the books data.
func (b *BooksView) SetBooks(books []types.Book) {
	b.Books = books
	b.buildTree()
	b.flattenVisible()
}

// SetFocused sets the focus state.
func (b *BooksView) SetFocused(focused bool) {
	b.Focused = focused
}

// ToggleCompleted toggles showing completed books.
func (b *BooksView) ToggleCompleted() {
	b.ShowCompleted = !b.ShowCompleted
	b.buildTree()
	b.flattenVisible()
}

// SelectNext moves selection down.
func (b *BooksView) SelectNext() {
	if b.SelectedIndex < len(b.flatNodes)-1 {
		b.SelectedIndex++
		b.ensureVisible()
	}
}

// SelectPrev moves selection up.
func (b *BooksView) SelectPrev() {
	if b.SelectedIndex > 0 {
		b.SelectedIndex--
		b.ensureVisible()
	}
}

// ToggleExpand expands or collapses the selected book.
func (b *BooksView) ToggleExpand() {
	if b.SelectedIndex >= 0 && b.SelectedIndex < len(b.flatNodes) {
		node := b.flatNodes[b.SelectedIndex]
		if len(node.Children) > 0 {
			node.Expanded = !node.Expanded
			b.flattenVisible()
		}
	}
}

// ExpandAll expands all nodes.
func (b *BooksView) ExpandAll() {
	for _, node := range b.nodes {
		node.Expanded = true
	}
	b.flattenVisible()
}

// CollapseAll collapses all nodes.
func (b *BooksView) CollapseAll() {
	for _, node := range b.nodes {
		node.Expanded = false
	}
	b.flattenVisible()
}

// SelectedBook returns the currently selected book.
func (b *BooksView) SelectedBook() *types.Book {
	if b.SelectedIndex >= 0 && b.SelectedIndex < len(b.flatNodes) {
		node := b.flatNodes[b.SelectedIndex]
		if node.Book != nil {
			return node.Book
		}
	}
	return nil
}

// SelectedNode returns the currently selected node.
func (b *BooksView) SelectedNode() *BookNode {
	if b.SelectedIndex >= 0 && b.SelectedIndex < len(b.flatNodes) {
		return b.flatNodes[b.SelectedIndex]
	}
	return nil
}

func (b *BooksView) buildTree() {
	b.nodes = make([]*BookNode, 0)

	for i := range b.Books {
		book := &b.Books[i]

		// Filter completed books
		progress := float64(book.CurrentPage) / float64(book.TotalPages)
		if progress >= 1.0 && !b.ShowCompleted {
			continue
		}

		bookNode := &BookNode{
			Type:     NodeTypeBook,
			Book:     book,
			Level:    0,
			Expanded: false,
		}

		for j := range book.Chapters {
			chapter := &book.Chapters[j]
			chapterNode := &BookNode{
				Type:    NodeTypeChapter,
				Book:    book,
				Chapter: chapter,
				Level:   1,
			}
			bookNode.Children = append(bookNode.Children, chapterNode)
		}

		b.nodes = append(b.nodes, bookNode)
	}
}

func (b *BooksView) flattenVisible() {
	b.flatNodes = make([]*BookNode, 0)
	for _, node := range b.nodes {
		b.flatNodes = append(b.flatNodes, node)
		if node.Expanded {
			for _, child := range node.Children {
				b.flatNodes = append(b.flatNodes, child)
			}
		}
	}

	for i, node := range b.flatNodes {
		node.Index = i
	}

	if b.SelectedIndex >= len(b.flatNodes) {
		b.SelectedIndex = len(b.flatNodes) - 1
	}
	if b.SelectedIndex < 0 && len(b.flatNodes) > 0 {
		b.SelectedIndex = 0
	}
}

func (b *BooksView) ensureVisible() {
	visibleLines := b.Height - 4
	if visibleLines < 1 {
		visibleLines = 1
	}

	if b.SelectedIndex < b.ScrollOffset {
		b.ScrollOffset = b.SelectedIndex
	}
	if b.SelectedIndex >= b.ScrollOffset+visibleLines {
		b.ScrollOffset = b.SelectedIndex - visibleLines + 1
	}
}

// Render renders the books view.
func (b *BooksView) Render() string {
	if len(b.flatNodes) == 0 {
		return b.renderEmpty()
	}

	return b.renderListWithDetails()
}

func (b *BooksView) renderEmpty() string {
	th := theme.Current
	t := i18n.T()
	screen := layout.NewScreen(b.Width, b.Height)

	msg := lipgloss.NewStyle().
		Foreground(th.Color("text_muted")).
		Render(icons.Get("book") + " " + t.Books.NoBooks)

	x := (b.Width - lipgloss.Width(msg)) / 2
	y := b.Height / 2
	screen.DrawBlock(x, y, msg)

	return screen.String()
}

func (b *BooksView) renderListWithDetails() string {
	th := theme.Current
	t := i18n.T()

	// Split: 55% list, 45% details
	listWidth := b.Width * 55 / 100
	detailsWidth := b.Width - listWidth

	// List panel
	listFrame := layout.NewFrame(listWidth, b.Height)
	listFrame.SetTitle(icons.Get("book") + " " + t.Nav.Books)
	listFrame.SetBorder(layout.BorderSingle)
	listFrame.SetFocused(b.Focused)
	listFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	listFrame.SetContentLines(b.renderListLines(listFrame.ContentWidth()))

	// Details panel
	detailsFrame := layout.NewFrame(detailsWidth, b.Height)
	detailsFrame.SetTitle("Book Details")
	detailsFrame.SetBorder(layout.BorderSingle)
	detailsFrame.SetFocused(false)
	detailsFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	detailsFrame.SetContentLines(b.renderBookDetails(detailsFrame.ContentWidth()))

	return combineHorizontal(listFrame.Render(), detailsFrame.Render())
}

func (b *BooksView) renderListLines(width int) []string {
	th := theme.Current
	var lines []string

	visibleLines := b.Height - 4
	if visibleLines < 1 {
		visibleLines = 1
	}

	endIdx := b.ScrollOffset + visibleLines
	if endIdx > len(b.flatNodes) {
		endIdx = len(b.flatNodes)
	}

	for i := b.ScrollOffset; i < endIdx; i++ {
		node := b.flatNodes[i]
		line := b.renderNodeLine(node, width, i == b.SelectedIndex, th)
		lines = append(lines, line)
	}

	return lines
}

func (b *BooksView) renderNodeLine(node *BookNode, width int, selected bool, th *theme.Theme) string {
	var prefix strings.Builder
	t := i18n.T()

	// Indentation
	indent := strings.Repeat("  ", node.Level)
	prefix.WriteString(indent)

	// Expand/collapse indicator
	if node.Type == NodeTypeBook && len(node.Children) > 0 {
		if node.Expanded {
			prefix.WriteString(icons.Get("expanded") + " ")
		} else {
			prefix.WriteString(icons.Get("collapsed") + " ")
		}
	} else if node.Type == NodeTypeBook {
		prefix.WriteString("  ")
	} else {
		prefix.WriteString("  ")
	}

	// Icon and title based on node type
	var icon, title, statusStr string
	var progress float64

	switch node.Type {
	case NodeTypeBook:
		progress = float64(node.Book.CurrentPage) / float64(node.Book.TotalPages)
		if progress >= 1.0 {
			icon = icons.Get("check")
		} else if progress > 0 {
			icon = icons.Get("reading")
		} else {
			icon = icons.Get("book")
		}
		title = node.Book.Title
		// Use i18n format for pages
		statusStr = " " + i18n.Format(t.Books.Pages, map[string]interface{}{
			"current": node.Book.CurrentPage,
			"total":   node.Book.TotalPages,
		})

	case NodeTypeChapter:
		if node.Chapter.Status == "done" || node.Chapter.Status == "completed" {
			icon = icons.Get("check")
			progress = 1.0
		} else if node.Chapter.Status == "in_progress" || node.Chapter.Status == "reading" {
			icon = icons.Get("reading")
			progress = 0.5
		} else {
			icon = icons.Get("chapter")
			progress = 0
		}
		title = node.Chapter.Title
		if node.Chapter.HasNote {
			statusStr = " " + icons.Get("note")
		}
	}

	prefix.WriteString(icon + " ")
	prefixStr := prefix.String()
	prefixWidth := lipgloss.Width(prefixStr)

	statusWidth := lipgloss.Width(statusStr)
	titleWidth := width - prefixWidth - statusWidth - 2
	if titleWidth < 10 {
		titleWidth = 10
	}

	title = layout.TruncateWithEllipsis(title, titleWidth)

	// Style
	var titleStyle, statusStyle lipgloss.Style
	if selected && b.Focused {
		titleStyle = lipgloss.NewStyle().
			Foreground(th.Color("bg_primary")).
			Background(th.Color("accent")).
			Bold(true)
		statusStyle = titleStyle
	} else if selected {
		titleStyle = lipgloss.NewStyle().
			Foreground(th.Color("text_primary")).
			Background(th.Color("bg_secondary"))
		statusStyle = titleStyle
	} else {
		titleStyle = lipgloss.NewStyle().Foreground(th.Color("text_primary"))
		if progress >= 1.0 {
			statusStyle = lipgloss.NewStyle().Foreground(th.Color("success"))
		} else if progress >= 0.5 {
			statusStyle = lipgloss.NewStyle().Foreground(th.Color("warning"))
		} else {
			statusStyle = lipgloss.NewStyle().Foreground(th.Color("text_muted"))
		}
	}

	styledTitle := titleStyle.Render(layout.FitToWidth(title, titleWidth))
	styledStatus := statusStyle.Render(statusStr)

	line := prefixStr + styledTitle + styledStatus
	return layout.FitToWidth(line, width)
}

func (b *BooksView) renderBookDetails(width int) []string {
	th := theme.Current
	t := i18n.T()
	var lines []string

	node := b.SelectedNode()
	if node == nil {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(th.Color("text_muted")).
			Italic(true).
			Render("Select a book to see details"))
		return lines
	}

	// Get the book
	var book *types.Book
	if node.Book != nil {
		book = node.Book
	}

	if book == nil {
		return lines
	}

	labelStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(th.Color("text_primary"))
	mutedStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))

	// Title
	titleLine := labelStyle.Render("Title: ") + valueStyle.Render(book.Title)
	lines = append(lines, layout.TruncateWithEllipsis(titleLine, width))

	// Author
	if book.Author != "" {
		authorLine := labelStyle.Render("Author: ") + valueStyle.Render(book.Author)
		lines = append(lines, layout.TruncateWithEllipsis(authorLine, width))
	}
	lines = append(lines, "")

	// Reading Progress
	progress := float64(book.CurrentPage) / float64(book.TotalPages)
	var statusLabel string
	if progress >= 1.0 {
		statusLabel = t.Books.Completed
	} else {
		statusLabel = t.Books.Reading
	}
	lines = append(lines, labelStyle.Render(statusLabel))
	lines = append(lines, "")

	// Progress bar
	progressBar := b.renderProgressBar(progress, width-4, th)
	lines = append(lines, "  "+progressBar)
	lines = append(lines, "")

	// Pages info
	pagesLabel := i18n.Format(t.Books.Pages, map[string]interface{}{
		"current": book.CurrentPage,
		"total":   book.TotalPages,
	})
	pagesRemaining := book.TotalPages - book.CurrentPage
	lines = append(lines, labelStyle.Render("Pages: ")+valueStyle.Render(pagesLabel))
	if pagesRemaining > 0 {
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("  %d pages remaining", pagesRemaining)))
	}
	lines = append(lines, "")

	// Chapters
	chaptersLabel := i18n.Format(t.Books.Chapters, map[string]interface{}{
		"count": len(book.Chapters),
	})
	lines = append(lines, labelStyle.Render("Chapters: ")+valueStyle.Render(chaptersLabel))

	// Count completed chapters
	completedChapters := 0
	for _, ch := range book.Chapters {
		if ch.Status == "done" || ch.Status == "completed" {
			completedChapters++
		}
	}
	if len(book.Chapters) > 0 {
		lines = append(lines, mutedStyle.Render(fmt.Sprintf("  %d/%d completed", completedChapters, len(book.Chapters))))
	}
	lines = append(lines, "")

	// Pomodoros and Notes
	if book.Pomodoros > 0 || book.Notes > 0 {
		pomodorosLine := labelStyle.Render("Pomodoros: ") + valueStyle.Render(fmt.Sprintf("%d", book.Pomodoros))
		notesLine := labelStyle.Render("Notes: ") + valueStyle.Render(fmt.Sprintf("%d", book.Notes))
		lines = append(lines, pomodorosLine)
		lines = append(lines, notesLine)
		lines = append(lines, "")
	}

	// Target date
	if book.TargetDate != nil {
		targetLine := labelStyle.Render("Target: ") + valueStyle.Render(book.TargetDate.Format("2006-01-02"))
		lines = append(lines, targetLine)

		// Calculate reading pace needed
		if pagesRemaining > 0 {
			daysLeft := int(time.Until(*book.TargetDate).Hours() / 24)
			if daysLeft > 0 {
				pagesPerDay := float64(pagesRemaining) / float64(daysLeft)
				paceStr := fmt.Sprintf("  %.1f pages/day needed", pagesPerDay)
				lines = append(lines, mutedStyle.Render(paceStr))
			}
		}
	}

	return lines
}

func (b *BooksView) renderProgressBar(progress float64, width int, th *theme.Theme) string {
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
