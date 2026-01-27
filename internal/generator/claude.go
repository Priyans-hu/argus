package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
)

// ClaudeGenerator generates CLAUDE.md files
type ClaudeGenerator struct {
	compact bool // Generate compact output for token efficiency
}

// NewClaudeGenerator creates a new Claude generator
func NewClaudeGenerator() *ClaudeGenerator {
	return &ClaudeGenerator{}
}

// SetCompact enables compact mode for smaller output
func (g *ClaudeGenerator) SetCompact(compact bool) {
	g.compact = compact
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

	// Architecture section for monorepos (skip in compact mode if not monorepo)
	if !g.compact || analysis.MonorepoInfo != nil {
		g.writeArchitecture(&buf, analysis.MonorepoInfo)
	}

	// Architecture diagram (simplified in compact mode)
	if g.compact {
		g.writeArchitectureDiagramCompact(&buf, analysis.ArchitectureInfo)
	} else {
		g.writeArchitectureDiagram(&buf, analysis.ArchitectureInfo)
	}

	// Tech Stack Summary
	g.writeTechStack(&buf, &analysis.TechStack)

	// Project Structure
	g.writeStructure(&buf, &analysis.Structure)

	// Key Files (limit in compact mode)
	if g.compact {
		g.writeKeyFilesCompact(&buf, analysis.KeyFiles)
	} else {
		g.writeKeyFiles(&buf, analysis.KeyFiles)
	}

	// Configuration System (skip in compact mode)
	if !g.compact {
		g.writeConfigurationSystem(&buf, analysis.ConfigFiles)
	}

	// Development Setup
	g.writeDevelopmentSetup(&buf, analysis.DevelopmentInfo)

	// Available Commands (detailed) - skip in compact, Quick Reference has essentials
	if !g.compact {
		g.writeCommands(&buf, analysis.Commands)
	}

	// CLI Output & Verbosity (skip in compact mode)
	if !g.compact {
		g.writeCLIOutput(&buf, analysis.CLIInfo)
	}

	// API Endpoints (limit in compact mode)
	if g.compact {
		g.writeEndpointsCompact(&buf, analysis.Endpoints)
	} else {
		g.writeEndpoints(&buf, analysis.Endpoints)
	}

	// Conventions (includes git conventions)
	g.writeConventions(&buf, analysis.Conventions, analysis.GitConventions)

	// Guidelines based on tech stack
	g.writeGuidelines(&buf, &analysis.TechStack)

	// Detected patterns from deep analysis (limited in compact mode)
	if g.compact {
		g.writePatternsCompact(&buf, analysis.CodePatterns)
	} else {
		g.writePatterns(&buf, analysis.CodePatterns)
	}

	// Dependencies summary (skip in compact mode)
	if !g.compact {
		g.writeDependencies(&buf, analysis.Dependencies)
	}

	// AI Usage Insights (from Claude Code session logs)
	g.writeUsageInsights(&buf, analysis.UsageInsights)

	// Import references to .claude/ rules (if claude-code format is also being generated)
	g.writeImports(&buf, analysis)

	return buf.Bytes(), nil
}

