package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/internal/merger"
)

// scanPair writes AGENTS.md and the CLAUDE.md shim into root, as a default
// scan does.
func scanPair(t *testing.T, root string) {
	t.Helper()
	generate(t, root, "agents")
	generate(t, root, "claude")
}

func TestHandWrittenAgentsFile(t *testing.T) {
	const notes = "# Team agents guide\nKEEP\n"
	skip := "⚠️  AGENTS.md exists without argus markers, left unchanged (add an empty ARGUS:AUTO pair where the generated block should go, or use --force to overwrite)\n"
	tests := []struct {
		name, existing string
		force          bool
		keep           bool // existing text survives
		notice         string
	}{
		{"skipped", notes, false, true, skip},
		{"empty marker pair opts in", notes + merger.AutoStartMarker + "\n" + merger.AutoEndMarker + "\n", false, true, ""},
		{"force replaces it", notes, true, false, "⚠️  AGENTS.md had no argus markers, replaced with a fresh generated file (--force); the old text was not kept\n"},
		{"BOM only is created", "\ufeff\n", false, false, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags(t)
			_, errOut := captureOutput(t)
			force = tt.force
			root := t.TempDir()
			writeFiles(t, root, map[string]string{"AGENTS.md": tt.existing})

			generate(t, root, "agents")
			got := readFile(t, root, "AGENTS.md")
			if strings.Contains(got, "KEEP") != tt.keep {
				t.Errorf("text kept = %v, want %v:\n%s", !tt.keep, tt.keep, got)
			}
			if wrote := got != tt.existing; wrote != (tt.notice != skip) {
				t.Errorf("file written = %v:\n%s", wrote, got)
			}
			if errOut.String() != tt.notice {
				t.Errorf("stderr = %q, want %q", errOut, tt.notice)
			}
		})
	}
}

func TestHandMadeShimIsUpToDate(t *testing.T) {
	resetFlags(t)
	out, errOut := captureOutput(t)
	root := t.TempDir()
	const own = "@AGENTS.md\n\n# Claude notes\n"
	writeFiles(t, root, map[string]string{"CLAUDE.md": own})

	generate(t, root, "claude")
	if got := readFile(t, root, "CLAUDE.md"); got != own || out.String() != "✅ CLAUDE.md is up to date\n" || errOut.Len() != 0 {
		t.Errorf("hand-made shim: file %q, printed %q / %q", got, out, errOut)
	}
}

func TestSymlinkedContextFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks needs extra privileges on Windows")
	}
	const docs = "# Docs\nKEEP\n"
	tests := []struct {
		name         string
		link, target string // link points at target, both relative to root
		wantAgents   bool   // AGENTS.md is written as a regular file
		wantShim     bool   // CLAUDE.md is written as the shim
		notice       string
	}{
		{"CLAUDE.md links to AGENTS.md", "CLAUDE.md", "AGENTS.md", true, false, ""},
		{"AGENTS.md links to CLAUDE.md", "AGENTS.md", "CLAUDE.md", false, false,
			"⚠️  AGENTS.md is a symlink to CLAUDE.md, both left unchanged (remove the symlink: argus then writes AGENTS.md and makes CLAUDE.md import it)\n"},
		{"AGENTS.md links elsewhere", "AGENTS.md", "docs/guide.md", false, true,
			"⚠️  AGENTS.md is a symlink, left unchanged (replace it with a regular file to let argus write it)\n"},
		{"CLAUDE.md links elsewhere", "CLAUDE.md", "docs/guide.md", true, false,
			"⚠️  CLAUDE.md is a symlink, left unchanged (replace it with a regular file to let argus write it)\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags(t)
			_, errOut := captureOutput(t)
			force = true // never writes through or replaces a link
			root := t.TempDir()
			writeFiles(t, root, map[string]string{"docs/guide.md": docs})
			if tt.target == "CLAUDE.md" {
				writeFiles(t, root, map[string]string{"CLAUDE.md": docs})
			}
			if err := os.Symlink(tt.target, filepath.Join(root, tt.link)); err != nil {
				t.Fatal(err)
			}

			scanPair(t, root)
			if fi, err := os.Lstat(filepath.Join(root, tt.link)); err != nil || fi.Mode()&os.ModeSymlink == 0 {
				t.Errorf("%s is no longer a symlink (err %v)", tt.link, err)
			}
			if got := readFile(t, root, "docs/guide.md"); got != docs {
				t.Errorf("written through a link:\n%s", got)
			}
			if tt.target == "CLAUDE.md" && readFile(t, root, "CLAUDE.md") != docs {
				t.Error("CLAUDE.md behind the AGENTS.md link was written")
			}
			if got := readFile(t, root, "AGENTS.md"); strings.Contains(got, merger.AutoStartMarker+"\n# demo") != tt.wantAgents {
				t.Errorf("AGENTS.md written = %v:\n%s", !tt.wantAgents, got)
			}
			if got := readFile(t, root, "CLAUDE.md"); (got == shim) != tt.wantShim {
				t.Errorf("CLAUDE.md shim written = %v:\n%s", !tt.wantShim, got)
			}
			if errOut.String() != tt.notice {
				t.Errorf("stderr = %q, want %q", errOut, tt.notice)
			}
		})
	}
}

func TestDanglingShimLinkIsUpToDate(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks needs extra privileges on Windows")
	}
	resetFlags(t)
	out, errOut := captureOutput(t)
	root := t.TempDir()
	if err := os.Symlink("AGENTS.md", filepath.Join(root, "CLAUDE.md")); err != nil {
		t.Fatal(err)
	}
	generate(t, root, "claude")
	if out.String() != "✅ CLAUDE.md is up to date\n" || errOut.Len() != 0 {
		t.Errorf("printed %q / %q", out, errOut)
	}
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); err == nil {
		t.Error("a write went through the dangling link")
	}
}
