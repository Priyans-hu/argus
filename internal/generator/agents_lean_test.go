package generator

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestRenderLean_Golden(t *testing.T) {
	a := &types.Analysis{
		ProjectName: "argus",
		ReadmeContent: &types.ReadmeContent{
			Description:   "The all-seeing code analyzer.\n\nHelp AI grok your codebase.",
			Features:      []string{"Fast scans"},
			Prerequisites: []string{"Go 1.24"},
		},
		TechStack: types.TechStack{
			Languages:  []types.Language{{Name: "Shell", Percentage: 5}, {Name: "Go", Percentage: 92.4}, {Name: "Makefile", Percentage: 2.6}},
			Frameworks: []types.Framework{{Name: "Cobra", Version: "1.8"}, {Name: "Cobra"}},
		},
		ArchitectureInfo: &types.ArchitectureInfo{EntryPoint: "cmd/argus/main.go"},
		Commands: []types.Command{
			{Name: "make build", Description: "Build the binary"},
			{Name: "make test"},
			{Name: "make lint"},
			{Name: "go mod download"},
			{Name: "make deploy"},
		},
		Conventions: []types.Convention{
			{Category: "naming", Description: "Files use snake_case"},
			{Category: "testing", Description: "Table-driven tests", Example: "tests := []struct{}{}"},
			{Category: "custom", Description: "Keep handlers thin"},
		},
		GitConventions: &types.GitConventions{
			Repository: &types.GitRepository{Owner: "Priyans-hu", Name: "argus", RemoteURL: "https://github.com/Priyans-hu/argus", Platform: "github"},
			CommitConvention: &types.CommitConvention{
				Style: "conventional", Format: "<type>(<scope>): <description>", Types: []string{"feat", "fix"}, Example: "feat: add lean output",
			},
			BranchConvention: &types.BranchConvention{Prefixes: []string{"feat", "fix"}, Examples: []string{"feat/user-auth"}},
		},
		Endpoints: []types.Endpoint{
			{Method: "GET", Path: "/users/:id", File: "api/users.go"},
			{Method: "POST", Path: "/users", File: "api/users.go"},
			{Method: "GET", Path: "/health", File: "api/health.go"},
		},
	}
	want := strings.Join([]string{
		"# argus",
		"",
		"The all-seeing code analyzer. Help AI grok your codebase.",
		"",
		"## Commands",
		"",
		"```bash",
		"# Setup",
		"go mod download",
		"",
		"# Build",
		"make build",
		"",
		"# Test",
		"make test",
		"",
		"# Lint",
		"make lint",
		"```",
		"",
		"## Stack",
		"",
		"- Languages: Go 92%, Shell 5%",
		"- Frameworks: Cobra",
		"- Entry: cmd/argus/main.go",
		"",
		"## Conventions",
		"",
		"- Keep handlers thin",
		"- Table-driven tests",
		"",
		"## Git",
		"",
		"- Repository: Priyans-hu/argus",
		"- Commit style: Conventional Commits",
		"- Commit format: `<type>(<scope>): <description>`",
		"- Commit types: `feat`, `fix`",
		"- Branch prefixes: feat/, fix/",
		"",
		"## API",
		"",
		"- 3 endpoints",
		"- Defined in: api/health.go, api/users.go",
		"- Resources: /health, /users",
	}, "\n") + "\n"

	got := string(renderLean(a))
	if got != want {
		t.Errorf("renderLean mismatch\ngot:\n%s\nwant:\n%s", got, want)
	}
	if strings.Contains(got, "](") {
		t.Errorf("repository must be plain text, got a link:\n%s", got)
	}
}

