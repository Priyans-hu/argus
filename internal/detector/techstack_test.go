package detector

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func detectStack(t *testing.T, files map[string]string) *types.TechStack {
	t.Helper()
	root, infos := writeProject(t, files)
	return stableResult(t, func() *types.TechStack {
		stack, err := NewTechStackDetector(root, infos, ReadDependencies(root, infos)).Detect()
		if err != nil {
			t.Fatal(err)
		}
		return stack
	})
}

func TestTechStackLanguagesTieBreakByName(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"a.ts": "", "b.ts": "", "main.go": "", "util.go": "", "tool.py": "",
	})

	want := []types.Language{{Name: "Go", Percentage: 40}, {Name: "TypeScript", Percentage: 40}, {Name: "Python", Percentage: 20}}
	if !reflect.DeepEqual(stack.Languages, want) {
		t.Errorf("Languages = %+v, want %+v", stack.Languages, want)
	}
}

func TestTechStackPackageJSONIsStable(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"package.json": `{
			"dependencies": {"next": "14.0.0", "react": "^18.2.0", "zustand": "4.0.0", "prisma": "^5.0.0", "@prisma/client": "^5.1.0", "pg": "8.0.0", "redis": "4.0.0"},
			"devDependencies": {"typescript": "5.3.0", "vitest": "1.0.0", "jest": "29.0.0", "eslint": "8.0.0", "prettier": "3.0.0"}
		}`,
	})

	var prisma []types.Framework
	for _, fw := range stack.Frameworks {
		if fw.Name == "Prisma" {
			prisma = append(prisma, fw)
		}
	}
	if len(prisma) != 1 || prisma[0].Version != "5.1.0" {
		t.Errorf("Prisma = %+v, want one entry with the @prisma/client version 5.1.0", prisma)
	}
	if want := []string{"PostgreSQL", "Redis"}; !reflect.DeepEqual(stack.Databases, want) {
		t.Errorf("Databases = %v, want %v", stack.Databases, want)
	}
}

func TestTechStackGoAndRustFrameworksAreStable(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"go.mod":     "module x\n\ngo 1.24\n\nrequire (\n\tgithub.com/spf13/cobra v1.0.0\n\tgithub.com/gin-gonic/gin v1.0.0\n\tgorm.io/gorm v1.0.0\n)\n",
		"Cargo.toml": "[dependencies]\ntokio = \"1\"\nserde = \"1\"\naxum = \"0.7\"\nclap = \"4\"\n",
	})
	if len(stack.Frameworks) != 7 {
		t.Errorf("Frameworks = %+v, want 7", stack.Frameworks)
	}
}

func TestTechStackPythonFrameworksAreDeduplicated(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"requirements.txt": "fastapi==0.110\npydantic\n",
		"pyproject.toml":   "[project]\ndependencies = [\"fastapi\", \"pydantic\", \"pytest\"]\n",
	})

	var names []string
	for _, fw := range stack.Frameworks {
		names = append(names, fw.Name)
	}
	if want := []string{"FastAPI", "Pydantic", "pytest"}; !reflect.DeepEqual(names, want) {
		t.Errorf("Frameworks = %v, want %v", names, want)
	}
}

// frameworkNames returns the stack's frameworks as name@version.
func frameworkNames(stack *types.TechStack) []string {
	var names []string
	for _, fw := range stack.Frameworks {
		names = append(names, fw.Name+"@"+fw.Version)
	}
	return names
}

func TestTechStackPeerOnlyFrameworkHasNoVersion(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"package.json": `{"peerDependencies": {"react": ">=18"}, "devDependencies": {"vitest": "^2.0.0"}}`,
	})
	if got, want := frameworkNames(stack), []string{"React@", "Vitest@2.0.0"}; !reflect.DeepEqual(got, want) {
		t.Errorf("frameworks = %v, want %v", got, want)
	}
}

