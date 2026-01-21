package detector

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// DevelopmentDetector detects development setup information
type DevelopmentDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewDevelopmentDetector creates a new development detector
func NewDevelopmentDetector(rootPath string, files []types.FileInfo) *DevelopmentDetector {
	return &DevelopmentDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect extracts development setup information
func (d *DevelopmentDetector) Detect() *types.DevelopmentInfo {
	info := &types.DevelopmentInfo{}

	info.Prerequisites = d.detectPrerequisites()
	info.SetupSteps = d.detectSetupSteps()
	info.GitHooks = d.detectGitHooks()

	// Return nil if nothing detected
	if len(info.Prerequisites) == 0 && len(info.SetupSteps) == 0 && len(info.GitHooks) == 0 {
		return nil
	}

	return info
}

// detectPrerequisites detects required tools and runtimes
func (d *DevelopmentDetector) detectPrerequisites() []types.Prerequisite {
	var prereqs []types.Prerequisite

	// Check Go version from go.mod
	if goVer := d.detectGoVersion(); goVer != "" {
		prereqs = append(prereqs, types.Prerequisite{
			Name:    "Go",
			Version: goVer,
		})
	}

	// Check Node version from package.json or .nvmrc
	if nodeVer := d.detectNodeVersion(); nodeVer != "" {
		prereqs = append(prereqs, types.Prerequisite{
			Name:    "Node.js",
			Version: nodeVer,
		})
	}

	// Check Python version from pyproject.toml or .python-version
	if pyVer := d.detectPythonVersion(); pyVer != "" {
		prereqs = append(prereqs, types.Prerequisite{
			Name:    "Python",
			Version: pyVer,
		})
	}

	// Check for Docker
	if d.hasFile("Dockerfile") || d.hasFile("docker-compose.yml") || d.hasFile("docker-compose.yaml") {
		prereqs = append(prereqs, types.Prerequisite{
			Name: "Docker",
		})
	}

	// Check .tool-versions (asdf)
	prereqs = append(prereqs, d.detectFromToolVersions()...)

	return prereqs
}

// detectGoVersion extracts Go version from go.mod
func (d *DevelopmentDetector) detectGoVersion() string {
	content, err := os.ReadFile(filepath.Join(d.rootPath, "go.mod"))
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`(?m)^go\s+(\d+\.\d+)`)
	if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
		return matches[1] + "+"
	}
	return ""
}

// detectNodeVersion extracts Node version from package.json or .nvmrc
func (d *DevelopmentDetector) detectNodeVersion() string {
	// Try .nvmrc first
	if content, err := os.ReadFile(filepath.Join(d.rootPath, ".nvmrc")); err == nil {
		ver := strings.TrimSpace(string(content))
		if ver != "" {
			return ver
		}
	}

	// Try package.json engines
	pkgPath := filepath.Join(d.rootPath, "package.json")
	content, err := os.ReadFile(pkgPath)
	if err != nil {
		return ""
	}

	var pkg struct {
		Engines struct {
			Node string `json:"node"`
		} `json:"engines"`
	}

	if json.Unmarshal(content, &pkg) == nil && pkg.Engines.Node != "" {
		return pkg.Engines.Node
	}

	// If package.json exists but no version, assume Node is needed
	if d.hasFile("package.json") {
		return "18+"
	}

	return ""
}

// detectPythonVersion extracts Python version
func (d *DevelopmentDetector) detectPythonVersion() string {
	// Try .python-version
	if content, err := os.ReadFile(filepath.Join(d.rootPath, ".python-version")); err == nil {
		ver := strings.TrimSpace(string(content))
		if ver != "" {
			return ver
		}
	}

	// Try pyproject.toml
	content, err := os.ReadFile(filepath.Join(d.rootPath, "pyproject.toml"))
	if err != nil {
		return ""
	}

	re := regexp.MustCompile(`requires-python\s*=\s*["']([^"']+)["']`)
	if matches := re.FindStringSubmatch(string(content)); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// detectFromToolVersions parses .tool-versions (asdf)
func (d *DevelopmentDetector) detectFromToolVersions() []types.Prerequisite {
	var prereqs []types.Prerequisite

	content, err := os.ReadFile(filepath.Join(d.rootPath, ".tool-versions"))
	if err != nil {
		return prereqs
	}

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := capitalizeFirst(parts[0])
			version := parts[1]

			// Skip if already detected
			skip := false
			for _, p := range prereqs {
				if strings.EqualFold(p.Name, name) {
					skip = true
					break
				}
			}
			if !skip {
				prereqs = append(prereqs, types.Prerequisite{
					Name:    name,
					Version: version,
				})
			}
		}
	}

	return prereqs
}

// detectSetupSteps detects setup instructions
func (d *DevelopmentDetector) detectSetupSteps() []types.SetupStep {
	var steps []types.SetupStep

	// Detect from Makefile
	steps = append(steps, d.detectMakefileSetup()...)

	// If no Makefile setup, infer from project type
	if len(steps) == 0 {
		steps = d.inferSetupSteps()
	}

	return steps
}

