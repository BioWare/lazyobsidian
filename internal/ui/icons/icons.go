// Package icons provides icon support for different terminal configurations.
package icons

import (
	"sync"
)

// Mode represents the icon display mode.
type Mode string

const (
	// ModeEmoji uses emoji icons (requires emoji support).
	ModeEmoji Mode = "emoji"
	// ModeNerd uses Nerd Font icons (requires Nerd Font).
	ModeNerd Mode = "nerd"
	// ModeNerdMinimal uses minimal Nerd Font icons.
	ModeNerdMinimal Mode = "nerd_minimal"
	// ModeASCII uses ASCII-only icons (maximum compatibility).
	ModeASCII Mode = "ascii"
)

// IconSet contains all icons for a specific mode.
type IconSet struct {
	// Navigation
	Dashboard string
	Calendar  string
	Goals     string
	Courses   string
	Books     string
	Wishlist  string
	Graph     string
	Stats     string
	Settings  string

	// Task statuses
	TaskOpen       string
	TaskDone       string
	TaskCancelled  string
	TaskInProgress string
	TaskDeferred   string
	TaskQuestion   string

	// Pomodoro
	PomodoroWork  string
	PomodoroBreak string
	PomodoroReady string

	// UI elements
	Selected      string
	Unselected    string
	Expanded      string
	Collapsed     string
	Branch        string
	LastBranch    string
	Indent        string
	Bullet        string
	Arrow         string
	ArrowRight    string
	ArrowLeft     string
	ArrowUp       string
	ArrowDown     string
	Check         string
	Cross         string
	Warning       string
	Error         string
	Info          string
	Success       string

	// Progress
	ProgressFull    string
	ProgressPartial string
	ProgressEmpty   string

	// Calendar
	CalendarToday string
	CalendarEvent string

	// Books & Courses
	Book     string
	Chapter  string
	Course   string
	Section  string
	Lesson   string

	// Misc
	Star      string
	StarEmpty string
	Heart     string
	Fire      string
	Clock     string
	Timer     string
	Note      string
	Folder    string
	File      string
	Link      string
	Tag       string
	Search    string
	Filter    string
	Sort      string
	Edit      string
	Delete    string
	Add       string
	Remove    string
	Refresh   string
	Sync      string
}

var (
	currentMode Mode
	currentSet  *IconSet
	customIcons map[string]string
	mu          sync.RWMutex

	// Pre-defined icon sets
	emojiSet       = newEmojiSet()
	nerdSet        = newNerdSet()
	nerdMinimalSet = newNerdMinimalSet()
	asciiSet       = newASCIISet()
)

// SetMode sets the current icon mode.
func SetMode(mode Mode) {
	mu.Lock()
	defer mu.Unlock()
	currentMode = mode
	switch mode {
	case ModeEmoji:
		currentSet = emojiSet
	case ModeNerd:
		currentSet = nerdSet
	case ModeNerdMinimal:
		currentSet = nerdMinimalSet
	default:
		currentSet = asciiSet
	}
}

// GetMode returns the current icon mode.
func GetMode() Mode {
	mu.RLock()
	defer mu.RUnlock()
	return currentMode
}

// SetCustom sets a custom icon override.
func SetCustom(name, icon string) {
	mu.Lock()
	defer mu.Unlock()
	if customIcons == nil {
		customIcons = make(map[string]string)
	}
	customIcons[name] = icon
}

// ClearCustom clears all custom icon overrides.
func ClearCustom() {
	mu.Lock()
	defer mu.Unlock()
	customIcons = nil
}

// Current returns the current icon set.
func Current() *IconSet {
	mu.RLock()
	defer mu.RUnlock()
	if currentSet == nil {
		return nerdMinimalSet
	}
	return currentSet
}

// Get returns an icon by name, checking custom overrides first.
func Get(name string) string {
	mu.RLock()
	defer mu.RUnlock()
	if customIcons != nil {
		if icon, ok := customIcons[name]; ok {
			return icon
		}
	}
	return getFromSet(name, currentSet)
}

