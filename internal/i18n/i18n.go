// Package i18n provides internationalization support for LazyObsidian.
package i18n

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed locales/*.yaml
var embeddedLocales embed.FS

// Translations holds all translations for a language.
type Translations struct {
	Language string `yaml:"language"`
	Name     string `yaml:"name"`

	Nav      NavTranslations      `yaml:"nav"`
	Actions  ActionsTranslations  `yaml:"actions"`
	Pomodoro PomodoroTranslations `yaml:"pomodoro"`
	Tasks    TasksTranslations    `yaml:"tasks"`
	Time     TimeTranslations     `yaml:"time"`
	Calendar CalendarTranslations `yaml:"calendar"`
	Goals    GoalsTranslations    `yaml:"goals"`
	Courses  CoursesTranslations  `yaml:"courses"`
	Books    BooksTranslations    `yaml:"books"`
	Stats    StatsTranslations    `yaml:"stats"`
	Messages MessagesTranslations `yaml:"messages"`
	Help     HelpTranslations     `yaml:"help"`
}

// NavTranslations holds navigation-related translations.
type NavTranslations struct {
	Dashboard string `yaml:"dashboard"`
	Calendar  string `yaml:"calendar"`
	Goals     string `yaml:"goals"`
	Courses   string `yaml:"courses"`
	Books     string `yaml:"books"`
	Wishlist  string `yaml:"wishlist"`
	Graph     string `yaml:"graph"`
	Stats     string `yaml:"stats"`
	Settings  string `yaml:"settings"`
}

// ActionsTranslations holds action-related translations.
type ActionsTranslations struct {
	Save   string `yaml:"save"`
	Cancel string `yaml:"cancel"`
	Delete string `yaml:"delete"`
	Edit   string `yaml:"edit"`
	Create string `yaml:"create"`
	Open   string `yaml:"open"`
	Close  string `yaml:"close"`
	Start  string `yaml:"start"`
	Stop   string `yaml:"stop"`
	Pause  string `yaml:"pause"`
	Resume string `yaml:"resume"`
}

// PomodoroTranslations holds pomodoro-related translations.
type PomodoroTranslations struct {
	Work          string `yaml:"work"`
	Break         string `yaml:"break"`
	ShortBreak    string `yaml:"short_break"`
	LongBreak     string `yaml:"long_break"`
	Ready         string `yaml:"ready"`
	Running       string `yaml:"running"`
	Paused        string `yaml:"paused"`
	SessionsToday string `yaml:"sessions_today"`
	DailyGoal     string `yaml:"daily_goal"`
	TimeLeft      string `yaml:"time_left"`
}

// TasksTranslations holds task-related translations.
type TasksTranslations struct {
	Open       string `yaml:"open"`
	Done       string `yaml:"done"`
	Cancelled  string `yaml:"cancelled"`
	InProgress string `yaml:"in_progress"`
	Deferred   string `yaml:"deferred"`
	Question   string `yaml:"question"`
	Progress   string `yaml:"progress"`
	NoTasks    string `yaml:"no_tasks"`
}

// TimeTranslations holds time-related translations.
type TimeTranslations struct {
	Minutes   string `yaml:"minutes"`
	Hours     string `yaml:"hours"`
	Days      string `yaml:"days"`
	DaysLeft  string `yaml:"days_left"`
	Today     string `yaml:"today"`
	Yesterday string `yaml:"yesterday"`
	Tomorrow  string `yaml:"tomorrow"`
}

// CalendarLayersTranslations holds calendar layer translations.
type CalendarLayersTranslations struct {
	Tasks    string `yaml:"tasks"`
	Goals    string `yaml:"goals"`
	Journal  string `yaml:"journal"`
	Activity string `yaml:"activity"`
}

// CalendarTranslations holds calendar-related translations.
type CalendarTranslations struct {
	Year   string                     `yaml:"year"`
	Month  string                     `yaml:"month"`
	Day    string                     `yaml:"day"`
	Week   string                     `yaml:"week"`
	Layers CalendarLayersTranslations `yaml:"layers"`
}

// GoalsTranslations holds goals-related translations.
type GoalsTranslations struct {
	Progress string `yaml:"progress"`
	Due      string `yaml:"due"`
	Pace     string `yaml:"pace"`
	NoGoals  string `yaml:"no_goals"`
}

// CoursesTranslations holds courses-related translations.
type CoursesTranslations struct {
	Active    string `yaml:"active"`
	Completed string `yaml:"completed"`
	Lessons   string `yaml:"lessons"`
	Sections  string `yaml:"sections"`
	NoCourses string `yaml:"no_courses"`
}

// BooksTranslations holds books-related translations.
type BooksTranslations struct {
	Reading   string `yaml:"reading"`
	Completed string `yaml:"completed"`
	Pages     string `yaml:"pages"`
	Chapters  string `yaml:"chapters"`
	NoBooks   string `yaml:"no_books"`
}

// StatsTranslations holds stats-related translations.
type StatsTranslations struct {
	TotalFocus      string `yaml:"total_focus"`
	TotalPomodoros  string `yaml:"total_pomodoros"`
	TasksCompleted  string `yaml:"tasks_completed"`
	CurrentStreak   string `yaml:"current_streak"`
	LongestStreak   string `yaml:"longest_streak"`
	ByCategory      string `yaml:"by_category"`
}

// MessagesTranslations holds message translations.
type MessagesTranslations struct {
	NoDailyNote       string `yaml:"no_daily_note"`
	TaskCompleted     string `yaml:"task_completed"`
	PomodoroStarted   string `yaml:"pomodoro_started"`
	PomodoroCompleted string `yaml:"pomodoro_completed"`
	BreakStarted      string `yaml:"break_started"`
	BreakCompleted    string `yaml:"break_completed"`
}

// HelpTranslations holds help-related translations.
type HelpTranslations struct {
	Title       string `yaml:"title"`
	Navigation  string `yaml:"navigation"`
	Actions     string `yaml:"actions"`
	Global      string `yaml:"global"`
	PressAnyKey string `yaml:"press_any_key"`
}

var (
	current *Translations
	mu      sync.RWMutex
)

// Current returns the current translations.
func Current() *Translations {
	mu.RLock()
	defer mu.RUnlock()
	if current == nil {
		// Load default English if not initialized
		current, _ = Load("en")
	}
	return current
}

// SetLanguage sets the current language.
func SetLanguage(lang string) error {
	t, err := Load(lang)
	if err != nil {
		return err
	}
	mu.Lock()
	current = t
	mu.Unlock()
	return nil
}

// Load loads translations for a language.
func Load(lang string) (*Translations, error) {
	// Try embedded locales first
	data, err := embeddedLocales.ReadFile("locales/" + lang + ".yaml")
	if err == nil {
		var t Translations
		if err := yaml.Unmarshal(data, &t); err != nil {
			return nil, fmt.Errorf("failed to parse embedded locale %s: %w", lang, err)
		}
		return &t, nil
	}

	// Try external locales directory
	localesDir, err := getLocalesDir()
	if err == nil {
		path := filepath.Join(localesDir, lang+".yaml")
		data, err := os.ReadFile(path)
		if err == nil {
			var t Translations
			if err := yaml.Unmarshal(data, &t); err != nil {
				return nil, fmt.Errorf("failed to parse locale %s: %w", lang, err)
			}
			return &t, nil
		}
	}

	// Fallback to English
	if lang != "en" {
		return Load("en")
	}

	return defaultEnglish(), nil
}

// LoadFromFile loads translations from a specific file.
func LoadFromFile(path string) (*Translations, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var t Translations
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

// AvailableLanguages returns a list of available language codes.
func AvailableLanguages() []string {
	languages := []string{}
	seen := make(map[string]bool)

	// From embedded
	entries, _ := embeddedLocales.ReadDir("locales")
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			lang := strings.TrimSuffix(entry.Name(), ".yaml")
			if !seen[lang] {
				languages = append(languages, lang)
				seen[lang] = true
			}
		}
	}

	// From external directory
	localesDir, err := getLocalesDir()
	if err == nil {
		entries, _ := os.ReadDir(localesDir)
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
				lang := strings.TrimSuffix(entry.Name(), ".yaml")
				if !seen[lang] {
					languages = append(languages, lang)
					seen[lang] = true
				}
			}
		}
	}

	return languages
}

func getLocalesDir() (string, error) {
	// Try relative to executable
	execPath, err := os.Executable()
	if err == nil {
		localesDir := filepath.Join(filepath.Dir(execPath), "locales")
		if info, err := os.Stat(localesDir); err == nil && info.IsDir() {
			return localesDir, nil
		}
	}

	// Try XDG config directory
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}

	localesDir := filepath.Join(configDir, "lazyobsidian", "locales")
	return localesDir, nil
}

func defaultEnglish() *Translations {
	return &Translations{
		Language: "en",
		Name:     "English",
		Nav: NavTranslations{
			Dashboard: "Dashboard",
			Calendar:  "Calendar",
			Goals:     "Goals",
			Courses:   "Courses",
			Books:     "Books",
			Wishlist:  "Wishlist",
			Graph:     "Graph",
			Stats:     "Statistics",
			Settings:  "Settings",
		},
		Actions: ActionsTranslations{
			Save:   "Save",
			Cancel: "Cancel",
			Delete: "Delete",
			Edit:   "Edit",
			Create: "Create",
			Open:   "Open",
			Close:  "Close",
			Start:  "Start",
			Stop:   "Stop",
			Pause:  "Pause",
			Resume: "Resume",
		},
		Pomodoro: PomodoroTranslations{
			Work:          "Work",
			Break:         "Break",
			ShortBreak:    "Short Break",
			LongBreak:     "Long Break",
			Ready:         "Ready",
			Running:       "Running",
			Paused:        "Paused",
			SessionsToday: "Sessions today: {count}",
			DailyGoal:     "Daily goal: {current}/{target}",
			TimeLeft:      "{minutes}:{seconds}",
		},
		Tasks: TasksTranslations{
			Open:       "Open",
			Done:       "Done",
			Cancelled:  "Cancelled",
			InProgress: "In Progress",
			Deferred:   "Deferred",
			Question:   "Question",
			Progress:   "Progress: {completed}/{total}",
			NoTasks:    "No tasks",
		},
		Time: TimeTranslations{
			Minutes:   "{n} min",
			Hours:     "{n}h",
			Days:      "{n} days",
			DaysLeft:  "{n} days left",
			Today:     "Today",
			Yesterday: "Yesterday",
			Tomorrow:  "Tomorrow",
		},
		Calendar: CalendarTranslations{
			Year:  "Year",
			Month: "Month",
			Day:   "Day",
			Week:  "Week",
			Layers: CalendarLayersTranslations{
				Tasks:    "Tasks",
				Goals:    "Goals",
				Journal:  "Journal",
				Activity: "Activity",
			},
		},
		Goals: GoalsTranslations{
			Progress: "Progress: {percent}%",
			Due:      "Due: {date}",
			Pace:     "Pace: {hours}/day",
			NoGoals:  "No goals",
		},
		Courses: CoursesTranslations{
			Active:    "Active",
			Completed: "Completed",
			Lessons:   "{completed}/{total} lessons",
			Sections:  "{count} sections",
			NoCourses: "No courses",
		},
		Books: BooksTranslations{
			Reading:   "Currently Reading",
			Completed: "Completed",
			Pages:     "p.{current}/{total}",
			Chapters:  "{count} chapters",
			NoBooks:   "No books",
		},
		Stats: StatsTranslations{
			TotalFocus:     "Total Focus Time",
			TotalPomodoros: "Total Pomodoros",
			TasksCompleted: "Tasks Completed",
			CurrentStreak:  "Current Streak",
			LongestStreak:  "Longest Streak",
			ByCategory:     "By Category",
		},
		Messages: MessagesTranslations{
			NoDailyNote:       "No daily note. Press Enter to create.",
			TaskCompleted:     "Task completed!",
			PomodoroStarted:   "Pomodoro started!",
			PomodoroCompleted: "Pomodoro completed!",
			BreakStarted:      "Break started!",
			BreakCompleted:    "Break completed!",
		},
		Help: HelpTranslations{
			Title:       "Help",
			Navigation:  "Navigation",
			Actions:     "Actions",
			Global:      "Global",
			PressAnyKey: "Press any key to close",
		},
	}
}

// T is a shortcut for Current() to make translations easier to use.
func T() *Translations {
	return Current()
}

// Format replaces placeholders in a string with values.
// Example: Format("Hello {name}!", map[string]interface{}{"name": "World"}) -> "Hello World!"
func Format(template string, values map[string]interface{}) string {
	result := template
	for key, value := range values {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprint(value))
	}
	return result
}
