package detector

import "github.com/Priyans-hu/argus/pkg/types"

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
