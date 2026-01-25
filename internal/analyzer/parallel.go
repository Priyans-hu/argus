package analyzer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
)

// ParallelAnalyzer performs analysis using concurrent detector execution
type ParallelAnalyzer struct {
	rootPath string
	config   *types.Config
}

// NewParallelAnalyzer creates a new parallel analyzer
func NewParallelAnalyzer(rootPath string, config *types.Config) *ParallelAnalyzer {
	return &ParallelAnalyzer{
		rootPath: rootPath,
		config:   config,
	}
}

// detectorResult holds the result of a detector execution
type detectorResult struct {
	name string
	err  error
}

// Analyze performs parallel codebase analysis
func (pa *ParallelAnalyzer) Analyze(ctx context.Context) (*types.Analysis, error) {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	startTime := time.Now()
	slog.Debug("starting parallel analysis", "rootPath", pa.rootPath)

	// Get absolute path and walk file tree first (required by all detectors)
	walker := NewWalker(pa.rootPath)
	files, err := walker.Walk(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	slog.Debug("file walk complete", "fileCount", len(files), "duration", time.Since(startTime))

	// Initialize analysis
	analysis := &types.Analysis{
		ProjectName: getProjectName(pa.rootPath),
		RootPath:    pa.rootPath,
	}

	// Phase 1: Run essential detectors that others depend on
	// These must complete before parallel phase
	phase1Start := time.Now()
	if err := pa.runPhase1(ctx, files, analysis); err != nil {
		return nil, err
	}
	slog.Debug("phase1 complete", "duration", time.Since(phase1Start))

	// Phase 2: Run remaining detectors in parallel
	phase2Start := time.Now()
	if err := pa.runPhase2(ctx, files, analysis); err != nil {
		return nil, err
	}
	slog.Debug("phase2 complete", "duration", time.Since(phase2Start))

	slog.Debug("parallel analysis complete", "totalDuration", time.Since(startTime))
	return analysis, nil
}

// runPhase1 runs detectors that must complete before others can start
func (pa *ParallelAnalyzer) runPhase1(ctx context.Context, files []types.FileInfo, analysis *types.Analysis) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var wg sync.WaitGroup
	errChan := make(chan detectorResult, 2)

	// TechStack and Structure can run in parallel with each other
	wg.Add(2)

	// Tech stack detector
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			errChan <- detectorResult{"techstack", ctx.Err()}
			return
		default:
		}
		techDetector := detector.NewTechStackDetector(pa.rootPath, files)
		techStack, err := techDetector.Detect()
		if err != nil {
			errChan <- detectorResult{"techstack", err}
			return
		}
		analysis.TechStack = *techStack
	}()

	// Structure detector
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			errChan <- detectorResult{"structure", ctx.Err()}
			return
		default:
		}
		structureDetector := detector.NewStructureDetector(pa.rootPath, files)
		structure, err := structureDetector.Detect()
		if err != nil {
			errChan <- detectorResult{"structure", err}
			return
		}
		analysis.Structure = *structure
		analysis.KeyFiles = structureDetector.DetectKeyFiles()
	}()

	// Wait for phase 1 to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for result := range errChan {
		if result.err != nil {
			return fmt.Errorf("failed to detect %s: %w", result.name, result.err)
		}
	}

	return nil
}

