package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// MonorepoDetector detects monorepo structure and workspace configuration
type MonorepoDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewMonorepoDetector creates a new monorepo detector
func NewMonorepoDetector(rootPath string, files []types.FileInfo) *MonorepoDetector {
	return &MonorepoDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes the project for monorepo patterns
func (d *MonorepoDetector) Detect() *types.MonorepoInfo {
	info := &types.MonorepoInfo{}

	// Check for monorepo tools
	d.detectMonorepoTools(info)

	// Check for workspace structure
	d.detectWorkspaces(info)

	// Check for apps/packages directories
	d.detectAppPackageStructure(info)

	// Only return if it looks like a monorepo
	if !info.IsMonorepo {
		return nil
	}

	return info
}

// detectMonorepoTools checks for monorepo tooling
func (d *MonorepoDetector) detectMonorepoTools(info *types.MonorepoInfo) {
	// Check for turbo.json
	if _, err := os.Stat(filepath.Join(d.rootPath, "turbo.json")); err == nil {
		info.IsMonorepo = true
		info.Tool = "Turborepo"
	}

	// Check for lerna.json
	if _, err := os.Stat(filepath.Join(d.rootPath, "lerna.json")); err == nil {
		info.IsMonorepo = true
		if info.Tool == "" {
			info.Tool = "Lerna"
		}
	}

	// Check for nx.json
	if _, err := os.Stat(filepath.Join(d.rootPath, "nx.json")); err == nil {
		info.IsMonorepo = true
		if info.Tool == "" {
			info.Tool = "Nx"
		}
	}

	// Check for pnpm-workspace.yaml
	if _, err := os.Stat(filepath.Join(d.rootPath, "pnpm-workspace.yaml")); err == nil {
		info.IsMonorepo = true
		info.PackageManager = "pnpm"
	}
}

// detectWorkspaces checks package.json for workspace configuration
func (d *MonorepoDetector) detectWorkspaces(info *types.MonorepoInfo) {
	pkgPath := filepath.Join(d.rootPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return
	}

	var pkg struct {
		Workspaces interface{} `json:"workspaces"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	if pkg.Workspaces != nil {
		info.IsMonorepo = true

		// Parse workspaces (can be array or object with packages key)
		switch ws := pkg.Workspaces.(type) {
		case []interface{}:
			for _, w := range ws {
				if s, ok := w.(string); ok {
					info.WorkspacePaths = append(info.WorkspacePaths, s)
				}
			}
		case map[string]interface{}:
			if packages, ok := ws["packages"].([]interface{}); ok {
				for _, w := range packages {
					if s, ok := w.(string); ok {
						info.WorkspacePaths = append(info.WorkspacePaths, s)
					}
				}
			}
		}
	}
}

// detectAppPackageStructure detects apps/ and packages/ structure
func (d *MonorepoDetector) detectAppPackageStructure(info *types.MonorepoInfo) {
	// Check for common monorepo directory structures
	monoDirs := []struct {
		dir     string
		purpose string
	}{
		{"apps", "Applications"},
		{"packages", "Shared packages"},
		{"libs", "Libraries"},
		{"services", "Microservices"},
		{"tools", "Development tools"},
	}

	for _, md := range monoDirs {
		dirPath := filepath.Join(d.rootPath, md.dir)
		if stat, err := os.Stat(dirPath); err == nil && stat.IsDir() {
			// Check if it has subdirectories (actual packages)
			entries, err := os.ReadDir(dirPath)
			if err != nil {
				continue
			}

			var subDirs []string
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					subDirs = append(subDirs, entry.Name())
				}
			}

			if len(subDirs) > 0 {
				info.IsMonorepo = true
				pkg := types.WorkspacePackage{
					Name:        md.dir,
					Path:        md.dir,
					Description: md.purpose,
					SubPackages: subDirs,
				}
				info.Packages = append(info.Packages, pkg)
			}
		}
	}
}
