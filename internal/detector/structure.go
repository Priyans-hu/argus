package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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
		"src":          "Source code",
		"source":       "Source code",
		"lib":          "Library code",
		"pkg":          "Packages",
		"internal":     "Internal packages",
		"vendor":       "Vendored dependencies",
		"node_modules": "Node.js dependencies",

		// Frontend
		"app":        "Application pages/routes",
		"apps":       "Application modules",
		"pages":      "Page components",
		"components": "UI components",
		"ui":         "UI components",
		"views":      "View components",
		"layouts":    "Layout components",
		"templates":  "Templates",
		"widgets":    "Widget components",
		"primitives": "Base UI primitives",
		"blocks":     "Complex UI blocks",
		"shared":     "Shared utilities and components",

		// Backend
		"api":          "API endpoints",
		"routes":       "Route handlers",
		"router":       "Route definitions",
		"routers":      "Route definitions",
		"controllers":  "Controllers",
		"controller":   "Controllers",
		"handlers":     "Request handlers",
		"handler":      "Request handlers",
		"services":     "Business logic services",
		"service":      "Business logic services",
		"usecases":     "Use case implementations",
		"usecase":      "Use case implementations",
		"models":       "Data models",
		"model":        "Data models",
		"entities":     "Database entities",
		"entity":       "Database entities",
		"schemas":      "Schema definitions",
		"schema":       "Schema definitions",
		"middleware":   "Middleware",
		"middlewares":  "Middleware",
		"interceptors": "Request interceptors",
		"guards":       "Auth guards",
		"validators":   "Input validators",
		"serializers":  "Data serializers",
		"resources":    "REST resources",
		"repositories": "Data repositories",
		"repository":   "Data repositories",
		"dao":          "Data access objects",

		// Data & Database
		"db":           "Database",
		"database":     "Database",
		"data":         "Data layer",
		"data_adapter": "Data adapters",
		"adapters":     "Adapters",
		"adapter":      "Adapters",
		"migrations":   "Database migrations",
		"migrate":      "Database migrations",
		"seeds":        "Database seeds",
		"seeders":      "Database seeders",
		"prisma":       "Prisma schema and migrations",
		"drizzle":      "Drizzle ORM",

		// Config
		"config":        "Configuration",
		"configs":       "Configuration",
		"configuration": "Configuration",
		"settings":      "Settings",
		"env":           "Environment configuration",
		"environments":  "Environment configs",

		// Utilities
		"utils":        "Utilities",
		"util":         "Utilities",
		"utilities":    "Utilities",
		"helpers":      "Helper functions",
		"helper":       "Helper functions",
		"tools":        "Tools",
		"scripts":      "Scripts",
		"bin":          "Binary/executable scripts",
		"cmd":          "Command entrypoints",
		"cli":          "CLI commands",
		"common":       "Common utilities",
		"core":         "Core functionality",
		"analyzer":     "Analysis logic",
		"analyser":     "Analysis logic",
		"parser":       "Parsing logic",
		"parsers":      "Parsing logic",
		"generator":    "Code generation",
		"generators":   "Code generation",
		"merger":       "Merge utilities",
		"walker":       "File tree walking",
		"scanner":      "Scanning utilities",
		"detector":     "Detection logic",
		"detectors":    "Detection logic",
		"builder":      "Builder utilities",
		"builders":     "Builder utilities",
		"formatter":    "Formatting utilities",
		"formatters":   "Formatting utilities",
		"transformer":  "Data transformation",
		"transformers": "Data transformation",
		"converter":    "Data conversion",
		"converters":   "Data conversion",
		"processor":    "Data processing",
		"processors":   "Data processing",
		"runner":       "Execution runner",
		"runners":      "Execution runners",
		"executor":     "Command execution",
		"executors":    "Command execution",

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
		"__mocks__":   "Jest mocks",
		"stubs":       "Test stubs",
		"factories":   "Test factories",

		// Assets
		"public": "Public static assets",
		"static": "Static files",
		"assets": "Assets (images, fonts, etc.)",
		"images": "Images",
		"img":    "Images",
		"icons":  "Icons",
		"fonts":  "Fonts",
		"styles": "Stylesheets",
		"css":    "CSS files",
		"scss":   "SCSS files",
		"sass":   "Sass files",
		"less":   "Less files",
		"media":  "Media files",

		// Documentation
		"docs":          "Documentation",
		"doc":           "Documentation",
		"documentation": "Documentation",

		// Types & Interfaces
		"types":      "Type definitions",
		"interfaces": "Interfaces",
		"typings":    "Type definitions",
		"@types":     "Type definitions",
		"dtos":       "Data transfer objects",
		"dto":        "Data transfer objects",

		// State Management
		"store":     "State management",
		"stores":    "State stores",
		"state":     "State management",
		"redux":     "Redux store",
		"context":   "React context",
		"contexts":  "React contexts",
		"providers": "Context providers",
		"atoms":     "Jotai/Recoil atoms",
		"slices":    "Redux slices",

		// Features & Modules
		"features": "Feature modules",
		"feature":  "Feature modules",
		"modules":  "Modules",
		"module":   "Modules",
		"domains":  "Domain modules",
		"domain":   "Domain modules",
		"packages": "Monorepo packages",

		// Hooks
		"hooks":       "Custom hooks",
		"hook":        "Custom hooks",
		"composables": "Vue composables",

		// Constants & Enums
		"constants": "Constants",
		"consts":    "Constants",
		"enums":     "Enumerations",
		"enum":      "Enumerations",

		// Localization
		"locales":      "Localization files",
		"locale":       "Localization files",
		"i18n":         "Internationalization",
		"translations": "Translations",
		"lang":         "Language files",

		// Workers & Jobs
		"workers":    "Web workers",
		"worker":     "Web workers",
		"jobs":       "Background jobs",
		"job":        "Background jobs",
		"queues":     "Job queues",
		"queue":      "Job queues",
		"tasks":      "Async tasks",
		"cron":       "Cron jobs",
		"schedulers": "Task schedulers",

		// GraphQL
		"graphql":       "GraphQL schema/resolvers",
		"resolvers":     "Resolver implementations",
		"resolver":      "Resolution logic",
		"mutations":     "GraphQL mutations",
		"queries":       "GraphQL queries",
		"subscriptions": "GraphQL subscriptions",

		// Error handling
		"errors":     "Error definitions",
		"error":      "Error handling",
		"exceptions": "Exception definitions",

		// History & Tracking
		"history":   "History tracking",
		"audit":     "Audit logging",
		"changelog": "Change tracking",

		// Terminal UI
		"tui":      "Terminal UI",
		"terminal": "Terminal utilities",
		"console":  "Console utilities",
		"prompt":   "Interactive prompts",

		// Events & Messaging
		"events":      "Event handlers",
		"event":       "Event handlers",
		"listeners":   "Event listeners",
		"subscribers": "Event subscribers",
		"publishers":  "Event publishers",
		"emitters":    "Event emitters",

		// Auth & Security
		"auth":           "Authentication",
		"authentication": "Authentication",
		"authorization":  "Authorization",
		"security":       "Security",
		"permissions":    "Permission handling",
		"policies":       "Auth policies",

		// Logging & Monitoring
		"logger":     "Logging utilities",
		"logging":    "Logging utilities",
		"logs":       "Log files",
		"metrics":    "Metrics collection",
		"monitoring": "Monitoring utilities",
		"telemetry":  "Telemetry",
		"tracing":    "Distributed tracing",
		"alerting":   "Alert handling",

		// Email & Notifications
		"email":           "Email handling",
		"emails":          "Email templates",
		"email_templates": "Email templates",
		"mail":            "Mail handling",
		"notifications":   "Notifications",
		"sms":             "SMS handling",

		// Integrations
		"integrations": "Third-party integrations",
		"external":     "External services",
		"clients":      "API clients",
		"client":       "API clients",
		"connectors":   "Service connectors",
		"gateways":     "Service gateways",

		// Server & Deployment
		"server":     "Server configuration",
		"deploy":     "Deployment configs",
		"deployment": "Deployment configs",
		"k8s":        "Kubernetes configs",
		"kubernetes": "Kubernetes configs",
		"terraform":  "Terraform configs",
		"ansible":    "Ansible configs",
		"ci":         "CI/CD configuration",
		".github":    "GitHub configuration",
		".circleci":  "CircleCI configuration",

		// Misc
		"tmp":       "Temporary files",
		"temp":      "Temporary files",
		"cache":     "Cache files",
		"build":     "Build output",
		"dist":      "Distribution files",
		"out":       "Output files",
		"output":    "Output files",
		"generated": "Generated files",
		"gen":       "Generated files",
		"proto":     "Protocol buffer definitions",
		"protos":    "Protocol buffer definitions",
		"grpc":      "gRPC definitions",
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
		// Entry points - Go
		"main.go": {"Entry point", "Go application entry"},

		// Entry points - JavaScript/TypeScript
		"main.ts":   {"Entry point", "TypeScript entry"},
		"main.js":   {"Entry point", "JavaScript entry"},
		"index.ts":  {"Entry point", "TypeScript index"},
		"index.js":  {"Entry point", "JavaScript index"},
		"app.ts":    {"Application", "Application setup"},
		"app.js":    {"Application", "Application setup"},
		"server.ts": {"Server", "Server setup"},
		"server.js": {"Server", "Server setup"},

		// Entry points - Python
		"manage.py":   {"Entry point", "Django management script"},
		"wsgi.py":     {"Entry point", "WSGI application entry"},
		"asgi.py":     {"Entry point", "ASGI application entry"},
		"app.py":      {"Entry point", "Flask/FastAPI application"},
		"main.py":     {"Entry point", "Python application entry"},
		"__main__.py": {"Entry point", "Python module entry"},
		"setup.py":    {"Package setup", "Python package setup"},

		// Config
		"package.json":     {"Package config", "Node.js dependencies and scripts"},
		"tsconfig.json":    {"TypeScript config", "TypeScript compiler options"},
		"go.mod":           {"Go module", "Go dependencies"},
		"Cargo.toml":       {"Cargo config", "Rust dependencies"},
		"requirements.txt": {"Python deps", "Python dependencies"},
		"pyproject.toml":   {"Python config", "Python project config"},
		"poetry.lock":      {"Poetry lock", "Poetry dependency lock"},
		"Pipfile":          {"Pipenv config", "Pipenv dependencies"},

		// Database
		"schema.prisma":       {"Database schema", "Prisma database schema"},
		"docker-compose.yml":  {"Docker config", "Docker services"},
		"docker-compose.yaml": {"Docker config", "Docker services"},
		"Dockerfile":          {"Docker", "Container definition"},

		// Environment
		".env.example": {"Env template", "Environment variables template"},
		".env.sample":  {"Env template", "Environment variables template"},

		// CI/CD
		".github/workflows/ci.yml": {"CI config", "GitHub Actions CI"},
		".gitlab-ci.yml":           {"CI config", "GitLab CI"},
		"Jenkinsfile":              {"CI config", "Jenkins pipeline"},

		// Documentation
		"README.md":       {"Documentation", "Project documentation"},
		"CONTRIBUTING.md": {"Contributing", "Contribution guidelines"},

		// Auth/Security
		"middleware.ts": {"Middleware", "Request middleware"},
		"auth.ts":       {"Authentication", "Auth utilities"},
		"auth.js":       {"Authentication", "Auth utilities"},
	}

	// Track which file types we've already added to avoid duplicates
	seen := make(map[string]bool)

	for _, f := range d.files {
		if f.IsDir {
			continue
		}

		// Special handling for README.md - only include root level
		if f.Name == "README.md" {
			if f.Path == "README.md" && !seen["README.md"] {
				keyFiles = append(keyFiles, types.KeyFile{
					Path:        f.Path,
					Purpose:     "Documentation",
					Description: "Project documentation",
				})
				seen["README.md"] = true
			}
			continue
		}

		// Special handling for CONTRIBUTING.md - only root level
		if f.Name == "CONTRIBUTING.md" {
			if f.Path == "CONTRIBUTING.md" && !seen["CONTRIBUTING.md"] {
				keyFiles = append(keyFiles, types.KeyFile{
					Path:        f.Path,
					Purpose:     "Contributing",
					Description: "Contribution guidelines",
				})
				seen["CONTRIBUTING.md"] = true
			}
			continue
		}

		// Check exact matches (non-doc files)
		if kf, ok := keyFilePatterns[f.Name]; ok {
			// Skip if we've already seen this type of file
			if seen[f.Name] {
				continue
			}
			keyFiles = append(keyFiles, types.KeyFile{
				Path:        f.Path,
				Purpose:     kf.purpose,
				Description: kf.desc,
			})
			seen[f.Name] = true
			continue
		}

		// Check path patterns (for nested config files like .github/workflows/ci.yml)
		for pattern, kf := range keyFilePatterns {
			if strings.HasSuffix(f.Path, pattern) && !seen[pattern] {
				keyFiles = append(keyFiles, types.KeyFile{
					Path:        f.Path,
					Purpose:     kf.purpose,
					Description: kf.desc,
				})
				seen[pattern] = true
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

	// Try Go commands
	goModPath := filepath.Join(rootPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		commands = append(commands, detectGoCommands(rootPath)...)
	}

	// Try Makefile
	makefilePath := filepath.Join(rootPath, "Makefile")
	if data, err := readFile(makefilePath); err == nil {
		makeTargets := parseMakefileTargets(string(data))
		commands = append(commands, makeTargets...)
	}

	// Try Cargo.toml (Rust)
	cargoPath := filepath.Join(rootPath, "Cargo.toml")
	if _, err := os.Stat(cargoPath); err == nil {
		commands = append(commands, detectCargoCommands()...)
	}

	// Try Python
	commands = append(commands, detectPythonCommands(rootPath)...)

	// Try Cobra CLI commands (Go)
	cobraCommands := detectCobraCommands(rootPath)
	commands = append(commands, cobraCommands...)

	return commands
}

// detectGoCommands returns standard Go commands
func detectGoCommands(rootPath string) []types.Command {
	var commands []types.Command

	// Check for test files
	hasTests := false
	_ = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && hasTestSuffix(d.Name()) {
			hasTests = true
			return filepath.SkipAll
		}
		return nil
	})

	commands = append(commands, types.Command{
		Name:        "go build ./...",
		Description: "Build all packages",
	})

	if hasTests {
		commands = append(commands, types.Command{
			Name:        "go test ./...",
			Description: "Run all tests",
		})
		commands = append(commands, types.Command{
			Name:        "go test -v ./...",
			Description: "Run all tests with verbose output",
		})
	}

	commands = append(commands, types.Command{
		Name:        "go fmt ./...",
		Description: "Format all Go files",
	})

	return commands
}

