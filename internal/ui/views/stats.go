// Package views implements the UI views for LazyObsidian.
package views

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/i18n"
	"github.com/BioWare/lazyobsidian/internal/ui/icons"
	"github.com/BioWare/lazyobsidian/internal/ui/layout"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
	"github.com/BioWare/lazyobsidian/pkg/types"
)

// StatsViewMode represents different statistics views.
type StatsViewMode int

const (
	StatsViewOverview StatsViewMode = iota
	StatsViewHeatmap
	StatsViewCategories
)

// StatsView represents the statistics view.
type StatsView struct {
	Width  int
	Height int

	// Data
	Stats        types.Stats
	WeeklyData   [7]int       // Last 7 days
	MonthlyData  [31]int      // Last 31 days
	YearlyData   map[string]int // date string -> count

	// UI state
	Mode    StatsViewMode
	Focused bool
}

// NewStatsView creates a new statistics view.
func NewStatsView(width, height int) *StatsView {
	return &StatsView{
		Width:      width,
		Height:     height,
		Mode:       StatsViewOverview,
		YearlyData: make(map[string]int),
	}
}

// SetSize updates the view dimensions.
func (s *StatsView) SetSize(width, height int) {
	s.Width = width
	s.Height = height
}

// SetStats sets the statistics data.
func (s *StatsView) SetStats(stats types.Stats) {
	s.Stats = stats
	if stats.ByDay != nil {
		s.YearlyData = stats.ByDay
	}
}

// SetWeeklyData sets the last 7 days pomodoro data.
func (s *StatsView) SetWeeklyData(data [7]int) {
	s.WeeklyData = data
}

// SetMonthlyData sets the last 31 days pomodoro data.
func (s *StatsView) SetMonthlyData(data [31]int) {
	s.MonthlyData = data
}

// SetFocused sets the focus state.
func (s *StatsView) SetFocused(focused bool) {
	s.Focused = focused
}

// NextMode switches to the next view mode.
func (s *StatsView) NextMode() {
	s.Mode = (s.Mode + 1) % 3
}

// PrevMode switches to the previous view mode.
func (s *StatsView) PrevMode() {
	s.Mode = (s.Mode - 1 + 3) % 3
}

// Render renders the statistics view.
func (s *StatsView) Render() string {
	switch s.Mode {
	case StatsViewHeatmap:
		return s.renderHeatmap()
	case StatsViewCategories:
		return s.renderCategories()
	default:
		return s.renderOverview()
	}
}

func (s *StatsView) renderOverview() string {
	th := theme.Current
	t := i18n.T()

	// Create main frame
	frame := layout.NewFrame(s.Width, s.Height)
	frame.SetTitle(icons.Get("stats") + " " + t.Nav.Stats)
	frame.SetBorder(layout.BorderSingle)
	frame.SetFocused(s.Focused)
	frame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)

	var lines []string
	contentWidth := frame.ContentWidth()

	// Summary cards row
	lines = append(lines, s.renderSummaryCards(contentWidth, th, t)...)
	lines = append(lines, "")

	// Weekly chart
	lines = append(lines, s.renderWeeklyChart(contentWidth, th)...)
	lines = append(lines, "")

	// Top categories
	lines = append(lines, s.renderTopCategories(contentWidth, th, t)...)

	// Navigation hint
	lines = append(lines, "")
	navHint := lipgloss.NewStyle().
		Foreground(th.Color("text_muted")).
		Render("[Tab] Switch view  [h] Heatmap  [c] Categories")
	lines = append(lines, navHint)

	frame.SetContentLines(lines)
	return frame.Render()
}

func (s *StatsView) renderSummaryCards(width int, th *theme.Theme, t *i18n.Translations) []string {
	var lines []string

	// Calculate card width (4 cards)
	cardWidth := (width - 6) / 4 // 6 = 3 gaps of 2 chars
	if cardWidth < 12 {
		cardWidth = 12
	}

	// Card data
	cards := []struct {
		label string
		value string
		icon  string
	}{
		{
			label: t.Stats.TotalFocus,
			value: formatDuration(s.Stats.TotalFocusTime),
			icon:  icons.Get("time"),
		},
		{
			label: t.Stats.TotalPomodoros,
			value: fmt.Sprintf("%d", s.Stats.TotalPomodoros),
			icon:  icons.Get("pomodoro"),
		},
		{
			label: t.Stats.TasksCompleted,
			value: fmt.Sprintf("%d", s.Stats.TasksCompleted),
			icon:  icons.Get("check"),
		},
		{
			label: t.Stats.CurrentStreak,
			value: fmt.Sprintf("%d days", s.Stats.CurrentStreak),
			icon:  icons.Get("fire"),
		},
	}

	labelStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
	valueStyle := lipgloss.NewStyle().Foreground(th.Color("text_primary")).Bold(true)
	iconStyle := lipgloss.NewStyle().Foreground(th.Color("accent"))

	// Build card line 1 (icons + values)
	var line1Parts []string
	for _, card := range cards {
		cardContent := iconStyle.Render(card.icon) + " " + valueStyle.Render(card.value)
		cardContent = layout.PadCenter(cardContent, cardWidth)
		line1Parts = append(line1Parts, cardContent)
	}
	lines = append(lines, strings.Join(line1Parts, "  "))

	// Build card line 2 (labels)
	var line2Parts []string
	for _, card := range cards {
		label := layout.TruncateWithEllipsis(card.label, cardWidth)
		label = layout.PadCenter(labelStyle.Render(label), cardWidth)
		line2Parts = append(line2Parts, label)
	}
	lines = append(lines, strings.Join(line2Parts, "  "))

	return lines
}

