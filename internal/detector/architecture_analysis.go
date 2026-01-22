package detector

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// ArchitectureAnalyzer performs deep architecture analysis on Go projects
type ArchitectureAnalyzer struct {
	rootPath string
	files    []types.FileInfo
}

// NewArchitectureAnalyzer creates a new architecture analyzer
func NewArchitectureAnalyzer(rootPath string, files []types.FileInfo) *ArchitectureAnalyzer {
	return &ArchitectureAnalyzer{
		rootPath: rootPath,
		files:    files,
	}
}

// PackageInfo holds information about a Go package
type PackageInfo struct {
	Path       string
	Name       string
	ImportPath string
	Imports    []string
	Files      []string
	IsMain     bool
	IsTest     bool
	HasTypes   bool
	HasFuncs   bool
}

// DependencyInfo holds dependency analysis results
type DependencyInfo struct {
	Packages        []PackageInfo
	Dependencies    map[string][]string // package -> dependencies
	Dependents      map[string][]string // package -> packages that depend on it
	CyclicDeps      [][]string          // groups of packages in cyclic dependencies
	LayerViolations []LayerViolation
}

// LayerViolation represents an architectural layer violation
type LayerViolation struct {
	From        string
	To          string
	FromLayer   string
	ToLayer     string
	Description string
}

// Analyze performs comprehensive architecture analysis
func (a *ArchitectureAnalyzer) Analyze() *DependencyInfo {
	info := &DependencyInfo{
		Dependencies: make(map[string][]string),
		Dependents:   make(map[string][]string),
	}

	// Find all Go packages
	packages := a.findPackages()
	info.Packages = packages

	// Build dependency graph
	for _, pkg := range packages {
		for _, imp := range pkg.Imports {
			// Only track internal imports
			if a.isInternalImport(imp) {
				info.Dependencies[pkg.Path] = append(info.Dependencies[pkg.Path], imp)
				info.Dependents[imp] = append(info.Dependents[imp], pkg.Path)
			}
		}
	}

	// Detect cyclic dependencies
	info.CyclicDeps = a.detectCycles(info.Dependencies)

	// Detect layer violations
	info.LayerViolations = a.detectLayerViolations(info.Dependencies)

	return info
}

// findPackages discovers all Go packages in the project
func (a *ArchitectureAnalyzer) findPackages() []PackageInfo {
	packageMap := make(map[string]*PackageInfo)

	for _, f := range a.files {
		if f.IsDir || !strings.HasSuffix(f.Name, ".go") {
			continue
		}

		dir := filepath.Dir(f.Path)
		if dir == "" {
			dir = "."
		}

		// Initialize package if not seen
		if _, ok := packageMap[dir]; !ok {
			packageMap[dir] = &PackageInfo{
				Path:    dir,
				Imports: []string{},
				Files:   []string{},
			}
		}

		pkg := packageMap[dir]
		pkg.Files = append(pkg.Files, f.Name)

		// Parse file to extract imports
		fullPath := filepath.Join(a.rootPath, f.Path)
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, fullPath, nil, parser.ImportsOnly)
		if err != nil {
			continue
		}

		// Set package name
		if pkg.Name == "" {
			pkg.Name = node.Name.Name
		}

		// Check if main package
		if node.Name.Name == "main" {
			pkg.IsMain = true
		}

		// Check if test file
		if strings.HasSuffix(f.Name, "_test.go") {
			pkg.IsTest = true
		}

		// Extract imports
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			if !sliceContainsString(pkg.Imports, importPath) {
				pkg.Imports = append(pkg.Imports, importPath)
			}
		}

		// Parse again for types and functions
		node, err = parser.ParseFile(fset, fullPath, nil, 0)
		if err != nil {
			continue
		}

		// Check for types and functions
		ast.Inspect(node, func(n ast.Node) bool {
			switch n.(type) {
			case *ast.TypeSpec:
				pkg.HasTypes = true
			case *ast.FuncDecl:
				pkg.HasFuncs = true
			}
			return true
		})
	}

	// Convert to slice
	var packages []PackageInfo
	for _, pkg := range packageMap {
		packages = append(packages, *pkg)
	}

	// Sort by path
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Path < packages[j].Path
	})

	return packages
}

