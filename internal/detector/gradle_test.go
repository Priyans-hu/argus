package detector

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func commandNames(commands []types.Command) []string {
	var names []string
	for _, c := range commands {
		names = append(names, c.Name)
	}
	return names
}

func TestGradleCommands(t *testing.T) {
	spindle := map[string]string{
		"gradlew": "", "settings.gradle.kts": "", "app/build.gradle.kts": "",
		"app/src/main/AndroidManifest.xml": "<manifest/>", "app/src/main/java/A.kt": "",
	}
	withTests := map[string]string{"app/src/test/java/ATest.kt": ""}
	for k, v := range spindle {
		withTests[k] = v
	}
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"Android app without tests", spindle, []string{"./gradlew assembleDebug"}},
		{"Android app with unit tests", withTests, []string{"./gradlew assembleDebug", "./gradlew test"}},
		{"JVM project without a wrapper", map[string]string{"build.gradle.kts": "", "src/main/java/A.java": "", "src/test/java/ATest.java": ""},
			[]string{"gradle build", "gradle test"}},
		{"no Gradle build", map[string]string{"app/build.gradle.kts": ""}, nil},
	}
	for _, tt := range tests {
		root, files := writeProject(t, tt.files)
		if got := commandNames(GradleCommands(root, files)); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s: %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSubdirCargoCommands(t *testing.T) {
	root, files := writeProject(t, map[string]string{
		"package.json":         "{}",
		"src-tauri/Cargo.toml": "[package]\nname = \"app\"\n\n[dependencies]\ntauri = \"2\"\n",
		"a/b/Cargo.toml":       "[package]\nname = \"deep\"\n",
	})
	got := commandNames(SubdirCargoCommands(root, files))
	for _, want := range []string{"cargo build (in src-tauri)", "cargo test (in src-tauri)"} {
		if !slices.Contains(got, want) {
			t.Errorf("commands %v lack %q", got, want)
		}
	}
	for _, name := range got {
		if !strings.HasSuffix(name, " (in src-tauri)") {
			t.Errorf("command %q is not from src-tauri (a/b is two directories deep)", name)
		}
	}
}
