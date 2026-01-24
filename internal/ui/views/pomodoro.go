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
)

// PomodoroTimerState represents the state of the pomodoro timer for fullscreen view.
type PomodoroTimerState int

const (
	TimerStateIdle PomodoroTimerState = iota
	TimerStateRunning
	TimerStatePaused
	TimerStateBreak
)

// PomodoroContext represents a context/category for pomodoro sessions.
type PomodoroContext struct {
	Name  string
	Color lipgloss.Color
	Icon  string
}

// PomodoroView represents the fullscreen pomodoro view.
type PomodoroView struct {
	Width  int
	Height int

	// Timer state
	State        PomodoroTimerState
	Remaining    time.Duration
	TotalTime    time.Duration
	IsShortBreak bool

	// Session info
	SessionsToday int
	DailyGoal     int
	CurrentStreak int

	// Context
	Context         *PomodoroContext
	AvailableCtx    []PomodoroContext
	ContextSelected int

	// Configuration
	WorkDuration    time.Duration
	ShortBreak      time.Duration
	LongBreak       time.Duration
	SessionsForLong int

	// UI state
	ShowContextPicker bool
	Focused           bool
}

// NewPomodoroView creates a new pomodoro fullscreen view.
func NewPomodoroView(width, height int) *PomodoroView {
	return &PomodoroView{
		Width:           width,
		Height:          height,
		State:           TimerStateIdle,
		Remaining:       25 * time.Minute,
		TotalTime:       25 * time.Minute,
		WorkDuration:    25 * time.Minute,
		ShortBreak:      5 * time.Minute,
		LongBreak:       15 * time.Minute,
		SessionsForLong: 4,
		DailyGoal:       8,
		AvailableCtx: []PomodoroContext{
			{Name: "Work", Color: lipgloss.Color("#ff6b6b"), Icon: "work"},
			{Name: "Study", Color: lipgloss.Color("#4ecdc4"), Icon: "book"},
			{Name: "Code", Color: lipgloss.Color("#45b7d1"), Icon: "code"},
			{Name: "Write", Color: lipgloss.Color("#96ceb4"), Icon: "edit"},
			{Name: "Other", Color: lipgloss.Color("#a8a8a8"), Icon: "note"},
		},
	}
}

// SetSize updates the view dimensions.
func (p *PomodoroView) SetSize(width, height int) {
	p.Width = width
	p.Height = height
}

// SetState updates the timer state.
func (p *PomodoroView) SetState(state PomodoroTimerState) {
	p.State = state
}

// SetRemaining updates the remaining time.
func (p *PomodoroView) SetRemaining(remaining time.Duration) {
	p.Remaining = remaining
}

// SetTotalTime sets the total duration for current session.
func (p *PomodoroView) SetTotalTime(total time.Duration) {
	p.TotalTime = total
}

// SetSessions updates session counts.
func (p *PomodoroView) SetSessions(today, goal int) {
	p.SessionsToday = today
	p.DailyGoal = goal
}

// SetContext sets the current pomodoro context.
func (p *PomodoroView) SetContext(ctx *PomodoroContext) {
	p.Context = ctx
}

// ToggleContextPicker toggles the context picker visibility.
func (p *PomodoroView) ToggleContextPicker() {
	p.ShowContextPicker = !p.ShowContextPicker
}

// SelectNextContext moves selection down in context picker.
func (p *PomodoroView) SelectNextContext() {
	if p.ContextSelected < len(p.AvailableCtx)-1 {
		p.ContextSelected++
	}
}

// SelectPrevContext moves selection up in context picker.
func (p *PomodoroView) SelectPrevContext() {
	if p.ContextSelected > 0 {
		p.ContextSelected--
	}
}

// ConfirmContextSelection confirms the current context selection.
func (p *PomodoroView) ConfirmContextSelection() {
	if p.ContextSelected >= 0 && p.ContextSelected < len(p.AvailableCtx) {
		ctx := p.AvailableCtx[p.ContextSelected]
		p.Context = &ctx
	}
	p.ShowContextPicker = false
}

