package detector

import (
	"cmp"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// TechStackDetector detects the technology stack
type TechStackDetector struct {
	rootPath string
	files    []types.FileInfo
	deps     DepSet
}

// NewTechStackDetector creates a new tech stack detector; frameworks come
// from deps, the declared dependencies
func NewTechStackDetector(rootPath string, files []types.FileInfo, deps []types.Dependency) *TechStackDetector {
	return &TechStackDetector{
		rootPath: rootPath,
		files:    files,
		deps:     NewDepSet(deps),
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

	// Detect from declared dependencies
	d.detectFromPackageJSON(stack)
	d.detectFromGoMod(stack)
	d.detectFromPython(stack)
	d.detectFromCargo(stack)

	d.detectFromAppManifests(stack)

	// Detect from config files
	d.detectFromConfigFiles(stack)

	return stack, nil
}

// detectLanguages detects languages from file extensions
func (d *TechStackDetector) detectLanguages(stack *types.TechStack) {
	extToLang := map[string]string{
		".js":     "JavaScript",
		".jsx":    "JavaScript",
		".ts":     "TypeScript",
		".tsx":    "TypeScript",
		".py":     "Python",
		".go":     "Go",
		".rs":     "Rust",
		".rb":     "Ruby",
		".java":   "Java",
		".kt":     "Kotlin",
		".swift":  "Swift",
		".php":    "PHP",
		".cs":     "C#",
		".cpp":    "C++",
		".c":      "C",
		".vue":    "Vue",
		".svelte": "Svelte",
		".dart":   "Dart",
		".zig":    "Zig",
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
	// Languages[0] is used as the primary language, so ties need a fixed winner
	slices.SortFunc(stack.Languages, func(a, b types.Language) int {
		return cmp.Or(cmp.Compare(b.Percentage, a.Percentage), cmp.Compare(a.Name, b.Name))
	})
}

// detectFromPackageJSON detects npm frameworks and databases
func (d *TechStackDetector) detectFromPackageJSON(stack *types.TechStack) {
	// Detect frameworks
	frameworkMap := map[string]struct{ name, category string }{
		// Frontend frameworks
		"next":          {"Next.js", "frontend"},
		"react":         {"React", "frontend"},
		"vue":           {"Vue.js", "frontend"},
		"nuxt":          {"Nuxt.js", "frontend"},
		"svelte":        {"Svelte", "frontend"},
		"@sveltejs/kit": {"SvelteKit", "frontend"},
		"@angular/core": {"Angular", "frontend"},
		"solid-js":      {"SolidJS", "frontend"},
		"astro":         {"Astro", "frontend"},

		// Backend frameworks
		"express":      {"Express.js", "backend"},
		"fastify":      {"Fastify", "backend"},
		"koa":          {"Koa", "backend"},
		"hono":         {"Hono", "backend"},
		"@nestjs/core": {"NestJS", "backend"},

		// Full-stack
		"remix":           {"Remix", "fullstack"},
		"@tauri-apps/cli": {"Tauri", "fullstack"},
		"@tauri-apps/api": {"Tauri", "fullstack"},

		// Styling
		"tailwindcss":       {"TailwindCSS", "styling"},
		"@chakra-ui/react":  {"Chakra UI", "styling"},
		"@mui/material":     {"Material UI", "styling"},
		"styled-components": {"Styled Components", "styling"},

		// State management
		"redux":                 {"Redux", "state"},
		"zustand":               {"Zustand", "state"},
		"@tanstack/react-query": {"React Query", "state"},
		"swr":                   {"SWR", "state"},

		// Testing
		"jest":             {"Jest", "testing"},
		"vitest":           {"Vitest", "testing"},
		"mocha":            {"Mocha", "testing"},
		"@playwright/test": {"Playwright", "testing"},
		"cypress":          {"Cypress", "testing"},

		// ORM/Database
		"prisma":         {"Prisma", "database"},
		"@prisma/client": {"Prisma", "database"},
		"drizzle-orm":    {"Drizzle", "database"},
		"mongoose":       {"Mongoose", "database"},
		"typeorm":        {"TypeORM", "database"},
		"sequelize":      {"Sequelize", "database"},

		// Tools
		"typescript": {"TypeScript", "language"},
		"eslint":     {"ESLint", "tooling"},
		"prettier":   {"Prettier", "tooling"},
	}

	// Sorted, so a framework listed under two packages (prisma, @prisma/client)
	// always takes its version from the same one
	for _, dep := range slices.Sorted(maps.Keys(frameworkMap)) {
		if version, ok := d.deps[dep]; ok {
			fw := frameworkMap[dep]
			addFramework(stack, types.Framework{
				Name:     fw.name,
				Version:  cleanVersion(version),
				Category: fw.category,
			})
		}
	}

	// Detect databases from dependencies
	dbDeps := map[string]string{
		"pg":                    "PostgreSQL",
		"mysql2":                "MySQL",
		"mongodb":               "MongoDB",
		"redis":                 "Redis",
		"@supabase/supabase-js": "Supabase",
		"firebase":              "Firebase",
	}

	for _, dep := range slices.Sorted(maps.Keys(dbDeps)) {
		if d.deps.Has(dep) && !slices.Contains(stack.Databases, dbDeps[dep]) {
			stack.Databases = append(stack.Databases, dbDeps[dep])
		}
	}
}

// GoMod represents go.mod structure
type GoMod struct {
	Module  string
	Go      string
	Require []string
}

// detectFromGoMod detects Go frameworks; the root go.mod gives the Go version
func (d *TechStackDetector) detectFromGoMod(stack *types.TechStack) {
	if data, err := os.ReadFile(filepath.Join(d.rootPath, "go.mod")); err == nil {
		goVerRegex := regexp.MustCompile(`go\s+(\d+\.\d+)`)
		if matches := goVerRegex.FindStringSubmatch(string(data)); len(matches) > 1 {
			for i, lang := range stack.Languages {
				if lang.Name == "Go" {
					stack.Languages[i].Version = matches[1]
					break
				}
			}
		}
	}

	// Detect frameworks
	goFrameworks := map[string]struct{ name, category string }{
		"github.com/gin-gonic/gin":    {"Gin", "backend"},
		"github.com/labstack/echo":    {"Echo", "backend"},
		"github.com/gofiber/fiber":    {"Fiber", "backend"},
		"github.com/gorilla/mux":      {"Gorilla Mux", "backend"},
		"github.com/go-chi/chi":       {"Chi", "backend"},
		"github.com/spf13/cobra":      {"Cobra", "cli"},
		"github.com/urfave/cli":       {"urfave/cli", "cli"},
		"gorm.io/gorm":                {"GORM", "database"},
		"github.com/jmoiron/sqlx":     {"sqlx", "database"},
		"go.mongodb.org/mongo-driver": {"MongoDB Driver", "database"},
	}

	for _, pkg := range slices.Sorted(maps.Keys(goFrameworks)) {
		if d.deps.HasModule(pkg) {
			addFramework(stack, types.Framework{Name: goFrameworks[pkg].name, Category: goFrameworks[pkg].category})
		}
	}
}

// detectFromPython detects Python frameworks
func (d *TechStackDetector) detectFromPython(stack *types.TechStack) {
	pyFrameworks := map[string]struct{ name, category string }{
		"django":     {"Django", "backend"},
		"flask":      {"Flask", "backend"},
		"fastapi":    {"FastAPI", "backend"},
		"starlette":  {"Starlette", "backend"},
		"pytest":     {"pytest", "testing"},
		"sqlalchemy": {"SQLAlchemy", "database"},
		"alembic":    {"Alembic", "database"},
		"celery":     {"Celery", "task"},
		"pandas":     {"pandas", "data"},
		"numpy":      {"NumPy", "data"},
		"pydantic":   {"Pydantic", "validation"},
	}

	for _, pkg := range slices.Sorted(maps.Keys(pyFrameworks)) {
		if d.deps.Has(pkg) {
			addFramework(stack, types.Framework{Name: pyFrameworks[pkg].name, Category: pyFrameworks[pkg].category})
		}
	}
}

// detectFromCargo detects Rust frameworks
func (d *TechStackDetector) detectFromCargo(stack *types.TechStack) {
	rustFrameworks := map[string]struct{ name, category string }{
		"actix-web": {"Actix Web", "backend"},
		"axum":      {"Axum", "backend"},
		"rocket":    {"Rocket", "backend"},
		"tokio":     {"Tokio", "async"},
		"serde":     {"Serde", "serialization"},
		"diesel":    {"Diesel", "database"},
		"sqlx":      {"SQLx", "database"},
		"clap":      {"Clap", "cli"},
		"tauri":     {"Tauri", "fullstack"},
	}

	for _, pkg := range slices.Sorted(maps.Keys(rustFrameworks)) {
		if d.deps.Has(pkg) {
			addFramework(stack, types.Framework{Name: rustFrameworks[pkg].name, Category: rustFrameworks[pkg].category})
		}
	}
}

var flutterSDK = regexp.MustCompile(`(?m)^\s*sdk:\s*flutter\b`)

// detectFromAppManifests detects Flutter from a pubspec.yaml that depends on
// the Flutter SDK and Tauri from a tauri.conf.json, at the root or up to two
// directories deep in project source.
func (d *TechStackDetector) detectFromAppManifests(stack *types.TechStack) {
	for _, f := range d.files {
		rel := filepath.ToSlash(f.Path)
		if f.IsDir || strings.Count(rel, "/") > 2 || !IsProjectSource(rel) {
			continue
		}
		switch f.Name {
		case "pubspec.yaml":
			if data, err := os.ReadFile(filepath.Join(d.rootPath, f.Path)); err == nil && flutterSDK.Match(data) {
				addFramework(stack, types.Framework{Name: "Flutter", Category: "frontend"})
			}
		case "tauri.conf.json":
			addFramework(stack, types.Framework{Name: "Tauri", Category: "fullstack"})
		}
	}
}

// addFramework appends fw unless the stack already lists a framework of
// that name (prisma and @prisma/client, for one).
func addFramework(stack *types.TechStack, fw types.Framework) {
	if !slices.ContainsFunc(stack.Frameworks, func(f types.Framework) bool { return f.Name == fw.Name }) {
		stack.Frameworks = append(stack.Frameworks, fw)
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
