package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/go-python/gpython/ast"
	"github.com/go-python/gpython/parser"
	"github.com/go-python/gpython/py"
)

// PythonASTDetector uses AST parsing for accurate Python pattern detection
type PythonASTDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewPythonASTDetector creates a new AST-based Python detector
func NewPythonASTDetector(rootPath string, files []types.FileInfo) *PythonASTDetector {
	return &PythonASTDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes Python code using AST and returns patterns
func (d *PythonASTDetector) Detect() []types.PatternInfo {
	// Track patterns across all Python files
	imports := make(map[string][]string)       // import name -> files using it
	decorators := make(map[string][]string)    // decorator name -> files using it
	classPatterns := make(map[string][]string) // class pattern -> files
	funcPatterns := make(map[string][]string)  // function pattern -> files

	// Parse all Python files
	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		// Check for Python extension
		ext := strings.ToLower(f.Extension)
		if ext != ".py" {
			continue
		}

		// Skip common non-source directories
		if strings.Contains(f.Path, "__pycache__") ||
			strings.Contains(f.Path, ".venv") ||
			strings.Contains(f.Path, "venv/") ||
			strings.Contains(f.Path, "site-packages") {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Parse the file
		mod, err := parser.Parse(strings.NewReader(string(content)), fullPath, py.ExecMode)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		// Extract patterns from AST
		module, ok := mod.(*ast.Module)
		if !ok {
			continue
		}

		d.extractFromModule(module, f.Path, imports, decorators, classPatterns, funcPatterns)
	}

	// Convert to PatternInfo
	var patterns []types.PatternInfo
	patterns = append(patterns, d.importsToPatterns(imports)...)
	patterns = append(patterns, d.decoratorsToPatterns(decorators)...)
	patterns = append(patterns, d.classToPatterns(classPatterns)...)
	patterns = append(patterns, d.funcToPatterns(funcPatterns)...)

	return patterns
}

// extractFromModule walks the module AST and extracts patterns
func (d *PythonASTDetector) extractFromModule(module *ast.Module, filePath string,
	imports, decorators, classPatterns, funcPatterns map[string][]string) {

	for _, stmt := range module.Body {
		d.visitStmt(stmt, filePath, imports, decorators, classPatterns, funcPatterns)
	}
}

// visitStmt visits a statement and extracts patterns
func (d *PythonASTDetector) visitStmt(stmt ast.Stmt, filePath string,
	imports, decorators, classPatterns, funcPatterns map[string][]string) {

	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.Import:
		// import x, y, z
		for _, alias := range s.Names {
			name := string(alias.Name)
			imports[name] = pyAppendUnique(imports[name], filePath)
		}

	case *ast.ImportFrom:
		// from x import y
		if s.Module != "" {
			moduleName := string(s.Module)
			imports[moduleName] = pyAppendUnique(imports[moduleName], filePath)
			// Also track specific imports
			for _, alias := range s.Names {
				fullName := moduleName + "." + string(alias.Name)
				imports[fullName] = pyAppendUnique(imports[fullName], filePath)
			}
		}

	case *ast.FunctionDef:
		// Check for async
		// Note: gpython may not have AsyncFunctionDef, check decorators instead

		// Extract decorators
		for _, dec := range s.DecoratorList {
			decName := d.extractDecoratorName(dec)
			if decName != "" {
				decorators[decName] = pyAppendUnique(decorators[decName], filePath)
			}
		}

		// Check for type hints
		if s.Returns != nil {
			funcPatterns["type hints"] = pyAppendUnique(funcPatterns["type hints"], filePath)
		}
		if s.Args != nil && len(s.Args.Args) > 0 {
			for _, arg := range s.Args.Args {
				if arg.Annotation != nil {
					funcPatterns["type hints"] = pyAppendUnique(funcPatterns["type hints"], filePath)
					break
				}
			}
		}

		// Visit function body
		for _, bodyStmt := range s.Body {
			d.visitStmt(bodyStmt, filePath, imports, decorators, classPatterns, funcPatterns)
		}

	case *ast.ClassDef:
		classPatterns["class definition"] = pyAppendUnique(classPatterns["class definition"], filePath)

		// Check for inheritance
		if len(s.Bases) > 0 {
			classPatterns["class inheritance"] = pyAppendUnique(classPatterns["class inheritance"], filePath)

			// Check for specific base classes
			for _, base := range s.Bases {
				baseName := d.extractExprName(base)
				if baseName != "" {
					classPatterns["extends:"+baseName] = pyAppendUnique(classPatterns["extends:"+baseName], filePath)
				}
			}
		}

		// Extract class decorators
		for _, dec := range s.DecoratorList {
			decName := d.extractDecoratorName(dec)
			if decName != "" {
				decorators[decName] = pyAppendUnique(decorators[decName], filePath)
			}
		}

		// Visit class body
		for _, bodyStmt := range s.Body {
			d.visitStmt(bodyStmt, filePath, imports, decorators, classPatterns, funcPatterns)
		}

	case *ast.If:
		for _, bodyStmt := range s.Body {
			d.visitStmt(bodyStmt, filePath, imports, decorators, classPatterns, funcPatterns)
		}
		for _, elseStmt := range s.Orelse {
			d.visitStmt(elseStmt, filePath, imports, decorators, classPatterns, funcPatterns)
		}

	case *ast.For:
		for _, bodyStmt := range s.Body {
			d.visitStmt(bodyStmt, filePath, imports, decorators, classPatterns, funcPatterns)
		}

	case *ast.While:
		for _, bodyStmt := range s.Body {
			d.visitStmt(bodyStmt, filePath, imports, decorators, classPatterns, funcPatterns)
		}

	case *ast.Try:
		for _, bodyStmt := range s.Body {
			d.visitStmt(bodyStmt, filePath, imports, decorators, classPatterns, funcPatterns)
		}
		for _, finallyStmt := range s.Finalbody {
			d.visitStmt(finallyStmt, filePath, imports, decorators, classPatterns, funcPatterns)
		}

	case *ast.With:
		for _, bodyStmt := range s.Body {
			d.visitStmt(bodyStmt, filePath, imports, decorators, classPatterns, funcPatterns)
		}
	}
}

