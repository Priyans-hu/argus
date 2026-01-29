package generator

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// WorkspaceInfo holds minimal info about a workspace for overview generation
type WorkspaceInfo struct {
	Path      string
	Name      string
	Languages []string
	Commands  int
	Endpoints int
}

// MonorepoOverviewGenerator creates a root-level overview file for monorepo projects
type MonorepoOverviewGenerator struct {
	format string // output format name (for reference)
}

// NewMonorepoOverviewGenerator creates a new overview generator
func NewMonorepoOverviewGenerator(format string) *MonorepoOverviewGenerator {
	return &MonorepoOverviewGenerator{format: format}
}

// Generate creates a root-level overview document referencing all workspaces
func (g *MonorepoOverviewGenerator) Generate(rootAnalysis *types.Analysis, workspaces []WorkspaceInfo) ([]byte, error) {
	var buf bytes.Buffer

	// Title
	fmt.Fprintf(&buf, "# %s (Monorepo)\n\n", rootAnalysis.ProjectName)

	// Project overview from README
	if rootAnalysis.ReadmeContent != nil && rootAnalysis.ReadmeContent.Description != "" {
		fmt.Fprintf(&buf, "%s\n\n", rootAnalysis.ReadmeContent.Description)
	}

	// Monorepo info
	if rootAnalysis.MonorepoInfo != nil {
		if rootAnalysis.MonorepoInfo.Tool != "" {
			fmt.Fprintf(&buf, "**Monorepo Tool:** %s\n", rootAnalysis.MonorepoInfo.Tool)
		}
		if rootAnalysis.MonorepoInfo.PackageManager != "" {
			fmt.Fprintf(&buf, "**Package Manager:** %s\n", rootAnalysis.MonorepoInfo.PackageManager)
		}
		fmt.Fprintln(&buf)
	}

	// Workspace table
	if len(workspaces) > 0 {
		fmt.Fprintln(&buf, "## Workspaces")
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "| Workspace | Path | Languages | Commands | Endpoints |")
		fmt.Fprintln(&buf, "|-----------|------|-----------|----------|-----------|")
		for _, ws := range workspaces {
			langs := "-"
			if len(ws.Languages) > 0 {
				langs = strings.Join(ws.Languages, ", ")
			}
			fmt.Fprintf(&buf, "| %s | `%s` | %s | %d | %d |\n",
				ws.Name, ws.Path, langs, ws.Commands, ws.Endpoints)
		}
		fmt.Fprintln(&buf)

		// Per-workspace context file reference
		fmt.Fprintln(&buf, "## Per-Workspace Context")
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "Each workspace has its own generated context files with detailed analysis:")
		fmt.Fprintln(&buf)
		for _, ws := range workspaces {
			fmt.Fprintf(&buf, "- `%s/CLAUDE.md`\n", ws.Path)
		}
		fmt.Fprintln(&buf)
	}

	// Quick reference: root-level commands
	if len(rootAnalysis.Commands) > 0 {
		fmt.Fprintln(&buf, "## Root Commands")
		fmt.Fprintln(&buf)
		fmt.Fprintln(&buf, "```bash")
		for _, cmd := range rootAnalysis.Commands {
			if cmd.Description != "" {
				fmt.Fprintf(&buf, "# %s\n", cmd.Description)
			}
			name := cmd.Name
			if cmd.Command != "" {
				name = cmd.Command
			}
			fmt.Fprintln(&buf, name)
		}
		fmt.Fprintln(&buf, "```")
		fmt.Fprintln(&buf)
	}

	// Tech stack summary
	g.writeTechStackSummary(&buf, rootAnalysis)

	// Shared conventions
	if len(rootAnalysis.Conventions) > 0 {
		fmt.Fprintln(&buf, "## Shared Conventions")
		fmt.Fprintln(&buf)
		for _, conv := range rootAnalysis.Conventions {
			fmt.Fprintf(&buf, "- %s\n", conv.Description)
		}
		fmt.Fprintln(&buf)
	}

	// Git conventions
	if rootAnalysis.GitConventions != nil {
		if rootAnalysis.GitConventions.CommitConvention != nil {
			c := rootAnalysis.GitConventions.CommitConvention
			fmt.Fprintln(&buf, "## Git Conventions")
			fmt.Fprintln(&buf)
			fmt.Fprintf(&buf, "- **Commit style:** %s\n", c.Style)
			if c.Format != "" {
				fmt.Fprintf(&buf, "- **Format:** `%s`\n", c.Format)
			}
			fmt.Fprintln(&buf)
		}
	}

	return buf.Bytes(), nil
}

func (g *MonorepoOverviewGenerator) writeTechStackSummary(buf *bytes.Buffer, analysis *types.Analysis) {
	ts := analysis.TechStack
	if len(ts.Languages) == 0 && len(ts.Frameworks) == 0 {
		return
	}

	fmt.Fprintln(buf, "## Tech Stack")
	fmt.Fprintln(buf)

	if len(ts.Languages) > 0 {
		var langs []string
		for _, l := range ts.Languages {
			if l.Version != "" {
				langs = append(langs, fmt.Sprintf("%s %s", l.Name, l.Version))
			} else {
				langs = append(langs, l.Name)
			}
		}
		fmt.Fprintf(buf, "**Languages:** %s\n", strings.Join(langs, ", "))
	}

	if len(ts.Frameworks) > 0 {
		var fws []string
		for _, f := range ts.Frameworks {
			fws = append(fws, f.Name)
		}
		fmt.Fprintf(buf, "**Frameworks:** %s\n", strings.Join(fws, ", "))
	}

	if len(ts.Databases) > 0 {
		fmt.Fprintf(buf, "**Databases:** %s\n", strings.Join(ts.Databases, ", "))
	}

	fmt.Fprintln(buf)
}
