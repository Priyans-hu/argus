package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/Priyans-hu/argus/pkg/types"
)

// ClaudeGenerator generates CLAUDE.md files
type ClaudeGenerator struct{}

// NewClaudeGenerator creates a new Claude generator
func NewClaudeGenerator() *ClaudeGenerator {
	return &ClaudeGenerator{}
}

// Name returns the generator name
func (g *ClaudeGenerator) Name() string {
	return "claude"
}

// OutputFile returns the output filename
func (g *ClaudeGenerator) OutputFile() string {
	return "CLAUDE.md"
}

// Generate creates the CLAUDE.md content
func (g *ClaudeGenerator) Generate(analysis *types.Analysis) ([]byte, error) {
	var buf bytes.Buffer

	// Header
	fmt.Fprintf(&buf, "# %s\n\n", analysis.ProjectName)

	// Project Overview from README
	g.writeProjectOverview(&buf, analysis.ReadmeContent)

	// Architecture section for monorepos
	g.writeArchitecture(&buf, analysis.MonorepoInfo)

	// Tech Stack Summary
	g.writeTechStack(&buf, &analysis.TechStack)

	// Project Structure
	g.writeStructure(&buf, &analysis.Structure)

	// Key Files
	g.writeKeyFiles(&buf, analysis.KeyFiles)

	// Available Commands
	g.writeCommands(&buf, analysis.Commands)

	// API Endpoints
	g.writeEndpoints(&buf, analysis.Endpoints)

	// Conventions
	g.writeConventions(&buf, analysis.Conventions)

	// Guidelines based on tech stack
	g.writeGuidelines(&buf, &analysis.TechStack)

	// Detected patterns from deep analysis
	g.writePatterns(&buf, analysis.CodePatterns)

	// Dependencies summary
	g.writeDependencies(&buf, analysis.Dependencies)

	return buf.Bytes(), nil
}

// writeProjectOverview writes the project overview from README
func (g *ClaudeGenerator) writeProjectOverview(buf *bytes.Buffer, readme *types.ReadmeContent) {
	if readme == nil {
		return
	}

	hasContent := readme.Description != "" || len(readme.Features) > 0

	if !hasContent {
		return
	}

	buf.WriteString("## Project Overview\n\n")

	// Description
	if readme.Description != "" {
		fmt.Fprintf(buf, "%s\n\n", readme.Description)
	}

	// Features
	if len(readme.Features) > 0 {
		buf.WriteString("### Key Features\n\n")
		for _, feature := range readme.Features {
			fmt.Fprintf(buf, "- %s\n", feature)
		}
		buf.WriteString("\n")
	}
}

// writeArchitecture writes the architecture section for monorepos
func (g *ClaudeGenerator) writeArchitecture(buf *bytes.Buffer, mono *types.MonorepoInfo) {
	if mono == nil || !mono.IsMonorepo {
		return
	}

	buf.WriteString("## Architecture\n\n")

	// Monorepo tool info
	if mono.Tool != "" {
		fmt.Fprintf(buf, "This is a **%s** monorepo", mono.Tool)
		if mono.PackageManager != "" {
			fmt.Fprintf(buf, " using **%s**", mono.PackageManager)
		}
		buf.WriteString(".\n\n")
	} else {
		buf.WriteString("This is a **monorepo** with multiple packages.\n\n")
	}

	// Workspace paths
	if len(mono.WorkspacePaths) > 0 {
		buf.WriteString("**Workspaces:** ")
		buf.WriteString("`" + strings.Join(mono.WorkspacePaths, "`, `") + "`\n\n")
	}

	// Package descriptions
	if len(mono.Packages) > 0 {
		buf.WriteString("### Key Packages\n\n")
		for _, pkg := range mono.Packages {
			fmt.Fprintf(buf, "- **`%s/`** - %s\n", pkg.Path, pkg.Description)
			if len(pkg.SubPackages) > 0 {
				// Show first few sub-packages
				shown := pkg.SubPackages
				if len(shown) > 5 {
					shown = shown[:5]
				}
				fmt.Fprintf(buf, "  - Contains: `%s`", strings.Join(shown, "`, `"))
				if len(pkg.SubPackages) > 5 {
					fmt.Fprintf(buf, " and %d more", len(pkg.SubPackages)-5)
				}
				buf.WriteString("\n")
			}
		}
		buf.WriteString("\n")
	}
}