// detectMakefileSetup extracts setup steps from Makefile
func (d *DevelopmentDetector) detectMakefileSetup() []types.SetupStep {
	var steps []types.SetupStep

	content, err := os.ReadFile(filepath.Join(d.rootPath, "Makefile"))
	if err != nil {
		return steps
	}

	lines := strings.Split(string(content), "\n")
	setupTargets := []string{"setup", "install", "deps", "init", "bootstrap"}

	targetRe := regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*):\s*`)
	commentRe := regexp.MustCompile(`^##\s*(.+)$`)

	var lastComment string
	for _, line := range lines {
		// Check for doc comment
		if match := commentRe.FindStringSubmatch(line); len(match) > 1 {
			lastComment = match[1]
			continue
		}

		// Check for target
		if match := targetRe.FindStringSubmatch(line); len(match) > 1 {
			target := match[1]
			for _, setupTarget := range setupTargets {
				if strings.EqualFold(target, setupTarget) {
					desc := lastComment
					if desc == "" {
						desc = "Run " + target
					}
					steps = append(steps, types.SetupStep{
						Command:     "make " + target,
						Description: desc,
					})
				}
			}
			lastComment = ""
		}
	}

	return steps
}

// inferSetupSteps infers setup steps from project type
func (d *DevelopmentDetector) inferSetupSteps() []types.SetupStep {
	var steps []types.SetupStep

	// Node.js project
	if d.hasFile("package.json") {
		pkgManager := "npm"
		if d.hasFile("yarn.lock") {
			pkgManager = "yarn"
		} else if d.hasFile("pnpm-lock.yaml") {
			pkgManager = "pnpm"
		} else if d.hasFile("bun.lockb") {
			pkgManager = "bun"
		}

		steps = append(steps, types.SetupStep{
			Command:     pkgManager + " install",
			Description: "Install dependencies",
		})
	}

	// Go project
	if d.hasFile("go.mod") {
		steps = append(steps, types.SetupStep{
			Command:     "go mod download",
			Description: "Download Go dependencies",
		})
	}

	// Python project
	if d.hasFile("requirements.txt") {
		steps = append(steps, types.SetupStep{
			Command:     "pip install -r requirements.txt",
			Description: "Install Python dependencies",
		})
	} else if d.hasFile("pyproject.toml") {
		steps = append(steps, types.SetupStep{
			Command:     "pip install -e .",
			Description: "Install Python package in editable mode",
		})
	}

	// Environment setup
	if d.hasFile(".env.example") {
		steps = append(steps, types.SetupStep{
			Command:     "cp .env.example .env",
			Description: "Copy environment template",
		})
	}

	return steps
}

// detectGitHooks detects git hook configuration
func (d *DevelopmentDetector) detectGitHooks() []types.GitHook {
	var hooks []types.GitHook

	// Check .githooks directory
	if entries, err := os.ReadDir(filepath.Join(d.rootPath, ".githooks")); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				actions := d.parseGitHookFile(filepath.Join(d.rootPath, ".githooks", entry.Name()))
				hooks = append(hooks, types.GitHook{
					Name:    entry.Name(),
					Actions: actions,
				})
			}
		}
	}

	// Check .husky directory
	if entries, err := os.ReadDir(filepath.Join(d.rootPath, ".husky")); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if !entry.IsDir() && !strings.HasPrefix(name, "_") && !strings.HasPrefix(name, ".") {
				actions := d.parseGitHookFile(filepath.Join(d.rootPath, ".husky", name))
				hooks = append(hooks, types.GitHook{
					Name:    name,
					Actions: actions,
				})
			}
		}
	}

	// Check for lefthook
	if d.hasFile(".lefthook.yml") || d.hasFile("lefthook.yml") {
		hooks = append(hooks, types.GitHook{
			Name:    "lefthook",
			Actions: []string{"See lefthook.yml for configuration"},
		})
	}

	return hooks
}

// parseGitHookFile extracts actions from a git hook file
func (d *DevelopmentDetector) parseGitHookFile(path string) []string {
	var actions []string

	content, err := os.ReadFile(path)
	if err != nil {
		return actions
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines, comments, and shebang
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract meaningful commands
		if strings.Contains(line, "fmt") || strings.Contains(line, "format") {
			actions = append(actions, "Format code")
		}
		if strings.Contains(line, "lint") {
			actions = append(actions, "Run linter")
		}
		if strings.Contains(line, "test") {
			actions = append(actions, "Run tests")
		}
		if strings.Contains(line, "goimports") {
			actions = append(actions, "Organize imports")
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, a := range actions {
		if !seen[a] {
			seen[a] = true
			unique = append(unique, a)
		}
	}

	return unique
}

// hasFile checks if a file exists in the root directory
func (d *DevelopmentDetector) hasFile(name string) bool {
	_, err := os.Stat(filepath.Join(d.rootPath, name))
	return err == nil
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
