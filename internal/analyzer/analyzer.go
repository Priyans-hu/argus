package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
)

// Analyzer coordinates the codebase analysis
type Analyzer struct {
	rootPath string
	config   *types.Config
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(rootPath string, config *types.Config) *Analyzer {
	return &Analyzer{
		rootPath: rootPath,
		config:   config,
	}
}

// Analyze performs the full codebase analysis
func (a *Analyzer) Analyze() (*types.Analysis, error) {
	// Get absolute path
	absPath, err := filepath.Abs(a.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Extract project name from directory
	projectName := filepath.Base(absPath)

	// Walk the file tree
	walker := NewWalker(absPath)
	files, err := walker.Walk()
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Initialize analysis
	analysis := &types.Analysis{
		ProjectName: projectName,
		RootPath:    absPath,
	}

	// Detect tech stack
	techDetector := detector.NewTechStackDetector(absPath, files)
	techStack, err := techDetector.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect tech stack: %w", err)
	}
	analysis.TechStack = *techStack

	// Detect structure
	structureDetector := detector.NewStructureDetector(absPath, files)
	structure, err := structureDetector.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect structure: %w", err)
	}
	analysis.Structure = *structure

	// Detect key files
	analysis.KeyFiles = structureDetector.DetectKeyFiles()

	// Detect commands
	analysis.Commands = detector.DetectCommands(absPath)

	// Detect dependencies
	analysis.Dependencies = a.detectDependencies(absPath)

	// Detect conventions (basic for now)
	analysis.Conventions = a.detectConventions(files)

	return analysis, nil
}

// detectDependencies extracts dependencies from package managers
func (a *Analyzer) detectDependencies(rootPath string) []types.Dependency {
	var deps []types.Dependency

	// Try package.json
	pkgPath := filepath.Join(rootPath, "package.json")
	if data, err := os.ReadFile(pkgPath); err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			for name, version := range pkg.Dependencies {
				deps = append(deps, types.Dependency{
					Name:    name,
					Version: version,
					Type:    "runtime",
				})
			}
			for name, version := range pkg.DevDependencies {
				deps = append(deps, types.Dependency{
					Name:    name,
					Version: version,
					Type:    "dev",
				})
			}
		}
	}

	// Try go.mod (simplified)
	modPath := filepath.Join(rootPath, "go.mod")
	if data, err := os.ReadFile(modPath); err == nil {
		lines := splitLines(string(data))
		inRequire := false
		for _, line := range lines {
			line = trimSpace(line)
			if line == "require (" {
				inRequire = true
				continue
			}
			if line == ")" {
				inRequire = false
				continue
			}
			if inRequire && line != "" {
				parts := splitFields(line)
				if len(parts) >= 2 {
					deps = append(deps, types.Dependency{
						Name:    parts[0],
						Version: parts[1],
						Type:    "runtime",
					})
				}
			}
		}
	}

	return deps
}

// detectConventions detects basic coding conventions
func (a *Analyzer) detectConventions(files []types.FileInfo) []types.Convention {
	var conventions []types.Convention

	// Check for TypeScript
	hasTS := false
	hasJS := false
	for _, f := range files {
		if f.Extension == ".ts" || f.Extension == ".tsx" {
			hasTS = true
		}
		if f.Extension == ".js" || f.Extension == ".jsx" {
			hasJS = true
		}
	}

	if hasTS {
		conventions = append(conventions, types.Convention{
			Category:    "language",
			Description: "TypeScript is used - prefer .ts/.tsx files over .js/.jsx",
		})
	}

	if hasTS && !hasJS {
		conventions = append(conventions, types.Convention{
			Category:    "language",
			Description: "Pure TypeScript project - avoid adding JavaScript files",
		})
	}

	// Check for common file patterns
	extCounts := CountByExtension(files)

	// React/JSX convention
	if extCounts[".jsx"] > 0 || extCounts[".tsx"] > 0 {
		conventions = append(conventions, types.Convention{
			Category:    "components",
			Description: "React components use JSX/TSX syntax",
		})
	}

	// Go conventions
	if extCounts[".go"] > 0 {
		conventions = append(conventions, types.Convention{
			Category:    "language",
			Description: "Go project - use gofmt for formatting",
		})
	}

	return conventions
}

// Helper functions to avoid importing strings package multiple times
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func splitFields(s string) []string {
	var fields []string
	start := -1
	for i, c := range s {
		if c == ' ' || c == '\t' {
			if start >= 0 {
				fields = append(fields, s[start:i])
				start = -1
			}
		} else {
			if start < 0 {
				start = i
			}
		}
	}
	if start >= 0 {
		fields = append(fields, s[start:])
	}
	return fields
}
