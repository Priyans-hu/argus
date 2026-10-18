package generator

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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

	// Check tech stack section
	if !strings.Contains(contentStr, "## Tech Stack") {
		t.Error("expected Tech Stack section")
	}

	// Check Go language
	if !strings.Contains(contentStr, "Go") {
		t.Error("expected Go language in output")
	}
}

func TestClaudeGenerator_NormalModeHasAllSections(t *testing.T) {
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
		Dependencies: []types.Dependency{
			{Name: "github.com/spf13/cobra", Version: "v1.8.0", Type: "runtime"},
			{Name: "gorm.io/gorm", Version: "v1.25.0", Type: "runtime"},
			{Name: "github.com/go-redis/redis/v8", Version: "v8.11.0", Type: "runtime"},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Normal mode should have all major sections
	sections := []string{
		"## Tech Stack",
		"## Available Commands",
		"## Configuration",
		"## Key Dependencies",
	}

	for _, section := range sections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("normal mode should have %s section", section)
		}
	}
}

// extractSection extracts content from a markdown section
func extractSection(content, header string) string {
	start := strings.Index(content, header)
	if start == -1 {
		return ""
	}

	// Find next ## header or end of content
	rest := content[start+len(header):]
	end := strings.Index(rest, "\n## ")
	if end == -1 {
		return rest
	}
	return rest[:end]
}