// parseMakefileTargets extracts targets from a Makefile
func parseMakefileTargets(content string) []types.Command {
	var commands []types.Command

	// Regex to match target definitions (target: [dependencies])
	targetRegex := regexp.MustCompile(`(?m)^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)

	// Common targets to detect with descriptions
	targetDescs := map[string]string{
		"build":    "Build the project",
		"test":     "Run tests",
		"clean":    "Clean build artifacts",
		"install":  "Install dependencies/binary",
		"run":      "Run the application",
		"dev":      "Start development mode",
		"lint":     "Run linter",
		"format":   "Format code",
		"fmt":      "Format code",
		"check":    "Run checks",
		"all":      "Build all targets",
		"help":     "Show available targets",
		"docker":   "Build Docker image",
		"deploy":   "Deploy the application",
		"release":  "Create a release",
		"coverage": "Run tests with coverage",
		"bench":    "Run benchmarks",
		"generate": "Generate code",
		"proto":    "Generate protobuf code",
		"migrate":  "Run database migrations",
		"seed":     "Seed the database",
	}

	matches := targetRegex.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		target := match[1]

		// Skip internal targets (starting with .)
		if target[0] == '.' {
			continue
		}

		// Skip if already added
		if seen[target] {
			continue
		}
		seen[target] = true

		desc := targetDescs[target]
		commands = append(commands, types.Command{
			Name:        "make " + target,
			Description: desc,
		})
	}

	return commands
}

// detectCargoCommands returns standard Cargo commands
func detectCargoCommands() []types.Command {
	return []types.Command{
		{Name: "cargo build", Description: "Build the project"},
		{Name: "cargo build --release", Description: "Build for release"},
		{Name: "cargo test", Description: "Run tests"},
		{Name: "cargo fmt", Description: "Format code"},
		{Name: "cargo clippy", Description: "Run linter"},
	}
}

// detectPythonCommands returns Python commands if applicable
func detectPythonCommands(rootPath string) []types.Command {
	var commands []types.Command

	// Check for poetry (check pyproject.toml for [tool.poetry])
	pyprojectPath := filepath.Join(rootPath, "pyproject.toml")
	hasPoetry := false
	if data, err := os.ReadFile(pyprojectPath); err == nil {
		if containsString(string(data), "[tool.poetry]") {
			hasPoetry = true
			commands = append(commands, types.Command{
				Name:        "poetry install",
				Description: "Install dependencies",
			})
			commands = append(commands, types.Command{
				Name:        "poetry shell",
				Description: "Activate virtual environment",
			})
		}
	}

	// Check for requirements.txt (if not using poetry)
	reqPath := filepath.Join(rootPath, "requirements.txt")
	hasRequirements := false
	if _, err := os.Stat(reqPath); err == nil && !hasPoetry {
		hasRequirements = true
		commands = append(commands, types.Command{
			Name:        "pip install -r requirements.txt",
			Description: "Install dependencies",
		})
	}

	// Check for Django (manage.py)
	managePath := filepath.Join(rootPath, "manage.py")
	if _, err := os.Stat(managePath); err == nil {
		prefix := "python manage.py"
		if hasPoetry {
			prefix = "poetry run python manage.py"
		}
		commands = append(commands, []types.Command{
			{
				Name:        prefix + " runserver",
				Description: "Start Django development server",
			},
			{
				Name:        prefix + " migrate",
				Description: "Run database migrations",
			},
			{
				Name:        prefix + " makemigrations",
				Description: "Create new migrations",
			},
			{
				Name:        prefix + " shell",
				Description: "Start Django shell",
			},
			{
				Name:        prefix + " test",
				Description: "Run tests",
			},
			{
				Name:        prefix + " createsuperuser",
				Description: "Create superuser",
			},
		}...)
	}

	// Check for Flask (wsgi.py or app.py with Flask import)
	wsgiPath := filepath.Join(rootPath, "wsgi.py")
	appPath := filepath.Join(rootPath, "app.py")
	hasFlask := false

	if _, err := os.Stat(wsgiPath); err == nil {
		hasFlask = true
	} else if data, err := os.ReadFile(appPath); err == nil && containsString(string(data), "Flask") {
		hasFlask = true
	}

	if hasFlask {
		runCmd := "flask run"
		if hasPoetry {
			runCmd = "poetry run flask run"
		}
		commands = append(commands, []types.Command{
			{
				Name:        runCmd,
				Description: "Start Flask development server",
			},
			{
				Name:        "flask shell",
				Description: "Start Flask shell",
			},
			{
				Name:        "flask routes",
				Description: "Show all registered routes",
			},
		}...)

		// Add gunicorn if wsgi.py exists
		if _, err := os.Stat(wsgiPath); err == nil {
			gunicornCmd := "gunicorn wsgi:app"
			if hasPoetry {
				gunicornCmd = "poetry run gunicorn wsgi:app"
			}
			commands = append(commands, types.Command{
				Name:        gunicornCmd,
				Description: "Run Flask with Gunicorn",
			})
		}
	}

	// Check for FastAPI (main.py or app.py with FastAPI import)
	mainPath := filepath.Join(rootPath, "main.py")
	hasFastAPI := false

	if data, err := os.ReadFile(mainPath); err == nil && containsString(string(data), "FastAPI") {
		hasFastAPI = true
	} else if data, err := os.ReadFile(appPath); err == nil && containsString(string(data), "FastAPI") {
		hasFastAPI = true
	}

	if hasFastAPI {
		// Determine the module:app pattern
		appModule := "main:app"
		if _, err := os.Stat(appPath); err == nil {
			appModule = "app:app"
		}

		uvicornCmd := "uvicorn " + appModule + " --reload"
		if hasPoetry {
			uvicornCmd = "poetry run uvicorn " + appModule + " --reload"
		}
		commands = append(commands, types.Command{
			Name:        uvicornCmd,
			Description: "Start FastAPI development server",
		})
	}

	// Check for pytest
	if hasRequirements || hasPoetry {
		var reqData []byte
		var err error

		if hasRequirements {
			reqData, err = os.ReadFile(reqPath)
		} else if hasPoetry {
			reqData, err = os.ReadFile(pyprojectPath)
		}

		if err == nil && containsString(string(reqData), "pytest") {
			pytestCmd := "pytest"
			if hasPoetry {
				pytestCmd = "poetry run pytest"
			}
			commands = append(commands, types.Command{
				Name:        pytestCmd,
				Description: "Run tests",
			})
			commands = append(commands, types.Command{
				Name:        pytestCmd + " -v",
				Description: "Run tests with verbose output",
			})
			commands = append(commands, types.Command{
				Name:        pytestCmd + " --cov",
				Description: "Run tests with coverage",
			})
		}
	}

	// Check for formatters (black, ruff)
	if hasRequirements || hasPoetry {
		var reqData []byte
		var err error

		if hasRequirements {
			reqData, err = os.ReadFile(reqPath)
		} else if hasPoetry {
			reqData, err = os.ReadFile(pyprojectPath)
		}

		if err == nil {
			prefix := ""
			if hasPoetry {
				prefix = "poetry run "
			}

			if containsString(string(reqData), "black") {
				commands = append(commands, types.Command{
					Name:        prefix + "black .",
					Description: "Format code with Black",
				})
			}

			if containsString(string(reqData), "ruff") {
				commands = append(commands, types.Command{
					Name:        prefix + "ruff format",
					Description: "Format code with Ruff",
				})
				commands = append(commands, types.Command{
					Name:        prefix + "ruff check",
					Description: "Lint code with Ruff",
				})
			}

			if containsString(string(reqData), "flake8") {
				commands = append(commands, types.Command{
					Name:        prefix + "flake8",
					Description: "Lint code with Flake8",
				})
			}

			if containsString(string(reqData), "mypy") {
				commands = append(commands, types.Command{
					Name:        prefix + "mypy .",
					Description: "Type check with mypy",
				})
			}
		}
	}

	// Check for pyproject.toml with pip (even without poetry)
	if _, err := os.Stat(pyprojectPath); err == nil && !hasPoetry {
		commands = append(commands, types.Command{
			Name:        "pip install -e .",
			Description: "Install package in editable mode",
		})
	}

	// Check for setup.py
	setupPath := filepath.Join(rootPath, "setup.py")
	if _, err := os.Stat(setupPath); err == nil {
		commands = append(commands, types.Command{
			Name:        "python setup.py install",
			Description: "Install the package",
		})
	}

	return commands
}

// detectCobraCommands finds Cobra CLI commands in Go projects
func detectCobraCommands(rootPath string) []types.Command {
	var commands []types.Command

	// Check if this is a Cobra project
	modPath := filepath.Join(rootPath, "go.mod")
	modData, err := os.ReadFile(modPath)
	if err != nil {
		return commands
	}
	if !containsString(string(modData), "spf13/cobra") {
		return commands
	}

	// Find the CLI name from cmd directory
	cmdDir := filepath.Join(rootPath, "cmd")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return commands
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		cliName := entry.Name()
		cmdSubDir := filepath.Join(cmdDir, cliName, "cmd")

		// Check if cmd subdir exists
		if _, err := os.Stat(cmdSubDir); os.IsNotExist(err) {
			continue
		}

		// Scan for command files
		cmdFiles, err := os.ReadDir(cmdSubDir)
		if err != nil {
			continue
		}

		for _, cmdFile := range cmdFiles {
			if cmdFile.IsDir() || !hasGoSuffix(cmdFile.Name()) {
				continue
			}
			// Skip test files and root.go
			if hasTestSuffix(cmdFile.Name()) || cmdFile.Name() == "root.go" {
				continue
			}

			// Parse the command name and description
			cmdName, cmdDesc := parseCobraCommand(filepath.Join(cmdSubDir, cmdFile.Name()))
			if cmdName != "" {
				commands = append(commands, types.Command{
					Name:        cliName + " " + cmdName,
					Description: cmdDesc,
				})
			}
		}
	}

	return commands
}

// parseCobraCommand extracts command name and description from a Cobra command file
func parseCobraCommand(filePath string) (string, string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", ""
	}

	contentStr := string(content)

	// Look for Use: "commandname" or Use: "commandname [args]"
	useRegex := regexp.MustCompile(`Use:\s*["']([^"'\s]+)`)
	useMatch := useRegex.FindStringSubmatch(contentStr)

	cmdName := ""
	if len(useMatch) >= 2 {
		cmdName = useMatch[1]
	} else {
		// Fallback: use filename without .go
		base := filepath.Base(filePath)
		cmdName = base[:len(base)-3]
	}

	// Look for Short: "description"
	shortRegex := regexp.MustCompile(`Short:\s*["']([^"']+)["']`)
	shortMatch := shortRegex.FindStringSubmatch(contentStr)

	cmdDesc := ""
	if len(shortMatch) >= 2 {
		cmdDesc = shortMatch[1]
	}

	return cmdName, cmdDesc
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findStr(s, substr) >= 0
}

func findStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func hasGoSuffix(name string) bool {
	return len(name) > 3 && name[len(name)-3:] == ".go"
}

func hasTestSuffix(name string) bool {
	return len(name) > 8 && name[len(name)-8:] == "_test.go"
}

func inferScriptDescription(name string) string {
	descriptions := map[string]string{
		"dev":        "Start development server",
		"start":      "Start the application",
		"build":      "Build for production",
		"test":       "Run tests",
		"lint":       "Run linter",
		"format":     "Format code",
		"typecheck":  "Run type checking",
		"preview":    "Preview production build",
		"deploy":     "Deploy the application",
		"db:push":    "Push database schema",
		"db:migrate": "Run database migrations",
		"db:seed":    "Seed the database",
		"generate":   "Generate code/types",
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
