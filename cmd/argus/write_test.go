package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/internal/analyzer"
	"github.com/Priyans-hu/argus/internal/config"
	"github.com/Priyans-hu/argus/internal/merger"
	"github.com/Priyans-hu/argus/pkg/types"
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
	out, errOut := captureOutput(t)
	root := monorepoFixture(t)
	writeFiles(t, root, map[string]string{"packages/api/.claude": "a file, not a directory\n"})

	scanMonorepo(t, root, "claude-code")

	if !strings.Contains(out.String(), "Generated context files for 1/2 workspaces") {
		t.Errorf("unexpected output:\n%s", out)
	}
	const warn = "[warn] packages/api (claude-code)"
	if !strings.Contains(errOut.String(), warn) || strings.Contains(out.String(), warn) {
		t.Errorf("want the warning on stderr only, got stdout:\n%s\nstderr:\n%s", out, errOut)
	}
}

func TestMonorepoMergesRootOverview(t *testing.T) {
	resetFlags(t)
	captureOutput(t)
	root := monorepoFixture(t)

	scanMonorepo(t, root, "claude")
	overview := readFile(t, root, "CLAUDE.md")
	checkMarked(t, overview)

	custom := merger.CustomStartMarker + "\nKEEP-CUSTOM\n" + merger.CustomEndMarker + "\n"
	writeFiles(t, root, map[string]string{
		"CLAUDE.md":              "# Root notes\nKEEP-ROOT-NOTE\n\n" + overview + "\n" + custom,
		"packages/web/CLAUDE.md": readFile(t, root, "packages/web/CLAUDE.md") + "\n" + custom,
	})
	scanMonorepo(t, root, "claude")
	second := snapshotTree(t, root)
	scanMonorepo(t, root, "claude")
	if diff := treeDiff(second, snapshotTree(t, root)); len(diff) > 0 {
		t.Errorf("third run changed %v", diff)
	}

	for rel, keep := range map[string][]string{
		"CLAUDE.md":              {"KEEP-ROOT-NOTE", "KEEP-CUSTOM"},
		"packages/web/CLAUDE.md": {"KEEP-CUSTOM"},
		"packages/api/CLAUDE.md": nil,
	} {
		checkMarked(t, second[rel])
		for _, k := range keep {
			if !strings.Contains(second[rel], k) {
				t.Errorf("%s lost %s", rel, k)
			}
		}
	}
}

func TestMonorepoAdoptsHandWrittenRootFile(t *testing.T) {
	resetFlags(t)
	_, errOut := captureOutput(t)
	root := monorepoFixture(t)
	const notes = "# Root notes\n- KEEP-ROOT-RULE: deploys go through the release train.\n"
	writeFiles(t, root, map[string]string{"CLAUDE.md": notes})

	scanMonorepo(t, root, "claude")

	got := readFile(t, root, "CLAUDE.md")
	checkMarked(t, got)
	if !strings.HasPrefix(got, merger.AutoStartMarker) || !strings.HasSuffix(got, "\n\n"+notes) {
		t.Errorf("root notes not kept below the generated block:\n%s", got)
	}
	if !strings.Contains(errOut.String(), "CLAUDE.md had no argus markers") {
		t.Errorf("no adoption notice: %q", errOut)
	}
}

