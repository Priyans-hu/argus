package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/internal/analyzer"
	"github.com/Priyans-hu/argus/internal/config"
	"gopkg.in/yaml.v3"
)

// skipNotice is the stderr line for n skipped files.
func skipNotice(n int) string {
	return fmt.Sprintf("⚠️  Skipped %d existing files that differ from the generated output (use --force to overwrite, -v to list)\n", n)
}

func TestWriteCreateOnly(t *testing.T) {
	const content = "name: demo\nrules:\n  - one\n"
	crlf := strings.ReplaceAll(content, "\n", "\r\n")
	tests := []struct {
		name        string
		existing    string // "" means the file is missing
		force, dry  bool
		wantStatus  fileStatus
		wantFile    string // "" means the file must not exist
		wantPreview bool
	}{
		{name: "missing", wantStatus: fileWritten, wantFile: content},
		{name: "identical", existing: content, wantStatus: fileUnchanged, wantFile: content},
		{name: "identical with CRLF", existing: crlf, wantStatus: fileUnchanged, wantFile: crlf},
		{name: "differs", existing: "name: mine\n", wantStatus: fileSkipped, wantFile: "name: mine\n"},
		{name: "differs with force", existing: "name: mine\n", force: true, wantStatus: fileWritten, wantFile: content},
		{name: "dry-run missing", dry: true, wantStatus: fileWritten, wantPreview: true},
		{name: "dry-run differs", existing: "name: mine\n", dry: true, wantStatus: fileSkipped, wantFile: "name: mine\n"},
		{name: "dry-run differs with force", existing: "name: mine\n", dry: true, force: true, wantStatus: fileWritten, wantFile: "name: mine\n", wantPreview: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags(t)
			out, _ := captureOutput(t)
			force = tt.force
			root := t.TempDir()
			if tt.existing != "" {
				writeFiles(t, root, map[string]string{".continue/config.yaml": tt.existing})
			}

			status, err := writeCreateOnly(root, "", ".continue/config.yaml", []byte(content), tt.dry)
			if err != nil {
				t.Fatal(err)
			}
			if status != tt.wantStatus {
				t.Errorf("status = %v, want %v", status, tt.wantStatus)
			}

			data, err := os.ReadFile(filepath.Join(root, ".continue", "config.yaml"))
			switch {
			case tt.wantFile == "" && !errors.Is(err, fs.ErrNotExist):
				t.Errorf("file exists (err %v), want none", err)
			case tt.wantFile != "" && string(data) != tt.wantFile:
				t.Errorf("file = %q, want %q", data, tt.wantFile)
			}

			wantSkipped := tt.wantStatus == fileSkipped
			if gotSkipped := len(skipped) == 1 && skipped[0] == ".continue/config.yaml"; gotSkipped != wantSkipped {
				t.Errorf("skipped = %v, want recorded: %v", skipped, wantSkipped)
			}
			if got := strings.Contains(out.String(), "Would write to"); got != tt.wantPreview {
				t.Errorf("preview printed = %v, want %v", got, tt.wantPreview)
			}
		})
	}
}

func TestClaudeCodeOutputReturnsWriteErrors(t *testing.T) {
	resetFlags(t)
	out, _ := captureOutput(t)
	root := t.TempDir()
	writeFiles(t, root, map[string]string{".claude": "a file, not a directory\n"})

	err := generateOutput(root, "", "claude-code", goAnalysis(), config.DefaultConfig(), false, false)
	if err == nil || !strings.Contains(err.Error(), ".claude") {
		t.Fatalf("err = %v, want an error naming .claude", err)
	}
	if strings.Contains(out.String(), "Generated") {
		t.Errorf("success line printed after a failed write:\n%s", out)
	}
}

