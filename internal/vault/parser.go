// Package vault handles parsing and watching the Obsidian vault.
package vault

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BioWare/lazyobsidian/internal/config"
	"github.com/BioWare/lazyobsidian/pkg/types"
)

var (
	// Regex patterns for parsing markdown
	taskPattern       = regexp.MustCompile(`^(\s*)-\s*\[([ x\-/>?])\]\s*(.+)$`)
	wikilinkPattern   = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	markdownLinkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	embedPattern      = regexp.MustCompile(`!\[\[([^\]]+)\]\]`)
	frontmatterStart  = regexp.MustCompile(`^---\s*$`)
	tagPattern        = regexp.MustCompile(`#([a-zA-Z0-9_/-]+)`)
	headingPattern    = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	pomodoroPattern   = regexp.MustCompile(`🍅`)
	schedulePattern   = regexp.MustCompile(`^-\s*(\d{1,2}:\d{2})(?:-(\d{1,2}:\d{2}))?\s*\|\s*(.+?)(?:\s*\|\s*(.+))?$`)
)

// Parser handles parsing markdown files from the vault.
type Parser struct {
	vaultPath string
	config    *config.Config
}

// NewParser creates a new vault parser.
func NewParser(vaultPath string, cfg *config.Config) *Parser {
	return &Parser{
		vaultPath: vaultPath,
		config:    cfg,
	}
}

// ParseFile parses a single markdown file.
func (p *Parser) ParseFile(path string) (*types.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	result := &types.File{
		Path:       path,
		Title:      strings.TrimSuffix(filepath.Base(path), ".md"),
		ModifiedAt: info.ModTime(),
		ParsedAt:   time.Now(),
		Tags:       []string{},
		Tasks:      []types.Task{},
		Links:      []types.Link{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inFrontmatter := false
	frontmatterLineCount := 0
	frontmatterLines := []string{}
	var contentBuilder strings.Builder

	// For task tree building
	var taskStack []*types.Task
	var allTasks []types.Task

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle frontmatter
		if lineNum == 1 && frontmatterStart.MatchString(line) {
			inFrontmatter = true
			frontmatterLineCount = 1
			continue
		}

		if inFrontmatter {
			frontmatterLineCount++
			if frontmatterStart.MatchString(line) && frontmatterLineCount > 1 {
				inFrontmatter = false
				result.Frontmatter = parseFrontmatter(frontmatterLines)
				continue
			}
			frontmatterLines = append(frontmatterLines, line)
			continue
		}

		// Parse tasks with indentation for subtasks
		if matches := taskPattern.FindStringSubmatch(line); matches != nil {
			indent := len(matches[1])
			indentLevel := indent / 2 // Assuming 2 spaces per indent level

			// Convert symbol to named status
			statusSymbol := matches[2]
			statusName := p.symbolToStatus(statusSymbol)

			task := types.Task{
				Line:     lineNum,
				Status:   statusName,
				Text:     matches[3],
				Subtasks: []types.Task{},
			}

			// Check for inline comment
			if idx := strings.Index(task.Text, " // "); idx != -1 {
				task.Comment = strings.TrimSpace(task.Text[idx+4:])
				task.Text = strings.TrimSpace(task.Text[:idx])
			}

			// Check for task note indicator
			if strings.Contains(task.Text, "📎") {
				task.HasNote = true
				task.Text = strings.ReplaceAll(task.Text, "📎", "")
				task.Text = strings.TrimSpace(task.Text)
			}

			// Count pomodoros in task text
			pomodoroCount := len(pomodoroPattern.FindAllString(task.Text, -1))
			if pomodoroCount > 0 {
				// Remove pomodoros from text for cleaner display
				task.Text = strings.TrimSpace(pomodoroPattern.ReplaceAllString(task.Text, ""))
			}

			// Build task tree
			if indentLevel == 0 {
				// Root level task
				allTasks = append(allTasks, task)
				taskStack = []*types.Task{&allTasks[len(allTasks)-1]}
			} else {
				// Subtask - find parent
				for len(taskStack) > indentLevel {
					taskStack = taskStack[:len(taskStack)-1]
				}
				if len(taskStack) > 0 {
					parent := taskStack[len(taskStack)-1]
					parent.Subtasks = append(parent.Subtasks, task)
					taskStack = append(taskStack, &parent.Subtasks[len(parent.Subtasks)-1])
				} else {
					// No valid parent, add as root
					allTasks = append(allTasks, task)
					taskStack = []*types.Task{&allTasks[len(allTasks)-1]}
				}
			}
		}

		// Extract wikilinks
		for range wikilinkPattern.FindAllStringSubmatch(line, -1) {
			result.Links = append(result.Links, types.Link{
				Type: types.LinkTypeWikilink,
				// TargetID will be resolved later when we have file IDs
			})
		}

		// Extract markdown links
		for _, match := range markdownLinkPattern.FindAllStringSubmatch(line, -1) {
			if strings.HasPrefix(match[2], "http") {
				continue // Skip external links
			}
			result.Links = append(result.Links, types.Link{
				Type: types.LinkTypeMarkdown,
			})
		}

		// Extract embeds
		for range embedPattern.FindAllStringSubmatch(line, -1) {
			result.Links = append(result.Links, types.Link{
				Type: types.LinkTypeEmbed,
			})
		}

		// Extract tags
		for _, match := range tagPattern.FindAllStringSubmatch(line, -1) {
			tag := match[1]
			if !containsString(result.Tags, tag) {
				result.Tags = append(result.Tags, tag)
			}
		}

		contentBuilder.WriteString(line)
		contentBuilder.WriteString("\n")
	}

	result.Tasks = allTasks
	result.Content = contentBuilder.String()

	// Determine file type based on path and frontmatter
	result.Type = p.determineFileType(path, result.Frontmatter)

	// Extract title from frontmatter or first heading
	if result.Frontmatter != nil {
		if title, ok := result.Frontmatter["title"].(string); ok && title != "" {
			result.Title = title
		}
	}

	return result, scanner.Err()
}

// ParseVault parses all markdown files in the vault.
func (p *Parser) ParseVault() ([]*types.File, error) {
	var files []*types.File

	err := filepath.Walk(p.vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-markdown files
		if info.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		file, err := p.ParseFile(path)
		if err != nil {
			// Log error but continue parsing other files
			return nil
		}

		files = append(files, file)
		return nil
	})

	return files, err
}

