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

	// Collect directories at multiple levels (up to 2 levels deep for src/, app/, etc.)
	dirCounts := make(map[string]int)
	expandDirs := map[string]bool{
		"src": true, "app": true, "apps": true, "packages": true,
		"lib": true, "internal": true, "pkg": true,
	}

	for _, f := range d.files {
		if f.IsDir {
			continue
		}
		dir := filepath.Dir(f.Path)
		if dir == "." {
			continue
		}

		parts := strings.Split(dir, string(filepath.Separator))
		topLevel := parts[0]

		// For expandable directories, also count second level
		if expandDirs[topLevel] && len(parts) >= 2 {
			secondLevel := parts[0] + "/" + parts[1]
			dirCounts[secondLevel]++
		} else {
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

	// For nested paths like src/components, check the last segment
	parts := strings.Split(dirLower, "/")
	lastPart := parts[len(parts)-1]

	purposes := map[string]string{
		// Source code
		"src":           "Source code",
		"source":        "Source code",
		"lib":           "Library code",
		"pkg":           "Packages",
		"internal":      "Internal packages",
		"vendor":        "Vendored dependencies",
		"node_modules":  "Node.js dependencies",

		// Frontend
		"app":           "Application pages/routes",
		"apps":          "Application modules",
		"pages":         "Page components",
		"components":    "UI components",
		"ui":            "UI components",
		"views":         "View components",
		"layouts":       "Layout components",
		"templates":     "Templates",
		"widgets":       "Widget components",
		"primitives":    "Base UI primitives",
		"blocks":        "Complex UI blocks",
		"shared":        "Shared utilities and components",

		// Backend
		"api":           "API endpoints",
		"routes":        "Route handlers",
		"router":        "Route definitions",
		"routers":       "Route definitions",
		"controllers":   "Controllers",
		"controller":    "Controllers",
		"handlers":      "Request handlers",
		"handler":       "Request handlers",
		"services":      "Business logic services",
		"service":       "Business logic services",
		"usecases":      "Use case implementations",
		"usecase":       "Use case implementations",
		"models":        "Data models",
		"model":         "Data models",
		"entities":      "Database entities",
		"entity":        "Database entities",
		"schemas":       "Schema definitions",
		"schema":        "Schema definitions",
		"middleware":    "Middleware",
		"middlewares":   "Middleware",
		"interceptors":  "Request interceptors",
		"guards":        "Auth guards",
		"validators":    "Input validators",
		"serializers":   "Data serializers",
		"resources":     "REST resources",
		"repositories":  "Data repositories",
		"repository":    "Data repositories",
		"dao":           "Data access objects",

		// Data & Database
		"db":            "Database",
		"database":      "Database",
		"data":          "Data layer",
		"data_adapter":  "Data adapters",
		"adapters":      "Adapters",
		"adapter":       "Adapters",
		"migrations":    "Database migrations",
		"migrate":       "Database migrations",
		"seeds":         "Database seeds",
		"seeders":       "Database seeders",
		"prisma":        "Prisma schema and migrations",
		"drizzle":       "Drizzle ORM",

		// Config
		"config":        "Configuration",
		"configs":       "Configuration",
		"configuration": "Configuration",
		"settings":      "Settings",
		"env":           "Environment configuration",
		"environments":  "Environment configs",

		// Utilities
		"utils":         "Utilities",
		"util":          "Utilities",
		"utilities":     "Utilities",
		"helpers":       "Helper functions",
		"helper":        "Helper functions",
		"tools":         "Tools",
		"scripts":       "Scripts",
		"bin":           "Binary/executable scripts",
		"cmd":           "Command entrypoints",
		"cli":           "CLI commands",
		"common":        "Common utilities",
		"core":          "Core functionality",
		"analyzer":      "Analysis logic",
		"analyser":      "Analysis logic",
		"parser":        "Parsing logic",
		"parsers":       "Parsing logic",
		"generator":     "Code generation",
		"generators":    "Code generation",
		"merger":        "Merge utilities",
		"walker":        "File tree walking",
		"scanner":       "Scanning utilities",
		"detector":      "Detection logic",
		"detectors":     "Detection logic",
		"builder":       "Builder utilities",
		"builders":      "Builder utilities",
		"formatter":     "Formatting utilities",
		"formatters":    "Formatting utilities",
		"transformer":   "Data transformation",
		"transformers":  "Data transformation",
		"converter":     "Data conversion",
		"converters":    "Data conversion",
		"processor":     "Data processing",
		"processors":    "Data processing",
		"runner":        "Execution runner",
		"runners":       "Execution runners",
		"executor":      "Command execution",
		"executors":     "Command execution",

		// Testing
		"test":          "Tests",
		"tests":         "Tests",
		"__tests__":     "Tests",
		"spec":          "Test specifications",
		"specs":         "Test specifications",
		"e2e":           "End-to-end tests",
		"integration":   "Integration tests",
		"unit":          "Unit tests",
		"fixtures":      "Test fixtures",
		"mocks":         "Mock data/services",
		"__mocks__":     "Jest mocks",
		"stubs":         "Test stubs",
		"factories":     "Test factories",

		// Assets
		"public":        "Public static assets",
		"static":        "Static files",
		"assets":        "Assets (images, fonts, etc.)",
		"images":        "Images",
		"img":           "Images",
		"icons":         "Icons",
		"fonts":         "Fonts",
		"styles":        "Stylesheets",
		"css":           "CSS files",
		"scss":          "SCSS files",
		"sass":          "Sass files",
		"less":          "Less files",
		"media":         "Media files",

		// Documentation
		"docs":          "Documentation",
		"doc":           "Documentation",
		"documentation": "Documentation",

		// Types & Interfaces
		"types":         "Type definitions",
		"interfaces":    "Interfaces",
		"typings":       "Type definitions",
		"@types":        "Type definitions",
		"dtos":          "Data transfer objects",
		"dto":           "Data transfer objects",

		// State Management
		"store":         "State management",
		"stores":        "State stores",
		"state":         "State management",
		"redux":         "Redux store",
		"context":       "React context",
		"contexts":      "React contexts",
		"providers":     "Context providers",
		"atoms":         "Jotai/Recoil atoms",
		"slices":        "Redux slices",

		// Features & Modules
		"features":      "Feature modules",
		"feature":       "Feature modules",
		"modules":       "Modules",
		"module":        "Modules",
		"domains":       "Domain modules",
		"domain":        "Domain modules",
		"packages":      "Monorepo packages",

		// Hooks
		"hooks":         "Custom hooks",
		"hook":          "Custom hooks",
		"composables":   "Vue composables",

		// Constants & Enums
		"constants":     "Constants",
		"consts":        "Constants",
		"enums":         "Enumerations",
		"enum":          "Enumerations",

		// Localization
		"locales":       "Localization files",
		"locale":        "Localization files",
		"i18n":          "Internationalization",
		"translations":  "Translations",
		"lang":          "Language files",

		// Workers & Jobs
		"workers":       "Web workers",
		"worker":        "Web workers",
		"jobs":          "Background jobs",
		"job":           "Background jobs",
		"queues":        "Job queues",
		"queue":         "Job queues",
		"tasks":         "Async tasks",
		"cron":          "Cron jobs",
		"schedulers":    "Task schedulers",

		// GraphQL
		"graphql":       "GraphQL schema/resolvers",
		"resolvers":     "GraphQL resolvers",
		"resolver":      "GraphQL resolvers",
		"mutations":     "GraphQL mutations",
		"queries":       "GraphQL queries",
		"subscriptions": "GraphQL subscriptions",

		// Events & Messaging
		"events":        "Event handlers",
		"event":         "Event handlers",
		"listeners":     "Event listeners",
		"subscribers":   "Event subscribers",
		"publishers":    "Event publishers",
		"emitters":      "Event emitters",

		// Auth & Security
		"auth":          "Authentication",
		"authentication": "Authentication",
		"authorization": "Authorization",
		"security":      "Security",
		"permissions":   "Permission handling",
		"policies":      "Auth policies",

		// Logging & Monitoring
		"logger":        "Logging utilities",
		"logging":       "Logging utilities",
		"logs":          "Log files",
		"metrics":       "Metrics collection",
		"monitoring":    "Monitoring utilities",
		"telemetry":     "Telemetry",
		"tracing":       "Distributed tracing",
		"alerting":      "Alert handling",

		// Email & Notifications
		"email":         "Email handling",
		"emails":        "Email templates",
		"email_templates": "Email templates",
		"mail":          "Mail handling",
		"notifications": "Notifications",
		"sms":           "SMS handling",

		// Integrations
		"integrations":  "Third-party integrations",
		"external":      "External services",
		"clients":       "API clients",
		"client":        "API clients",
		"connectors":    "Service connectors",
		"gateways":      "Service gateways",

		// Server & Deployment
		"server":        "Server configuration",
		"deploy":        "Deployment configs",
		"deployment":    "Deployment configs",
		"k8s":           "Kubernetes configs",
		"kubernetes":    "Kubernetes configs",
		"terraform":     "Terraform configs",
		"ansible":       "Ansible configs",
		"ci":            "CI/CD configuration",
		".github":       "GitHub configuration",
		".circleci":     "CircleCI configuration",

		// Misc
		"tmp":           "Temporary files",
		"temp":          "Temporary files",
		"cache":         "Cache files",
		"build":         "Build output",
		"dist":          "Distribution files",
		"out":           "Output files",
		"output":        "Output files",
		"generated":     "Generated files",
		"gen":           "Generated files",
		"proto":         "Protocol buffer definitions",
		"protos":        "Protocol buffer definitions",
		"grpc":          "gRPC definitions",
	}

	// Check full path first (for nested like src/components)
	if purpose, ok := purposes[dirLower]; ok {
		return purpose
	}

	// Then check last segment
	if purpose, ok := purposes[lastPart]; ok {
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
