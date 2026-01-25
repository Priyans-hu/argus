package detector

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
)

// JSASTDetector uses AST parsing for accurate JavaScript/TypeScript pattern detection
type JSASTDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewJSASTDetector creates a new AST-based JS/TS detector
func NewJSASTDetector(rootPath string, files []types.FileInfo) *JSASTDetector {
	return &JSASTDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes JavaScript/TypeScript code using AST and returns patterns
func (d *JSASTDetector) Detect() []types.PatternInfo {
	// Track patterns across all JS/TS files
	imports := make(map[string][]string)       // import path -> files using it
	reactHooks := make(map[string][]string)    // hook name -> files using it
	funcPatterns := make(map[string][]string)  // pattern -> files
	classPatterns := make(map[string][]string) // class pattern -> files

	// Parse all JS/TS files
	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		// Check for JS/TS extensions
		ext := strings.ToLower(f.Extension)
		if ext != ".js" && ext != ".jsx" && ext != ".ts" && ext != ".tsx" && ext != ".mjs" && ext != ".cjs" {
			continue
		}

		// Skip common non-source directories
		if strings.Contains(f.Path, "node_modules") ||
			strings.Contains(f.Path, ".min.") ||
			strings.Contains(f.Path, "dist/") ||
			strings.Contains(f.Path, "build/") {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// For TypeScript files, strip type annotations for parsing
		// (goja doesn't support TypeScript syntax directly)
		if ext == ".ts" || ext == ".tsx" {
			content = stripTypeAnnotations(content)
		}

		// Parse the file
		program, err := parser.ParseFile(nil, fullPath, string(content), 0)
		if err != nil {
			continue // Skip files that can't be parsed
		}

		// Extract patterns from AST
		d.extractFromProgram(program, f.Path, imports, reactHooks, funcPatterns, classPatterns)
	}

	// Convert to PatternInfo
	var patterns []types.PatternInfo
	patterns = append(patterns, d.importsToPatterns(imports)...)
	patterns = append(patterns, d.hooksToPatterns(reactHooks)...)
	patterns = append(patterns, d.funcToPatterns(funcPatterns)...)
	patterns = append(patterns, d.classToPatterns(classPatterns)...)

	return patterns
}

// extractFromProgram walks the AST and extracts patterns
func (d *JSASTDetector) extractFromProgram(program *ast.Program, filePath string,
	imports, reactHooks, funcPatterns, classPatterns map[string][]string) {

	for _, stmt := range program.Body {
		d.visitStatement(stmt, filePath, imports, reactHooks, funcPatterns, classPatterns)
	}
}

// visitStatement recursively visits statements
func (d *JSASTDetector) visitStatement(stmt ast.Statement, filePath string,
	imports, reactHooks, funcPatterns, classPatterns map[string][]string) {

	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.VariableStatement:
		// Check for require() calls
		for _, decl := range s.List {
			d.visitExpression(decl.Initializer, filePath, imports, reactHooks, funcPatterns)
		}

	case *ast.FunctionDeclaration:
		if s.Function != nil {
			if s.Function.Async {
				funcPatterns["async function"] = appendUnique(funcPatterns["async function"], filePath)
			}
			if s.Function.Generator {
				funcPatterns["generator function"] = appendUnique(funcPatterns["generator function"], filePath)
			}
			// Visit function body
			if s.Function.Body != nil {
				d.visitStatement(s.Function.Body, filePath, imports, reactHooks, funcPatterns, classPatterns)
			}
		}

	case *ast.ClassDeclaration:
		classPatterns["ES6 class"] = appendUnique(classPatterns["ES6 class"], filePath)
		if s.Class != nil && s.Class.SuperClass != nil {
			classPatterns["class inheritance"] = appendUnique(classPatterns["class inheritance"], filePath)
		}

	case *ast.BlockStatement:
		for _, inner := range s.List {
			d.visitStatement(inner, filePath, imports, reactHooks, funcPatterns, classPatterns)
		}

	case *ast.ExpressionStatement:
		d.visitExpression(s.Expression, filePath, imports, reactHooks, funcPatterns)

	case *ast.ReturnStatement:
		d.visitExpression(s.Argument, filePath, imports, reactHooks, funcPatterns)

	case *ast.IfStatement:
		d.visitStatement(s.Consequent, filePath, imports, reactHooks, funcPatterns, classPatterns)
		d.visitStatement(s.Alternate, filePath, imports, reactHooks, funcPatterns, classPatterns)

	case *ast.ForStatement:
		d.visitStatement(s.Body, filePath, imports, reactHooks, funcPatterns, classPatterns)

	case *ast.ForInStatement:
		d.visitStatement(s.Body, filePath, imports, reactHooks, funcPatterns, classPatterns)

	case *ast.ForOfStatement:
		d.visitStatement(s.Body, filePath, imports, reactHooks, funcPatterns, classPatterns)

	case *ast.WhileStatement:
		d.visitStatement(s.Body, filePath, imports, reactHooks, funcPatterns, classPatterns)

	case *ast.TryStatement:
		d.visitStatement(s.Body, filePath, imports, reactHooks, funcPatterns, classPatterns)
		if s.Catch != nil {
			d.visitStatement(s.Catch.Body, filePath, imports, reactHooks, funcPatterns, classPatterns)
		}
		d.visitStatement(s.Finally, filePath, imports, reactHooks, funcPatterns, classPatterns)
	}
}

