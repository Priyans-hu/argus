package generator

import (
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestContinueGenerator_Name(t *testing.T) {
	g := NewContinueGenerator()
	if g.Name() != "continue" {
		t.Errorf("expected name 'continue', got '%s'", g.Name())
	}
}

func TestContinueGenerator_OutputFile(t *testing.T) {
	g := NewContinueGenerator()
	if g.OutputFile() != ".continue/config.yaml" {
		t.Errorf("expected output file '.continue/config.yaml', got '%s'", g.OutputFile())
	}
}

func TestContinueGenerator_Generate_Basic(t *testing.T) {
	g := NewContinueGenerator()

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
	if !strings.Contains(contentStr, "# Continue.dev Configuration") {
		t.Error("expected header comment")
	}

	// Check project name (may be quoted due to hyphen)
	if !strings.Contains(contentStr, "name: test-project") && !strings.Contains(contentStr, "name: \"test-project\"") {
		t.Error("expected project name in config")
	}

	// Check version and schema
	if !strings.Contains(contentStr, "version: \"1.0.0\"") {
		t.Error("expected version in config")
	}
	if !strings.Contains(contentStr, "schema: v1") {
		t.Error("expected schema in config")
	}

	// Check rules section exists
	if !strings.Contains(contentStr, "rules:") {
		t.Error("expected rules section")
	}

	// Check context providers
	if !strings.Contains(contentStr, "context:") {
		t.Error("expected context section")
	}
	if !strings.Contains(contentStr, "provider: code") {
		t.Error("expected code context provider")
	}
}

func TestContinueGenerator_Generate_WithFrameworks(t *testing.T) {
	g := NewContinueGenerator()

	analysis := &types.Analysis{
		ProjectName: "react-app",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "TypeScript", Version: "5.0", Percentage: 80},
				{Name: "JavaScript", Percentage: 20},
			},
			Frameworks: []types.Framework{
				{Name: "React", Version: "18.0"},
				{Name: "Next.js", Version: "14.0"},
			},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Check framework rules
	if !strings.Contains(contentStr, "Frameworks in use") {
		t.Error("expected frameworks rule")
	}

	// Check React-specific rules
	if !strings.Contains(contentStr, "functional components") {
		t.Error("expected React functional components rule")
	}

	// Check docs section for React
	if !strings.Contains(contentStr, "docs:") {
		t.Error("expected docs section for React project")
	}
	if !strings.Contains(contentStr, "react.dev") {
		t.Error("expected React docs URL")
	}
}

func TestContinueGenerator_Generate_WithGitConventions(t *testing.T) {
	g := NewContinueGenerator()

	analysis := &types.Analysis{
		ProjectName: "git-project",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "Go", Percentage: 100},
			},
		},
		GitConventions: &types.GitConventions{
			CommitConvention: &types.CommitConvention{
				Style:  "conventional",
				Format: "<type>(<scope>): <description>",
			},
			BranchConvention: &types.BranchConvention{
				Format:   "<prefix>/<description>",
				Prefixes: []string{"feat", "fix", "chore"},
			},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Check git convention rules
	if !strings.Contains(contentStr, "conventional") {
		t.Error("expected commit style rule")
	}
	if !strings.Contains(contentStr, "Branch naming") {
		t.Error("expected branch naming rule")
	}
}

func TestContinueGenerator_Generate_WithConventions(t *testing.T) {
	g := NewContinueGenerator()

	analysis := &types.Analysis{
		ProjectName: "conv-project",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "Python", Percentage: 100},
			},
		},
		Conventions: []types.Convention{
			{Category: "testing", Description: "Use pytest for testing"},
			{Category: "style", Description: "Follow PEP 8"},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Check Python-specific rules
	if !strings.Contains(contentStr, "PEP 8") {
		t.Error("expected PEP 8 rule for Python")
	}
	if !strings.Contains(contentStr, "type hints") {
		t.Error("expected type hints rule for Python")
	}
}

func TestContinueGenerator_Generate_LargeProject(t *testing.T) {
	g := NewContinueGenerator()

	// Create a project with many directories (should trigger codebase provider)
	dirs := make([]types.Directory, 10)
	for i := 0; i < 10; i++ {
		dirs[i] = types.Directory{Path: "dir" + string(rune('a'+i))}
	}

	analysis := &types.Analysis{
		ProjectName: "large-project",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "Go", Percentage: 100},
			},
		},
		Structure: types.ProjectStructure{
			Directories: dirs,
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Large projects should have codebase provider
	if !strings.Contains(contentStr, "provider: codebase") {
		t.Error("expected codebase provider for large project")
	}
}

