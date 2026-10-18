package generator

import (
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// reviewerFixture is a project with testing context, the given languages
// and project test command.
func reviewerFixture(test string, langs ...types.Language) *types.Analysis {
	return &types.Analysis{
		ProjectName: "app",
		TechStack:   types.TechStack{Languages: langs},
		Commands:    []types.Command{{Name: "test", Command: test}},
		CodePatterns: &types.CodePatterns{Testing: []types.PatternInfo{
			{Name: "describe(", FileCount: 1, Examples: []string{"src/a.test.ts"}},
		}},
	}
}

func TestReviewerTestCommandFitsItsLanguage(t *testing.T) {
	tests := []struct {
		lang, test, agent, want string
	}{
		{"Rust", "vitest run", "rust-reviewer", "cargo test"},
		{"TypeScript", "vitest run", "ts-reviewer", "vitest run"},
		{"TypeScript", "pnpm run test", "ts-reviewer", "pnpm run test"},
		{"TypeScript", "eslint src/ tests/", "ts-reviewer", ""},
		{"Go", "vitest run", "go-reviewer", "go test ./..."},
		{"Python", "poetry run pytest -q", "python-reviewer", "pytest"},
		{"Python", "npm test", "python-reviewer", ""},
		{"Java", "./gradlew test", "java-reviewer", "./gradlew test"},
		{"Java", "npm test", "java-reviewer", ""},
	}
	for _, tt := range tests {
		t.Run(tt.agent+" "+tt.test, func(t *testing.T) {
			files := renderClaudeCode(t, reviewerFixture(tt.test, types.Language{Name: tt.lang, Percentage: 100}))
			content, ok := files[".claude/agents/"+tt.agent+".md"]
			if !ok {
				t.Fatalf("%s was not generated", tt.agent)
			}
			got := ""
			for _, line := range strings.Split(content, "\n") {
				if cmd, found := strings.CutPrefix(line, "Run tests: `"); found {
					got = strings.TrimSuffix(cmd, "`")
				}
			}
			if got != tt.want {
				t.Errorf("test line = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestJavaReviewerNeedsJavaCode(t *testing.T) {
	for _, pct := range []float64{11, 60} {
		analysis := reviewerFixture("./gradlew test", types.Language{Name: "Kotlin", Percentage: 100 - pct}, types.Language{Name: "Java", Percentage: pct})
		analysis.Structure.RootFiles = []string{"pubspec.yaml"}
		_, ok := renderClaudeCode(t, analysis)[".claude/agents/java-reviewer.md"]
		if ok != (pct >= 15) {
			t.Errorf("Java at %v%%: java-reviewer generated = %v", pct, ok)
		}
	}
}

func TestAgentsListNoSkills(t *testing.T) {
	analysis := renderFixture()
	analysis.Commands = nil
	for path, content := range renderClaudeCode(t, analysis) {
		if strings.HasPrefix(path, ".claude/agents/") && strings.Contains(content, "skills:") {
			t.Errorf("%s lists skills:\n%s", path, content)
		}
	}
}
