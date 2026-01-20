package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// ConventionDetector detects coding conventions in a codebase
type ConventionDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewConventionDetector creates a new convention detector
func NewConventionDetector(rootPath string, files []types.FileInfo) *ConventionDetector {
	return &ConventionDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// NamingPattern represents detected naming conventions
type NamingPattern string

const (
	PascalCase NamingPattern = "PascalCase"
	CamelCase  NamingPattern = "camelCase"
	KebabCase  NamingPattern = "kebab-case"
	SnakeCase  NamingPattern = "snake_case"
	Unknown    NamingPattern = "unknown"
)

// Detect analyzes the codebase and returns detected conventions
func (d *ConventionDetector) Detect() ([]types.Convention, error) {
	var conventions []types.Convention

	// Detect file naming conventions
	conventions = append(conventions, d.detectFileNaming()...)

	// Detect import styles
	conventions = append(conventions, d.detectImportStyles()...)

	// Detect TypeScript configuration
	conventions = append(conventions, d.detectTypeScriptConfig()...)

	// Detect test patterns
	conventions = append(conventions, d.detectTestPatterns()...)

	// Detect code style tools
	conventions = append(conventions, d.detectCodeStyleTools()...)

	// Detect component patterns
	conventions = append(conventions, d.detectComponentPatterns()...)

	return conventions, nil
}

// detectFileNaming analyzes file naming patterns
func (d *ConventionDetector) detectFileNaming() []types.Convention {
	var conventions []types.Convention

	// Count patterns by file type
	componentPatterns := make(map[NamingPattern]int)
	utilityPatterns := make(map[NamingPattern]int)
	testPatterns := make(map[NamingPattern]int)

	componentDirs := map[string]bool{
		"components": true, "ui": true, "views": true, "pages": true,
		"layouts": true, "features": true, "modules": true,
	}
	utilityDirs := map[string]bool{
		"utils": true, "lib": true, "helpers": true, "hooks": true,
		"services": true, "api": true,
	}

	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		// Skip non-source files
		if !isSourceFile(f.Extension) {
			continue
		}

		// Get the base name without extension
		baseName := strings.TrimSuffix(f.Name, f.Extension)

		// Skip index/main files
		if baseName == "index" || baseName == "main" {
			continue
		}

		pattern := detectNamingPattern(baseName)
		if pattern == Unknown {
			continue
		}

		// Categorize by directory
		dir := filepath.Dir(f.Path)
		parts := strings.Split(dir, string(filepath.Separator))

		isComponent := false
		isUtility := false
		isTest := false

		for _, part := range parts {
			partLower := strings.ToLower(part)
			if componentDirs[partLower] {
				isComponent = true
			}
			if utilityDirs[partLower] {
				isUtility = true
			}
			if partLower == "test" || partLower == "tests" || partLower == "__tests__" {
				isTest = true
			}
		}

		// Check if test file by name
		if strings.Contains(baseName, ".test") || strings.Contains(baseName, ".spec") ||
			strings.HasSuffix(baseName, "_test") {
			isTest = true
		}

		if isTest {
			testPatterns[pattern]++
		} else if isComponent {
			componentPatterns[pattern]++
		} else if isUtility {
			utilityPatterns[pattern]++
		}
	}

	// Report dominant patterns
	if pattern, count := dominantPattern(componentPatterns); count >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "naming",
			Description: formatNamingConvention("Components", pattern),
			Example:     getPatternExample(pattern, "UserCard"),
		})
	}

	if pattern, count := dominantPattern(utilityPatterns); count >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "naming",
			Description: formatNamingConvention("Utility files", pattern),
			Example:     getPatternExample(pattern, "formatDate"),
		})
	}

	return conventions
}