func TestContinueOutputStaysValidYAML(t *testing.T) {
	resetFlags(t)
	out, _ := captureOutput(t)
	root := t.TempDir()

	var first string
	for run := 1; run <= 2; run++ {
		if err := generateOutput(root, "", "continue", goAnalysis(), config.DefaultConfig(), false, false); err != nil {
			t.Fatal(err)
		}
		got := readFile(t, root, ".continue/config.yaml")
		var doc map[string]any
		if err := yaml.Unmarshal([]byte(got), &doc); err != nil {
			t.Fatalf("run %d: config.yaml is not a YAML mapping: %v", run, err)
		}
		if strings.Contains(got, "ARGUS:") {
			t.Errorf("run %d: config.yaml has argus markers", run)
		}
		if run == 1 {
			first = got
		} else if got != first {
			t.Errorf("run %d changed config.yaml", run)
		}
	}
	if !strings.Contains(out.String(), "✅ Generated .continue/config.yaml\n") ||
		!strings.Contains(out.String(), "✅ .continue/config.yaml is up to date\n") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestClaudeCodeSkipsHandEditedFiles(t *testing.T) {
	resetFlags(t)
	out, errOut := captureOutput(t)
	root := t.TempDir()
	run := func() {
		t.Helper()
		if err := generateOutput(root, "", "claude-code", goAnalysis(), config.DefaultConfig(), false, false); err != nil {
			t.Fatal(err)
		}
		reportSkipped()
	}

	run()
	if !strings.Contains(out.String(), "Claude Code configs in .claude/") {
		t.Fatalf("first run printed:\n%s", out)
	}
	edited := []string{".claude/agents/go-reviewer.md", ".claude/settings.json"}
	original := map[string]string{}
	for _, rel := range edited {
		original[rel] = readFile(t, root, rel)
		writeFiles(t, root, map[string]string{rel: "KEEP " + rel + "\n"})
	}

	verbose = true
	errOut.Reset()
	run()
	for _, rel := range edited {
		if got := readFile(t, root, rel); got != "KEEP "+rel+"\n" {
			t.Errorf("%s was overwritten: %q", rel, got)
		}
	}
	want := skipNotice(2) + "  .claude/agents/go-reviewer.md\n  .claude/settings.json\n"
	if errOut.String() != want {
		t.Errorf("stderr =\n%q\nwant\n%q", errOut, want)
	}

	verbose, force = false, true
	errOut.Reset()
	run()
	for _, rel := range edited {
		if got := readFile(t, root, rel); got != original[rel] {
			t.Errorf("--force did not restore %s", rel)
		}
	}
	if errOut.Len() != 0 {
		t.Errorf("--force printed notices: %q", errOut)
	}

	force = false
	out.Reset()
	run()
	if out.String() != "✅ Claude Code configs are up to date\n" {
		t.Errorf("final run printed %q", out)
	}
}

func TestSkipLineIsAggregatedAcrossFormats(t *testing.T) {
	resetFlags(t)
	_, errOut := captureOutput(t)
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".continue/config.yaml": "name: mine\n",
		".claude/settings.json": "{}\n",
	})

	for _, format := range []string{"continue", "claude-code"} {
		if err := generateOutput(root, "", format, goAnalysis(), config.DefaultConfig(), false, false); err != nil {
			t.Fatal(err)
		}
	}
	reportSkipped()

	want := skipNotice(2)
	if errOut.String() != want {
		t.Errorf("stderr = %q, want %q", errOut, want)
	}
}

func TestWatchReportsEachSkipOnce(t *testing.T) {
	resetFlags(t)
	_, errOut := captureOutput(t)
	watchNotified = map[string]bool{}
	root := t.TempDir()
	writeFiles(t, root, map[string]string{".continue/config.yaml": "name: mine\n"})

	for range 3 {
		if err := generateOutput(root, "", "continue", goAnalysis(), config.DefaultConfig(), false, false); err != nil {
			t.Fatal(err)
		}
		reportSkipped()
	}

	want := skipNotice(1)
	if errOut.String() != want {
		t.Errorf("stderr = %q, want one skip line", errOut)
	}
}

// monorepoFixture is an npm workspaces repo with packages/web and packages/api.
func monorepoFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"package.json":              `{"name":"mono","private":true,"workspaces":["packages/*"]}`,
		"packages/web/package.json": `{"name":"web","scripts":{"build":"next build"},"dependencies":{"next":"14.0.0"}}`,
		"packages/api/package.json": `{"name":"api","scripts":{"start":"node index.js"},"dependencies":{"express":"4.18.0"}}`,
	})
	return root
}

// scanMonorepo analyzes root and runs the monorepo scan with its config.
func scanMonorepo(t *testing.T, root string, formats ...string) {
	t.Helper()
	cfg, err := config.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	analysis, err := analyzer.NewParallelAnalyzer(root, nil).Analyze(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if err := runMonorepoScan(context.Background(), root, cfg, formats, analysis); err != nil {
		t.Fatal(err)
	}
}

func TestMonorepoAppliesRootClaudeCodeConfig(t *testing.T) {
	resetFlags(t)
	out, _ := captureOutput(t)
	root := monorepoFixture(t)
	writeFiles(t, root, map[string]string{
		".argus.yaml": "claude_code:\n  agents: false\n  skills: false\n  rules: true\n  mcp: false\n  hooks: false\n",
	})

	scanMonorepo(t, root, "claude-code")

	for _, ws := range []string{"packages/web", "packages/api"} {
		if _, err := os.Stat(filepath.Join(root, ws, ".claude", "agents")); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s has agents despite agents: false in the root config", ws)
		}
		if _, err := os.Stat(filepath.Join(root, ws, ".claude", "rules")); err != nil {
			t.Errorf("%s has no rules: %v", ws, err)
		}
	}
	if !strings.Contains(out.String(), "Generated context files for 2/2 workspaces") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestMonorepoDoesNotCountFailedWorkspaces(t *testing.T) {
	resetFlags(t)
	out, _ := captureOutput(t)
	root := monorepoFixture(t)
	writeFiles(t, root, map[string]string{"packages/api/.claude": "a file, not a directory\n"})

	scanMonorepo(t, root, "claude-code")

	if !strings.Contains(out.String(), "Generated context files for 1/2 workspaces") {
		t.Errorf("unexpected output:\n%s", out)
	}
}
