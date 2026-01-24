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
	"github.com/BioWare/lazyobsidian/internal/i18n"
	"github.com/BioWare/lazyobsidian/internal/logging"
	"github.com/BioWare/lazyobsidian/internal/pomodoro"
	"github.com/BioWare/lazyobsidian/internal/ui/icons"
	"github.com/BioWare/lazyobsidian/internal/ui/layout"
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

// DashboardModule represents focusable modules in the dashboard.
type DashboardModule int

const (
	ModuleTodayFocus DashboardModule = iota
	ModulePomodoro
	ModuleWeekStats
	ModuleCourses
	ModuleBook
	ModuleRecentNotes
	ModuleCount // Total number of modules
)

// App is the main application model.
type App struct {
	config      *config.Config
	cache       *cache.Cache
	parser      *vault.Parser
	writer      *vault.Writer
	watcher     *watcher.Watcher
	width       int
	height      int
	currentView View
	focus       FocusArea
	sidebar     *Sidebar
	quitting    bool
	err         error

	// Dashboard navigation
	focusedModule  DashboardModule
	selectedTask   int // Selected task index in Today's Focus

	// Pomodoro timer
	pomodoroTimer *pomodoro.Timer

	// Data loaded from vault
	todayTasks      []types.Task
	todayNotePath   string // Path to today's daily note
	activeCourses   []types.Course
	currentBook     *types.Book
	recentNotes     []views.RecentNote
	dailyGoal       *types.DailyGoal
	weeklyStats     views.WeeklyStats

	// Goals data
	goals           []types.Goal
}

// New creates a new App instance.
func New(cfg *config.Config, c *cache.Cache, p *vault.Parser, w *watcher.Watcher) *App {
	// Initialize Pomodoro timer
	timer := pomodoro.NewTimer(pomodoro.Config{
		WorkMinutes:        cfg.Pomodoro.WorkMinutes,
		ShortBreakMinutes:  cfg.Pomodoro.ShortBreak,
		LongBreakMinutes:   cfg.Pomodoro.LongBreak,
		SessionsBeforeLong: cfg.Pomodoro.SessionsBeforeLong,
		DailyGoal:          cfg.Pomodoro.DailyGoal,
	})

	// Initialize vault writer
	writer := vault.NewWriter(cfg.Vault.Path, cfg, p)

	return &App{
		config:        cfg,
		cache:         c,
		parser:        p,
		writer:        writer,
		watcher:       w,
		currentView:   ViewDashboard,
		focus:         FocusSidebar,
		sidebar:       NewSidebar(),
		pomodoroTimer: timer,
	}
}

// Custom message types
type dataLoadedMsg struct{}
type fileChangedMsg struct {
	path string
}
type tickMsg time.Time

// pomodoroTickMsg is sent every second when the pomodoro timer is running.
type pomodoroTickMsg time.Time

// Init implements tea.Model.
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.loadInitialData(),
		a.startFileWatcher(),
		a.tickPomodoro(),
	)
}

// tickPomodoro returns a command that sends tick messages every second.
func (a *App) tickPomodoro() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return pomodoroTickMsg(t)
	})
}

