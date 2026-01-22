package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// ArchitectureDetector detects high-level architecture patterns
type ArchitectureDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewArchitectureDetector creates a new architecture detector
func NewArchitectureDetector(rootPath string, files []types.FileInfo) *ArchitectureDetector {
	return &ArchitectureDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes and returns architecture information
func (d *ArchitectureDetector) Detect() *types.ArchitectureInfo {
	info := &types.ArchitectureInfo{}

	// Detect architecture style based on directory structure
	info.Style = d.detectArchitectureStyle()

	// Detect layers and their relationships
	info.Layers = d.detectLayers()

	// Find entry point
	info.EntryPoint = d.findEntryPoint()

	// Generate text diagram
	info.Diagram = d.generateDiagram(info)

	return info
}

// detectArchitectureStyle determines the architecture pattern
func (d *ArchitectureDetector) detectArchitectureStyle() string {
	dirs := make(map[string]bool)
	pyFiles := make(map[string]bool)
	for _, f := range d.files {
		if f.IsDir {
			dirs[strings.ToLower(f.Name)] = true
		}
		// Track Python entry point files
		if !f.IsDir && strings.HasSuffix(f.Name, ".py") {
			pyFiles[strings.ToLower(f.Name)] = true
		}
	}

	// Check for Django
	if pyFiles["manage.py"] || pyFiles["wsgi.py"] || dirs["apps"] || hasPyFile(d.rootPath, "settings.py") {
		return "Django"
	}

	// Check for Flask
	if (pyFiles["app.py"] || pyFiles["wsgi.py"] || pyFiles["asgi.py"]) && (dirs["blueprints"] || dirs["routes"] || dirs["views"]) {
		return "Flask (Blueprints)"
	}
	if pyFiles["app.py"] || pyFiles["wsgi.py"] {
		return "Flask"
	}

	// Check for FastAPI
	if (pyFiles["main.py"] || pyFiles["app.py"]) && (dirs["routers"] || dirs["schemas"] || dirs["api"]) {
		return "FastAPI"
	}

	// Check for common Go project structure
	hasCmd := dirs["cmd"]
	hasInternal := dirs["internal"]
	hasPkg := dirs["pkg"]

	if hasCmd && hasInternal {
		return "Standard Go Layout"
	}

	// Check for Clean Architecture
	if dirs["domain"] && (dirs["infrastructure"] || dirs["adapters"] || dirs["ports"]) {
		return "Clean Architecture"
	}

	// Check for Hexagonal
	if dirs["ports"] && dirs["adapters"] {
		return "Hexagonal Architecture"
	}

	// Check for MVC
	if dirs["models"] && dirs["views"] && dirs["controllers"] {
		return "MVC"
	}

	// Check for Feature-based
	if dirs["features"] || dirs["modules"] {
		return "Feature-based"
	}

	// Default for Go projects
	if hasPkg || hasInternal {
		return "Go Package Layout"
	}

	return ""
}

// hasPyFile checks if a Python file exists anywhere in the project
func hasPyFile(rootPath, filename string) bool {
	found := false
	_ = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || found {
			return nil
		}
		if !d.IsDir() && strings.ToLower(d.Name()) == filename {
			found = true
		}
		return nil
	})
	return found
}

