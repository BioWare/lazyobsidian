// Package types contains shared types used across the application.
package types

import "time"

// TaskStatus represents the status of a task.
type TaskStatus struct {
	Symbol string `yaml:"symbol"`
	Name   string `yaml:"name"`
	Icon   string `yaml:"icon"`
	Color  string `yaml:"color"`
}

// Task represents a task parsed from markdown.
type Task struct {
	ID        int64
	FileID    int64
	Line      int
	Text      string
	Status    string
	ParentID  *int64
	HasNote   bool
	Subtasks  []Task
	Comment   string // inline comment after " // "
	CreatedAt time.Time
	UpdatedAt time.Time
}

// File represents a parsed markdown file.
type File struct {
	ID              int64
	Path            string
	Type            FileType
	Title           string
	Frontmatter     map[string]interface{}
	Tags            []string
	Tasks           []Task
	Links           []Link
	Content         string
	ModifiedAt      time.Time
	ParsedAt        time.Time
}

// FileType represents the type of a file.
type FileType string

const (
	FileTypeNote    FileType = "note"
	FileTypeDaily   FileType = "daily"
	FileTypeGoal    FileType = "goal"
	FileTypeCourse  FileType = "course"
	FileTypeBook    FileType = "book"
	FileTypeUnknown FileType = "unknown"
)

// Link represents a link between files.
type Link struct {
	SourceID int64
	TargetID int64
	Type     LinkType
}

// LinkType represents the type of a link.
type LinkType string

const (
	LinkTypeWikilink LinkType = "wikilink"
	LinkTypeMarkdown LinkType = "markdown"
	LinkTypeEmbed    LinkType = "embed"
)

// Goal represents a goal with hierarchical structure.
type Goal struct {
	ID          int64
	FileID      int64
	Title       string
	Description string
	DueDate     *time.Time
	Progress    float64 // 0.0 - 1.0
	Children    []Goal
	ParentID    *int64
	Pomodoros   int // aggregated from children
	OwnPomodoros int
}

// Course represents a course being tracked.
type Course struct {
	ID           int64
	FileID       int64
	Title        string
	Source       string // udemy, coursera, youtube, etc.
	URL          string
	TotalLessons int
	Completed    int
	Sections     []CourseSection
	TargetDate   *time.Time
	Pomodoros    int
	Notes        int
}

// CourseSection represents a section within a course.
type CourseSection struct {
	Title    string
	Lessons  []CourseLesson
	Progress float64
}

// CourseLesson represents a lesson within a course section.
type CourseLesson struct {
	Title    string
	Status   string
	Duration int // minutes
	HasNote  bool
}

// Book represents a book being read.
type Book struct {
	ID           int64
	FileID       int64
	Title        string
	Author       string
	TotalPages   int
	CurrentPage  int
	Chapters     []BookChapter
	TargetDate   *time.Time
	Pomodoros    int
	Notes        int
}

// BookChapter represents a chapter in a book.
type BookChapter struct {
	Title   string
	Status  string
	HasNote bool
}

// PomodoroSession represents a completed pomodoro session.
type PomodoroSession struct {
	ID        int64
	FileID    *int64
	TaskID    *int64
	StartedAt time.Time
	EndedAt   time.Time
	Duration  int // minutes
	Type      PomodoroType
	Context   string
}

// PomodoroType represents whether it's a work or break session.
type PomodoroType string

const (
	PomodoroTypeWork       PomodoroType = "work"
	PomodoroTypeShortBreak PomodoroType = "short_break"
	PomodoroTypeLongBreak  PomodoroType = "long_break"
)

// DailyGoal represents the daily pomodoro goal tracking.
type DailyGoal struct {
	Date      time.Time
	Target    int
	Completed int
}

// Stats represents aggregated statistics.
type Stats struct {
	TotalFocusTime    time.Duration
	TotalPomodoros    int
	TasksCompleted    int
	CurrentStreak     int
	LongestStreak     int
	ByCategory        map[string]time.Duration
	ByDay             map[string]int // date -> pomodoros
}

// NavItem represents an item in the sidebar navigation.
type NavItem struct {
	ID       string
	Label    string
	Icon     string
	Shortcut string
}