func (s *StatsView) renderWeeklyChart(width int, th *theme.Theme) []string {
	var lines []string

	titleStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary")).Bold(true)
	lines = append(lines, titleStyle.Render("Last 7 Days"))

	// Find max value
	maxVal := 1
	for _, v := range s.WeeklyData {
		if v > maxVal {
			maxVal = v
		}
	}

	// Chart height
	chartHeight := 5
	chartWidth := width - 10

	// Day labels
	dayLabels := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	barWidth := chartWidth / 7
	if barWidth < 3 {
		barWidth = 3
	}

	// Render bars (top to bottom)
	for row := chartHeight; row > 0; row-- {
		threshold := float64(row) / float64(chartHeight)
		var barLine strings.Builder
		for i, val := range s.WeeklyData {
			normalized := float64(val) / float64(maxVal)
			var bar string
			if normalized >= threshold {
				bar = lipgloss.NewStyle().Foreground(th.Color("success")).Render(strings.Repeat("█", barWidth-1))
			} else {
				bar = strings.Repeat(" ", barWidth-1)
			}
			barLine.WriteString(bar)
			if i < 6 {
				barLine.WriteString(" ")
			}
		}
		lines = append(lines, barLine.String())
	}

	// Day labels row
	var labelLine strings.Builder
	labelStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
	for i, label := range dayLabels {
		labelLine.WriteString(labelStyle.Render(layout.PadCenter(label, barWidth-1)))
		if i < 6 {
			labelLine.WriteString(" ")
		}
	}
	lines = append(lines, labelLine.String())

	// Values row
	var valueLine strings.Builder
	valueStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary"))
	for i, val := range s.WeeklyData {
		valStr := fmt.Sprintf("%d", val)
		valueLine.WriteString(valueStyle.Render(layout.PadCenter(valStr, barWidth-1)))
		if i < 6 {
			valueLine.WriteString(" ")
		}
	}
	lines = append(lines, valueLine.String())

	return lines
}

func (s *StatsView) renderTopCategories(width int, th *theme.Theme, t *i18n.Translations) []string {
	var lines []string

	titleStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary")).Bold(true)
	lines = append(lines, titleStyle.Render(t.Stats.ByCategory))

	if len(s.Stats.ByCategory) == 0 {
		mutedStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted")).Italic(true)
		lines = append(lines, mutedStyle.Render("  No category data available"))
		return lines
	}

	// Sort categories by duration
	type catDur struct {
		name     string
		duration time.Duration
	}
	var sorted []catDur
	for name, dur := range s.Stats.ByCategory {
		sorted = append(sorted, catDur{name, dur})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].duration > sorted[j].duration
	})

	// Show top 5
	maxShow := 5
	if len(sorted) < maxShow {
		maxShow = len(sorted)
	}

	// Find max for bar scaling
	maxDur := sorted[0].duration
	barMaxWidth := width - 25
	if barMaxWidth < 10 {
		barMaxWidth = 10
	}

	categoryColors := []lipgloss.Color{
		th.Color("success"),
		th.Color("info"),
		th.Color("warning"),
		th.Color("accent"),
		th.Color("text_secondary"),
	}

	for i := 0; i < maxShow; i++ {
		cat := sorted[i]

		// Name
		nameStyle := lipgloss.NewStyle().Foreground(th.Color("text_primary"))
		name := layout.FitToWidth(nameStyle.Render(cat.name), 12)

		// Duration
		durStr := formatDuration(cat.duration)
		durStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary"))

		// Bar
		ratio := float64(cat.duration) / float64(maxDur)
		barWidth := int(float64(barMaxWidth) * ratio)
		if barWidth < 1 {
			barWidth = 1
		}
		barStyle := lipgloss.NewStyle().Foreground(categoryColors[i%len(categoryColors)])
		bar := barStyle.Render(strings.Repeat("█", barWidth))

		line := fmt.Sprintf("  %s %s %s", name, durStyle.Render(layout.FitToWidth(durStr, 8)), bar)
		lines = append(lines, line)
	}

	return lines
}

