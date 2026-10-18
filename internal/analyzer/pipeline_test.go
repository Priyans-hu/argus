package analyzer

import (
	"context"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// richProject exercises every pipeline step: a Go CLI (project tools), JS
// with React hooks, Python with a framework import, manifests and a README.
var richProject = map[string]string{
	"go.mod": "module example.com/rich\n\ngo 1.24\n\nrequire github.com/spf13/cobra v1.8.0\n",
	"cmd/rich/main.go": "package main\n\nimport (\n\t\"fmt\"\n\n\t\"github.com/spf13/cobra\"\n)\n\n" +
		"func main() {\n\tcmd := &cobra.Command{Use: \"rich\"}\n\tcmd.Flags().BoolP(\"verbose\", \"v\", false, \"verbose output\")\n" +
		"\tif err := cmd.Execute(); err != nil {\n\t\tfmt.Println(err)\n\t}\n}\n",
	"internal/store/store.go": "package store\n\nimport (\n\t\"errors\"\n\t\"sync\"\n)\n\nvar errMissing = errors.New(\"missing\")\n\n" +
		"type Store struct {\n\tmu sync.Mutex\n}\n\nfunc (s *Store) Get(err error) bool {\n\ts.mu.Lock()\n\tdefer s.mu.Unlock()\n\treturn errors.Is(err, errMissing)\n}\n",
	"internal/store/store_test.go": "package store\n\nimport \"testing\"\n\nfunc TestGet(t *testing.T) {\n\tt.Run(\"missing\", func(t *testing.T) {})\n}\n",
	"package.json": `{"name": "rich", "scripts": {"dev": "vite", "build": "vite build", "test": "vitest"},
		"dependencies": {"react": "18.2.0"}, "devDependencies": {"vitest": "2.0.0"}}`,
	"web/App.js":     "var useState = require(\"react\").useState;\n\nfunction App() {\n  var n = useState(0);\n  useCart();\n  return n;\n}\n",
	"web/useCart.js": "function useCart() {\n  return useState([]);\n}\n",
	"pyproject.toml": "[project]\nname = \"rich\"\ndependencies = [\"flask>=3\"]\n",
	"api/app.py":     "from flask import Flask\n\napp = Flask(__name__)\n\n\n@app.route(\"/health\")\ndef health():\n    return \"ok\"\n",
	"README.md":      "# Rich\n\nA command line tool for rich fixtures.\n\n## Usage\n\n```bash\nrich --verbose\n```\n",
	"Makefile":       "build:\n\tgo build -o bin/rich ./cmd/rich\n\ntest:\n\tgo test ./...\n",
}

// writeRichProject writes richProject into a git repository named rich.
func writeRichProject(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "rich")
	writeTree(t, root, richProject)
	commitTree(t, root)
	return root
}

func hasPattern(patterns []types.PatternInfo, name string) bool {
	return slices.ContainsFunc(patterns, func(p types.PatternInfo) bool { return p.Name == name })
}

func TestAnalyzersRunOnePipeline(t *testing.T) {
	root := writeRichProject(t)

	first, _ := analyzeBoth(t, root, nil)
	switch {
	case len(first.ProjectTools) == 0:
		t.Error("no project tools")
	case first.CodePatterns == nil || !hasPattern(first.CodePatterns.StateManagement, "useCart"):
		t.Errorf("React hooks did not join state management: %+v", first.CodePatterns)
	case !hasPattern(first.CodePatterns.PythonPatterns, "Flask"):
		t.Errorf("Python AST patterns missing: %+v", first.CodePatterns.PythonPatterns)
	case first.GitConventions == nil || first.ReadmeContent == nil:
		t.Error("git conventions or README missing")
	case first.RootPath != root || first.ProjectName != "rich":
		t.Errorf("root %q name %q, want %q rich", first.RootPath, first.ProjectName, root)
	}

	for i := range 20 {
		if sequential, _ := analyzeBoth(t, root, nil); !reflect.DeepEqual(sequential, first) {
			t.Fatalf("run %d differs from run 0", i+1)
		}
	}
}

func TestWatchAnalysesMatchScan(t *testing.T) {
	root := writeRichProject(t)
	ctx := context.Background()

	ia := NewIncrementalAnalyzer(root, nil)
	full, err := ia.AnalyzeFull(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, scan := analyzeBoth(t, root, nil)
	if !reflect.DeepEqual(full, scan) {
		t.Fatalf("watch's full analysis differs from scan's:\nwatch %+v\nscan  %+v", full, scan)
	}

	changed := filepath.Join(root, "web", "useTotal.js")
	writeTree(t, root, map[string]string{"web/useTotal.js": "function Total() {\n  var total = useTotal();\n  return total;\n}\n"})
	incremental, impacts, err := ia.AnalyzeIncremental(ctx, changed)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(impacts, ImpactConventions) {
		t.Fatalf("impacts = %v, want conventions", impacts)
	}
	_, fresh := analyzeBoth(t, root, nil)
	if !reflect.DeepEqual(incremental.CodePatterns, fresh.CodePatterns) || !hasPattern(fresh.CodePatterns.StateManagement, "useTotal") {
		t.Errorf("code patterns after the change:\nwatch %+v\nscan  %+v", incremental.CodePatterns, fresh.CodePatterns)
	}
	if !reflect.DeepEqual(incremental.ProjectTools, fresh.ProjectTools) {
		t.Errorf("project tools after the change:\nwatch %+v\nscan  %+v", incremental.ProjectTools, fresh.ProjectTools)
	}
}

func TestCloneAnalysisKeepsProjectTools(t *testing.T) {
	src := &types.Analysis{ProjectTools: []types.ProjectTool{{Name: "rich", Description: "tool"}}}
	clone := (&IncrementalAnalyzer{}).cloneAnalysis(src)
	if !reflect.DeepEqual(clone.ProjectTools, src.ProjectTools) {
		t.Errorf("clone project tools = %+v, want %+v", clone.ProjectTools, src.ProjectTools)
	}
}