// Render renders the fullscreen pomodoro view.
func (p *PomodoroView) Render() string {
	if p.ShowContextPicker {
		return p.renderContextPicker()
	}
	return p.renderTimer()
}

func (p *PomodoroView) renderTimer() string {
	th := theme.Current
	t := i18n.T()
	screen := layout.NewScreen(p.Width, p.Height)

	// Calculate vertical positioning
	centerY := p.Height / 2

	// Render big timer digits
	timerY := centerY - 5
	p.renderBigTimer(screen, timerY, th)

	// Render status below timer
	statusY := timerY + 8
	p.renderStatus(screen, statusY, th, t)

	// Render progress bar
	progressY := statusY + 2
	p.renderProgressBar(screen, progressY, th)

	// Render context info
	contextY := progressY + 3
	p.renderContextInfo(screen, contextY, th)

	// Render session stats
	statsY := p.Height - 5
	p.renderSessionStats(screen, statsY, th, t)

	// Render help at bottom
	helpY := p.Height - 2
	p.renderHelp(screen, helpY, th)

	return screen.String()
}

func (p *PomodoroView) renderBigTimer(screen *layout.Screen, y int, th *theme.Theme) {
	minutes := int(p.Remaining.Minutes())
	seconds := int(p.Remaining.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	// Create big ASCII art digits
	bigDigits := p.createBigDigits(timeStr)

	// Choose color based on state
	var style lipgloss.Style
	switch p.State {
	case TimerStateRunning:
		style = lipgloss.NewStyle().Foreground(th.Color("success"))
	case TimerStatePaused:
		style = lipgloss.NewStyle().Foreground(th.Color("warning"))
	case TimerStateBreak:
		style = lipgloss.NewStyle().Foreground(th.Color("info"))
	default:
		style = lipgloss.NewStyle().Foreground(th.Color("text_primary"))
	}

	// Center and render each line of the big timer
	for i, line := range bigDigits {
		styledLine := style.Render(line)
		x := (p.Width - lipgloss.Width(line)) / 2
		if x < 0 {
			x = 0
		}
		screen.DrawBlock(x, y+i, styledLine)
	}
}

func (p *PomodoroView) createBigDigits(timeStr string) []string {
	// ASCII art digit patterns (5 lines high)
	digits := map[rune][]string{
		'0': {
			" ██████  ",
			"██    ██ ",
			"██    ██ ",
			"██    ██ ",
			" ██████  ",
		},
		'1': {
			"   ██    ",
			"  ███    ",
			"   ██    ",
			"   ██    ",
			" ██████  ",
		},
		'2': {
			" ██████  ",
			"      ██ ",
			" ██████  ",
			"██       ",
			" ██████  ",
		},
		'3': {
			" ██████  ",
			"      ██ ",
			"  █████  ",
			"      ██ ",
			" ██████  ",
		},
		'4': {
			"██    ██ ",
			"██    ██ ",
			" ███████ ",
			"      ██ ",
			"      ██ ",
		},
		'5': {
			" ██████  ",
			"██       ",
			" ██████  ",
			"      ██ ",
			" ██████  ",
		},
		'6': {
			" ██████  ",
			"██       ",
			" ██████  ",
			"██    ██ ",
			" ██████  ",
		},
		'7': {
			" ██████  ",
			"      ██ ",
			"     ██  ",
			"    ██   ",
			"   ██    ",
		},
		'8': {
			" ██████  ",
			"██    ██ ",
			" ██████  ",
			"██    ██ ",
			" ██████  ",
		},
		'9': {
			" ██████  ",
			"██    ██ ",
			" ██████  ",
			"      ██ ",
			" ██████  ",
		},
		':': {
			"         ",
			"   ██    ",
			"         ",
			"   ██    ",
			"         ",
		},
	}

	lines := make([]string, 5)
	for _, char := range timeStr {
		if pattern, ok := digits[char]; ok {
			for i := 0; i < 5; i++ {
				lines[i] += pattern[i]
			}
		}
	}

	return lines
}

func (p *PomodoroView) renderStatus(screen *layout.Screen, y int, th *theme.Theme, t *i18n.Translations) {
	var statusText string
	var style lipgloss.Style

	switch p.State {
	case TimerStateIdle:
		statusText = icons.Get("pomodoro") + " " + t.Pomodoro.Ready
		style = lipgloss.NewStyle().Foreground(th.Color("text_secondary"))
	case TimerStateRunning:
		statusText = icons.Get("pomodoro") + " " + t.Pomodoro.Work + " - " + t.Pomodoro.Running
		style = lipgloss.NewStyle().Foreground(th.Color("success")).Bold(true)
	case TimerStatePaused:
		statusText = icons.Get("pause") + " " + t.Pomodoro.Paused
		style = lipgloss.NewStyle().Foreground(th.Color("warning")).Bold(true)
	case TimerStateBreak:
		breakType := t.Pomodoro.ShortBreak
		if !p.IsShortBreak {
			breakType = t.Pomodoro.LongBreak
		}
		statusText = icons.Get("break") + " " + breakType
		style = lipgloss.NewStyle().Foreground(th.Color("info")).Bold(true)
	}

	styledStatus := style.Render(statusText)
	x := (p.Width - lipgloss.Width(statusText)) / 2
	if x < 0 {
		x = 0
	}
	screen.DrawBlock(x, y, styledStatus)
}

func (p *PomodoroView) renderProgressBar(screen *layout.Screen, y int, th *theme.Theme) {
	barWidth := p.Width - 20
	if barWidth < 20 {
		barWidth = 20
	}
	if barWidth > 60 {
		barWidth = 60
	}

	progress := 1.0
	if p.TotalTime > 0 {
		progress = 1.0 - float64(p.Remaining)/float64(p.TotalTime)
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	filledWidth := int(float64(barWidth) * progress)
	emptyWidth := barWidth - filledWidth

	// Choose color based on state
	var filledColor lipgloss.Color
	switch p.State {
	case TimerStateRunning:
		filledColor = th.Color("success")
	case TimerStatePaused:
		filledColor = th.Color("warning")
	case TimerStateBreak:
		filledColor = th.Color("info")
	default:
		filledColor = th.Color("text_secondary")
	}

	filledStyle := lipgloss.NewStyle().Foreground(filledColor)
	emptyStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))

	filled := filledStyle.Render(strings.Repeat("█", filledWidth))
	empty := emptyStyle.Render(strings.Repeat("░", emptyWidth))
	bar := "│" + filled + empty + "│"

	// Add percentage
	percent := int(progress * 100)
	percentStr := fmt.Sprintf(" %d%%", percent)
	bar += lipgloss.NewStyle().Foreground(th.Color("text_secondary")).Render(percentStr)

	x := (p.Width - lipgloss.Width(bar)) / 2
	if x < 0 {
		x = 0
	}
	screen.DrawBlock(x, y, bar)
}

