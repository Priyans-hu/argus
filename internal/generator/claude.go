package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"

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
	buf.WriteString(fmt.Sprintf("# %s\n\n", analysis.ProjectName))

	// Tech Stack Summary
	g.writeTechStack(&buf, &analysis.TechStack)

	// Project Structure
	g.writeStructure(&buf, &analysis.Structure)

	// Key Files
	g.writeKeyFiles(&buf, analysis.KeyFiles)

	// Available Commands
	g.writeCommands(&buf, analysis.Commands)

	// Conventions
	g.writeConventions(&buf, analysis.Conventions)

	// Dependencies summary
	g.writeDependencies(&buf, analysis.Dependencies)

	return buf.Bytes(), nil
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
				buf.WriteString(fmt.Sprintf("- **%s** %s (%.1f%%)\n", lang.Name, lang.Version, lang.Percentage))
			} else {
				buf.WriteString(fmt.Sprintf("- **%s** (%.1f%%)\n", lang.Name, lang.Percentage))
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
				catName = strings.Title(cat)
			}

			buf.WriteString(fmt.Sprintf("**%s:**\n", catName))
			for _, fw := range fws {
				if fw.Version != "" {
					buf.WriteString(fmt.Sprintf("- %s %s\n", fw.Name, fw.Version))
				} else {
					buf.WriteString(fmt.Sprintf("- %s\n", fw.Name))
				}
			}
			buf.WriteString("\n")
		}
	}

	// Databases
	if len(stack.Databases) > 0 {
		buf.WriteString("### Databases\n\n")
		for _, db := range stack.Databases {
			buf.WriteString(fmt.Sprintf("- %s\n", db))
		}
		buf.WriteString("\n")
	}

	// Tools
	if len(stack.Tools) > 0 {
		buf.WriteString("### Tools\n\n")
		for _, tool := range stack.Tools {
			buf.WriteString(fmt.Sprintf("- %s\n", tool))
		}
		buf.WriteString("\n")
	}
}

// writeStructure writes the project structure section
func (g *ClaudeGenerator) writeStructure(buf *bytes.Buffer, structure *types.ProjectStructure) {
	buf.WriteString("## Project Structure\n\n")

	buf.WriteString("```\n")
	buf.WriteString(".\n")

	// Sort directories by path
	dirs := make([]types.Directory, len(structure.Directories))
	copy(dirs, structure.Directories)
	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].Path < dirs[j].Path
	})

	for _, dir := range dirs {
		if dir.Purpose != "" {
			buf.WriteString(fmt.Sprintf("├── %s/          # %s\n", dir.Path, dir.Purpose))
		} else {
			buf.WriteString(fmt.Sprintf("├── %s/\n", dir.Path))
		}
	}

	// Root files
	if len(structure.RootFiles) > 0 {
		rootFiles := make([]string, len(structure.RootFiles))
		copy(rootFiles, structure.RootFiles)
		sort.Strings(rootFiles)

		for i, f := range rootFiles {
			if i == len(rootFiles)-1 {
				buf.WriteString(fmt.Sprintf("└── %s\n", f))
			} else {
				buf.WriteString(fmt.Sprintf("├── %s\n", f))
			}
		}
	}

	buf.WriteString("```\n\n")
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
		buf.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", kf.Path, kf.Purpose, desc))
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
			buf.WriteString(fmt.Sprintf("# %s\n", cmd.Description))
		}
		buf.WriteString(fmt.Sprintf("%s\n", cmd.Name))
		if cmd.Command != "" && cmd.Command != cmd.Name {
			buf.WriteString(fmt.Sprintf("# → %s\n", cmd.Command))
		}
		buf.WriteString("\n")
	}

	buf.WriteString("```\n\n")
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
		buf.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(cat)))
		for _, conv := range convs {
			buf.WriteString(fmt.Sprintf("- %s\n", conv.Description))
			if conv.Example != "" {
				buf.WriteString(fmt.Sprintf("  ```\n  %s\n  ```\n", conv.Example))
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
			buf.WriteString(fmt.Sprintf("- `%s` %s\n", runtime[i].Name, runtime[i].Version))
		}
		if len(runtime) > 20 {
			buf.WriteString(fmt.Sprintf("\n*...and %d more*\n", len(runtime)-20))
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
			buf.WriteString(fmt.Sprintf("- `%s` %s\n", dev[i].Name, dev[i].Version))
		}
		if len(dev) > 15 {
			buf.WriteString(fmt.Sprintf("\n*...and %d more*\n", len(dev)-15))
		}
		buf.WriteString("\n")
	}
}

// claudeTemplate is an alternative template-based approach (not used currently)
var claudeTemplate = template.Must(template.New("claude").Parse(`# {{.ProjectName}}

## Tech Stack
{{range .TechStack.Languages}}
- **{{.Name}}** {{if .Version}}{{.Version}} {{end}}({{printf "%.1f" .Percentage}}%)
{{- end}}

## Project Structure

` + "```" + `
{{range .Structure.Directories}}├── {{.Path}}/{{if .Purpose}}          # {{.Purpose}}{{end}}
{{end}}` + "```" + `
`))
