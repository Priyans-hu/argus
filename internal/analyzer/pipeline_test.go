package analyzer

import (
	"context"
	"path"
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

// patternSource makes every detector that reads only project source report
// something: component naming (conventions), console logging (patterns),
// React context (framework patterns), goroutines (code patterns), a custom
// hook (JS AST), a Flask import (Python AST) and model weights (ML).
var patternSource = map[string]string{
	"components/UserCard.tsx": reactComponent("UserCard"),
	"components/UserList.tsx": reactComponent("UserList"),
	"components/UserForm.tsx": reactComponent("UserForm"),
	"cart.js":                 "function Cart() {\n  return useCart();\n}\n",
	"app.py":                  "from flask import Flask\n\napp = Flask(__name__)\n",
	"model.pt":                "weights",
	"worker.go":               "package worker\n\nfunc Start() {\n\tgo func() {}()\n}\n",
}

func reactComponent(name string) string {
	return "export const " + name + " = () => {\n  const user = useContext(UserContext);\n  console.log(user);\n  return <div>{user}</div>;\n};\n"
}

// reactProject returns a React project with patternSource under dir, or
// without it when dir is empty.
func reactProject(dir string) map[string]string {
	files := map[string]string{
		"package.json": `{"name": "app", "dependencies": {"react": "18.2.0"}}`,
		"src/index.ts": "export const x = 1;\n",
	}
	if dir != "" {
		for rel, content := range patternSource {
			files[path.Join(dir, rel)] = content
		}
	}
	return files
}

func TestPatternDetectorsReadOnlyProjectSource(t *testing.T) {
	analyze := func(root string, files map[string]string) *types.Analysis {
		writeTree(t, root, files)
		a, _ := analyzeBoth(t, root, nil)
		return a
	}
	same := func(name string, got, want *types.Analysis) {
		t.Helper()
		if !reflect.DeepEqual(got.Conventions, want.Conventions) || !reflect.DeepEqual(got.CodePatterns, want.CodePatterns) {
			t.Errorf("%s:\nconventions %+v\nwant        %+v\npatterns    %+v\nwant        %+v",
				name, got.Conventions, want.Conventions, got.CodePatterns, want.CodePatterns)
		}
	}

	baseline := analyze(t.TempDir(), reactProject(""))
	control := analyze(t.TempDir(), reactProject("src"))
	hasConvention := func(category string) bool {
		return slices.ContainsFunc(control.Conventions, func(c types.Convention) bool { return c.Category == category })
	}
	cp := control.CodePatterns
	for what, found := range map[string]bool{
		"component naming":        hasConvention("naming"),
		"console logging":         hasConvention("logging"),
		"React framework pattern": hasConvention("react"),
		"goroutines":              hasPattern(cp.GoPatterns, "go func"),
		"JS AST hook":             hasPattern(cp.StateManagement, "useCart"),
		"Python AST import":       hasPattern(cp.PythonPatterns, "Flask"),
		"model weights":           len(cp.MLPatterns) > 0,
	} {
		if !found {
			t.Errorf("the control analysis misses the %s", what)
		}
	}

	for _, dir := range []string{"docs", "examples", "testdata", "vendor", "__mocks__", "fixtures", "generated", "apps/docs", "packages/ui/examples"} {
		same(dir+"/", analyze(t.TempDir(), reactProject(dir)), baseline)
	}
	// the JS and Python AST detectors skip test files, which feed only testing patterns
	withTests := reactProject("")
	withTests["src/cart.test.js"] = patternSource["cart.js"]
	withTests["tests/test_app.py"] = patternSource["app.py"]
	if got := analyze(t.TempDir(), withTests); !reflect.DeepEqual(got.CodePatterns, baseline.CodePatterns) {
		t.Errorf("test files changed the code patterns: %+v", got.CodePatterns)
	}
	// T3-D18: the scan root's own name is never judged
	for _, name := range []string{"docs", "examples"} {
		same("root named "+name, analyze(filepath.Join(t.TempDir(), name), reactProject("src")), control)
	}

	// a workspace in apps/docs is skipped by the root analysis and scanned
	// fully as its own root
	root := t.TempDir()
	files := map[string]string{"package.json": `{"name": "m", "workspaces": ["apps/*"]}`}
	for rel, content := range reactProject("src") {
		files["apps/docs/"+rel] = content
	}
	mono := analyze(root, files)
	if hasPattern(mono.CodePatterns.GoPatterns, "go func") {
		t.Error("the root analysis reads apps/docs")
	}
	results := NewMonorepoAnalyzer(root, nil, false, 2).AnalyzeWorkspaces(context.Background(), mono.MonorepoInfo)
	if len(results) != 1 || results[0].Analysis == nil {
		t.Fatalf("workspaces = %+v, want apps/docs", results)
	}
	same("workspace apps/docs", results[0].Analysis, control)
}

// React hooks join state management only with React or Preact declared;
// richProject declares React and gets them.
func TestReactHooksNeedReact(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{"package.json": `{"name": "app"}`, "src/cart.js": patternSource["cart.js"]})
	if a, _ := analyzeBoth(t, root, nil); hasPattern(a.CodePatterns.StateManagement, "useCart") {
		t.Errorf("hooks reported without React: %+v", a.CodePatterns.StateManagement)
	}
}

func TestGoLibrariesAreFiledByCategory(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"go.mod":        "module example.com/m\n\ngo 1.24\n",
		"api/api.go":    "package api\n\nimport (\n\t_ \"github.com/gin-gonic/gin\"\n\t_ \"gorm.io/gorm\"\n\t_ \"github.com/spf13/cobra\"\n)\n",
		"api/a_test.go": "package api\n\nimport _ \"github.com/stretchr/testify/assert\"\n",
	})
	a, _ := analyzeBoth(t, root, nil)
	cp := a.CodePatterns
	for _, c := range []struct {
		field          []types.PatternInfo
		name, category string
	}{
		{cp.Routing, "gin", "routing"},
		{cp.DatabaseORM, "gorm", "database"},
		{cp.Testing, "testify", "testing"},
		{cp.GoPatterns, "cobra", "go-cli"},
	} {
		if !slices.ContainsFunc(c.field, func(p types.PatternInfo) bool { return p.Name == c.name && p.Category == c.category }) {
			t.Errorf("%s (%s) not filed: %+v", c.name, c.category, c.field)
		}
	}
}