// ParseDailyNote parses a daily note file for the given date.
func (p *Parser) ParseDailyNote(date time.Time) (*types.File, error) {
	folder := p.config.Daily.Folder
	format := p.config.Daily.FilenameFormat
	if format == "" {
		format = "2006-01-02"
	}

	filename := date.Format(format) + ".md"
	path := filepath.Join(p.vaultPath, folder, filename)

	return p.ParseFile(path)
}

// DailyNoteExists checks if a daily note exists for the given date.
func (p *Parser) DailyNoteExists(date time.Time) bool {
	folder := p.config.Daily.Folder
	format := p.config.Daily.FilenameFormat
	if format == "" {
		format = "2006-01-02"
	}

	filename := date.Format(format) + ".md"
	path := filepath.Join(p.vaultPath, folder, filename)

	_, err := os.Stat(path)
	return err == nil
}

func (p *Parser) determineFileType(path string, frontmatter map[string]interface{}) types.FileType {
	// Check frontmatter type field first
	if frontmatter != nil {
		if t, ok := frontmatter["type"].(string); ok {
			switch t {
			case "daily":
				return types.FileTypeDaily
			case "goal", "yearly_plan", "monthly_plan", "weekly_plan":
				return types.FileTypeGoal
			case "course":
				return types.FileTypeCourse
			case "book":
				return types.FileTypeBook
			}
		}
	}

	// Check path against configured folders
	if p.config != nil {
		relPath, err := filepath.Rel(p.vaultPath, path)
		if err == nil {
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) > 0 {
				folder := parts[0]
				switch folder {
				case p.config.Folders.Daily, p.config.Daily.Folder:
					return types.FileTypeDaily
				case p.config.Folders.Goals:
					return types.FileTypeGoal
				case p.config.Folders.Courses:
					return types.FileTypeCourse
				case p.config.Folders.Books:
					return types.FileTypeBook
				}
			}
		}
	}

	return types.FileTypeNote
}

func parseFrontmatter(lines []string) map[string]interface{} {
	result := make(map[string]interface{})
	var currentKey string
	var inList bool
	var listItems []string

	for _, line := range lines {
		// Check for list item
		if strings.HasPrefix(strings.TrimSpace(line), "- ") && currentKey != "" {
			inList = true
			item := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- "))
			listItems = append(listItems, item)
			continue
		}

		// If we were in a list, save it
		if inList && !strings.HasPrefix(strings.TrimSpace(line), "- ") {
			result[currentKey] = listItems
			listItems = nil
			inList = false
		}

		// Parse key: value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			currentKey = strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			if value == "" {
				// Could be start of list or nested object
				continue
			}

			// Handle quoted strings
			if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) ||
				(strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
				value = value[1 : len(value)-1]
			}

			// Handle arrays like [item1, item2]
			if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
				inner := value[1 : len(value)-1]
				items := strings.Split(inner, ",")
				var cleanItems []string
				for _, item := range items {
					cleanItems = append(cleanItems, strings.TrimSpace(item))
				}
				result[currentKey] = cleanItems
				continue
			}

			// Handle booleans
			if value == "true" {
				result[currentKey] = true
				continue
			}
			if value == "false" {
				result[currentKey] = false
				continue
			}

			result[currentKey] = value
		}
	}

	// Handle trailing list
	if inList && currentKey != "" {
		result[currentKey] = listItems
	}

	return result
}

func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// symbolToStatus converts a task status symbol to its named status.
func (p *Parser) symbolToStatus(symbol string) string {
	// Check config for custom statuses first
	if p.config != nil {
		for _, status := range p.config.Tasks.Statuses {
			if status.Symbol == symbol {
				return status.Name
			}
		}
	}

	// Fallback to default mapping
	switch symbol {
	case " ":
		return "open"
	case "x":
		return "done"
	case "-":
		return "cancelled"
	case "/":
		return "in_progress"
	case ">":
		return "deferred"
	case "?":
		return "question"
	default:
		return "open"
	}
}

