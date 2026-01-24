// Package views implements the UI views for LazyObsidian.
package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/i18n"
	"github.com/BioWare/lazyobsidian/internal/ui/icons"
	"github.com/BioWare/lazyobsidian/internal/ui/layout"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// CourseViewMode represents the viewing mode for courses.
type CourseViewMode int

const (
	CourseViewList CourseViewMode = iota
	CourseViewDetails
)

// CourseNode represents a node in the course tree (course, section, or lesson).
type CourseNode struct {
	Type     CourseNodeType
	Course   *types.Course
	Section  *types.CourseSection
	Lesson   *types.CourseLesson
	Level    int
	Expanded bool
	Index    int
	Children []*CourseNode
}

// CourseNodeType represents the type of node in the tree.
type CourseNodeType int

const (
	NodeTypeCourse CourseNodeType = iota
	NodeTypeSection
	NodeTypeLesson
)

// CoursesView represents the courses list/tree view.
type CoursesView struct {
	Width  int
	Height int

	// Data
	Courses   []types.Course
	nodes     []*CourseNode
	flatNodes []*CourseNode

	// UI state
	Mode          CourseViewMode
	SelectedIndex int
	ScrollOffset  int
	Focused       bool
	ShowCompleted bool
}

// NewCoursesView creates a new courses view.
func NewCoursesView(width, height int) *CoursesView {
	return &CoursesView{
		Width:         width,
		Height:        height,
		Mode:          CourseViewList,
		ShowCompleted: false,
	}
}

// SetSize updates the view dimensions.
func (c *CoursesView) SetSize(width, height int) {
	c.Width = width
	c.Height = height
}

// SetCourses sets the courses data.
func (c *CoursesView) SetCourses(courses []types.Course) {
	c.Courses = courses
	c.buildTree()
	c.flattenVisible()
}

// SetFocused sets the focus state.
func (c *CoursesView) SetFocused(focused bool) {
	c.Focused = focused
}

// ToggleCompleted toggles showing completed courses.
func (c *CoursesView) ToggleCompleted() {
	c.ShowCompleted = !c.ShowCompleted
	c.buildTree()
	c.flattenVisible()
}

// SelectNext moves selection down.
func (c *CoursesView) SelectNext() {
	if c.SelectedIndex < len(c.flatNodes)-1 {
		c.SelectedIndex++
		c.ensureVisible()
	}
}

// SelectPrev moves selection up.
func (c *CoursesView) SelectPrev() {
	if c.SelectedIndex > 0 {
		c.SelectedIndex--
		c.ensureVisible()
	}
}

// ToggleExpand expands or collapses the selected node.
func (c *CoursesView) ToggleExpand() {
	if c.SelectedIndex >= 0 && c.SelectedIndex < len(c.flatNodes) {
		node := c.flatNodes[c.SelectedIndex]
		if len(node.Children) > 0 {
			node.Expanded = !node.Expanded
			c.flattenVisible()
		}
	}
}

// ExpandAll expands all nodes.
func (c *CoursesView) ExpandAll() {
	c.setExpandAll(c.nodes, true)
	c.flattenVisible()
}

// CollapseAll collapses all nodes.
func (c *CoursesView) CollapseAll() {
	c.setExpandAll(c.nodes, false)
	c.flattenVisible()
}

func (c *CoursesView) setExpandAll(nodes []*CourseNode, expanded bool) {
	for _, node := range nodes {
		node.Expanded = expanded
		c.setExpandAll(node.Children, expanded)
	}
}

// SelectedCourse returns the currently selected course.
func (c *CoursesView) SelectedCourse() *types.Course {
	if c.SelectedIndex >= 0 && c.SelectedIndex < len(c.flatNodes) {
		node := c.flatNodes[c.SelectedIndex]
		if node.Course != nil {
			return node.Course
		}
	}
	return nil
}

// SelectedNode returns the currently selected node.
func (c *CoursesView) SelectedNode() *CourseNode {
	if c.SelectedIndex >= 0 && c.SelectedIndex < len(c.flatNodes) {
		return c.flatNodes[c.SelectedIndex]
	}
	return nil
}

func (c *CoursesView) buildTree() {
	c.nodes = make([]*CourseNode, 0)

	for i := range c.Courses {
		course := &c.Courses[i]

		// Filter based on completion
		progress := float64(course.Completed) / float64(course.TotalLessons)
		if progress >= 1.0 && !c.ShowCompleted {
			continue
		}

		courseNode := &CourseNode{
			Type:     NodeTypeCourse,
			Course:   course,
			Level:    0,
			Expanded: false,
		}

		for j := range course.Sections {
			section := &course.Sections[j]
			sectionNode := &CourseNode{
				Type:     NodeTypeSection,
				Course:   course,
				Section:  section,
				Level:    1,
				Expanded: false,
			}

			for k := range section.Lessons {
				lesson := &section.Lessons[k]
				lessonNode := &CourseNode{
					Type:    NodeTypeLesson,
					Course:  course,
					Section: section,
					Lesson:  lesson,
					Level:   2,
				}
				sectionNode.Children = append(sectionNode.Children, lessonNode)
			}

			courseNode.Children = append(courseNode.Children, sectionNode)
		}

		c.nodes = append(c.nodes, courseNode)
	}
}

