package generator

import (
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestMonorepoOverviewGenerator_Generate(t *testing.T) {
	analysis := &types.Analysis{
		ProjectName: "my-monorepo",
		MonorepoInfo: &types.MonorepoInfo{
			IsMonorepo:     true,
			Tool:           "Turborepo",
			PackageManager: "pnpm",
		},
		Commands: []types.Command{
			{Name: "pnpm build", Description: "Build all packages"},
			{Name: "pnpm test", Description: "Run tests"},
		},
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "TypeScript", Version: "5.0"},
			},
			Frameworks: []types.Framework{
				{Name: "React", Category: "frontend"},
				{Name: "Next.js", Category: "frontend"},
			},
		},
		Conventions: []types.Convention{
			{Category: "naming", Description: "Use camelCase for variables"},
		},
	}

	workspaces := []WorkspaceInfo{
		{
			Path:      "apps/web",
			Name:      "@myapp/web",
			Languages: []string{"TypeScript"},
			Commands:  5,
			Endpoints: 3,
		},
		{
			Path:      "apps/api",
			Name:      "@myapp/api",
			Languages: []string{"TypeScript"},
			Commands:  4,
			Endpoints: 12,
		},
		{
			Path:      "packages/shared",
			Name:      "@myapp/shared",
			Languages: []string{"TypeScript"},
			Commands:  2,
			Endpoints: 0,
		},
	}

	gen := NewMonorepoOverviewGenerator("claude")
	content, err := gen.Generate(analysis, workspaces)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	output := string(content)

	// Check title
	if !strings.Contains(output, "# my-monorepo (Monorepo)") {
		t.Error("missing project title")
	}

	// Check monorepo info
	if !strings.Contains(output, "Turborepo") {
		t.Error("missing monorepo tool")
	}
	if !strings.Contains(output, "pnpm") {
		t.Error("missing package manager")
	}

	// Check workspace table
	if !strings.Contains(output, "## Workspaces") {
		t.Error("missing workspaces section")
	}
	if !strings.Contains(output, "@myapp/web") {
		t.Error("missing workspace @myapp/web in table")
	}
	if !strings.Contains(output, "apps/api") {
		t.Error("missing workspace path apps/api in table")
	}

	// Check per-workspace context references
	if !strings.Contains(output, "apps/web/CLAUDE.md") {
		t.Error("missing per-workspace CLAUDE.md reference")
	}

	// Check root commands
	if !strings.Contains(output, "## Root Commands") {
		t.Error("missing root commands section")
	}
	if !strings.Contains(output, "pnpm build") {
		t.Error("missing build command")
	}

	// Check tech stack
	if !strings.Contains(output, "## Tech Stack") {
		t.Error("missing tech stack section")
	}
	if !strings.Contains(output, "TypeScript 5.0") {
		t.Error("missing language version")
	}

	// Check shared conventions
	if !strings.Contains(output, "## Shared Conventions") {
		t.Error("missing shared conventions section")
	}
	if !strings.Contains(output, "camelCase") {
		t.Error("missing convention content")
	}
}

func TestMonorepoOverviewGenerator_EmptyWorkspaces(t *testing.T) {
	analysis := &types.Analysis{
		ProjectName: "empty-mono",
	}

	gen := NewMonorepoOverviewGenerator("claude")
	content, err := gen.Generate(analysis, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "# empty-mono (Monorepo)") {
		t.Error("missing project title")
	}

	// Should not have workspace sections
	if strings.Contains(output, "## Workspaces") {
		t.Error("should not have workspaces section when empty")
	}
}

func TestMonorepoOverviewGenerator_WithGitConventions(t *testing.T) {
	analysis := &types.Analysis{
		ProjectName: "git-mono",
		GitConventions: &types.GitConventions{
			CommitConvention: &types.CommitConvention{
				Style:  "conventional",
				Format: "<type>(<scope>): <description>",
			},
		},
	}

	gen := NewMonorepoOverviewGenerator("claude")
	content, err := gen.Generate(analysis, nil)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "## Git Conventions") {
		t.Error("missing git conventions section")
	}
	if !strings.Contains(output, "conventional") {
		t.Error("missing commit style")
	}
}
