package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Priyans-hu/argus/pkg/types"
)

// PyProject represents pyproject.toml structure
type PyProject struct {
	Project     ProjectSection     `toml:"project"`
	Tool        ToolSection        `toml:"tool"`
	BuildSystem BuildSystemSection `toml:"build-system"`
}

// ProjectSection represents [project] section
type ProjectSection struct {
	Name                 string                       `toml:"name"`
	Version              string                       `toml:"version"`
	Description          string                       `toml:"description"`
	Authors              []Author                     `toml:"authors"`
	License              string                       `toml:"license"`
	Readme               string                       `toml:"readme"`
	RequiresPython       string                       `toml:"requires-python"`
	Dependencies         []string                     `toml:"dependencies"`
	OptionalDependencies map[string][]string          `toml:"optional-dependencies"`
	Scripts              map[string]string            `toml:"scripts"`
	EntryPoints          map[string]map[string]string `toml:"entry-points"`
}

// Author represents an author entry
type Author struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

// ToolSection represents [tool] section with various tool configs
type ToolSection struct {
	Poetry     PoetrySection     `toml:"poetry"`
	Pytest     PytestSection     `toml:"pytest"`
	Black      BlackSection      `toml:"black"`
	Ruff       RuffSection       `toml:"ruff"`
	Mypy       MypySection       `toml:"mypy"`
	Setuptools SetuptoolsSection `toml:"setuptools"`
	Coverage   CoverageSection   `toml:"coverage"`
	PDM        PDMSection        `toml:"pdm"`
}

// PoetrySection represents [tool.poetry] section
type PoetrySection struct {
	Name            string                       `toml:"name"`
	Version         string                       `toml:"version"`
	Description     string                       `toml:"description"`
	Authors         []string                     `toml:"authors"`
	License         string                       `toml:"license"`
	Readme          string                       `toml:"readme"`
	Packages        []PoetryPackage              `toml:"packages"`
	Dependencies    map[string]interface{}       `toml:"dependencies"`
	DevDependencies map[string]interface{}       `toml:"dev-dependencies"`
	Extras          map[string][]string          `toml:"extras"`
	Scripts         map[string]string            `toml:"scripts"`
	Plugins         map[string]map[string]string `toml:"plugins"`
	Group           map[string]PoetryGroup       `toml:"group"`
}

// PoetryPackage represents a poetry package entry
type PoetryPackage struct {
	Include string `toml:"include"`
	From    string `toml:"from"`
}

// PoetryGroup represents a poetry dependency group
type PoetryGroup struct {
	Dependencies map[string]interface{} `toml:"dependencies"`
}

// PytestSection represents [tool.pytest] section
type PytestSection struct {
	Ini_Options map[string]interface{} `toml:"ini_options"`
}

// BlackSection represents [tool.black] section
type BlackSection struct {
	LineLength    int      `toml:"line-length"`
	TargetVersion []string `toml:"target-version"`
	Include       string   `toml:"include"`
	Exclude       string   `toml:"exclude"`
}

// RuffSection represents [tool.ruff] section
type RuffSection struct {
	LineLength int             `toml:"line-length"`
	Select     []string        `toml:"select"`
	Ignore     []string        `toml:"ignore"`
	FixableAll bool            `toml:"fixable"`
	Lint       RuffLintSection `toml:"lint"` // [tool.ruff.lint] subsection
}

// RuffLintSection represents [tool.ruff.lint] section
type RuffLintSection struct {
	Select []string `toml:"select"`
	Ignore []string `toml:"ignore"`
}

// MypySection represents [tool.mypy] section
type MypySection struct {
	PythonVersion        string   `toml:"python_version"`
	Strict               bool     `toml:"strict"`
	StrictOptional       bool     `toml:"strict_optional"`
	IgnoreMissingImports bool     `toml:"ignore_missing_imports"`
	Plugins              []string `toml:"plugins"`
}

// CoverageSection represents [tool.coverage] section
type CoverageSection struct {
	Run    map[string]interface{} `toml:"run"`
	Report map[string]interface{} `toml:"report"`
}