// detectImportStyles checks for path aliases and import patterns
func (d *ConventionDetector) detectImportStyles() []types.Convention {
	var conventions []types.Convention

	// Check tsconfig.json for path aliases
	tsconfigPath := filepath.Join(d.rootPath, "tsconfig.json")
	if data, err := os.ReadFile(tsconfigPath); err == nil {
		var tsconfig struct {
			CompilerOptions struct {
				BaseURL string              `json:"baseUrl"`
				Paths   map[string][]string `json:"paths"`
			} `json:"compilerOptions"`
		}

		if json.Unmarshal(data, &tsconfig) == nil {
			if len(tsconfig.CompilerOptions.Paths) > 0 {
				// Find common aliases
				aliases := []string{}
				for alias := range tsconfig.CompilerOptions.Paths {
					// Clean up the alias pattern
					cleanAlias := strings.TrimSuffix(alias, "/*")
					aliases = append(aliases, cleanAlias)
				}

				if len(aliases) > 0 {
					conventions = append(conventions, types.Convention{
						Category:    "imports",
						Description: "Path aliases configured: " + strings.Join(aliases, ", "),
						Example:     "import { Button } from '" + aliases[0] + "/components/Button'",
					})
				}
			}

			if tsconfig.CompilerOptions.BaseURL != "" {
				conventions = append(conventions, types.Convention{
					Category:    "imports",
					Description: "Absolute imports enabled with baseUrl: " + tsconfig.CompilerOptions.BaseURL,
				})
			}
		}
	}

	// Check for import patterns in source files (sample a few files)
	atImports := 0
	tildeImports := 0
	relativeImports := 0
	sampledFiles := 0

	importRegex := regexp.MustCompile(`(?:import|from)\s+['"]([^'"]+)['"]`)

	for _, f := range d.files {
		if !isSourceFile(f.Extension) {
			continue
		}
		if sampledFiles >= 20 {
			break
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		matches := importRegex.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			importPath := match[1]

			if strings.HasPrefix(importPath, "@/") || strings.HasPrefix(importPath, "@") {
				atImports++
			} else if strings.HasPrefix(importPath, "~/") {
				tildeImports++
			} else if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
				relativeImports++
			}
		}
		sampledFiles++
	}

	// Report import style if clear pattern emerges
	if atImports > relativeImports && atImports >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "imports",
			Description: "Prefer @/ path alias for imports over relative paths",
			Example:     "import { utils } from '@/lib/utils'",
		})
	} else if tildeImports > relativeImports && tildeImports >= 5 {
		conventions = append(conventions, types.Convention{
			Category:    "imports",
			Description: "Prefer ~/ path alias for imports over relative paths",
			Example:     "import { utils } from '~/lib/utils'",
		})
	}

	return conventions
}

// detectTypeScriptConfig analyzes TypeScript configuration
func (d *ConventionDetector) detectTypeScriptConfig() []types.Convention {
	var conventions []types.Convention

	tsconfigPath := filepath.Join(d.rootPath, "tsconfig.json")
	data, err := os.ReadFile(tsconfigPath)
	if err != nil {
		return conventions
	}

	var tsconfig struct {
		CompilerOptions struct {
			Strict                 bool `json:"strict"`
			StrictNullChecks       bool `json:"strictNullChecks"`
			NoImplicitAny          bool `json:"noImplicitAny"`
			NoUnusedLocals         bool `json:"noUnusedLocals"`
			NoUnusedParameters     bool `json:"noUnusedParameters"`
			ExactOptionalPropertyTypes bool `json:"exactOptionalPropertyTypes"`
		} `json:"compilerOptions"`
	}

	if err := json.Unmarshal(data, &tsconfig); err != nil {
		return conventions
	}

	opts := tsconfig.CompilerOptions

	if opts.Strict {
		conventions = append(conventions, types.Convention{
			Category:    "typescript",
			Description: "TypeScript strict mode enabled - maintain strict type safety",
		})
	}

	if opts.NoImplicitAny {
		conventions = append(conventions, types.Convention{
			Category:    "typescript",
			Description: "Explicit types required - avoid 'any' type",
		})
	}

	if opts.NoUnusedLocals || opts.NoUnusedParameters {
		conventions = append(conventions, types.Convention{
			Category:    "typescript",
			Description: "Unused variables/parameters not allowed - clean up dead code",
		})
	}

	return conventions
}