func TestTechStackReadsSubdirectoryManifests(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"web/package.json":     `{"dependencies": {"next": "14.0.0"}}`,
		"src-tauri/Cargo.toml": "[dependencies]\ntokio = \"1\"\n",
		"go.mod":               "module m\n\ngo 1.24\n\nrequire github.com/labstack/echo/v4 v4.11.0\n",
	})
	if got, want := frameworkNames(stack), []string{"Next.js@14.0.0", "Echo@", "Tokio@"}; !reflect.DeepEqual(got, want) {
		t.Errorf("frameworks = %v, want %v", got, want)
	}
}

func TestTechStackComparesExactNames(t *testing.T) {
	stack := detectStack(t, map[string]string{
		"requirements.txt": "pytest-cov\ndjango-extensions\n",
		"go.mod": "module m\n\ngo 1.24\n\n// github.com/gin-gonic/gin is not used\nrequire (\n" +
			"\tgithub.com/gin-gonic/gin v1.9.0 // indirect\n\tgithub.com/spf13/cobra-cli v1.0.0\n)\n",
	})
	if len(stack.Frameworks) != 0 {
		t.Errorf("frameworks = %v, want none", frameworkNames(stack))
	}
}

func TestTechStackDartAndZigLanguages(t *testing.T) {
	files := map[string]string{}
	for i := range 140 {
		files[fmt.Sprintf("lib/f%d.dart", i)] = "void main() {}\n"
	}
	for i := range 80 {
		files[fmt.Sprintf("android/f%d.kt", i)] = "fun main() {}\n"
	}
	for i := range 10 {
		files[fmt.Sprintf("android/f%d.java", i)] = "class A {}\n"
	}
	if got := detectStack(t, files).Languages; len(got) == 0 || got[0].Name != "Dart" {
		t.Errorf("languages = %+v, want Dart first", got)
	}

	files = map[string]string{"a.swift": "", "b.zig": "", "c.zig": ""}
	if got := detectStack(t, files).Languages; len(got) == 0 || got[0].Name != "Zig" {
		t.Errorf("languages = %+v, want Zig first", got)
	}
}

func TestTechStackAppFrameworks(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string // "name category"
	}{
		{"Flutter pubspec", map[string]string{"app/pubspec.yaml": "name: app\ndependencies:\n  flutter:\n    sdk: flutter\n"}, []string{"Flutter frontend"}},
		{"Dart-only pubspec", map[string]string{"pubspec.yaml": "name: lib\ndependencies:\n  path: ^1.8.0\n"}, nil},
		{"@tauri-apps/cli", map[string]string{"package.json": `{"devDependencies": {"@tauri-apps/cli": "2.0.0"}}`}, []string{"Tauri fullstack"}},
		{"@tauri-apps/api", map[string]string{"package.json": `{"dependencies": {"@tauri-apps/api": "2.0.0"}}`}, []string{"Tauri fullstack"}},
		{"tauri crate", map[string]string{"src-tauri/Cargo.toml": "[package]\nname = \"app\"\n\n[dependencies]\ntauri = \"2\"\n"}, []string{"Tauri fullstack"}},
		{"tauri.conf.json", map[string]string{"src-tauri/tauri.conf.json": "{}"}, []string{"Tauri fullstack"}},
		{"SvelteKit", map[string]string{"package.json": `{"devDependencies": {"@sveltejs/kit": "2.0.0"}}`}, []string{"SvelteKit frontend"}},
		{"claude-glance shape", map[string]string{
			"package.json":         `{"devDependencies": {"@sveltejs/kit": "2.0.0", "@tauri-apps/cli": "2.0.0"}}`,
			"src-tauri/Cargo.toml": "[package]\nname = \"app\"\n\n[dependencies]\ntauri = \"2\"\n",
		}, []string{"SvelteKit frontend", "Tauri fullstack"}},
	}
	for _, tt := range tests {
		var got []string
		for _, fw := range detectStack(t, tt.files).Frameworks {
			got = append(got, fw.Name+" "+fw.Category)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: frameworks %v, want %v", tt.name, got, tt.want)
		}
	}
}
