package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// ProjectType represents the detected project type
type ProjectType string

const (
	ProjectTypeApp       ProjectType = "application"
	ProjectTypeLibrary   ProjectType = "library"
	ProjectTypeCLI       ProjectType = "cli"
	ProjectTypeAPI       ProjectType = "api"
	ProjectTypeML        ProjectType = "ml"
	ProjectTypeDocs      ProjectType = "documentation"
	ProjectTypeMonorepo  ProjectType = "monorepo"
	ProjectTypeFramework ProjectType = "framework"
	ProjectTypeTutorial  ProjectType = "tutorial"
	ProjectTypeConfig    ProjectType = "config"
	ProjectTypeUnknown   ProjectType = "unknown"
)

// ProjectTypeInfo contains project type inference results
type ProjectTypeInfo struct {
	PrimaryType   ProjectType
	SecondaryType ProjectType
	Confidence    float64 // 0.0 to 1.0
	Indicators    []string
	Description   string
}

// ProjectTypeDetector infers the project type
type ProjectTypeDetector struct {
	rootPath string
	files    []types.FileInfo
}

// NewProjectTypeDetector creates a new project type detector
func NewProjectTypeDetector(rootPath string, files []types.FileInfo) *ProjectTypeDetector {
	return &ProjectTypeDetector{rootPath: rootPath, files: files}
}

// Detect infers the project type
func (d *ProjectTypeDetector) Detect() *ProjectTypeInfo {
	scores := make(map[ProjectType]float64)
	indicators := make(map[ProjectType][]string)

	// Score each project type
	d.scoreDocumentation(scores, indicators)
	d.scoreLibrary(scores, indicators)
	d.scoreCLI(scores, indicators)
	d.scoreAPI(scores, indicators)
	d.scoreML(scores, indicators)
	d.scoreMonorepo(scores, indicators)
	d.scoreFramework(scores, indicators)
	d.scoreTutorial(scores, indicators)
	d.scoreApplication(scores, indicators)
	d.scoreConfig(scores, indicators)

	// Find primary and secondary types
	var primary, secondary ProjectType
	var primaryScore, secondaryScore float64

	for pt, score := range scores {
		if score > primaryScore {
			secondary = primary
			secondaryScore = primaryScore
			primary = pt
			primaryScore = score
		} else if score > secondaryScore {
			secondary = pt
			secondaryScore = score
		}
	}

	if primary == "" {
		primary = ProjectTypeUnknown
	}

	// Calculate confidence
	confidence := 0.0
	if primaryScore > 0 {
		totalScore := 0.0
		for _, s := range scores {
			totalScore += s
		}
		confidence = primaryScore / totalScore
		if confidence > 1.0 {
			confidence = 1.0
		}
	}

	return &ProjectTypeInfo{
		PrimaryType:   primary,
		SecondaryType: secondary,
		Confidence:    confidence,
		Indicators:    indicators[primary],
		Description:   d.getDescription(primary),
	}
}

// scoreDocumentation scores for documentation projects
func (d *ProjectTypeDetector) scoreDocumentation(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	score := 0.0
	var inds []string

	// Count markdown files
	mdCount := 0
	for _, f := range d.files {
		if strings.HasSuffix(f.Name, ".md") || strings.HasSuffix(f.Name, ".mdx") {
			mdCount++
		}
	}

	totalFiles := len(d.files)
	if totalFiles > 0 {
		mdRatio := float64(mdCount) / float64(totalFiles)
		if mdRatio > 0.5 {
			score += 3.0
			inds = append(inds, "Majority of files are markdown")
		} else if mdRatio > 0.3 {
			score += 1.5
		}
	}

	// Check for documentation tools
	docTools := []struct {
		files []string
		name  string
		score float64
	}{
		{[]string{"mkdocs.yml", "mkdocs.yaml"}, "MkDocs", 2.5},
		{[]string{"docusaurus.config.js", "docusaurus.config.ts"}, "Docusaurus", 2.5},
		{[]string{"_config.yml"}, "Jekyll", 2.0},
		{[]string{"book.toml"}, "mdBook", 2.0},
		{[]string{"docs/", "documentation/"}, "Docs directory", 1.0},
		{[]string{".vuepress/"}, "VuePress", 2.5},
		{[]string{"astro.config.mjs"}, "Astro (possibly docs)", 1.0},
		{[]string{"sphinx.conf", "conf.py"}, "Sphinx", 2.0},
	}

	for _, tool := range docTools {
		for _, file := range tool.files {
			checkPath := filepath.Join(d.rootPath, file)
			if _, err := os.Stat(checkPath); err == nil {
				score += tool.score
				inds = append(inds, tool.name+" configuration found")
				break
			}
		}
	}

	// Check README for "documentation" mentions
	readmePath := filepath.Join(d.rootPath, "README.md")
	if content, err := os.ReadFile(readmePath); err == nil {
		lower := strings.ToLower(string(content))
		if strings.Contains(lower, "documentation") && strings.Contains(lower, "guide") {
			score += 1.0
		}
	}

	scores[ProjectTypeDocs] = score
	indicators[ProjectTypeDocs] = inds
}