// isInternalImport checks if an import is internal to the project
func (a *ArchitectureAnalyzer) isInternalImport(importPath string) bool {
	// Read go.mod to get module path
	goModPath := filepath.Join(a.rootPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return false
	}

	// Extract module path from go.mod
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimPrefix(line, "module ")
			modulePath = strings.TrimSpace(modulePath)
			return strings.HasPrefix(importPath, modulePath)
		}
	}

	return false
}

// detectCycles finds cyclic dependencies using DFS
func (a *ArchitectureAnalyzer) detectCycles(deps map[string][]string) [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{}

	var dfs func(node string)
	dfs = func(node string) {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, neighbor := range deps[node] {
			if !visited[neighbor] {
				dfs(neighbor)
			} else if recStack[neighbor] {
				// Found a cycle
				cycleStart := -1
				for i, p := range path {
					if p == neighbor {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle := make([]string, len(path)-cycleStart)
					copy(cycle, path[cycleStart:])
					cycles = append(cycles, cycle)
				}
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
	}

	for node := range deps {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// detectLayerViolations checks for architectural layer violations
func (a *ArchitectureAnalyzer) detectLayerViolations(deps map[string][]string) []LayerViolation {
	var violations []LayerViolation

	// Define layer hierarchy (higher layers can depend on lower layers, not vice versa)
	layers := map[string]int{
		"cmd":        5, // Entry points
		"api":        4, // API/handlers
		"service":    3, // Business logic
		"repository": 2, // Data access
		"domain":     1, // Domain models
		"pkg":        0, // Shared utilities
		"internal":   0, // Internal utilities
	}

	// Get layer for a package path
	getLayer := func(path string) (string, int) {
		parts := strings.Split(path, string(filepath.Separator))
		for _, part := range parts {
			if level, ok := layers[part]; ok {
				return part, level
			}
		}
		return "", -1
	}

	// Check each dependency
	for from, imports := range deps {
		fromLayer, fromLevel := getLayer(from)
		if fromLevel < 0 {
			continue
		}

		for _, to := range imports {
			toLayer, toLevel := getLayer(to)
			if toLevel < 0 {
				continue
			}

			// Violation: lower layer depending on higher layer
			if fromLevel < toLevel {
				violations = append(violations, LayerViolation{
					From:        from,
					To:          to,
					FromLayer:   fromLayer,
					ToLayer:     toLayer,
					Description: fromLayer + " should not depend on " + toLayer,
				})
			}
		}
	}

	return violations
}

// GetArchitecturePatterns returns detected architectural patterns
func (a *ArchitectureAnalyzer) GetArchitecturePatterns() []types.PatternInfo {
	info := a.Analyze()
	var patterns []types.PatternInfo

	// Report cyclic dependencies
	if len(info.CyclicDeps) > 0 {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Cyclic Dependencies",
			Category:    "architecture-issue",
			Description: "Cyclic dependencies detected between packages",
			FileCount:   len(info.CyclicDeps),
			Examples:    formatCycles(info.CyclicDeps, 3),
		})
	}

	// Report layer violations
	if len(info.LayerViolations) > 0 {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Layer Violations",
			Category:    "architecture-issue",
			Description: "Architectural layer violations detected",
			FileCount:   len(info.LayerViolations),
			Examples:    formatViolations(info.LayerViolations, 3),
		})
	}

	// Detect Clean Architecture
	if hasCleanArchitecture(info.Packages) {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Clean Architecture",
			Category:    "architecture-pattern",
			Description: "Project follows Clean Architecture pattern",
			FileCount:   len(info.Packages),
		})
	}

	// Detect Hexagonal Architecture
	if hasHexagonalArchitecture(info.Packages) {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Hexagonal Architecture",
			Category:    "architecture-pattern",
			Description: "Project follows Hexagonal/Ports & Adapters pattern",
			FileCount:   len(info.Packages),
		})
	}

	// Detect Standard Go Layout
	if hasStandardGoLayout(info.Packages) {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Standard Go Layout",
			Category:    "architecture-pattern",
			Description: "Project follows Standard Go Project Layout",
			FileCount:   len(info.Packages),
		})
	}

	// Report package metrics
	patterns = append(patterns, types.PatternInfo{
		Name:        "Package Count",
		Category:    "architecture-metric",
		Description: "Number of Go packages in project",
		FileCount:   len(info.Packages),
	})

	// Report highly coupled packages
	highlyCoupled := findHighlyCoupledPackages(info.Dependencies, info.Dependents)
	if len(highlyCoupled) > 0 {
		patterns = append(patterns, types.PatternInfo{
			Name:        "Highly Coupled Packages",
			Category:    "architecture-metric",
			Description: "Packages with many dependencies or dependents",
			FileCount:   len(highlyCoupled),
			Examples:    highlyCoupled,
		})
	}

	return patterns
}

