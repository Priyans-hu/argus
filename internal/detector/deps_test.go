package detector

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// readDeps writes files and reads their dependencies 50 times, failing
// unless every run agrees.
func readDeps(t *testing.T, files map[string]string) []types.Dependency {
	t.Helper()
	root, infos := writeProject(t, files)
	return stableResult(t, func() []types.Dependency { return ReadDependencies(root, infos) })
}

func TestReadDependenciesFromPackageJSON(t *testing.T) {
	got := readDeps(t, map[string]string{
		"package.json": `{"dependencies": {"a": "1.0.0"}, "devDependencies": {"b": "2.0.0"},
			"peerDependencies": {"react": ">=18", "c": "^3"}, "optionalDependencies": {"d": "4.0.0"}}`,
	})
	want := []types.Dependency{
		{Name: "b", Version: "2.0.0", Type: "dev"},
		{Name: "d", Version: "4.0.0", Type: "optional"},
		{Name: "c", Version: "^3", Type: "peer"},
		{Name: "react", Version: ">=18", Type: "peer"},
		{Name: "a", Version: "1.0.0", Type: "runtime"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dependencies = %+v, want %+v", got, want)
	}
}

func TestReadDependenciesFromGoModDirectRequiresOnly(t *testing.T) {
	got := readDeps(t, map[string]string{
		"go.mod": "module example.com/m\n\ngo 1.24\n\n// github.com/gin-gonic/gin is not used\nrequire github.com/a/single v1.0.0\n\n" +
			"require (\n\tgithub.com/b/direct v1.2.3\n\tgithub.com/c/indirect v0.1.0 // indirect\n\tgithub.com/aws/x/internal/y v1.0.0\n)\n\n" +
			"replace github.com/b/direct => ../direct\n",
	})
	want := []types.Dependency{
		{Name: "github.com/a/single", Version: "v1.0.0", Type: "runtime"},
		{Name: "github.com/b/direct", Version: "v1.2.3", Type: "runtime"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dependencies = %+v, want %+v", got, want)
	}
}

func TestReadDependenciesFromPython(t *testing.T) {
	got := readDeps(t, map[string]string{
		"pyproject.toml": `[project]
dependencies = ["fastapi>=0.110", "Pydantic[email]==2.5.0", "uvicorn ; python_version>='3.9'", "requests (>=2.8)"]

[project.optional-dependencies]
dev = ["pytest-mock>=3"]

[tool.poetry.dependencies]
python = "^3.11"
Django_Extensions = { version = "^3", optional = true }

[tool.poetry.dev-dependencies]
black = "^24"

[tool.poetry.group.test.dependencies]
coverage = "^7"

[dependency-groups]
lint = ["ruff>=0.4", {include-group = "test"}]
`,
		"requirements.txt": "# comment\n-r other.txt\n--index-url https://example.com/simple\n-e .\n" +
			"git+https://example.com/repo.git#egg=zzz\nlocal/pkg\n./local/pkg\nzzz @ https://example.com/zzz.whl\n" +
			"Flask[async]==3.0.0 ; python_version >= \"3.8\"\nSQLAlchemy>=2,<3\npython_dotenv  # inline comment\n" +
			"Django[argon2]>=4.2 ; python_version>\"3.8\"\nhashed==1.0 \\\n    --hash=sha256:abc\n",
		"requirements-dev.txt": "pytest==8.0.0\n",
		"requirements.lock":    "shouldnotappear==1.0\n",
	})
	want := []types.Dependency{
		{Name: "black", Version: "^24", Type: "dev"},
		{Name: "coverage", Version: "^7", Type: "dev"},
		{Name: "ruff", Version: ">=0.4", Type: "dev"},
		{Name: "pytest-mock", Version: ">=3", Type: "optional"},
		{Name: "django", Version: ">=4.2", Type: "runtime"},
		{Name: "django-extensions", Version: "^3", Type: "runtime"},
		{Name: "fastapi", Version: ">=0.110", Type: "runtime"},
		{Name: "flask", Version: "3.0.0", Type: "runtime"},
		{Name: "hashed", Version: "1.0", Type: "runtime"},
		{Name: "pydantic", Version: "2.5.0", Type: "runtime"},
		{Name: "pytest", Version: "8.0.0", Type: "runtime"},
		{Name: "python-dotenv", Version: "", Type: "runtime"},
		{Name: "requests", Version: ">=2.8", Type: "runtime"},
		{Name: "sqlalchemy", Version: ">=2,<3", Type: "runtime"},
		{Name: "uvicorn", Version: "", Type: "runtime"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dependencies =\n%+v\nwant\n%+v", got, want)
	}
}

func TestReadDependenciesFromCargo(t *testing.T) {
	got := readDeps(t, map[string]string{
		"Cargo.toml": `[package]
name = "x"

[dependencies]
serde = { version = "1", features = ["derive"] }
renamed = { package = "real-name", version = "2" }
local = { path = "../local" }
Serde_Json = "1"

[dev-dependencies]
mockall = "0.12"

[build-dependencies]
cc = "1"

[workspace.dependencies]
anyhow = "1"
serde = "0.9"
`,
	})
	want := []types.Dependency{
		{Name: "cc", Version: "1", Type: "build"},
		{Name: "mockall", Version: "0.12", Type: "dev"},
		{Name: "anyhow", Version: "1", Type: "runtime"},
		{Name: "local", Version: "", Type: "runtime"},
		{Name: "real-name", Version: "2", Type: "runtime"},
		{Name: "serde", Version: "1", Type: "runtime"},
		{Name: "serde-json", Version: "1", Type: "runtime"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dependencies = %+v, want %+v", got, want)
	}
}

func TestReadDependenciesDepthAndScope(t *testing.T) {
	pkg := func(name string) string { return `{"dependencies": {"` + name + `": "1.0.0"}}` }
	got := readDeps(t, map[string]string{
		"package.json":               `{"dependencies": {"root": "1.0.0", "shared": "1.0.0"}}`,
		"web/package.json":           `{"dependencies": {"w1": "1.0.0", "shared": "2.0.0"}}`,
		"packages/app/package.json":  pkg("w2"),
		"a/b/c/package.json":         pkg("w3"),
		"docs/package.json":          pkg("wdocs"),
		"apps/docs/package.json":     pkg("wappdocs"),
		"examples/x/package.json":    pkg("wexample"),
		"vendor/x/package.json":      pkg("wvendor"),
		"third_party/t/package.json": pkg("wthird"),
	})
	var names []string
	for _, d := range got {
		names = append(names, d.Name+"@"+d.Version)
	}
	if want := []string{"root@1.0.0", "shared@1.0.0", "w1@1.0.0", "w2@1.0.0"}; !reflect.DeepEqual(names, want) {
		t.Errorf("dependencies = %v, want %v", names, want)
	}
}

func TestReadDependenciesInRootNamedDocs(t *testing.T) {
	root := filepath.Join(t.TempDir(), "docs")
	if err := os.MkdirAll(root, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"dependencies": {"site": "1.0.0"}}`), 0644); err != nil {
		t.Fatal(err)
	}
	got := ReadDependencies(root, []types.FileInfo{{Path: "package.json", Name: "package.json", Extension: ".json"}})
	if len(got) != 1 || got[0].Name != "site" {
		t.Errorf("dependencies = %+v, want site", got)
	}
}

func TestDepSet(t *testing.T) {
	set := NewDepSet([]types.Dependency{
		{Name: "react", Version: ">=18", Type: "peer"},
		{Name: "next", Version: "14.0.0", Type: "runtime"},
		{Name: "@trpc/server", Version: "10.0.0", Type: "runtime"},
		{Name: "github.com/labstack/echo/v4", Version: "v4.1.0", Type: "runtime"},
		{Name: "github.com/spf13/cobra-cli", Version: "v1.0.0", Type: "runtime"},
	})
	checks := []struct {
		name string
		got  bool
		want bool
	}{
		{"Has next", set.Has("express", "next"), true},
		{"Has nothing", set.Has("express"), false},
		{"HasPrefix @trpc/", set.HasPrefix("@trpc/"), true},
		{"HasPrefix @tanstack/", set.HasPrefix("@tanstack/"), false},
		{"HasModule echo", set.HasModule("github.com/labstack/echo"), true},
		{"HasModule cobra", set.HasModule("github.com/spf13/cobra"), false},
		{"peer version empty", set["react"] == "", true},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %v, want %v", c.name, c.got, c.want)
		}
	}
}

func TestReadDependenciesFromRequirementsDirectory(t *testing.T) {
	got := readDeps(t, map[string]string{
		"requirements/base.txt":         "Django==4.2.0\nrequests>=2.31\n",
		"requirements/dev.in":           "-r base.txt\npytest\n",
		"requirements/README.md":        "docs\n",
		"backend/requirements/prod.txt": "gunicorn==21.0.0\n",
	})
	var names []string
	for _, d := range got {
		names = append(names, d.Name)
	}
	if want := []string{"django", "gunicorn", "pytest", "requests"}; !reflect.DeepEqual(names, want) {
		t.Errorf("dependencies = %v, want %v", names, want)
	}

	// a Django app laid out this way keeps its library patterns
	root, infos := writeProject(t, map[string]string{
		"requirements/base.txt": "Django==4.2.0\nrequests>=2.31\n",
		"app/models.py":         "from django.db import models\n\nclass A(models.Model):\n    pass\n",
		"app/api.py":            "import requests\n\nrequests.get(\"x\")\n",
	})
	cp := NewCodePatternDetector(root, infos, ReadDependencies(root, infos)).Detect()
	for _, name := range []string{"django.db", "models.Model"} {
		if patternExamples(cp.DatabaseORM, name) == nil {
			t.Errorf("%s missing: %+v", name, cp.DatabaseORM)
		}
	}
	if patternExamples(cp.DataFetching, "requests.") == nil {
		t.Errorf("requests. missing: %+v", cp.DataFetching)
	}
}