func getFromSet(name string, set *IconSet) string {
	if set == nil {
		set = nerdMinimalSet
	}
	switch name {
	// Navigation
	case "dashboard":
		return set.Dashboard
	case "calendar":
		return set.Calendar
	case "goals":
		return set.Goals
	case "courses":
		return set.Courses
	case "books":
		return set.Books
	case "wishlist":
		return set.Wishlist
	case "graph":
		return set.Graph
	case "stats":
		return set.Stats
	case "settings":
		return set.Settings

	// Tasks
	case "task_open":
		return set.TaskOpen
	case "task_done":
		return set.TaskDone
	case "task_cancelled":
		return set.TaskCancelled
	case "task_in_progress":
		return set.TaskInProgress
	case "task_deferred":
		return set.TaskDeferred
	case "task_question":
		return set.TaskQuestion

	// Pomodoro
	case "pomodoro_work":
		return set.PomodoroWork
	case "pomodoro_break":
		return set.PomodoroBreak
	case "pomodoro_ready":
		return set.PomodoroReady

	// UI
	case "selected":
		return set.Selected
	case "unselected":
		return set.Unselected
	case "expanded":
		return set.Expanded
	case "collapsed":
		return set.Collapsed
	case "branch":
		return set.Branch
	case "last_branch":
		return set.LastBranch
	case "indent":
		return set.Indent
	case "bullet":
		return set.Bullet
	case "arrow":
		return set.Arrow
	case "arrow_right":
		return set.ArrowRight
	case "arrow_left":
		return set.ArrowLeft
	case "arrow_up":
		return set.ArrowUp
	case "arrow_down":
		return set.ArrowDown
	case "check":
		return set.Check
	case "cross":
		return set.Cross
	case "warning":
		return set.Warning
	case "error":
		return set.Error
	case "info":
		return set.Info
	case "success":
		return set.Success

	// Progress
	case "progress_full":
		return set.ProgressFull
	case "progress_partial":
		return set.ProgressPartial
	case "progress_empty":
		return set.ProgressEmpty

	// Calendar
	case "calendar_today":
		return set.CalendarToday
	case "calendar_event":
		return set.CalendarEvent

	// Books & Courses
	case "book":
		return set.Book
	case "chapter":
		return set.Chapter
	case "course":
		return set.Course
	case "section":
		return set.Section
	case "lesson":
		return set.Lesson

	// Misc
	case "star":
		return set.Star
	case "star_empty":
		return set.StarEmpty
	case "heart":
		return set.Heart
	case "fire":
		return set.Fire
	case "clock":
		return set.Clock
	case "timer":
		return set.Timer
	case "note":
		return set.Note
	case "folder":
		return set.Folder
	case "file":
		return set.File
	case "link":
		return set.Link
	case "tag":
		return set.Tag
	case "search":
		return set.Search
	case "filter":
		return set.Filter
	case "sort":
		return set.Sort
	case "edit":
		return set.Edit
	case "delete":
		return set.Delete
	case "add":
		return set.Add
	case "remove":
		return set.Remove
	case "refresh":
		return set.Refresh
	case "sync":
		return set.Sync

	default:
		return "?"
	}
}

