// Package cache provides SQLite-based caching for parsed vault data.
package cache

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/BioWare/lazyobsidian/pkg/types"
)

// Cache provides fast access to parsed vault data.
type Cache struct {
	db *sql.DB
}

// New creates a new cache instance.
func New(vaultPath string) (*Cache, error) {
	cacheDir := filepath.Join(vaultPath, ".lazyobsidian")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(cacheDir, "cache.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	cache := &Cache{db: db}
	if err := cache.initSchema(); err != nil {
		db.Close()
		return nil, err
	}

	return cache, nil
}

func (c *Cache) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT UNIQUE NOT NULL,
		type TEXT NOT NULL,
		title TEXT NOT NULL,
		frontmatter_json TEXT,
		tags TEXT,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_id INTEGER NOT NULL,
		target_id INTEGER NOT NULL,
		type TEXT NOT NULL,
		FOREIGN KEY (source_id) REFERENCES files(id),
		FOREIGN KEY (target_id) REFERENCES files(id)
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER NOT NULL,
		line INTEGER NOT NULL,
		text TEXT NOT NULL,
		status TEXT NOT NULL,
		parent_id INTEGER,
		has_note BOOLEAN DEFAULT FALSE,
		comment TEXT,
		FOREIGN KEY (file_id) REFERENCES files(id)
	);

	CREATE TABLE IF NOT EXISTS pomodoro (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id INTEGER,
		task_id INTEGER,
		started_at DATETIME NOT NULL,
		ended_at DATETIME NOT NULL,
		duration INTEGER NOT NULL,
		type TEXT NOT NULL,
		context TEXT,
		FOREIGN KEY (file_id) REFERENCES files(id),
		FOREIGN KEY (task_id) REFERENCES tasks(id)
	);

	CREATE TABLE IF NOT EXISTS daily_goals (
		date DATE PRIMARY KEY,
		target INTEGER NOT NULL,
		completed INTEGER DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_files_path ON files(path);
	CREATE INDEX IF NOT EXISTS idx_files_type ON files(type);
	CREATE INDEX IF NOT EXISTS idx_tasks_file ON tasks(file_id);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_pomodoro_date ON pomodoro(started_at);
	`

	_, err := c.db.Exec(schema)
	return err
}

// Close closes the database connection.
func (c *Cache) Close() error {
	return c.db.Close()
}

// SaveFile inserts or updates a parsed file in the cache.
func (c *Cache) SaveFile(file *types.File) (int64, error) {
	// Marshal frontmatter to JSON
	var frontmatterJSON []byte
	var err error
	if file.Frontmatter != nil {
		frontmatterJSON, err = json.Marshal(file.Frontmatter)
		if err != nil {
			return 0, err
		}
	}

	// Convert tags to comma-separated string
	tags := strings.Join(file.Tags, ",")

	// Upsert file record
	result, err := c.db.Exec(`
		INSERT INTO files (path, type, title, frontmatter_json, tags, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			type = excluded.type,
			title = excluded.title,
			frontmatter_json = excluded.frontmatter_json,
			tags = excluded.tags,
			updated_at = excluded.updated_at
	`, file.Path, string(file.Type), file.Title, string(frontmatterJSON), tags, file.ModifiedAt)
	if err != nil {
		return 0, err
	}

	// Get the file ID
	var fileID int64
	if file.ID > 0 {
		fileID = file.ID
	} else {
		fileID, err = result.LastInsertId()
		if err != nil {
			// If INSERT ON CONFLICT did UPDATE, LastInsertId returns 0
			// We need to query for the ID
			err = c.db.QueryRow("SELECT id FROM files WHERE path = ?", file.Path).Scan(&fileID)
			if err != nil {
				return 0, err
			}
		}
	}

	// Delete existing tasks for this file and re-insert
	_, err = c.db.Exec("DELETE FROM tasks WHERE file_id = ?", fileID)
	if err != nil {
		return fileID, err
	}

	// Save tasks recursively
	err = c.saveTasks(file.Tasks, fileID, nil)
	if err != nil {
		return fileID, err
	}

	return fileID, nil
}

// saveTasks recursively saves tasks and their subtasks.
func (c *Cache) saveTasks(tasks []types.Task, fileID int64, parentID *int64) error {
	for _, task := range tasks {
		result, err := c.db.Exec(`
			INSERT INTO tasks (file_id, line, text, status, parent_id, has_note, comment)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, fileID, task.Line, task.Text, task.Status, parentID, task.HasNote, task.Comment)
		if err != nil {
			return err
		}

		if len(task.Subtasks) > 0 {
			taskID, err := result.LastInsertId()
			if err != nil {
				return err
			}
			err = c.saveTasks(task.Subtasks, fileID, &taskID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetFile retrieves a cached file by path.
func (c *Cache) GetFile(path string) (*types.File, error) {
	row := c.db.QueryRow(`
		SELECT id, path, type, title, frontmatter_json, tags, updated_at
		FROM files WHERE path = ?
	`, path)

	var file types.File
	var frontmatterJSON sql.NullString
	var tags sql.NullString
	var fileType string

	err := row.Scan(&file.ID, &file.Path, &fileType, &file.Title, &frontmatterJSON, &tags, &file.ModifiedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	file.Type = types.FileType(fileType)

	// Parse frontmatter JSON
	if frontmatterJSON.Valid && frontmatterJSON.String != "" {
		err = json.Unmarshal([]byte(frontmatterJSON.String), &file.Frontmatter)
		if err != nil {
			return nil, err
		}
	}

	// Parse tags
	if tags.Valid && tags.String != "" {
		file.Tags = strings.Split(tags.String, ",")
	}

	// Load tasks for this file
	file.Tasks, err = c.getTasksForFile(file.ID, nil)
	if err != nil {
		return nil, err
	}

	return &file, nil
}

// getTasksForFile retrieves tasks for a file, optionally filtered by parent.
func (c *Cache) getTasksForFile(fileID int64, parentID *int64) ([]types.Task, error) {
	var rows *sql.Rows
	var err error

	if parentID == nil {
		rows, err = c.db.Query(`
			SELECT id, line, text, status, has_note, comment
			FROM tasks WHERE file_id = ? AND parent_id IS NULL
			ORDER BY line
		`, fileID)
	} else {
		rows, err = c.db.Query(`
			SELECT id, line, text, status, has_note, comment
			FROM tasks WHERE file_id = ? AND parent_id = ?
			ORDER BY line
		`, fileID, *parentID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []types.Task
	for rows.Next() {
		var task types.Task
		var comment sql.NullString

		err = rows.Scan(&task.ID, &task.Line, &task.Text, &task.Status, &task.HasNote, &comment)
		if err != nil {
			return nil, err
		}

		if comment.Valid {
			task.Comment = comment.String
		}
		task.FileID = fileID

		// Recursively get subtasks
		task.Subtasks, err = c.getTasksForFile(fileID, &task.ID)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// GetFilesByType retrieves all files of a given type.
func (c *Cache) GetFilesByType(fileType types.FileType) ([]*types.File, error) {
	rows, err := c.db.Query(`
		SELECT id, path, type, title, frontmatter_json, tags, updated_at
		FROM files WHERE type = ?
		ORDER BY updated_at DESC
	`, string(fileType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*types.File
	for rows.Next() {
		var file types.File
		var frontmatterJSON sql.NullString
		var tags sql.NullString
		var ft string

		err = rows.Scan(&file.ID, &file.Path, &ft, &file.Title, &frontmatterJSON, &tags, &file.ModifiedAt)
		if err != nil {
			return nil, err
		}

		file.Type = types.FileType(ft)

		if frontmatterJSON.Valid && frontmatterJSON.String != "" {
			json.Unmarshal([]byte(frontmatterJSON.String), &file.Frontmatter)
		}

		if tags.Valid && tags.String != "" {
			file.Tags = strings.Split(tags.String, ",")
		}

		// Load tasks
		file.Tasks, _ = c.getTasksForFile(file.ID, nil)

		files = append(files, &file)
	}

	return files, rows.Err()
}

// GetPendingTasks retrieves all incomplete tasks across all files.
func (c *Cache) GetPendingTasks() ([]types.Task, error) {
	rows, err := c.db.Query(`
		SELECT t.id, t.file_id, t.line, t.text, t.status, t.has_note, t.comment, f.path, f.title
		FROM tasks t
		JOIN files f ON t.file_id = f.id
		WHERE t.status != 'x' AND t.parent_id IS NULL
		ORDER BY f.updated_at DESC, t.line
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []types.Task
	for rows.Next() {
		var task types.Task
		var comment sql.NullString
		var filePath, fileTitle string

		err = rows.Scan(&task.ID, &task.FileID, &task.Line, &task.Text, &task.Status,
			&task.HasNote, &comment, &filePath, &fileTitle)
		if err != nil {
			return nil, err
		}

		if comment.Valid {
			task.Comment = comment.String
		}

		// Load subtasks
		task.Subtasks, _ = c.getTasksForFile(task.FileID, &task.ID)

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

// SavePomodoroSession logs a completed pomodoro session.
func (c *Cache) SavePomodoroSession(session *types.PomodoroSession) error {
	_, err := c.db.Exec(`
		INSERT INTO pomodoro (file_id, task_id, started_at, ended_at, duration, type, context)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, session.FileID, session.TaskID, session.StartedAt, session.EndedAt,
		session.Duration, string(session.Type), session.Context)
	return err
}

// GetDailyGoal retrieves the daily goal for a specific date.
func (c *Cache) GetDailyGoal(date time.Time) (*types.DailyGoal, error) {
	dateStr := date.Format("2006-01-02")

	row := c.db.QueryRow(`
		SELECT date, target, completed FROM daily_goals WHERE date = ?
	`, dateStr)

	var goal types.DailyGoal
	var dateVal string
	err := row.Scan(&dateVal, &goal.Target, &goal.Completed)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	goal.Date, _ = time.Parse("2006-01-02", dateVal)
	return &goal, nil
}

// SetDailyGoal sets the daily goal target for a specific date.
func (c *Cache) SetDailyGoal(date time.Time, target int) error {
	dateStr := date.Format("2006-01-02")
	_, err := c.db.Exec(`
		INSERT INTO daily_goals (date, target, completed)
		VALUES (?, ?, 0)
		ON CONFLICT(date) DO UPDATE SET target = excluded.target
	`, dateStr, target)
	return err
}

// IncrementDailyPomodoros increments the completed pomodoro count for today.
func (c *Cache) IncrementDailyPomodoros(date time.Time) error {
	dateStr := date.Format("2006-01-02")
	_, err := c.db.Exec(`
		INSERT INTO daily_goals (date, target, completed)
		VALUES (?, 5, 1)
		ON CONFLICT(date) DO UPDATE SET completed = completed + 1
	`, dateStr)
	return err
}

// GetPomodorosForDate retrieves all pomodoro sessions for a specific date.
func (c *Cache) GetPomodorosForDate(date time.Time) ([]types.PomodoroSession, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	rows, err := c.db.Query(`
		SELECT id, file_id, task_id, started_at, ended_at, duration, type, context
		FROM pomodoro
		WHERE started_at >= ? AND started_at < ?
		ORDER BY started_at
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []types.PomodoroSession
	for rows.Next() {
		var session types.PomodoroSession
		var fileID, taskID sql.NullInt64
		var context sql.NullString
		var pomType string

		err = rows.Scan(&session.ID, &fileID, &taskID, &session.StartedAt,
			&session.EndedAt, &session.Duration, &pomType, &context)
		if err != nil {
			return nil, err
		}

		if fileID.Valid {
			session.FileID = &fileID.Int64
		}
		if taskID.Valid {
			session.TaskID = &taskID.Int64
		}
		if context.Valid {
			session.Context = context.String
		}
		session.Type = types.PomodoroType(pomType)

		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// InvalidateFile marks a file as needing re-parsing by deleting it from cache.
func (c *Cache) InvalidateFile(path string) error {
	// Get file ID first
	var fileID int64
	err := c.db.QueryRow("SELECT id FROM files WHERE path = ?", path).Scan(&fileID)
	if err == sql.ErrNoRows {
		return nil // File not in cache, nothing to invalidate
	}
	if err != nil {
		return err
	}

	// Delete tasks for this file
	_, err = c.db.Exec("DELETE FROM tasks WHERE file_id = ?", fileID)
	if err != nil {
		return err
	}

	// Delete the file record
	_, err = c.db.Exec("DELETE FROM files WHERE id = ?", fileID)
	return err
}

// GetRecentFiles retrieves the most recently modified files.
func (c *Cache) GetRecentFiles(limit int) ([]*types.File, error) {
	rows, err := c.db.Query(`
		SELECT id, path, type, title, frontmatter_json, tags, updated_at
		FROM files
		ORDER BY updated_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []*types.File
	for rows.Next() {
		var file types.File
		var frontmatterJSON sql.NullString
		var tags sql.NullString
		var ft string

		err = rows.Scan(&file.ID, &file.Path, &ft, &file.Title, &frontmatterJSON, &tags, &file.ModifiedAt)
		if err != nil {
			return nil, err
		}

		file.Type = types.FileType(ft)

		if frontmatterJSON.Valid && frontmatterJSON.String != "" {
			json.Unmarshal([]byte(frontmatterJSON.String), &file.Frontmatter)
		}

		if tags.Valid && tags.String != "" {
			file.Tags = strings.Split(tags.String, ",")
		}

		files = append(files, &file)
	}

	return files, rows.Err()
}

// GetWeeklyStats retrieves pomodoro statistics for the past week.
func (c *Cache) GetWeeklyStats() (pomodorosByDay [7]int, totalPomodoros int, totalMinutes int, err error) {
	now := time.Now()
	// Find Monday of current week
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday becomes 7
	}
	monday := now.AddDate(0, 0, -(weekday - 1))
	monday = time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, now.Location())

	rows, err := c.db.Query(`
		SELECT date(started_at) as day, COUNT(*), SUM(duration)
		FROM pomodoro
		WHERE started_at >= ? AND type = 'work'
		GROUP BY day
		ORDER BY day
	`, monday)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var dayStr string
		var count, minutes int
		if err = rows.Scan(&dayStr, &count, &minutes); err != nil {
			return
		}

		day, _ := time.Parse("2006-01-02", dayStr)
		dayIndex := int(day.Weekday())
		if dayIndex == 0 {
			dayIndex = 6 // Sunday is index 6
		} else {
			dayIndex-- // Monday is index 0
		}

		if dayIndex >= 0 && dayIndex < 7 {
			pomodorosByDay[dayIndex] = count
		}
		totalPomodoros += count
		totalMinutes += minutes
	}

	return pomodorosByDay, totalPomodoros, totalMinutes, rows.Err()
}
