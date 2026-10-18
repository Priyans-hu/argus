package analyzer

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
)

// Impact categories for file changes
const (
	ImpactTechStack   = "techstack"
	ImpactStructure   = "structure"
	ImpactCommands    = "commands"
	ImpactConventions = "conventions"
	ImpactEndpoints   = "endpoints"
	ImpactConfig      = "config"
	ImpactDevelopment = "development"
	ImpactReadme      = "readme"
	ImpactGit         = "git"
	ImpactAll         = "all"
)

// IncrementalAnalyzer performs incremental analysis by only running affected detectors
type IncrementalAnalyzer struct {
	rootPath    string
	config      *types.Config
	cache       *types.Analysis
	cacheMu     sync.RWMutex
	lastUpdated time.Time
}

// NewIncrementalAnalyzer creates a new incremental analyzer; cfg may be nil
func NewIncrementalAnalyzer(rootPath string, cfg *types.Config) *IncrementalAnalyzer {
	absPath, _ := filepath.Abs(rootPath)
	return &IncrementalAnalyzer{rootPath: absPath, config: cfg}
}

// AnalyzeFull performs a full analysis, the same as scan's, and caches the result
func (ia *IncrementalAnalyzer) AnalyzeFull(ctx context.Context) (*types.Analysis, error) {
	analysis, err := NewParallelAnalyzer(ia.rootPath, ia.config).Analyze(ctx)
	if err != nil {
		return nil, err
	}

	ia.cacheMu.Lock()
	ia.cache = analysis
	ia.lastUpdated = time.Now()
	ia.cacheMu.Unlock()

	return analysis, nil
}

// AnalyzeIncremental performs incremental analysis based on changed file
func (ia *IncrementalAnalyzer) AnalyzeIncremental(ctx context.Context, changedFile string) (*types.Analysis, []string, error) {
	// If no cache, do full analysis
	ia.cacheMu.RLock()
	if ia.cache == nil {
		ia.cacheMu.RUnlock()
		analysis, err := ia.AnalyzeFull(ctx)
		return analysis, []string{ImpactAll}, err
	}
	ia.cacheMu.RUnlock()

	// Determine impact
	impacts := DetermineImpact(changedFile)

	// If all impacts, do full analysis
	for _, impact := range impacts {
		if impact == ImpactAll {
			analysis, err := ia.AnalyzeFull(ctx)
			return analysis, impacts, err
		}
	}

	// Fresh file list and dependencies
	p, err := newPipeline(ctx, ia.rootPath, ia.config)
	if err != nil {
		return nil, nil, err
	}

	// Clone cached analysis
	ia.cacheMu.Lock()
	analysis := ia.cloneAnalysis(ia.cache)
	ia.cacheMu.Unlock()

	// Run only affected detectors
	for _, impact := range impacts {
		if err := ia.runDetector(impact, p, analysis); err != nil {
			// On error, fall back to full analysis
			fullAnalysis, fullErr := ia.AnalyzeFull(ctx)
			return fullAnalysis, []string{ImpactAll}, fullErr
		}
	}
	analysis.ProjectTools = p.projectTools(analysis)
	normalize(analysis)

	// Update cache
	ia.cacheMu.Lock()
	ia.cache = analysis
	ia.lastUpdated = time.Now()
	ia.cacheMu.Unlock()

	return analysis, impacts, nil
}

// GetCache returns the cached analysis (thread-safe)
func (ia *IncrementalAnalyzer) GetCache() *types.Analysis {
	ia.cacheMu.RLock()
	defer ia.cacheMu.RUnlock()
	return ia.cache
}

// LastUpdated returns when the cache was last updated
func (ia *IncrementalAnalyzer) LastUpdated() time.Time {
	ia.cacheMu.RLock()
	defer ia.cacheMu.RUnlock()
	return ia.lastUpdated
}