func (c *CoursesView) flattenVisible() {
	c.flatNodes = make([]*CourseNode, 0)
	c.flattenNodes(c.nodes)

	for i, node := range c.flatNodes {
		node.Index = i
	}

	if c.SelectedIndex >= len(c.flatNodes) {
		c.SelectedIndex = len(c.flatNodes) - 1
	}
	if c.SelectedIndex < 0 && len(c.flatNodes) > 0 {
		c.SelectedIndex = 0
	}
}

func (c *CoursesView) flattenNodes(nodes []*CourseNode) {
	for _, node := range nodes {
		c.flatNodes = append(c.flatNodes, node)
		if node.Expanded {
			c.flattenNodes(node.Children)
		}
	}
}

func (c *CoursesView) ensureVisible() {
	visibleLines := c.Height - 4
	if visibleLines < 1 {
		visibleLines = 1
	}

	if c.SelectedIndex < c.ScrollOffset {
		c.ScrollOffset = c.SelectedIndex
	}
	if c.SelectedIndex >= c.ScrollOffset+visibleLines {
		c.ScrollOffset = c.SelectedIndex - visibleLines + 1
	}
}

// Render renders the courses view.
func (c *CoursesView) Render() string {
	if len(c.flatNodes) == 0 {
		return c.renderEmpty()
	}

	switch c.Mode {
	case CourseViewDetails:
		return c.renderDetailsMode()
	default:
		return c.renderListMode()
	}
}

func (c *CoursesView) renderEmpty() string {
	th := theme.Current
	t := i18n.T()
	screen := layout.NewScreen(c.Width, c.Height)

	msg := lipgloss.NewStyle().
		Foreground(th.Color("text_muted")).
		Render(icons.Get("courses") + " " + t.Courses.NoCourses)

	x := (c.Width - lipgloss.Width(msg)) / 2
	y := c.Height / 2
	screen.DrawBlock(x, y, msg)

	return screen.String()
}

func (c *CoursesView) renderListMode() string {
	th := theme.Current
	t := i18n.T()

	// Split: 60% list, 40% details
	listWidth := c.Width * 60 / 100
	detailsWidth := c.Width - listWidth

	// List panel
	listFrame := layout.NewFrame(listWidth, c.Height)
	listFrame.SetTitle(icons.Get("courses") + " " + t.Nav.Courses)
	listFrame.SetBorder(layout.BorderSingle)
	listFrame.SetFocused(c.Focused)
	listFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	listFrame.SetContentLines(c.renderListLines(listFrame.ContentWidth()))

	// Details panel
	detailsFrame := layout.NewFrame(detailsWidth, c.Height)
	detailsFrame.SetTitle("Course Details")
	detailsFrame.SetBorder(layout.BorderSingle)
	detailsFrame.SetFocused(false)
	detailsFrame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	detailsFrame.SetContentLines(c.renderCourseDetails(detailsFrame.ContentWidth()))

	// Combine horizontally
	return combineHorizontal(listFrame.Render(), detailsFrame.Render())
}

func (c *CoursesView) renderDetailsMode() string {
	// Full-width details for selected course
	th := theme.Current

	frame := layout.NewFrame(c.Width, c.Height)
	frame.SetTitle("Course Details")
	frame.SetBorder(layout.BorderSingle)
	frame.SetFocused(c.Focused)
	frame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)
	frame.SetContentLines(c.renderFullDetails(frame.ContentWidth()))

	return frame.Render()
}

func (c *CoursesView) renderListLines(width int) []string {
	th := theme.Current
	var lines []string

	visibleLines := c.Height - 4
	if visibleLines < 1 {
		visibleLines = 1
	}

	endIdx := c.ScrollOffset + visibleLines
	if endIdx > len(c.flatNodes) {
		endIdx = len(c.flatNodes)
	}

	for i := c.ScrollOffset; i < endIdx; i++ {
		node := c.flatNodes[i]
		line := c.renderNodeLine(node, width, i == c.SelectedIndex, th)
		lines = append(lines, line)
	}

	return lines
}

