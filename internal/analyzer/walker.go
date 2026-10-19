package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Priyans-hu/argus/internal/generator"
	"github.com/Priyans-hu/argus/pkg/types"
)

// defaultIgnore is matched by entry name at any depth.
var defaultIgnore = []string{
	".git", "node_modules", "vendor", "__pycache__", ".venv", "venv",
	"dist", "build", ".next", ".nuxt", "target", "bin", "obj", ".idea", ".vscode",
	"*.log", "*.lock", "package-lock.json", "yarn.lock", "pnpm-lock.yaml",
	"go.sum", "Cargo.lock", "*.min.js", "*.min.css", "*.map",
}

// Walker walks the directory tree and collects file information
type Walker struct {
	rootPath string
	ignore   []string // patterns from .argus.yaml
}

// NewWalker creates a new file walker; cfg may be nil
func NewWalker(rootPath string, cfg *types.Config) *Walker {
	w := &Walker{rootPath: rootPath}
	if cfg != nil {
		w.ignore = cfg.Ignore
	}
	return w
}

// Walk walks the directory tree and returns file information
func (w *Walker) Walk(ctx context.Context) ([]types.FileInfo, error) {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	rules := w.rules()
	var files []types.FileInfo

	err := filepath.Walk(w.rootPath, func(path string, info os.FileInfo, err error) error {
		// Check for cancellation during walk
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

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

		// argus's own outputs are not project files; then the ignore rules
		if generator.IsGeneratedPath(relPath) || slices.ContainsFunc(rules, func(r ignoreRule) bool {
			return r.matches(relPath, info.IsDir())
		}) {
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

// rules returns the built-in defaults, the root .gitignore lines and the
// configured patterns. Nested .gitignore files are not read.
func (w *Walker) rules() []ignoreRule {
	lines := slices.Clone(defaultIgnore)
	if data, err := os.ReadFile(filepath.Join(w.rootPath, ".gitignore")); err == nil {
		lines = append(lines, strings.Split(string(data), "\n")...)
	}
	lines = append(lines, w.ignore...)

	var rules []ignoreRule
	for _, line := range lines {
		if r, ok := parseIgnoreRule(line); ok {
			rules = append(rules, r)
		}
	}
	return rules
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
