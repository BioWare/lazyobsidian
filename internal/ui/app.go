// Package ui implements the terminal user interface using BubbleTea.
package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BioWare/lazyobsidian/internal/cache"
	"github.com/BioWare/lazyobsidian/internal/config"
	"github.com/BioWare/lazyobsidian/internal/ui/theme"
	"github.com/BioWare/lazyobsidian/internal/ui/views"
	"github.com/BioWare/lazyobsidian/internal/vault"
	"github.com/BioWare/lazyobsidian/internal/watcher"
	"github.com/BioWare/lazyobsidian/pkg/types"
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
	cache       *cache.Cache
	parser      *vault.Parser
	watcher     *watcher.Watcher
	width       int
	height      int
	currentView View
	focus       FocusArea
	sidebar     *Sidebar
	quitting    bool
	err         error

	// Data loaded from vault
	todayTasks    []types.Task
	activeCourses []types.Course
	currentBook   *types.Book
	recentNotes   []views.RecentNote
	dailyGoal     *types.DailyGoal
	weeklyStats   views.WeeklyStats
}

// New creates a new App instance.
func New(cfg *config.Config, c *cache.Cache, p *vault.Parser, w *watcher.Watcher) *App {
	return &App{
		config:      cfg,
		cache:       c,
		parser:      p,
		watcher:     w,
		currentView: ViewDashboard,
		focus:       FocusSidebar,
		sidebar:     NewSidebar(),
	}
}

// Custom message types
type dataLoadedMsg struct{}
type fileChangedMsg struct {
	path string
}
type tickMsg time.Time

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.loadInitialData(),
		a.startFileWatcher(),
	)
}

// loadInitialData loads data from the vault and cache.
func (a *App) loadInitialData() tea.Cmd {
	return func() tea.Msg {
		// Parse and cache today's daily note
		today := time.Now()
		if a.parser.DailyNoteExists(today) {
			file, err := a.parser.ParseDailyNote(today)
			if err == nil && file != nil {
				a.cache.SaveFile(file)
				a.todayTasks = file.Tasks
			}
		}

		// Load courses from cache/vault
		courseFiles, _ := a.cache.GetFilesByType(types.FileTypeCourse)
		if len(courseFiles) == 0 {
			// Parse from vault if cache is empty
			files, _ := a.parser.ParseVault()
			for _, f := range files {
				a.cache.SaveFile(f)
			}
			courseFiles, _ = a.cache.GetFilesByType(types.FileTypeCourse)
		}

		// Convert course files to Course structs
		for _, f := range courseFiles {
			course := a.parseCourseFromFile(f)
			if course != nil {
				a.activeCourses = append(a.activeCourses, *course)
			}
		}

		// Load books
		bookFiles, _ := a.cache.GetFilesByType(types.FileTypeBook)
		for _, f := range bookFiles {
			book := a.parseBookFromFile(f)
			if book != nil {
				a.currentBook = book
				break // Just get the first/most recent book
			}
		}

		// Load recent notes
		recentFiles, _ := a.cache.GetRecentFiles(5)
		for _, f := range recentFiles {
			a.recentNotes = append(a.recentNotes, views.RecentNote{
				Title:    f.Title,
				Path:     f.Path,
				Modified: f.ModifiedAt,
			})
		}

		// Load daily goal
		a.dailyGoal, _ = a.cache.GetDailyGoal(today)
		if a.dailyGoal == nil {
			// Create default daily goal
			a.cache.SetDailyGoal(today, a.config.Pomodoro.DailyGoal)
			a.dailyGoal = &types.DailyGoal{
				Date:      today,
				Target:    a.config.Pomodoro.DailyGoal,
				Completed: 0,
			}
		}

		// Load weekly stats
		byDay, totalPom, totalMin, _ := a.cache.GetWeeklyStats()
		a.weeklyStats = views.WeeklyStats{
			Pomodoros: totalPom,
			FocusTime: time.Duration(totalMin) * time.Minute,
			ByDay:     byDay,
			Streak:    0, // TODO: Calculate streak from data
		}

		return dataLoadedMsg{}
	}
}