func newEmojiSet() *IconSet {
	return &IconSet{
		// Navigation
		Dashboard: "📊",
		Calendar:  "📅",
		Goals:     "🎯",
		Courses:   "📚",
		Books:     "📖",
		Wishlist:  "💝",
		Graph:     "🕸️",
		Stats:     "📈",
		Settings:  "⚙️",

		// Task statuses
		TaskOpen:       "○",
		TaskDone:       "✅",
		TaskCancelled:  "❌",
		TaskInProgress: "🔄",
		TaskDeferred:   "⏸️",
		TaskQuestion:   "❓",

		// Pomodoro
		PomodoroWork:  "🍅",
		PomodoroBreak: "☕",
		PomodoroReady: "⏱️",

		// UI elements
		Selected:      "▶",
		Unselected:    " ",
		Expanded:      "▼",
		Collapsed:     "▶",
		Branch:        "├",
		LastBranch:    "└",
		Indent:        "│",
		Bullet:        "•",
		Arrow:         "→",
		ArrowRight:    "→",
		ArrowLeft:     "←",
		ArrowUp:       "↑",
		ArrowDown:     "↓",
		Check:         "✓",
		Cross:         "✗",
		Warning:       "⚠️",
		Error:         "🔴",
		Info:          "ℹ️",
		Success:       "🟢",

		// Progress
		ProgressFull:    "█",
		ProgressPartial: "▓",
		ProgressEmpty:   "░",

		// Calendar
		CalendarToday: "📍",
		CalendarEvent: "📌",

		// Books & Courses
		Book:    "📕",
		Chapter: "📑",
		Course:  "🎓",
		Section: "📂",
		Lesson:  "📝",

		// Misc
		Star:      "⭐",
		StarEmpty: "☆",
		Heart:     "❤️",
		Fire:      "🔥",
		Clock:     "🕐",
		Timer:     "⏱️",
		Note:      "📝",
		Folder:    "📁",
		File:      "📄",
		Link:      "🔗",
		Tag:       "🏷️",
		Search:    "🔍",
		Filter:    "🔬",
		Sort:      "↕️",
		Edit:      "✏️",
		Delete:    "🗑️",
		Add:       "➕",
		Remove:    "➖",
		Refresh:   "🔄",
		Sync:      "🔁",
	}
}

func newNerdSet() *IconSet {
	return &IconSet{
		// Navigation (using Nerd Font icons)
		Dashboard: "",
		Calendar:  "",
		Goals:     "",
		Courses:   "",
		Books:     "",
		Wishlist:  "",
		Graph:     "",
		Stats:     "",
		Settings:  "",

		// Task statuses
		TaskOpen:       "",
		TaskDone:       "",
		TaskCancelled:  "",
		TaskInProgress: "",
		TaskDeferred:   "",
		TaskQuestion:   "",

		// Pomodoro
		PomodoroWork:  "",
		PomodoroBreak: "",
		PomodoroReady: "",

		// UI elements
		Selected:      "",
		Unselected:    " ",
		Expanded:      "",
		Collapsed:     "",
		Branch:        "├",
		LastBranch:    "└",
		Indent:        "│",
		Bullet:        "",
		Arrow:         "",
		ArrowRight:    "",
		ArrowLeft:     "",
		ArrowUp:       "",
		ArrowDown:     "",
		Check:         "",
		Cross:         "",
		Warning:       "",
		Error:         "",
		Info:          "",
		Success:       "",

		// Progress
		ProgressFull:    "█",
		ProgressPartial: "▓",
		ProgressEmpty:   "░",

		// Calendar
		CalendarToday: "",
		CalendarEvent: "",

		// Books & Courses
		Book:    "",
		Chapter: "",
		Course:  "",
		Section: "",
		Lesson:  "",

		// Misc
		Star:      "",
		StarEmpty: "",
		Heart:     "",
		Fire:      "",
		Clock:     "",
		Timer:     "",
		Note:      "",
		Folder:    "",
		File:      "",
		Link:      "",
		Tag:       "",
		Search:    "",
		Filter:    "",
		Sort:      "",
		Edit:      "",
		Delete:    "",
		Add:       "",
		Remove:    "",
		Refresh:   "",
		Sync:      "",
	}
}