// extractDecoratorName extracts the name of a decorator expression
func (d *PythonASTDetector) extractDecoratorName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Name:
		return string(e.Id)
	case *ast.Attribute:
		// e.g., app.route, pytest.fixture
		valueName := d.extractExprName(e.Value)
		if valueName != "" {
			return valueName + "." + string(e.Attr)
		}
		return string(e.Attr)
	case *ast.Call:
		// e.g., @app.route("/path")
		return d.extractDecoratorName(e.Func)
	}
	return ""
}

// extractExprName extracts a simple name from an expression
func (d *PythonASTDetector) extractExprName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Name:
		return string(e.Id)
	case *ast.Attribute:
		valueName := d.extractExprName(e.Value)
		if valueName != "" {
			return valueName + "." + string(e.Attr)
		}
		return string(e.Attr)
	}
	return ""
}

// importsToPatterns converts import map to PatternInfo slice
func (d *PythonASTDetector) importsToPatterns(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	// Framework detection based on imports
	frameworkImports := map[string]string{
		"flask":        "Flask",
		"django":       "Django",
		"fastapi":      "FastAPI",
		"starlette":    "Starlette",
		"tornado":      "Tornado",
		"bottle":       "Bottle",
		"pyramid":      "Pyramid",
		"aiohttp":      "aiohttp",
		"sanic":        "Sanic",
		"requests":     "Requests",
		"httpx":        "httpx",
		"sqlalchemy":   "SQLAlchemy",
		"peewee":       "Peewee",
		"tortoise":     "Tortoise ORM",
		"mongoengine":  "MongoEngine",
		"pymongo":      "PyMongo",
		"redis":        "Redis",
		"celery":       "Celery",
		"dramatiq":     "Dramatiq",
		"rq":           "RQ",
		"pytest":       "pytest",
		"unittest":     "unittest",
		"nose":         "nose",
		"pandas":       "pandas",
		"numpy":        "NumPy",
		"scipy":        "SciPy",
		"matplotlib":   "Matplotlib",
		"seaborn":      "Seaborn",
		"plotly":       "Plotly",
		"tensorflow":   "TensorFlow",
		"torch":        "PyTorch",
		"keras":        "Keras",
		"sklearn":      "scikit-learn",
		"transformers": "Transformers",
		"langchain":    "LangChain",
		"openai":       "OpenAI",
		"anthropic":    "Anthropic",
		"pydantic":     "Pydantic",
		"attrs":        "attrs",
		"dataclasses":  "dataclasses",
		"typing":       "typing",
		"asyncio":      "asyncio",
		"click":        "Click",
		"typer":        "Typer",
		"argparse":     "argparse",
		"fire":         "Fire",
		"rich":         "Rich",
		"loguru":       "Loguru",
		"structlog":    "structlog",
		"alembic":      "Alembic",
		"boto3":        "Boto3 (AWS)",
		"google.cloud": "Google Cloud",
		"azure":        "Azure",
	}

	for importPath, files := range imports {
		// Check direct match or prefix match
		for prefix, framework := range frameworkImports {
			if importPath == prefix || strings.HasPrefix(importPath, prefix+".") {
				patterns = append(patterns, types.PatternInfo{
					Category:    "Python Frameworks",
					Name:        framework,
					Description: "Detected via import: " + importPath,
					FileCount:   len(files),
					Examples:    limitSlice(dedupe(files), 3),
				})
				break
			}
		}
	}

	return patterns
}