// visitExpression recursively visits expressions
func (d *JSASTDetector) visitExpression(expr ast.Expression, filePath string,
	imports, reactHooks, funcPatterns map[string][]string) {

	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		// Check for require() calls
		if ident, ok := e.Callee.(*ast.Identifier); ok {
			name := string(ident.Name)
			if name == "require" && len(e.ArgumentList) > 0 {
				if lit, ok := e.ArgumentList[0].(*ast.StringLiteral); ok {
					importPath := lit.Value.String()
					imports[importPath] = appendUnique(imports[importPath], filePath)
				}
			}
			// Check for React hooks
			if strings.HasPrefix(name, "use") && len(name) > 3 {
				reactHooks[name] = appendUnique(reactHooks[name], filePath)
			}
		}
		// Visit callee and arguments
		d.visitExpression(e.Callee, filePath, imports, reactHooks, funcPatterns)
		for _, arg := range e.ArgumentList {
			d.visitExpression(arg, filePath, imports, reactHooks, funcPatterns)
		}

	case *ast.ArrowFunctionLiteral:
		funcPatterns["arrow function"] = appendUnique(funcPatterns["arrow function"], filePath)
		if e.Async {
			funcPatterns["async arrow"] = appendUnique(funcPatterns["async arrow"], filePath)
		}

	case *ast.AwaitExpression:
		funcPatterns["await"] = appendUnique(funcPatterns["await"], filePath)
		d.visitExpression(e.Argument, filePath, imports, reactHooks, funcPatterns)

	case *ast.FunctionLiteral:
		if e.Async {
			funcPatterns["async function"] = appendUnique(funcPatterns["async function"], filePath)
		}
		if e.Generator {
			funcPatterns["generator function"] = appendUnique(funcPatterns["generator function"], filePath)
		}

	case *ast.AssignExpression:
		d.visitExpression(e.Right, filePath, imports, reactHooks, funcPatterns)

	case *ast.BinaryExpression:
		d.visitExpression(e.Left, filePath, imports, reactHooks, funcPatterns)
		d.visitExpression(e.Right, filePath, imports, reactHooks, funcPatterns)

	case *ast.ConditionalExpression:
		d.visitExpression(e.Consequent, filePath, imports, reactHooks, funcPatterns)
		d.visitExpression(e.Alternate, filePath, imports, reactHooks, funcPatterns)

	case *ast.ArrayLiteral:
		for _, item := range e.Value {
			d.visitExpression(item, filePath, imports, reactHooks, funcPatterns)
		}

	case *ast.ObjectLiteral:
		for _, prop := range e.Value {
			if p, ok := prop.(*ast.PropertyKeyed); ok {
				d.visitExpression(p.Value, filePath, imports, reactHooks, funcPatterns)
			}
		}
	}
}

// stripTypeAnnotations removes TypeScript type annotations for parsing
// This is a simplified approach - it won't handle all TS syntax
func stripTypeAnnotations(content []byte) []byte {
	src := string(content)

	// Remove type imports: import type { ... } from '...'
	src = strings.ReplaceAll(src, "import type ", "// import type ")

	// Remove interface declarations
	lines := strings.Split(src, "\n")
	var result []string
	inInterface := false
	braceCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip interface blocks
		if strings.HasPrefix(trimmed, "interface ") || strings.HasPrefix(trimmed, "export interface ") {
			inInterface = true
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")
			continue
		}

		if inInterface {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount <= 0 {
				inInterface = false
			}
			continue
		}

		// Skip type declarations
		if strings.HasPrefix(trimmed, "type ") || strings.HasPrefix(trimmed, "export type ") {
			continue
		}

		result = append(result, line)
	}

	return []byte(strings.Join(result, "\n"))
}

