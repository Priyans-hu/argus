package detector

import (
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// GradleCommands returns the commands of a Gradle build at root (a gradlew
// wrapper or a build or settings script there): the debug APK for Android
// projects (an AndroidManifest.xml in project source), else the build, and
// the unit tests when a src/test directory exists.
func GradleCommands(root string, files []types.FileInfo) []types.Command {
	gradle := "gradle"
	if fileExists(filepath.Join(root, "gradlew")) {
		gradle = "./gradlew"
	} else if !hasAnyFile(root, "build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts") {
		return nil
	}
	android, tests := false, false
	for _, f := range files {
		rel := filepath.ToSlash(f.Path)
		if f.IsDir || !IsProjectSource(rel) {
			continue
		}
		android = android || f.Name == "AndroidManifest.xml"
		tests = tests || strings.HasPrefix(rel, "src/test/") || strings.Contains(rel, "/src/test/")
	}
	commands := []types.Command{{Name: gradle + " build", Description: "Build the project"}}
	if android {
		commands = []types.Command{{Name: gradle + " assembleDebug", Description: "Build the debug APK"}}
	}
	if tests {
		commands = append(commands, types.Command{Name: gradle + " test", Description: "Run unit tests"})
	}
	return commands
}

// SubdirCargoCommands returns the Cargo commands of each crate one directory
// below root, each suffixed with (in <dir>).
func SubdirCargoCommands(root string, files []types.FileInfo) []types.Command {
	var commands []types.Command
	for _, m := range FindManifests(files) {
		if m.Kind != "Cargo.toml" || strings.Count(m.Path, "/") != 1 {
			continue
		}
		for _, c := range NewCargoDetector(filepath.Join(root, filepath.FromSlash(m.Dir))).DetectCargoCommands() {
			c.Name += " (in " + m.Dir + ")"
			commands = append(commands, c)
		}
	}
	return commands
}

func hasAnyFile(dir string, names ...string) bool {
	for _, name := range names {
		if fileExists(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}