func newNerdMinimalSet() *IconSet {
	return &IconSet{
		// Navigation (simple icons)
		Dashboard: " ",
		Calendar:  " ",
		Goals:     " ",
		Courses:   " ",
		Books:     " ",
		Wishlist:  " ",
		Graph:     " ",
		Stats:     " ",
		Settings:  " ",

		// Task statuses
		TaskOpen:       "○",
		TaskDone:       "✓",
		TaskCancelled:  "⊘",
		TaskInProgress: "◐",
		TaskDeferred:   "→",
		TaskQuestion:   "?",

		// Pomodoro
		PomodoroWork:  "🍅",
		PomodoroBreak: "☕",
		PomodoroReady: "⏱",

		// UI elements
		Selected:      "▶",
		Unselected:    " ",
		Expanded:      "▼",
		Collapsed:     "▶",
		Branch:        "├",
		LastBranch:    "└",
		Indent:        "│",
		Bullet:        "•",
		Arrow:         "→",
		ArrowRight:    "→",
		ArrowLeft:     "←",
		ArrowUp:       "↑",
		ArrowDown:     "↓",
		Check:         "✓",
		Cross:         "✗",
		Warning:       "⚠",
		Error:         "✗",
		Info:          "i",
		Success:       "✓",

		// Progress
		ProgressFull:    "█",
		ProgressPartial: "▓",
		ProgressEmpty:   "░",

		// Calendar
		CalendarToday: "•",
		CalendarEvent: "◆",

		// Books & Courses
		Book:    "📖",
		Chapter: "§",
		Course:  "📚",
		Section: "▸",
		Lesson:  "•",

		// Misc
		Star:      "★",
		StarEmpty: "☆",
		Heart:     "♥",
		Fire:      "🔥",
		Clock:     "⏰",
		Timer:     "⏱",
		Note:      "📝",
		Folder:    "📁",
		File:      "📄",
		Link:      "🔗",
		Tag:       "#",
		Search:    "🔍",
		Filter:    "⊛",
		Sort:      "↕",
		Edit:      "✎",
		Delete:    "✗",
		Add:       "+",
		Remove:    "-",
		Refresh:   "↻",
		Sync:      "⇄",
	}
}

func newASCIISet() *IconSet {
	return &IconSet{
		// Navigation (ASCII only)
		Dashboard: "[D]",
		Calendar:  "[C]",
		Goals:     "[G]",
		Courses:   "[U]",
		Books:     "[B]",
		Wishlist:  "[W]",
		Graph:     "[R]",
		Stats:     "[S]",
		Settings:  "[*]",

		// Task statuses
		TaskOpen:       "[ ]",
		TaskDone:       "[x]",
		TaskCancelled:  "[-]",
		TaskInProgress: "[/]",
		TaskDeferred:   "[>]",
		TaskQuestion:   "[?]",

		// Pomodoro
		PomodoroWork:  "[W]",
		PomodoroBreak: "[B]",
		PomodoroReady: "[R]",

		// UI elements
		Selected:      ">",
		Unselected:    " ",
		Expanded:      "v",
		Collapsed:     ">",
		Branch:        "|--",
		LastBranch:    "`--",
		Indent:        "|  ",
		Bullet:        "*",
		Arrow:         "->",
		ArrowRight:    "->",
		ArrowLeft:     "<-",
		ArrowUp:       "^",
		ArrowDown:     "v",
		Check:         "[x]",
		Cross:         "[X]",
		Warning:       "[!]",
		Error:         "[X]",
		Info:          "[i]",
		Success:       "[+]",

		// Progress
		ProgressFull:    "#",
		ProgressPartial: "=",
		ProgressEmpty:   "-",

		// Calendar
		CalendarToday: "*",
		CalendarEvent: "+",

		// Books & Courses
		Book:    "[B]",
		Chapter: " -",
		Course:  "[C]",
		Section: " >",
		Lesson:  " *",

		// Misc
		Star:      "*",
		StarEmpty: "o",
		Heart:     "<3",
		Fire:      "~*~",
		Clock:     "()",
		Timer:     "[]",
		Note:      "[N]",
		Folder:    "[/]",
		File:      "[.]",
		Link:      "@",
		Tag:       "#",
		Search:    "[?]",
		Filter:    "[F]",
		Sort:      "[S]",
		Edit:      "[E]",
		Delete:    "[D]",
		Add:       "[+]",
		Remove:    "[-]",
		Refresh:   "[R]",
		Sync:      "[~]",
	}
}

// I is a shortcut for Get to make icon access easier.
func I(name string) string {
	return Get(name)
}

// Init initializes the icons package with the specified mode.
func Init(mode string) {
	switch Mode(mode) {
	case ModeEmoji:
		SetMode(ModeEmoji)
	case ModeNerd:
		SetMode(ModeNerd)
	case ModeNerdMinimal:
		SetMode(ModeNerdMinimal)
	case ModeASCII:
		SetMode(ModeASCII)
	default:
		SetMode(ModeNerdMinimal)
	}
}