func TestContinueGenerator_Generate_AllLanguages(t *testing.T) {
	languages := []struct {
		name     string
		expected string
	}{
		{"Go", "gofmt"},
		{"TypeScript", "const"},
		{"JavaScript", "const"},
		{"Python", "PEP 8"},
		{"Rust", "Result"},
		{"Java", "Javadoc"},
		{"Ruby", "Ruby style"},
		{"C#", ".NET"},
	}

	for _, tc := range languages {
		t.Run(tc.name, func(t *testing.T) {
			g := NewContinueGenerator()

			analysis := &types.Analysis{
				ProjectName: "test-" + tc.name,
				TechStack: types.TechStack{
					Languages: []types.Language{
						{Name: tc.name, Percentage: 100},
					},
				},
			}

			content, err := g.Generate(analysis)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, tc.expected) {
				t.Errorf("expected '%s' rule for %s language", tc.expected, tc.name)
			}
		})
	}
}

func TestContinueGenerator_Generate_DocsSection(t *testing.T) {
	tests := []struct {
		framework   string
		expectedDoc string
	}{
		{"React", "react.dev"},
		{"Next.js", "nextjs.org"},
		{"Vue", "vuejs.org"},
		{"Express", "expressjs.com"},
		{"FastAPI", "fastapi.tiangolo.com"},
	}

	for _, tc := range tests {
		t.Run(tc.framework, func(t *testing.T) {
			g := NewContinueGenerator()

			analysis := &types.Analysis{
				ProjectName: "test-" + tc.framework,
				TechStack: types.TechStack{
					Languages: []types.Language{
						{Name: "JavaScript", Percentage: 100},
					},
					Frameworks: []types.Framework{
						{Name: tc.framework},
					},
				},
			}

			content, err := g.Generate(analysis)
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			contentStr := string(content)
			if !strings.Contains(contentStr, tc.expectedDoc) {
				t.Errorf("expected docs URL containing '%s' for %s", tc.expectedDoc, tc.framework)
			}
		})
	}
}

func TestContinueMultiFileGenerator_Name(t *testing.T) {
	g := NewContinueMultiFileGenerator()
	if g.Name() != "continue-full" {
		t.Errorf("expected name 'continue-full', got '%s'", g.Name())
	}
}

func TestContinueMultiFileGenerator_Generate(t *testing.T) {
	g := NewContinueMultiFileGenerator()

	analysis := &types.Analysis{
		ProjectName: "multi-test",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "Go", Version: "1.21", Percentage: 100},
			},
			Frameworks: []types.Framework{
				{Name: "Gin"},
			},
		},
		Structure: types.ProjectStructure{
			Directories: []types.Directory{
				{Path: "cmd", Purpose: "Commands"},
				{Path: "internal", Purpose: "Private packages"},
			},
		},
		Conventions: []types.Convention{
			{Category: "testing", Description: "Use table-driven tests"},
		},
		KeyFiles: []types.KeyFile{
			{Path: "main.go", Purpose: "Entry point"},
		},
	}

	files, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should generate config.yaml
	if _, exists := files[".continue/config.yaml"]; !exists {
		t.Error("expected .continue/config.yaml to be generated")
	}

	// Should generate rules file
	if _, exists := files[".continue/rules/project.md"]; !exists {
		t.Error("expected .continue/rules/project.md to be generated")
	}

	// Should generate prompts file
	if _, exists := files[".continue/prompts/review.md"]; !exists {
		t.Error("expected .continue/prompts/review.md to be generated")
	}

	// Check rules file content
	rulesContent := string(files[".continue/rules/project.md"])
	if !strings.Contains(rulesContent, "multi-test") {
		t.Error("expected project name in rules file")
	}
	if !strings.Contains(rulesContent, "Go") {
		t.Error("expected language in rules file")
	}

	// Check prompts file content
	promptsContent := string(files[".continue/prompts/review.md"])
	if !strings.Contains(promptsContent, "Code Review") {
		t.Error("expected code review title in prompts file")
	}
	if !strings.Contains(promptsContent, "Go-specific") {
		t.Error("expected Go-specific checks in prompts file")
	}
}

func TestSanitizeYAMLString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with spaces", "with spaces"},
		{"with:colon", "\"with:colon\""},
		{"with{brace", "\"with{brace\""},
		{"with\"quote", "\"with\\\"quote\""},
		{"normal-text", "\"normal-text\""}, // hyphen triggers quoting due to YAML special chars
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := sanitizeYAMLString(tc.input)
			if result != tc.expected {
				t.Errorf("sanitizeYAMLString(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}
