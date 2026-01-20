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
	for _, f := range d.files {
		if f.IsDir {
			dirs[strings.ToLower(f.Name)] = true
		}
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
		{"internal", "Private packages"},
		{"pkg", "Public packages"},
		{"domain", "Business logic"},
		{"services", "Service layer"},
		{"handlers", "Request handlers"},
		{"models", "Data models"},
		{"repository", "Data access"},
		{"config", "Configuration"},
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
				if pkg != "" && pkg != layer {
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

	for _, entry := range entries {
		if entry.IsDir() {
			mainPath := filepath.Join(cmdDir, entry.Name(), "main.go")
			if _, err := os.Stat(mainPath); err == nil {
				return "cmd/" + entry.Name() + "/main.go"
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
