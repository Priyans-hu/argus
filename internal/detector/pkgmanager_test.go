package detector

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectPackageManager(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		dir   string // package directory, relative to the root
		want  string
	}{
		{"bun.lock", map[string]string{"package.json": "{}", "bun.lock": ""}, "", "bun"},
		{"bun.lockb", map[string]string{"package.json": "{}", "bun.lockb": ""}, "", "bun"},
		{"pnpm-lock.yaml", map[string]string{"package.json": "{}", "pnpm-lock.yaml": ""}, "", "pnpm"},
		{"yarn.lock", map[string]string{"package.json": "{}", "yarn.lock": ""}, "", "yarn"},
		{"package-lock.json", map[string]string{"package.json": "{}", "package-lock.json": ""}, "", "npm"},
		{"npm-shrinkwrap.json", map[string]string{"package.json": "{}", "npm-shrinkwrap.json": ""}, "", "npm"},
		{"packageManager beats a stray lockfile", map[string]string{"package.json": `{"packageManager": "pnpm@9.1.0"}`, "package-lock.json": ""}, "", "pnpm"},
		{"packageManager with a hash", map[string]string{"package.json": `{"packageManager": "pnpm@9.1.0+sha512.abc"}`}, "", "pnpm"},
		{"devEngines.packageManager.name", map[string]string{"package.json": `{"devEngines": {"packageManager": {"name": "yarn", "version": "4.1.0"}}}`}, "", "yarn"},
		{"unknown name falls through", map[string]string{"package.json": `{"packageManager": "deno@2.0.0"}`, "yarn.lock": ""}, "", "yarn"},
		{"devEngines array falls through", map[string]string{"package.json": `{"devEngines": {"packageManager": [{"name": "pnpm"}]}}`, "bun.lock": ""}, "", "bun"},
		{"nothing", map[string]string{"package.json": "{}"}, "", "npm"},
		{"pnpm-workspace.yaml alone", map[string]string{"package.json": "{}", "pnpm-workspace.yaml": ""}, "", "pnpm"},
		{"workspace inherits the root lockfile", map[string]string{"package.json": "{}", "bun.lock": "", "packages/a/package.json": "{}"}, "packages/a", "bun"},
		{"workspace inherits the root packageManager", map[string]string{"package.json": `{"packageManager": "yarn@4.0.0"}`, "packages/a/package.json": "{}"}, "packages/a", "yarn"},
		{"own lockfile beats an ancestor's lockfile", map[string]string{"package.json": "{}", "yarn.lock": "", "web/package.json": "{}", "web/bun.lock": ""}, "web", "bun"},
		{"ancestor's packageManager beats an own lockfile", map[string]string{"package.json": `{"packageManager": "pnpm@9.0.0"}`, "web/package.json": "{}", "web/bun.lock": ""}, "web", "pnpm"},
	}
	for _, tt := range tests {
		root, _ := writeProject(t, tt.files)
		if got := DetectPackageManager(filepath.Join(root, tt.dir), root); got != tt.want {
			t.Errorf("%s: DetectPackageManager = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestInstallAndScriptCommands(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"npm without a lockfile", map[string]string{"package.json": "{}"}, []string{"npm install"}},
		{"npm with a lockfile", map[string]string{"package.json": "{}", "package-lock.json": ""}, []string{"npm install", "npm ci"}},
		{"yarn classic", map[string]string{"package.json": "{}", "yarn.lock": ""}, []string{"yarn install", "yarn install --frozen-lockfile"}},
		{"yarn berry by .yarnrc.yml", map[string]string{"package.json": "{}", "yarn.lock": "", ".yarnrc.yml": ""}, []string{"yarn install", "yarn install --immutable"}},
		{"yarn berry by packageManager", map[string]string{"package.json": `{"packageManager": "yarn@4.1.0"}`, "yarn.lock": ""}, []string{"yarn install", "yarn install --immutable"}},
		{"pnpm", map[string]string{"package.json": "{}", "pnpm-lock.yaml": ""}, []string{"pnpm install", "pnpm install --frozen-lockfile"}},
		{"pnpm declared without a lockfile", map[string]string{"package.json": `{"packageManager": "pnpm@9.0.0"}`}, []string{"pnpm install"}},
		{"bun", map[string]string{"package.json": "{}", "bun.lock": ""}, []string{"bun install", "bun install --frozen-lockfile"}},
	}
	for _, tt := range tests {
		root, _ := writeProject(t, tt.files)
		var got []string
		for _, c := range detectNodePackageManagerCommands(root) {
			got = append(got, c.Name)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: installs %v, want %v", tt.name, got, tt.want)
		}
	}

	for pm, want := range map[string][]string{
		"npm":  {"npm start", "npm test", "npm run build"},
		"yarn": {"yarn start", "yarn test", "yarn run build"},
		"pnpm": {"pnpm start", "pnpm test", "pnpm run build"},
		"bun":  {"bun run start", "bun run test", "bun run build"},
	} {
		got := []string{scriptCommand(pm, "start"), scriptCommand(pm, "test"), scriptCommand(pm, "build")}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s scripts = %v, want %v", pm, got, want)
		}
	}
}

func TestPackageManagerIsSharedBySetupAndTechStack(t *testing.T) {
	root, files := writeProject(t, map[string]string{"package.json": `{"scripts": {"test": "vitest"}}`, "bun.lock": ""})
	if got := DetectCommands(root); got[0].Name != "bun run test" {
		t.Errorf("first command = %q, want bun run test", got[0].Name)
	}
	info := NewDevelopmentDetector(root, files).Detect()
	if info == nil || len(info.SetupSteps) == 0 || info.SetupSteps[0].Command != "bun install" {
		t.Errorf("setup steps = %+v, want bun install first", info.SetupSteps)
	}
	stack := detectStack(t, map[string]string{"package.json": "{}", "pnpm-lock.yaml": ""})
	if stack.PackageManager != "pnpm" {
		t.Errorf("tech stack package manager = %q, want pnpm", stack.PackageManager)
	}
	if stack := detectStack(t, map[string]string{"go.mod": "module m\n"}); stack.PackageManager != "" {
		t.Errorf("package manager without package.json = %q", stack.PackageManager)
	}
}