// detectLayers identifies architectural layers and their dependencies
func (d *ArchitectureDetector) detectLayers() []types.ArchitectureLayer {
	var layers []types.ArchitectureLayer

	// Group directories by depth and purpose, deduplicating
	topLevelDirs := make(map[string]map[string]bool)

	for _, f := range d.files {
		if !f.IsDir {
			continue
		}

		parts := strings.Split(f.Path, string(filepath.Separator))
		if len(parts) == 0 {
			continue
		}

		topLevel := parts[0]
		if topLevelDirs[topLevel] == nil {
			topLevelDirs[topLevel] = make(map[string]bool)
		}
		if len(parts) > 1 {
			topLevelDirs[topLevel][parts[1]] = true
		}
	}

	// Define layer order and purposes
	layerOrder := []struct {
		name    string
		purpose string
	}{
		{"cmd", "Entry points / CLI"},
		{"api", "API handlers"},
		{"routers", "API routes"},
		{"blueprints", "Flask blueprints"},
		{"apps", "Django apps"},
		{"internal", "Private packages"},
		{"pkg", "Public packages"},
		{"domain", "Business logic"},
		{"services", "Service layer"},
		{"handlers", "Request handlers"},
		{"views", "View layer"},
		{"controllers", "Controllers"},
		{"models", "Data models"},
		{"schemas", "Data schemas"},
		{"repository", "Data access"},
		{"config", "Configuration"},
		{"middleware", "Middleware"},
	}

	for _, layerDef := range layerOrder {
		subDirMap, exists := topLevelDirs[layerDef.name]
		if !exists || len(subDirMap) == 0 {
			continue
		}

		// Convert map to sorted slice
		var subDirs []string
		for dir := range subDirMap {
			subDirs = append(subDirs, dir)
		}
		sort.Strings(subDirs)

		layer := types.ArchitectureLayer{
			Name:     layerDef.name,
			Purpose:  layerDef.purpose,
			Packages: subDirs,
		}

		// Detect dependencies by analyzing imports
		layer.DependsOn = d.detectLayerDependencies(layerDef.name)

		layers = append(layers, layer)
	}

	return layers
}

// detectLayerDependencies analyzes imports to find layer dependencies
func (d *ArchitectureDetector) detectLayerDependencies(layer string) []string {
	deps := make(map[string]bool)
	// Match individual import lines: "path/to/pkg" or alias "path/to/pkg"
	importLineRegex := regexp.MustCompile(`"([^"]+/internal/([^"/]+))"`)
	// Valid Go package names: alphanumeric and underscore only
	validPkgRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)

	layerPath := filepath.Join(d.rootPath, layer)
	_ = filepath.WalkDir(layerPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Find all imports that reference /internal/
		matches := importLineRegex.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) >= 3 {
				pkg := match[2]
				// Validate it's a real package name (not regex artifacts)
				if pkg != "" && pkg != layer && validPkgRegex.MatchString(pkg) {
					deps[pkg] = true
				}
			}
		}

		return nil
	})

	var result []string
	for dep := range deps {
		result = append(result, dep)
	}
	sort.Strings(result)
	return result
}

