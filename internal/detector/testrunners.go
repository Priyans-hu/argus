package detector

import (
	"cmp"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// jsExec runs a package binary with each package manager.
var jsExec = map[string]string{"npm": "npx", "bun": "bunx", "pnpm": "pnpm exec", "yarn": "yarn"}

// DetectTestRunners returns, per test framework and directory, how to run
// all tests and one test, sorted by Dir then Framework; root is absolute,
// files the walked list, deps the declared dependencies and commands the
// project's commands (a pytest command run through poetry or uv gives its
// prefix).
func DetectTestRunners(root string, files []types.FileInfo, deps []types.Dependency, commands []types.Command) []types.TestRunner {
	var runners []types.TestRunner
	manifests := FindManifests(files)
	src := ProjectSourceFiles(files)

	for _, m := range manifests {
		switch m.Kind {
		case "package.json":
			if r, ok := jsTestRunner(root, m, src); ok {
				runners = append(runners, r)
			}
		case "go.mod":
			if slices.ContainsFunc(src, func(f types.FileInfo) bool {
				return strings.HasSuffix(f.Name, "_test.go") && under(filepath.ToSlash(f.Path), m.Dir)
			}) {
				runners = append(runners, types.TestRunner{Framework: "go", Dir: m.Dir, RunAll: "go test ./...", RunOne: "go test -run <TestName> ./<pkg>"})
			}
		case "Cargo.toml":
			runners = append(runners, types.TestRunner{Framework: "cargo", Dir: m.Dir, RunAll: "cargo test", RunOne: "cargo test <name>"})
		}
	}

	if r, ok := pytestRunner(root, manifests, src, deps, commands); ok {
		runners = append(runners, r)
	}
	if gradle, android, tests, ok := gradleBuild(root, files); ok && tests {
		task := "test"
		if android {
			task = "testDebugUnitTest"
		}
		runners = append(runners, types.TestRunner{Framework: "gradle", RunAll: gradle + " " + task,
			RunOne: gradle + " " + task + ` --tests "<fully.qualified.Class>"`})
	}

	slices.SortFunc(runners, func(a, b types.TestRunner) int {
		return cmp.Or(cmp.Compare(a.Dir, b.Dir), cmp.Compare(a.Framework, b.Framework))
	})
	return slices.CompactFunc(runners, func(a, b types.TestRunner) bool { return a.Dir == b.Dir && a.Framework == b.Framework })
}

// jsTestRunner finds vitest, which wins, or jest in the package.json m by
// its own dependencies or a config file beside it.
func jsTestRunner(root string, m Manifest, src []types.FileInfo) (types.TestRunner, bool) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(m.Path)))
	if err != nil {
		return types.TestRunner{}, false
	}
	declared := NewDepSet(npmDeps(data))
	hasConfig := func(prefix string) bool {
		return slices.ContainsFunc(src, func(f types.FileInfo) bool {
			return strings.HasPrefix(f.Name, prefix) && path.Dir(filepath.ToSlash(f.Path)) == cmp.Or(m.Dir, ".")
		})
	}
	framework := ""
	switch {
	case declared.Has("vitest") || hasConfig("vitest.config."):
		framework = "vitest"
	case declared.Has("jest", "@jest/globals") || hasConfig("jest.config."):
		framework = "jest"
	default:
		return types.TestRunner{}, false
	}

	dir := filepath.Join(root, filepath.FromSlash(m.Dir))
	pm := DetectPackageManager(dir, root)
	exec := jsExec[pm] + " " + framework
	if framework == "vitest" {
		exec += " run"
	}
	runAll := exec
	if hasTestScript(data) {
		runAll = scriptCommand(pm, "test")
	}
	return types.TestRunner{Framework: framework, Dir: m.Dir, RunAll: runAll, RunOne: exec + " <file>"}, true
}

func hasTestScript(packageJSON []byte) bool {
	var pkg struct {
		Scripts map[string]any `json:"scripts"`
	}
	if json.Unmarshal(packageJSON, &pkg) != nil {
		return false
	}
	_, ok := pkg.Scripts["test"].(string)
	return ok
}

// pytestRunner finds pytest by a declared dependency, [tool.pytest] in a
// pyproject.toml, a pytest.ini or a conftest.py; it runs from the
// shallowest Python manifest's directory.
func pytestRunner(root string, manifests []Manifest, src []types.FileInfo, deps []types.Dependency, commands []types.Command) (types.TestRunner, bool) {
	found := NewDepSet(deps).Has("pytest") || slices.ContainsFunc(src, func(f types.FileInfo) bool {
		return f.Name == "pytest.ini" || f.Name == "conftest.py"
	})
	dir, hasDir := "", false
	for _, m := range manifests {
		if m.Kind != "pyproject.toml" && m.Kind != "requirements.txt" {
			continue
		}
		if !hasDir {
			dir, hasDir = m.Dir, true
		}
		if m.Kind == "pyproject.toml" && !found {
			data, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(m.Path)))
			found = strings.Contains(string(data), "[tool.pytest")
		}
	}
	if !found {
		return types.TestRunner{}, false
	}
	prefix := ""
	for _, c := range commands {
		for _, p := range []string{"poetry run ", "uv run "} {
			if strings.HasPrefix(c.Name, p+"pytest") || strings.HasPrefix(c.Command, p+"pytest") {
				prefix = cmp.Or(prefix, p)
			}
		}
	}
	return types.TestRunner{Framework: "pytest", Dir: dir, RunAll: prefix + "pytest", RunOne: prefix + "pytest <file> -k <name>"}, true
}

// under reports whether the slash path rel is inside the slash directory
// dir, the root when dir is empty.
func under(rel, dir string) bool {
	return dir == "" || strings.HasPrefix(rel, dir+"/")
}

// TestRunnerCommands returns each runner's commands not already in
// commands: running all tests, then one test, suffixed with (in <dir>)
// outside the root.
func TestRunnerCommands(runners []types.TestRunner, commands []types.Command) []types.Command {
	var added []types.Command
	have := func(name string) bool {
		return slices.ContainsFunc(slices.Concat(commands, added), func(c types.Command) bool { return c.Name == name })
	}
	for _, r := range runners {
		suffix := ""
		if r.Dir != "" {
			suffix = " (in " + r.Dir + ")"
		}
		for _, c := range []types.Command{
			{Name: r.RunAll + suffix, Description: "Run all " + r.Framework + " tests"},
			{Name: r.RunOne + suffix, Description: "Run one " + r.Framework + " test"},
		} {
			if !have(c.Name) {
				added = append(added, c)
			}
		}
	}
	return added
}
