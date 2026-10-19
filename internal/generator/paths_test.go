package generator

import "testing"

func TestIsGeneratedPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"CLAUDE.md", true},
		{"AGENTS.md", true},
		{".cursorrules", true},
		{".cursor/rules/argus.mdc", true},
		{".github/copilot-instructions.md", true},
		{"./CLAUDE.md", true},
		{".claude", true},
		{".claude/settings.json", true},
		{".claude/agents/planner.md", true},
		{".continue", true},
		{".continue/config.yaml", true},
		{"packages/web/CLAUDE.md", true},
		{"packages/web/AGENTS.md", true},
		{"packages/web/.github/copilot-instructions.md", true},
		{"packages/web/.claude/rules/testing.md", true},
		{"apps/api/.continue/config.yaml", true},
		{"packages/web/.cursor/rules/argus.mdc", true},
		{"packages/web/.cursorrules", true},
		{`packages\web\CLAUDE.md`, true},
		{`.claude\agents\planner.md`, true},
		{`.github\copilot-instructions.md`, true},
		{".github", false},
		{".github/workflows/ci.yml", false},
		{".cursor", false},
		{".cursor/rules", false},
		{".cursor/rules/team.mdc", false},
		{"docs/claude.md", false},
		{"claude.md", false},
		{"CLAUDE.md.bak", false},
		{"MY_CLAUDE.md", false},
		{"docs/notes-AGENTS.md", false},
		{".claudex", false},
		{".claudex/settings.json", false},
		{"src/.continuerc", false},
		{"copilot-instructions.md", false},
		{"cmd/argus/main.go", false},
		{".argus-123456", true},
		{"packages/web/.argus-42", true},
		{`docs\.argus-7`, true},
		{".argus-", false},
		{".argus-12.md", false},
		{".argus-tmp", false},
		{".argus.yaml", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsGeneratedPath(tt.path); got != tt.want {
			t.Errorf("IsGeneratedPath(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestIsMergeFile(t *testing.T) {
	tests := []struct {
		rel  string
		want bool
	}{
		{"CLAUDE.md", true},
		{"AGENTS.md", true},
		{".cursor/rules/argus.mdc", true},
		{".github/copilot-instructions.md", true},
		{".cursorrules", false},
		{".continue/config.yaml", false},
		{".claude/settings.json", false},
		{".claude/agents/planner.md", false},
		{"packages/web/CLAUDE.md", false},
		{"./CLAUDE.md", false},
		{"claude.md", false},
	}
	for _, tt := range tests {
		if got := IsMergeFile(tt.rel); got != tt.want {
			t.Errorf("IsMergeFile(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}