func (p *PomodoroView) renderContextInfo(screen *layout.Screen, y int, th *theme.Theme) {
	if p.Context == nil {
		text := lipgloss.NewStyle().
			Foreground(th.Color("text_muted")).
			Italic(true).
			Render("Press 'c' to select context")
		x := (p.Width - lipgloss.Width(text)) / 2
		screen.DrawBlock(x, y, text)
		return
	}

	icon := icons.Get(p.Context.Icon)
	contextStyle := lipgloss.NewStyle().
		Foreground(p.Context.Color).
		Bold(true)
	text := contextStyle.Render(icon + " " + p.Context.Name)
	x := (p.Width - lipgloss.Width(text)) / 2
	screen.DrawBlock(x, y, text)
}

func (p *PomodoroView) renderSessionStats(screen *layout.Screen, y int, th *theme.Theme, t *i18n.Translations) {
	_ = t // translations available for future use

	// Sessions today
	sessionsText := fmt.Sprintf("%s %d/%d", icons.Get("pomodoro"), p.SessionsToday, p.DailyGoal)
	sessionsStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary"))

	// Progress towards daily goal
	goalProgress := float64(p.SessionsToday) / float64(p.DailyGoal)
	if goalProgress > 1 {
		goalProgress = 1
	}

	miniBarWidth := 10
	filledWidth := int(float64(miniBarWidth) * goalProgress)
	emptyWidth := miniBarWidth - filledWidth

	var barColor lipgloss.Color
	if goalProgress >= 1 {
		barColor = th.Color("success")
	} else if goalProgress >= 0.5 {
		barColor = th.Color("warning")
	} else {
		barColor = th.Color("text_muted")
	}

	miniBar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("●", filledWidth)) +
		lipgloss.NewStyle().Foreground(th.Color("text_muted")).Render(strings.Repeat("○", emptyWidth))

	statsLine := sessionsStyle.Render(sessionsText) + "  " + miniBar
	x := (p.Width - lipgloss.Width(statsLine)) / 2
	screen.DrawBlock(x, y, statsLine)

	// Show "Daily goal reached!" message if goal is met
	if p.SessionsToday >= p.DailyGoal {
		successMsg := lipgloss.NewStyle().
			Foreground(th.Color("success")).
			Bold(true).
			Render(icons.Get("check") + " Daily goal reached!")
		x := (p.Width - lipgloss.Width(successMsg)) / 2
		screen.DrawBlock(x, y+1, successMsg)
	}
}

