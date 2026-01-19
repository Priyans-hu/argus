package analyzer

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// Walker walks the directory tree and collects file information
type Walker struct {
	rootPath     string
	ignorePatterns []string
	defaultIgnore  []string
}

// NewWalker creates a new file walker
func NewWalker(rootPath string) *Walker {
	return &Walker{
		rootPath: rootPath,
		defaultIgnore: []string{
			".git",
			"node_modules",
			"vendor",
			"__pycache__",
			".venv",
			"venv",
			"dist",
			"build",
			".next",
			".nuxt",
			"target",
			"bin",
			"obj",
			".idea",
			".vscode",
			"*.log",
			"*.lock",
			"package-lock.json",
			"yarn.lock",
			"pnpm-lock.yaml",
			"go.sum",
			"Cargo.lock",
			"*.min.js",
			"*.min.css",
			"*.map",
		},
	}
}

// Walk walks the directory tree and returns file information
func (w *Walker) Walk() ([]types.FileInfo, error) {
	// Load .gitignore patterns
	w.loadGitignore()

	var files []types.FileInfo

	err := filepath.Walk(w.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		// Get relative path
		relPath, err := filepath.Rel(w.rootPath, path)
		if err != nil {
			return nil
		}

		// Skip root
		if relPath == "." {
			return nil
		}

		// Check if should ignore
		if w.shouldIgnore(relPath, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		files = append(files, types.FileInfo{
			Path:      relPath,
			Name:      info.Name(),
			Extension: strings.ToLower(filepath.Ext(info.Name())),
			Size:      info.Size(),
			IsDir:     info.IsDir(),
		})

		return nil
	})

	return files, err
}

// loadGitignore loads patterns from .gitignore file
func (w *Walker) loadGitignore() {
	gitignorePath := filepath.Join(w.rootPath, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		return // No .gitignore, use defaults only
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		w.ignorePatterns = append(w.ignorePatterns, line)
	}
}

// shouldIgnore checks if a path should be ignored
func (w *Walker) shouldIgnore(path string, isDir bool) bool {
	name := filepath.Base(path)

	// Check default ignores
	for _, pattern := range w.defaultIgnore {
		if matched := matchPattern(pattern, name, path, isDir); matched {
			return true
		}
	}

	// Check .gitignore patterns
	for _, pattern := range w.ignorePatterns {
		if matched := matchPattern(pattern, name, path, isDir); matched {
			return true
		}
	}

	return false
}

// matchPattern matches a gitignore-style pattern
func matchPattern(pattern, name, path string, isDir bool) bool {
	// Handle negation (we ignore negation for simplicity)
	if strings.HasPrefix(pattern, "!") {
		return false
	}

	// Handle directory-only patterns
	if strings.HasSuffix(pattern, "/") {
		if !isDir {
			return false
		}
		pattern = strings.TrimSuffix(pattern, "/")
	}

	// Handle patterns starting with /
	if strings.HasPrefix(pattern, "/") {
		pattern = strings.TrimPrefix(pattern, "/")
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	// Handle glob patterns
	if strings.Contains(pattern, "*") {
		matched, _ := filepath.Match(pattern, name)
		if matched {
			return true
		}
		matched, _ = filepath.Match(pattern, path)
		return matched
	}

	// Simple name match
	if name == pattern {
		return true
	}

	// Path contains pattern
	if strings.Contains(path, pattern) {
		return true
	}

	return false
}

// GetDirectories returns only directories from the file list
func GetDirectories(files []types.FileInfo) []types.FileInfo {
	var dirs []types.FileInfo
	for _, f := range files {
		if f.IsDir {
			dirs = append(dirs, f)
		}
	}
	return dirs
}

// GetFilesByExtension returns files with a specific extension
func GetFilesByExtension(files []types.FileInfo, ext string) []types.FileInfo {
	var result []types.FileInfo
	for _, f := range files {
		if !f.IsDir && f.Extension == ext {
			result = append(result, f)
		}
	}
	return result
}

// CountByExtension counts files by extension
func CountByExtension(files []types.FileInfo) map[string]int {
	counts := make(map[string]int)
	for _, f := range files {
		if !f.IsDir && f.Extension != "" {
			counts[f.Extension]++
		}
	}
	return counts
}
