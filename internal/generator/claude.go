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

	// Quick Reference (commands table)
	g.writeQuickReference(&buf, analysis.Commands)

	// Architecture section for monorepos
	g.writeArchitecture(&buf, analysis.MonorepoInfo)

	// Architecture diagram
	g.writeArchitectureDiagram(&buf, analysis.ArchitectureInfo)

	// Tech Stack Summary
	g.writeTechStack(&buf, &analysis.TechStack)

	// Project Structure
	g.writeStructure(&buf, &analysis.Structure)

	// Key Files
	g.writeKeyFiles(&buf, analysis.KeyFiles)

	// Configuration System
	g.writeConfigurationSystem(&buf, analysis.ConfigFiles)

	// Development Setup
	g.writeDevelopmentSetup(&buf, analysis.DevelopmentInfo)

	// Available Commands (detailed)
	g.writeCommands(&buf, analysis.Commands)

	// CLI Output & Verbosity
	g.writeCLIOutput(&buf, analysis.CLIInfo)

	// API Endpoints
	g.writeEndpoints(&buf, analysis.Endpoints)

	// Conventions (includes git conventions)
	g.writeConventions(&buf, analysis.Conventions, analysis.GitConventions)

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

// writeArchitectureDiagram writes the architecture diagram section
func (g *ClaudeGenerator) writeArchitectureDiagram(buf *bytes.Buffer, arch *types.ArchitectureInfo) {
	if arch == nil || arch.Diagram == "" {
		return
	}

	// Only write if we have meaningful architecture info
	if arch.Style == "" && len(arch.Layers) == 0 {
		return
	}

	buf.WriteString("## Architecture\n\n")

	if arch.Style != "" {
		fmt.Fprintf(buf, "**Style:** %s\n\n", arch.Style)
	}

	if arch.EntryPoint != "" {
		fmt.Fprintf(buf, "**Entry Point:** `%s`\n\n", arch.EntryPoint)
	}

	// Write diagram
	if arch.Diagram != "" {
		buf.WriteString(arch.Diagram)
		buf.WriteString("\n")
	}

	// Write layer dependencies
	if len(arch.Layers) > 0 {
		hasDepends := false
		for _, layer := range arch.Layers {
			if len(layer.DependsOn) > 0 {
				hasDepends = true
				break
			}
		}

		if hasDepends {
			buf.WriteString("**Package Dependencies:**\n")
			for _, layer := range arch.Layers {
				if len(layer.DependsOn) > 0 {
					fmt.Fprintf(buf, "- `%s` → `%s`\n", layer.Name, strings.Join(layer.DependsOn, "`, `"))
				}
			}
			buf.WriteString("\n")
		}
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

// writeEndpoints writes the API endpoints section grouped by resource
func (g *ClaudeGenerator) writeEndpoints(buf *bytes.Buffer, endpoints []types.Endpoint) {
	if len(endpoints) == 0 {
		return
	}

	buf.WriteString("## API Endpoints\n\n")

	// Group endpoints by resource (first path segment after /)
	grouped := groupEndpointsByResource(endpoints)

	// Get sorted resource names
	var resources []string
	for resource := range grouped {
		resources = append(resources, resource)
	}
	sort.Strings(resources)

	totalShown := 0
	maxEndpoints := 100 // Show up to 100 endpoints total

	for _, resource := range resources {
		if totalShown >= maxEndpoints {
			break
		}

		eps := grouped[resource]
		if len(eps) == 0 {
			continue
		}

		// Write resource header
		displayResource := resource
		if displayResource == "" || displayResource == "/" {
			displayResource = "Root"
		}
		fmt.Fprintf(buf, "### %s\n\n", displayResource)
		buf.WriteString("| Method | Path | File |\n")
		buf.WriteString("|--------|------|------|\n")

		for _, ep := range eps {
			if totalShown >= maxEndpoints {
				break
			}

			file := ep.File
			if ep.Line > 0 {
				file = fmt.Sprintf("%s:%d", ep.File, ep.Line)
			}

			// Add auth indicator if present
			path := ep.Path
			if ep.Auth != "" {
				path = fmt.Sprintf("%s 🔒", ep.Path)
			}

			fmt.Fprintf(buf, "| %s | `%s` | `%s` |\n", ep.Method, path, file)
			totalShown++
		}

		buf.WriteString("\n")
	}

	remaining := len(endpoints) - totalShown
	if remaining > 0 {
		fmt.Fprintf(buf, "*...and %d more endpoints*\n\n", remaining)
	}
}

// groupEndpointsByResource groups endpoints by their resource path prefix
func groupEndpointsByResource(endpoints []types.Endpoint) map[string][]types.Endpoint {
	grouped := make(map[string][]types.Endpoint)

	for _, ep := range endpoints {
		resource := extractResourcePrefix(ep.Path)
		grouped[resource] = append(grouped[resource], ep)
	}

	// Sort endpoints within each group by path, then method
	for resource := range grouped {
		eps := grouped[resource]
		sort.Slice(eps, func(i, j int) bool {
			if eps[i].Path == eps[j].Path {
				return methodPriority(eps[i].Method) < methodPriority(eps[j].Method)
			}
			return eps[i].Path < eps[j].Path
		})
		grouped[resource] = eps
	}

	return grouped
}

// extractResourcePrefix extracts the resource name from a path
// /api/users/123 -> /api/users
// /users/:id -> /users
// /v1/products/categories -> /v1/products
func extractResourcePrefix(path string) string {
	// Clean the path
	path = strings.Trim(path, "/")
	if path == "" {
		return "/"
	}

	parts := strings.Split(path, "/")

	// Handle API versioning prefixes
	startIdx := 0
	if len(parts) > 0 {
		first := strings.ToLower(parts[0])
		// Check for common prefixes like api, v1, v2
		if first == "api" || (len(first) >= 2 && first[0] == 'v' && isDigit(first[1])) {
			startIdx = 1
		}
	}

	// Build resource path
	var resourceParts []string

	// Include prefix (api, v1, etc.)
	for i := 0; i < startIdx && i < len(parts); i++ {
		resourceParts = append(resourceParts, parts[i])
	}

	// Add the main resource (first non-prefix, non-param segment)
	for i := startIdx; i < len(parts) && i < startIdx+2; i++ {
		part := parts[i]
		// Skip dynamic segments
		if strings.HasPrefix(part, ":") || strings.HasPrefix(part, "{") ||
			strings.HasPrefix(part, "[") || part == "*" {
			break
		}
		resourceParts = append(resourceParts, part)
		// Stop at 2 meaningful segments
		if len(resourceParts)-startIdx >= 1 {
			break
		}
	}

	if len(resourceParts) == 0 {
		return "/"
	}

	return "/" + strings.Join(resourceParts, "/")
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func methodPriority(method string) int {
	priority := map[string]int{
		"GET": 1, "POST": 2, "PUT": 3, "PATCH": 4, "DELETE": 5, "ALL": 6,
	}
	if p, ok := priority[method]; ok {
		return p
	}
	return 99
}

// writeConventions writes the conventions section
func (g *ClaudeGenerator) writeConventions(buf *bytes.Buffer, conventions []types.Convention, gitConventions *types.GitConventions) {
	hasConventions := len(conventions) > 0
	hasGitConventions := gitConventions != nil && (gitConventions.CommitConvention != nil || gitConventions.BranchConvention != nil)

	if !hasConventions && !hasGitConventions {
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

	// Write git conventions
	g.writeGitConventions(buf, gitConventions)
}

// writeGitConventions writes git commit and branch conventions
func (g *ClaudeGenerator) writeGitConventions(buf *bytes.Buffer, git *types.GitConventions) {
	if git == nil {
		return
	}

	// Commit conventions
	if git.CommitConvention != nil {
		cc := git.CommitConvention
		buf.WriteString("### Git\n\n")

		// Commit message format
		if cc.Style != "" {
			styleName := cc.Style
			switch cc.Style {
			case "conventional":
				styleName = "Conventional Commits"
			case "angular":
				styleName = "Angular style"
			case "gitmoji":
				styleName = "Gitmoji"
			case "jira":
				styleName = "Jira ticket prefix"
			}
			fmt.Fprintf(buf, "- Commit style: **%s**\n", styleName)
		}

		if cc.Format != "" {
			fmt.Fprintf(buf, "  - Format: `%s`\n", cc.Format)
		}

		if len(cc.Types) > 0 {
			fmt.Fprintf(buf, "  - Types: `%s`\n", strings.Join(cc.Types, "`, `"))
		}

		if len(cc.Scopes) > 0 {
			fmt.Fprintf(buf, "  - Scopes: `%s`\n", strings.Join(cc.Scopes, "`, `"))
		}

		if cc.Example != "" {
			fmt.Fprintf(buf, "  - Example: `%s`\n", cc.Example)
		}
	}

	// Branch naming conventions
	if git.BranchConvention != nil {
		bc := git.BranchConvention

		// If no commit convention was written, add Git header
		if git.CommitConvention == nil {
			buf.WriteString("### Git\n\n")
		}

		if len(bc.Prefixes) > 0 {
			fmt.Fprintf(buf, "- Branch naming uses prefixes: %s\n", strings.Join(bc.Prefixes, ", "))
			if len(bc.Examples) > 0 {
				fmt.Fprintf(buf, "  ```\n  %s\n  ```\n", strings.Join(bc.Examples, ", "))
			}
		}
	}

	buf.WriteString("\n")
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

// writeQuickReference writes a quick reference section for common commands
func (g *ClaudeGenerator) writeQuickReference(buf *bytes.Buffer, commands []types.Command) {
	if len(commands) == 0 {
		return
	}

	// Classify commands
	classified := g.classifyAllCommands(commands)

	// Check if we have any meaningful categories
	categories := []string{"Development", "Build", "Test", "Lint", "Format", "Setup"}
	hasContent := false
	for _, cat := range categories {
		if cmds, ok := classified[cat]; ok && len(cmds) > 0 {
			hasContent = true
			break
		}
	}

	if !hasContent {
		return
	}

	buf.WriteString("## Quick Reference\n\n")
	buf.WriteString("```bash\n")

	for _, cat := range categories {
		cmds, ok := classified[cat]
		if !ok || len(cmds) == 0 {
			continue
		}

		fmt.Fprintf(buf, "# %s\n", cat)
		for _, cmd := range cmds {
			// Format command with description as comment
			if cmd.Description != "" {
				fmt.Fprintf(buf, "%-24s # %s\n", cmd.Name, cmd.Description)
			} else {
				fmt.Fprintf(buf, "%s\n", cmd.Name)
			}
		}
		buf.WriteString("\n")
	}

	buf.WriteString("```\n\n")
}

// classifyAllCommands groups commands by category
func (g *ClaudeGenerator) classifyAllCommands(commands []types.Command) map[string][]types.Command {
	result := make(map[string][]types.Command)

	for _, cmd := range commands {
		category := g.categorizeCommand(cmd)
		result[category] = append(result[category], cmd)
	}

	return result
}

// categorizeCommand determines the category of a command
func (g *ClaudeGenerator) categorizeCommand(cmd types.Command) string {
	nameLower := strings.ToLower(cmd.Name)
	cmdLower := strings.ToLower(cmd.Command)

	switch {
	case strings.Contains(nameLower, "test") || strings.Contains(cmdLower, "test"):
		return "Test"
	case strings.Contains(nameLower, "lint") || strings.Contains(cmdLower, "lint"):
		return "Lint"
	case strings.Contains(nameLower, "build") || strings.Contains(cmdLower, "build"):
		return "Build"
	case strings.Contains(nameLower, "fmt") || strings.Contains(nameLower, "format") ||
		strings.Contains(cmdLower, "fmt") || strings.Contains(cmdLower, "format"):
		return "Format"
	case nameLower == "dev" || nameLower == "start" || strings.Contains(nameLower, "serve") ||
		strings.Contains(nameLower, "watch"):
		return "Development"
	case strings.Contains(nameLower, "setup") || strings.Contains(nameLower, "install") ||
		strings.Contains(nameLower, "deps"):
		return "Setup"
	default:
		return "Other"
	}
}

// writeDevelopmentSetup writes the development setup section
func (g *ClaudeGenerator) writeDevelopmentSetup(buf *bytes.Buffer, devInfo *types.DevelopmentInfo) {
	if devInfo == nil {
		return
	}

	hasContent := len(devInfo.Prerequisites) > 0 ||
		len(devInfo.SetupSteps) > 0 ||
		len(devInfo.GitHooks) > 0

	if !hasContent {
		return
	}

	buf.WriteString("## Development Setup\n\n")

	// Prerequisites
	if len(devInfo.Prerequisites) > 0 {
		buf.WriteString("### Prerequisites\n\n")
		for _, p := range devInfo.Prerequisites {
			if p.Version != "" {
				fmt.Fprintf(buf, "- %s %s\n", p.Name, p.Version)
			} else {
				fmt.Fprintf(buf, "- %s\n", p.Name)
			}
		}
		buf.WriteString("\n")
	}

	// Setup Steps
	if len(devInfo.SetupSteps) > 0 {
		buf.WriteString("### Initial Setup\n\n")
		buf.WriteString("```bash\n")
		for _, step := range devInfo.SetupSteps {
			if step.Command != "" {
				if step.Description != "" {
					fmt.Fprintf(buf, "%-24s # %s\n", step.Command, step.Description)
				} else {
					fmt.Fprintf(buf, "%s\n", step.Command)
				}
			}
		}
		buf.WriteString("```\n\n")
	}

	// Git Hooks
	if len(devInfo.GitHooks) > 0 {
		buf.WriteString("### Git Hooks\n\n")
		for _, hook := range devInfo.GitHooks {
			fmt.Fprintf(buf, "- **%s**", hook.Name)
			if len(hook.Actions) > 0 {
				fmt.Fprintf(buf, ": %s", strings.Join(hook.Actions, ", "))
			}
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}
}

// writeConfigurationSystem writes the configuration files section
func (g *ClaudeGenerator) writeConfigurationSystem(buf *bytes.Buffer, configs []types.ConfigFileInfo) {
	if len(configs) == 0 {
		return
	}

	buf.WriteString("## Configuration\n\n")

	// Sort configs by type for better organization
	sort.Slice(configs, func(i, j int) bool {
		if configs[i].Type == configs[j].Type {
			return configs[i].Path < configs[j].Path
		}
		return configs[i].Type < configs[j].Type
	})

	buf.WriteString("| File | Type | Purpose |\n")
	buf.WriteString("|------|------|--------|\n")

	for _, cfg := range configs {
		fmt.Fprintf(buf, "| `%s` | %s | %s |\n", cfg.Path, cfg.Type, cfg.Purpose)
	}
	buf.WriteString("\n")
}

// writeCLIOutput writes the CLI output and verbosity section
func (g *ClaudeGenerator) writeCLIOutput(buf *bytes.Buffer, cliInfo *types.CLIInfo) {
	if cliInfo == nil {
		return
	}

	hasContent := cliInfo.VerboseFlag != "" || cliInfo.DryRunFlag != "" || len(cliInfo.Indicators) > 0

	if !hasContent {
		return
	}

	buf.WriteString("## CLI Output & Verbosity\n\n")

	// Output levels table
	if cliInfo.VerboseFlag != "" || cliInfo.DryRunFlag != "" {
		buf.WriteString("| Flag | Output |\n")
		buf.WriteString("|------|--------|\n")
		buf.WriteString("| (none) | Progress indicators, success/error messages |\n")
		if cliInfo.VerboseFlag != "" {
			fmt.Fprintf(buf, "| `%s` | Detailed analysis results, file-by-file processing |\n", cliInfo.VerboseFlag)
		}
		if cliInfo.DryRunFlag != "" {
			fmt.Fprintf(buf, "| `%s` | Preview output without writing files |\n", cliInfo.DryRunFlag)
		}
		buf.WriteString("\n")
	}

	// Output indicators
	if len(cliInfo.Indicators) > 0 {
		buf.WriteString("### Output Indicators\n\n")
		for _, ind := range cliInfo.Indicators {
			fmt.Fprintf(buf, "- `%s` - %s\n", ind.Symbol, ind.Meaning)
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
