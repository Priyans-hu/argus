package generator

import (
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// GeneratorContext holds extracted contextual data from Analysis
// for use by context-aware generators
type GeneratorContext struct {
	ProjectName string

	// File examples by category
	TestFiles   []string          // _test.go, .test.ts, .spec.ts files
	ConfigFiles map[string]string // type -> path (eslint, golangci, jest, etc.)
	EntryPoints []string          // main.go, index.ts

	// Pattern examples with files
	TestingPatterns map[string][]string // pattern name -> example files
	AuthPatterns    map[string][]string
	APIPatterns     map[string][]string
	ErrorPatterns   map[string][]string

	// Git conventions
	CommitStyle  string
	BranchFormat string

	// Classified commands
	TestCommand   string
	LintCommand   string
	BuildCommand  string
	FormatCommand string

	// Architecture info
	EntryPoint string
	Layers     []types.ArchitectureLayer
	Diagram    string
}

// BuildContext extracts relevant context from Analysis for generators
func BuildContext(analysis *types.Analysis) *GeneratorContext {
	ctx := &GeneratorContext{
		ProjectName:     analysis.ProjectName,
		ConfigFiles:     make(map[string]string),
		TestingPatterns: make(map[string][]string),
		AuthPatterns:    make(map[string][]string),
		APIPatterns:     make(map[string][]string),
		ErrorPatterns:   make(map[string][]string),
	}

	// Extract test files from structure
	ctx.TestFiles = findTestFiles(analysis)

	// Extract config files from KeyFiles
	ctx.ConfigFiles = findConfigFiles(analysis)

	// Extract entry points
	ctx.EntryPoints = findEntryPoints(analysis)

	// Extract pattern examples
	if analysis.CodePatterns != nil {
		ctx.TestingPatterns = extractPatternExamples(analysis.CodePatterns.Testing)
		ctx.AuthPatterns = extractPatternExamples(analysis.CodePatterns.Authentication)
		ctx.APIPatterns = extractPatternExamples(analysis.CodePatterns.APIPatterns)
	}

	// Extract error handling patterns from conventions
	ctx.ErrorPatterns = findErrorPatterns(analysis)

	// Extract git conventions
	if analysis.GitConventions != nil {
		if cc := analysis.GitConventions.CommitConvention; cc != nil {
			ctx.CommitStyle = cc.Style
		}
		if bc := analysis.GitConventions.BranchConvention; bc != nil {
			ctx.BranchFormat = bc.Format
		}
	}

	// Classify commands
	ctx.classifyCommands(analysis.Commands)

	// Extract architecture info
	if analysis.ArchitectureInfo != nil {
		ctx.EntryPoint = analysis.ArchitectureInfo.EntryPoint
		ctx.Layers = analysis.ArchitectureInfo.Layers
		ctx.Diagram = analysis.ArchitectureInfo.Diagram
	}

	return ctx
}

// classifyCommands categorizes commands by type
func (ctx *GeneratorContext) classifyCommands(commands []types.Command) {
	for _, cmd := range commands {
		cmdType := classifyCommand(cmd.Name, cmd.Command)
		cmdStr := cmd.Command
		if cmdStr == "" {
			cmdStr = cmd.Name
		}

		switch cmdType {
		case "test":
			if ctx.TestCommand == "" {
				ctx.TestCommand = cmdStr
			}
		case "lint":
			if ctx.LintCommand == "" {
				ctx.LintCommand = cmdStr
			}
		case "build":
			if ctx.BuildCommand == "" {
				ctx.BuildCommand = cmdStr
			}
		case "format":
			if ctx.FormatCommand == "" {
				ctx.FormatCommand = cmdStr
			}
		}
	}
}

// HasTestingContext returns true if testing context is available
func (ctx *GeneratorContext) HasTestingContext() bool {
	return len(ctx.TestFiles) > 0 || len(ctx.TestingPatterns) > 0
}

// HasLintingContext returns true if linting config is available
func (ctx *GeneratorContext) HasLintingContext() bool {
	_, hasESLint := ctx.ConfigFiles["eslint"]
	_, hasGoLint := ctx.ConfigFiles["golangci"]
	_, hasPylint := ctx.ConfigFiles["pylint"]
	return hasESLint || hasGoLint || hasPylint || ctx.LintCommand != ""
}

// HasArchitectureContext returns true if architecture info is available
func (ctx *GeneratorContext) HasArchitectureContext() bool {
	return len(ctx.Layers) > 0 || ctx.EntryPoint != ""
}

// GetTestConfig returns the test config file path if available
func (ctx *GeneratorContext) GetTestConfig() string {
	if cfg, ok := ctx.ConfigFiles["jest"]; ok {
		return cfg
	}
	if cfg, ok := ctx.ConfigFiles["vitest"]; ok {
		return cfg
	}
	if cfg, ok := ctx.ConfigFiles["pytest"]; ok {
		return cfg
	}
	return ""
}

// GetLintConfig returns the lint config file path if available
func (ctx *GeneratorContext) GetLintConfig() string {
	if cfg, ok := ctx.ConfigFiles["eslint"]; ok {
		return cfg
	}
	if cfg, ok := ctx.ConfigFiles["golangci"]; ok {
		return cfg
	}
	if cfg, ok := ctx.ConfigFiles["pylint"]; ok {
		return cfg
	}
	if cfg, ok := ctx.ConfigFiles["pyproject"]; ok {
		return cfg
	}
	return ""
}

// Helper functions

// findTestFiles extracts test file paths from the analysis
func findTestFiles(analysis *types.Analysis) []string {
	var testFiles []string

	// Check CodePatterns.Testing for example files
	if analysis.CodePatterns != nil {
		for _, pattern := range analysis.CodePatterns.Testing {
			testFiles = append(testFiles, pattern.Examples...)
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	unique := []string{}
	for _, f := range testFiles {
		if !seen[f] {
			seen[f] = true
			unique = append(unique, f)
		}
	}

	// Limit to 5 examples
	if len(unique) > 5 {
		unique = unique[:5]
	}

	return unique
}

// findConfigFiles extracts config file paths from KeyFiles
func findConfigFiles(analysis *types.Analysis) map[string]string {
	configs := make(map[string]string)

	configPatterns := map[string][]string{
		"jest":      {"jest.config.js", "jest.config.ts", "jest.config.mjs", "jest.config.cjs"},
		"vitest":    {"vitest.config.ts", "vitest.config.js", "vitest.config.mjs"},
		"eslint":    {".eslintrc", ".eslintrc.js", ".eslintrc.json", ".eslintrc.yaml", "eslint.config.js", "eslint.config.mjs"},
		"prettier":  {".prettierrc", ".prettierrc.js", ".prettierrc.json", "prettier.config.js"},
		"golangci":  {".golangci.yml", ".golangci.yaml", ".golangci.toml"},
		"tsconfig":  {"tsconfig.json"},
		"pytest":    {"pytest.ini", "pyproject.toml", "setup.cfg"},
		"pyproject": {"pyproject.toml"},
		"pylint":    {".pylintrc", "pylintrc", "pyproject.toml"},
	}

	// Check KeyFiles
	for _, kf := range analysis.KeyFiles {
		fileName := filepath.Base(kf.Path)
		for configType, patterns := range configPatterns {
			for _, pattern := range patterns {
				if fileName == pattern {
					if _, exists := configs[configType]; !exists {
						configs[configType] = kf.Path
					}
				}
			}
		}
	}

	// Also check root files
	for _, rf := range analysis.Structure.RootFiles {
		fileName := filepath.Base(rf)
		for configType, patterns := range configPatterns {
			for _, pattern := range patterns {
				if fileName == pattern {
					if _, exists := configs[configType]; !exists {
						configs[configType] = rf
					}
				}
			}
		}
	}

	return configs
}

// findEntryPoints extracts entry point files from KeyFiles
func findEntryPoints(analysis *types.Analysis) []string {
	var entryPoints []string

	entryPatterns := []string{"main.go", "index.ts", "index.js", "app.ts", "app.js", "main.py", "__main__.py", "main.rs", "lib.rs"}

	// Check KeyFiles for entry points
	for _, kf := range analysis.KeyFiles {
		fileName := filepath.Base(kf.Path)
		for _, pattern := range entryPatterns {
			if fileName == pattern {
				entryPoints = append(entryPoints, kf.Path)
				break
			}
		}
	}

	// Check architecture entry point
	if analysis.ArchitectureInfo != nil && analysis.ArchitectureInfo.EntryPoint != "" {
		// Add if not already present
		found := false
		for _, ep := range entryPoints {
			if ep == analysis.ArchitectureInfo.EntryPoint {
				found = true
				break
			}
		}
		if !found {
			entryPoints = append(entryPoints, analysis.ArchitectureInfo.EntryPoint)
		}
	}

	return entryPoints
}

// extractPatternExamples converts PatternInfo to pattern name -> examples map
func extractPatternExamples(patterns []types.PatternInfo) map[string][]string {
	result := make(map[string][]string)
	for _, p := range patterns {
		if len(p.Examples) > 0 {
			result[p.Name] = p.Examples
		}
	}
	return result
}

// findErrorPatterns extracts error handling patterns from conventions
func findErrorPatterns(analysis *types.Analysis) map[string][]string {
	patterns := make(map[string][]string)

	// Look for error-handling conventions
	for _, conv := range analysis.Conventions {
		if conv.Category == "error-handling" && conv.Example != "" {
			patterns["error-handling"] = append(patterns["error-handling"], conv.Example)
		}
	}

	// Check CodePatterns for API error patterns
	if analysis.CodePatterns != nil {
		for _, p := range analysis.CodePatterns.APIPatterns {
			if strings.Contains(strings.ToLower(p.Name), "error") {
				patterns[p.Name] = p.Examples
			}
		}
	}

	return patterns
}