// scoreLibrary scores for library/package projects
func (d *ProjectTypeDetector) scoreLibrary(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	score := 0.0
	var inds []string

	// Check for package publishing configs
	libIndicators := []struct {
		file  string
		name  string
		score float64
	}{
		{"setup.py", "Python setup.py", 2.0},
		{"setup.cfg", "Python setup.cfg", 1.5},
		{"pyproject.toml", "Python pyproject.toml", 1.5},
		{"Cargo.toml", "Rust Cargo.toml", 1.5},
		{"go.mod", "Go module", 1.0},
		{"package.json", "npm package.json", 1.0},
		{"pom.xml", "Maven POM", 1.5},
		{"build.gradle", "Gradle build", 1.5},
		{".gemspec", "Ruby gem", 2.0},
	}

	for _, ind := range libIndicators {
		if strings.HasSuffix(ind.file, "*") {
			// Glob pattern
			pattern := filepath.Join(d.rootPath, ind.file)
			matches, _ := filepath.Glob(pattern)
			if len(matches) > 0 {
				score += ind.score
				inds = append(inds, ind.name+" found")
			}
		} else {
			checkPath := filepath.Join(d.rootPath, ind.file)
			if _, err := os.Stat(checkPath); err == nil {
				score += ind.score
				inds = append(inds, ind.name+" found")
			}
		}
	}

	// Check for src/ directory structure (common in libraries)
	srcDir := filepath.Join(d.rootPath, "src")
	if info, err := os.Stat(srcDir); err == nil && info.IsDir() {
		score += 0.5
		inds = append(inds, "Has src/ directory")
	}

	// Check for lib/ directory
	libDir := filepath.Join(d.rootPath, "lib")
	if info, err := os.Stat(libDir); err == nil && info.IsDir() {
		score += 0.5
	}

	// Check package.json for "main" or "exports" (indicates library)
	pkgPath := filepath.Join(d.rootPath, "package.json")
	if content, err := os.ReadFile(pkgPath); err == nil {
		contentStr := string(content)
		if strings.Contains(contentStr, `"main"`) || strings.Contains(contentStr, `"exports"`) {
			if !strings.Contains(contentStr, `"bin"`) { // Not a CLI
				score += 1.0
				inds = append(inds, "package.json has main/exports")
			}
		}
	}

	// Reduce score if it has app-like indicators
	for _, f := range d.files {
		if f.Name == "Dockerfile" || f.Name == "docker-compose.yml" {
			score -= 0.5
		}
	}

	scores[ProjectTypeLibrary] = score
	indicators[ProjectTypeLibrary] = inds
}

