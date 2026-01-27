package usage

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ClaudeProjectsDir returns the base directory for Claude Code project logs
func ClaudeProjectsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".claude", "projects")
}

// EncodeProjectPath converts an absolute project path to Claude's encoded directory name.
// Claude Code encodes paths by replacing "/" and "." with "-": /Users/foo.bar/project → -Users-foo-bar-project
func EncodeProjectPath(absPath string) string {
	encoded := strings.ReplaceAll(absPath, string(filepath.Separator), "-")
	encoded = strings.ReplaceAll(encoded, ".", "-")
	return encoded
}

// DiscoverProjectDir finds the Claude Code project directory for a given project path.
// Returns the full path to the directory under ~/.claude/projects/, or empty string if not found.
func DiscoverProjectDir(absPath string) string {
	baseDir := ClaudeProjectsDir()
	if baseDir == "" {
		return ""
	}

	encoded := EncodeProjectPath(absPath)
	projectDir := filepath.Join(baseDir, encoded)

	if info, err := os.Stat(projectDir); err == nil && info.IsDir() {
		return projectDir
	}

	return ""
}

// DiscoverSessions lists JSONL session files in a Claude project directory.
// If since is non-zero, only files modified after that time are returned.
func DiscoverSessions(projectDir string, since time.Time) ([]string, error) {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, err
	}

	var sessions []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}

		// Apply date filter using file modification time
		if !since.IsZero() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if info.ModTime().Before(since) {
				continue
			}
		}

		sessions = append(sessions, filepath.Join(projectDir, entry.Name()))
	}

	return sessions, nil
}

// DiscoverSubagents lists subagent JSONL files for a given session.
// Session directories are named by UUID and contain a subagents/ subdirectory.
func DiscoverSubagents(projectDir, sessionID string) ([]string, error) {
	subagentDir := filepath.Join(projectDir, sessionID, "subagents")

	if info, err := os.Stat(subagentDir); err != nil || !info.IsDir() {
		return nil, nil
	}

	entries, err := os.ReadDir(subagentDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".jsonl") {
			files = append(files, filepath.Join(subagentDir, entry.Name()))
		}
	}

	return files, nil
}