// detectTestPatterns analyzes testing conventions
func (d *ConventionDetector) detectTestPatterns() []types.Convention {
	var conventions []types.Convention

	testFileCount := 0
	specFileCount := 0
	underscoreTestCount := 0
	colocatedTests := 0
	separateTestDir := 0

	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		baseName := strings.TrimSuffix(f.Name, f.Extension)
		dir := filepath.Dir(f.Path)

		// Count test file naming patterns
		if strings.HasSuffix(baseName, ".test") {
			testFileCount++
		} else if strings.HasSuffix(baseName, ".spec") {
			specFileCount++
		} else if strings.HasSuffix(baseName, "_test") {
			underscoreTestCount++
		} else {
			continue // Not a test file
		}

		// Check if colocated or in test directory
		if strings.Contains(dir, "__tests__") || strings.Contains(dir, "/test/") ||
			strings.Contains(dir, "/tests/") || strings.HasPrefix(dir, "test") {
			separateTestDir++
		} else {
			colocatedTests++
		}
	}

	// Report test naming convention
	if testFileCount > specFileCount && testFileCount > underscoreTestCount && testFileCount >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "testing",
			Description: "Test files use .test suffix",
			Example:     "Button.test.tsx, utils.test.ts",
		})
	} else if specFileCount > testFileCount && specFileCount > underscoreTestCount && specFileCount >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "testing",
			Description: "Test files use .spec suffix",
			Example:     "Button.spec.tsx, utils.spec.ts",
		})
	} else if underscoreTestCount > testFileCount && underscoreTestCount > specFileCount && underscoreTestCount >= 2 {
		conventions = append(conventions, types.Convention{
			Category:    "testing",
			Description: "Test files use _test suffix (Go style)",
			Example:     "handler_test.go, utils_test.go",
		})
	}

	// Report test location convention
	if colocatedTests > separateTestDir && colocatedTests >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "testing",
			Description: "Tests are colocated with source files",
		})
	} else if separateTestDir > colocatedTests && separateTestDir >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "testing",
			Description: "Tests are in dedicated test directories",
		})
	}

	return conventions
}

// detectCodeStyleTools checks for linting/formatting tools
func (d *ConventionDetector) detectCodeStyleTools() []types.Convention {
	var conventions []types.Convention

	// Check for ESLint
	eslintFiles := []string{
		".eslintrc", ".eslintrc.js", ".eslintrc.json", ".eslintrc.yml",
		"eslint.config.js", "eslint.config.mjs",
	}
	for _, f := range eslintFiles {
		if _, err := os.Stat(filepath.Join(d.rootPath, f)); err == nil {
			conventions = append(conventions, types.Convention{
				Category:    "code-style",
				Description: "ESLint configured - follow linting rules",
			})
			break
		}
	}

	// Check for Prettier
	prettierFiles := []string{
		".prettierrc", ".prettierrc.js", ".prettierrc.json", ".prettierrc.yml",
		"prettier.config.js", "prettier.config.mjs",
	}
	for _, f := range prettierFiles {
		if _, err := os.Stat(filepath.Join(d.rootPath, f)); err == nil {
			conventions = append(conventions, types.Convention{
				Category:    "code-style",
				Description: "Prettier configured - code formatting is automated",
			})
			break
		}
	}

	// Check for EditorConfig
	if _, err := os.Stat(filepath.Join(d.rootPath, ".editorconfig")); err == nil {
		conventions = append(conventions, types.Convention{
			Category:    "code-style",
			Description: "EditorConfig present - editor settings are standardized",
		})
	}

	// Check for Go formatting
	hasGo := false
	for _, f := range d.files {
		if f.Extension == ".go" {
			hasGo = true
			break
		}
	}
	if hasGo {
		conventions = append(conventions, types.Convention{
			Category:    "code-style",
			Description: "Go project - use 'go fmt' or 'gofmt' for formatting",
		})
	}

	return conventions
}

