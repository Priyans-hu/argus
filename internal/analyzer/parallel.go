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

	// Walk the file tree and read dependencies first (required by all detectors)
	p, err := newPipeline(ctx, pa.rootPath, pa.config)
	if err != nil {
		return nil, err
	}

	slog.Debug("file walk complete", "fileCount", len(p.files), "duration", time.Since(startTime))

	analysis := p.newAnalysis()

	// Phase 1: Run essential detectors that others depend on
	// These must complete before parallel phase
	phase1Start := time.Now()
	if err := pa.runPhase1(ctx, p, analysis); err != nil {
		return nil, err
	}
	slog.Debug("phase1 complete", "duration", time.Since(phase1Start))

	// Phase 2: Run remaining detectors in parallel
	phase2Start := time.Now()
	if err := pa.runPhase2(ctx, p, analysis); err != nil {
		return nil, err
	}
	slog.Debug("phase2 complete", "duration", time.Since(phase2Start))

	normalize(analysis)
	slog.Debug("parallel analysis complete", "totalDuration", time.Since(startTime))
	return analysis, nil
}

// runPhase1 runs detectors that must complete before others can start
func (pa *ParallelAnalyzer) runPhase1(ctx context.Context, p *pipeline, analysis *types.Analysis) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// TechStack and Structure can run in parallel with each other
	wg.Add(2)

	// Tech stack detector
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
		}
		techStack, err := p.techStack()
		if err != nil {
			errChan <- err
			return
		}
		analysis.TechStack = techStack
	}()

	// Structure detector
	go func() {
		defer wg.Done()
		select {
		case <-ctx.Done():
			errChan <- ctx.Err()
			return
		default:
		}
		structureDetector := detector.NewStructureDetector(p.root, p.files)
		structure, err := structureDetector.Detect()
		if err != nil {
			errChan <- fmt.Errorf("failed to detect structure: %w", err)
			return
		}
		analysis.Structure = *structure
		analysis.KeyFiles = structureDetector.DetectKeyFiles()
	}()

	// Wait for phase 1 to complete
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// runPhase2 runs all remaining detectors in parallel
func (pa *ParallelAnalyzer) runPhase2(ctx context.Context, p *pipeline, analysis *types.Analysis) error {
	// Check for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	root, files := p.root, p.files
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 10)

	// Commands (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		commands, runners := p.commands()
		mu.Lock()
		analysis.Commands, analysis.TestRunners = commands, runners
		mu.Unlock()
	}()

	// Conventions, patterns and framework patterns, in that order
	wg.Add(1)
	go func() {
		defer wg.Done()
		conventions, err := p.conventions()
		if err != nil {
			errChan <- err
			return
		}
		mu.Lock()
		analysis.Conventions = conventions
		mu.Unlock()
	}()

	// Endpoints (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		endpointDetector := detector.NewEndpointDetector(root, files)
		endpoints, err := endpointDetector.Detect()
		if err != nil {
			errChan <- fmt.Errorf("failed to detect endpoints: %w", err)
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
		readmeDetector := detector.NewReadmeDetector(root)
		content := readmeDetector.Detect()
		mu.Lock()
		analysis.ReadmeContent = content
		mu.Unlock()
	}()

	// Monorepo (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		monorepoDetector := detector.NewMonorepoDetector(root, files)
		info := monorepoDetector.Detect()
		mu.Lock()
		analysis.MonorepoInfo = info
		mu.Unlock()
	}()

	// Code patterns: keyword, ML and AST detectors
	wg.Add(1)
	go func() {
		defer wg.Done()
		patterns := p.codePatterns()
		mu.Lock()
		analysis.CodePatterns = patterns
		mu.Unlock()
	}()

	// Git conventions (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		gitDetector := detector.NewGitDetectorGoGit(root)
		conventions := gitDetector.Detect()
		mu.Lock()
		analysis.GitConventions = conventions
		mu.Unlock()
	}()

	// Architecture (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		archDetector := detector.NewArchitectureDetector(root, files)
		info := archDetector.Detect()
		mu.Lock()
		analysis.ArchitectureInfo = info
		mu.Unlock()
	}()

	// Development info (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		devDetector := detector.NewDevelopmentDetector(root, files)
		info := devDetector.Detect()
		mu.Lock()
		analysis.DevelopmentInfo = info
		mu.Unlock()
	}()

	// Config files (no dependencies)
	wg.Add(1)
	go func() {
		defer wg.Done()
		configDetector := detector.NewConfigDetector(root, files)
		configs := configDetector.Detect()
		mu.Lock()
		analysis.ConfigFiles = configs
		mu.Unlock()
	}()

	// Wait for all phase 2 detectors
	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	// Phase 3: CLI info needs the tech stack, project tools need CLI info and the README
	cliDetector := detector.NewCLIDetector(root, files, &analysis.TechStack)
	analysis.CLIInfo = cliDetector.Detect()
	analysis.ProjectTools = p.projectTools(analysis)

	return nil
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