// DetermineImpact determines which detectors need to run for a file change
func DetermineImpact(changedFile string) []string {
	name := filepath.Base(changedFile)
	ext := strings.ToLower(filepath.Ext(changedFile))
	dir := filepath.Dir(changedFile)

	// Config file changes - full reload
	if name == ".argus.yaml" {
		return []string{ImpactAll}
	}

	// Dependencies are evidence for the tech stack and the pattern detectors
	depFiles := map[string]bool{
		"go.mod": true, "go.sum": true,
		"package.json": true, "package-lock.json": true,
		"yarn.lock": true, "pnpm-lock.yaml": true,
		"Cargo.toml": true, "Cargo.lock": true,
		"pyproject.toml": true, "requirements.txt": true,
		"Pipfile": true, "Pipfile.lock": true,
		"pom.xml": true, "build.gradle": true,
		"Gemfile": true, "Gemfile.lock": true,
	}
	if depFiles[name] || (strings.HasPrefix(name, "requirements") && ext == ".txt") {
		return []string{ImpactAll}
	}

	// Makefile changes
	if name == "Makefile" || name == "makefile" || name == "GNUmakefile" {
		return []string{ImpactCommands, ImpactDevelopment}
	}

	// README changes
	if strings.EqualFold(name, "README.md") || strings.EqualFold(name, "README") {
		return []string{ImpactReadme}
	}

	// Git hooks
	if strings.Contains(dir, ".githooks") || strings.Contains(dir, ".husky") {
		return []string{ImpactDevelopment}
	}
	if name == "lefthook.yml" || name == ".lefthook.yml" || name == ".pre-commit-config.yaml" {
		return []string{ImpactDevelopment}
	}

	// GitHub workflows and config
	if strings.Contains(dir, ".github") {
		return []string{ImpactConfig}
	}

	// Config files
	configFiles := map[string]bool{
		".golangci.yml": true, ".golangci.yaml": true,
		".eslintrc.json": true, ".eslintrc.js": true, ".eslintrc.yml": true,
		"eslint.config.js": true, "eslint.config.mjs": true,
		".prettierrc": true, ".prettierrc.json": true,
		"tsconfig.json": true, "jest.config.js": true, "jest.config.ts": true,
		"vite.config.ts": true, "vite.config.js": true,
		"Dockerfile": true, "docker-compose.yml": true, "docker-compose.yaml": true,
		".goreleaser.yml": true, ".goreleaser.yaml": true,
		".env.example": true, ".editorconfig": true,
		".nvmrc": true, ".python-version": true, ".tool-versions": true,
	}
	if configFiles[name] {
		return []string{ImpactConfig, ImpactDevelopment}
	}

	// Source code changes - affects patterns and endpoints
	sourceExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".py": true, ".java": true, ".kt": true, ".rs": true, ".rb": true,
		".cs": true, ".cpp": true, ".c": true, ".h": true, ".hpp": true,
		".swift": true, ".php": true, ".vue": true, ".svelte": true,
	}
	if sourceExts[ext] {
		return []string{ImpactConventions, ImpactEndpoints}
	}

	// Directory structure changes (new/deleted directories)
	// This is harder to detect from file path alone, so we check extension
	if ext == "" {
		// Might be a directory change
		return []string{ImpactStructure}
	}

	// Default: minimal impact
	return []string{}
}

// runDetector runs a specific detector and updates the analysis
func (ia *IncrementalAnalyzer) runDetector(impact string, p *pipeline, analysis *types.Analysis) error {
	root, files := p.root, p.files
	switch impact {
	case ImpactTechStack:
		techStack, err := p.techStack()
		if err != nil {
			return err
		}
		analysis.TechStack = techStack
		analysis.Dependencies = p.deps

	case ImpactStructure:
		structureDetector := detector.NewStructureDetector(root, files)
		structure, err := structureDetector.Detect()
		if err != nil {
			return err
		}
		analysis.Structure = *structure
		analysis.KeyFiles = structureDetector.DetectKeyFiles()

		// Also update monorepo info
		monorepoDetector := detector.NewMonorepoDetector(root, files)
		analysis.MonorepoInfo = monorepoDetector.Detect()

		// And architecture
		archDetector := detector.NewArchitectureDetector(root, files)
		analysis.ArchitectureInfo = archDetector.Detect()

	case ImpactCommands:
		analysis.Commands = p.commands()

	case ImpactConventions:
		conventions, err := p.conventions()
		if err != nil {
			return err
		}
		analysis.Conventions = conventions
		analysis.CodePatterns = p.codePatterns()

	case ImpactEndpoints:
		endpointDetector := detector.NewEndpointDetector(root, files)
		endpoints, err := endpointDetector.Detect()
		if err != nil {
			return err
		}
		analysis.Endpoints = endpoints

	case ImpactConfig:
		configDetector := detector.NewConfigDetector(root, files)
		analysis.ConfigFiles = configDetector.Detect()

	case ImpactDevelopment:
		devDetector := detector.NewDevelopmentDetector(root, files)
		analysis.DevelopmentInfo = devDetector.Detect()

		// CLI info depends on tech stack
		cliDetector := detector.NewCLIDetector(root, files, &analysis.TechStack)
		analysis.CLIInfo = cliDetector.Detect()

	case ImpactReadme:
		readmeDetector := detector.NewReadmeDetector(root)
		analysis.ReadmeContent = readmeDetector.Detect()

	case ImpactGit:
		gitDetector := detector.NewGitDetectorGoGit(root)
		analysis.GitConventions = gitDetector.Detect()
	}

	return nil
}

