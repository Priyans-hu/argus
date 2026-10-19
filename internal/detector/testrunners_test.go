package detector

import (
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func detectRunners(t *testing.T, files map[string]string, commands []types.Command) []types.TestRunner {
	t.Helper()
	root, infos := writeProject(t, files)
	return stableResult(t, func() []types.TestRunner {
		return DetectTestRunners(root, infos, ReadDependencies(root, infos), commands)
	})
}

func TestDetectTestRunners(t *testing.T) {
	vitest := `{"devDependencies": {"vitest": "2.0.0"}}`
	spindle := map[string]string{"gradlew": "", "settings.gradle.kts": "", "app/src/main/AndroidManifest.xml": "<manifest/>"}
	tests := []struct {
		name     string
		files    map[string]string
		commands []types.Command
		want     []types.TestRunner
	}{
		{"vitest with bun and a test script", map[string]string{"package.json": `{"scripts": {"test": "vitest"}, "devDependencies": {"vitest": "2.0.0"}}`, "bun.lock": ""}, nil,
			[]types.TestRunner{{Framework: "vitest", RunAll: "bun run test", RunOne: "bunx vitest run <file>"}}},
		{"vitest with npm", map[string]string{"package.json": `{"scripts": {"test": "vitest"}, "devDependencies": {"vitest": "2.0.0"}}`}, nil,
			[]types.TestRunner{{Framework: "vitest", RunAll: "npm test", RunOne: "npx vitest run <file>"}}},
		{"vitest with pnpm, no test script", map[string]string{"package.json": vitest, "pnpm-lock.yaml": ""}, nil,
			[]types.TestRunner{{Framework: "vitest", RunAll: "pnpm exec vitest run", RunOne: "pnpm exec vitest run <file>"}}},
		{"vitest config with yarn", map[string]string{"package.json": "{}", "vitest.config.ts": "", "yarn.lock": ""}, nil,
			[]types.TestRunner{{Framework: "vitest", RunAll: "yarn vitest run", RunOne: "yarn vitest run <file>"}}},
		{"jest", map[string]string{"package.json": `{"devDependencies": {"jest": "29.0.0"}}`}, nil,
			[]types.TestRunner{{Framework: "jest", RunAll: "npx jest", RunOne: "npx jest <file>"}}},
		{"vitest wins over jest", map[string]string{"package.json": `{"devDependencies": {"jest": "29.0.0", "vitest": "2.0.0"}}`}, nil,
			[]types.TestRunner{{Framework: "vitest", RunAll: "npx vitest run", RunOne: "npx vitest run <file>"}}},
		{"vitest in a sub-directory", map[string]string{"package.json": "{}", "web/package.json": vitest}, nil,
			[]types.TestRunner{{Framework: "vitest", Dir: "web", RunAll: "npx vitest run", RunOne: "npx vitest run <file>"}}},
		{"go", map[string]string{"go.mod": "module m\n", "a/a_test.go": "package a\n"}, nil,
			[]types.TestRunner{{Framework: "go", RunAll: "go test ./...", RunOne: "go test -run <TestName> ./<pkg>"}}},
		{"go without tests", map[string]string{"go.mod": "module m\n", "a/a.go": "package a\n"}, nil, nil},
		{"pytest", map[string]string{"requirements.txt": "pytest==8.0.0\n"}, nil,
			[]types.TestRunner{{Framework: "pytest", RunAll: "pytest", RunOne: "pytest <file> -k <name>"}}},
		{"pytest through poetry", map[string]string{"pyproject.toml": "[tool.pytest.ini_options]\n"}, []types.Command{{Name: "poetry run pytest"}},
			[]types.TestRunner{{Framework: "pytest", RunAll: "poetry run pytest", RunOne: "poetry run pytest <file> -k <name>"}}},
		{"pytest through uv", map[string]string{"pyproject.toml": "[project]\n", "tests/conftest.py": ""}, []types.Command{{Name: "uv run pytest -q"}},
			[]types.TestRunner{{Framework: "pytest", RunAll: "uv run pytest", RunOne: "uv run pytest <file> -k <name>"}}},
		{"cargo in src-tauri", map[string]string{"package.json": "{}", "src-tauri/Cargo.toml": "[package]\nname = \"app\"\n"}, nil,
			[]types.TestRunner{{Framework: "cargo", Dir: "src-tauri", RunAll: "cargo test", RunOne: "cargo test <name>"}}},
		{"gradle", map[string]string{"build.gradle.kts": "", "src/test/java/ATest.java": ""}, nil,
			[]types.TestRunner{{Framework: "gradle", RunAll: "gradle test", RunOne: `gradle test --tests "<fully.qualified.Class>"`}}},
		{"Android", merge(spindle, map[string]string{"app/src/test/java/ATest.kt": ""}), nil,
			[]types.TestRunner{{Framework: "gradle", RunAll: "./gradlew testDebugUnitTest", RunOne: `./gradlew testDebugUnitTest --tests "<fully.qualified.Class>"`}}},
		{"Android without tests", spindle, nil, nil},
		{"sorted by Dir then Framework", map[string]string{"go.mod": "module m\n", "x_test.go": "", "package.json": vitest, "web/package.json": vitest}, nil,
			[]types.TestRunner{
				{Framework: "go", RunAll: "go test ./...", RunOne: "go test -run <TestName> ./<pkg>"},
				{Framework: "vitest", RunAll: "npx vitest run", RunOne: "npx vitest run <file>"},
				{Framework: "vitest", Dir: "web", RunAll: "npx vitest run", RunOne: "npx vitest run <file>"},
			}},
	}
	for _, tt := range tests {
		if got := detectRunners(t, tt.files, tt.commands); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s:\n got %+v\nwant %+v", tt.name, got, tt.want)
		}
	}
}

func merge(a, b map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func TestTestRunnerCommands(t *testing.T) {
	runners := []types.TestRunner{
		{Framework: "go", RunAll: "go test ./...", RunOne: "go test -run <TestName> ./<pkg>"},
		{Framework: "cargo", Dir: "src-tauri", RunAll: "cargo test", RunOne: "cargo test <name>"},
	}
	got := commandNames(TestRunnerCommands(runners, []types.Command{{Name: "go test ./..."}, {Name: "cargo test (in src-tauri)"}}))
	want := []string{"go test -run <TestName> ./<pkg>", "cargo test <name> (in src-tauri)"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("commands = %v, want %v", got, want)
	}
}
