package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	resetFlags(t)
	out, errOut := captureOutput(t)
	restoreFlags(t, checkCmd, "format", "verbose")
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"go.mod": "module example.com/fx\n\ngo 1.22\n", "main.go": "package main\n\nfunc main() {}\n"})

	check := func(want string, wantCode int, args ...string) {
		t.Helper()
		out.Reset()
		errOut.Reset()
		before := snapshotTree(t, root)
		cliOut, code := runCLI(t, append([]string{"check"}, append(args, root)...)...)
		if out.String() != want || code != wantCode {
			t.Errorf("check %v: exit %d, stdout %q, want exit %d, %q (stderr %q %q)", args, code, out, wantCode, want, errOut, cliOut)
		}
		if code == 1 && strings.Contains(cliOut, "Error:") {
			t.Errorf("drift printed an error line: %q", cliOut)
		}
		if diff := treeDiff(before, snapshotTree(t, root)); len(diff) > 0 {
			t.Errorf("check changed %v", diff)
		}
	}

	check("AGENTS.md: missing\nCLAUDE.md: missing\n", 1)
	writeFormats(t, root, analyze(t, root, true), false, "agents", "claude")
	check("AGENTS.md: up to date\nCLAUDE.md: up to date\n", 0)
	if errOut.String() != "argus check: all 2 files up to date\n" {
		t.Errorf("summary %q", errOut)
	}

	agents := readFile(t, root, "AGENTS.md")
	writeFiles(t, root, map[string]string{"AGENTS.md": agents + "\nTeam notes outside the block\n"})
	check("AGENTS.md: up to date\nCLAUDE.md: up to date\n", 0)

	writeFiles(t, root, map[string]string{"AGENTS.md": strings.Replace(agents, "## Stack", "## Stack edited", 1)})
	if err := os.Remove(filepath.Join(root, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	check("AGENTS.md: out of date\nCLAUDE.md: missing\n", 1)
	if errOut.String() != "argus check: 2 of 2 files out of date (run argus scan to update)\n" {
		t.Errorf("summary %q", errOut)
	}

	writeFiles(t, root, map[string]string{"AGENTS.md": "# Hand-written\n"})
	check("AGENTS.md: skipped\nCLAUDE.md: missing\n", 1, "--format", "claude")
}

func TestCheckErrorsExitTwo(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		args  []string
		want  string
	}{
		{"scan-only flag", nil, []string{"--ai"}, "Error: unknown flag: --ai"},
		{"ai enabled", map[string]string{".argus.yaml": "ai:\n  enabled: true\n"}, nil, "Error: ai is enabled in .argus.yaml: check cannot reproduce AI output\n"},
		{"shallow clone", map[string]string{".git/shallow": ""}, nil, "Error: check needs full git history, this clone is shallow (use fetch-depth: 0)\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags(t)
			captureOutput(t)
			root := t.TempDir()
			writeFiles(t, root, tt.files)
			got, code := runCLI(t, append(append([]string{"check"}, tt.args...), root)...)
			if code != 2 || !strings.Contains(got, tt.want) {
				t.Errorf("exit %d, output %q, want 2 and %q", code, got, tt.want)
			}
		})
	}
}

// check analyzes with .argus.yaml like scan, so ignored files cause no drift.
func TestCheckUsesConfig(t *testing.T) {
	resetFlags(t)
	captureOutput(t)
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		".argus.yaml":  "ignore:\n  - extra\n",
		"go.mod":       "module example.com/fx\n\ngo 1.22\n",
		"main.go":      "package main\n\nfunc main() {}\n",
		"extra/a.py":   "print('a')\n",
		"extra/b.py":   "print('b')\n",
		"extra/c/d.py": "print('d')\n",
	})
	if _, code := runCLI(t, "scan", root); code != 0 {
		t.Fatalf("scan exit %d", code)
	}
	if got, code := runCLI(t, "check", root); code != 0 {
		t.Errorf("check after scan: exit %d\n%s", code, got)
	}
}

// A workspace output check cannot read is an error, not a pass.
func TestCheckFailsOnUnreadableWorkspace(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs unix permission bits that apply to the user")
	}
	resetFlags(t)
	captureOutput(t)
	root := monorepoFixture(t)
	writeFiles(t, root, map[string]string{".argus.yaml": "monorepo:\n  per_workspace: true\n  root_overview: true\n"})
	if _, code := runCLI(t, "scan", root); code != 0 {
		t.Fatalf("scan exit %d", code)
	}
	locked := filepath.Join(root, "packages", "web", "AGENTS.md")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0644) })
	if got, code := runCLI(t, "check", root); code != 2 || !strings.Contains(got, "packages/web/AGENTS.md") {
		t.Errorf("exit %d, want 2:\n%s", code, got)
	}
}