// renderFixture returns an analysis that reaches the map-backed and tie-broken
// parts of every generator: many convention categories, equal language shares
// and pattern counts, a dependency matching two notable patterns, extra model
// specs and endpoints that differ only by file.
func renderFixture() *types.Analysis {
	return &types.Analysis{
		ProjectName: "shop",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "TypeScript", Percentage: 45},
				{Name: "Go", Version: "1.22", Percentage: 45},
				{Name: "Python", Percentage: 10},
			},
			Frameworks: []types.Framework{
				{Name: "Gin", Category: "backend"},
				{Name: "React", Version: "18", Category: "frontend"},
				{Name: "Vitest", Category: "testing"},
				{Name: "Cobra", Category: "cli"},
			},
			Databases: []string{"PostgreSQL", "Redis"},
			Tools:     []string{"Docker", "Make"},
		},
		Structure: types.ProjectStructure{
			Directories: []types.Directory{
				{Path: "api", Purpose: "API handlers", FileCount: 6},
				{Path: "cmd", Purpose: "Command entrypoints", FileCount: 1},
				{Path: "internal", FileCount: 3},
				{Path: "scripts", Purpose: "Scripts", FileCount: 2},
				{Path: "web", Purpose: "Frontend", FileCount: 8},
				{Path: "web/components", Purpose: "UI components", FileCount: 5},
			},
			RootFiles: []string{"Makefile", "README.md", "go.mod", "package.json"},
		},
		Conventions: []types.Convention{
			{Category: "react", Description: "Functional components with hooks"},
			{Category: "testing", Description: "Test files use _test suffix (Go style)", Example: "handler_test.go"},
			{Category: "gin", Description: "Gin HTTP handlers with gin.Context"},
			{Category: "naming", Description: "Files use kebab-case"},
			{Category: "custom", Description: "Keep handlers thin"},
			{Category: "error-handling", Description: "Go-style explicit error checking", Example: "if err != nil { return err }"},
			{Category: "logging", Description: "Structured logging with slog"},
			{Category: "code-style", Description: "Run gofmt before committing"},
			{Description: "Keep functions small"},
		},
		Dependencies: []types.Dependency{
			{Name: "gorm.io/driver/postgres", Version: "v1.5.0", Type: "runtime"},
			{Name: "github.com/redis/go-redis/v9", Version: "v9.0.0", Type: "runtime"},
			{Name: "github.com/stretchr/testify", Version: "v1.9.0", Type: "dev"},
		},
		Commands: []types.Command{
			{Name: "make build", Description: "Build the project"},
			{Name: "make test", Description: "Run tests"},
			{Name: "make lint", Description: "Run linter"},
			{Name: "npm run dev", Description: "Start dev server"},
		},
		KeyFiles: []types.KeyFile{
			{Path: "cmd/shop/main.go", Purpose: "Entry point", Description: "Go application entry"},
			{Path: "go.mod", Purpose: "Go module"},
			{Path: "package.json", Purpose: "Node package"},
			{Path: "vitest.config.ts", Purpose: "Test config"},
		},
		Endpoints: []types.Endpoint{
			{Method: "OPTIONS", Path: "/api/users", File: "api/users.go", Line: 50},
			{Method: "GET", Path: "/api/users", File: "api/users_v2.go", Line: 9},
			{Method: "HEAD", Path: "/api/users", File: "api/users.go", Line: 48},
			{Method: "POST", Path: "/api/users", File: "api/users.go", Line: 30},
			{Method: "GET", Path: "/api/users", File: "api/users.go", Line: 12},
			{Method: "DELETE", Path: "/api/orders/:id", File: "api/orders.go", Line: 40, Auth: "jwt"},
			{Method: "GET", Path: "/health", File: "api/health.go", Line: 5},
		},
		ReadmeContent: &types.ReadmeContent{
			Description: "An online shop.",
			Features:    []string{"Cart", "Checkout"},
			ModelSpecs: map[string]string{
				"vocab_size": "32000", "parameters": "7B", "license": "MIT", "heads": "32", "context_length": "4096",
			},
		},
		CodePatterns: &types.CodePatterns{
			Testing: []types.PatternInfo{
				{Name: "t.Run", Category: "testing", Description: "Go subtest", FileCount: 4, Examples: []string{"api/users_test.go"}},
				{Name: "func Test", Category: "testing", Description: "Go test function", FileCount: 4, Examples: []string{"api/orders_test.go", "api/users_test.go"}},
				{Name: "describe(", Category: "testing", Description: "Vitest suite", FileCount: 2, Examples: []string{"web/cart.test.ts"}},
				{Name: "httptest", Category: "testing", Description: "Go HTTP testing", FileCount: 1, Examples: []string{"api/health_test.go"}},
			},
			Authentication: []types.PatternInfo{
				{Name: "middleware", Category: "authentication", Description: "Auth middleware", FileCount: 3, Examples: []string{"api/middleware.go"}},
				{Name: "jwt.", Category: "authentication", Description: "JWT handling", FileCount: 2, Examples: []string{"api/auth.go"}},
				{Name: "Bearer", Category: "authentication", Description: "Bearer token", FileCount: 2, Examples: []string{"api/middleware.go"}},
			},
			APIPatterns: []types.PatternInfo{
				{Name: "grpc", Category: "api", Description: "gRPC protocol", FileCount: 3, Examples: []string{"api/grpc.go"}},
				{Name: "REST", Category: "api", Description: "RESTful API design", FileCount: 3, Examples: []string{"api/users.go"}},
				{Name: "OpenAPI", Category: "api", Description: "OpenAPI/Swagger spec", FileCount: 1, Examples: []string{"api/openapi.yaml"}},
			},
			GoPatterns: []types.PatternInfo{
				{Name: "fmt.Errorf", Category: "errors", Description: "Error formatting", FileCount: 5, Examples: []string{"api/orders.go"}},
				{Name: "sync.Mutex", Category: "concurrency", Description: "Mutex synchronization", FileCount: 1},
				{Name: "context.Context", Category: "concurrency", Description: "Context for cancellation", FileCount: 5, Examples: []string{"api/users.go"}},
				{Name: "go func", Category: "concurrency", Description: "Goroutines", FileCount: 2},
			},
			DataFetching: []types.PatternInfo{
				{Name: "useQuery", Category: "data-fetching", Description: "React Query", FileCount: 3},
				{Name: "fetch", Category: "data-fetching", Description: "Fetch API", FileCount: 3},
			},
			DatabaseORM: []types.PatternInfo{
				{Name: "gorm.Open", Category: "database", Description: "GORM ORM", FileCount: 1, Examples: []string{"internal/db.go"}},
			},
		},
		GitConventions: &types.GitConventions{
			CommitConvention: &types.CommitConvention{
				Style: "conventional", Format: "<type>(<scope>): <description>",
				Types: []string{"feat", "fix"}, Example: "feat(api): add users",
			},
			BranchConvention: &types.BranchConvention{
				Prefixes: []string{"feat", "fix"}, Format: "<prefix>/<description>", Examples: []string{"feat/cart"},
			},
		},
		ArchitectureInfo: &types.ArchitectureInfo{
			Style:      "Standard Go Layout",
			EntryPoint: "cmd/shop/main.go",
			Layers: []types.ArchitectureLayer{
				{Name: "cmd", Purpose: "Entry points", Packages: []string{"shop"}, DependsOn: []string{"api", "internal"}},
				{Name: "api", Purpose: "HTTP handlers", Packages: []string{"users", "orders"}, DependsOn: []string{"internal"}},
				{Name: "internal", Purpose: "Private packages", Packages: []string{"db"}},
			},
			Diagram: "```\ncmd\n │\napi\n```",
		},
		DevelopmentInfo: &types.DevelopmentInfo{
			Prerequisites: []types.Prerequisite{{Name: "Go", Version: "1.22+"}, {Name: "Node.js", Version: "20+"}},
			SetupSteps:    []types.SetupStep{{Description: "Install dependencies", Command: "make setup"}},
			GitHooks:      []types.GitHook{{Name: "pre-commit", Actions: []string{"Run linter", "Run tests"}}},
		},
		ConfigFiles: []types.ConfigFileInfo{
			{Path: "go.mod", Type: "Go Modules", Purpose: "Go dependencies"},
			{Path: "Makefile", Type: "Build", Purpose: "Make build automation"},
			{Path: ".github/workflows/", Type: "CI/CD", Purpose: "GitHub Actions workflows"},
		},
		CLIInfo: &types.CLIInfo{
			VerboseFlag: "-v, --verbose",
			DryRunFlag:  "-n, --dry-run",
			Indicators:  []types.Indicator{{Symbol: "✅", Meaning: "Success"}, {Symbol: "⚠️", Meaning: "Warning"}},
		},
	}
}

