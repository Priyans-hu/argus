package detector

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
	"golang.org/x/mod/modfile"
)

// GoASTDetector uses AST parsing for accurate Go pattern detection
type GoASTDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewGoASTDetector creates a new AST-based Go detector
func NewGoASTDetector(rootPath string, files []types.FileInfo) *GoASTDetector {
	return &GoASTDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes Go code using AST and returns patterns
func (d *GoASTDetector) Detect() []types.PatternInfo {
	var patterns []types.PatternInfo

	// Track imports across all Go files
	imports := make(map[string][]string)          // import path -> files using it
	functionPatterns := make(map[string][]string) // pattern -> files
	structPatterns := make(map[string][]string)   // pattern -> files

	// Parse all Go files
	for _, f := range d.files {
		if f.IsDir || !strings.HasSuffix(f.Name, ".go") {
			continue
		}

		// Skip test files for some patterns
		isTest := strings.HasSuffix(f.Name, "_test.go")

		fullPath := filepath.Join(d.rootPath, f.Path)
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, fullPath, nil, parser.ImportsOnly)
		if err != nil {
			continue
		}

		// Extract imports
		for _, imp := range node.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			imports[importPath] = append(imports[importPath], f.Path)
		}

		// Parse again with full detail for non-test files
		if !isTest {
			node, err = parser.ParseFile(fset, fullPath, nil, parser.ParseComments)
			if err != nil {
				continue
			}

			// Walk the AST
			ast.Inspect(node, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.CallExpr:
					// Detect function calls
					if sel, ok := x.Fun.(*ast.SelectorExpr); ok {
						pattern := d.extractCallPattern(sel)
						if pattern != "" {
							functionPatterns[pattern] = append(functionPatterns[pattern], f.Path)
						}
					}
				case *ast.StructType:
					// Detect struct embedding
					for _, field := range x.Fields.List {
						if len(field.Names) == 0 { // Embedded field
							if ident, ok := field.Type.(*ast.Ident); ok {
								pattern := "embedded:" + ident.Name
								structPatterns[pattern] = append(structPatterns[pattern], f.Path)
							}
						}
					}
				}
				return true
			})
		}
	}

	// Convert imports to patterns
	patterns = append(patterns, d.importsToPatterns(imports)...)

	// Add function call patterns
	patterns = append(patterns, d.callsToPatterns(functionPatterns)...)

	// Detect CLI frameworks
	patterns = append(patterns, d.detectCLIFrameworks(imports)...)

	// Detect logging libraries
	patterns = append(patterns, d.detectLogging(imports)...)

	// Detect testing frameworks
	patterns = append(patterns, d.detectTesting(imports)...)

	// Detect web frameworks
	patterns = append(patterns, d.detectWebFrameworks(imports)...)

	// Detect database libraries
	patterns = append(patterns, d.detectDatabase(imports)...)

	// Detect configuration libraries
	patterns = append(patterns, d.detectConfig(imports)...)

	return patterns
}

// extractCallPattern extracts a pattern from a selector expression
func (d *GoASTDetector) extractCallPattern(sel *ast.SelectorExpr) string {
	if x, ok := sel.X.(*ast.Ident); ok {
		return x.Name + "." + sel.Sel.Name
	}
	return ""
}

// importsToPatt converts import usage to patterns
func (d *GoASTDetector) importsToPatterns(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	// Standard library patterns
	stdPatterns := map[string]string{
		"context":       "Context for cancellation and timeouts",
		"sync":          "Synchronization primitives",
		"encoding/json": "JSON encoding/decoding",
		"encoding/xml":  "XML encoding/decoding",
		"net/http":      "HTTP client/server",
		"database/sql":  "SQL database access",
		"html/template": "HTML templating",
		"text/template": "Text templating",
		"crypto/tls":    "TLS cryptography",
		"crypto/sha256": "SHA256 hashing",
	}

	for importPath, files := range imports {
		if desc, ok := stdPatterns[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        filepath.Base(importPath),
				Category:    "go-stdlib",
				Description: desc,
				FileCount:   len(dedupe(files)),
				Examples:    limitSlice(dedupe(files), 3),
			})
		}
	}

	return patterns
}

// callsToPatterns converts function call patterns to PatternInfo
func (d *GoASTDetector) callsToPatterns(calls map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	// Known patterns
	knownPatterns := map[string]string{
		"fmt.Errorf":           "Error formatting",
		"errors.New":           "Error creation",
		"errors.Is":            "Error comparison (Go 1.13+)",
		"errors.As":            "Error type assertion (Go 1.13+)",
		"json.Marshal":         "JSON serialization",
		"json.Unmarshal":       "JSON deserialization",
		"context.WithCancel":   "Cancelable context",
		"context.WithTimeout":  "Context with timeout",
		"context.WithDeadline": "Context with deadline",
	}

	for pattern, files := range calls {
		if desc, ok := knownPatterns[pattern]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        pattern,
				Category:    "go-patterns",
				Description: desc,
				FileCount:   len(dedupe(files)),
				Examples:    limitSlice(dedupe(files), 3),
			})
		}
	}

	return patterns
}

