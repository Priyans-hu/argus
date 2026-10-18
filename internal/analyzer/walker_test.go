package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// walkPaths walks root and zeroes directory sizes, which depend on how many
// entries a directory holds rather than on what the walk returns.
func walkPaths(t *testing.T, root string) []types.FileInfo {
	t.Helper()
	files, err := NewWalker(root).Walk(context.Background())
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
