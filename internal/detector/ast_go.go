package detector

import (
	"cmp"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
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

// goLibrary is a library recognized by an import of its module or of a
// package below it, so echo/v4 and gin/binding match and gin-contrib does not.
type goLibrary struct {
	module, name, category, description string
}

// goLibraries are filed by the analyzers: go-web in Routing, go-database in
// Database & ORM, go-testing in Testing, the rest in Go Patterns.
var goLibraries = []goLibrary{
	{"github.com/spf13/cobra", "cobra", "go-cli", "Cobra CLI framework"},
	{"github.com/urfave/cli", "urfave/cli", "go-cli", "urfave/cli framework"},
	{"gopkg.in/alecthomas/kingpin", "kingpin", "go-cli", "Kingpin CLI framework"},
	{"gopkg.in/alecthomas/kingpin.v2", "kingpin", "go-cli", "Kingpin CLI framework"},
	{"github.com/alecthomas/kingpin", "kingpin", "go-cli", "Kingpin CLI framework"},
	{"github.com/spf13/pflag", "pflag", "go-cli", "POSIX/GNU flags"},
	{"github.com/spf13/viper", "viper", "go-config", "Viper configuration"},
	{"github.com/joho/godotenv", "godotenv", "go-config", "GoDotEnv .env loader"},
	{"github.com/kelseyhightower/envconfig", "envconfig", "go-config", "envconfig struct tags"},
	{"github.com/caarlos0/env", "env", "go-config", "env struct tags"},
	{"github.com/sirupsen/logrus", "logrus", "go-logging", "Logrus structured logging"},
	{"go.uber.org/zap", "zap", "go-logging", "Zap high-performance logging"},
	{"github.com/rs/zerolog", "zerolog", "go-logging", "Zerolog logging"},
	{"log/slog", "slog", "go-logging", "Structured logging (Go 1.21+)"},
	{"github.com/go-resty/resty", "resty", "go-rest", "Resty HTTP client"},
	{"resty.dev", "resty", "go-rest", "Resty HTTP client"},
	{"github.com/golang-jwt/jwt", "jwt", "go-auth", "JWT handling"},
	{"github.com/google/wire", "wire", "go-di", "Google Wire dependency injection"},
	{"go.uber.org/dig", "dig", "go-di", "Uber Dig dependency injection"},
	{"go.uber.org/fx", "fx", "go-di", "Uber Fx application framework"},
	{"github.com/stretchr/testify", "testify", "go-testing", "Testify assertions and mocks"},
	{"github.com/golang/mock", "gomock", "go-testing", "GoMock mocking"},
	{"go.uber.org/mock", "gomock", "go-testing", "GoMock mocking"},
	{"github.com/onsi/ginkgo", "ginkgo", "go-testing", "Ginkgo BDD testing"},
	{"github.com/onsi/gomega", "gomega", "go-testing", "Gomega matchers"},
	{"github.com/gin-gonic/gin", "gin", "go-web", "Gin web framework"},
	{"github.com/labstack/echo", "echo", "go-web", "Echo web framework"},
	{"github.com/gofiber/fiber", "fiber", "go-web", "Fiber web framework"},
	{"github.com/go-chi/chi", "chi", "go-web", "Chi router"},
	{"github.com/gorilla/mux", "mux", "go-web", "Gorilla Mux router"},
	{"gorm.io/gorm", "gorm", "go-database", "GORM ORM"},
	{"github.com/jmoiron/sqlx", "sqlx", "go-database", "sqlx database library"},
	{"github.com/jackc/pgx", "pgx", "go-database", "pgx PostgreSQL driver"},
	{"go.mongodb.org/mongo-driver", "mongo", "go-database", "MongoDB Go driver"},
	{"github.com/uptrace/bun", "bun", "go-database", "Bun ORM"},
	{"entgo.io/ent", "ent", "go-database", "Ent ORM"},
	{"github.com/volatiletech/sqlboiler", "sqlboiler", "go-database", "SQLBoiler ORM"},
}

// goStdlib are standard library imports reported by their base name.
var goStdlib = map[string]string{
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

// goCalls are the package-qualified calls reported from non-test files.
var goCalls = map[string]string{
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

// Detect parses every Go file once and reports libraries and standard
// packages from imports and known calls, sorted by name and category. Test
// files count only for go-testing libraries.
func (d *GoASTDetector) Detect() []types.PatternInfo {
	type key struct{ name, category string }
	found := map[key][]string{}
	descriptions := map[key]string{}
	add := func(k key, description, file string) {
		if !slices.Contains(found[k], file) {
			found[k] = append(found[k], file)
		}
		descriptions[k] = cmp.Or(descriptions[k], description)
	}

	for _, f := range d.files {
		if f.IsDir || !strings.HasSuffix(f.Name, ".go") {
			continue
		}
		isTest := IsTestFile(f.Path)
		mode := parser.SkipObjectResolution
		if isTest {
			mode |= parser.ImportsOnly
		}
		node, err := parser.ParseFile(token.NewFileSet(), filepath.Join(d.rootPath, f.Path), nil, mode)
		if err != nil {
			continue
		}

		for _, imp := range node.Imports {
			path := strings.Trim(imp.Path.Value, `"`)
			for _, lib := range goLibraries {
				if (path == lib.module || strings.HasPrefix(path, lib.module+"/")) && (!isTest || lib.category == "go-testing") {
					add(key{lib.name, lib.category}, lib.description, f.Path)
				}
			}
			if desc, ok := goStdlib[path]; ok && !isTest {
				add(key{filepath.Base(path), "go-stdlib"}, desc, f.Path)
			}
		}
		if isTest {
			continue
		}
		ast.Inspect(node, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if x, ok := sel.X.(*ast.Ident); ok {
						name := x.Name + "." + sel.Sel.Name
						if desc, ok := goCalls[name]; ok {
							add(key{name, "go-patterns"}, desc, f.Path)
						}
					}
				}
			}
			return true
		})
	}

	var patterns []types.PatternInfo
	for k, files := range found {
		patterns = append(patterns, types.PatternInfo{
			Name: k.name, Category: k.category, Description: descriptions[k],
			FileCount: len(files), Examples: limitSlice(files, 3),
		})
	}
	slices.SortFunc(patterns, func(a, b types.PatternInfo) int {
		return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Category, b.Category))
	})
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