// writeProjectOverview writes the project overview from README
func (g *ClaudeGenerator) writeProjectOverview(buf *bytes.Buffer, readme *types.ReadmeContent) {
	if readme == nil {
		return
	}

	hasContent := readme.Description != "" || len(readme.Features) > 0 || len(readme.ModelSpecs) > 0

	if !hasContent {
		return
	}

	buf.WriteString("## Project Overview\n\n")

	// Description
	if readme.Description != "" {
		fmt.Fprintf(buf, "%s\n\n", readme.Description)
	}

	// Model Specs (for ML projects)
	if len(readme.ModelSpecs) > 0 {
		buf.WriteString("### Model Specifications\n\n")
		// Order specs for consistency
		specOrder := []string{"parameters", "architecture", "layers", "heads", "embedding", "context_length"}
		for _, key := range specOrder {
			if value, ok := readme.ModelSpecs[key]; ok {
				displayName := toTitleCase(strings.ReplaceAll(key, "_", " "))
				fmt.Fprintf(buf, "- **%s:** %s\n", displayName, value)
			}
		}
		// Any remaining specs not in our order
		for key, value := range readme.ModelSpecs {
			found := false
			for _, ordered := range specOrder {
				if key == ordered {
					found = true
					break
				}
			}
			if !found {
				displayName := toTitleCase(strings.ReplaceAll(key, "_", " "))
				fmt.Fprintf(buf, "- **%s:** %s\n", displayName, value)
			}
		}
		buf.WriteString("\n")
	}

	// Features
	if len(readme.Features) > 0 {
		buf.WriteString("### Key Features\n\n")
		for _, feature := range readme.Features {
			fmt.Fprintf(buf, "- %s\n", feature)
		}
		buf.WriteString("\n")
	}

	// Prerequisites
	if len(readme.Prerequisites) > 0 {
		buf.WriteString("### Prerequisites\n\n")
		for _, prereq := range readme.Prerequisites {
			fmt.Fprintf(buf, "- %s\n", prereq)
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
	if arch == nil {
		return
	}

	// Only write if we have meaningful architecture info (layers or style)
	// Skip generic diagrams that don't add value
	if arch.Style == "" && len(arch.Layers) == 0 {
		return
	}

	// Skip if diagram is too basic (just entry point -> external services)
	if arch.Diagram != "" && len(arch.Layers) <= 1 && !strings.Contains(arch.Diagram, "│") {
		// Diagram has no internal structure, skip it
		arch.Diagram = ""
	}

	buf.WriteString("## Architecture\n\n")

	if arch.Style != "" {
		fmt.Fprintf(buf, "**Style:** %s\n\n", arch.Style)
	}

	if arch.EntryPoint != "" {
		fmt.Fprintf(buf, "**Entry Point:** `%s`\n\n", arch.EntryPoint)
	}

	// Write diagram only if it has meaningful structure
	if arch.Diagram != "" && len(arch.Layers) > 1 {
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

	for i, cmd := range commands {
		// Add blank line before commands with descriptions (groups)
		hasDescription := cmd.Description != ""
		if hasDescription && i > 0 {
			buf.WriteString("\n")
		}
		if hasDescription {
			fmt.Fprintf(buf, "# %s\n", cmd.Description)
		}
		fmt.Fprintf(buf, "%s\n", cmd.Name)
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

	hasContent := false

	// Repository information
	if git.Repository != nil && git.Repository.RemoteURL != "" {
		buf.WriteString("### Git\n\n")
		hasContent = true

		repo := git.Repository
		if repo.Platform != "" && repo.Owner != "" && repo.Name != "" {
			fmt.Fprintf(buf, "- Repository: [%s/%s](%s)\n", repo.Owner, repo.Name, repo.RemoteURL)
		} else {
			fmt.Fprintf(buf, "- Repository: %s\n", repo.RemoteURL)
		}
	}

	// Commit conventions
	if git.CommitConvention != nil {
		cc := git.CommitConvention

		if !hasContent {
			buf.WriteString("### Git\n\n")
			hasContent = true
		}

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

		// If no content was written yet, add Git header
		if !hasContent {
			buf.WriteString("### Git\n\n")
		}

		if len(bc.Prefixes) > 0 {
			fmt.Fprintf(buf, "- Branch naming uses prefixes: %s\n", strings.Join(bc.Prefixes, ", "))
			if len(bc.Examples) > 0 {
				fmt.Fprintf(buf, "  ```\n  %s\n  ```\n", strings.Join(bc.Examples, ", "))
			}
		}
	}

	if hasContent {
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
		{"Go Patterns", patterns.GoPatterns},
		{"Rust Patterns", patterns.RustPatterns},
		{"Python Patterns", patterns.PythonPatterns},
		{"ML & Data Science", patterns.MLPatterns},
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

// writeDependencies writes only notable dependencies with context
// Skip this section in compact mode as it's less useful for token efficiency
func (g *ClaudeGenerator) writeDependencies(buf *bytes.Buffer, deps []types.Dependency) {
	// In compact mode, skip dependencies as they add tokens without much value
	if g.compact || len(deps) == 0 {
		return
	}

	// Only include notable dependencies that help understand the project
	// Skip generic/common packages that don't add context
	notablePatterns := map[string]string{
		// Databases
		"gorm":     "ORM",
		"sqlx":     "SQL",
		"mongo":    "MongoDB",
		"redis":    "Redis",
		"postgres": "PostgreSQL",
		"mysql":    "MySQL",
		"sqlite":   "SQLite",
		// Messaging
		"kafka":    "Kafka",
		"rabbitmq": "RabbitMQ",
		"nats":     "NATS",
		"pubsub":   "Pub/Sub",
		// Cloud
		"aws-sdk": "AWS",
		"azure":   "Azure",
		"gcloud":  "GCP",
		// Observability
		"prometheus":    "Metrics",
		"sentry":        "Error tracking",
		"opentelemetry": "Tracing",
		"elastic":       "APM",
		// Testing
		"testify":  "Testing",
		"gomock":   "Mocking",
		"httptest": "HTTP testing",
	}

	var notable []string
	seen := make(map[string]bool)

	for _, d := range deps {
		nameLower := strings.ToLower(d.Name)
		for pattern, category := range notablePatterns {
			if strings.Contains(nameLower, pattern) && !seen[category] {
				notable = append(notable, category)
				seen[category] = true
				break
			}
		}
	}

	// Only write section if we have notable dependencies
	if len(notable) > 0 {
		buf.WriteString("## Key Dependencies\n\n")
		sort.Strings(notable)
		buf.WriteString(strings.Join(notable, ", ") + "\n\n")
	}
}

// writeQuickReference writes a quick reference section for common commands
func (g *ClaudeGenerator) writeQuickReference(buf *bytes.Buffer, commands []types.Command) {
	if len(commands) == 0 {
		return
	}

	// Get prioritized commands (top 15 most important)
	maxCommands := 15
	if g.compact {
		maxCommands = 10
	}
	quickRef := detector.GetQuickReferenceCommands(commands, maxCommands)

	if len(quickRef) == 0 {
		return
	}

	// Group by category for organized output
	grouped := detector.GroupCommandsByCategory(quickRef)

	buf.WriteString("## Quick Reference\n\n")
	buf.WriteString("```bash\n")

	// Output in priority order
	categoryOrder := []string{"Build", "Test", "Lint", "Format", "Run", "Setup", "Clean", "Generate", "Deploy", "Docker", "Database", "Other"}
	for _, cat := range categoryOrder {
		cmds, ok := grouped[cat]
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

// Note: Command classification is now handled by detector.GroupCommandsByCategory

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

// toTitleCase converts the first letter of each word to uppercase
// This replaces the deprecated strings.Title function
func toTitleCase(s string) string {
	if s == "" {
		return s
	}
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			r := []rune(word)
			r[0] = unicode.ToUpper(r[0])
			words[i] = string(r)
		}
	}
	return strings.Join(words, " ")
}

// =============================================================================
// Compact Mode Methods
// =============================================================================

// writeArchitectureDiagramCompact writes a simplified architecture section
func (g *ClaudeGenerator) writeArchitectureDiagramCompact(buf *bytes.Buffer, arch *types.ArchitectureInfo) {
	if arch == nil || (arch.Style == "" && arch.EntryPoint == "") {
		return
	}

	buf.WriteString("## Architecture\n\n")

	if arch.Style != "" {
		fmt.Fprintf(buf, "**Style:** %s", arch.Style)
		if arch.EntryPoint != "" {
			fmt.Fprintf(buf, " | **Entry:** `%s`", arch.EntryPoint)
		}
		buf.WriteString("\n\n")
	} else if arch.EntryPoint != "" {
		fmt.Fprintf(buf, "**Entry Point:** `%s`\n\n", arch.EntryPoint)
	}

	// Skip diagram in compact mode - too verbose
}

// writeKeyFilesCompact writes only the most important key files (max 5)
func (g *ClaudeGenerator) writeKeyFilesCompact(buf *bytes.Buffer, keyFiles []types.KeyFile) {
	if len(keyFiles) == 0 {
		return
	}

	buf.WriteString("## Key Files\n\n")

	// Prioritize entry points, configs, and main files
	priority := []string{"main", "entry", "config", "readme", "contributing"}
	selected := make([]types.KeyFile, 0, 5)

	// First pass: get priority files
	for _, kf := range keyFiles {
		lower := strings.ToLower(kf.Path + kf.Purpose)
		for _, p := range priority {
			if strings.Contains(lower, p) {
				selected = append(selected, kf)
				break
			}
		}
		if len(selected) >= 5 {
			break
		}
	}

	// If not enough, fill with remaining
	if len(selected) < 5 {
		for _, kf := range keyFiles {
			found := false
			for _, s := range selected {
				if s.Path == kf.Path {
					found = true
					break
				}
			}
			if !found {
				selected = append(selected, kf)
				if len(selected) >= 5 {
					break
				}
			}
		}
	}

	for _, kf := range selected {
		if kf.Description != "" {
			fmt.Fprintf(buf, "- `%s` - %s\n", kf.Path, kf.Description)
		} else if kf.Purpose != "" {
			fmt.Fprintf(buf, "- `%s` - %s\n", kf.Path, kf.Purpose)
		} else {
			fmt.Fprintf(buf, "- `%s`\n", kf.Path)
		}
	}

	if len(keyFiles) > 5 {
		fmt.Fprintf(buf, "\n*...and %d more files*\n", len(keyFiles)-5)
	}
	buf.WriteString("\n")
}

// writeEndpointsCompact writes only unique endpoint patterns (max 10)
func (g *ClaudeGenerator) writeEndpointsCompact(buf *bytes.Buffer, endpoints []types.Endpoint) {
	if len(endpoints) == 0 {
		return
	}

	buf.WriteString("## API Endpoints\n\n")

	// Group by method and show unique paths
	methodPaths := make(map[string][]string)
	for _, ep := range endpoints {
		method := ep.Method
		if method == "" {
			method = "GET"
		}
		methodPaths[method] = append(methodPaths[method], ep.Path)
	}

	// Show summary
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	count := 0
	for _, method := range methods {
		paths := methodPaths[method]
		if len(paths) == 0 {
			continue
		}
		// Show max 3 examples per method
		limit := 3
		if len(paths) < limit {
			limit = len(paths)
		}
		fmt.Fprintf(buf, "**%s:** ", method)
		examples := make([]string, limit)
		for i := 0; i < limit; i++ {
			examples[i] = "`" + paths[i] + "`"
		}
		buf.WriteString(strings.Join(examples, ", "))
		if len(paths) > limit {
			fmt.Fprintf(buf, " *+%d more*", len(paths)-limit)
		}
		buf.WriteString("\n")
		count++
		if count >= 5 {
			break
		}
	}
	buf.WriteString("\n")
}

// writePatternsCompact writes only the top 5 most relevant patterns per category
func (g *ClaudeGenerator) writePatternsCompact(buf *bytes.Buffer, patterns *types.CodePatterns) {
	if patterns == nil {
		return
	}

	// Check if any patterns exist
	hasPatterns := len(patterns.StateManagement) > 0 ||
		len(patterns.DataFetching) > 0 ||
		len(patterns.Routing) > 0 ||
		len(patterns.Testing) > 0 ||
		len(patterns.Authentication) > 0 ||
		len(patterns.APIPatterns) > 0 ||
		len(patterns.DatabaseORM) > 0 ||
		len(patterns.GoPatterns) > 0

	if !hasPatterns {
		return
	}

	buf.WriteString("## Detected Patterns\n\n")
	buf.WriteString("*Top patterns detected in the codebase:*\n\n")

	// Helper to write limited patterns
	writeTopPatterns := func(title string, patterns []types.PatternInfo, limit int) {
		if len(patterns) == 0 {
			return
		}

		// Sort by file count (most common first)
		sorted := make([]types.PatternInfo, len(patterns))
		copy(sorted, patterns)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].FileCount > sorted[j].FileCount
		})

		buf.WriteString("### " + title + "\n\n")
		count := 0
		for _, p := range sorted {
			if count >= limit {
				break
			}
			if p.FileCount > 1 {
				fmt.Fprintf(buf, "- **%s** (%d files)\n", p.Name, p.FileCount)
			} else if len(p.Examples) > 0 {
				fmt.Fprintf(buf, "- **%s** - `%s`\n", p.Name, p.Examples[0])
			} else {
				fmt.Fprintf(buf, "- **%s**\n", p.Name)
			}
			count++
		}
		if len(patterns) > limit {
			fmt.Fprintf(buf, "\n*...and %d more*\n", len(patterns)-limit)
		}
		buf.WriteString("\n")
	}

	// Write only the most relevant categories with limited patterns
	writeTopPatterns("Go Patterns", patterns.GoPatterns, 3)
	writeTopPatterns("Data Fetching", patterns.DataFetching, 3)
	writeTopPatterns("Testing", patterns.Testing, 3)
	writeTopPatterns("API Patterns", patterns.APIPatterns, 3)
	writeTopPatterns("Database & ORM", patterns.DatabaseORM, 3)
}

// writeImports writes import references to .claude/ rules
// This uses Claude Code's @import syntax to reference external files
func (g *ClaudeGenerator) writeImports(buf *bytes.Buffer, analysis *types.Analysis) {
	// Build list of available rule imports
	var imports []string

	// Add standard rules that are typically generated
	if analysis.GitConventions != nil {
		imports = append(imports, "@.claude/rules/git-workflow.md")
	}
	if analysis.CodePatterns != nil && len(analysis.CodePatterns.Testing) > 0 {
		imports = append(imports, "@.claude/rules/testing.md")
	}
	if len(analysis.Conventions) > 0 {
		imports = append(imports, "@.claude/rules/coding-style.md")
	}
	if analysis.ArchitectureInfo != nil && analysis.ArchitectureInfo.Style != "" {
		imports = append(imports, "@.claude/rules/architecture.md")
	}
	// Security rules are always generated
	imports = append(imports, "@.claude/rules/security.md")

	if len(imports) == 0 {
		return
	}

	buf.WriteString("## Additional Rules\n\n")
	buf.WriteString("*The following rules are imported from `.claude/rules/` for context-specific guidance:*\n\n")

	for _, imp := range imports {
		fmt.Fprintf(buf, "- %s\n", imp)
	}
	buf.WriteString("\n")
}

// writeUsageInsights writes the AI usage insights section
func (g *ClaudeGenerator) writeUsageInsights(buf *bytes.Buffer, insights *types.UsageInsights) {
	if insights == nil {
		return
	}

	buf.WriteString("## AI Usage Insights\n\n")

	dateRange := fmt.Sprintf("%s - %s",
		insights.DateRange.Start.Format("Jan 2, 2006"),
		insights.DateRange.End.Format("Jan 2, 2006"))
	fmt.Fprintf(buf, "*Based on %d Claude Code sessions (%s)*\n\n", insights.SessionCount, dateRange)

	// Hot Files
	if len(insights.HotFiles) > 0 {
		buf.WriteString("### Hot Files\n\n")
		if g.compact {
			buf.WriteString("Files most accessed by AI — prioritize keeping them well-documented:\n")
		} else {
			buf.WriteString("These files are most frequently accessed by AI. Prioritize keeping them well-documented:\n")
		}

		limit := len(insights.HotFiles)
		if g.compact && limit > 5 {
			limit = 5
		} else if limit > 10 {
			limit = 10
		}

		for _, hf := range insights.HotFiles[:limit] {
			parts := []string{}
			if hf.ReadCount > 0 {
				parts = append(parts, fmt.Sprintf("%d reads", hf.ReadCount))
			}
			if hf.EditCount > 0 {
				parts = append(parts, fmt.Sprintf("%d edits", hf.EditCount))
			}
			if hf.WriteCount > 0 {
				parts = append(parts, fmt.Sprintf("%d writes", hf.WriteCount))
			}
			detail := strings.Join(parts, ", ")
			fmt.Fprintf(buf, "- `%s` — %d ops (%s)\n", hf.Path, hf.TotalOps, detail)
		}
		buf.WriteString("\n")
	}

	// Pain Points
	if len(insights.PainPoints) > 0 {
		buf.WriteString("### AI Pain Points\n\n")
		buf.WriteString("Files that cause AI difficulty — consider adding explicit conventions:\n")

		for _, pp := range insights.PainPoints {
			fmt.Fprintf(buf, "- `%s` — %s\n", pp.File, pp.Description)
		}
		buf.WriteString("\n")
	}
}
