package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// TechStackDetector detects the technology stack
type TechStackDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewTechStackDetector creates a new tech stack detector
func NewTechStackDetector(rootPath string, files []types.FileInfo) *TechStackDetector {
	return &TechStackDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes the codebase and returns the tech stack
func (d *TechStackDetector) Detect() (*types.TechStack, error) {
	stack := &types.TechStack{
		Languages:  []types.Language{},
		Frameworks: []types.Framework{},
		Databases:  []string{},
		Tools:      []string{},
	}

	// Detect languages from file extensions
	d.detectLanguages(stack)

	// Detect from package managers
	d.detectFromPackageJSON(stack)
	d.detectFromGoMod(stack)
	d.detectFromPython(stack)
	d.detectFromCargo(stack)

	// Detect from config files
	d.detectFromConfigFiles(stack)

	return stack, nil
}

// detectLanguages detects languages from file extensions
func (d *TechStackDetector) detectLanguages(stack *types.TechStack) {
	extToLang := map[string]string{
		".js":    "JavaScript",
		".jsx":   "JavaScript",
		".ts":    "TypeScript",
		".tsx":   "TypeScript",
		".py":    "Python",
		".go":    "Go",
		".rs":    "Rust",
		".rb":    "Ruby",
		".java":  "Java",
		".kt":    "Kotlin",
		".swift": "Swift",
		".php":   "PHP",
		".cs":    "C#",
		".cpp":   "C++",
		".c":     "C",
		".vue":   "Vue",
		".svelte": "Svelte",
	}

	langCounts := make(map[string]int)
	totalFiles := 0

	for _, f := range d.files {
		if f.IsDir {
			continue
		}
		if lang, ok := extToLang[f.Extension]; ok {
			langCounts[lang]++
			totalFiles++
		}
	}

	for lang, count := range langCounts {
		percentage := float64(count) / float64(totalFiles) * 100
		if percentage >= 1 { // Only include if >= 1%
			stack.Languages = append(stack.Languages, types.Language{
				Name:       lang,
				Percentage: percentage,
			})
		}
	}
}

// PackageJSON represents package.json structure
type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

// detectFromPackageJSON detects from package.json
func (d *TechStackDetector) detectFromPackageJSON(stack *types.TechStack) {
	pkgPath := filepath.Join(d.rootPath, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return
	}

	var pkg PackageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	// Merge all dependencies
	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	// Detect frameworks
	frameworkMap := map[string]struct{ name, category string }{
		// Frontend frameworks
		"next":         {"Next.js", "frontend"},
		"react":        {"React", "frontend"},
		"vue":          {"Vue.js", "frontend"},
		"nuxt":         {"Nuxt.js", "frontend"},
		"svelte":       {"Svelte", "frontend"},
		"@angular/core": {"Angular", "frontend"},
		"solid-js":     {"SolidJS", "frontend"},
		"astro":        {"Astro", "frontend"},

		// Backend frameworks
		"express":      {"Express.js", "backend"},
		"fastify":      {"Fastify", "backend"},
		"koa":          {"Koa", "backend"},
		"hono":         {"Hono", "backend"},
		"@nestjs/core": {"NestJS", "backend"},

		// Full-stack
		"remix":        {"Remix", "fullstack"},

		// Styling
		"tailwindcss":  {"TailwindCSS", "styling"},
		"@chakra-ui/react": {"Chakra UI", "styling"},
		"@mui/material": {"Material UI", "styling"},
		"styled-components": {"Styled Components", "styling"},

		// State management
		"redux":        {"Redux", "state"},
		"zustand":      {"Zustand", "state"},
		"@tanstack/react-query": {"React Query", "state"},
		"swr":          {"SWR", "state"},

		// Testing
		"jest":         {"Jest", "testing"},
		"vitest":       {"Vitest", "testing"},
		"mocha":        {"Mocha", "testing"},
		"@playwright/test": {"Playwright", "testing"},
		"cypress":      {"Cypress", "testing"},

		// ORM/Database
		"prisma":       {"Prisma", "database"},
		"@prisma/client": {"Prisma", "database"},
		"drizzle-orm":  {"Drizzle", "database"},
		"mongoose":     {"Mongoose", "database"},
		"typeorm":      {"TypeORM", "database"},
		"sequelize":    {"Sequelize", "database"},

		// Tools
		"typescript":   {"TypeScript", "language"},
		"eslint":       {"ESLint", "tooling"},
		"prettier":     {"Prettier", "tooling"},
	}

	seen := make(map[string]bool)
	for dep, version := range allDeps {
		if fw, ok := frameworkMap[dep]; ok {
			if !seen[fw.name] {
				stack.Frameworks = append(stack.Frameworks, types.Framework{
					Name:     fw.name,
					Version:  cleanVersion(version),
					Category: fw.category,
				})
				seen[fw.name] = true
			}
		}
	}

	// Detect databases from dependencies
	dbDeps := map[string]string{
		"pg":           "PostgreSQL",
		"mysql2":       "MySQL",
		"mongodb":      "MongoDB",
		"redis":        "Redis",
		"@supabase/supabase-js": "Supabase",
		"firebase":     "Firebase",
	}

	for dep, db := range dbDeps {
		if _, ok := allDeps[dep]; ok {
			stack.Databases = append(stack.Databases, db)
		}
	}
}

// GoMod represents go.mod structure
type GoMod struct {
	Module  string
	Go      string
	Require []string
}

// detectFromGoMod detects from go.mod
func (d *TechStackDetector) detectFromGoMod(stack *types.TechStack) {
	modPath := filepath.Join(d.rootPath, "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		return
	}

	content := string(data)

	// Extract Go version
	goVerRegex := regexp.MustCompile(`go\s+(\d+\.\d+)`)
	if matches := goVerRegex.FindStringSubmatch(content); len(matches) > 1 {
		for i, lang := range stack.Languages {
			if lang.Name == "Go" {
				stack.Languages[i].Version = matches[1]
				break
			}
		}
	}

	// Detect frameworks
	goFrameworks := map[string]struct{ name, category string }{
		"github.com/gin-gonic/gin":   {"Gin", "backend"},
		"github.com/labstack/echo":   {"Echo", "backend"},
		"github.com/gofiber/fiber":   {"Fiber", "backend"},
		"github.com/gorilla/mux":     {"Gorilla Mux", "backend"},
		"github.com/go-chi/chi":      {"Chi", "backend"},
		"github.com/spf13/cobra":     {"Cobra", "cli"},
		"github.com/urfave/cli":      {"urfave/cli", "cli"},
		"gorm.io/gorm":               {"GORM", "database"},
		"github.com/jmoiron/sqlx":    {"sqlx", "database"},
		"go.mongodb.org/mongo-driver": {"MongoDB Driver", "database"},
	}

	for pkg, fw := range goFrameworks {
		if strings.Contains(content, pkg) {
			stack.Frameworks = append(stack.Frameworks, types.Framework{
				Name:     fw.name,
				Category: fw.category,
			})
		}
	}
}

// detectFromPython detects from Python files
func (d *TechStackDetector) detectFromPython(stack *types.TechStack) {
	// Check requirements.txt
	reqPath := filepath.Join(d.rootPath, "requirements.txt")
	if data, err := os.ReadFile(reqPath); err == nil {
		d.parsePythonDeps(string(data), stack)
	}

	// Check pyproject.toml
	pyprojectPath := filepath.Join(d.rootPath, "pyproject.toml")
	if data, err := os.ReadFile(pyprojectPath); err == nil {
		d.parsePythonDeps(string(data), stack)
	}
}

func (d *TechStackDetector) parsePythonDeps(content string, stack *types.TechStack) {
	pyFrameworks := map[string]struct{ name, category string }{
		"django":      {"Django", "backend"},
		"flask":       {"Flask", "backend"},
		"fastapi":     {"FastAPI", "backend"},
		"starlette":   {"Starlette", "backend"},
		"pytest":      {"pytest", "testing"},
		"sqlalchemy":  {"SQLAlchemy", "database"},
		"alembic":     {"Alembic", "database"},
		"celery":      {"Celery", "task"},
		"pandas":      {"pandas", "data"},
		"numpy":       {"NumPy", "data"},
		"pydantic":    {"Pydantic", "validation"},
	}

	contentLower := strings.ToLower(content)
	for pkg, fw := range pyFrameworks {
		if strings.Contains(contentLower, pkg) {
			stack.Frameworks = append(stack.Frameworks, types.Framework{
				Name:     fw.name,
				Category: fw.category,
			})
		}
	}
}

// detectFromCargo detects from Cargo.toml (Rust)
func (d *TechStackDetector) detectFromCargo(stack *types.TechStack) {
	cargoPath := filepath.Join(d.rootPath, "Cargo.toml")
	data, err := os.ReadFile(cargoPath)
	if err != nil {
		return
	}

	content := string(data)

	rustFrameworks := map[string]struct{ name, category string }{
		"actix-web": {"Actix Web", "backend"},
		"axum":      {"Axum", "backend"},
		"rocket":    {"Rocket", "backend"},
		"tokio":     {"Tokio", "async"},
		"serde":     {"Serde", "serialization"},
		"diesel":    {"Diesel", "database"},
		"sqlx":      {"SQLx", "database"},
		"clap":      {"Clap", "cli"},
	}

	for pkg, fw := range rustFrameworks {
		if strings.Contains(content, pkg) {
			stack.Frameworks = append(stack.Frameworks, types.Framework{
				Name:     fw.name,
				Category: fw.category,
			})
		}
	}
}

// detectFromConfigFiles detects from various config files
func (d *TechStackDetector) detectFromConfigFiles(stack *types.TechStack) {
	configChecks := []struct {
		file   string
		tool   string
		isTool bool
	}{
		{"tsconfig.json", "TypeScript", false},
		{"tailwind.config.js", "TailwindCSS", false},
		{"tailwind.config.ts", "TailwindCSS", false},
		{".eslintrc", "ESLint", true},
		{".eslintrc.js", "ESLint", true},
		{".eslintrc.json", "ESLint", true},
		{".prettierrc", "Prettier", true},
		{"prettier.config.js", "Prettier", true},
		{"docker-compose.yml", "Docker Compose", true},
		{"docker-compose.yaml", "Docker Compose", true},
		{"Dockerfile", "Docker", true},
		{".github/workflows", "GitHub Actions", true},
		{"vercel.json", "Vercel", true},
		{"netlify.toml", "Netlify", true},
	}

	for _, check := range configChecks {
		checkPath := filepath.Join(d.rootPath, check.file)
		if _, err := os.Stat(checkPath); err == nil {
			if check.isTool {
				// Avoid duplicates
				found := false
				for _, t := range stack.Tools {
					if t == check.tool {
						found = true
						break
					}
				}
				if !found {
					stack.Tools = append(stack.Tools, check.tool)
				}
			}
		}
	}
}

// cleanVersion removes ^ ~ and other prefixes from version strings
func cleanVersion(version string) string {
	version = strings.TrimPrefix(version, "^")
	version = strings.TrimPrefix(version, "~")
	version = strings.TrimPrefix(version, ">=")
	version = strings.TrimPrefix(version, ">")
	version = strings.TrimPrefix(version, "<=")
	version = strings.TrimPrefix(version, "<")
	return version
}