// writeTechStack writes the tech stack section
func (g *ClaudeGenerator) writeTechStack(buf *bytes.Buffer, stack *types.TechStack) {
	buf.WriteString("## Tech Stack\n\n")

	// Languages
	if len(stack.Languages) > 0 {
		buf.WriteString("### Languages\n\n")
		// Sort by percentage descending
		langs := make([]types.Language, len(stack.Languages))
		copy(langs, stack.Languages)
		sort.Slice(langs, func(i, j int) bool {
			return langs[i].Percentage > langs[j].Percentage
		})

		for _, lang := range langs {
			if lang.Version != "" {
				fmt.Fprintf(buf, "- **%s** %s (%.1f%%)\n", lang.Name, lang.Version, lang.Percentage)
			} else {
				fmt.Fprintf(buf, "- **%s** (%.1f%%)\n", lang.Name, lang.Percentage)
			}
		}
		buf.WriteString("\n")
	}

	// Frameworks by category
	if len(stack.Frameworks) > 0 {
		buf.WriteString("### Frameworks & Libraries\n\n")

		// Group by category
		byCategory := make(map[string][]types.Framework)
		for _, fw := range stack.Frameworks {
			cat := fw.Category
			if cat == "" {
				cat = "other"
			}
			byCategory[cat] = append(byCategory[cat], fw)
		}

		// Order categories
		categoryOrder := []string{"frontend", "backend", "fullstack", "database", "testing", "styling", "state", "cli", "tooling", "other"}
		categoryNames := map[string]string{
			"frontend":  "Frontend",
			"backend":   "Backend",
			"fullstack": "Full-Stack",
			"database":  "Database/ORM",
			"testing":   "Testing",
			"styling":   "Styling",
			"state":     "State Management",
			"cli":       "CLI",
			"tooling":   "Tooling",
			"other":     "Other",
		}

		for _, cat := range categoryOrder {
			fws, ok := byCategory[cat]
			if !ok || len(fws) == 0 {
				continue
			}

			catName := categoryNames[cat]
			if catName == "" {
				catName = titleCase(cat)
			}

			fmt.Fprintf(buf, "**%s:**\n", catName)
			for _, fw := range fws {
				if fw.Version != "" {
					fmt.Fprintf(buf, "- %s %s\n", fw.Name, fw.Version)
				} else {
					fmt.Fprintf(buf, "- %s\n", fw.Name)
				}
			}
			buf.WriteString("\n")
		}
	}

	// Databases
	if len(stack.Databases) > 0 {
		buf.WriteString("### Databases\n\n")
		for _, db := range stack.Databases {
			fmt.Fprintf(buf, "- %s\n", db)
		}
		buf.WriteString("\n")
	}

	// Tools
	if len(stack.Tools) > 0 {
		buf.WriteString("### Tools\n\n")
		for _, tool := range stack.Tools {
			fmt.Fprintf(buf, "- %s\n", tool)
		}
		buf.WriteString("\n")
	}
}

// writeStructure writes the project structure section
func (g *ClaudeGenerator) writeStructure(buf *bytes.Buffer, structure *types.ProjectStructure) {
	buf.WriteString("## Project Structure\n\n")

	buf.WriteString("```\n")

	// Build tree from directories
	tree := buildDirectoryTree(structure.Directories)

	// Render tree
	renderTree(buf, tree, "", true, structure.RootFiles)

	buf.WriteString("```\n\n")
}

// treeNode represents a node in the directory tree
type treeNode struct {
	name     string
	purpose  string
	children map[string]*treeNode
}

// buildDirectoryTree builds a tree structure from flat directory paths
func buildDirectoryTree(dirs []types.Directory) *treeNode {
	root := &treeNode{name: ".", children: make(map[string]*treeNode)}

	for _, dir := range dirs {
		parts := strings.Split(dir.Path, "/")
		current := root

		for i, part := range parts {
			if _, exists := current.children[part]; !exists {
				current.children[part] = &treeNode{
					name:     part,
					children: make(map[string]*treeNode),
				}
			}
			current = current.children[part]

			// Set purpose on the deepest node
			if i == len(parts)-1 {
				current.purpose = dir.Purpose
			}
		}
	}

	return root
}