// cloneAnalysis creates a deep copy of the analysis
func (ia *IncrementalAnalyzer) cloneAnalysis(src *types.Analysis) *types.Analysis {
	if src == nil {
		return nil
	}

	// Create new analysis with same basic fields
	dst := &types.Analysis{
		ProjectName:  src.ProjectName,
		RootPath:     src.RootPath,
		TechStack:    src.TechStack,
		Structure:    src.Structure,
		MonorepoInfo: src.MonorepoInfo,
	}

	// Copy slices
	if src.KeyFiles != nil {
		dst.KeyFiles = make([]types.KeyFile, len(src.KeyFiles))
		copy(dst.KeyFiles, src.KeyFiles)
	}
	if src.Commands != nil {
		dst.Commands = make([]types.Command, len(src.Commands))
		copy(dst.Commands, src.Commands)
	}
	if src.Dependencies != nil {
		dst.Dependencies = make([]types.Dependency, len(src.Dependencies))
		copy(dst.Dependencies, src.Dependencies)
	}
	if src.Conventions != nil {
		dst.Conventions = make([]types.Convention, len(src.Conventions))
		copy(dst.Conventions, src.Conventions)
	}
	if src.Endpoints != nil {
		dst.Endpoints = make([]types.Endpoint, len(src.Endpoints))
		copy(dst.Endpoints, src.Endpoints)
	}
	if src.ConfigFiles != nil {
		dst.ConfigFiles = make([]types.ConfigFileInfo, len(src.ConfigFiles))
		copy(dst.ConfigFiles, src.ConfigFiles)
	}
	dst.ProjectTools = slices.Clone(src.ProjectTools)

	// Copy pointers (shallow copy is OK since we replace whole objects)
	dst.ReadmeContent = src.ReadmeContent
	dst.CodePatterns = src.CodePatterns
	dst.GitConventions = src.GitConventions
	dst.ArchitectureInfo = src.ArchitectureInfo
	dst.DevelopmentInfo = src.DevelopmentInfo
	dst.CLIInfo = src.CLIInfo

	return dst
}

// ImpactDescription returns a human-readable description of impact categories
func ImpactDescription(impacts []string) string {
	if len(impacts) == 0 {
		return "no sections"
	}

	descriptions := make([]string, 0, len(impacts))
	for _, impact := range impacts {
		switch impact {
		case ImpactAll:
			return "all sections"
		case ImpactTechStack:
			descriptions = append(descriptions, "tech stack")
		case ImpactStructure:
			descriptions = append(descriptions, "structure")
		case ImpactCommands:
			descriptions = append(descriptions, "commands")
		case ImpactConventions:
			descriptions = append(descriptions, "conventions")
		case ImpactEndpoints:
			descriptions = append(descriptions, "endpoints")
		case ImpactConfig:
			descriptions = append(descriptions, "config")
		case ImpactDevelopment:
			descriptions = append(descriptions, "development")
		case ImpactReadme:
			descriptions = append(descriptions, "readme")
		case ImpactGit:
			descriptions = append(descriptions, "git")
		}
	}

	return strings.Join(descriptions, ", ")
}