// assertInOrder fails unless each of want appears in s after the previous one.
func assertInOrder(t *testing.T, s string, want ...string) {
	t.Helper()
	rest := s
	for _, w := range want {
		i := strings.Index(rest, w)
		if i < 0 {
			t.Errorf("%q is missing or out of order in:\n%s", w, s)
			return
		}
		rest = rest[i+len(w):]
	}
}

func TestGeneratorsRenderIdenticalBytes(t *testing.T) {
	tests := []struct {
		name string
		gen  interface {
			Generate(*types.Analysis) ([]byte, error)
		}
	}{
		{"claude", NewClaudeGenerator()},
		{"cursor", NewCursorGenerator()},
		{"copilot", NewCopilotGenerator()},
		{"continue", NewContinueGenerator()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := renderFixture()
			first, err := tt.gen.Generate(analysis)
			if err != nil {
				t.Fatal(err)
			}
			for i := 1; i < 20; i++ {
				got, err := tt.gen.Generate(analysis)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, first) {
					t.Fatalf("render %d differs from the first:\n%s\n--- first ---\n%s", i, got, first)
				}
			}
		})
	}
}

func TestClaudeRendersMapsAndTiesInFixedOrder(t *testing.T) {
	content, err := NewClaudeGenerator().Generate(renderFixture())
	if err != nil {
		t.Fatal(err)
	}
	out := string(content)

	t.Run("model specs", func(t *testing.T) {
		assertInOrder(t, out, "- **Parameters:** 7B", "- **Heads:** 32", "- **Context Length:** 4096",
			"- **License:** MIT", "- **Vocab Size:** 32000")
	})
	t.Run("languages with equal shares", func(t *testing.T) {
		assertInOrder(t, out, "- **Go** 1.22 (45.0%)", "- **TypeScript** (45.0%)", "- **Python** (10.0%)")
	})
	t.Run("convention categories", func(t *testing.T) {
		assertInOrder(t, out, "### Naming", "### Testing", "### Code-style", "### Logging",
			"### Error-handling", "### Custom", "### General", "### Gin", "### React", "### Git")
	})
	t.Run("endpoints", func(t *testing.T) {
		assertInOrder(t, out,
			"| GET | `/api/users` | `api/users.go:12` |",
			"| GET | `/api/users` | `api/users_v2.go:9` |",
			"| POST | `/api/users` | `api/users.go:30` |",
			"| HEAD | `/api/users` | `api/users.go:48` |",
			"| OPTIONS | `/api/users` | `api/users.go:50` |")
	})
	t.Run("dependency matching two patterns", func(t *testing.T) {
		// gorm.io/driver/postgres matches "gorm" first
		if section := extractSection(out, "## Key Dependencies"); strings.TrimSpace(section) != "ORM, Redis, Testing" {
			t.Errorf("Key Dependencies = %q, want ORM, Redis, Testing", section)
		}
	})
}