func (s *StatsView) renderHeatmap() string {
	th := theme.Current
	t := i18n.T()

	frame := layout.NewFrame(s.Width, s.Height)
	frame.SetTitle(icons.Get("stats") + " Activity Heatmap")
	frame.SetBorder(layout.BorderSingle)
	frame.SetFocused(s.Focused)
	frame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)

	var lines []string
	contentWidth := frame.ContentWidth()

	// Year heatmap (GitHub style)
	lines = append(lines, s.renderYearHeatmap(contentWidth, th)...)
	lines = append(lines, "")

	// Legend
	lines = append(lines, s.renderHeatmapLegend(th)...)
	lines = append(lines, "")

	// Stats summary
	lines = append(lines, s.renderHeatmapStats(th, t)...)

	// Navigation hint
	lines = append(lines, "")
	navHint := lipgloss.NewStyle().
		Foreground(th.Color("text_muted")).
		Render("[Tab] Switch view  [o] Overview  [c] Categories")
	lines = append(lines, navHint)

	frame.SetContentLines(lines)
	return frame.Render()
}

func (s *StatsView) renderYearHeatmap(width int, th *theme.Theme) []string {
	var lines []string

	titleStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary")).Bold(true)
	lines = append(lines, titleStyle.Render("Year Activity"))
	lines = append(lines, "")

	// Heatmap colors (0-4 intensity levels)
	heatColors := []lipgloss.Color{
		th.Color("text_muted"),      // 0
		th.Color("heatmap_1"),       // 1-2
		th.Color("heatmap_2"),       // 3-5
		th.Color("heatmap_3"),       // 6-9
		th.Color("heatmap_4"),       // 10+
	}

	// Generate last 52 weeks (7 rows for days of week)
	now := time.Now()

	// Month labels row
	monthLabels := strings.Repeat(" ", 4) // space for day labels
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	weeksPerMonth := (width - 4) / 52
	if weeksPerMonth < 1 {
		weeksPerMonth = 1
	}

	currentMonth := -1
	for week := 0; week < 52; week++ {
		weekStart := now.AddDate(0, 0, -((52-week)*7))
		if int(weekStart.Month()) != currentMonth {
			currentMonth = int(weekStart.Month())
			if week < 52-3 { // don't show label if it won't fit
				monthLabels += months[currentMonth-1]
			}
		}
		monthLabels += " "
	}
	monthLabelStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
	lines = append(lines, monthLabelStyle.Render(layout.TruncateToWidth(monthLabels, width)))

	// Day rows (Mon, Tue, Wed, Thu, Fri, Sat, Sun)
	dayLabels := []string{"Mon", "   ", "Wed", "   ", "Fri", "   ", "Sun"}

	for row := 0; row < 7; row++ {
		var rowStr strings.Builder

		// Day label
		dayLabelStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
		rowStr.WriteString(dayLabelStyle.Render(dayLabels[row]))
		rowStr.WriteString(" ")

		// 52 weeks of data
		for week := 0; week < 52; week++ {
			// Calculate date
			daysAgo := (52-week)*7 + (6 - row) // adjust for day of week
			date := now.AddDate(0, 0, -daysAgo)
			dateStr := date.Format("2006-01-02")

			// Get activity count
			count := 0
			if s.YearlyData != nil {
				count = s.YearlyData[dateStr]
			}

			// Determine color intensity
			var colorIdx int
			switch {
			case count == 0:
				colorIdx = 0
			case count <= 2:
				colorIdx = 1
			case count <= 5:
				colorIdx = 2
			case count <= 9:
				colorIdx = 3
			default:
				colorIdx = 4
			}

			cellStyle := lipgloss.NewStyle().Foreground(heatColors[colorIdx])
			rowStr.WriteString(cellStyle.Render("█"))
		}

		lines = append(lines, rowStr.String())
	}

	return lines
}

func (s *StatsView) renderHeatmapLegend(th *theme.Theme) []string {
	var lines []string

	legendStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))

	heatColors := []lipgloss.Color{
		th.Color("text_muted"),
		th.Color("heatmap_1"),
		th.Color("heatmap_2"),
		th.Color("heatmap_3"),
		th.Color("heatmap_4"),
	}

	legend := legendStyle.Render("Less ")
	for _, c := range heatColors {
		style := lipgloss.NewStyle().Foreground(c)
		legend += style.Render("█")
	}
	legend += legendStyle.Render(" More")

	lines = append(lines, legend)
	return lines
}