// PDMSection represents [tool.pdm] section
type PDMSection struct {
	Version      map[string]interface{} `toml:"version"`
	Distribution bool                   `toml:"distribution"`
}

// SetuptoolsSection represents [tool.setuptools] section
type SetuptoolsSection struct {
	Packages   []string          `toml:"packages"`
	PackageDir map[string]string `toml:"package-dir"`
}

// BuildSystemSection represents [build-system] section
type BuildSystemSection struct {
	Requires     []string `toml:"requires"`
	BuildBackend string   `toml:"build-backend"`
}

// PyProjectDetector detects Python project info from pyproject.toml
type PyProjectDetector struct {
	rootPath string
}

// NewPyProjectDetector creates a new pyproject.toml detector
func NewPyProjectDetector(rootPath string) *PyProjectDetector {
	return &PyProjectDetector{rootPath: rootPath}
}

// Detect parses pyproject.toml and returns project info
func (d *PyProjectDetector) Detect() *PyProjectInfo {
	pyprojectPath := filepath.Join(d.rootPath, "pyproject.toml")

	content, err := os.ReadFile(pyprojectPath)
	if err != nil {
		return nil
	}

	var pyproject PyProject
	if _, err := toml.Decode(string(content), &pyproject); err != nil {
		return nil
	}

	info := &PyProjectInfo{
		HasPyProject: true,
	}

	// Extract project info
	if pyproject.Project.Name != "" {
		info.Name = pyproject.Project.Name
		info.Version = pyproject.Project.Version
		info.Description = pyproject.Project.Description
		info.PythonVersion = pyproject.Project.RequiresPython
		info.Dependencies = pyproject.Project.Dependencies
		info.Scripts = pyproject.Project.Scripts
	}

	// Check for Poetry
	if pyproject.Tool.Poetry.Name != "" {
		info.UsesPoetry = true
		info.Name = pyproject.Tool.Poetry.Name
		info.Version = pyproject.Tool.Poetry.Version
		info.Description = pyproject.Tool.Poetry.Description
		info.Scripts = pyproject.Tool.Poetry.Scripts

		// Extract poetry dependencies
		info.Dependencies = extractPoetryDeps(pyproject.Tool.Poetry.Dependencies)
		info.DevDependencies = extractPoetryDeps(pyproject.Tool.Poetry.DevDependencies)

		// Extract group dependencies
		for groupName, group := range pyproject.Tool.Poetry.Group {
			deps := extractPoetryDeps(group.Dependencies)
			if groupName == "dev" || groupName == "test" {
				info.DevDependencies = append(info.DevDependencies, deps...)
			} else {
				info.Dependencies = append(info.Dependencies, deps...)
			}
		}
	}

	// Detect tools
	info.Tools = d.detectTools(pyproject)

	// Detect build backend
	info.BuildBackend = pyproject.BuildSystem.BuildBackend

	return info
}

// PyProjectInfo holds parsed pyproject.toml information
type PyProjectInfo struct {
	HasPyProject    bool
	Name            string
	Version         string
	Description     string
	PythonVersion   string
	Dependencies    []string
	DevDependencies []string
	Scripts         map[string]string
	Tools           []string
	BuildBackend    string
	UsesPoetry      bool
}