// scoreCLI scores for CLI tool projects
func (d *ProjectTypeDetector) scoreCLI(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	score := 0.0
	var inds []string

	// Check for CLI frameworks
	cliIndicators := []struct {
		pattern string
		imports []string
		name    string
		score   float64
	}{
		{"cmd/", nil, "Go cmd/ directory", 2.0},
		{"", []string{"github.com/spf13/cobra", "github.com/urfave/cli"}, "Go CLI framework", 2.5},
		{"", []string{"import click", "from click", "import typer", "from typer", "import argparse"}, "Python CLI framework", 2.0},
		{"", []string{"commander", "yargs", "inquirer", "ora", "chalk"}, "Node.js CLI libraries", 2.0},
	}

	// Check cmd/ directory
	cmdDir := filepath.Join(d.rootPath, "cmd")
	if info, err := os.Stat(cmdDir); err == nil && info.IsDir() {
		score += 2.0
		inds = append(inds, "Has cmd/ directory")
	}

	// Check for CLI imports in files
	for _, f := range d.files {
		content, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		contentStr := string(content)

		for _, ind := range cliIndicators {
			if len(ind.imports) == 0 {
				continue
			}
			for _, imp := range ind.imports {
				if strings.Contains(contentStr, imp) {
					score += ind.score
					inds = append(inds, ind.name+" detected")
					break
				}
			}
		}
	}

	// Check package.json for bin
	pkgPath := filepath.Join(d.rootPath, "package.json")
	if content, err := os.ReadFile(pkgPath); err == nil {
		if strings.Contains(string(content), `"bin"`) {
			score += 2.0
			inds = append(inds, "package.json has bin field")
		}
	}

	// Check for main.go in root (common for CLI tools)
	mainGo := filepath.Join(d.rootPath, "main.go")
	if _, err := os.Stat(mainGo); err == nil {
		score += 1.0
	}

	scores[ProjectTypeCLI] = score
	indicators[ProjectTypeCLI] = inds
}

// scoreAPI scores for API/backend projects
func (d *ProjectTypeDetector) scoreAPI(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	score := 0.0
	var inds []string

	// Check for API frameworks
	apiPatterns := []struct {
		imports []string
		name    string
		score   float64
	}{
		{[]string{"from fastapi", "import fastapi"}, "FastAPI", 2.5},
		{[]string{"from flask", "import flask"}, "Flask", 2.0},
		{[]string{"from django", "import django"}, "Django", 2.0},
		{[]string{"express", "fastify", "koa", "hapi"}, "Node.js API framework", 2.0},
		{[]string{"gin", "echo", "fiber", "chi"}, "Go API framework", 2.0},
		{[]string{"actix", "axum", "rocket"}, "Rust API framework", 2.0},
		{[]string{"spring", "quarkus", "micronaut"}, "Java API framework", 2.0},
	}

	for _, f := range d.files {
		if !strings.HasSuffix(f.Name, ".py") && !strings.HasSuffix(f.Name, ".go") &&
			!strings.HasSuffix(f.Name, ".js") && !strings.HasSuffix(f.Name, ".ts") &&
			!strings.HasSuffix(f.Name, ".rs") && !strings.HasSuffix(f.Name, ".java") {
			continue
		}

		content, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		contentStr := string(content)

		for _, pattern := range apiPatterns {
			for _, imp := range pattern.imports {
				if strings.Contains(contentStr, imp) {
					score += pattern.score
					inds = append(inds, pattern.name+" detected")
					break
				}
			}
		}
	}

	// Check for OpenAPI/Swagger
	apiSpecFiles := []string{"openapi.yaml", "openapi.yml", "openapi.json", "swagger.yaml", "swagger.yml", "swagger.json"}
	for _, file := range apiSpecFiles {
		if _, err := os.Stat(filepath.Join(d.rootPath, file)); err == nil {
			score += 1.5
			inds = append(inds, "OpenAPI/Swagger spec found")
			break
		}
	}

	// Check for routes directory
	for _, dirName := range []string{"routes", "api", "endpoints", "controllers"} {
		if _, err := os.Stat(filepath.Join(d.rootPath, dirName)); err == nil {
			score += 1.0
			inds = append(inds, dirName+"/ directory found")
		}
	}

	scores[ProjectTypeAPI] = score
	indicators[ProjectTypeAPI] = inds
}

// scoreML scores for ML projects
func (d *ProjectTypeDetector) scoreML(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	mlDetector := NewMLDetector(d.rootPath, d.files)
	info := mlDetector.Detect()

	score := 0.0
	var inds []string

	if info.IsMLProject {
		score += 3.0
		inds = append(inds, "ML frameworks detected")
	}

	if len(info.Frameworks) > 0 {
		score += float64(len(info.Frameworks)) * 0.5
		for _, fw := range info.Frameworks {
			inds = append(inds, fw.Name+" detected")
		}
	}

	if len(info.ModelFiles) > 0 {
		score += 1.5
		inds = append(inds, "Model files found")
	}

	if info.NotebookCount > 0 {
		score += float64(info.NotebookCount) * 0.2
		inds = append(inds, "Jupyter notebooks found")
	}

	scores[ProjectTypeML] = score
	indicators[ProjectTypeML] = inds
}

