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

	// Detect additional commands from pyproject.toml (Python)
	pyprojectDetector := detector.NewPyProjectDetector(absPath)
	if pyInfo := pyprojectDetector.Detect(); pyInfo != nil && pyInfo.HasPyProject {
		analysis.Commands = append(analysis.Commands, detectPyProjectCommands(pyInfo)...)
	}

	// Detect additional commands from Cargo.toml (Rust)
	cargoDetector := detector.NewCargoDetector(absPath)
	if cargoInfo := cargoDetector.Detect(); cargoInfo != nil && cargoInfo.HasCargo {
		// Replace basic cargo commands with enhanced ones
		analysis.Commands = filterNonCargoCommands(analysis.Commands)
		analysis.Commands = append(analysis.Commands, cargoDetector.DetectCargoCommands()...)
	}

	// Detect dependencies
	analysis.Dependencies = a.detectDependencies(absPath)

	// Detect conventions
	conventionDetector := detector.NewConventionDetector(absPath, files)
	conventions, err := conventionDetector.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect conventions: %w", err)
	}
	analysis.Conventions = conventions

	// Detect patterns (branch naming, comments, logging, error handling, architecture)
	patternDetector := detector.NewPatternDetector(absPath, files)
	patterns, err := patternDetector.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect patterns: %w", err)
	}
	analysis.Conventions = append(analysis.Conventions, patterns...)

	// Detect framework-specific patterns
	frameworkDetector := detector.NewFrameworkDetector(absPath, files)
	frameworkPatterns, err := frameworkDetector.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect framework patterns: %w", err)
	}
	analysis.Conventions = append(analysis.Conventions, frameworkPatterns...)

	// Detect API endpoints
	endpointDetector := detector.NewEndpointDetector(absPath, files)
	endpoints, err := endpointDetector.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect endpoints: %w", err)
	}
	analysis.Endpoints = endpoints

	// Parse README for project overview
	readmeDetector := detector.NewReadmeDetector(absPath)
	analysis.ReadmeContent = readmeDetector.Detect()

	// Detect monorepo structure
	monorepoDetector := detector.NewMonorepoDetector(absPath, files)
	analysis.MonorepoInfo = monorepoDetector.Detect()

	// Deep code pattern analysis
	codePatternDetector := detector.NewCodePatternDetector(absPath, files)
	analysis.CodePatterns = codePatternDetector.Detect()

	// Detect git conventions (commit messages, branch naming) - using go-git library
	gitDetector := detector.NewGitDetectorGoGit(absPath)
	analysis.GitConventions = gitDetector.Detect()

	// Detect architecture patterns
	archDetector := detector.NewArchitectureDetector(absPath, files)
	analysis.ArchitectureInfo = archDetector.Detect()

	// Detect development setup info
	devDetector := detector.NewDevelopmentDetector(absPath, files)
	analysis.DevelopmentInfo = devDetector.Detect()

	// Detect config files
	configDetector := detector.NewConfigDetector(absPath, files)
	analysis.ConfigFiles = configDetector.Detect()

	// Detect CLI info (if applicable)
	cliDetector := detector.NewCLIDetector(absPath, files, &analysis.TechStack)
	analysis.CLIInfo = cliDetector.Detect()

	// Detect project-specific tools
	toolsDetector := detector.NewProjectToolsDetector(
		absPath,
		files,
		analysis.ProjectName,
		&analysis.TechStack,
		analysis.ReadmeContent,
		analysis.CLIInfo,
	)
	analysis.ProjectTools = toolsDetector.Detect()

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
				// Skip indirect dependencies
				if contains(line, "// indirect") {
					continue
				}

				parts := splitFields(line)
				if len(parts) >= 2 {
					pkgName := parts[0]

					// Filter out internal/vendor packages
					if isInternalPackage(pkgName) {
						continue
					}

					deps = append(deps, types.Dependency{
						Name:    pkgName,
						Version: parts[1],
						Type:    "runtime",
					})
				}
			}
		}
	}

	return deps
}

// isInternalPackage checks if a package path is internal/vendor
func isInternalPackage(pkg string) bool {
	// Skip internal subpackages
	if contains(pkg, "/internal/") {
		return true
	}
	// Skip service-specific subpackages (AWS SDK pattern)
	if contains(pkg, "/service/internal/") {
		return true
	}
	// Skip feature subpackages
	if contains(pkg, "/feature/") {
		return true
	}
	return false
}

// contains checks if s contains substr (simple implementation)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
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

// detectPyProjectCommands extracts commands from pyproject.toml
func detectPyProjectCommands(info *detector.PyProjectInfo) []types.Command {
	var commands []types.Command

	// Add script commands
	for name, cmd := range info.Scripts {
		commands = append(commands, types.Command{
			Name:        name,
			Command:     cmd,
			Description: "Script from pyproject.toml",
		})
	}

	// Add tool-specific commands
	for _, tool := range info.Tools {
		switch tool {
		case "pytest":
			commands = append(commands, types.Command{
				Name:        "pytest",
				Description: "Run tests with pytest",
			})
			commands = append(commands, types.Command{
				Name:        "pytest -v",
				Description: "Run tests with verbose output",
			})
		case "black":
			commands = append(commands, types.Command{
				Name:        "black .",
				Description: "Format code with Black",
			})
		case "ruff":
			commands = append(commands, types.Command{
				Name:        "ruff check .",
				Description: "Lint code with Ruff",
			})
			commands = append(commands, types.Command{
				Name:        "ruff format .",
				Description: "Format code with Ruff",
			})
		case "mypy":
			commands = append(commands, types.Command{
				Name:        "mypy .",
				Description: "Type check with mypy",
			})
		case "poetry":
			commands = append(commands, types.Command{
				Name:        "poetry install",
				Description: "Install dependencies with Poetry",
			})
			commands = append(commands, types.Command{
				Name:        "poetry shell",
				Description: "Activate Poetry virtual environment",
			})
		case "pdm":
			commands = append(commands, types.Command{
				Name:        "pdm install",
				Description: "Install dependencies with PDM",
			})
		case "hatch":
			commands = append(commands, types.Command{
				Name:        "hatch run",
				Description: "Run commands in Hatch environment",
			})
		case "coverage":
			commands = append(commands, types.Command{
				Name:        "coverage run -m pytest",
				Description: "Run tests with coverage",
			})
			commands = append(commands, types.Command{
				Name:        "coverage report",
				Description: "Show coverage report",
			})
		}
	}

	return commands
}

// filterNonCargoCommands removes basic cargo commands to be replaced with enhanced ones
func filterNonCargoCommands(commands []types.Command) []types.Command {
	var filtered []types.Command
	for _, cmd := range commands {
		// Keep non-cargo commands
		if !isBasicCargoCommand(cmd.Name) {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

// isBasicCargoCommand checks if a command is a basic cargo command
func isBasicCargoCommand(name string) bool {
	basicCargo := []string{
		"cargo build",
		"cargo build --release",
		"cargo test",
		"cargo fmt",
		"cargo clippy",
	}
	for _, basic := range basicCargo {
		if name == basic {
			return true
		}
	}
	return false
}
