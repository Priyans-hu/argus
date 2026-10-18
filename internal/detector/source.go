package detector

import "strings"

// nonSourceDirs hold vendored, documentation, example, fixture and generated
// code, which is not the project's own source.
var nonSourceDirs = map[string]bool{
	"vendor": true, "third_party": true, "docs": true, "doc": true, "documentation": true,
	"examples": true, "example": true, "samples": true, "demo": true, "testdata": true,
	"fixtures": true, "__fixtures__": true, "__mocks__": true,
	"generated": true, "__generated__": true, ".generated": true,
}

// IsProjectSource reports whether rel, a path relative to the scan root, has
// no directory component in nonSourceDirs. The scan root's own name is never
// part of rel, so a checkout named docs or demo still counts.
func IsProjectSource(rel string) bool {
	parts := strings.Split(strings.ReplaceAll(rel, `\`, "/"), "/")
	for _, dir := range parts[:len(parts)-1] {
		if nonSourceDirs[strings.ToLower(dir)] {
			return false
		}
	}
	return true
}