func (c *CoursesView) renderNodeLine(node *CourseNode, width int, selected bool, th *theme.Theme) string {
	var prefix strings.Builder
	t := i18n.T()

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

	// Icon based on node type and status
	var icon, title, statusStr string
	var progress float64

	switch node.Type {
	case NodeTypeCourse:
		icon = icons.Get("courses")
		title = node.Course.Title
		if node.Course.TotalLessons > 0 {
			progress = float64(node.Course.Completed) / float64(node.Course.TotalLessons)
		}
		statusStr = fmt.Sprintf(" %d/%d", node.Course.Completed, node.Course.TotalLessons)

	case NodeTypeSection:
		icon = icons.Get("folder")
		title = node.Section.Title
		progress = node.Section.Progress
		lessonCount := len(node.Section.Lessons)
		completed := 0
		for _, l := range node.Section.Lessons {
			if l.Status == "done" || l.Status == "completed" {
				completed++
			}
		}
		statusStr = fmt.Sprintf(" %d/%d", completed, lessonCount)

	case NodeTypeLesson:
		if node.Lesson.Status == "done" || node.Lesson.Status == "completed" {
			icon = icons.Get("check")
			progress = 1.0
		} else if node.Lesson.Status == "in_progress" {
			icon = icons.Get("progress")
			progress = 0.5
		} else {
			icon = icons.Get("open")
			progress = 0
		}
		title = node.Lesson.Title
		if node.Lesson.Duration > 0 {
			statusStr = fmt.Sprintf(" %dm", node.Lesson.Duration)
		}
		if node.Lesson.HasNote {
			statusStr += " " + icons.Get("note")
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
	if selected && c.Focused {
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

	_ = t // available for future use

	styledTitle := titleStyle.Render(layout.FitToWidth(title, titleWidth))
	styledStatus := statusStyle.Render(statusStr)

	line := prefixStr + styledTitle + styledStatus
	return layout.FitToWidth(line, width)
}

func (c *CoursesView) renderCourseDetails(width int) []string {
	th := theme.Current
	t := i18n.T()
	var lines []string

	node := c.SelectedNode()
	if node == nil {
		lines = append(lines, lipgloss.NewStyle().
			Foreground(th.Color("text_muted")).
			Italic(true).
			Render("Select a course to see details"))
		return lines
	}

	// Get the course (from any node type)
	var course *types.Course
	if node.Course != nil {
		course = node.Course
	}

	if course == nil {
		return lines
	}

	labelStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(th.Color("text_primary"))
	mutedStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))

	// Title
	titleLine := labelStyle.Render("Title: ") + valueStyle.Render(course.Title)
	lines = append(lines, layout.TruncateWithEllipsis(titleLine, width))
	lines = append(lines, "")

	// Source
	if course.Source != "" {
		sourceLine := labelStyle.Render("Source: ") + valueStyle.Render(course.Source)
		lines = append(lines, layout.TruncateWithEllipsis(sourceLine, width))
	}

	// URL
	if course.URL != "" {
		urlLine := labelStyle.Render("URL: ") + mutedStyle.Render(layout.TruncateWithEllipsis(course.URL, width-5))
		lines = append(lines, urlLine)
	}
	lines = append(lines, "")

	// Progress
	progress := float64(course.Completed) / float64(course.TotalLessons)
	lines = append(lines, labelStyle.Render("Progress:"))
	progressBar := c.renderProgressBar(progress, width-4, th)
	lines = append(lines, "  "+progressBar)
	lines = append(lines, "")

	// Lessons
	lessonsLabel := i18n.Format(t.Courses.Lessons, map[string]interface{}{
		"completed": course.Completed,
		"total":     course.TotalLessons,
	})
	lines = append(lines, labelStyle.Render(lessonsLabel))

	// Sections
	sectionsLabel := i18n.Format(t.Courses.Sections, map[string]interface{}{
		"count": len(course.Sections),
	})
	lines = append(lines, mutedStyle.Render(sectionsLabel))
	lines = append(lines, "")

	// Pomodoros and Notes
	pomodorosLine := labelStyle.Render("Pomodoros: ") + valueStyle.Render(fmt.Sprintf("%d", course.Pomodoros))
	notesLine := labelStyle.Render("Notes: ") + valueStyle.Render(fmt.Sprintf("%d", course.Notes))
	lines = append(lines, pomodorosLine)
	lines = append(lines, notesLine)

	// Target date
	if course.TargetDate != nil {
		lines = append(lines, "")
		targetLine := labelStyle.Render("Target: ") + valueStyle.Render(course.TargetDate.Format("2006-01-02"))
		lines = append(lines, targetLine)
	}

	return lines
}

func (c *CoursesView) renderFullDetails(width int) []string {
	// Similar to renderCourseDetails but with more space
	return c.renderCourseDetails(width)
}

func (c *CoursesView) renderProgressBar(progress float64, width int, th *theme.Theme) string {
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

// Helper function to combine two rendered frames horizontally
func combineHorizontal(left, right string) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	var result strings.Builder
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	for i := 0; i < maxLines; i++ {
		var leftLine, rightLine string
		if i < len(leftLines) {
			leftLine = leftLines[i]
		}
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}
		result.WriteString(leftLine)
		result.WriteString(rightLine)
		if i < maxLines-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}