// detectCLIFrameworks detects CLI frameworks from imports
func (d *GoASTDetector) detectCLIFrameworks(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	frameworks := map[string]string{
		"github.com/spf13/cobra":           "Cobra CLI framework",
		"github.com/urfave/cli":            "urfave/cli framework",
		"github.com/urfave/cli/v2":         "urfave/cli v2 framework",
		"gopkg.in/alecthomas/kingpin":      "Kingpin CLI framework",
		"github.com/alecthomas/kingpin/v2": "Kingpin v2 CLI framework",
		"github.com/spf13/pflag":           "POSIX/GNU flags",
	}

	for importPath, files := range imports {
		if desc, ok := frameworks[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        filepath.Base(importPath),
				Category:    "go-cli",
				Description: desc,
				FileCount:   len(dedupe(files)),
				Examples:    limitSlice(dedupe(files), 3),
			})
		}
	}

	return patterns
}

// detectLogging detects logging libraries
func (d *GoASTDetector) detectLogging(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	loggers := map[string]string{
		"github.com/sirupsen/logrus": "Logrus structured logging",
		"go.uber.org/zap":            "Zap high-performance logging",
		"github.com/rs/zerolog":      "Zerolog logging",
		"log/slog":                   "Structured logging (Go 1.21+)",
	}

	for importPath, files := range imports {
		if desc, ok := loggers[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        filepath.Base(importPath),
				Category:    "go-logging",
				Description: desc,
				FileCount:   len(dedupe(files)),
				Examples:    limitSlice(dedupe(files), 3),
			})
		}
	}

	return patterns
}

// detectTesting detects testing frameworks
func (d *GoASTDetector) detectTesting(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	testFrameworks := map[string]string{
		"github.com/stretchr/testify/assert":  "Testify assertions",
		"github.com/stretchr/testify/require": "Testify require",
		"github.com/stretchr/testify/mock":    "Testify mocking",
		"github.com/golang/mock/gomock":       "GoMock mocking",
		"github.com/onsi/ginkgo":              "Ginkgo BDD testing",
		"github.com/onsi/gomega":              "Gomega matchers",
	}

	for importPath, files := range imports {
		if desc, ok := testFrameworks[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        filepath.Base(importPath),
				Category:    "go-testing",
				Description: desc,
				FileCount:   len(dedupe(files)),
				Examples:    limitSlice(dedupe(files), 3),
			})
		}
	}

	return patterns
}

// detectWebFrameworks detects web frameworks
func (d *GoASTDetector) detectWebFrameworks(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	frameworks := map[string]string{
		"github.com/gin-gonic/gin":    "Gin web framework",
		"github.com/labstack/echo":    "Echo web framework",
		"github.com/labstack/echo/v4": "Echo v4 web framework",
		"github.com/gofiber/fiber":    "Fiber web framework",
		"github.com/gofiber/fiber/v2": "Fiber v2 web framework",
		"github.com/gorilla/mux":      "Gorilla Mux router",
		"github.com/go-chi/chi":       "Chi router",
		"github.com/go-chi/chi/v5":    "Chi v5 router",
	}

	for importPath, files := range imports {
		if desc, ok := frameworks[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        filepath.Base(importPath),
				Category:    "go-web",
				Description: desc,
				FileCount:   len(dedupe(files)),
				Examples:    limitSlice(dedupe(files), 3),
			})
		}
	}

	return patterns
}

// detectDatabase detects database libraries
func (d *GoASTDetector) detectDatabase(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	databases := map[string]string{
		"gorm.io/gorm":                      "GORM ORM",
		"github.com/jmoiron/sqlx":           "sqlx database library",
		"github.com/jackc/pgx/v5":           "pgx PostgreSQL driver",
		"github.com/jackc/pgx/v4":           "pgx v4 PostgreSQL driver",
		"go.mongodb.org/mongo-driver/mongo": "MongoDB Go driver",
		"github.com/uptrace/bun":            "Bun ORM",
		"entgo.io/ent":                      "Ent ORM",
		"github.com/volatiletech/sqlboiler": "SQLBoiler ORM",
	}

	for importPath, files := range imports {
		if desc, ok := databases[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        filepath.Base(importPath),
				Category:    "go-database",
				Description: desc,
				FileCount:   len(dedupe(files)),
				Examples:    limitSlice(dedupe(files), 3),
			})
		}
	}

	return patterns
}

// detectConfig detects configuration libraries
func (d *GoASTDetector) detectConfig(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	configs := map[string]string{
		"github.com/spf13/viper":               "Viper configuration",
		"github.com/joho/godotenv":             "GoDotEnv .env loader",
		"github.com/kelseyhightower/envconfig": "envconfig struct tags",
		"github.com/caarlos0/env":              "env struct tags",
	}

	for importPath, files := range imports {
		if desc, ok := configs[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        filepath.Base(importPath),
				Category:    "go-config",
				Description: desc,
				FileCount:   len(dedupe(files)),
				Examples:    limitSlice(dedupe(files), 3),
			})
		}
	}

	return patterns
}

// ParseGoMod parses go.mod file and returns module information
func ParseGoMod(rootPath string) (*modfile.File, error) {
	goModPath := filepath.Join(rootPath, "go.mod")
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}

	return modfile.Parse(goModPath, content, nil)
}

// dedupe removes duplicate strings from a slice
func dedupe(strs []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range strs {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
