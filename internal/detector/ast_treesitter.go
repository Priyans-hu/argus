package detector

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// TreeSitterDetector uses tree-sitter for multi-language AST parsing
type TreeSitterDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewTreeSitterDetector creates a new tree-sitter based detector
func NewTreeSitterDetector(rootPath string, files []types.FileInfo) *TreeSitterDetector {
	return &TreeSitterDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes code using tree-sitter and returns patterns
func (d *TreeSitterDetector) Detect() []types.PatternInfo {
	var patterns []types.PatternInfo

	// Track imports and patterns by language
	jsImports := make(map[string][]string)
	tsImports := make(map[string][]string)
	pyImports := make(map[string][]string)

	// Parse files by language
	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		fullPath := filepath.Join(d.rootPath, f.Path)

		switch {
		case strings.HasSuffix(f.Name, ".js") || strings.HasSuffix(f.Name, ".jsx"):
			if !strings.Contains(f.Path, "node_modules") && !strings.Contains(f.Path, ".min.") {
				d.parseJavaScript(fullPath, f.Path, jsImports)
			}
		case strings.HasSuffix(f.Name, ".ts") || strings.HasSuffix(f.Name, ".tsx"):
			if !strings.Contains(f.Path, "node_modules") {
				d.parseTypeScript(fullPath, f.Path, tsImports)
			}
		case strings.HasSuffix(f.Name, ".py"):
			if !strings.Contains(f.Path, "__pycache__") && !strings.Contains(f.Path, ".pyc") {
				d.parsePython(fullPath, f.Path, pyImports)
			}
		}
	}

	// Convert imports to patterns
	patterns = append(patterns, d.jsImportsToPatterns(jsImports)...)
	patterns = append(patterns, d.tsImportsToPatterns(tsImports)...)
	patterns = append(patterns, d.pyImportsToPatterns(pyImports)...)

	return patterns
}

// parseJavaScript parses JS/JSX files using tree-sitter
func (d *TreeSitterDetector) parseJavaScript(fullPath, relPath string, imports map[string][]string) {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return
	}

	parser := sitter.NewParser()
	parser.SetLanguage(javascript.GetLanguage())

	tree, err := parser.ParseCtx(context.TODO(), nil, content)
	if err != nil {
		return
	}
	defer tree.Close()

	d.extractJSImports(tree.RootNode(), content, relPath, imports)
}

// parseTypeScript parses TS/TSX files using tree-sitter
func (d *TreeSitterDetector) parseTypeScript(fullPath, relPath string, imports map[string][]string) {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return
	}

	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	tree, err := parser.ParseCtx(context.TODO(), nil, content)
	if err != nil {
		return
	}
	defer tree.Close()

	d.extractJSImports(tree.RootNode(), content, relPath, imports)
}

// parsePython parses Python files using tree-sitter
func (d *TreeSitterDetector) parsePython(fullPath, relPath string, imports map[string][]string) {
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return
	}

	parser := sitter.NewParser()
	parser.SetLanguage(python.GetLanguage())

	tree, err := parser.ParseCtx(context.TODO(), nil, content)
	if err != nil {
		return
	}
	defer tree.Close()

	d.extractPyImports(tree.RootNode(), content, relPath, imports)
}

// extractJSImports extracts import statements from JS/TS AST
func (d *TreeSitterDetector) extractJSImports(node *sitter.Node, content []byte, relPath string, imports map[string][]string) {
	// Walk the AST looking for import declarations
	cursor := sitter.NewTreeCursor(node)
	defer cursor.Close()

	d.walkJSTree(cursor, content, relPath, imports)
}

// walkJSTree recursively walks the JS/TS AST
func (d *TreeSitterDetector) walkJSTree(cursor *sitter.TreeCursor, content []byte, relPath string, imports map[string][]string) {
	node := cursor.CurrentNode()

	// Check for import_statement or call_expression (require)
	switch node.Type() {
	case "import_statement":
		// Extract the source from import statement
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "string" || child.Type() == "string_fragment" {
				importPath := strings.Trim(child.Content(content), "\"'")
				imports[importPath] = append(imports[importPath], relPath)
			}
		}
	case "call_expression":
		// Check for require() calls
		if node.ChildCount() >= 2 {
			funcNode := node.Child(0)
			if funcNode != nil && funcNode.Content(content) == "require" {
				argsNode := node.Child(1)
				if argsNode != nil && argsNode.Type() == "arguments" {
					for i := 0; i < int(argsNode.ChildCount()); i++ {
						arg := argsNode.Child(i)
						if arg.Type() == "string" || arg.Type() == "string_fragment" {
							importPath := strings.Trim(arg.Content(content), "\"'")
							imports[importPath] = append(imports[importPath], relPath)
						}
					}
				}
			}
		}
	}

	// Recurse into children
	if cursor.GoToFirstChild() {
		for {
			d.walkJSTree(cursor, content, relPath, imports)
			if !cursor.GoToNextSibling() {
				break
			}
		}
		cursor.GoToParent()
	}
}