// decoratorsToPatterns converts decorator map to PatternInfo slice
func (d *PythonASTDetector) decoratorsToPatterns(decorators map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	decoratorDescriptions := map[string]string{
		"app.route":           "Flask route decorator",
		"router.get":          "FastAPI GET endpoint",
		"router.post":         "FastAPI POST endpoint",
		"router.put":          "FastAPI PUT endpoint",
		"router.delete":       "FastAPI DELETE endpoint",
		"api_view":            "Django REST framework view",
		"pytest.fixture":      "pytest fixture",
		"pytest.mark":         "pytest marker",
		"property":            "Python property decorator",
		"staticmethod":        "Static method",
		"classmethod":         "Class method",
		"abstractmethod":      "Abstract method",
		"dataclass":           "Dataclass decorator",
		"validator":           "Pydantic validator",
		"field_validator":     "Pydantic field validator",
		"cached_property":     "Cached property",
		"lru_cache":           "LRU cache decorator",
		"functools.lru_cache": "LRU cache decorator",
		"contextmanager":      "Context manager",
		"asynccontextmanager": "Async context manager",
	}

	for decorator, files := range decorators {
		desc := decoratorDescriptions[decorator]
		if desc == "" {
			// Check for partial matches
			for key, value := range decoratorDescriptions {
				if strings.Contains(decorator, key) {
					desc = value
					break
				}
			}
			if desc == "" {
				desc = "Python decorator"
			}
		}
		patterns = append(patterns, types.PatternInfo{
			Category:    "Python Decorators",
			Name:        "@" + decorator,
			Description: desc,
			FileCount:   len(files),
			Examples:    limitSlice(dedupe(files), 3),
		})
	}

	return patterns
}

// classToPatterns converts class patterns to PatternInfo slice
func (d *PythonASTDetector) classToPatterns(classPatterns map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	baseClassDescriptions := map[string]string{
		"extends:Model":     "Django Model",
		"extends:BaseModel": "Pydantic BaseModel",
		"extends:APIView":   "Django REST APIView",
		"extends:TestCase":  "Test case class",
		"extends:Exception": "Custom exception",
		"extends:Enum":      "Enum class",
		"extends:ABC":       "Abstract base class",
	}

	for pattern, files := range classPatterns {
		desc := baseClassDescriptions[pattern]
		if desc == "" {
			if strings.HasPrefix(pattern, "extends:") {
				desc = "Inherits from " + strings.TrimPrefix(pattern, "extends:")
			} else {
				desc = "Python class pattern"
			}
		}

		name := pattern
		if strings.HasPrefix(pattern, "extends:") {
			name = "Inherits " + strings.TrimPrefix(pattern, "extends:")
		}

		patterns = append(patterns, types.PatternInfo{
			Category:    "Python Classes",
			Name:        name,
			Description: desc,
			FileCount:   len(files),
			Examples:    limitSlice(dedupe(files), 3),
		})
	}

	return patterns
}

// funcToPatterns converts function patterns to PatternInfo slice
func (d *PythonASTDetector) funcToPatterns(funcPatterns map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	patternDescriptions := map[string]string{
		"async def":  "Async function definitions",
		"type hints": "Type annotations",
		"generator":  "Generator functions",
	}

	for pattern, files := range funcPatterns {
		desc := patternDescriptions[pattern]
		if desc == "" {
			desc = "Python function pattern"
		}
		patterns = append(patterns, types.PatternInfo{
			Category:    "Python Patterns",
			Name:        pattern,
			Description: desc,
			FileCount:   len(files),
			Examples:    limitSlice(dedupe(files), 3),
		})
	}

	return patterns
}

// pyAppendUnique appends a string to slice if not already present
func pyAppendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