// renderTree renders the tree structure to the buffer
func renderTree(buf *bytes.Buffer, node *treeNode, prefix string, isRoot bool, rootFiles []string) {
	if isRoot {
		buf.WriteString(".\n")
	}

	// Get sorted children names
	var childNames []string
	for name := range node.children {
		childNames = append(childNames, name)
	}
	sort.Strings(childNames)

	// Count total items (dirs + root files if at root level)
	totalItems := len(childNames)
	if isRoot {
		totalItems += len(rootFiles)
	}

	// Render directories
	for i, name := range childNames {
		child := node.children[name]
		isLast := (i == len(childNames)-1) && (isRoot && len(rootFiles) == 0 || !isRoot)

		// Choose connector
		connector := "├── "
		if isLast && (!isRoot || len(rootFiles) == 0) {
			connector = "└── "
		}

		// Format line with purpose comment
		if child.purpose != "" {
			fmt.Fprintf(buf, "%s%s%s/          # %s\n", prefix, connector, name, child.purpose)
		} else {
			fmt.Fprintf(buf, "%s%s%s/\n", prefix, connector, name)
		}

		// Render children with updated prefix
		newPrefix := prefix
		if isLast && (!isRoot || len(rootFiles) == 0) {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		if len(child.children) > 0 {
			renderTree(buf, child, newPrefix, false, nil)
		}
	}

	// Render root files at the end
	if isRoot && len(rootFiles) > 0 {
		sortedFiles := make([]string, len(rootFiles))
		copy(sortedFiles, rootFiles)
		sort.Strings(sortedFiles)

		for i, f := range sortedFiles {
			connector := "├── "
			if i == len(sortedFiles)-1 {
				connector = "└── "
			}
			fmt.Fprintf(buf, "%s%s%s\n", prefix, connector, f)
		}
	}
}

// writeKeyFiles writes the key files section
func (g *ClaudeGenerator) writeKeyFiles(buf *bytes.Buffer, keyFiles []types.KeyFile) {
	if len(keyFiles) == 0 {
		return
	}

	buf.WriteString("## Key Files\n\n")
	buf.WriteString("| File | Purpose | Description |\n")
	buf.WriteString("|------|---------|-------------|\n")

	for _, kf := range keyFiles {
		desc := kf.Description
		if desc == "" {
			desc = "-"
		}
		fmt.Fprintf(buf, "| `%s` | %s | %s |\n", kf.Path, kf.Purpose, desc)
	}
	buf.WriteString("\n")
}

// writeCommands writes the available commands section
func (g *ClaudeGenerator) writeCommands(buf *bytes.Buffer, commands []types.Command) {
	if len(commands) == 0 {
		return
	}

	buf.WriteString("## Available Commands\n\n")
	buf.WriteString("```bash\n")

	for _, cmd := range commands {
		if cmd.Description != "" {
			fmt.Fprintf(buf, "# %s\n", cmd.Description)
		}
		fmt.Fprintf(buf, "%s\n", cmd.Name)
		if cmd.Command != "" && cmd.Command != cmd.Name {
			fmt.Fprintf(buf, "# → %s\n", cmd.Command)
		}
		buf.WriteString("\n")
	}

	buf.WriteString("```\n\n")
}

// writeEndpoints writes the API endpoints section
func (g *ClaudeGenerator) writeEndpoints(buf *bytes.Buffer, endpoints []types.Endpoint) {
	if len(endpoints) == 0 {
		return
	}

	buf.WriteString("## API Endpoints\n\n")
	buf.WriteString("| Method | Path | File | Auth |\n")
	buf.WriteString("|--------|------|------|------|\n")

	// Limit to 50 endpoints to avoid huge tables
	limit := 50
	if len(endpoints) < limit {
		limit = len(endpoints)
	}

	for i := 0; i < limit; i++ {
		ep := endpoints[i]
		auth := ep.Auth
		if auth == "" {
			auth = "-"
		}
		file := ep.File
		if ep.Line > 0 {
			file = fmt.Sprintf("%s:%d", ep.File, ep.Line)
		}
		fmt.Fprintf(buf, "| %s | `%s` | `%s` | %s |\n", ep.Method, ep.Path, file, auth)
	}

	if len(endpoints) > 50 {
		fmt.Fprintf(buf, "\n*...and %d more endpoints*\n", len(endpoints)-50)
	}

	buf.WriteString("\n")
}

// writeConventions writes the conventions section
func (g *ClaudeGenerator) writeConventions(buf *bytes.Buffer, conventions []types.Convention) {
	if len(conventions) == 0 {
		return
	}

	buf.WriteString("## Coding Conventions\n\n")

	// Group by category
	byCategory := make(map[string][]types.Convention)
	for _, conv := range conventions {
		cat := conv.Category
		if cat == "" {
			cat = "general"
		}
		byCategory[cat] = append(byCategory[cat], conv)
	}

	for cat, convs := range byCategory {
		fmt.Fprintf(buf, "### %s\n\n", titleCase(cat))
		for _, conv := range convs {
			fmt.Fprintf(buf, "- %s\n", conv.Description)
			if conv.Example != "" {
				fmt.Fprintf(buf, "  ```\n  %s\n  ```\n", conv.Example)
			}
		}
		buf.WriteString("\n")
	}
}

// writeGuidelines writes actionable coding guidelines based on tech stack
func (g *ClaudeGenerator) writeGuidelines(buf *bytes.Buffer, stack *types.TechStack) {
	var dos []string
	var donts []string

	// Check for languages and frameworks
	hasGo := false
	hasTypeScript := false
	hasJavaScript := false
	hasPython := false
	hasReact := false
	hasVue := false
	hasAngular := false

	for _, lang := range stack.Languages {
		switch strings.ToLower(lang.Name) {
		case "go":
			hasGo = true
		case "typescript":
			hasTypeScript = true
		case "javascript":
			hasJavaScript = true
		case "python":
			hasPython = true
		}
	}

	for _, fw := range stack.Frameworks {
		switch strings.ToLower(fw.Name) {
		case "react", "next.js", "nextjs":
			hasReact = true
		case "vue", "nuxt", "nuxt.js":
			hasVue = true
		case "angular":
			hasAngular = true
		}
	}

	// Go guidelines
	if hasGo {
		dos = append(dos,
			"Use `gofmt` or `goimports` for consistent formatting",
			"Handle all errors explicitly with `if err != nil`",
			"Use meaningful variable names; short names for short scopes",
			"Write doc comments for exported functions starting with function name",
			"Prefer composition over inheritance",
		)
		donts = append(donts,
			"Don't use `panic()` for regular error handling",
			"Don't ignore errors with `_`",
			"Don't use global state unnecessarily",
		)
	}

	// TypeScript/JavaScript guidelines
	if hasTypeScript || hasJavaScript {
		dos = append(dos,
			"Use `const` for variables that don't change, `let` for those that do",
			"Use async/await over raw promises when possible",
			"Handle errors properly in try/catch blocks",
		)
		donts = append(donts,
			"Don't use `var`, prefer `const`/`let`",
			"Don't use `any` type without good reason (TypeScript)",
			"Don't nest ternary operators",
		)
	}

	// React guidelines
	if hasReact {
		dos = append(dos,
			"Use functional components with hooks",
			"Extract reusable logic into custom hooks",
			"Use `const` for component declarations",
			"Add `data-testid` to interactive elements for testing",
		)
		donts = append(donts,
			"Don't use class components for new code",
			"Don't overuse `useEffect` - consider if it's truly needed",
			"Don't mutate state directly, use setter functions",
		)
	}

	// Vue guidelines
	if hasVue {
		dos = append(dos,
			"Use Composition API with `<script setup>` for new components",
			"Use reactive/ref for state management",
			"Extract reusable logic into composables",
		)
		donts = append(donts,
			"Don't mix Options API and Composition API in same component",
			"Don't mutate props directly",
		)
	}

	// Angular guidelines
	if hasAngular {
		dos = append(dos,
			"Use standalone components for new code",
			"Use signals for reactive state",
			"Use dependency injection for services",
		)
		donts = append(donts,
			"Don't subscribe without unsubscribing (use async pipe or takeUntil)",
			"Don't put business logic in components",
		)
	}

	// Python guidelines
	if hasPython {
		dos = append(dos,
			"Follow PEP 8 style guide",
			"Use type hints for function signatures",
			"Use snake_case for functions and variables",
			"Write docstrings for functions and classes",
			"Use context managers (`with` statements) for resource management",
		)
		donts = append(donts,
			"Don't use mutable default arguments",
			"Don't catch generic `Exception` without re-raising",
			"Don't use `from module import *`",
		)
	}

	// Only write section if we have guidelines
	if len(dos) == 0 && len(donts) == 0 {
		return
	}

	buf.WriteString("## Guidelines\n\n")

	if len(dos) > 0 {
		buf.WriteString("### Do\n\n")
		for _, d := range dos {
			fmt.Fprintf(buf, "- %s\n", d)
		}
		buf.WriteString("\n")
	}

	if len(donts) > 0 {
		buf.WriteString("### Don't\n\n")
		for _, d := range donts {
			fmt.Fprintf(buf, "- %s\n", d)
		}
		buf.WriteString("\n")
	}
}

// writePatterns writes detected code patterns from deep analysis
func (g *ClaudeGenerator) writePatterns(buf *bytes.Buffer, patterns *types.CodePatterns) {
	if patterns == nil {
		return
	}

	// Collect all patterns that have findings
	type patternSection struct {
		title    string
		patterns []types.PatternInfo
	}

	sections := []patternSection{
		{"State Management", patterns.StateManagement},
		{"Data Fetching", patterns.DataFetching},
		{"Routing", patterns.Routing},
		{"Forms", patterns.Forms},
		{"Testing", patterns.Testing},
		{"Styling", patterns.Styling},
		{"Authentication", patterns.Authentication},
		{"API Patterns", patterns.APIPatterns},
		{"Database & ORM", patterns.DatabaseORM},
		{"Utilities", patterns.Utilities},
	}

	// Check if any patterns were detected
	hasPatterns := false
	for _, s := range sections {
		if len(s.patterns) > 0 {
			hasPatterns = true
			break
		}
	}

	if !hasPatterns {
		return
	}

	buf.WriteString("## Detected Patterns\n\n")
	buf.WriteString("*The following patterns were detected by scanning the codebase:*\n\n")

	for _, section := range sections {
		if len(section.patterns) == 0 {
			continue
		}

		fmt.Fprintf(buf, "### %s\n\n", section.title)

		// Group by significance (file count)
		for _, p := range section.patterns {
			if p.FileCount > 0 {
				fmt.Fprintf(buf, "- **%s** - %s", p.Name, p.Description)
				if p.FileCount > 1 {
					fmt.Fprintf(buf, " (%d files)", p.FileCount)
				}
				buf.WriteString("\n")

				// Show example files for significant patterns
				if p.FileCount >= 3 && len(p.Examples) > 0 {
					buf.WriteString("  - Found in: ")
					for i, ex := range p.Examples {
						if i > 0 {
							buf.WriteString(", ")
						}
						fmt.Fprintf(buf, "`%s`", ex)
					}
					buf.WriteString("\n")
				}
			}
		}
		buf.WriteString("\n")
	}
}

// writeDependencies writes a summary of key dependencies
func (g *ClaudeGenerator) writeDependencies(buf *bytes.Buffer, deps []types.Dependency) {
	if len(deps) == 0 {
		return
	}

	// Group by type
	runtime := []types.Dependency{}
	dev := []types.Dependency{}

	for _, d := range deps {
		if d.Type == "dev" || d.Type == "devDependencies" {
			dev = append(dev, d)
		} else {
			runtime = append(runtime, d)
		}
	}

	buf.WriteString("## Dependencies\n\n")

	if len(runtime) > 0 {
		buf.WriteString("### Runtime\n\n")
		// Limit to top 20 to avoid huge lists
		limit := 20
		if len(runtime) < limit {
			limit = len(runtime)
		}
		for i := 0; i < limit; i++ {
			fmt.Fprintf(buf, "- `%s` %s\n", runtime[i].Name, runtime[i].Version)
		}
		if len(runtime) > 20 {
			fmt.Fprintf(buf, "\n*...and %d more*\n", len(runtime)-20)
		}
		buf.WriteString("\n")
	}

	if len(dev) > 0 {
		buf.WriteString("### Development\n\n")
		limit := 15
		if len(dev) < limit {
			limit = len(dev)
		}
		for i := 0; i < limit; i++ {
			fmt.Fprintf(buf, "- `%s` %s\n", dev[i].Name, dev[i].Version)
		}
		if len(dev) > 15 {
			fmt.Fprintf(buf, "\n*...and %d more*\n", len(dev)-15)
		}
		buf.WriteString("\n")
	}
}

// titleCase converts the first letter of a string to uppercase
func titleCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