// runPhase2 runs all remaining detectors in parallel
func (pa *ParallelAnalyzer) runPhase2(ctx context.Context, files []types.FileInfo, analysis *types.Analysis) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan detectorResult, 12)

	// Commands (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		commands := detector.DetectCommands(pa.rootPath)

		// Add pyproject.toml commands (Python)
		pyprojectDetector := detector.NewPyProjectDetector(pa.rootPath)
		if pyInfo := pyprojectDetector.Detect(); pyInfo != nil && pyInfo.HasPyProject {
			commands = append(commands, detectPyProjectCommands(pyInfo)...)
		}

		// Add Cargo.toml commands (Rust)
		cargoDetector := detector.NewCargoDetector(pa.rootPath)
		if cargoInfo := cargoDetector.Detect(); cargoInfo != nil && cargoInfo.HasCargo {
			commands = filterNonCargoCommands(commands)
			commands = append(commands, cargoDetector.DetectCargoCommands()...)
		}

		mu.Lock()
		analysis.Commands = commands
		mu.Unlock()
	}()

	// Dependencies (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		a := &Analyzer{rootPath: pa.rootPath}
		deps := a.detectDependencies(pa.rootPath)
		mu.Lock()
		analysis.Dependencies = deps
		mu.Unlock()
	}()

	// Conventions (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		conventionDetector := detector.NewConventionDetector(pa.rootPath, files)
		conventions, err := conventionDetector.Detect()
		if err != nil {
			errChan <- detectorResult{"conventions", err}
			return
		}
		mu.Lock()
		analysis.Conventions = conventions
		mu.Unlock()
	}()

	// Patterns (appends to conventions)
	wg.Add(1)
	go func() {
		defer wg.Done()
		patternDetector := detector.NewPatternDetector(pa.rootPath, files)
		patterns, err := patternDetector.Detect()
		if err != nil {
			errChan <- detectorResult{"patterns", err}
			return
		}
		mu.Lock()
		analysis.Conventions = append(analysis.Conventions, patterns...)
		mu.Unlock()
	}()

	// Framework patterns (appends to conventions)
	wg.Add(1)
	go func() {
		defer wg.Done()
		frameworkDetector := detector.NewFrameworkDetector(pa.rootPath, files)
		frameworkPatterns, err := frameworkDetector.Detect()
		if err != nil {
			errChan <- detectorResult{"framework", err}
			return
		}
		mu.Lock()
		analysis.Conventions = append(analysis.Conventions, frameworkPatterns...)
		mu.Unlock()
	}()

	// Endpoints (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		endpointDetector := detector.NewEndpointDetector(pa.rootPath, files)
		endpoints, err := endpointDetector.Detect()
		if err != nil {
			errChan <- detectorResult{"endpoints", err}
			return
		}
		mu.Lock()
		analysis.Endpoints = endpoints
		mu.Unlock()
	}()

	// README (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		readmeDetector := detector.NewReadmeDetector(pa.rootPath)
		content := readmeDetector.Detect()
		mu.Lock()
		analysis.ReadmeContent = content
		mu.Unlock()
	}()

	// Monorepo (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		monorepoDetector := detector.NewMonorepoDetector(pa.rootPath, files)
		info := monorepoDetector.Detect()
		mu.Lock()
		analysis.MonorepoInfo = info
		mu.Unlock()
	}()

	// Code patterns (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		codePatternDetector := detector.NewCodePatternDetector(pa.rootPath, files)
		patterns := codePatternDetector.Detect()

		// Add ML-specific patterns
		mlDetector := detector.NewMLDetector(pa.rootPath, files)
		if mlPatterns := mlDetector.GetMLPatterns(); len(mlPatterns) > 0 {
			patterns.MLPatterns = mlPatterns
		}

		// Add AST-based JavaScript/TypeScript patterns
		jsASTDetector := detector.NewJSASTDetector(pa.rootPath, files)
		if jsPatterns := jsASTDetector.Detect(); len(jsPatterns) > 0 {
			patterns.DataFetching = mergePatterns(patterns.DataFetching, filterByCategory(jsPatterns, "JavaScript Frameworks"))
			patterns.StateManagement = mergePatterns(patterns.StateManagement, filterByCategory(jsPatterns, "React Hooks"))
		}

		// Add AST-based Python patterns
		pyASTDetector := detector.NewPythonASTDetector(pa.rootPath, files)
		if pyPatterns := pyASTDetector.Detect(); len(pyPatterns) > 0 {
			patterns.PythonPatterns = mergePatterns(patterns.PythonPatterns, pyPatterns)
		}

		mu.Lock()
		analysis.CodePatterns = patterns
		mu.Unlock()
	}()

	// Git conventions (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		gitDetector := detector.NewGitDetectorGoGit(pa.rootPath)
		conventions := gitDetector.Detect()
		mu.Lock()
		analysis.GitConventions = conventions
		mu.Unlock()
	}()

	// Architecture (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		archDetector := detector.NewArchitectureDetector(pa.rootPath, files)
		info := archDetector.Detect()
		mu.Lock()
		analysis.ArchitectureInfo = info
		mu.Unlock()
	}()

	// Development info (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		devDetector := detector.NewDevelopmentDetector(pa.rootPath, files)
		info := devDetector.Detect()
		mu.Lock()
		analysis.DevelopmentInfo = info
		mu.Unlock()
	}()

	// Config files (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		configDetector := detector.NewConfigDetector(pa.rootPath, files)
		configs := configDetector.Detect()
		mu.Lock()
		analysis.ConfigFiles = configs
		mu.Unlock()
	}()

	// Wait for all phase 2 detectors
	wg.Wait()
	close(errChan)

	// Check for errors
	for result := range errChan {
		if result.err != nil {
			return fmt.Errorf("failed to detect %s: %w", result.name, result.err)
		}
	}

	// Phase 3: CLI detector depends on TechStack (which is now available)
	cliDetector := detector.NewCLIDetector(pa.rootPath, files, &analysis.TechStack)
	analysis.CLIInfo = cliDetector.Detect()

	return nil
}

// getProjectName extracts project name from path
func getProjectName(rootPath string) string {
	// Use filepath.Base for simple extraction
	return baseFileName(rootPath)
}

// baseFileName returns the last element of path
func baseFileName(path string) string {
	// Find the last separator
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}

// mergePatterns merges two PatternInfo slices, avoiding duplicates by name
func mergePatterns(existing, new []types.PatternInfo) []types.PatternInfo {
	if len(new) == 0 {
		return existing
	}
	if len(existing) == 0 {
		return new
	}

	// Build set of existing pattern names
	seen := make(map[string]bool)
	for _, p := range existing {
		seen[p.Name] = true
	}

	// Add new patterns that don't exist
	result := existing
	for _, p := range new {
		if !seen[p.Name] {
			result = append(result, p)
			seen[p.Name] = true
		}
	}
	return result
}

// filterByCategory filters PatternInfo slice by category name
func filterByCategory(patterns []types.PatternInfo, category string) []types.PatternInfo {
	var result []types.PatternInfo
	for _, p := range patterns {
		if p.Category == category {
			result = append(result, p)
		}
	}
	return result
}