// statusToSymbol converts a named status to its symbol.
func (p *Parser) statusToSymbol(status string) string {
	// Check config for custom statuses first
	if p.config != nil {
		for _, s := range p.config.Tasks.Statuses {
			if s.Name == status {
				return s.Symbol
			}
		}
	}

	// Fallback to default mapping
	switch status {
	case "open":
		return " "
	case "done":
		return "x"
	case "cancelled":
		return "-"
	case "in_progress":
		return "/"
	case "deferred":
		return ">"
	case "question":
		return "?"
	default:
		return " "
	}
}

// CountTasks counts completed and total tasks recursively.
func CountTasks(tasks []types.Task) (completed, total int) {
	for _, task := range tasks {
		total++
		if task.Status == "done" {
			completed++
		}
		c, t := CountTasks(task.Subtasks)
		completed += c
		total += t
	}
	return completed, total
}

// FlattenTasks flattens a task tree into a slice.
func FlattenTasks(tasks []types.Task) []types.Task {
	var result []types.Task
	for _, task := range tasks {
		result = append(result, task)
		result = append(result, FlattenTasks(task.Subtasks)...)
	}
	return result
}

// ParseGoals parses goal files from the Goals folder.
func (p *Parser) ParseGoals() ([]types.Goal, error) {
	goalsFolder := p.config.Folders.Goals
	if goalsFolder == "" {
		goalsFolder = "Goals"
	}

	goalsPath := filepath.Join(p.vaultPath, goalsFolder)

	var goals []types.Goal

	err := filepath.Walk(goalsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		file, err := p.ParseFile(path)
		if err != nil {
			return nil
		}

		goal := types.Goal{
			Title:       file.Title,
			Description: "",
			Progress:    0,
			Children:    []types.Goal{},
		}

		// Extract description from frontmatter
		if file.Frontmatter != nil {
			if desc, ok := file.Frontmatter["description"].(string); ok {
				goal.Description = desc
			}
			if progress, ok := file.Frontmatter["progress"].(float64); ok {
				goal.Progress = progress
			}
			if due, ok := file.Frontmatter["due"].(string); ok {
				if t, err := time.Parse("2006-01-02", due); err == nil {
					goal.DueDate = &t
				}
			}
		}

		// Calculate progress from tasks if not set
		if goal.Progress == 0 && len(file.Tasks) > 0 {
			completed, total := CountTasks(file.Tasks)
			if total > 0 {
				goal.Progress = float64(completed) / float64(total)
			}
		}

		goals = append(goals, goal)
		return nil
	})

	return goals, err
}

// ParseCourses parses course files from the Courses folder.
func (p *Parser) ParseCourses() ([]types.Course, error) {
	coursesFolder := p.config.Folders.Courses
	if coursesFolder == "" {
		coursesFolder = "Courses"
	}

	coursesPath := filepath.Join(p.vaultPath, coursesFolder)

	var courses []types.Course

	err := filepath.Walk(coursesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		file, err := p.ParseFile(path)
		if err != nil {
			return nil
		}

		course := types.Course{
			Title:    file.Title,
			Sections: []types.CourseSection{},
		}

		// Extract from frontmatter
		if file.Frontmatter != nil {
			if source, ok := file.Frontmatter["source"].(string); ok {
				course.Source = source
			}
			if url, ok := file.Frontmatter["url"].(string); ok {
				course.URL = url
			}
			if total, ok := file.Frontmatter["total_lessons"].(int); ok {
				course.TotalLessons = total
			}
		}

		// Calculate progress from tasks
		if len(file.Tasks) > 0 {
			completed, total := CountTasks(file.Tasks)
			course.Completed = completed
			if course.TotalLessons == 0 {
				course.TotalLessons = total
			}
		}

		courses = append(courses, course)
		return nil
	})

	return courses, err
}

// ParseBooks parses book files from the Books folder.
func (p *Parser) ParseBooks() ([]types.Book, error) {
	booksFolder := p.config.Folders.Books
	if booksFolder == "" {
		booksFolder = "Books"
	}

	booksPath := filepath.Join(p.vaultPath, booksFolder)

	var books []types.Book

	err := filepath.Walk(booksPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		file, err := p.ParseFile(path)
		if err != nil {
			return nil
		}

		book := types.Book{
			Title:    file.Title,
			Chapters: []types.BookChapter{},
		}

		// Extract from frontmatter
		if file.Frontmatter != nil {
			if author, ok := file.Frontmatter["author"].(string); ok {
				book.Author = author
			}
			if currentPage, ok := file.Frontmatter["current_page"].(int); ok {
				book.CurrentPage = currentPage
			}
			if totalPages, ok := file.Frontmatter["total_pages"].(int); ok {
				book.TotalPages = totalPages
			}
		}

		books = append(books, book)
		return nil
	})

	return books, err
}
