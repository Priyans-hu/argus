package analyzer

import (
	"context"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// walkPaths walks root and zeroes directory sizes, which depend on how many
// entries a directory holds rather than on what the walk returns.
func walkPaths(t *testing.T, root string) []types.FileInfo {
	t.Helper()
	files, err := NewWalker(root, nil).Walk(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i := range files {
		if files[i].IsDir {
			files[i].Size = 0
		}
	}
	return files
}

func TestWalkSkipsGeneratedOutputs(t *testing.T) {
	project := map[string]string{
		"go.mod":                   "module example.com/demo\n",
		"main.go":                  "package main\n",
		"packages/a/index.js":      "export {}\n",
		".github/workflows/ci.yml": "name: ci\n",
		"docs/claude.md":           "# notes\n",
		".claudex/notes.md":        "# notes\n",
	}
	outputs := map[string]string{
		"CLAUDE.md":                       "# demo\n",
		"AGENTS.md":                       "# demo\n",
		".cursorrules":                    "rules\n",
		".github/copilot-instructions.md": "# demo\n",
		".claude/settings.json":           "{}\n",
		".claude/agents/planner.md":       "---\n",
		".continue/config.yaml":           "name: demo\n",
		"packages/a/CLAUDE.md":            "# a\n",
		"packages/a/.claude/rules/x.md":   "# x\n",
	}

	clean, generated := t.TempDir(), t.TempDir()
	writeTree(t, clean, project)
	writeTree(t, generated, project)
	writeTree(t, generated, outputs)

	want := walkPaths(t, clean)
	got := walkPaths(t, generated)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("walk with outputs differs from walk without them:\ngot  %v\nwant %v", got, want)
	}

	for _, rel := range []string{"docs/claude.md", ".claudex/notes.md", "packages/a/index.js"} {
		if !slices.ContainsFunc(got, func(f types.FileInfo) bool { return f.Path == filepath.FromSlash(rel) }) {
			t.Errorf("walk lost %s", rel)
		}
	}
}

func TestWalkMatchesIgnoreRulesByName(t *testing.T) {
	root := t.TempDir()
	walked := []string{
		"internal/binder/x.go", "internal/objects/y.go", "lib/distance/z.go", "src/binary_search.go",
		"docs/build-notes.md", ".github/workflows/ci.yml", ".gitignore", ".githooks/pre-commit",
		"app/build.gradle.kts", "x/scripts/old/o.txt", "e/a/b/f.txt", "x/docs/legacy/l.md",
	}
	skipped := []string{
		"node_modules/a.js", "web/node_modules/b.js", ".git/config", "build/out.js", "src/build/gen.go",
		"bin/tool", "dist/x.js", "target/y", "vendor/z.go", "app.log", "package-lock.json",
		"secret.txt", "src/secret.txt", "docs/legacy/l.md",
		"generated.go", "pkg/generated.go", "c/x.tmp", "scripts/old/o.txt", "a/b/f.txt",
	}
	files := map[string]string{".gitignore": "secret.txt\ndocs/legacy/\n"}
	for _, rel := range append(walked, skipped...) {
		if files[rel] == "" {
			files[rel] = "x\n"
		}
	}
	writeTree(t, root, files)

	cfg := &types.Config{Ignore: []string{"generated.go", "*.tmp", "/scripts/old", "a/b"}}
	got, err := NewWalker(root, cfg).Walk(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]bool{}
	for _, f := range got {
		paths[filepath.ToSlash(f.Path)] = true
	}
	for _, rel := range walked {
		if !paths[rel] {
			t.Errorf("%s was not walked", rel)
		}
	}
	for _, rel := range skipped {
		if paths[rel] {
			t.Errorf("%s was walked", rel)
		}
	}
}

func TestIgnoreReachesEveryAnalyzer(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"package.json":                `{"name": "m", "workspaces": ["packages/*"]}`,
		"src/a.ts":                    "export const a = 1;\n",
		"legacy/old.ts":               "export const o = 1;\n",
		"packages/web/package.json":   `{"name": "web"}`,
		"packages/web/src/a.ts":       "export const a = 1;\n",
		"packages/web/legacy/old.ts":  "export const o = 1;\n",
		"packages/web/legacy/more.ts": "export const m = 1;\n",
	})
	ctx := context.Background()
	hasLegacy := func(a *types.Analysis) bool {
		return slices.ContainsFunc(a.Structure.Directories, func(d types.Directory) bool { return d.Path == "legacy" })
	}

	control, err := NewParallelAnalyzer(root, nil).Analyze(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !hasLegacy(control) {
		t.Fatalf("control analysis lists no legacy directory: %+v", control.Structure.Directories)
	}

	cfg := &types.Config{Ignore: []string{"legacy"}}
	sequential, err := NewAnalyzer(root, cfg).Analyze(ctx)
	if err != nil {
		t.Fatal(err)
	}
	parallel, err := NewParallelAnalyzer(root, cfg).Analyze(ctx)
	if err != nil {
		t.Fatal(err)
	}
	watch, err := NewIncrementalAnalyzer(root, cfg).AnalyzeFull(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for name, a := range map[string]*types.Analysis{"sequential": sequential, "parallel": parallel, "watch": watch} {
		if hasLegacy(a) {
			t.Errorf("%s analysis lists the ignored legacy directory", name)
		}
	}

	for _, par := range []bool{false, true} {
		results := NewMonorepoAnalyzer(root, cfg, par, 2, nil).AnalyzeWorkspaces(ctx, control.MonorepoInfo)
		if len(results) != 1 || results[0].Analysis == nil {
			t.Fatalf("workspaces = %+v, want packages/web", results)
		}
		if hasLegacy(results[0].Analysis) {
			t.Errorf("workspace analysis (parallel %v) lists the ignored legacy directory", par)
		}
	}
}
