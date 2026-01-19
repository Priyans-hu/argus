package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// StructureDetector detects project structure
type StructureDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewStructureDetector creates a new structure detector
func NewStructureDetector(rootPath string, files []types.FileInfo) *StructureDetector {
	return &StructureDetector{
		rootPath: rootPath,
		files:    files,
	}
}

// Detect analyzes and returns the project structure
func (d *StructureDetector) Detect() (*types.ProjectStructure, error) {
	structure := &types.ProjectStructure{
		Directories: []types.Directory{},
		RootFiles:   []string{},
	}

	// Collect root files
	for _, f := range d.files {
		if !f.IsDir && !strings.Contains(f.Path, string(filepath.Separator)) {
			structure.RootFiles = append(structure.RootFiles, f.Name)
		}
	}

	// Collect and analyze directories
	dirCounts := make(map[string]int)
	for _, f := range d.files {
		if f.IsDir {
			continue
		}
		dir := filepath.Dir(f.Path)
		if dir != "." {
			// Get top-level directory
			parts := strings.Split(dir, string(filepath.Separator))
			topLevel := parts[0]
			dirCounts[topLevel]++
		}
	}

	// Create directory entries with purposes
	for dir, count := range dirCounts {
		purpose := inferDirectoryPurpose(dir)
		structure.Directories = append(structure.Directories, types.Directory{
			Path:      dir,
			Purpose:   purpose,
			FileCount: count,
		})
	}

	return structure, nil
}

// inferDirectoryPurpose guesses the purpose of a directory from its name
func inferDirectoryPurpose(dirName string) string {
	dirLower := strings.ToLower(dirName)

	purposes := map[string]string{
		// Source code
		"src":         "Source code",
		"source":      "Source code",
		"lib":         "Library code",
		"pkg":         "Packages",
		"internal":    "Internal packages",

		// Frontend
		"app":         "Application pages/routes",
		"pages":       "Page components",
		"components":  "UI components",
		"ui":          "UI components",
		"views":       "View components",
		"layouts":     "Layout components",
		"templates":   "Templates",

		// Backend
		"api":         "API endpoints",
		"routes":      "Route handlers",
		"controllers": "Controllers",
		"handlers":    "Request handlers",
		"services":    "Business logic services",
		"models":      "Data models",
		"entities":    "Database entities",
		"schemas":     "Schema definitions",
		"middleware":  "Middleware",
		"middlewares": "Middleware",

		// Data
		"db":          "Database",
		"database":    "Database",
		"migrations":  "Database migrations",
		"seeds":       "Database seeds",
		"prisma":      "Prisma schema and migrations",

		// Config
		"config":      "Configuration",
		"configs":     "Configuration",
		"settings":    "Settings",
		"env":         "Environment configuration",

		// Utilities
		"utils":       "Utilities",
		"util":        "Utilities",
		"helpers":     "Helper functions",
		"tools":       "Tools",
		"scripts":     "Scripts",
		"bin":         "Binary/executable scripts",
		"cmd":         "Command entrypoints",

		// Testing
		"test":        "Tests",
		"tests":       "Tests",
		"__tests__":   "Tests",
		"spec":        "Test specifications",
		"specs":       "Test specifications",
		"e2e":         "End-to-end tests",
		"integration": "Integration tests",
		"unit":        "Unit tests",
		"fixtures":    "Test fixtures",
		"mocks":       "Mock data/services",

		// Assets
		"public":      "Public static assets",
		"static":      "Static files",
		"assets":      "Assets (images, fonts, etc.)",
		"images":      "Images",
		"img":         "Images",
		"icons":       "Icons",
		"fonts":       "Fonts",
		"styles":      "Stylesheets",
		"css":         "CSS files",
		"scss":        "SCSS files",

		// Documentation
		"docs":        "Documentation",
		"doc":         "Documentation",
		"documentation": "Documentation",

		// Types
		"types":       "Type definitions",
		"interfaces":  "Interfaces",
		"typings":     "Type definitions",
		"@types":      "Type definitions",

		// State
		"store":       "State management",
		"stores":      "State stores",
		"state":       "State management",
		"redux":       "Redux store",
		"context":     "React context",
		"contexts":    "React contexts",

		// Features
		"features":    "Feature modules",
		"modules":     "Modules",
		"domains":     "Domain modules",

		// Hooks
		"hooks":       "Custom hooks",

		// Constants
		"constants":   "Constants",
		"consts":      "Constants",

		// Localization
		"locales":     "Localization files",
		"i18n":        "Internationalization",
		"translations": "Translations",

		// Workers
		"workers":     "Web workers",
		"jobs":        "Background jobs",
		"queues":      "Job queues",

		// GraphQL
		"graphql":     "GraphQL schema/resolvers",
		"resolvers":   "GraphQL resolvers",
		"mutations":   "GraphQL mutations",
		"queries":     "GraphQL queries",
	}

	if purpose, ok := purposes[dirLower]; ok {
		return purpose
	}

	return ""
}