// scoreMonorepo scores for monorepo projects
func (d *ProjectTypeDetector) scoreMonorepo(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	score := 0.0
	var inds []string

	// Check for monorepo tools
	monoIndicators := []struct {
		file  string
		name  string
		score float64
	}{
		{"pnpm-workspace.yaml", "pnpm workspace", 3.0},
		{"lerna.json", "Lerna", 3.0},
		{"turbo.json", "Turborepo", 3.0},
		{"nx.json", "Nx", 3.0},
		{"rush.json", "Rush", 3.0},
		{"packages/", "packages/ directory", 2.0},
		{"apps/", "apps/ directory", 1.5},
	}

	for _, ind := range monoIndicators {
		checkPath := filepath.Join(d.rootPath, ind.file)
		if _, err := os.Stat(checkPath); err == nil {
			score += ind.score
			inds = append(inds, ind.name+" found")
		}
	}

	// Check package.json for workspaces
	pkgPath := filepath.Join(d.rootPath, "package.json")
	if content, err := os.ReadFile(pkgPath); err == nil {
		if strings.Contains(string(content), `"workspaces"`) {
			score += 2.5
			inds = append(inds, "package.json has workspaces")
		}
	}

	// Check Cargo.toml for workspace
	cargoPath := filepath.Join(d.rootPath, "Cargo.toml")
	if content, err := os.ReadFile(cargoPath); err == nil {
		if strings.Contains(string(content), "[workspace]") {
			score += 2.5
			inds = append(inds, "Cargo workspace detected")
		}
	}

	scores[ProjectTypeMonorepo] = score
	indicators[ProjectTypeMonorepo] = inds
}

// scoreFramework scores for framework projects
func (d *ProjectTypeDetector) scoreFramework(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	score := 0.0
	var inds []string

	// Check README for framework indicators
	readmePath := filepath.Join(d.rootPath, "README.md")
	if content, err := os.ReadFile(readmePath); err == nil {
		lower := strings.ToLower(string(content))
		frameworkKeywords := []string{"framework", "toolkit", "sdk", "platform", "engine"}
		for _, kw := range frameworkKeywords {
			if strings.Contains(lower, kw) {
				score += 1.0
				inds = append(inds, "README mentions "+kw)
			}
		}
	}

	// Check for plugin/extension architecture
	pluginDirs := []string{"plugins", "extensions", "addons", "middleware"}
	for _, dir := range pluginDirs {
		if _, err := os.Stat(filepath.Join(d.rootPath, dir)); err == nil {
			score += 1.5
			inds = append(inds, dir+"/ directory found")
		}
	}

	scores[ProjectTypeFramework] = score
	indicators[ProjectTypeFramework] = inds
}

