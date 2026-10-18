package detector

import (
	"cmp"
	"maps"
	"slices"

	"github.com/Priyans-hu/argus/pkg/types"
)

// longestFirst returns the keys of m longest first, alphabetical among equal
// lengths, so the most specific prefix wins ("next/router" before "next").
func longestFirst[V any](m map[string]V) []string {
	keys := slices.Collect(maps.Keys(m))
	slices.SortFunc(keys, func(a, b string) int {
		return cmp.Or(cmp.Compare(len(b), len(a)), cmp.Compare(a, b))
	})
	return keys
}

// ASTDetector is the interface for language-specific AST detectors
type ASTDetector interface {
	// Detect analyzes source files and returns detected patterns
	Detect() []types.PatternInfo
}

// ImportInfo represents an import statement
type ImportInfo struct {
	Name   string // import name/path
	Alias  string // import alias (if any)
	IsType bool   // true for TypeScript type-only imports
}

// FunctionInfo represents a function definition
type FunctionInfo struct {
	Name       string
	IsAsync    bool
	IsExported bool
	HasTypes   bool // has type annotations
	Line       int
}

// ClassInfo represents a class definition
type ClassInfo struct {
	Name       string
	BaseClass  string
	IsExported bool
	Line       int
}

// DecoratorInfo represents a decorator/annotation
type DecoratorInfo struct {
	Name      string
	Arguments []string
	Line      int
}