// detectComponentPatterns detects React/Vue/Svelte component conventions
func (d *ConventionDetector) detectComponentPatterns() []types.Convention {
	var conventions []types.Convention

	// Count component file types
	jsxCount := 0
	tsxCount := 0
	vueCount := 0
	svelteCount := 0

	// Check for barrel exports (index files)
	barrelExports := 0

	for _, f := range d.files {
		switch f.Extension {
		case ".jsx":
			jsxCount++
		case ".tsx":
			tsxCount++
		case ".vue":
			vueCount++
		case ".svelte":
			svelteCount++
		}

		if f.Name == "index.ts" || f.Name == "index.js" {
			dir := filepath.Dir(f.Path)
			if strings.Contains(dir, "components") || strings.Contains(dir, "ui") {
				barrelExports++
			}
		}
	}

	// React conventions
	if tsxCount > 0 || jsxCount > 0 {
		if tsxCount > jsxCount {
			conventions = append(conventions, types.Convention{
				Category:    "components",
				Description: "React components use TypeScript (.tsx)",
			})
		}

		// Check for function vs class components by sampling files
		funcComponents := 0
		for _, f := range d.files {
			if f.Extension != ".tsx" && f.Extension != ".jsx" {
				continue
			}
			fullPath := filepath.Join(d.rootPath, f.Path)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				continue
			}

			contentStr := string(content)
			// Simple heuristic: check for function/arrow components
			if strings.Contains(contentStr, "export function") ||
				strings.Contains(contentStr, "export const") ||
				strings.Contains(contentStr, "export default function") {
				funcComponents++
			}

			if funcComponents >= 5 {
				break
			}
		}

		if funcComponents >= 3 {
			conventions = append(conventions, types.Convention{
				Category:    "components",
				Description: "Use functional components (not class components)",
			})
		}
	}

	// Vue conventions
	if vueCount > 0 {
		conventions = append(conventions, types.Convention{
			Category:    "components",
			Description: "Vue single-file components (.vue)",
		})
	}

	// Svelte conventions
	if svelteCount > 0 {
		conventions = append(conventions, types.Convention{
			Category:    "components",
			Description: "Svelte components (.svelte)",
		})
	}

	// Barrel exports
	if barrelExports >= 3 {
		conventions = append(conventions, types.Convention{
			Category:    "structure",
			Description: "Components use barrel exports (index.ts) for cleaner imports",
			Example:     "import { Button, Card } from '@/components'",
		})
	}

	return conventions
}

// Helper functions

func isSourceFile(ext string) bool {
	sourceExts := map[string]bool{
		".js": true, ".jsx": true, ".ts": true, ".tsx": true,
		".go": true, ".py": true, ".rs": true, ".rb": true,
		".vue": true, ".svelte": true,
	}
	return sourceExts[ext]
}

func detectNamingPattern(name string) NamingPattern {
	// Skip if too short
	if len(name) < 2 {
		return Unknown
	}

	// Check for kebab-case (has hyphen, no underscore, lowercase)
	if strings.Contains(name, "-") && !strings.Contains(name, "_") {
		return KebabCase
	}

	// Check for snake_case (has underscore, no hyphen, lowercase)
	if strings.Contains(name, "_") && !strings.Contains(name, "-") {
		if name == strings.ToLower(name) {
			return SnakeCase
		}
	}

	// Check for PascalCase (starts with uppercase, no separators)
	if !strings.Contains(name, "-") && !strings.Contains(name, "_") {
		if name[0] >= 'A' && name[0] <= 'Z' {
			return PascalCase
		}
		// Check for camelCase (starts with lowercase)
		if name[0] >= 'a' && name[0] <= 'z' {
			// Must have some uppercase to be camelCase (not all lowercase)
			for _, c := range name[1:] {
				if c >= 'A' && c <= 'Z' {
					return CamelCase
				}
			}
		}
	}

	return Unknown
}

func dominantPattern(patterns map[NamingPattern]int) (NamingPattern, int) {
	var maxPattern NamingPattern
	maxCount := 0

	for pattern, count := range patterns {
		if count > maxCount {
			maxCount = count
			maxPattern = pattern
		}
	}

	return maxPattern, maxCount
}

func formatNamingConvention(fileType string, pattern NamingPattern) string {
	return fileType + " use " + string(pattern) + " naming"
}

func getPatternExample(pattern NamingPattern, baseName string) string {
	switch pattern {
	case PascalCase:
		return "UserCard.tsx, DatePicker.tsx"
	case CamelCase:
		return "formatDate.ts, useAuth.ts"
	case KebabCase:
		return "user-card.tsx, date-picker.tsx"
	case SnakeCase:
		return "user_card.tsx, date_picker.tsx"
	default:
		return ""
	}
}
