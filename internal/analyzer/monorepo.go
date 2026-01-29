package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Priyans-hu/argus/pkg/types"
)

// WorkspaceResult holds the analysis result for a single workspace
type WorkspaceResult struct {
	Path     string          // relative path from monorepo root
	Name     string          // workspace name (from package.json or directory name)
	Analysis *types.Analysis // full analysis of the workspace
	Error    error           // non-nil if analysis failed
}

// MonorepoAnalyzer orchestrates per-workspace analysis in a monorepo
type MonorepoAnalyzer struct {
	rootPath      string
	parallel      bool
	maxConcurrent int
}

// NewMonorepoAnalyzer creates a new monorepo analyzer
func NewMonorepoAnalyzer(rootPath string, parallel bool, maxConcurrent int) *MonorepoAnalyzer {
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}
	return &MonorepoAnalyzer{
		rootPath:      rootPath,
		parallel:      parallel,
		maxConcurrent: maxConcurrent,
	}
}

// AnalyzeWorkspaces runs full analysis on each resolved workspace directory
func (ma *MonorepoAnalyzer) AnalyzeWorkspaces(ctx context.Context, monoInfo *types.MonorepoInfo) []WorkspaceResult {
	workspaceDirs := ma.resolveWorkspaces(monoInfo)
	if len(workspaceDirs) == 0 {
		slog.Debug("MonorepoAnalyzer: no workspace directories resolved")
		return nil
	}

	slog.Debug("MonorepoAnalyzer: analyzing workspaces",
		"count", len(workspaceDirs), "parallel", ma.parallel)

	if !ma.parallel {
		return ma.analyzeSequential(ctx, workspaceDirs)
	}
	return ma.analyzeParallel(ctx, workspaceDirs)
}

// resolveWorkspaces expands glob patterns in WorkspacePaths and Packages into actual directories
func (ma *MonorepoAnalyzer) resolveWorkspaces(info *types.MonorepoInfo) []string {
	seen := make(map[string]bool)
	var resolved []string

	// Resolve from WorkspacePaths (e.g., "packages/*", "apps/*")
	for _, pattern := range info.WorkspacePaths {
		matches, err := filepath.Glob(filepath.Join(ma.rootPath, pattern))
		if err != nil {
			slog.Debug("MonorepoAnalyzer: glob failed", "pattern", pattern, "error", err)
			continue
		}
		for _, match := range matches {
			fi, err := os.Stat(match)
			if err != nil || !fi.IsDir() {
				continue
			}
			rel, err := filepath.Rel(ma.rootPath, match)
			if err != nil {
				continue
			}
			if !seen[rel] {
				seen[rel] = true
				resolved = append(resolved, rel)
			}
		}
	}

	// Resolve from Packages (detected apps/, packages/, etc.)
	for _, pkg := range info.Packages {
		for _, sub := range pkg.SubPackages {
			dir := filepath.Join(pkg.Path, sub)
			absDir := filepath.Join(ma.rootPath, dir)
			fi, err := os.Stat(absDir)
			if err != nil || !fi.IsDir() {
				continue
			}
			if !seen[dir] {
				seen[dir] = true
				resolved = append(resolved, dir)
			}
		}
	}

	return resolved
}

func (ma *MonorepoAnalyzer) analyzeSequential(ctx context.Context, dirs []string) []WorkspaceResult {
	results := make([]WorkspaceResult, 0, len(dirs))
	for _, dir := range dirs {
		select {
		case <-ctx.Done():
			return results
		default:
		}
		result := ma.analyzeOne(ctx, dir)
		results = append(results, result)
	}
	return results
}

func (ma *MonorepoAnalyzer) analyzeParallel(ctx context.Context, dirs []string) []WorkspaceResult {
	results := make([]WorkspaceResult, len(dirs))
	sem := make(chan struct{}, ma.maxConcurrent)
	var wg sync.WaitGroup

	for i, dir := range dirs {
		wg.Add(1)
		go func(idx int, wsDir string) {
			defer wg.Done()
			sem <- struct{}{} // acquire
			defer func() { <-sem }()

			select {
			case <-ctx.Done():
				results[idx] = WorkspaceResult{
					Path:  wsDir,
					Name:  filepath.Base(wsDir),
					Error: ctx.Err(),
				}
				return
			default:
			}

			results[idx] = ma.analyzeOne(ctx, wsDir)
		}(i, dir)
	}

	wg.Wait()
	return results
}

func (ma *MonorepoAnalyzer) analyzeOne(ctx context.Context, relDir string) WorkspaceResult {
	absDir := filepath.Join(ma.rootPath, relDir)
	name := workspaceName(absDir, relDir)

	start := time.Now()
	slog.Debug("MonorepoAnalyzer: analyzing workspace", "workspace", relDir)

	var analysis *types.Analysis
	var err error

	if ma.parallel {
		pa := NewParallelAnalyzer(absDir, nil)
		analysis, err = pa.Analyze(ctx)
	} else {
		a := NewAnalyzer(absDir, nil)
		analysis, err = a.Analyze(ctx)
	}

	if err != nil {
		slog.Debug("MonorepoAnalyzer: workspace analysis failed",
			"workspace", relDir, "error", err)
		return WorkspaceResult{Path: relDir, Name: name, Error: err}
	}

	slog.Debug("MonorepoAnalyzer: workspace done",
		"workspace", relDir, "duration", time.Since(start))

	return WorkspaceResult{Path: relDir, Name: name, Analysis: analysis}
}

// workspaceName tries to extract a meaningful name from the workspace
func workspaceName(absDir, relDir string) string {
	// Try package.json "name" field
	pkgPath := filepath.Join(absDir, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		name := extractJSONField(data, "name")
		if name != "" {
			return name
		}
	}

	// Fallback to directory name
	return filepath.Base(relDir)
}

// extractJSONField is a simple helper to get a string field from JSON without importing encoding/json
func extractJSONField(data []byte, field string) string {
	key := fmt.Sprintf(`"%s"`, field)
	idx := strings.Index(string(data), key)
	if idx < 0 {
		return ""
	}
	rest := string(data[idx+len(key):])
	// skip `:` and whitespace
	rest = strings.TrimLeft(rest, ": \t\n\r")
	if len(rest) == 0 || rest[0] != '"' {
		return ""
	}
	rest = rest[1:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	return rest[:end]
}