// extractPyImports extracts import statements from Python AST
func (d *TreeSitterDetector) extractPyImports(node *sitter.Node, content []byte, relPath string, imports map[string][]string) {
	cursor := sitter.NewTreeCursor(node)
	defer cursor.Close()

	d.walkPyTree(cursor, content, relPath, imports)
}

// walkPyTree recursively walks the Python AST
func (d *TreeSitterDetector) walkPyTree(cursor *sitter.TreeCursor, content []byte, relPath string, imports map[string][]string) {
	node := cursor.CurrentNode()

	switch node.Type() {
	case "import_statement":
		// import foo, bar
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "dotted_name" {
				importPath := child.Content(content)
				imports[importPath] = append(imports[importPath], relPath)
			}
		}
	case "import_from_statement":
		// from foo import bar
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "dotted_name" {
				importPath := child.Content(content)
				imports[importPath] = append(imports[importPath], relPath)
				break // Only get the module name, not the imported items
			}
		}
	}

	// Recurse into children
	if cursor.GoToFirstChild() {
		for {
			d.walkPyTree(cursor, content, relPath, imports)
			if !cursor.GoToNextSibling() {
				break
			}
		}
		cursor.GoToParent()
	}
}

// jsImportsToPatterns converts JS imports to pattern info
func (d *TreeSitterDetector) jsImportsToPatterns(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	// Known JS/React frameworks and libraries
	jsFrameworks := map[string]string{
		"react":                  "React UI library",
		"react-dom":              "React DOM rendering",
		"next":                   "Next.js framework",
		"next/router":            "Next.js routing",
		"next/link":              "Next.js navigation",
		"next/image":             "Next.js image optimization",
		"vue":                    "Vue.js framework",
		"@angular/core":          "Angular framework",
		"svelte":                 "Svelte framework",
		"express":                "Express.js web framework",
		"fastify":                "Fastify web framework",
		"koa":                    "Koa web framework",
		"axios":                  "Axios HTTP client",
		"lodash":                 "Lodash utility library",
		"moment":                 "Moment.js date library",
		"dayjs":                  "Day.js date library",
		"date-fns":               "date-fns date library",
		"redux":                  "Redux state management",
		"@reduxjs/toolkit":       "Redux Toolkit",
		"mobx":                   "MobX state management",
		"zustand":                "Zustand state management",
		"jotai":                  "Jotai state management",
		"recoil":                 "Recoil state management",
		"@tanstack/react-query":  "TanStack Query (React Query)",
		"swr":                    "SWR data fetching",
		"tailwindcss":            "Tailwind CSS",
		"styled-components":      "styled-components CSS-in-JS",
		"@emotion/react":         "Emotion CSS-in-JS",
		"jest":                   "Jest testing framework",
		"mocha":                  "Mocha testing framework",
		"vitest":                 "Vitest testing framework",
		"@testing-library/react": "React Testing Library",
		"cypress":                "Cypress E2E testing",
		"playwright":             "Playwright E2E testing",
		"prisma":                 "Prisma ORM",
		"@prisma/client":         "Prisma Client",
		"mongoose":               "Mongoose MongoDB ODM",
		"sequelize":              "Sequelize ORM",
		"typeorm":                "TypeORM",
		"drizzle-orm":            "Drizzle ORM",
		"zod":                    "Zod schema validation",
		"yup":                    "Yup schema validation",
		"joi":                    "Joi schema validation",
		"socket.io":              "Socket.IO real-time",
		"ws":                     "WebSocket library",
		"graphql":                "GraphQL",
		"@apollo/client":         "Apollo Client GraphQL",
		"trpc":                   "tRPC",
		"@trpc/server":           "tRPC Server",
		"@trpc/client":           "tRPC Client",
	}

	for importPath, files := range imports {
		// Check exact match first
		if desc, ok := jsFrameworks[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        importPath,
				Category:    "js-framework",
				Description: desc,
				FileCount:   len(dedupeStrings(files)),
				Examples:    limitStrings(dedupeStrings(files), 3),
			})
			continue
		}

		// Check prefix match for scoped packages
		for framework, desc := range jsFrameworks {
			if strings.HasPrefix(importPath, framework+"/") {
				patterns = append(patterns, types.PatternInfo{
					Name:        framework,
					Category:    "js-framework",
					Description: desc,
					FileCount:   len(dedupeStrings(files)),
					Examples:    limitStrings(dedupeStrings(files), 3),
				})
				break
			}
		}
	}

	return patterns
}

