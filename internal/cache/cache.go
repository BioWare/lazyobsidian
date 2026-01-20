// Package cache provides SQLite-based caching for parsed vault data.
package cache

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
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

// TODO: Add methods for CRUD operations on files, tasks, pomodoro, etc.
