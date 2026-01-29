package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// CopilotGenerator generates .github/copilot-instructions.md files
type CopilotGenerator struct{}

// NewCopilotGenerator creates a new Copilot generator
func NewCopilotGenerator() *CopilotGenerator {
	return &CopilotGenerator{}
}

// Name returns the generator name
func (g *CopilotGenerator) Name() string {
	return "copilot"
}

// OutputFile returns the output filename
func (g *CopilotGenerator) OutputFile() string {
	return ".github/copilot-instructions.md"
}

// Generate creates the copilot-instructions.md content
func (g *CopilotGenerator) Generate(analysis *types.Analysis) ([]byte, error) {
	var buf bytes.Buffer

	// Title
	fmt.Fprintf(&buf, "# %s - Copilot Instructions\n\n", analysis.ProjectName)

	// Overview
	g.writeOverview(&buf, analysis)

	// Tech stack
	g.writeTechStack(&buf, &analysis.TechStack)

	// Architecture
	g.writeArchitecture(&buf, &analysis.Structure)

	// Coding standards
	g.writeCodingStandards(&buf, analysis.Conventions, &analysis.TechStack)

	// Patterns to follow
	g.writePatterns(&buf, analysis.Conventions)

	// Don'ts - common mistakes to avoid
	g.writeDonts(&buf, &analysis.TechStack)

	// AI enrichment
	g.writeAIInsights(&buf, analysis.AIEnrichment)

	return buf.Bytes(), nil
}

// writeOverview writes project overview
func (g *CopilotGenerator) writeOverview(buf *bytes.Buffer, analysis *types.Analysis) {
	buf.WriteString("## Overview\n\n")

	// Determine project type from tech stack
	projectType := g.determineProjectType(&analysis.TechStack)
	fmt.Fprintf(buf, "This is a %s project", projectType)

	if len(analysis.TechStack.Languages) > 0 {
		fmt.Fprintf(buf, " primarily written in %s", analysis.TechStack.Languages[0].Name)
	}

	buf.WriteString(".\n\n")
}

func (g *CopilotGenerator) determineProjectType(stack *types.TechStack) string {
	// Check frameworks for project type
	for _, fw := range stack.Frameworks {
		name := strings.ToLower(fw.Name)
		switch {
		case strings.Contains(name, "react") || strings.Contains(name, "vue") || strings.Contains(name, "angular"):
			return "frontend"
		case strings.Contains(name, "next") || strings.Contains(name, "nuxt"):
			return "full-stack"
		case strings.Contains(name, "express") || strings.Contains(name, "fastapi") || strings.Contains(name, "spring") || strings.Contains(name, "gin"):
			return "backend"
		case strings.Contains(name, "cobra") || strings.Contains(name, "click"):
			return "CLI"
		}
	}

	// Check languages as fallback
	if len(stack.Languages) > 0 {
		lang := strings.ToLower(stack.Languages[0].Name)
		switch lang {
		case "go", "rust", "java", "python":
			return "backend"
		case "typescript", "javascript":
			return "web"
		}
	}

	return "software"
}

// writeTechStack writes tech stack section
func (g *CopilotGenerator) writeTechStack(buf *bytes.Buffer, stack *types.TechStack) {
	buf.WriteString("## Tech Stack\n\n")

	// Languages
	if len(stack.Languages) > 0 {
		buf.WriteString("**Languages:**\n")
		for _, lang := range stack.Languages {
			if lang.Version != "" {
				fmt.Fprintf(buf, "- %s %s\n", lang.Name, lang.Version)
			} else {
				fmt.Fprintf(buf, "- %s\n", lang.Name)
			}
		}
		buf.WriteString("\n")
	}

	// Frameworks
	if len(stack.Frameworks) > 0 {
		buf.WriteString("**Frameworks/Libraries:**\n")
		for _, fw := range stack.Frameworks {
			if fw.Version != "" {
				fmt.Fprintf(buf, "- %s %s\n", fw.Name, fw.Version)
			} else {
				fmt.Fprintf(buf, "- %s\n", fw.Name)
			}
		}
		buf.WriteString("\n")
	}

	// Databases
	if len(stack.Databases) > 0 {
		buf.WriteString("**Databases:**\n")
		for _, db := range stack.Databases {
			fmt.Fprintf(buf, "- %s\n", db)
		}
		buf.WriteString("\n")
	}
}

// writeArchitecture writes the architecture section
func (g *CopilotGenerator) writeArchitecture(buf *bytes.Buffer, structure *types.ProjectStructure) {
	if len(structure.Directories) == 0 {
		return
	}

	buf.WriteString("## Architecture\n\n")

	// Sort directories
	dirs := make([]types.Directory, len(structure.Directories))
	copy(dirs, structure.Directories)
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Path < dirs[j].Path
	})

	buf.WriteString("**Directory Structure:**\n")
	for _, dir := range dirs {
		if dir.Purpose != "" {
			fmt.Fprintf(buf, "- `%s/` - %s\n", dir.Path, dir.Purpose)
		}
	}
	buf.WriteString("\n")
}

