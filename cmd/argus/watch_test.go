package main

import (
	"path/filepath"
	"testing"
)

func TestIsGeneratedFile(t *testing.T) {
	root := filepath.Join(t.TempDir(), "proj")
	underClaude := filepath.Join(t.TempDir(), ".claude", "proj")
	tests := []struct {
		root string
		rel  string
		want bool
	}{
		{root, "CLAUDE.md", true},
		{root, "AGENTS.md", true},
		{root, ".cursorrules", true},
		{root, ".github/copilot-instructions.md", true},
		{root, ".claude/settings.json", true},
		{root, ".continue/config.yaml", true},
		{root, "packages/web/CLAUDE.md", true},
		{root, "src/main.go", false},
		{root, ".github/workflows/ci.yml", false},
		{root, "docs/claude.md", false},
		{underClaude, "src/main.go", false},
		{underClaude, "CLAUDE.md", true},
		{underClaude, ".continue/config.yaml", true},
	}
	for _, tt := range tests {
		path := filepath.Join(tt.root, filepath.FromSlash(tt.rel))
		if got := isGeneratedFile(tt.root, path); got != tt.want {
			t.Errorf("isGeneratedFile(%q, %q) = %v, want %v", tt.root, path, got, tt.want)
		}
	}
}