// startFileWatcher starts listening for file changes.
func (a *App) startFileWatcher() tea.Cmd {
	if a.watcher == nil {
		return nil
	}

	return func() tea.Msg {
		// Start the watcher if not already started
		a.watcher.Start()

		// Wait for first event (blocks)
		select {
		case event := <-a.watcher.Events:
			return fileChangedMsg{path: event.Path}
		case <-time.After(time.Second):
			// Timeout, return empty to re-poll
			return nil
		}
	}
}

// waitForFileEvent waits for the next file event.
func (a *App) waitForFileEvent() tea.Cmd {
	if a.watcher == nil {
		return nil
	}

	return func() tea.Msg {
		select {
		case event := <-a.watcher.Events:
			return fileChangedMsg{path: event.Path}
		case <-time.After(time.Second):
			return nil
		}
	}
}

// parseCourseFromFile converts a File to a Course.
func (a *App) parseCourseFromFile(f *types.File) *types.Course {
	if f == nil || f.Frontmatter == nil {
		return nil
	}

	course := &types.Course{
		FileID: f.ID,
		Title:  f.Title,
	}

	if source, ok := f.Frontmatter["source"].(string); ok {
		course.Source = source
	}
	if url, ok := f.Frontmatter["url"].(string); ok {
		course.URL = url
	}

	// Count completed and total lessons from tasks
	completed, total := vault.CountTasks(f.Tasks)
	course.TotalLessons = total
	course.Completed = completed

	return course
}

// parseBookFromFile converts a File to a Book.
func (a *App) parseBookFromFile(f *types.File) *types.Book {
	if f == nil || f.Frontmatter == nil {
		return nil
	}

	book := &types.Book{
		FileID: f.ID,
		Title:  f.Title,
	}

	if author, ok := f.Frontmatter["author"].(string); ok {
		book.Author = author
	}
	if pages, ok := f.Frontmatter["total_pages"].(string); ok {
		fmt.Sscanf(pages, "%d", &book.TotalPages)
	}
	if current, ok := f.Frontmatter["current_page"].(string); ok {
		fmt.Sscanf(current, "%d", &book.CurrentPage)
	}

	// Count completed chapters from tasks
	completed, total := vault.CountTasks(f.Tasks)
	if book.TotalPages == 0 {
		book.TotalPages = total
		book.CurrentPage = completed
	}

	return book
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

	case dataLoadedMsg:
		// Data loaded, nothing special to do
		return a, nil

	case fileChangedMsg:
		// Re-parse the changed file
		if msg.path != "" {
			file, err := a.parser.ParseFile(msg.path)
			if err == nil && file != nil {
				a.cache.SaveFile(file)

				// If it's today's daily note, update tasks
				if file.Type == types.FileTypeDaily {
					a.todayTasks = file.Tasks
				}
			}
		}
		// Continue listening for file events
		return a, a.waitForFileEvent()

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

		// Set real data from vault
		dashboard.Tasks = a.todayTasks
		dashboard.ActiveCourses = a.activeCourses
		dashboard.CurrentBook = a.currentBook
		dashboard.RecentNotes = a.recentNotes
		dashboard.WeeklyStats = a.weeklyStats

		// Set pomodoro state
		dailyGoal := 5
		dailyDone := 0
		if a.dailyGoal != nil {
			dailyGoal = a.dailyGoal.Target
			dailyDone = a.dailyGoal.Completed
		}
		dashboard.PomodoroState = views.PomodoroState{
			State:     "ready",
			Remaining: time.Duration(a.config.Pomodoro.WorkMinutes) * time.Minute,
			DailyGoal: dailyGoal,
			DailyDone: dailyDone,
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

	// Initialize cache
	c, err := cache.New(cfg.Vault.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}
	defer c.Close()

	// Initialize parser
	parser := vault.NewParser(cfg.Vault.Path, cfg)

	// Initialize file watcher
	w, err := watcher.New(cfg.Vault.Path)
	if err != nil {
		// Non-fatal: continue without file watching
		w = nil
	}
	if w != nil {
		defer w.Stop()
	}

	app := New(cfg, c, parser, w)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