// writeCodingStandards writes coding standards based on tech stack
func (g *CopilotGenerator) writeCodingStandards(buf *bytes.Buffer, conventions []types.Convention, stack *types.TechStack) {
	buf.WriteString("## Coding Standards\n\n")

	// Extract code style conventions
	for _, conv := range conventions {
		if conv.Category == "code-style" || conv.Category == "naming" {
			fmt.Fprintf(buf, "- %s\n", conv.Description)
		}
	}

	// Add language-specific standards
	if len(stack.Languages) > 0 {
		primary := strings.ToLower(stack.Languages[0].Name)
		g.writeLanguageStandards(buf, primary)
	}

	buf.WriteString("\n")
}

func (g *CopilotGenerator) writeLanguageStandards(buf *bytes.Buffer, language string) {
	switch language {
	case "go":
		buf.WriteString("- Use `gofmt` for formatting\n")
		buf.WriteString("- Always handle errors explicitly\n")
		buf.WriteString("- Export only what needs to be public\n")
	case "typescript", "javascript":
		buf.WriteString("- Use camelCase for variables and functions\n")
		buf.WriteString("- Use PascalCase for classes and types\n")
		buf.WriteString("- Prefer const over let\n")
	case "python":
		buf.WriteString("- Follow PEP 8 style guide\n")
		buf.WriteString("- Use type hints for function signatures\n")
		buf.WriteString("- Use snake_case for functions and variables\n")
	case "java":
		buf.WriteString("- Follow Java naming conventions\n")
		buf.WriteString("- Use camelCase for methods\n")
		buf.WriteString("- Add Javadoc to public methods\n")
	case "rust":
		buf.WriteString("- Use `cargo fmt` for formatting\n")
		buf.WriteString("- Handle Result/Option properly, avoid unwrap()\n")
		buf.WriteString("- Use snake_case for functions\n")
	case "c#":
		buf.WriteString("- Follow .NET naming conventions\n")
		buf.WriteString("- Use PascalCase for public members\n")
		buf.WriteString("- Add XML documentation to public APIs\n")
	}
}

// writePatterns writes patterns to follow
func (g *CopilotGenerator) writePatterns(buf *bytes.Buffer, conventions []types.Convention) {
	// Collect relevant patterns
	var patterns []types.Convention
	for _, conv := range conventions {
		switch conv.Category {
		case "error-handling", "logging", "architecture", "async", "git", "testing":
			patterns = append(patterns, conv)
		}
	}

	if len(patterns) == 0 {
		return
	}

	buf.WriteString("## Patterns to Follow\n\n")

	// Group by category
	byCategory := make(map[string][]types.Convention)
	for _, p := range patterns {
		byCategory[p.Category] = append(byCategory[p.Category], p)
	}

	categoryOrder := []string{"error-handling", "logging", "architecture", "async", "testing", "git"}

	for _, cat := range categoryOrder {
		convs, ok := byCategory[cat]
		if !ok || len(convs) == 0 {
			continue
		}

		fmt.Fprintf(buf, "**%s:**\n", titleCase(strings.ReplaceAll(cat, "-", " ")))
		for _, conv := range convs {
			fmt.Fprintf(buf, "- %s\n", conv.Description)
		}
		buf.WriteString("\n")
	}
}

// writeDonts writes common mistakes to avoid
func (g *CopilotGenerator) writeDonts(buf *bytes.Buffer, stack *types.TechStack) {
	buf.WriteString("## Avoid\n\n")

	// Common things to avoid
	buf.WriteString("- Don't add unnecessary comments for obvious code\n")
	buf.WriteString("- Don't ignore errors or use empty catch blocks\n")
	buf.WriteString("- Don't commit sensitive data or credentials\n")
	buf.WriteString("- Don't introduce breaking changes without discussion\n")

	// Language-specific don'ts
	if len(stack.Languages) > 0 {
		primary := strings.ToLower(stack.Languages[0].Name)
		switch primary {
		case "go":
			buf.WriteString("- Don't use panic() for regular error handling\n")
			buf.WriteString("- Don't use global state unnecessarily\n")
		case "typescript", "javascript":
			buf.WriteString("- Don't use `any` type without good reason\n")
			buf.WriteString("- Don't use var, prefer const/let\n")
		case "python":
			buf.WriteString("- Don't use mutable default arguments\n")
			buf.WriteString("- Don't catch generic Exception without re-raising\n")
		case "rust":
			buf.WriteString("- Don't use unwrap() in production code\n")
			buf.WriteString("- Don't ignore compiler warnings\n")
		case "java":
			buf.WriteString("- Don't catch Exception without proper handling\n")
			buf.WriteString("- Don't expose internal implementation details\n")
		}
	}

	buf.WriteString("\n")
}

// writeAIInsights writes AI-enriched insights if available
func (g *CopilotGenerator) writeAIInsights(buf *bytes.Buffer, enrichment *types.AIEnrichment) {
	if enrichment == nil {
		return
	}

	buf.WriteString("## AI Insights\n\n")

	if enrichment.ProjectSummary != "" {
		fmt.Fprintf(buf, "%s\n\n", enrichment.ProjectSummary)
	}

	if len(enrichment.BestPractices) > 0 {
		buf.WriteString("**Best Practices:**\n")
		for _, bp := range enrichment.BestPractices {
			fmt.Fprintf(buf, "- %s: %s\n", bp.Title, bp.Description)
		}
		buf.WriteString("\n")
	}
}
