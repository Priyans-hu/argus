package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"slices"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
)

// Analyzer coordinates the codebase analysis
type Analyzer struct {
	rootPath string
	config   *types.Config
	rootDeps []types.Dependency // a monorepo root's, for a workspace
}

// NewAnalyzer creates a new analyzer
func NewAnalyzer(rootPath string, config *types.Config) *Analyzer {
	return &Analyzer{
		rootPath: rootPath,
		config:   config,
	}
}

// Analyze performs the full codebase analysis
func (a *Analyzer) Analyze(ctx context.Context) (*types.Analysis, error) {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	slog.Debug("starting analysis", "rootPath", a.rootPath)
	p, err := newPipeline(ctx, a.rootPath, a.config, a.rootDeps)
	if err != nil {
		return nil, err
	}
	slog.Debug("file walk complete", "fileCount", len(p.files))
	root, files := p.root, p.files
	analysis := p.newAnalysis()

	// Detect tech stack
	if analysis.TechStack, err = p.techStack(); err != nil {
		return nil, err
	}

	// Detect structure
	structureDetector := detector.NewStructureDetector(root, files)
	structure, err := structureDetector.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect structure: %w", err)
	}
	analysis.Structure = *structure

	// Detect key files
	analysis.KeyFiles = structureDetector.DetectKeyFiles()

	analysis.Commands = p.commands()

	// Conventions, patterns (branch naming, comments, logging, error
	// handling, architecture) and framework patterns
	if analysis.Conventions, err = p.conventions(); err != nil {
		return nil, err
	}

	// Detect API endpoints
	endpointDetector := detector.NewEndpointDetector(root, files)
	endpoints, err := endpointDetector.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect endpoints: %w", err)
	}
	analysis.Endpoints = endpoints

	// Parse README for project overview
	readmeDetector := detector.NewReadmeDetector(root)
	analysis.ReadmeContent = readmeDetector.Detect()

	// Detect monorepo structure
	monorepoDetector := detector.NewMonorepoDetector(root, files)
	analysis.MonorepoInfo = monorepoDetector.Detect()

	analysis.CodePatterns = p.codePatterns()

	// Detect git conventions (commit messages, branch naming) - using go-git library
	gitDetector := detector.NewGitDetectorGoGit(root)
	analysis.GitConventions = gitDetector.Detect()

	// Detect architecture patterns
	archDetector := detector.NewArchitectureDetector(root, files)
	analysis.ArchitectureInfo = archDetector.Detect()

	// Detect development setup info
	devDetector := detector.NewDevelopmentDetector(root, files)
	analysis.DevelopmentInfo = devDetector.Detect()

	// Detect config files
	configDetector := detector.NewConfigDetector(root, files)
	analysis.ConfigFiles = configDetector.Detect()

	// Detect CLI info (if applicable)
	cliDetector := detector.NewCLIDetector(root, files, &analysis.TechStack)
	analysis.CLIInfo = cliDetector.Detect()

	analysis.ProjectTools = p.projectTools(analysis)

	normalize(analysis)
	return analysis, nil
}

// detectPyProjectCommands extracts commands from pyproject.toml
func detectPyProjectCommands(info *detector.PyProjectInfo) []types.Command {
	var commands []types.Command

	// Add script commands
	for _, name := range slices.Sorted(maps.Keys(info.Scripts)) {
		commands = append(commands, types.Command{
			Name:        name,
			Command:     info.Scripts[name],
			Description: "Script from pyproject.toml",
		})
	}

	// Add tool-specific commands
	for _, tool := range info.Tools {
		switch tool {
		case "pytest":
			commands = append(commands, types.Command{
				Name:        "pytest",
				Description: "Run tests with pytest",
			})
			commands = append(commands, types.Command{
				Name:        "pytest -v",
				Description: "Run tests with verbose output",
			})
		case "black":
			commands = append(commands, types.Command{
				Name:        "black .",
				Description: "Format code with Black",
			})
		case "ruff":
			commands = append(commands, types.Command{
				Name:        "ruff check .",
				Description: "Lint code with Ruff",
			})
			commands = append(commands, types.Command{
				Name:        "ruff format .",
				Description: "Format code with Ruff",
			})
		case "mypy":
			commands = append(commands, types.Command{
				Name:        "mypy .",
				Description: "Type check with mypy",
			})
		case "poetry":
			commands = append(commands, types.Command{
				Name:        "poetry install",
				Description: "Install dependencies with Poetry",
			})
			commands = append(commands, types.Command{
				Name:        "poetry shell",
				Description: "Activate Poetry virtual environment",
			})
		case "pdm":
			commands = append(commands, types.Command{
				Name:        "pdm install",
				Description: "Install dependencies with PDM",
			})
		case "hatch":
			commands = append(commands, types.Command{
				Name:        "hatch run",
				Description: "Run commands in Hatch environment",
			})
		case "coverage":
			commands = append(commands, types.Command{
				Name:        "coverage run -m pytest",
				Description: "Run tests with coverage",
			})
			commands = append(commands, types.Command{
				Name:        "coverage report",
				Description: "Show coverage report",
			})
		}
	}

	return commands
}

// filterNonCargoCommands removes basic cargo commands to be replaced with enhanced ones
func filterNonCargoCommands(commands []types.Command) []types.Command {
	var filtered []types.Command
	for _, cmd := range commands {
		// Keep non-cargo commands
		if !isBasicCargoCommand(cmd.Name) {
			filtered = append(filtered, cmd)
		}
	}
	return filtered
}

// isBasicCargoCommand checks if a command is a basic cargo command
func isBasicCargoCommand(name string) bool {
	basicCargo := []string{
		"cargo build",
		"cargo build --release",
		"cargo test",
		"cargo fmt",
		"cargo clippy",
	}
	for _, basic := range basicCargo {
		if name == basic {
			return true
		}
	}
	return false
}
