// Package ui implements the terminal user interface using BubbleTea.
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/config"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
	"github.com/BioWare/lazyobsidian/internal/ui/views"
)

// View represents the current active view/page.
type View string

const (
	ViewDashboard View = "dashboard"
	ViewCalendar  View = "calendar"
	ViewGoals     View = "goals"
	ViewCourses   View = "courses"
	ViewBooks     View = "books"
	ViewWishlist  View = "wishlist"
	ViewGraph     View = "graph"
	ViewStats     View = "stats"
	ViewSettings  View = "settings"
)

// FocusArea represents which panel has focus.
type FocusArea int

const (
	FocusSidebar FocusArea = iota
	FocusMain
)

// App is the main application model.
type App struct {
	config      *config.Config
	width       int
	height      int
	currentView View
	focus       FocusArea
	sidebar     *Sidebar
	quitting    bool
	err         error
}

// New creates a new App instance.
func New(cfg *config.Config) *App {
	return &App{
		config:      cfg,
		currentView: ViewDashboard,
		focus:       FocusSidebar,
		sidebar:     NewSidebar(),
	}
}

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return a.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		return a, nil

	case error:
		a.err = msg
		return a, nil
	}

	return a, nil
}

func (a *App) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keybindings
	switch msg.String() {
	case "q", "ctrl+c":
		a.quitting = true
		return a, tea.Quit

	case "tab":
		// Toggle focus between sidebar and main panel
		if a.focus == FocusSidebar {
			a.focus = FocusMain
		} else {
			a.focus = FocusSidebar
		}
		return a, nil

	case "?":
		// TODO: Show help modal
		return a, nil

	case "/":
		// TODO: Show global search
		return a, nil
	}

	// Handle navigation based on focus
	if a.focus == FocusSidebar {
		return a.handleSidebarKeys(msg)
	}

	return a.handleMainKeys(msg)
}

func (a *App) handleSidebarKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		a.sidebar.MoveDown()
		a.currentView = a.sidebar.CurrentView()

	case "k", "up":
		a.sidebar.MoveUp()
		a.currentView = a.sidebar.CurrentView()

	case "enter", "l", "right":
		a.focus = FocusMain

	case "g":
		// Handle gg (go to top)
		a.sidebar.MoveToTop()
		a.currentView = a.sidebar.CurrentView()

	case "G":
		a.sidebar.MoveToBottom()
		a.currentView = a.sidebar.CurrentView()

	// Quick navigation with numbers
	case "1":
		a.sidebar.SetIndex(0)
		a.currentView = a.sidebar.CurrentView()
	case "2":
		a.sidebar.SetIndex(1)
		a.currentView = a.sidebar.CurrentView()
	case "3":
		a.sidebar.SetIndex(2)
		a.currentView = a.sidebar.CurrentView()
	case "4":
		a.sidebar.SetIndex(3)
		a.currentView = a.sidebar.CurrentView()
	case "5":
		a.sidebar.SetIndex(4)
		a.currentView = a.sidebar.CurrentView()
	case "6":
		a.sidebar.SetIndex(5)
		a.currentView = a.sidebar.CurrentView()
	case "7":
		a.sidebar.SetIndex(6)
		a.currentView = a.sidebar.CurrentView()
	case "8":
		a.sidebar.SetIndex(7)
		a.currentView = a.sidebar.CurrentView()
	case "9":
		a.sidebar.SetIndex(8)
		a.currentView = a.sidebar.CurrentView()
	}

	return a, nil
}

func (a *App) handleMainKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left", "esc":
		a.focus = FocusSidebar

	case "p":
		// TODO: Start pomodoro
	}

	// TODO: Delegate to current view's handler
	return a, nil
}

// View implements tea.Model.
func (a *App) View() string {
	if a.quitting {
		return ""
	}

	if a.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", a.err)
	}

	if a.width == 0 || a.height == 0 {
		return "Loading..."
	}

	return a.renderLayout()
}

func (a *App) renderLayout() string {
	// Calculate dimensions
	sidebarWidth := a.calculateSidebarWidth()
	mainWidth := a.width - sidebarWidth - 1 // -1 for border

	// Render components
	sidebar := a.sidebar.Render(sidebarWidth, a.height-4, a.focus == FocusSidebar)
	main := a.renderMainPanel(mainWidth, a.height-4)

	// Build layout
	header := a.renderHeader()
	content := joinHorizontal(sidebar, main)
	footer := a.renderFooter()

	return header + "\n" + content + "\n" + footer
}

func (a *App) calculateSidebarWidth() int {
	if !a.config.Display.DynamicWindows.Enabled {
		return a.width * a.config.Display.DynamicWindows.SidebarNormal / 100
	}

	if a.focus == FocusSidebar {
		return a.width * a.config.Display.DynamicWindows.SidebarNormal / 100
	}
	return a.width * a.config.Display.DynamicWindows.SidebarMinimized / 100
}