// tsImportsToPatterns converts TS imports to pattern info
func (d *TreeSitterDetector) tsImportsToPatterns(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	// TypeScript-specific patterns
	tsFrameworks := map[string]string{
		"typescript":        "TypeScript language",
		"@types/node":       "Node.js type definitions",
		"@types/react":      "React type definitions",
		"ts-node":           "TypeScript execution",
		"tsx":               "TypeScript execution (tsx)",
		"@nestjs/core":      "NestJS framework",
		"@nestjs/common":    "NestJS common utilities",
		"type-graphql":      "TypeGraphQL",
		"typedi":            "TypeDI dependency injection",
		"class-validator":   "class-validator validation",
		"class-transformer": "class-transformer serialization",
		"inversify":         "InversifyJS DI container",
		"tsyringe":          "TSyringe DI container",
	}

	for importPath, files := range imports {
		if desc, ok := tsFrameworks[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        importPath,
				Category:    "ts-framework",
				Description: desc,
				FileCount:   len(dedupeStrings(files)),
				Examples:    limitStrings(dedupeStrings(files), 3),
			})
			continue
		}

		// Check prefix match
		for framework, desc := range tsFrameworks {
			if strings.HasPrefix(importPath, framework+"/") {
				patterns = append(patterns, types.PatternInfo{
					Name:        framework,
					Category:    "ts-framework",
					Description: desc,
					FileCount:   len(dedupeStrings(files)),
					Examples:    limitStrings(dedupeStrings(files), 3),
				})
				break
			}
		}
	}

	// Also include JS patterns for TS files
	patterns = append(patterns, d.jsImportsToPatterns(imports)...)

	return patterns
}

// pyImportsToPatterns converts Python imports to pattern info
func (d *TreeSitterDetector) pyImportsToPatterns(imports map[string][]string) []types.PatternInfo {
	var patterns []types.PatternInfo

	// Known Python frameworks and libraries
	pyFrameworks := map[string]string{
		"django":       "Django web framework",
		"flask":        "Flask web framework",
		"fastapi":      "FastAPI web framework",
		"starlette":    "Starlette ASGI framework",
		"tornado":      "Tornado web framework",
		"aiohttp":      "aiohttp async HTTP",
		"requests":     "Requests HTTP library",
		"httpx":        "HTTPX async HTTP client",
		"sqlalchemy":   "SQLAlchemy ORM",
		"alembic":      "Alembic migrations",
		"pydantic":     "Pydantic data validation",
		"marshmallow":  "Marshmallow serialization",
		"celery":       "Celery task queue",
		"redis":        "Redis client",
		"pymongo":      "PyMongo MongoDB driver",
		"motor":        "Motor async MongoDB driver",
		"psycopg2":     "psycopg2 PostgreSQL driver",
		"asyncpg":      "asyncpg PostgreSQL driver",
		"pytest":       "pytest testing framework",
		"unittest":     "unittest testing",
		"nose":         "nose testing framework",
		"hypothesis":   "Hypothesis property testing",
		"numpy":        "NumPy numerical computing",
		"pandas":       "Pandas data analysis",
		"scipy":        "SciPy scientific computing",
		"matplotlib":   "Matplotlib plotting",
		"seaborn":      "Seaborn statistical plots",
		"plotly":       "Plotly interactive plots",
		"tensorflow":   "TensorFlow ML framework",
		"torch":        "PyTorch ML framework",
		"keras":        "Keras deep learning",
		"scikit-learn": "scikit-learn ML library",
		"sklearn":      "scikit-learn ML library",
		"transformers": "Hugging Face Transformers",
		"langchain":    "LangChain LLM framework",
		"openai":       "OpenAI API client",
		"anthropic":    "Anthropic API client",
		"boto3":        "AWS SDK for Python",
		"botocore":     "AWS SDK core",
		"google.cloud": "Google Cloud SDK",
		"azure":        "Azure SDK",
		"click":        "Click CLI framework",
		"typer":        "Typer CLI framework",
		"argparse":     "argparse CLI parsing",
		"logging":      "Python logging",
		"structlog":    "structlog structured logging",
		"loguru":       "Loguru logging",
		"asyncio":      "asyncio async programming",
		"aiofiles":     "aiofiles async file I/O",
		"uvicorn":      "Uvicorn ASGI server",
		"gunicorn":     "Gunicorn WSGI server",
		"poetry":       "Poetry dependency management",
		"pipenv":       "Pipenv dependency management",
	}

	for importPath, files := range imports {
		// Get the top-level module name
		topModule := strings.Split(importPath, ".")[0]

		if desc, ok := pyFrameworks[topModule]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        topModule,
				Category:    "py-framework",
				Description: desc,
				FileCount:   len(dedupeStrings(files)),
				Examples:    limitStrings(dedupeStrings(files), 3),
			})
		} else if desc, ok := pyFrameworks[importPath]; ok {
			patterns = append(patterns, types.PatternInfo{
				Name:        importPath,
				Category:    "py-framework",
				Description: desc,
				FileCount:   len(dedupeStrings(files)),
				Examples:    limitStrings(dedupeStrings(files), 3),
			})
		}
	}

	return patterns
}

// Helper functions

func dedupeStrings(strs []string) []string {
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

func limitStrings(strs []string, n int) []string {
	if len(strs) <= n {
		return strs
	}
	return strs[:n]
}
