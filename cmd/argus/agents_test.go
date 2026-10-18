package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/internal/config"
	"github.com/spf13/cobra"
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

func TestFullRestoresDetailedBody(t *testing.T) {
	resetFlags(t)
	captureOutput(t)
	detailed := []string{"## Quick Reference", "## Tech Stack"}
	for _, full := range []bool{false, true} {
		root := t.TempDir()
		if err := generateOutput(root, "", "agents", goAnalysis(), config.DefaultConfig(), false, full); err != nil {
			t.Fatal(err)
		}
		got := readFile(t, root, "AGENTS.md")
		for _, h := range detailed {
			if strings.Contains(got, h) != full {
				t.Errorf("full=%v: %q present = %v", full, h, !full)
			}
		}
		if strings.Contains(got, "@.claude/rules") || strings.Contains(got, "## Additional Rules") {
			t.Errorf("full=%v: rule imports in AGENTS.md:\n%s", full, got)
		}
	}
}

func TestCompactIsDeprecated(t *testing.T) {
	const msg = "the lean output is now the default, use --full for the detailed output"
	for _, c := range []*cobra.Command{scanCmd, syncCmd} {
		if f := c.Flags().Lookup("compact"); f == nil || f.Deprecated != msg || !f.Hidden {
			t.Errorf("%s --compact is not a hidden deprecated flag", c.Name())
		}
	}
	restoreFlags(t, scanCmd, "compact", "full")
	errOut, code := runCLI(t, "scan", "--compact", "--full", t.TempDir())
	if code == 0 || !strings.Contains(errOut, "Flag --compact has been deprecated, "+msg) {
		t.Errorf("scan --compact --full: exit %d, output:\n%s", code, errOut)
	}
}
