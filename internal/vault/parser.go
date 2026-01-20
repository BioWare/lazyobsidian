// Package vault handles parsing and watching the Obsidian vault.
package vault

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BioWare/lazyobsidian/pkg/types"
)

var (
	// Regex patterns for parsing markdown
	taskPattern      = regexp.MustCompile(`^(\s*)-\s*\[([ x\-/>?])\]\s*(.+)$`)
	wikilinkPattern  = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]+)?\]\]`)
	frontmatterStart = regexp.MustCompile(`^---\s*$`)
	frontmatterEnd   = regexp.MustCompile(`^---\s*$`)
	tagPattern       = regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)
)

// Parser handles parsing markdown files from the vault.
type Parser struct {
	vaultPath string
}

// NewParser creates a new vault parser.
func NewParser(vaultPath string) *Parser {
	return &Parser{
		vaultPath: vaultPath,
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
	frontmatterLines := []string{}
	var contentBuilder strings.Builder

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle frontmatter
		if lineNum == 1 && frontmatterStart.MatchString(line) {
			inFrontmatter = true
			continue
		}

		if inFrontmatter {
			if frontmatterEnd.MatchString(line) {
				inFrontmatter = false
				result.Frontmatter = parseFrontmatter(frontmatterLines)
				continue
			}
			frontmatterLines = append(frontmatterLines, line)
			continue
		}

		// Parse tasks
		if matches := taskPattern.FindStringSubmatch(line); matches != nil {
			task := types.Task{
				Line:   lineNum,
				Status: matches[2],
				Text:   matches[3],
			}

			// Check for inline comment
			if idx := strings.Index(task.Text, " // "); idx != -1 {
				task.Comment = strings.TrimSpace(task.Text[idx+4:])
				task.Text = strings.TrimSpace(task.Text[:idx])
			}

			result.Tasks = append(result.Tasks, task)
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

	result.Content = contentBuilder.String()

	// Determine file type based on path and frontmatter
	result.Type = p.determineFileType(path, result.Frontmatter)

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

func (p *Parser) determineFileType(path string, frontmatter map[string]interface{}) types.FileType {
	// Check frontmatter type field first
	if frontmatter != nil {
		if t, ok := frontmatter["type"].(string); ok {
			switch t {
			case "daily":
				return types.FileTypeDaily
			case "goal", "yearly_plan", "monthly_plan":
				return types.FileTypeGoal
			case "course":
				return types.FileTypeCourse
			case "book":
				return types.FileTypeBook
			}
		}
	}

	// TODO: Check path against configured folders
	return types.FileTypeNote
}

func parseFrontmatter(lines []string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			result[key] = value
		}
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