func (s *StatsView) renderHeatmapStats(th *theme.Theme, t *i18n.Translations) []string {
	var lines []string

	labelStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary"))
	valueStyle := lipgloss.NewStyle().Foreground(th.Color("text_primary")).Bold(true)

	// Calculate totals from yearly data
	totalDays := 0
	totalPomodoros := 0
	maxDay := 0
	for _, count := range s.YearlyData {
		if count > 0 {
			totalDays++
			totalPomodoros += count
			if count > maxDay {
				maxDay = count
			}
		}
	}

	line1 := fmt.Sprintf("%s %s  %s %s  %s %s",
		labelStyle.Render("Active days:"),
		valueStyle.Render(fmt.Sprintf("%d", totalDays)),
		labelStyle.Render("Total:"),
		valueStyle.Render(fmt.Sprintf("%d", totalPomodoros)),
		labelStyle.Render("Best day:"),
		valueStyle.Render(fmt.Sprintf("%d", maxDay)),
	)
	lines = append(lines, line1)

	// Streak info
	line2 := fmt.Sprintf("%s %s  %s %s",
		labelStyle.Render(t.Stats.CurrentStreak+":"),
		valueStyle.Render(fmt.Sprintf("%d days", s.Stats.CurrentStreak)),
		labelStyle.Render(t.Stats.LongestStreak+":"),
		valueStyle.Render(fmt.Sprintf("%d days", s.Stats.LongestStreak)),
	)
	lines = append(lines, line2)

	return lines
}

func (s *StatsView) renderCategories() string {
	th := theme.Current
	t := i18n.T()

	frame := layout.NewFrame(s.Width, s.Height)
	frame.SetTitle(icons.Get("stats") + " " + t.Stats.ByCategory)
	frame.SetBorder(layout.BorderSingle)
	frame.SetFocused(s.Focused)
	frame.SetColors(
		th.Color("border_default"),
		th.Color("border_active"),
		th.Color("text_primary"),
		th.Color("bg_primary"),
	)

	var lines []string
	contentWidth := frame.ContentWidth()

	if len(s.Stats.ByCategory) == 0 {
		mutedStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted")).Italic(true)
		lines = append(lines, mutedStyle.Render("No category data available"))
	} else {
		lines = append(lines, s.renderDetailedCategories(contentWidth, th)...)
	}

	// Navigation hint
	lines = append(lines, "")
	navHint := lipgloss.NewStyle().
		Foreground(th.Color("text_muted")).
		Render("[Tab] Switch view  [o] Overview  [h] Heatmap")
	lines = append(lines, navHint)

	frame.SetContentLines(lines)
	return frame.Render()
}

func (s *StatsView) renderDetailedCategories(width int, th *theme.Theme) []string {
	var lines []string

	// Sort categories by duration
	type catDur struct {
		name     string
		duration time.Duration
	}
	var sorted []catDur
	totalDuration := time.Duration(0)
	for name, dur := range s.Stats.ByCategory {
		sorted = append(sorted, catDur{name, dur})
		totalDuration += dur
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].duration > sorted[j].duration
	})

	categoryColors := []lipgloss.Color{
		th.Color("success"),
		th.Color("info"),
		th.Color("warning"),
		th.Color("accent"),
		th.Color("error"),
		th.Color("text_secondary"),
	}

	maxDur := sorted[0].duration
	barMaxWidth := width - 35
	if barMaxWidth < 10 {
		barMaxWidth = 10
	}

	for i, cat := range sorted {
		// Percentage
		pct := float64(cat.duration) / float64(totalDuration) * 100
		pctStr := fmt.Sprintf("%5.1f%%", pct)

		// Name
		nameStyle := lipgloss.NewStyle().Foreground(th.Color("text_primary"))
		name := layout.FitToWidth(nameStyle.Render(cat.name), 15)

		// Duration
		durStr := formatDuration(cat.duration)
		durStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary"))

		// Bar
		ratio := float64(cat.duration) / float64(maxDur)
		barWidth := int(float64(barMaxWidth) * ratio)
		if barWidth < 1 {
			barWidth = 1
		}
		barStyle := lipgloss.NewStyle().Foreground(categoryColors[i%len(categoryColors)])
		bar := barStyle.Render(strings.Repeat("█", barWidth))

		pctStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
		line := fmt.Sprintf("%s %s %s %s", pctStyle.Render(pctStr), name, durStyle.Render(layout.FitToWidth(durStr, 10)), bar)
		lines = append(lines, line)
	}

	// Total
	lines = append(lines, "")
	totalStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary")).Bold(true)
	lines = append(lines, totalStyle.Render(fmt.Sprintf("Total: %s", formatDuration(totalDuration))))

	return lines
}

// Helper function to format duration
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours >= 24 {
		days := hours / 24
		hours = hours % 24
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