func (a *App) renderHeader() string {
	title := "LazyObsidian"
	vaultPath := truncatePath(a.config.Vault.Path, 30)

	// Simple header for now
	return fmt.Sprintf(" %s%s%s", title, spacer(a.width-len(title)-len(vaultPath)-4), vaultPath)
}

func (a *App) renderMainPanel(width, height int) string {
	// Title bar
	title := a.getViewTitle()
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3D3428"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#D4C9B5"))

	titleBar := titleStyle.Render(" " + title)

	// Content
	var content string
	contentHeight := height - 2 // minus title and border

	switch a.currentView {
	case ViewDashboard:
		dashboard := views.NewDashboard(width-2, contentHeight)
		// Set demo data for now
		dashboard.PomodoroState = views.PomodoroState{
			State:     "ready",
			Remaining: 25 * time.Minute,
			DailyGoal: 5,
			DailyDone: 0,
		}
		dashboard.WeeklyStats = views.WeeklyStats{
			Pomodoros: 18,
			FocusTime: 7*time.Hour + 30*time.Minute,
			Streak:    12,
			ByDay:     [7]int{4, 5, 3, 2, 4, 0, 0},
		}
		content = dashboard.Render()

	case ViewCalendar:
		content = renderPlaceholderContent("Calendar - Coming soon", width-2, contentHeight)
	case ViewGoals:
		content = renderPlaceholderContent("Goals - Coming soon", width-2, contentHeight)
	case ViewCourses:
		content = renderPlaceholderContent("Courses - Coming soon", width-2, contentHeight)
	case ViewBooks:
		content = renderPlaceholderContent("Books - Coming soon", width-2, contentHeight)
	case ViewWishlist:
		content = renderPlaceholderContent("Wishlist - Coming soon", width-2, contentHeight)
	case ViewGraph:
		content = renderPlaceholderContent("Graph - Coming soon", width-2, contentHeight)
	case ViewStats:
		content = renderPlaceholderContent("Statistics - Coming soon", width-2, contentHeight)
	case ViewSettings:
		content = renderPlaceholderContent("Settings - Coming soon", width-2, contentHeight)
	default:
		content = renderPlaceholderContent("Unknown view", width-2, contentHeight)
	}

	// Build panel with border
	var lines []string
	lines = append(lines, borderStyle.Render("┌─")+titleBar+borderStyle.Render(" "+strings.Repeat("─", width-len(title)-6)+"┐"))

	contentLines := strings.Split(content, "\n")
	for i := 0; i < contentHeight; i++ {
		line := ""
		if i < len(contentLines) {
			line = contentLines[i]
		}
		// Ensure line fits width
		if len(line) > width-4 {
			line = line[:width-5] + "…"
		}
		paddedLine := line + strings.Repeat(" ", max(0, width-len(line)-4))
		lines = append(lines, borderStyle.Render("│ ")+paddedLine+borderStyle.Render(" │"))
	}

	lines = append(lines, borderStyle.Render("└"+strings.Repeat("─", width-2)+"┘"))

	return strings.Join(lines, "\n")
}

func (a *App) getViewTitle() string {
	switch a.currentView {
	case ViewDashboard:
		return "Dashboard"
	case ViewCalendar:
		return "Calendar"
	case ViewGoals:
		return "Goals"
	case ViewCourses:
		return "Courses"
	case ViewBooks:
		return "Books"
	case ViewWishlist:
		return "Wishlist"
	case ViewGraph:
		return "Graph"
	case ViewStats:
		return "Statistics"
	case ViewSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}

func renderPlaceholderContent(text string, width, height int) string {
	var lines []string
	centerY := height / 2
	for i := 0; i < height; i++ {
		if i == centerY {
			padding := (width - len(text)) / 2
			line := strings.Repeat(" ", padding) + text
			lines = append(lines, line)
		} else {
			lines = append(lines, "")
		}
	}
	return strings.Join(lines, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a *App) renderFooter() string {
	hints := "[Tab] focus  [j/k] nav  [Enter] select  [p] pomodoro  [/] search  [?] help  [q] quit"
	return " " + hints
}

// Helper functions

func joinHorizontal(left, right string) string {
	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	var result []string
	for i := 0; i < maxLines; i++ {
		leftLine := ""
		rightLine := ""
		if i < len(leftLines) {
			leftLine = leftLines[i]
		}
		if i < len(rightLines) {
			rightLine = rightLines[i]
		}
		result = append(result, leftLine+"│"+rightLine)
	}

	return strings.Join(result, "\n")
}

func spacer(n int) string {
	if n <= 0 {
		return ""
	}
	return repeatStr(" ", n)
}

func repeatStr(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}

// Run starts the TUI application.
func Run(cfg *config.Config) error {
	// Initialize theme
	t, err := theme.LoadBuiltin(cfg.Theme.Current)
	if err != nil {
		t, _ = theme.LoadBuiltin("corsair-light")
	}
	t.Apply()

	app := New(cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
