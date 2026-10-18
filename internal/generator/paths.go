package generator

import (
	"slices"
	"strings"
)

// Paths argus writes, relative to an output root: files by slash path,
// directories by name. The files carry ARGUS:AUTO markers and are merged on
// rescans; everything under the directories is create-only.
var (
	generatedFiles = []string{"CLAUDE.md", "AGENTS.md", ".cursorrules", ".github/copilot-instructions.md"}
	generatedDirs  = []string{".claude", ".continue"}
)

// TempPrefix starts the temp file a write goes through before its rename;
// os.CreateTemp completes the name with digits. A crash can leave one behind.
const TempPrefix = ".argus-"

// IsGeneratedPath reports whether the relative path p is an argus output at
// any depth (workspace outputs included), so scans and watch can ignore it.
// Both slash and backslash separators are accepted.
func IsGeneratedPath(p string) bool {
	p = strings.TrimPrefix(strings.ReplaceAll(p, `\`, "/"), "./")
	base := p[strings.LastIndex(p, "/")+1:]
	if digits, ok := strings.CutPrefix(base, TempPrefix); ok && digits != "" && strings.Trim(digits, "0123456789") == "" {
		return true
	}
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

// IsMergeFile reports whether rel, a slash path relative to an output root,
// is an output that argus merges into rather than creating only once.
func IsMergeFile(rel string) bool {
	return slices.Contains(generatedFiles, rel)
}