// Helper functions

func sliceContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func formatCycles(cycles [][]string, max int) []string {
	var result []string
	for i, cycle := range cycles {
		if i >= max {
			break
		}
		result = append(result, strings.Join(cycle, " -> ")+" -> "+cycle[0])
	}
	return result
}

func formatViolations(violations []LayerViolation, max int) []string {
	var result []string
	for i, v := range violations {
		if i >= max {
			break
		}
		result = append(result, v.From+" -> "+v.To+" ("+v.Description+")")
	}
	return result
}

func hasCleanArchitecture(packages []PackageInfo) bool {
	// Clean Architecture markers
	markers := []string{"domain", "usecase", "repository", "delivery", "infrastructure"}
	found := 0
	for _, pkg := range packages {
		for _, marker := range markers {
			if strings.Contains(pkg.Path, marker) {
				found++
				break
			}
		}
	}
	return found >= 3
}

func hasHexagonalArchitecture(packages []PackageInfo) bool {
	// Hexagonal Architecture markers
	markers := []string{"port", "adapter", "application", "domain", "infrastructure"}
	found := 0
	for _, pkg := range packages {
		for _, marker := range markers {
			if strings.Contains(pkg.Path, marker) {
				found++
				break
			}
		}
	}
	return found >= 3
}

func hasStandardGoLayout(packages []PackageInfo) bool {
	// Standard Go Layout markers
	hasCmd := false
	hasInternal := false
	hasPkg := false

	for _, pkg := range packages {
		if strings.HasPrefix(pkg.Path, "cmd") {
			hasCmd = true
		}
		if strings.HasPrefix(pkg.Path, "internal") {
			hasInternal = true
		}
		if strings.HasPrefix(pkg.Path, "pkg") {
			hasPkg = true
		}
	}

	return hasCmd && (hasInternal || hasPkg)
}

func findHighlyCoupledPackages(deps, dependents map[string][]string) []string {
	var result []string
	threshold := 5 // Package with more than 5 deps or dependents is considered highly coupled

	seen := make(map[string]bool)
	for pkg, imports := range deps {
		if len(imports) > threshold && !seen[pkg] {
			result = append(result, pkg+" ("+string(rune('0'+len(imports)))+" deps)")
			seen[pkg] = true
		}
	}

	for pkg, users := range dependents {
		if len(users) > threshold && !seen[pkg] {
			result = append(result, pkg+" ("+string(rune('0'+len(users)))+" dependents)")
			seen[pkg] = true
		}
	}

	// Limit results
	if len(result) > 5 {
		result = result[:5]
	}

	return result
}