func TestPolicyFor(t *testing.T) {
	tests := []struct {
		rel  string
		want writePolicy
	}{
		{"CLAUDE.md", policyMerge},
		{"AGENTS.md", policyMerge},
		{".cursorrules", policyMerge},
		{".github/copilot-instructions.md", policyMerge},
		{".continue/config.yaml", policyCreateOnly},
		{".claude/settings.json", policyCreateOnly},
		{".claude/mcp.json", policyCreateOnly},
		{".claude/agents/planner.md", policyCreateOnly},
		{".claude/rules/testing.md", policyCreateOnly},
		{".claude/skills/test/SKILL.md", policyCreateOnly},
	}
	for _, tt := range tests {
		if got := policyFor(tt.rel); got != tt.want {
			t.Errorf("policyFor(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}

// checkMarked fails unless content holds one AUTO pair and ends with
// exactly one newline.
func checkMarked(t *testing.T, content string) {
	t.Helper()
	if strings.Count(content, merger.AutoStartMarker) != 1 || strings.Count(content, merger.AutoEndMarker) != 1 {
		t.Errorf("want exactly one ARGUS:AUTO pair:\n%s", content)
	}
	if !strings.HasSuffix(content, "\n") || strings.HasSuffix(content, "\n\n") {
		t.Errorf("want exactly one trailing newline, got %q", content[max(0, len(content)-20):])
	}
}

// generate runs one format into root with the default config.
func generate(t *testing.T, root, format string) {
	t.Helper()
	if err := generateOutput(root, "", format, goAnalysis(), config.DefaultConfig(), dryRun, false); err != nil {
		t.Fatal(err)
	}
}

func TestMergeFilesAreMarkedAndStable(t *testing.T) {
	files := map[string]string{"claude": "CLAUDE.md", "cursor": ".cursorrules", "copilot": ".github/copilot-instructions.md"}
	for format, rel := range files {
		t.Run(format, func(t *testing.T) {
			resetFlags(t)
			out, errOut := captureOutput(t)
			root := t.TempDir()

			generate(t, root, format)
			first := readFile(t, root, rel)
			checkMarked(t, first)
			if out.String() != "✅ Generated "+rel+"\n" {
				t.Errorf("first run printed %q", out)
			}

			for run := 2; run <= 3; run++ {
				out.Reset()
				generate(t, root, format)
				if got := readFile(t, root, rel); got != first {
					t.Errorf("run %d changed %s", run, rel)
				}
				if out.String() != "✅ "+rel+" is up to date\n" {
					t.Errorf("run %d printed %q", run, out)
				}
			}
			if errOut.Len() != 0 {
				t.Errorf("unexpected notices: %q", errOut)
			}
		})
	}
}

func TestMergeDefusesMarkerTextFromReadme(t *testing.T) {
	resetFlags(t)
	out, errOut := captureOutput(t)
	root := t.TempDir()
	analysis := goAnalysis()
	analysis.ReadmeContent = &types.ReadmeContent{Features: []string{"Replaces the " + merger.AutoStartMarker + " block"}}

	for run := 1; run <= 3; run++ {
		out.Reset()
		if err := generateOutput(root, "", "claude", analysis, config.DefaultConfig(), false, false); err != nil {
			t.Fatal(err)
		}
		got := readFile(t, root, "CLAUDE.md")
		checkMarked(t, got)
		if !strings.Contains(got, "\n- Replaces the <!--ARGUS:AUTO --> block\n") {
			t.Errorf("run %d: defused bullet missing:\n%s", run, got)
		}
		if run > 1 && out.String() != "✅ CLAUDE.md is up to date\n" {
			t.Errorf("run %d printed %q", run, out)
		}
	}
	if errOut.Len() != 0 {
		t.Errorf("unexpected notices: %q", errOut)
	}
}

func TestMergeKeepsExistingText(t *testing.T) {
	const (
		auto     = merger.AutoStartMarker + "\n# demo\nold generated body\n" + merger.AutoEndMarker
		custom   = merger.CustomStartMarker + "\n## Custom Notes\n\nKEEP-CUSTOM\n\n" + merger.CustomEndMarker
		previous = merger.CustomStartMarker + "\n## Previous Content\n\nThe following content was preserved from your original file:\n\nKEEP-PREVIOUS\n" + merger.CustomEndMarker
	)
	tests := []struct {
		name     string
		existing string
		keep     []string // must each appear once, in this order
		adopted  bool
	}{
		{"unmarked", "# demo\n\nKEEP-BODY\n", []string{merger.AutoStartMarker, "KEEP-BODY"}, true},
		{"custom only", "# demo\n\nKEEP-BODY\n\n" + custom, []string{merger.AutoStartMarker, "KEEP-BODY", "KEEP-CUSTOM"}, true},
		{"auto and custom", auto + "\n\n" + custom, []string{merger.AutoStartMarker, "KEEP-CUSTOM"}, false},
		{
			"text above, between and below",
			"# Team rules\nKEEP-TOP\n\n" + auto + "\n\nKEEP-MID\n\n" + custom + "\n\nKEEP-BOTTOM\n",
			[]string{"KEEP-TOP", merger.AutoStartMarker, "KEEP-MID", "KEEP-CUSTOM", "KEEP-BOTTOM"},
			false,
		},
		{"previous content", auto + "\n\n" + previous, []string{merger.AutoStartMarker, "KEEP-PREVIOUS"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags(t)
			out, errOut := captureOutput(t)
			root := t.TempDir()
			writeFiles(t, root, map[string]string{"CLAUDE.md": tt.existing})

			generate(t, root, "claude")
			got := readFile(t, root, "CLAUDE.md")
			checkMarked(t, got)
			last := -1
			for _, keep := range tt.keep {
				if n := strings.Count(got, keep); n != 1 {
					t.Errorf("%q appears %d times", keep, n)
				}
				if i := strings.Index(got, keep); i < last {
					t.Errorf("%q moved out of order", keep)
				} else {
					last = i
				}
			}
			if strings.Contains(got, "old generated body") {
				t.Error("the old AUTO block was kept")
			}
			notice := "ℹ️  CLAUDE.md had no argus markers: generated block added on top, your text kept below it (remove any part an older argus run wrote)\n"
			if want := map[bool]string{true: notice, false: ""}[tt.adopted]; errOut.String() != want {
				t.Errorf("stderr = %q, want %q", errOut, want)
			}

			out.Reset()
			errOut.Reset()
			generate(t, root, "claude")
			if again := readFile(t, root, "CLAUDE.md"); again != got {
				t.Errorf("second run changed the file:\n%s", again)
			}
			if out.String() != "✅ CLAUDE.md is up to date\n" || errOut.Len() != 0 {
				t.Errorf("second run printed %q / %q", out, errOut)
			}
		})
	}
}

func TestMergeLeavesMalformedMarkersAlone(t *testing.T) {
	resetFlags(t)
	out, errOut := captureOutput(t)
	root := t.TempDir()
	malformed := "# Team rules\n" + merger.AutoStartMarker + "\nonly a start\n"
	writeFiles(t, root, map[string]string{"CLAUDE.md": malformed})

	generate(t, root, "claude")
	if got := readFile(t, root, "CLAUDE.md"); got != malformed {
		t.Errorf("malformed file was changed:\n%s", got)
	}
	want := "⚠️  CLAUDE.md: unbalanced or repeated ARGUS:AUTO markers, left unchanged (fix them or use --force to overwrite)\n"
	if errOut.String() != want || out.Len() != 0 {
		t.Errorf("printed %q / %q, want only the warning", out, errOut)
	}

	const replaced = "⚠️  CLAUDE.md: unbalanced or repeated ARGUS:AUTO markers, replaced with a fresh generated file (--force); the old text was not kept\n"
	force, dryRun = true, true
	errOut.Reset()
	generate(t, root, "claude")
	if got := readFile(t, root, "CLAUDE.md"); got != malformed || errOut.String() != replaced || !strings.Contains(out.String(), "Would write to") {
		t.Errorf("dry-run --force: file %q, printed %q / %q", got, out, errOut)
	}

	dryRun = false
	errOut.Reset()
	generate(t, root, "claude")
	got := readFile(t, root, "CLAUDE.md")
	checkMarked(t, got)
	if strings.Contains(got, "only a start") || errOut.String() != replaced {
		t.Errorf("--force did not write a clean file:\n%s\nstderr %q", got, errOut)
	}

	writeFiles(t, root, map[string]string{"CLAUDE.md": merger.AutoStartMarker + "\nstale\n" + merger.AutoEndMarker + "\n"})
	errOut.Reset()
	generate(t, root, "claude")
	if got := readFile(t, root, "CLAUDE.md"); strings.Contains(got, "stale") || errOut.Len() != 0 {
		t.Errorf("--force on a healthy file: file %q, stderr %q", got, errOut)
	}
}

func TestWatchWarnsAboutMalformedMarkersOnce(t *testing.T) {
	resetFlags(t)
	_, errOut := captureOutput(t)
	watchNotified = map[string]bool{}
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"CLAUDE.md": merger.AutoEndMarker + "\n"})

	for range 3 {
		generate(t, root, "claude")
		reportSkipped()
	}
	if n := strings.Count(errOut.String(), "unbalanced or repeated ARGUS:AUTO markers"); n != 1 {
		t.Errorf("warning printed %d times, want once:\n%s", n, errOut)
	}
}

func TestMergeDisabledReplacesWholeFile(t *testing.T) {
	resetFlags(t)
	captureOutput(t)
	mergeMode = false
	root := t.TempDir()
	writeFiles(t, root, map[string]string{
		"CLAUDE.md": "KEEP-TOP\n\n" + merger.AutoStartMarker + "\nold\n" + merger.AutoEndMarker + "\n\n" +
			merger.CustomStartMarker + "\nKEEP-CUSTOM\n" + merger.CustomEndMarker + "\n",
	})

	generate(t, root, "claude")
	got := readFile(t, root, "CLAUDE.md")
	checkMarked(t, got)
	if strings.Contains(got, "KEEP") {
		t.Errorf("--merge=false kept old text:\n%s", got)
	}
}

func TestAddCustomIsIdempotentForMergeFilesOnly(t *testing.T) {
	resetFlags(t)
	captureOutput(t)
	addCustomBlock = true
	root := t.TempDir()

	for range 2 {
		for _, format := range []string{"claude", "continue", "claude-code"} {
			generate(t, root, format)
		}
	}
	if n := strings.Count(readFile(t, root, "CLAUDE.md"), merger.CustomStartMarker); n != 1 {
		t.Errorf("CLAUDE.md has %d CUSTOM blocks, want 1", n)
	}
	for _, rel := range []string{".continue/config.yaml", ".claude/agents/planner.md", ".claude/settings.json"} {
		if strings.Contains(readFile(t, root, rel), "ARGUS:") {
			t.Errorf("%s got argus markers", rel)
		}
	}
}

func TestMergeFileWithCRLFIsNotRewritten(t *testing.T) {
	resetFlags(t)
	out, _ := captureOutput(t)
	root := t.TempDir()
	generate(t, root, "claude")
	crlf := strings.ReplaceAll(readFile(t, root, "CLAUDE.md"), "\n", "\r\n")
	writeFiles(t, root, map[string]string{"CLAUDE.md": crlf})

	out.Reset()
	generate(t, root, "claude")
	if readFile(t, root, "CLAUDE.md") != crlf {
		t.Error("CRLF copy of an up-to-date file was rewritten")
	}
	if out.String() != "✅ CLAUDE.md is up to date\n" {
		t.Errorf("printed %q", out)
	}
}

func TestMergeKeepsCRLFFilesCRLF(t *testing.T) {
	resetFlags(t)
	out, errOut := captureOutput(t)
	ref := t.TempDir()
	generate(t, ref, "claude")
	block := strings.TrimSuffix(readFile(t, ref, "CLAUDE.md"), "\n")

	const (
		top   = "# Team rules\r\nKEEP-TOP\r\n\r\n"
		stale = merger.AutoStartMarker + "\r\nstale\r\n" + merger.AutoEndMarker
	)
	tests := []struct {
		name, bottom, block string
	}{
		{"all CRLF", "\r\n\r\nKEEP-BOTTOM\r\n", strings.ReplaceAll(block, "\n", "\r\n")},
		{"one LF line", "\r\n\r\nKEEP-BOTTOM\n", block},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeFiles(t, root, map[string]string{"CLAUDE.md": top + stale + tt.bottom})
			want := top + tt.block + tt.bottom
			for run := 1; run <= 2; run++ {
				out.Reset()
				generate(t, root, "claude")
				if got := readFile(t, root, "CLAUDE.md"); got != want {
					t.Errorf("run %d wrote %q, want %q", run, got, want)
				}
			}
			if out.String() != "✅ CLAUDE.md is up to date\n" || errOut.Len() != 0 {
				t.Errorf("run 2 printed %q / %q", out, errOut)
			}
		})
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	resetFlags(t)
	out, _ := captureOutput(t)
	dryRun = true
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"CLAUDE.md": "# Team rules\n"})

	for _, format := range []string{"claude", "cursor", "copilot", "continue", "claude-code"} {
		generate(t, root, format)
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || readFile(t, root, "CLAUDE.md") != "# Team rules\n" {
		t.Errorf("dry-run changed the tree: %v", entries)
	}
	for _, rel := range []string{"CLAUDE.md", ".cursorrules", ".continue/config.yaml", ".claude/settings.json"} {
		if header := "📄 Would write to " + filepath.Join(root, filepath.FromSlash(rel)) + ":"; !strings.Contains(out.String(), header) {
			t.Errorf("no preview for %s", rel)
		}
	}
}

func TestMergeWritesThroughSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("creating symlinks needs extra privileges on Windows")
	}
	resetFlags(t)
	captureOutput(t)
	root := t.TempDir()
	writeFiles(t, root, map[string]string{"docs/real.md": "# Team rules\nKEEP\n"})
	link := filepath.Join(root, "CLAUDE.md")
	if err := os.Symlink(filepath.Join("docs", "real.md"), link); err != nil {
		t.Fatal(err)
	}

	generate(t, root, "claude")
	if fi, err := os.Lstat(link); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("CLAUDE.md is no longer a symlink (err %v)", err)
	}
	got := readFile(t, root, "docs/real.md")
	checkMarked(t, got)
	if !strings.Contains(got, "KEEP") {
		t.Errorf("target lost its text:\n%s", got)
	}
}

func TestMergeReplacesFileThroughTempFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission bits and inodes work differently on Windows")
	}
	const stale = merger.AutoStartMarker + "\nstale\n" + merger.AutoEndMarker + "\n"
	tests := []struct {
		name              string
		fileMode, dirMode os.FileMode
		renamed, wantErr  bool
	}{
		{"writable file", 0640, 0755, true, false},
		{"read-only directory", 0640, 0555, false, false},
		{"read-only file", 0444, 0755, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.renamed && os.Geteuid() == 0 {
				t.Skip("root ignores permission bits")
			}
			resetFlags(t)
			captureOutput(t)
			root := t.TempDir()
			target := filepath.Join(root, "CLAUDE.md")
			writeFiles(t, root, map[string]string{"CLAUDE.md": stale})
			if err := os.Chmod(target, tt.fileMode); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(root, tt.dirMode); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(root, 0755) })
			before, err := os.Stat(target)
			if err != nil {
				t.Fatal(err)
			}

			err = generateOutput(root, "", "claude", goAnalysis(), config.DefaultConfig(), false, false)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, want an error: %v", err, tt.wantErr)
			}
			after, err := os.Stat(target)
			if err != nil {
				t.Fatal(err)
			}
			if after.Mode().Perm() != tt.fileMode {
				t.Errorf("mode = %v, want %v", after.Mode().Perm(), tt.fileMode)
			}
			if renamed := !os.SameFile(before, after); renamed != tt.renamed {
				t.Errorf("replaced by a rename = %v, want %v", renamed, tt.renamed)
			}
			got := readFile(t, root, "CLAUDE.md")
			if updated := got != stale; updated == tt.wantErr {
				t.Errorf("file updated = %v, want %v:\n%s", updated, !tt.wantErr, got)
			}
			if entries, err := os.ReadDir(root); err != nil || len(entries) != 1 {
				t.Errorf("directory holds %v (err %v), want only CLAUDE.md", entries, err)
			}
		})
	}
}
