package generator

import (
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestClaudeGenerator_Name(t *testing.T) {
	g := NewClaudeGenerator()
	if g.Name() != "claude" {
		t.Errorf("expected name 'claude', got '%s'", g.Name())
	}
}

func TestClaudeGenerator_OutputFile(t *testing.T) {
	g := NewClaudeGenerator()
	if g.OutputFile() != "CLAUDE.md" {
		t.Errorf("expected output file 'CLAUDE.md', got '%s'", g.OutputFile())
	}
}

func TestClaudeGenerator_Generate_Basic(t *testing.T) {
	g := NewClaudeGenerator()

	analysis := &types.Analysis{
		ProjectName: "test-project",
		RootPath:    "/test/path",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "Go", Version: "1.21", Percentage: 100},
			},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Check header
	if !strings.Contains(contentStr, "# test-project") {
		t.Error("expected project name header")
	}
}

func TestClaudeGenerator_Generate_Endpoints(t *testing.T) {
	g := NewClaudeGenerator()

	analysis := &types.Analysis{
		ProjectName: "test-project",
		Endpoints: []types.Endpoint{
			{Method: "GET", Path: "/api/users", File: "handlers/users.go", Line: 42},
			{Method: "POST", Path: "/api/users", File: "handlers/users.go", Line: 58},
			{Method: "GET", Path: "/api/posts", File: "handlers/posts.go", Line: 10},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "## API Endpoints") {
		t.Error("expected API Endpoints section")
	}

	if !strings.Contains(contentStr, "/api/users") {
		t.Error("expected /api/users endpoint")
	}

	if !strings.Contains(contentStr, "handlers/users.go:42") {
		t.Error("expected file:line reference")
	}
}

func TestClaudeGenerator_Generate_Dependencies(t *testing.T) {
	g := NewClaudeGenerator()

	analysis := &types.Analysis{
		ProjectName: "test-project",
		Dependencies: []types.Dependency{
			{Name: "gorm.io/gorm", Version: "v1.25.0", Type: "runtime"},
			{Name: "github.com/go-redis/redis/v8", Version: "v8.11.0", Type: "runtime"},
			{Name: "github.com/spf13/cobra", Version: "v1.8.0", Type: "runtime"},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "## Key Dependencies") {
		t.Error("expected Key Dependencies section")
	}

	if !strings.Contains(contentStr, "ORM") {
		t.Error("expected ORM in dependencies")
	}

	if !strings.Contains(contentStr, "Redis") {
		t.Error("expected Redis in dependencies")
	}
}

func TestClaudeGenerator_Generate_Imports(t *testing.T) {
	g := NewClaudeGenerator()

	analysis := &types.Analysis{
		ProjectName: "test-project",
		GitConventions: &types.GitConventions{
			Repository: &types.GitRepository{RemoteURL: "https://github.com/test/test"},
		},
		Conventions: []types.Convention{
			{Category: "code-style", Description: "Use gofmt"},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "## Additional Rules") {
		t.Error("expected Additional Rules section")
	}

	if !strings.Contains(contentStr, "@.claude/rules/git-workflow.md") {
		t.Error("expected git-workflow rule import")
	}

	if !strings.Contains(contentStr, "@.claude/rules/security.md") {
		t.Error("expected security rule import")
	}
}

func TestClaudeGenerator_Generate_LeanOutput(t *testing.T) {
	g := NewClaudeGenerator()

	analysis := &types.Analysis{
		ProjectName: "test-project",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "Go", Percentage: 100},
			},
		},
		Commands: []types.Command{
			{Command: "make build", Description: "Build"},
		},
		ConfigFiles: []types.ConfigFileInfo{
			{Path: "go.mod", Type: "Go Modules"},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Lean output should NOT have these verbose sections
	removedSections := []string{
		"## Tech Stack",
		"## Quick Reference",
		"## Architecture",
		"## Project Structure",
		"## Key Files",
		"## Configuration",
		"## Development Setup",
		"## Available Commands",
		"## CLI Output",
		"## Coding Conventions",
		"## Guidelines",
		"## Detected Patterns",
	}

	for _, section := range removedSections {
		if strings.Contains(contentStr, section) {
			t.Errorf("lean output should NOT have %s section", section)
		}
	}
}