// detectTools detects configured development tools
func (d *PyProjectDetector) detectTools(pyproject PyProject) []string {
	var tools []string

	// Check for common tools in [tool.*] section
	if pyproject.Tool.Pytest.Ini_Options != nil {
		tools = append(tools, "pytest")
	}
	if pyproject.Tool.Black.LineLength > 0 || pyproject.Tool.Black.TargetVersion != nil {
		tools = append(tools, "black")
	}
	// Check both [tool.ruff] and [tool.ruff.lint] sections
	if len(pyproject.Tool.Ruff.Select) > 0 || pyproject.Tool.Ruff.LineLength > 0 ||
		len(pyproject.Tool.Ruff.Lint.Select) > 0 {
		tools = append(tools, "ruff")
	}
	// Check for mypy config (various fields indicate presence)
	if pyproject.Tool.Mypy.PythonVersion != "" || pyproject.Tool.Mypy.StrictOptional ||
		pyproject.Tool.Mypy.Strict || len(pyproject.Tool.Mypy.Plugins) > 0 {
		tools = append(tools, "mypy")
	}
	// Check for coverage config
	if pyproject.Tool.Coverage.Run != nil || pyproject.Tool.Coverage.Report != nil {
		tools = append(tools, "coverage")
	}

	// Check build backend for additional tools
	backend := pyproject.BuildSystem.BuildBackend
	switch backend {
	case "poetry.core.masonry.api":
		if !contains(tools, "poetry") {
			tools = append(tools, "poetry")
		}
	case "setuptools.build_meta":
		tools = append(tools, "setuptools")
	case "hatchling.build":
		tools = append(tools, "hatch")
	case "flit_core.buildapi":
		tools = append(tools, "flit")
	case "pdm.pep517.api", "pdm.backend", "pdm-backend":
		tools = append(tools, "pdm")
	}

	return tools
}

// extractPoetryDeps extracts dependency names from poetry's dependency format
func extractPoetryDeps(deps map[string]interface{}) []string {
	var result []string
	for name := range deps {
		if name != "python" { // Skip python version constraint
			result = append(result, name)
		}
	}
	return result
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// DetectPythonPatterns detects Python patterns from pyproject.toml
func (d *PyProjectDetector) DetectPythonPatterns() []types.PatternInfo {
	info := d.Detect()
	if info == nil {
		return nil
	}

	var patterns []types.PatternInfo

	// Add build backend as pattern
	if info.BuildBackend != "" {
		backendName := strings.Split(info.BuildBackend, ".")[0]
		patterns = append(patterns, types.PatternInfo{
			Name:        backendName,
			Category:    "py-build",
			Description: "Python build backend: " + info.BuildBackend,
			FileCount:   1,
			Examples:    []string{"pyproject.toml"},
		})
	}

	// Add tools as patterns
	toolDescriptions := map[string]string{
		"pytest":     "pytest testing framework",
		"black":      "Black code formatter",
		"ruff":       "Ruff fast Python linter",
		"mypy":       "mypy static type checker",
		"poetry":     "Poetry dependency management",
		"setuptools": "setuptools build system",
		"hatch":      "Hatch project manager",
		"flit":       "Flit simple Python packaging",
		"pdm":        "PDM package manager",
	}

	for _, tool := range info.Tools {
		desc := toolDescriptions[tool]
		if desc == "" {
			desc = tool + " Python tool"
		}
		patterns = append(patterns, types.PatternInfo{
			Name:        tool,
			Category:    "py-tool",
			Description: desc,
			FileCount:   1,
			Examples:    []string{"pyproject.toml"},
		})
	}

	// Detect framework dependencies
	frameworkDeps := map[string]string{
		"django":       "Django web framework",
		"flask":        "Flask web framework",
		"fastapi":      "FastAPI web framework",
		"starlette":    "Starlette ASGI framework",
		"celery":       "Celery task queue",
		"sqlalchemy":   "SQLAlchemy ORM",
		"pydantic":     "Pydantic data validation",
		"numpy":        "NumPy numerical computing",
		"pandas":       "Pandas data analysis",
		"tensorflow":   "TensorFlow ML framework",
		"torch":        "PyTorch ML framework",
		"transformers": "Hugging Face Transformers",
		"langchain":    "LangChain LLM framework",
	}

	allDeps := append(info.Dependencies, info.DevDependencies...)
	for _, dep := range allDeps {
		depLower := strings.ToLower(dep)
		if desc, ok := frameworkDeps[depLower]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        dep,
				Category:    "py-framework",
				Description: desc,
				FileCount:   1,
				Examples:    []string{"pyproject.toml"},
			})
		}
	}

	return patterns
}