// importsToPatterns converts import map to PatternInfo slice
func (d *JSASTDetector) importsToPatterns(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	// Framework detection based on imports
	frameworkImports := map[string]string{
		"react":             "React",
		"react-dom":         "React DOM",
		"next":              "Next.js",
		"next/router":       "Next.js Router",
		"next/link":         "Next.js",
		"vue":               "Vue.js",
		"@vue/":             "Vue.js",
		"angular":           "Angular",
		"@angular/":         "Angular",
		"svelte":            "Svelte",
		"express":           "Express.js",
		"fastify":           "Fastify",
		"koa":               "Koa",
		"hapi":              "Hapi",
		"@nestjs/":          "NestJS",
		"axios":             "Axios",
		"@tanstack/":        "TanStack Query",
		"react-query":       "React Query",
		"swr":               "SWR",
		"redux":             "Redux",
		"@reduxjs/":         "Redux Toolkit",
		"zustand":           "Zustand",
		"mobx":              "MobX",
		"jotai":             "Jotai",
		"recoil":            "Recoil",
		"tailwindcss":       "Tailwind CSS",
		"styled-components": "styled-components",
		"@emotion/":         "Emotion",
		"@mui/":             "Material-UI",
		"@chakra-ui/":       "Chakra UI",
		"jest":              "Jest",
		"vitest":            "Vitest",
		"mocha":             "Mocha",
		"cypress":           "Cypress",
		"playwright":        "Playwright",
		"prisma":            "Prisma",
		"@prisma/":          "Prisma",
		"drizzle-orm":       "Drizzle ORM",
		"typeorm":           "TypeORM",
		"sequelize":         "Sequelize",
		"mongoose":          "Mongoose",
		"graphql":           "GraphQL",
		"@apollo/":          "Apollo",
		"socket.io":         "Socket.IO",
		"ws":                "WebSocket",
		"zod":               "Zod",
		"yup":               "Yup",
		"joi":               "Joi",
		"lodash":            "Lodash",
		"dayjs":             "Day.js",
		"moment":            "Moment.js",
		"date-fns":          "date-fns",
	}

	for importPath, files := range imports {
		for prefix, framework := range frameworkImports {
			if strings.HasPrefix(importPath, prefix) || importPath == prefix {
				patterns = append(patterns, types.PatternInfo{
					Category:    "JavaScript Frameworks",
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

// hooksToPatterns converts hooks map to PatternInfo slice
func (d *JSASTDetector) hooksToPatterns(hooks map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	hookDescriptions := map[string]string{
		"useState":        "React state management",
		"useEffect":       "React side effects",
		"useContext":      "React context consumption",
		"useReducer":      "React reducer pattern",
		"useCallback":     "React memoized callbacks",
		"useMemo":         "React memoized values",
		"useRef":          "React refs",
		"useLayoutEffect": "React synchronous effects",
		"useQuery":        "Data fetching (React Query/TanStack)",
		"useMutation":     "Data mutation (React Query/TanStack)",
		"useRouter":       "Routing hook",
		"useParams":       "Route parameters",
		"useNavigate":     "Navigation hook",
		"useSelector":     "Redux state selection",
		"useDispatch":     "Redux dispatch",
		"useForm":         "Form handling",
		"useAuth":         "Authentication hook",
	}

	for hook, files := range hooks {
		desc := hookDescriptions[hook]
		if desc == "" {
			desc = "Custom React hook"
		}
		patterns = append(patterns, types.PatternInfo{
			Category:    "React Hooks",
			Name:        hook,
			Description: desc,
			FileCount:   len(files),
			Examples:    limitSlice(dedupe(files), 3),
		})
	}

	return patterns
}

// funcToPatterns converts function patterns to PatternInfo slice
func (d *JSASTDetector) funcToPatterns(funcPatterns map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	patternDescriptions := map[string]string{
		"async function":     "Async/await pattern",
		"generator function": "Generator functions",
		"arrow function":     "Arrow function syntax",
		"async arrow":        "Async arrow functions",
		"await":              "Await expressions",
	}

	for pattern, files := range funcPatterns {
		desc := patternDescriptions[pattern]
		if desc == "" {
			desc = "JavaScript pattern"
		}
		patterns = append(patterns, types.PatternInfo{
			Category:    "JavaScript Patterns",
			Name:        pattern,
			Description: desc,
			FileCount:   len(files),
			Examples:    limitSlice(dedupe(files), 3),
		})
	}

	return patterns
}

// classToPatterns converts class patterns to PatternInfo slice
func (d *JSASTDetector) classToPatterns(classPatterns map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	for pattern, files := range classPatterns {
		patterns = append(patterns, types.PatternInfo{
			Category:    "JavaScript Patterns",
			Name:        pattern,
			Description: "ES6+ class feature",
			FileCount:   len(files),
			Examples:    limitSlice(dedupe(files), 3),
		})
	}

	return patterns
}

// appendUnique appends a string to slice if not already present
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}