// scoreTutorial scores for tutorial/example projects
func (d *ProjectTypeDetector) scoreTutorial(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	score := 0.0
	var inds []string

	// Check for tutorial indicators
	tutorialPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)tutorial`),
		regexp.MustCompile(`(?i)example`),
		regexp.MustCompile(`(?i)demo`),
		regexp.MustCompile(`(?i)sample`),
		regexp.MustCompile(`(?i)starter`),
		regexp.MustCompile(`(?i)template`),
		regexp.MustCompile(`(?i)boilerplate`),
	}

	// Check directory name
	dirName := filepath.Base(d.rootPath)
	for _, pattern := range tutorialPatterns {
		if pattern.MatchString(dirName) {
			score += 2.0
			inds = append(inds, "Directory name suggests tutorial/example")
			break
		}
	}

	// Check README
	readmePath := filepath.Join(d.rootPath, "README.md")
	if content, err := os.ReadFile(readmePath); err == nil {
		contentStr := string(content)
		for _, pattern := range tutorialPatterns {
			if pattern.MatchString(contentStr) {
				score += 1.0
				inds = append(inds, "README suggests tutorial/example")
				break
			}
		}
	}

	// Check for numbered directories (01-intro, 02-basics, etc.)
	numberedDirPattern := regexp.MustCompile(`^\d{1,2}[-_]`)
	numberedCount := 0
	entries, _ := os.ReadDir(d.rootPath)
	for _, entry := range entries {
		if entry.IsDir() && numberedDirPattern.MatchString(entry.Name()) {
			numberedCount++
		}
	}
	if numberedCount >= 3 {
		score += 2.0
		inds = append(inds, "Has numbered tutorial directories")
	}

	scores[ProjectTypeTutorial] = score
	indicators[ProjectTypeTutorial] = inds
}

// scoreApplication scores for application projects
func (d *ProjectTypeDetector) scoreApplication(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	score := 0.0
	var inds []string

	// Check for application indicators
	appIndicators := []struct {
		file  string
		name  string
		score float64
	}{
		{"Dockerfile", "Dockerfile", 1.5},
		{"docker-compose.yml", "Docker Compose", 1.5},
		{"docker-compose.yaml", "Docker Compose", 1.5},
		{".env.example", "Environment config", 1.0},
		{".env.sample", "Environment config", 1.0},
		{"Procfile", "Heroku Procfile", 1.5},
		{"app.yaml", "App config", 1.0},
		{"vercel.json", "Vercel config", 1.0},
		{"netlify.toml", "Netlify config", 1.0},
	}

	for _, ind := range appIndicators {
		if _, err := os.Stat(filepath.Join(d.rootPath, ind.file)); err == nil {
			score += ind.score
			inds = append(inds, ind.name+" found")
		}
	}

	// Check for UI frameworks (indicates frontend app)
	uiFrameworks := []string{"react", "vue", "angular", "svelte", "next", "nuxt"}
	pkgPath := filepath.Join(d.rootPath, "package.json")
	if content, err := os.ReadFile(pkgPath); err == nil {
		lower := strings.ToLower(string(content))
		for _, fw := range uiFrameworks {
			if strings.Contains(lower, fw) {
				score += 1.0
				inds = append(inds, "Frontend framework detected")
				break
			}
		}
	}

	scores[ProjectTypeApp] = score
	indicators[ProjectTypeApp] = inds
}

// scoreConfig scores for config-only projects (dotfiles, etc.)
func (d *ProjectTypeDetector) scoreConfig(scores map[ProjectType]float64, indicators map[ProjectType][]string) {
	score := 0.0
	var inds []string

	// Count dotfiles
	dotfileCount := 0
	configCount := 0
	for _, f := range d.files {
		if strings.HasPrefix(f.Name, ".") && !strings.HasPrefix(f.Name, ".git") {
			dotfileCount++
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		if ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".toml" || ext == ".ini" {
			configCount++
		}
	}

	totalFiles := len(d.files)
	if totalFiles > 0 {
		dotfileRatio := float64(dotfileCount) / float64(totalFiles)
		if dotfileRatio > 0.5 {
			score += 3.0
			inds = append(inds, "Majority dotfiles")
		}

		configRatio := float64(configCount) / float64(totalFiles)
		if configRatio > 0.7 {
			score += 2.0
			inds = append(inds, "Majority config files")
		}
	}

	// Check for common dotfile repo patterns
	dirName := strings.ToLower(filepath.Base(d.rootPath))
	if strings.Contains(dirName, "dotfile") || dirName == ".config" {
		score += 2.0
		inds = append(inds, "Directory name suggests config repo")
	}

	scores[ProjectTypeConfig] = score
	indicators[ProjectTypeConfig] = inds
}

// getDescription returns a human-readable description for a project type
func (d *ProjectTypeDetector) getDescription(pt ProjectType) string {
	descriptions := map[ProjectType]string{
		ProjectTypeApp:       "A standalone application with deployment configuration",
		ProjectTypeLibrary:   "A reusable library/package meant to be imported by other projects",
		ProjectTypeCLI:       "A command-line interface tool",
		ProjectTypeAPI:       "A backend API or web service",
		ProjectTypeML:        "A machine learning or data science project",
		ProjectTypeDocs:      "A documentation website or knowledge base",
		ProjectTypeMonorepo:  "A monorepo containing multiple packages or applications",
		ProjectTypeFramework: "A framework or toolkit for building applications",
		ProjectTypeTutorial:  "A tutorial, example, or learning project",
		ProjectTypeConfig:    "A configuration or dotfiles repository",
		ProjectTypeUnknown:   "Could not determine project type",
	}

	if desc, ok := descriptions[pt]; ok {
		return desc
	}
	return ""
}