func (p *PomodoroView) renderHelp(screen *layout.Screen, y int, th *theme.Theme) {
	var helpItems []string

	switch p.State {
	case TimerStateIdle:
		helpItems = []string{
			"[Enter] Start",
			"[c] Context",
			"[+/-] Adjust time",
			"[q] Quit",
		}
	case TimerStateRunning:
		helpItems = []string{
			"[Space] Pause",
			"[s] Stop",
			"[+/-] Adjust time",
			"[q] Quit",
		}
	case TimerStatePaused:
		helpItems = []string{
			"[Space] Resume",
			"[s] Stop",
			"[+/-] Adjust time",
			"[q] Quit",
		}
	case TimerStateBreak:
		helpItems = []string{
			"[Enter] Skip break",
			"[s] Stop",
			"[q] Quit",
		}
	}

	helpStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
	keyStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary"))

	var parts []string
	for _, item := range helpItems {
		// Find the key part [X]
		if strings.HasPrefix(item, "[") {
			idx := strings.Index(item, "]")
			if idx > 0 {
				key := keyStyle.Render(item[:idx+1])
				desc := helpStyle.Render(item[idx+1:])
				parts = append(parts, key+desc)
				continue
			}
		}
		parts = append(parts, helpStyle.Render(item))
	}

	helpLine := strings.Join(parts, "  ")
	x := (p.Width - lipgloss.Width(helpLine)) / 2
	if x < 0 {
		x = 0
	}
	screen.DrawBlock(x, y, helpLine)
}

