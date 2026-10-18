package generator

import (
	"slices"
	"strings"
)

// Paths argus writes, relative to an output root: files by slash path,
// directories by name.
var (
	generatedFiles = []string{"CLAUDE.md", "AGENTS.md", ".cursorrules", ".github/copilot-instructions.md"}
	generatedDirs  = []string{".claude", ".continue"}
)

// IsGeneratedPath reports whether the relative path p is an argus output at
// any depth (workspace outputs included), so scans and watch can ignore it.
// Both slash and backslash separators are accepted.
func IsGeneratedPath(p string) bool {
	p = strings.TrimPrefix(strings.ReplaceAll(p, `\`, "/"), "./")
	for _, f := range generatedFiles {
		if p == f || strings.HasSuffix(p, "/"+f) {
			return true
		}
	}
	for _, part := range strings.Split(p, "/") {
		if slices.Contains(generatedDirs, part) {
			return true
		}
	}
	return false
}