// loadInitialData loads data from the vault and cache.
func (a *App) loadInitialData() tea.Cmd {
	return func() tea.Msg {
		logging.Info("Loading initial data from vault...")

		// Parse and cache today's daily note
		today := time.Now()
		logging.Debug("Checking for daily note: %s", today.Format("2006-01-02"))

		if a.parser.DailyNoteExists(today) {
			logging.Debug("Daily note exists, parsing...")
			file, err := a.parser.ParseDailyNote(today)
			if err != nil {
				logging.Error("Failed to parse daily note: %v", err)
			} else if file != nil {
				a.cache.SaveFile(file)
				a.todayTasks = file.Tasks
				a.todayNotePath = file.Path
				logging.Info("Loaded %d tasks from daily note: %s", len(a.todayTasks), file.Path)
			}
		} else {
			logging.Debug("No daily note found for today")
		}

		// Load courses from cache/vault
		courseFiles, err := a.cache.GetFilesByType(types.FileTypeCourse)
		if err != nil {
			logging.Error("Failed to get course files from cache: %v", err)
		}
		logging.Debug("Found %d course files in cache", len(courseFiles))

		if len(courseFiles) == 0 {
			// Parse from vault if cache is empty
			logging.Info("Cache empty, parsing vault...")
			files, err := a.parser.ParseVault()
			if err != nil {
				logging.Error("Failed to parse vault: %v", err)
			} else {
				logging.Info("Parsed %d files from vault", len(files))
				for _, f := range files {
					a.cache.SaveFile(f)
				}
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
		logging.Info("Loaded %d active courses", len(a.activeCourses))

		// Load books
		bookFiles, _ := a.cache.GetFilesByType(types.FileTypeBook)
		logging.Debug("Found %d book files", len(bookFiles))
		for _, f := range bookFiles {
			book := a.parseBookFromFile(f)
			if book != nil {
				a.currentBook = book
				logging.Info("Current book: %s", book.Title)
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
		logging.Debug("Loaded %d recent notes", len(a.recentNotes))

		// Load goals
		goals, err := a.parser.ParseGoals()
		if err != nil {
			logging.Error("Failed to parse goals: %v", err)
		} else {
			a.goals = goals
			logging.Info("Loaded %d goals", len(a.goals))
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
			logging.Debug("Created default daily goal: %d pomodoros", a.config.Pomodoro.DailyGoal)
		} else {
			logging.Debug("Daily goal: %d/%d pomodoros", a.dailyGoal.Completed, a.dailyGoal.Target)
		}

		// Load weekly stats
		byDay, totalPom, totalMin, _ := a.cache.GetWeeklyStats()
		a.weeklyStats = views.WeeklyStats{
			Pomodoros: totalPom,
			FocusTime: time.Duration(totalMin) * time.Minute,
			ByDay:     byDay,
			Streak:    0, // TODO: Calculate streak from data
		}
		logging.Debug("Weekly stats: %d pomodoros, %d minutes", totalPom, totalMin)

		logging.Info("Initial data load complete")
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

	case pomodoroTickMsg:
		// The pomodoro timer runs its own goroutine, we just need to refresh the UI
		// Continue ticking to keep the UI updated
		return a, a.tickPomodoro()

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
	key := msg.String()
	logging.Debug("Main panel key: %s, view: %s, module: %d", key, a.currentView, a.focusedModule)

	switch key {
	case "h", "left", "esc":
		a.focus = FocusSidebar
		logging.Debug("Focus switched to sidebar")
		return a, nil

	case "p":
		// Toggle pomodoro timer
		state := a.pomodoroTimer.State()
		switch state {
		case pomodoro.StateIdle:
			a.pomodoroTimer.Start("")
		case pomodoro.StateRunning:
			a.pomodoroTimer.Pause()
		case pomodoro.StatePaused:
			a.pomodoroTimer.Resume()
		case pomodoro.StateBreak:
			a.pomodoroTimer.Stop()
			a.pomodoroTimer.Start("")
		}
		logging.Debug("Pomodoro toggled, state: %d", a.pomodoroTimer.State())
		return a, nil

	case "P":
		// Stop pomodoro timer
		a.pomodoroTimer.Stop()
		logging.Debug("Pomodoro stopped")
		return a, nil

	case "+", "=":
		// Add minute to timer
		a.pomodoroTimer.AdjustTime(1)
		return a, nil

	case "-", "_":
		// Subtract minute from timer
		a.pomodoroTimer.AdjustTime(-1)
		return a, nil
	}

	// Delegate to current view's handler
	switch a.currentView {
	case ViewDashboard:
		return a.handleDashboardKeys(msg)
	default:
		// Other views not implemented yet
		logging.Debug("View %s navigation not implemented", a.currentView)
	}

	return a, nil
}

// handleDashboardKeys handles keyboard input for the dashboard view.
func (a *App) handleDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "j", "down":
		// Move down within current module or to next module
		if a.focusedModule == ModuleTodayFocus && len(a.todayTasks) > 0 {
			// Navigate within tasks
			if a.selectedTask < len(a.todayTasks)-1 {
				a.selectedTask++
				logging.Debug("Task selection: %d", a.selectedTask)
			} else {
				// Move to next module
				a.focusedModule = ModulePomodoro
				logging.Debug("Module changed to: %d", a.focusedModule)
			}
		} else if a.focusedModule < ModuleCount-1 {
			a.focusedModule++
			logging.Debug("Module changed to: %d", a.focusedModule)
		}

	case "k", "up":
		// Move up within current module or to previous module
		if a.focusedModule == ModuleTodayFocus && a.selectedTask > 0 {
			a.selectedTask--
			logging.Debug("Task selection: %d", a.selectedTask)
		} else if a.focusedModule > 0 {
			a.focusedModule--
			// If entering TodayFocus, select last task
			if a.focusedModule == ModuleTodayFocus && len(a.todayTasks) > 0 {
				a.selectedTask = len(a.todayTasks) - 1
			}
			logging.Debug("Module changed to: %d", a.focusedModule)
		}

	case "l", "right", "enter":
		// Enter/interact with current module
		logging.Debug("Enter pressed on module: %d", a.focusedModule)
		switch a.focusedModule {
		case ModuleTodayFocus:
			if len(a.todayTasks) > 0 && a.selectedTask < len(a.todayTasks) {
				// Toggle task status
				return a.toggleTask(a.selectedTask)
			}
		case ModulePomodoro:
			// TODO: Start/pause pomodoro
			logging.Debug("Pomodoro interaction (not implemented)")
		}

	case "x", " ":
		// Toggle task completion (spacebar or x)
		if a.focusedModule == ModuleTodayFocus && len(a.todayTasks) > 0 {
			return a.toggleTask(a.selectedTask)
		}

	case "g":
		// Go to top
		a.focusedModule = ModuleTodayFocus
		a.selectedTask = 0
		logging.Debug("Jump to top")

	case "G":
		// Go to bottom
		a.focusedModule = ModuleRecentNotes
		logging.Debug("Jump to bottom")
	}

	return a, nil
}

// toggleTask toggles the completion status of a task.
func (a *App) toggleTask(index int) (tea.Model, tea.Cmd) {
	if index < 0 || index >= len(a.todayTasks) {
		return a, nil
	}

	task := &a.todayTasks[index]
	logging.Info("Toggling task: %s (current status: %s)", task.Text, task.Status)

	// Save to file
	if a.todayNotePath != "" {
		if err := a.writer.ToggleTask(a.todayNotePath, task); err != nil {
			logging.Error("Failed to toggle task in file: %v", err)
			a.err = err
			return a, nil
		}
		logging.Debug("Task saved to file: %s -> %s", task.Text, task.Status)
	} else {
		// Toggle in memory only (no daily note path)
		if task.Status == "done" {
			task.Status = "open"
		} else {
			task.Status = "done"
		}
		logging.Debug("Task toggled in memory only (no file path)")
	}

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
	// Reserve space for header and footer
	headerHeight := 1
	footerHeight := 1
	contentHeight := a.height - headerHeight - footerHeight

	// Calculate dimensions using percentage-based sidebar width
	sidebarWidth := a.calculateSidebarWidth()
	mainWidth := a.width - sidebarWidth

	// Render components
	sidebar := a.sidebar.Render(sidebarWidth, contentHeight, a.focus == FocusSidebar)
	main := a.renderMainPanel(mainWidth, contentHeight)

	// Build layout using lipgloss for proper alignment
	header := a.renderHeader()
	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
	footer := a.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
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

	// Apply theme styling with background
	bgColor := theme.Current.Color("bg_secondary")
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Current.Color("primary"))
	pathStyle := lipgloss.NewStyle().Foreground(theme.Current.Color("text_muted"))

	// Calculate proper spacing
	titleRendered := titleStyle.Render(title)
	pathRendered := pathStyle.Render(vaultPath)
	titleWidth := lipgloss.Width(titleRendered)
	pathWidth := lipgloss.Width(pathRendered)
	spacerWidth := a.width - titleWidth - pathWidth - 4

	if spacerWidth < 1 {
		spacerWidth = 1
	}

	// Build content without background first
	content := " " + titleRendered + spacer(spacerWidth) + pathRendered + " "

	// Fit to width and apply background to entire header
	content = layout.FitToWidth(content, a.width)
	headerStyle := lipgloss.NewStyle().Background(bgColor)
	return headerStyle.Render(content)
}

func (a *App) renderMainPanel(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	var content string

	// Views render their own frames
	switch a.currentView {
	case ViewDashboard:
		dashboard := views.NewDashboard(width, height)
		dashboard.FocusedModule = int(a.focusedModule)
		dashboard.SelectedTask = a.selectedTask

		// Set real data from vault
		dashboard.Tasks = a.todayTasks
		dashboard.ActiveCourses = a.activeCourses
		dashboard.CurrentBook = a.currentBook
		dashboard.RecentNotes = a.recentNotes
		dashboard.WeeklyStats = a.weeklyStats

		// Set pomodoro state from actual timer
		dailyGoal := a.pomodoroTimer.DailyGoal()
		dailyDone := a.pomodoroTimer.SessionsToday()
		if a.dailyGoal != nil {
			dailyGoal = a.dailyGoal.Target
			dailyDone = max(a.dailyGoal.Completed, a.pomodoroTimer.SessionsToday())
		}
		timerState := a.pomodoroTimer.State()
		stateStr := "ready"
		switch timerState {
		case pomodoro.StateRunning:
			stateStr = "running"
		case pomodoro.StatePaused:
			stateStr = "paused"
		case pomodoro.StateBreak:
			stateStr = "break"
		}
		dashboard.PomodoroState = views.PomodoroState{
			State:     stateStr,
			Remaining: a.pomodoroTimer.Remaining(),
			DailyGoal: dailyGoal,
			DailyDone: dailyDone,
			Context:   a.pomodoroTimer.Context(),
		}

		content = dashboard.Render()

	case ViewCalendar:
		calendar := views.NewCalendar(width, height)
		calendar.SetFocused(a.focus == FocusMain)
		// Convert today's tasks to calendar events
		events := views.ConvertTasksToEvents(a.todayTasks)
		for i := range events {
			events[i].Date = time.Now()
		}
		calendar.SetEvents(events)
		content = calendar.Render()

	case ViewGoals:
		goalsView := views.NewGoalsView(width, height)
		goalsView.SetFocused(a.focus == FocusMain)
		goalsView.SetGoals(a.goals)
		content = goalsView.Render()

	case ViewCourses:
		coursesView := views.NewCoursesView(width, height)
		coursesView.SetFocused(a.focus == FocusMain)
		coursesView.SetCourses(a.activeCourses)
		content = coursesView.Render()

	case ViewBooks:
		booksView := views.NewBooksView(width, height)
		booksView.SetFocused(a.focus == FocusMain)
		var books []types.Book
		if a.currentBook != nil {
			books = append(books, *a.currentBook)
		}
		booksView.SetBooks(books)
		content = booksView.Render()

	case ViewWishlist:
		content = a.renderWishlistPlaceholder(width, height)

	case ViewGraph:
		graphView := views.NewGraphView(width, height)
		graphView.SetFocused(a.focus == FocusMain)
		content = graphView.Render()

	case ViewStats:
		statsView := views.NewStatsView(width, height)
		statsView.SetFocused(a.focus == FocusMain)
		content = statsView.Render()

	case ViewSettings:
		settingsView := views.NewSettingsView(width, height)
		settingsView.SetFocused(a.focus == FocusMain)
		content = settingsView.Render()

	default:
		content = a.renderWishlistPlaceholder(width, height)
	}

	// Ensure main panel has proper background and dimensions
	bgColor := theme.Current.Color("bg_primary")
	mainStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Background(bgColor)

	return mainStyle.Render(content)
}

func (a *App) renderWishlistPlaceholder(width, height int) string {
	frame := layout.NewFrame(width, height)
	frame.SetTitle("Wishlist")
	frame.SetBorder(layout.BorderRounded)
	frame.SetFocused(a.focus == FocusMain)
	if theme.Current != nil {
		frame.SetColors(
			theme.Current.Color("border_default"),
			theme.Current.Color("border_active"),
			theme.Current.Color("text_primary"),
			theme.Current.Color("bg_primary"),
		)
	}
	frame.SetContentLines([]string{
		"",
		"  Coming soon...",
	})
	return frame.Render()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a *App) renderFooter() string {
	// Build help hints with proper styling
	bgColor := theme.Current.Color("bg_secondary")
	keyStyle := theme.S.HelpKey
	descStyle := theme.S.HelpDesc

	hints := []struct {
		key  string
		desc string
	}{
		{"Tab", "focus"},
		{"j/k", "nav"},
		{"Enter", "select"},
		{"p", "pomodoro"},
		{"/", "search"},
		{"?", "help"},
		{"q", "quit"},
	}

	var parts []string
	for _, h := range hints {
		parts = append(parts, keyStyle.Render("["+h.key+"]")+" "+descStyle.Render(h.desc))
	}

	footer := " " + strings.Join(parts, "  ")

	// Fit to width and apply background to entire footer
	footer = layout.FitToWidth(footer, a.width)
	footerStyle := lipgloss.NewStyle().Background(bgColor)
	return footerStyle.Render(footer)
}

// Helper functions

func spacer(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(" ", n)
}

func truncatePath(path string, maxLen int) string {
	pathWidth := lipgloss.Width(path)
	if pathWidth <= maxLen {
		return path
	}
	// Truncate from the beginning
	return "..." + path[len(path)-maxLen+3:]
}

// truncateToWidth truncates a string to fit within maxWidth.
// Delegates to layout.TruncateToWidth for consistent handling.
func truncateToWidth(s string, maxWidth int) string {
	return layout.TruncateToWidth(s, maxWidth)
}

// Run starts the TUI application.
func Run(cfg *config.Config) error {
	// Initialize theme
	themeName := cfg.Theme.Current
	if themeName == "" {
		themeName = "corsair-light"
	}
	logging.Info("Loading theme: %s", themeName)

	t, err := theme.LoadBuiltin(themeName)
	if err != nil {
		logging.Error("Failed to load theme %s: %v, falling back to corsair-light", themeName, err)
		t, _ = theme.LoadBuiltin("corsair-light")
	}
	t.Apply()
	logging.Info("Theme applied: %s (type: %s)", t.Name, t.Type)

	// Initialize i18n
	if cfg.Language != "" {
		if err := i18n.SetLanguage(cfg.Language); err != nil {
			logging.Error("Failed to set language %s: %v", cfg.Language, err)
		}
	}

	// Initialize icons
	iconMode := cfg.Icons.Mode
	if iconMode == "" {
		iconMode = "emoji"
	}
	icons.Init(iconMode)

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