func TestRenderLean_OversizedStaysCapped(t *testing.T) {
	cmdTemplates := []string{
		"npm install\npkg%d", "go build ./cmd/%d", "go test\n./pkg/%d", "golangci-lint run ./%d", "gofmt -w %d", "go run ./cmd/%d",
		"go clean -cache %d", "go generate ./%d", "kubectl apply -f %d", "docker build -t img%d .", "migrate up %d", "echo %d",
	}
	convCategories := []string{"custom", "naming", "imports", "testing", "logging", "structure"}
	many := func(n int, format string) []string {
		var out []string
		for i := range n {
			out = append(out, fmt.Sprintf(format, i))
		}
		return out
	}
	a := &types.Analysis{
		ProjectName:      "big",
		ReadmeContent:    &types.ReadmeContent{Description: strings.Repeat("Lorem ipsum dolor sit amet,\nconsectetur adipiscing elit. ", 34), Features: many(50, "feature %d")},
		TechStack:        types.TechStack{Databases: many(30, "db%d\nnext")},
		Structure:        types.ProjectStructure{Directories: []types.Directory{{Path: "cmd", Purpose: "entry"}}},
		KeyFiles:         []types.KeyFile{{Path: "main.go", Purpose: "entry"}},
		Dependencies:     []types.Dependency{{Name: "cobra", Version: "1.8"}},
		ArchitectureInfo: &types.ArchitectureInfo{EntryPoint: "cmd/\nmain.go"},
		MonorepoInfo:     &types.MonorepoInfo{IsMonorepo: true, Tool: "Turborepo", PackageManager: "pnpm", WorkspacePaths: many(50, "packages/p%d\n")},
		GitConventions: &types.GitConventions{
			Repository:       &types.GitRepository{Owner: "own\ner", Name: "repo"},
			CommitConvention: &types.CommitConvention{Style: "conventional", Format: "<type>:\n<desc>", Types: many(30, "t%d\nx"), Scopes: many(30, "s%d\nx"), Example: "feat: x\ny"},
			BranchConvention: &types.BranchConvention{Prefixes: many(20, "p%d\nx"), Examples: many(5, "p%d/x\ny")},
		},
	}
	for i := range 500 {
		a.Commands = append(a.Commands, types.Command{Name: fmt.Sprintf(cmdTemplates[i%len(cmdTemplates)], i)})
	}
	for i := range 300 {
		a.Conventions = append(a.Conventions, types.Convention{
			Category: convCategories[i%len(convCategories)], Description: fmt.Sprintf("Rule %d\nsecond line", i), Example: "code\nmore",
		})
	}
	for i := range 40 {
		a.TechStack.Languages = append(a.TechStack.Languages, types.Language{Name: fmt.Sprintf("Lang%d", i), Percentage: float64(i)})
	}
	for i := range 200 {
		a.TechStack.Frameworks = append(a.TechStack.Frameworks, types.Framework{Name: fmt.Sprintf("fw%d\nnext", i)})
	}
	for i := range 100 {
		a.Endpoints = append(a.Endpoints, types.Endpoint{Method: "GET", Path: fmt.Sprintf("/res%d/items/:id", i), File: fmt.Sprintf("api/h%d.go", i%20)})
	}

	got := string(renderLean(a))
	lines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	if len(lines) > 150 {
		t.Errorf("got %d lines, want at most 150", len(lines))
	}
	if len(lines[2]) > 400 || !utf8.ValidString(lines[2]) {
		t.Errorf("description line is %d bytes, want at most 400 valid UTF-8", len(lines[2]))
	}
	for _, s := range []string{"## Project Structure", "## Key Files", "## Dependencies", "@.claude/rules"} {
		if strings.Contains(got, s) {
			t.Errorf("output contains %q", s)
		}
	}
	if n := strings.Count(got, "```"); n != 2 {
		t.Errorf("got %d code fences, want exactly one pair", n)
	}
}

func TestRenderLean_CommandDedup(t *testing.T) {
	a := &types.Analysis{ProjectName: "p", Commands: []types.Command{
		{Name: "npm test (in a)"}, {Name: "npm test (in b)"}, {Name: "npm  test\n(in a)"}, {Name: "npm test (in a)"},
		{Name: "make build"}, {Name: "make build"},
		{Name: "make deploy"}, {Name: "docker build ."}, {Name: "make clean"}, {Name: "echo hi"},
	}}
	want := "# p\n\n## Commands\n\n```bash\n# Build\nmake build\n\n# Test\nnpm test (in a)\nnpm test (in b)\n```\n"
	if got := string(renderLean(a)); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestRenderLean_EmptyRendersTitleOnly(t *testing.T) {
	tests := []struct {
		name string
		a    *types.Analysis
	}{
		{"only name", &types.Analysis{ProjectName: "empty"}},
		{"empty optional parts", &types.Analysis{
			ProjectName:      "empty",
			ReadmeContent:    &types.ReadmeContent{Description: " \n "},
			MonorepoInfo:     &types.MonorepoInfo{WorkspacePaths: []string{"a"}},
			ArchitectureInfo: &types.ArchitectureInfo{},
			GitConventions:   &types.GitConventions{Repository: &types.GitRepository{Owner: "o"}},
			Commands:         []types.Command{{Name: "make deploy"}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(renderLean(tt.a)); got != "# empty\n" {
				t.Errorf("got %q, want only the title line", got)
			}
		})
	}
}

func TestLeanDescription(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"collapses whitespace", "  One.\n\nTwo  ", "One. Two"},
		{"cuts at last sentence end", "Wow! Fast? " + strings.Repeat("x", 395), "Wow! Fast?"},
		{"dot inside a word is no sentence end", "Built on Node.js " + strings.Repeat("x", 395), "Built on Node.js"},
		{"falls back to last space", strings.Repeat("word ", 100), strings.TrimSpace(strings.Repeat("word ", 80))},
		{"never splits a rune", "a" + strings.Repeat("é", 300), "a" + strings.Repeat("é", 199)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := leanDescription(tt.in)
			if got != tt.want || len(got) > 400 || !utf8.ValidString(got) {
				t.Errorf("got %q (%d bytes), want %q", got, len(got), tt.want)
			}
		})
	}
}
