package detector

import (
	"os"
	"path/filepath"

	"github.com/Priyans-hu/argus/pkg/types"
)

// ConfigDetector detects configuration files in the project
type ConfigDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewConfigDetector creates a new config detector
func NewConfigDetector(rootPath string, files []types.FileInfo) *ConfigDetector {
	return &ConfigDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// configFileMapping maps file names to their type and purpose
var configFileMapping = map[string]struct {
	Type    string
	Purpose string
}{
	// Argus
	".argus.yaml": {"Argus", "Code analyzer configuration"},

	// Go
	"go.mod":         {"Go Modules", "Go dependencies and module path"},
	".golangci.yml":  {"Linter", "Go linter configuration"},
	".golangci.yaml": {"Linter", "Go linter configuration"},

	// Node.js / JavaScript / TypeScript
	"package.json":        {"npm", "Project metadata and dependencies"},
	"tsconfig.json":       {"TypeScript", "TypeScript compiler options"},
	".eslintrc.json":      {"Linter", "ESLint configuration"},
	".eslintrc.js":        {"Linter", "ESLint configuration"},
	".eslintrc.yml":       {"Linter", "ESLint configuration"},
	"eslint.config.js":    {"Linter", "ESLint flat config"},
	"eslint.config.mjs":   {"Linter", "ESLint flat config"},
	".prettierrc":         {"Formatter", "Prettier configuration"},
	".prettierrc.json":    {"Formatter", "Prettier configuration"},
	".prettierrc.js":      {"Formatter", "Prettier configuration"},
	"prettier.config.js":  {"Formatter", "Prettier configuration"},
	"jest.config.js":      {"Testing", "Jest test configuration"},
	"jest.config.ts":      {"Testing", "Jest test configuration"},
	"vitest.config.ts":    {"Testing", "Vitest test configuration"},
	"vitest.config.js":    {"Testing", "Vitest test configuration"},
	"vite.config.ts":      {"Build", "Vite bundler configuration"},
	"vite.config.js":      {"Build", "Vite bundler configuration"},
	"next.config.js":      {"Framework", "Next.js configuration"},
	"next.config.mjs":     {"Framework", "Next.js configuration"},
	"nuxt.config.ts":      {"Framework", "Nuxt configuration"},
	"tailwind.config.js":  {"Styling", "Tailwind CSS configuration"},
	"tailwind.config.ts":  {"Styling", "Tailwind CSS configuration"},
	"postcss.config.js":   {"Styling", "PostCSS configuration"},
	"webpack.config.js":   {"Build", "Webpack bundler configuration"},
	"rollup.config.js":    {"Build", "Rollup bundler configuration"},
	"turbo.json":          {"Monorepo", "Turborepo configuration"},
	"lerna.json":          {"Monorepo", "Lerna configuration"},
	"nx.json":             {"Monorepo", "Nx workspace configuration"},
	"pnpm-workspace.yaml": {"Monorepo", "pnpm workspace configuration"},

	// Python
	"pyproject.toml":   {"Python", "Python project configuration"},
	"setup.py":         {"Python", "Python package setup"},
	"setup.cfg":        {"Python", "Python package configuration"},
	"requirements.txt": {"Python", "Python dependencies"},
	"tox.ini":          {"Testing", "Tox test configuration"},
	"pytest.ini":       {"Testing", "Pytest configuration"},
	".flake8":          {"Linter", "Flake8 linter configuration"},
	"mypy.ini":         {"Linter", "Mypy type checker configuration"},
	".mypy.ini":        {"Linter", "Mypy type checker configuration"},
	"ruff.toml":        {"Linter", "Ruff linter configuration"},

	// Rust
	"Cargo.toml": {"Rust", "Rust package configuration"},

	// Docker
	"Dockerfile":          {"Docker", "Container image definition"},
	"docker-compose.yml":  {"Docker", "Docker Compose services"},
	"docker-compose.yaml": {"Docker", "Docker Compose services"},

	// CI/CD
	".travis.yml":    {"CI/CD", "Travis CI configuration"},
	".gitlab-ci.yml": {"CI/CD", "GitLab CI configuration"},
	"Jenkinsfile":    {"CI/CD", "Jenkins pipeline"},

	// Build
	"Makefile":         {"Build", "Make build automation"},
	".goreleaser.yml":  {"Release", "GoReleaser configuration"},
	".goreleaser.yaml": {"Release", "GoReleaser configuration"},

	// Environment
	".env.example":  {"Environment", "Environment variables template"},
	".env.sample":   {"Environment", "Environment variables template"},
	".env.template": {"Environment", "Environment variables template"},

	// Git hooks
	".lefthook.yml":           {"Git Hooks", "Lefthook configuration"},
	"lefthook.yml":            {"Git Hooks", "Lefthook configuration"},
	".pre-commit-config.yaml": {"Git Hooks", "Pre-commit hooks configuration"},

	// Editor
	".editorconfig": {"Editor", "Editor configuration"},

	// Other
	".codecov.yml":  {"Coverage", "Codecov configuration"},
	"codecov.yml":   {"Coverage", "Codecov configuration"},
	"renovate.json": {"Dependencies", "Renovate bot configuration"},
	".nvmrc":        {"Node", "Node version specification"},
}

// Detect finds and documents configuration files in the project
func (d *ConfigDetector) Detect() []types.ConfigFileInfo {
	var configs []types.ConfigFileInfo

	// Check each known config file
	for filename, info := range configFileMapping {
		if d.fileExists(filename) {
			configs = append(configs, types.ConfigFileInfo{
				Path:    filename,
				Type:    info.Type,
				Purpose: info.Purpose,
			})
		}
	}

	// Check for GitHub workflows directory
	if d.dirExists(".github/workflows") {
		configs = append(configs, types.ConfigFileInfo{
			Path:    ".github/workflows/",
			Type:    "CI/CD",
			Purpose: "GitHub Actions workflows",
		})
	}

	// Check for .github/dependabot.yml
	if d.fileExists(".github/dependabot.yml") {
		configs = append(configs, types.ConfigFileInfo{
			Path:    ".github/dependabot.yml",
			Type:    "Dependencies",
			Purpose: "Dependabot configuration",
		})
	}

	return configs
}

// fileExists checks if a file exists in the root directory
func (d *ConfigDetector) fileExists(name string) bool {
	info, err := os.Stat(filepath.Join(d.rootPath, name))
	return err == nil && !info.IsDir()
}

// dirExists checks if a directory exists in the root directory
func (d *ConfigDetector) dirExists(name string) bool {
	info, err := os.Stat(filepath.Join(d.rootPath, name))
	return err == nil && info.IsDir()
}
