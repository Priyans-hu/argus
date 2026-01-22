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

func TestClaudeGenerator_SetCompact(t *testing.T) {
	g := NewClaudeGenerator()

	// Default should be false
	if g.compact {
		t.Error("expected compact to be false by default")
	}

	// Set to true
	g.SetCompact(true)
	if !g.compact {
		t.Error("expected compact to be true after SetCompact(true)")
	}

	// Set back to false
	g.SetCompact(false)
	if g.compact {
		t.Error("expected compact to be false after SetCompact(false)")
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

func TestClaudeGenerator_Generate_CompactMode(t *testing.T) {
	normalGen := NewClaudeGenerator()
	compactGen := NewClaudeGenerator()
	compactGen.SetCompact(true)

	analysis := &types.Analysis{
		ProjectName: "test-project",
		RootPath:    "/test/path",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "Go", Version: "1.21", Percentage: 100},
			},
		},
		Commands: []types.Command{
			{Command: "go build", Description: "Build the project"},
			{Command: "go test", Description: "Run tests"},
		},
		KeyFiles: []types.KeyFile{
			{Path: "main.go", Purpose: "Entry point"},
			{Path: "config.go", Purpose: "Configuration"},
			{Path: "handler.go", Purpose: "HTTP handlers"},
			{Path: "service.go", Purpose: "Business logic"},
			{Path: "model.go", Purpose: "Data models"},
			{Path: "utils.go", Purpose: "Utilities"},
			{Path: "extra.go", Purpose: "Extra file"},
		},
		ConfigFiles: []types.ConfigFileInfo{
			{Path: "go.mod", Type: "Go Modules"},
		},
		CodePatterns: &types.CodePatterns{
			Testing: []types.PatternInfo{
				{Name: "t.Fatal", FileCount: 5},
				{Name: "t.Error", FileCount: 4},
				{Name: "t.Run", FileCount: 3},
			},
			DataFetching: []types.PatternInfo{
				{Name: "http.Get", FileCount: 2},
			},
		},
	}

	normalContent, err := normalGen.Generate(analysis)
	if err != nil {
		t.Fatalf("Normal Generate failed: %v", err)
	}

	compactContent, err := compactGen.Generate(analysis)
	if err != nil {
		t.Fatalf("Compact Generate failed: %v", err)
	}

	normalStr := string(normalContent)
	compactStr := string(compactContent)

	// Compact should be smaller
	if len(compactContent) >= len(normalContent) {
		t.Errorf("compact output (%d) should be smaller than normal output (%d)",
			len(compactContent), len(normalContent))
	}

	// Compact should skip certain sections
	if strings.Contains(compactStr, "## Configuration") {
		t.Error("compact mode should skip Configuration section")
	}

	if strings.Contains(compactStr, "## Available Commands") {
		t.Error("compact mode should skip Available Commands section")
	}

	if strings.Contains(compactStr, "## Dependencies") {
		t.Error("compact mode should skip Dependencies section")
	}

	// Normal should have these sections
	if !strings.Contains(normalStr, "## Available Commands") {
		t.Error("normal mode should have Available Commands section")
	}
}

func TestClaudeGenerator_CompactKeyFiles(t *testing.T) {
	g := NewClaudeGenerator()
	g.SetCompact(true)

	analysis := &types.Analysis{
		ProjectName: "test-project",
		KeyFiles: []types.KeyFile{
			{Path: "main.go", Purpose: "Entry point"},
			{Path: "config.go", Purpose: "Configuration"},
			{Path: "handler.go", Purpose: "HTTP handlers"},
			{Path: "service.go", Purpose: "Business logic"},
			{Path: "model.go", Purpose: "Data models"},
			{Path: "utils.go", Purpose: "Utilities"},
			{Path: "extra1.go", Purpose: "Extra file 1"},
			{Path: "extra2.go", Purpose: "Extra file 2"},
			{Path: "extra3.go", Purpose: "Extra file 3"},
			{Path: "extra4.go", Purpose: "Extra file 4"},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Should show max 5 key files in compact mode
	keyFilesSection := extractSection(contentStr, "## Key Files")
	fileCount := strings.Count(keyFilesSection, "- `")

	if fileCount > 5 {
		t.Errorf("compact mode should show max 5 key files, got %d", fileCount)
	}

	// Should indicate more files exist
	if !strings.Contains(keyFilesSection, "more files") {
		t.Error("should indicate more files exist when truncated")
	}
}

func TestClaudeGenerator_CompactPatterns(t *testing.T) {
	g := NewClaudeGenerator()
	g.SetCompact(true)

	analysis := &types.Analysis{
		ProjectName: "test-project",
		CodePatterns: &types.CodePatterns{
			Testing: []types.PatternInfo{
				{Name: "t.Fatal", FileCount: 10},
				{Name: "t.Error", FileCount: 8},
				{Name: "t.Run", FileCount: 6},
				{Name: "assert.Equal", FileCount: 4},
				{Name: "require.NoError", FileCount: 2},
			},
			DataFetching: []types.PatternInfo{
				{Name: "http.Get", FileCount: 5},
				{Name: "http.Post", FileCount: 3},
			},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Should have Detected Patterns section
	if !strings.Contains(contentStr, "## Detected Patterns") {
		t.Error("should have Detected Patterns section")
	}

	// Should indicate "Top patterns"
	if !strings.Contains(contentStr, "Top patterns") {
		t.Error("should indicate these are top patterns")
	}

	// Should limit patterns shown - compact mode shows max 3 per category
	patternsSection := extractSection(contentStr, "## Detected Patterns")

	// Count testing patterns by looking for specific markers
	testingPatternCount := 0
	for _, p := range []string{"t.Fatal", "t.Error", "t.Run", "assert.Equal", "require.NoError"} {
		if strings.Contains(patternsSection, p) {
			testingPatternCount++
		}
	}

	// Should show max 3 testing patterns
	if testingPatternCount > 3 {
		t.Errorf("should show max 3 testing patterns in compact, got %d", testingPatternCount)
	}
}

func TestClaudeGenerator_CompactEndpoints(t *testing.T) {
	g := NewClaudeGenerator()
	g.SetCompact(true)

	analysis := &types.Analysis{
		ProjectName: "test-project",
		Endpoints: []types.Endpoint{
			{Method: "GET", Path: "/api/users"},
			{Method: "GET", Path: "/api/posts"},
			{Method: "GET", Path: "/api/comments"},
			{Method: "GET", Path: "/api/tags"},
			{Method: "GET", Path: "/api/categories"},
			{Method: "POST", Path: "/api/users"},
			{Method: "POST", Path: "/api/posts"},
			{Method: "PUT", Path: "/api/users/:id"},
			{Method: "DELETE", Path: "/api/users/:id"},
		},
	}

	content, err := g.Generate(analysis)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	contentStr := string(content)

	// Should have API Endpoints section
	if !strings.Contains(contentStr, "## API Endpoints") {
		t.Error("should have API Endpoints section")
	}

	// Should show method summaries
	if !strings.Contains(contentStr, "**GET:**") {
		t.Error("should show GET method summary")
	}

	// Should indicate more endpoints exist
	if !strings.Contains(contentStr, "+") || !strings.Contains(contentStr, "more") {
		t.Error("should indicate more endpoints exist when truncated")
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
		"## Dependencies",
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