func TestCursorPrimaryLanguageTieGoesToName(t *testing.T) {
	content, err := NewCursorGenerator().Generate(renderFixture())
	if err != nil {
		t.Fatal(err)
	}
	assertInOrder(t, string(content), "- Primary Language: Go 1.22\n", "- Other Languages: TypeScript, Python\n")
}

func TestClaudeRepositoryLine(t *testing.T) {
	commit := &types.CommitConvention{Style: "conventional", Format: "<type>(<scope>): <description>"}
	tests := []struct {
		name   string
		git    *types.GitConventions
		want   string // the Repository line, "" for none
		header bool   // whether a Git heading is written
	}{
		{
			name: "known host is linked",
			git: &types.GitConventions{Repository: &types.GitRepository{
				RemoteURL: "https://gitlab.com/g/sub/n", Owner: "g/sub", Name: "n", Platform: "gitlab"}},
			want:   "- Repository: [g/sub/n](https://gitlab.com/g/sub/n)\n",
			header: true,
		},
		{
			name:   "other host is plain",
			git:    &types.GitConventions{Repository: &types.GitRepository{Owner: "o", Name: "n"}},
			want:   "- Repository: o/n\n",
			header: true,
		},
		{
			name: "no owner and name, no line and no heading",
			git:  &types.GitConventions{Repository: &types.GitRepository{}},
		},
		{
			name:   "commit conventions without a repository",
			git:    &types.GitConventions{CommitConvention: commit},
			header: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := NewClaudeGenerator().Generate(&types.Analysis{ProjectName: "n", GitConventions: tt.git})
			if err != nil {
				t.Fatal(err)
			}
			out := string(content)
			if tt.want != "" && !strings.Contains(out, "### Git\n\n"+tt.want) {
				t.Errorf("missing %q under the Git heading in:\n%s", tt.want, out)
			}
			if tt.want == "" && strings.Contains(out, "Repository:") {
				t.Errorf("unexpected Repository line in:\n%s", out)
			}
			if got := strings.Contains(out, "### Git\n"); got != tt.header {
				t.Errorf("Git heading written = %v, want %v:\n%s", got, tt.header, out)
			}
		})
	}
}

func TestClaudeRemoteCredentialsNeverReachOutput(t *testing.T) {
	secret := strings.Repeat("k4", 4)
	remotes := []struct {
		url, want string
	}{
		{"https://" + "deploy:" + secret + "@" + "github.com/o/n.git", "- Repository: [o/n](https://github.com/o/n)\n"},
		{"git@github-personal:o/n.git", "- Repository: o/n\n"},
	}
	for _, remote := range remotes {
		dir := t.TempDir()
		repo, err := git.PlainInit(dir, false)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := repo.CreateRemote(&config.RemoteConfig{Name: "origin", URLs: []string{remote.url}}); err != nil {
			t.Fatal(err)
		}

		analysis := &types.Analysis{ProjectName: "n", GitConventions: detector.NewGitDetectorGoGit(dir).Detect()}
		data, err := json.Marshal(analysis.GitConventions)
		if err != nil {
			t.Fatal(err)
		}
		content, err := NewClaudeGenerator().Generate(analysis)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(content), remote.want) {
			t.Errorf("CLAUDE.md lacks %q:\n%s", remote.want, content)
		}
		for name, text := range map[string]string{"JSON": string(data), "CLAUDE.md": string(content)} {
			for _, leak := range []string{secret, "deploy:", "git@", "github-personal"} {
				if strings.Contains(text, leak) {
					t.Errorf("%s contains %q from the remote:\n%s", name, leak, text)
				}
			}
		}
	}
}
