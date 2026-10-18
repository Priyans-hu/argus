package main

import (
	"slices"
	"testing"

	"github.com/Priyans-hu/argus/internal/config"
)

// shim is the exact CLAUDE.md a first scan writes.
const shim = "<!-- ARGUS:AUTO -->\n@AGENTS.md\n<!-- /ARGUS:AUTO -->\n\n<!-- ARGUS:CUSTOM -->\n<!-- Claude Code only: notes here are not shared with other tools. argus keeps this block when it regenerates. -->\n<!-- /ARGUS:CUSTOM -->\n"

func TestExpandFormats(t *testing.T) {
	tests := []struct {
		in, want []string
	}{
		{[]string{"claude"}, []string{"agents", "claude"}},
		{[]string{"agents"}, []string{"agents"}},
		{[]string{"all"}, []string{"agents", "claude", "claude-code", "cursor", "copilot"}},
		{[]string{"all", "continue"}, []string{"agents", "claude", "claude-code", "cursor", "copilot", "continue"}},
		{[]string{"copilot", "claude", "claude", "agents"}, []string{"agents", "claude", "copilot"}},
		{[]string{"continue"}, []string{"continue"}},
		{[]string{"zed", "cursor", "bogus"}, []string{"cursor", "bogus", "zed"}},
	}
	for _, tt := range tests {
		if got := expandFormats(tt.in); !slices.Equal(got, tt.want) {
			t.Errorf("expandFormats(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestContinueIsDeprecated(t *testing.T) {
	const warning = "⚠️  format \"continue\" is deprecated and will be removed in a future release; it is no longer part of --format all\n"
	override := &config.Config{Monorepo: &config.MonorepoConfig{WorkspaceOverrides: map[string]*config.WorkspaceOverride{
		"apps/a": {Output: []string{"continue"}}, "apps/b": {Output: []string{"continue"}},
	}}}
	tests := []struct {
		name    string
		cfg     *config.Config
		formats []string
		want    string
	}{
		{"named", config.DefaultConfig(), []string{"claude", "continue"}, warning},
		{"all", config.DefaultConfig(), []string{"all"}, ""},
		{"not named", config.DefaultConfig(), []string{"claude"}, ""},
		{"workspace overrides", override, []string{"agents"}, warning},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags(t)
			_, errOut := captureOutput(t)
			warnContinue(tt.cfg, expandFormats(tt.formats))
			if errOut.String() != tt.want {
				t.Errorf("printed %q, want %q", errOut, tt.want)
			}
		})
	}
}

func TestClaudeShim(t *testing.T) {
	resetFlags(t)
	out, errOut := captureOutput(t)
	root := t.TempDir()

	generate(t, root, "claude")
	if got := readFile(t, root, "CLAUDE.md"); got != shim {
		t.Fatalf("CLAUDE.md = %q, want %q", got, shim)
	}
	out.Reset()
	generate(t, root, "claude")
	if got := readFile(t, root, "CLAUDE.md"); got != shim || out.String() != "✅ CLAUDE.md is up to date\n" || errOut.Len() != 0 {
		t.Errorf("second run changed the shim or printed %q %q", out, errOut)
	}
}
