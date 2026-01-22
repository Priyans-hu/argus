package analyzer

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/karrick/godirwalk"
)

// OptimizedWalker uses godirwalk for 2-3x faster directory traversal
type OptimizedWalker struct {
	rootPath       string
	ignorePatterns []string
	defaultIgnore  []string
}

// NewOptimizedWalker creates a new optimized file walker
func NewOptimizedWalker(rootPath string) *OptimizedWalker {
	return &OptimizedWalker{
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

// Walk walks the directory tree using godirwalk and returns file information
func (w *OptimizedWalker) Walk() ([]types.FileInfo, error) {
	// Load .gitignore patterns
	w.loadGitignore()

	var files []types.FileInfo

	err := godirwalk.Walk(w.rootPath, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			// Get relative path
			relPath, err := filepath.Rel(w.rootPath, path)
			if err != nil {
				return nil
			}

			// Skip root
			if relPath == "." {
				return nil
			}

			isDir := de.IsDir()

			// Check if should ignore
			if w.shouldIgnore(relPath, isDir) {
				if isDir {
					return godirwalk.SkipThis
				}
				return nil
			}

			// Get file info for size (godirwalk doesn't provide it directly)
			var size int64
			if !isDir {
				if info, err := os.Stat(path); err == nil {
					size = info.Size()
				}
			}

			files = append(files, types.FileInfo{
				Path:      relPath,
				Name:      de.Name(),
				Extension: strings.ToLower(filepath.Ext(de.Name())),
				Size:      size,
				IsDir:     isDir,
			})

			return nil
		},
		ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
			// Skip files we can't access
			return godirwalk.SkipNode
		},
		Unsorted:            true,  // Faster without sorting
		AllowNonDirectory:   false, // Only allow starting from directories
		FollowSymbolicLinks: false, // Don't follow symlinks for safety
	})

	return files, err
}

// loadGitignore loads patterns from .gitignore file
func (w *OptimizedWalker) loadGitignore() {
	gitignorePath := filepath.Join(w.rootPath, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		return // No .gitignore, use defaults only
	}
	defer func() { _ = file.Close() }()

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
func (w *OptimizedWalker) shouldIgnore(path string, isDir bool) bool {
	name := filepath.Base(path)

	// Check default ignores
	for _, pattern := range w.defaultIgnore {
		if matched := matchPatternOptimized(pattern, name, path, isDir); matched {
			return true
		}
	}

	// Check .gitignore patterns
	for _, pattern := range w.ignorePatterns {
		if matched := matchPatternOptimized(pattern, name, path, isDir); matched {
			return true
		}
	}

	return false
}

// matchPatternOptimized matches a gitignore-style pattern (optimized version)
func matchPatternOptimized(pattern, name, path string, isDir bool) bool {
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

// WalkWithCallback walks the directory tree with a custom callback
// This is useful for streaming results instead of collecting all at once
func (w *OptimizedWalker) WalkWithCallback(callback func(types.FileInfo) error) error {
	// Load .gitignore patterns
	w.loadGitignore()

	return godirwalk.Walk(w.rootPath, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			// Get relative path
			relPath, err := filepath.Rel(w.rootPath, path)
			if err != nil {
				return nil
			}

			// Skip root
			if relPath == "." {
				return nil
			}

			isDir := de.IsDir()

			// Check if should ignore
			if w.shouldIgnore(relPath, isDir) {
				if isDir {
					return godirwalk.SkipThis
				}
				return nil
			}

			// Get file info for size
			var size int64
			if !isDir {
				if info, err := os.Stat(path); err == nil {
					size = info.Size()
				}
			}

			return callback(types.FileInfo{
				Path:      relPath,
				Name:      de.Name(),
				Extension: strings.ToLower(filepath.Ext(de.Name())),
				Size:      size,
				IsDir:     isDir,
			})
		},
		ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Unsorted:            true,
		AllowNonDirectory:   false,
		FollowSymbolicLinks: false,
	})
}

// WalkDirectoriesOnly walks only directories (faster for structure analysis)
func (w *OptimizedWalker) WalkDirectoriesOnly() ([]types.FileInfo, error) {
	w.loadGitignore()

	var dirs []types.FileInfo

	err := godirwalk.Walk(w.rootPath, &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			if !de.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(w.rootPath, path)
			if err != nil {
				return nil
			}

			if relPath == "." {
				return nil
			}

			if w.shouldIgnore(relPath, true) {
				return godirwalk.SkipThis
			}

			dirs = append(dirs, types.FileInfo{
				Path:  relPath,
				Name:  de.Name(),
				IsDir: true,
			})

			return nil
		},
		ErrorCallback: func(path string, err error) godirwalk.ErrorAction {
			return godirwalk.SkipNode
		},
		Unsorted:            true,
		AllowNonDirectory:   false,
		FollowSymbolicLinks: false,
	})

	return dirs, err
}