// DetectKeyFiles identifies important files in the project
func (d *StructureDetector) DetectKeyFiles() []types.KeyFile {
	var keyFiles []types.KeyFile

	keyFilePatterns := map[string]struct{ purpose, desc string }{
		// Entry points
		"main.go":           {"Entry point", "Go application entry"},
		"main.ts":           {"Entry point", "TypeScript entry"},
		"main.js":           {"Entry point", "JavaScript entry"},
		"index.ts":          {"Entry point", "TypeScript index"},
		"index.js":          {"Entry point", "JavaScript index"},
		"app.ts":            {"Application", "Application setup"},
		"app.js":            {"Application", "Application setup"},
		"server.ts":         {"Server", "Server setup"},
		"server.js":         {"Server", "Server setup"},

		// Config
		"package.json":      {"Package config", "Node.js dependencies and scripts"},
		"tsconfig.json":     {"TypeScript config", "TypeScript compiler options"},
		"go.mod":            {"Go module", "Go dependencies"},
		"Cargo.toml":        {"Cargo config", "Rust dependencies"},
		"requirements.txt":  {"Python deps", "Python dependencies"},
		"pyproject.toml":    {"Python config", "Python project config"},

		// Database
		"schema.prisma":     {"Database schema", "Prisma database schema"},
		"docker-compose.yml": {"Docker config", "Docker services"},
		"docker-compose.yaml": {"Docker config", "Docker services"},
		"Dockerfile":        {"Docker", "Container definition"},

		// Environment
		".env.example":      {"Env template", "Environment variables template"},
		".env.sample":       {"Env template", "Environment variables template"},

		// CI/CD
		".github/workflows/ci.yml": {"CI config", "GitHub Actions CI"},
		".gitlab-ci.yml":   {"CI config", "GitLab CI"},
		"Jenkinsfile":      {"CI config", "Jenkins pipeline"},

		// Documentation
		"README.md":         {"Documentation", "Project documentation"},
		"CONTRIBUTING.md":   {"Contributing", "Contribution guidelines"},

		// Auth/Security
		"middleware.ts":     {"Middleware", "Request middleware"},
		"auth.ts":           {"Authentication", "Auth utilities"},
		"auth.js":           {"Authentication", "Auth utilities"},
	}

	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		// Check exact matches
		if kf, ok := keyFilePatterns[f.Name]; ok {
			keyFiles = append(keyFiles, types.KeyFile{
				Path:        f.Path,
				Purpose:     kf.purpose,
				Description: kf.desc,
			})
			continue
		}

		// Check path patterns
		for pattern, kf := range keyFilePatterns {
			if strings.HasSuffix(f.Path, pattern) {
				keyFiles = append(keyFiles, types.KeyFile{
					Path:        f.Path,
					Purpose:     kf.purpose,
					Description: kf.desc,
				})
				break
			}
		}
	}

	return keyFiles
}

// DetectCommands extracts available commands from package.json scripts
func DetectCommands(rootPath string) []types.Command {
	var commands []types.Command

	// Try package.json
	pkgPath := filepath.Join(rootPath, "package.json")
	if data, err := readJSON(pkgPath); err == nil {
		if pkg, ok := data.(map[string]interface{}); ok {
			if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
				for name, cmd := range scripts {
					if cmdStr, ok := cmd.(string); ok {
						commands = append(commands, types.Command{
							Name:        "npm run " + name,
							Command:     cmdStr,
							Description: inferScriptDescription(name),
						})
					}
				}
			}
		}
	}

	// Try Makefile
	makefilePath := filepath.Join(rootPath, "Makefile")
	if _, err := readFile(makefilePath); err == nil {
		// Basic Makefile parsing could be added here
		commands = append(commands, types.Command{
			Name:        "make",
			Description: "Run Makefile targets",
		})
	}

	return commands
}

func inferScriptDescription(name string) string {
	descriptions := map[string]string{
		"dev":       "Start development server",
		"start":     "Start the application",
		"build":     "Build for production",
		"test":      "Run tests",
		"lint":      "Run linter",
		"format":    "Format code",
		"typecheck": "Run type checking",
		"preview":   "Preview production build",
		"deploy":    "Deploy the application",
		"db:push":   "Push database schema",
		"db:migrate": "Run database migrations",
		"db:seed":   "Seed the database",
		"generate":  "Generate code/types",
	}

	if desc, ok := descriptions[name]; ok {
		return desc
	}
	return ""
}

func readJSON(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func readFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