func (p *PomodoroView) renderContextPicker() string {
	th := theme.Current
	screen := layout.NewScreen(p.Width, p.Height)

	// Title
	titleY := p.Height/2 - len(p.AvailableCtx)/2 - 3
	title := lipgloss.NewStyle().
		Foreground(th.Color("text_primary")).
		Bold(true).
		Render("Select Context")
	titleX := (p.Width - lipgloss.Width(title)) / 2
	screen.DrawBlock(titleX, titleY, title)

	// Context list
	listY := titleY + 2
	maxWidth := 0
	for _, ctx := range p.AvailableCtx {
		w := lipgloss.Width(icons.Get(ctx.Icon) + " " + ctx.Name)
		if w > maxWidth {
			maxWidth = w
		}
	}

	for i, ctx := range p.AvailableCtx {
		icon := icons.Get(ctx.Icon)
		itemText := icon + " " + ctx.Name

		var style lipgloss.Style
		if i == p.ContextSelected {
			style = lipgloss.NewStyle().
				Foreground(th.Color("bg_primary")).
				Background(ctx.Color).
				Bold(true).
				Padding(0, 1)
		} else {
			style = lipgloss.NewStyle().
				Foreground(ctx.Color).
				Padding(0, 1)
		}

		styledItem := style.Render(layout.FitToWidth(itemText, maxWidth))
		x := (p.Width - lipgloss.Width(styledItem)) / 2
		screen.DrawBlock(x, listY+i, styledItem)
	}

	// Help
	helpY := listY + len(p.AvailableCtx) + 2
	helpStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted"))
	helpText := helpStyle.Render("[j/k] Navigate  [Enter] Select  [Esc] Cancel")
	helpX := (p.Width - lipgloss.Width(helpText)) / 2
	screen.DrawBlock(helpX, helpY, helpText)

	return screen.String()
}

// PomodoroWidget is a compact version for the sidebar.
type PomodoroWidget struct {
	Width     int
	Height    int
	State     PomodoroTimerState
	Remaining time.Duration
	Context   string
}

// NewPomodoroWidget creates a new pomodoro widget.
func NewPomodoroWidget(width, height int) *PomodoroWidget {
	return &PomodoroWidget{
		Width:     width,
		Height:    height,
		State:     TimerStateIdle,
		Remaining: 25 * time.Minute,
	}
}

// SetSize updates widget dimensions.
func (w *PomodoroWidget) SetSize(width, height int) {
	w.Width = width
	w.Height = height
}

// Update updates the widget state.
func (w *PomodoroWidget) Update(state PomodoroTimerState, remaining time.Duration, context string) {
	w.State = state
	w.Remaining = remaining
	w.Context = context
}

// Render renders the compact pomodoro widget.
func (w *PomodoroWidget) Render() string {
	th := theme.Current
	t := i18n.T()

	var lines []string

	// Timer line
	minutes := int(w.Remaining.Minutes())
	seconds := int(w.Remaining.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	var timerStyle lipgloss.Style
	switch w.State {
	case TimerStateRunning:
		timerStyle = lipgloss.NewStyle().Foreground(th.Color("success")).Bold(true)
	case TimerStatePaused:
		timerStyle = lipgloss.NewStyle().Foreground(th.Color("warning")).Bold(true)
	case TimerStateBreak:
		timerStyle = lipgloss.NewStyle().Foreground(th.Color("info")).Bold(true)
	default:
		timerStyle = lipgloss.NewStyle().Foreground(th.Color("text_primary"))
	}

	timerLine := icons.Get("pomodoro") + " " + timerStyle.Render(timeStr)
	lines = append(lines, layout.FitToWidth(timerLine, w.Width))

	// Status line
	var statusText string
	switch w.State {
	case TimerStateIdle:
		statusText = "[" + t.Pomodoro.Ready + "]"
	case TimerStateRunning:
		statusText = "[" + t.Pomodoro.Running + "]"
	case TimerStatePaused:
		statusText = "[" + t.Pomodoro.Paused + "]"
	case TimerStateBreak:
		statusText = "[" + t.Pomodoro.Break + "]"
	}

	statusStyle := lipgloss.NewStyle().Foreground(th.Color("text_secondary"))
	lines = append(lines, layout.FitToWidth(statusStyle.Render(statusText), w.Width))

	// Context line (if set)
	if w.Context != "" && w.Height >= 3 {
		ctxStyle := lipgloss.NewStyle().Foreground(th.Color("text_muted")).Italic(true)
		lines = append(lines, layout.FitToWidth(ctxStyle.Render(w.Context), w.Width))
	}

	return strings.Join(lines, "\n")
}