// findEntryPoint finds the main entry point
func (d *ArchitectureDetector) findEntryPoint() string {
	// Python entry points
	pythonEntryPoints := []string{
		"manage.py",   // Django
		"wsgi.py",     // Flask/Django WSGI
		"asgi.py",     // FastAPI/Django ASGI
		"app.py",      // Flask/FastAPI
		"main.py",     // Generic Python
		"run.py",      // Common Flask pattern
		"__main__.py", // Python module entry
	}

	for _, entry := range pythonEntryPoints {
		if _, err := os.Stat(filepath.Join(d.rootPath, entry)); err == nil {
			return entry
		}
	}

	// Look for main.go in cmd/
	cmdDir := filepath.Join(d.rootPath, "cmd")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		// Try root main.go
		if _, err := os.Stat(filepath.Join(d.rootPath, "main.go")); err == nil {
			return "main.go"
		}
		return ""
	}

	// Get project name from go.mod for prioritization
	projectName := d.getProjectNameFromGoMod()

	// Utility command patterns to skip
	utilityPatterns := []string{"gen", "doc", "test", "mock", "tool", "util", "example", "demo"}

	// Find all valid cmd directories with main.go
	type cmdEntry struct {
		name string
		path string
		size int64
	}
	var candidates []cmdEntry

	for _, entry := range entries {
		if entry.IsDir() {
			mainPath := filepath.Join(cmdDir, entry.Name(), "main.go")
			if stat, err := os.Stat(mainPath); err == nil {
				candidates = append(candidates, cmdEntry{
					name: entry.Name(),
					path: "cmd/" + entry.Name() + "/main.go",
					size: stat.Size(),
				})
			}
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	// Priority 1: cmd/{project-name}/main.go
	if projectName != "" {
		for _, c := range candidates {
			if c.name == projectName {
				return c.path
			}
		}
	}

	// Priority 2: Skip utility commands and select the main binary
	var nonUtility []cmdEntry
	for _, c := range candidates {
		isUtility := false
		lowerName := strings.ToLower(c.name)
		for _, pattern := range utilityPatterns {
			if strings.Contains(lowerName, pattern) {
				isUtility = true
				break
			}
		}
		if !isUtility {
			nonUtility = append(nonUtility, c)
		}
	}

	// If we filtered to only utilities or nothing, use all candidates
	if len(nonUtility) == 0 {
		nonUtility = candidates
	}

	// Priority 3: Select the one with most code (largest main.go)
	if len(nonUtility) > 0 {
		largestIdx := 0
		for i, c := range nonUtility {
			if c.size > nonUtility[largestIdx].size {
				largestIdx = i
			}
		}
		return nonUtility[largestIdx].path
	}

	// Fallback: return first candidate
	return candidates[0].path
}

// getProjectNameFromGoMod extracts project name from go.mod
func (d *ArchitectureDetector) getProjectNameFromGoMod() string {
	goModPath := filepath.Join(d.rootPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}

	// Parse module line: "module github.com/user/project"
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			modulePath := strings.TrimPrefix(line, "module ")
			modulePath = strings.TrimSpace(modulePath)
			// Extract last component
			parts := strings.Split(modulePath, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}

	return ""
}

// generateDiagram creates a text-based architecture diagram
func (d *ArchitectureDetector) generateDiagram(info *types.ArchitectureInfo) string {
	if len(info.Layers) == 0 {
		return ""
	}

	var sb strings.Builder

	// Simple layered diagram
	sb.WriteString("```\n")

	// Entry point at top
	if info.EntryPoint != "" {
		entryName := filepath.Base(filepath.Dir(info.EntryPoint))
		if entryName == "." || entryName == "" {
			entryName = "main"
		}
		sb.WriteString("                    ┌─────────────┐\n")
		sb.WriteString("                    │   " + centerPad(entryName, 7) + "   │\n")
		sb.WriteString("                    └──────┬──────┘\n")
		sb.WriteString("                           │\n")
	}

	// Find internal packages
	var internalPkgs []string
	for _, layer := range info.Layers {
		if layer.Name == "internal" {
			internalPkgs = layer.Packages
			break
		}
	}

	if len(internalPkgs) > 0 {
		// Group into rows of 3
		rows := groupStrings(internalPkgs, 3)

		for i, row := range rows {
			// Draw boxes
			var line1, line2, line3 strings.Builder

			for j, pkg := range row {
				boxWidth := 13
				paddedName := centerPad(pkg, boxWidth-2)

				line1.WriteString("┌")
				line1.WriteString(strings.Repeat("─", boxWidth))
				line1.WriteString("┐")

				line2.WriteString("│ ")
				line2.WriteString(paddedName)
				line2.WriteString(" │")

				line3.WriteString("└")
				line3.WriteString(strings.Repeat("─", boxWidth))
				line3.WriteString("┘")

				if j < len(row)-1 {
					line1.WriteString("  ")
					line2.WriteString("  ")
					line3.WriteString("  ")
				}
			}

			// Center the row
			padding := (60 - line1.Len()) / 2
			if padding < 0 {
				padding = 0
			}
			prefix := strings.Repeat(" ", padding)

			sb.WriteString(prefix + line1.String() + "\n")
			sb.WriteString(prefix + line2.String() + "\n")
			sb.WriteString(prefix + line3.String() + "\n")

			// Add connector if not last row
			if i < len(rows)-1 {
				sb.WriteString(prefix + "       │              │\n")
			}
		}
	}

	// Add external dependencies indicator
	sb.WriteString("                           │\n")
	sb.WriteString("                           ▼\n")
	sb.WriteString("                   [External Services]\n")

	sb.WriteString("```\n")

	return sb.String()
}

// centerPad centers a string within a given width
func centerPad(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	leftPad := (width - len(s)) / 2
	rightPad := width - len(s) - leftPad
	return strings.Repeat(" ", leftPad) + s + strings.Repeat(" ", rightPad)
}

// groupStrings groups strings into chunks of given size
func groupStrings(strs []string, size int) [][]string {
	var result [][]string
	for i := 0; i < len(strs); i += size {
		end := i + size
		if end > len(strs) {
			end = len(strs)
		}
		result = append(result, strs[i:end])
	}
	return result
}
