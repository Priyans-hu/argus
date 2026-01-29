package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Priyans-hu/argus/internal/analyzer"
	"github.com/Priyans-hu/argus/internal/config"
	"github.com/Priyans-hu/argus/internal/generator"
	"github.com/Priyans-hu/argus/internal/merger"
	"github.com/Priyans-hu/argus/internal/usage"
	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var version = "dev"

// Flags
var (
	outputDir         string
	outputFormat      string
	dryRun            bool
	verbose           bool
	force             bool
	mergeMode         bool
	addCustomBlock    bool
	compactMode       bool
	parallel          bool
	usageMode         bool
	monorepoMode      bool
	insightsSince     string
	insightsFormat    string
	insightsSubagents bool
)

var rootCmd = &cobra.Command{
	Use:   "argus",
	Short: "Help AI grok your codebase",
	Long: `Argus - The All-Seeing Code Analyzer

Argus scans your codebase and generates optimized context files
for AI coding assistants (Claude Code, Cursor, Copilot, etc.).

No more manually writing CLAUDE.md or .cursorrules - Argus sees everything.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupLogging()
	},
}

// setupLogging configures slog based on verbose flag
func setupLogging() {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize Argus configuration",
	Long: `Initialize Argus in the specified directory (or current directory).
Creates a .argus.yaml configuration file with sensible defaults.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan codebase and generate context files",
	Long: `Scan the specified directory (or current directory if not specified)
and generate AI context files like CLAUDE.md, .cursorrules, etc.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

var syncCmd = &cobra.Command{
	Use:   "sync [path]",
	Short: "Regenerate context files using existing config",
	Long: `Regenerate context files based on .argus.yaml configuration.
Uses the output formats specified in the config file.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSync,
}

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Watch for changes and regenerate context files",
	Long: `Watch the specified directory (or current directory if not specified)
for file changes and automatically regenerate AI context files.

Press Ctrl+C to stop watching.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runWatch,
}

var insightsCmd = &cobra.Command{
	Use:   "insights [path]",
	Short: "Analyze Claude Code usage for a project",
	Long: `Analyze Claude Code JSONL session logs to show tool usage, file interaction
patterns, token consumption, cost estimates, and AI pain points.

Data is read locally from ~/.claude/projects/ and never sent anywhere.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInsights,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("argus version %s\n", version)
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade argus to the latest version",
	Long: `Check for and install the latest version of argus from GitHub releases.

This command will:
1. Check for the latest release on GitHub
2. Download the appropriate binary for your OS/architecture
3. Replace the current binary with the new version`,
	RunE: runUpgrade,
}

func init() {
	// Init command flags
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing config file")

	// Scan command flags
	scanCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated files")
	scanCmd.Flags().StringVarP(&outputFormat, "format", "f", "claude", "Output format: claude, claude-code, cursor, copilot, continue, all")
	scanCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be generated without writing files")
	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	scanCmd.Flags().BoolVarP(&mergeMode, "merge", "m", true, "Preserve custom sections when regenerating (default: true)")
	scanCmd.Flags().BoolVar(&addCustomBlock, "add-custom", false, "Add a custom section placeholder to output")
	scanCmd.Flags().BoolVarP(&compactMode, "compact", "c", false, "Generate compact output (~45% smaller, optimized for token efficiency)")
	scanCmd.Flags().BoolVarP(&parallel, "parallel", "p", true, "Run detectors in parallel for faster analysis (default: true)")
	scanCmd.Flags().BoolVar(&usageMode, "usage", false, "Include AI usage insights from Claude Code session logs")
	scanCmd.Flags().BoolVar(&monorepoMode, "monorepo", false, "Generate output per workspace in monorepo projects")

	// Sync command flags
	syncCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be generated without writing files")
	syncCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	syncCmd.Flags().BoolVarP(&mergeMode, "merge", "m", true, "Preserve custom sections when regenerating (default: true)")
	syncCmd.Flags().BoolVar(&addCustomBlock, "add-custom", false, "Add a custom section placeholder to output")
	syncCmd.Flags().BoolVarP(&compactMode, "compact", "c", false, "Generate compact output (~45% smaller, optimized for token efficiency)")
	syncCmd.Flags().BoolVarP(&parallel, "parallel", "p", true, "Run detectors in parallel for faster analysis (default: true)")
	syncCmd.Flags().BoolVar(&usageMode, "usage", false, "Include AI usage insights from Claude Code session logs")

	// Insights command flags
	insightsCmd.Flags().StringVarP(&insightsSince, "since", "s", "", "Date filter (e.g., 7d, 30d, 2025-01-01)")
	insightsCmd.Flags().StringVarP(&insightsFormat, "format", "f", "text", "Output format: text, json")
	insightsCmd.Flags().BoolVar(&insightsSubagents, "subagents", true, "Include subagent data")

	// Watch command flags
	watchCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	watchCmd.Flags().BoolVarP(&mergeMode, "merge", "m", true, "Preserve custom sections when regenerating (default: true)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(insightsCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(upgradeCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists and is a directory
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Check if config already exists
	configPath := filepath.Join(absPath, config.ConfigFileName)
	if config.Exists(absPath) && !force {
		return fmt.Errorf("config file already exists: %s\nUse --force to overwrite", configPath)
	}

	// Write config file with comments
	configContent := config.ConfigWithComments()
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	fmt.Printf("✅ Created %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Edit .argus.yaml to customize output formats and conventions")
	fmt.Println("  2. Run 'argus scan' to generate context files")
	fmt.Println("  3. Run 'argus sync' anytime to regenerate with your config")

	return nil
}

func runScan(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Load config if exists
	cfg, err := config.Load(absPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override format from flag if specified
	formats := cfg.Output
	if cmd.Flags().Changed("format") {
		if outputFormat == "all" {
			formats = []string{"claude", "claude-code", "cursor", "copilot", "continue"}
		} else {
			formats = []string{outputFormat}
		}
	}

	fmt.Printf("🔍 Scanning %s...\n", absPath)
	if parallel && verbose {
		fmt.Println("   Using parallel detector execution...")
	}

	// Create context with cancellation support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals for graceful cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Run analysis (parallel or sequential)
	var analysis *types.Analysis
	if parallel {
		pa := analyzer.NewParallelAnalyzer(absPath, nil)
		analysis, err = pa.Analyze(ctx)
	} else {
		a := analyzer.NewAnalyzer(absPath, nil)
		analysis, err = a.Analyze(ctx)
	}
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Add custom conventions from config
	for _, conv := range cfg.CustomConventions {
		analysis.Conventions = append(analysis.Conventions, types.Convention{
			Category:    "custom",
			Description: conv,
		})
	}

	if verbose {
		fmt.Printf("\n📊 Analysis Results:\n")
		fmt.Printf("   Project: %s\n", analysis.ProjectName)
		fmt.Printf("   Languages: %d\n", len(analysis.TechStack.Languages))
		fmt.Printf("   Frameworks: %d\n", len(analysis.TechStack.Frameworks))
		fmt.Printf("   Directories: %d\n", len(analysis.Structure.Directories))
		fmt.Printf("   Key Files: %d\n", len(analysis.KeyFiles))
		fmt.Printf("   Commands: %d\n", len(analysis.Commands))
		fmt.Printf("   Conventions: %d\n", len(analysis.Conventions))
	}

	// Include usage insights if requested
	if usageMode {
		if err := attachUsageInsights(ctx, absPath, analysis); err != nil {
			fmt.Printf("⚠️  Usage analysis: %v\n", err)
		}
	}

	// Check for monorepo mode
	isMonorepo := monorepoMode
	if !isMonorepo && cfg.Monorepo != nil && cfg.Monorepo.PerWorkspace {
		isMonorepo = true
	}

	if isMonorepo && analysis.MonorepoInfo != nil && analysis.MonorepoInfo.IsMonorepo {
		return runMonorepoScan(ctx, absPath, cfg, formats, analysis)
	}

	// Generate output for each format
	for _, format := range formats {
		if err := generateOutput(absPath, format, analysis, dryRun, compactMode); err != nil {
			return err
		}
	}

	return nil
}

func runMonorepoScan(ctx context.Context, absPath string, cfg *config.Config, formats []string, rootAnalysis *types.Analysis) error {
	maxConcurrent := 4
	if cfg.Monorepo != nil && cfg.Monorepo.MaxConcurrent > 0 {
		maxConcurrent = cfg.Monorepo.MaxConcurrent
	}

	ma := analyzer.NewMonorepoAnalyzer(absPath, parallel, maxConcurrent)
	wsResults := ma.AnalyzeWorkspaces(ctx, rootAnalysis.MonorepoInfo)

	if len(wsResults) == 0 {
		fmt.Println("No workspaces found to analyze, falling back to single-project mode")
		for _, format := range formats {
			if err := generateOutput(absPath, format, rootAnalysis, dryRun, compactMode); err != nil {
				return err
			}
		}
		return nil
	}

	fmt.Printf("   Found %d workspaces\n", len(wsResults))

	// Generate per-workspace output
	successCount := 0
	for _, ws := range wsResults {
		if ws.Error != nil {
			fmt.Printf("   [skip] %s: %v\n", ws.Path, ws.Error)
			continue
		}

		// Determine formats for this workspace
		wsFormats := formats
		if cfg.Monorepo != nil && cfg.Monorepo.WorkspaceOverrides != nil {
			if override, ok := cfg.Monorepo.WorkspaceOverrides[ws.Path]; ok {
				if len(override.Output) > 0 {
					wsFormats = override.Output
				}
				for _, conv := range override.CustomConventions {
					ws.Analysis.Conventions = append(ws.Analysis.Conventions, types.Convention{
						Category:    "custom",
						Description: conv,
					})
				}
			}
		}

		wsAbsPath := filepath.Join(absPath, ws.Path)
		for _, format := range wsFormats {
			if err := generateOutput(wsAbsPath, format, ws.Analysis, dryRun, compactMode); err != nil {
				fmt.Printf("   [warn] %s (%s): %v\n", ws.Path, format, err)
			}
		}
		successCount++
	}

	// Generate root overview
	rootOverview := true
	if cfg.Monorepo != nil {
		rootOverview = cfg.Monorepo.RootOverview
	}

	if rootOverview {
		var wsInfos []generator.WorkspaceInfo
		for _, ws := range wsResults {
			if ws.Error != nil {
				continue
			}
			var langs []string
			for _, l := range ws.Analysis.TechStack.Languages {
				langs = append(langs, l.Name)
			}
			wsInfos = append(wsInfos, generator.WorkspaceInfo{
				Path:      ws.Path,
				Name:      ws.Name,
				Languages: langs,
				Commands:  len(ws.Analysis.Commands),
				Endpoints: len(ws.Analysis.Endpoints),
			})
		}

		overviewGen := generator.NewMonorepoOverviewGenerator("claude")
		overview, err := overviewGen.Generate(rootAnalysis, wsInfos)
		if err != nil {
			return fmt.Errorf("overview generation failed: %w", err)
		}

		outPath := filepath.Join(absPath, "CLAUDE.md")
		if dryRun {
			fmt.Printf("\n   Would write overview to %s\n", outPath)
		} else {
			if err := os.WriteFile(outPath, overview, 0644); err != nil {
				return fmt.Errorf("failed to write overview: %w", err)
			}
			fmt.Printf("✅ Generated monorepo overview: %s\n", outPath)
		}
	}

	if !dryRun {
		fmt.Printf("✅ Generated context files for %d/%d workspaces\n", successCount, len(wsResults))
	}

	return nil
}

func runSync(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if config exists
	if !config.Exists(absPath) {
		return fmt.Errorf("no .argus.yaml found in %s\nRun 'argus init' first", absPath)
	}

	// Load config
	cfg, err := config.Load(absPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("🔄 Syncing %s...\n", absPath)
	if parallel && verbose {
		fmt.Println("   Using parallel detector execution...")
	}

	// Create context with cancellation support
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals for graceful cancellation
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Run analysis (parallel or sequential)
	var analysis *types.Analysis
	if parallel {
		pa := analyzer.NewParallelAnalyzer(absPath, nil)
		analysis, err = pa.Analyze(ctx)
	} else {
		a := analyzer.NewAnalyzer(absPath, nil)
		analysis, err = a.Analyze(ctx)
	}
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Add custom conventions from config
	for _, conv := range cfg.CustomConventions {
		analysis.Conventions = append(analysis.Conventions, types.Convention{
			Category:    "custom",
			Description: conv,
		})
	}

	if verbose {
		fmt.Printf("\n📊 Analysis Results:\n")
		fmt.Printf("   Project: %s\n", analysis.ProjectName)
		fmt.Printf("   Languages: %d\n", len(analysis.TechStack.Languages))
		fmt.Printf("   Frameworks: %d\n", len(analysis.TechStack.Frameworks))
		fmt.Printf("   Conventions: %d\n", len(analysis.Conventions))
	}

	// Include usage insights if requested
	if usageMode {
		if err := attachUsageInsights(ctx, absPath, analysis); err != nil {
			fmt.Printf("⚠️  Usage analysis: %v\n", err)
		}
	}

	// Generate output for each format in config
	for _, format := range cfg.Output {
		if err := generateOutput(absPath, format, analysis, dryRun, compactMode); err != nil {
			return err
		}
	}

	return nil
}

// Generator interface for different output formats
type contextGenerator interface {
	Generate(analysis *types.Analysis) ([]byte, error)
	OutputFile() string
}

func generateOutput(absPath, format string, analysis *types.Analysis, dryRun, compact bool) error {
	// Handle claude-code format separately (multi-file generator)
	if format == "claude-code" {
		return generateClaudeCodeOutput(absPath, analysis, dryRun)
	}

	var gen contextGenerator
	var outputFile string

	switch format {
	case "claude":
		g := generator.NewClaudeGenerator()
		g.SetCompact(compact)
		gen = g
		outputFile = g.OutputFile()
	case "cursor":
		g := generator.NewCursorGenerator()
		gen = g
		outputFile = g.OutputFile()
	case "copilot":
		g := generator.NewCopilotGenerator()
		gen = g
		outputFile = g.OutputFile()
	case "continue":
		g := generator.NewContinueGenerator()
		gen = g
		outputFile = g.OutputFile()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	content, err := gen.Generate(analysis)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	outPath := filepath.Join(absPath, outputFile)

	// Handle merge mode - preserve custom sections from existing file
	if mergeMode {
		existingContent, err := os.ReadFile(outPath)
		if err == nil && len(existingContent) > 0 {
			m := merger.NewMerger(true)
			content = m.Merge(existingContent, content)
			if verbose {
				fmt.Printf("   ℹ️  Merged with existing content (preserving custom sections)\n")
			}
		}
	}

	// Add custom block placeholder if requested
	if addCustomBlock {
		contentStr := string(content)
		if !strings.Contains(contentStr, merger.CustomStartMarker) {
			content = []byte(merger.AddCustomSectionPlaceholder(contentStr))
		}
	}

	if dryRun {
		fmt.Printf("\n📄 Would write to %s:\n", outPath)
		fmt.Println("---")
		fmt.Println(string(content))
		fmt.Println("---")
		return nil
	}

	// Ensure parent directory exists (for .github/copilot-instructions.md)
	if dir := filepath.Dir(outPath); dir != absPath {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if err := os.WriteFile(outPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Generated %s\n", outPath)
	return nil
}

// generateClaudeCodeOutput handles the multi-file claude-code format
func generateClaudeCodeOutput(absPath string, analysis *types.Analysis, dryRun bool) error {
	// Load config to get ClaudeCode settings
	cfg, err := config.Load(absPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	g := generator.NewClaudeCodeGenerator(cfg.ClaudeCode)
	files, err := g.Generate(analysis)
	if err != nil {
		return fmt.Errorf("claude-code generation failed: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("ℹ️  No Claude Code configs generated (check claude_code settings)")
		return nil
	}

	for _, file := range files {
		outPath := filepath.Join(absPath, file.Path)

		if dryRun {
			fmt.Printf("\n📄 Would write to %s:\n", outPath)
			fmt.Println("---")
			fmt.Println(string(file.Content))
			fmt.Println("---")
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", file.Path, err)
		}

		if err := os.WriteFile(outPath, file.Content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", file.Path, err)
		}

		if verbose {
			fmt.Printf("✅ Generated %s\n", outPath)
		}
	}

	if !dryRun {
		fmt.Printf("✅ Generated %d Claude Code configs in .claude/\n", len(files))
	}

	return nil
}

func runWatch(cmd *cobra.Command, args []string) error {
	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("path does not exist: %s", absPath)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Load config
	cfg, err := config.Load(absPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	// Add directories to watch (recursively)
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			// Skip common directories that don't need watching
			name := info.Name()
			if shouldIgnoreDir(name) {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to add directories to watcher: %w", err)
	}

	fmt.Printf("👁️  Watching %s for changes...\n", absPath)
	fmt.Printf("   Output formats: %s\n", strings.Join(cfg.Output, ", "))
	fmt.Println("   Mode: incremental updates")
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println()

	// Create context with cancellation for watch mode
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create incremental analyzer
	incAnalyzer := analyzer.NewIncrementalAnalyzer(absPath)

	// Do initial full generation
	fmt.Println("🔍 Running initial full analysis...")
	if err := regenerateWithAnalyzer(ctx, absPath, cfg, incAnalyzer, "", true); err != nil {
		fmt.Printf("⚠️  Initial generation failed: %v\n", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Debounce timer and last changed file
	var debounceTimer *time.Timer
	var lastChangedFile string
	debounceDelay := 500 * time.Millisecond

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			// Skip irrelevant events
			if !isRelevantChange(event) {
				continue
			}

			// Skip generated files to avoid infinite loops
			if isGeneratedFile(event.Name) {
				continue
			}

			// Store the changed file
			lastChangedFile = event.Name

			// Debounce: reset timer on each event
			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			changedFile := lastChangedFile // Capture for closure
			debounceTimer = time.AfterFunc(debounceDelay, func() {
				if err := regenerateWithAnalyzer(ctx, absPath, cfg, incAnalyzer, changedFile, false); err != nil {
					fmt.Printf("⚠️  Regeneration failed: %v\n", err)
				}
			})

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("⚠️  Watcher error: %v\n", err)

		case <-sigChan:
			fmt.Println("\n👋 Stopping watcher...")
			return nil
		}
	}
}

func regenerateWithAnalyzer(ctx context.Context, absPath string, cfg *config.Config, incAnalyzer *analyzer.IncrementalAnalyzer, changedFile string, isInitial bool) error {
	var analysis *types.Analysis
	var impacts []string
	var err error

	startTime := time.Now()

	if isInitial || changedFile == "" {
		// Full analysis for initial run
		analysis, err = incAnalyzer.AnalyzeFull(ctx)
		impacts = []string{analyzer.ImpactAll}
	} else {
		// Incremental analysis for file changes
		analysis, impacts, err = incAnalyzer.AnalyzeIncremental(ctx, changedFile)
	}

	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// Add custom conventions from config
	for _, conv := range cfg.CustomConventions {
		analysis.Conventions = append(analysis.Conventions, types.Convention{
			Category:    "custom",
			Description: conv,
		})
	}

	// Generate output for each format
	for _, format := range cfg.Output {
		if err := generateOutput(absPath, format, analysis, false, compactMode); err != nil {
			return err
		}
	}

	elapsed := time.Since(startTime)

	// Print status
	if isInitial {
		fmt.Printf("✅ Initial generation complete (%dms)\n\n", elapsed.Milliseconds())
	} else {
		impactDesc := analyzer.ImpactDescription(impacts)
		fmt.Printf("🔄 %s → updated: %s (%dms)\n", filepath.Base(changedFile), impactDesc, elapsed.Milliseconds())
	}

	return nil
}

func shouldIgnoreDir(name string) bool {
	ignoreDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		"vendor":       true,
		".next":        true,
		"dist":         true,
		"build":        true,
		"__pycache__":  true,
		".venv":        true,
		"venv":         true,
		".idea":        true,
		".vscode":      true,
		"target":       true, // Rust/Java
		"bin":          true,
		"obj":          true, // C#
	}
	return ignoreDirs[name]
}

func isRelevantChange(event fsnotify.Event) bool {
	// Only care about write, create, remove events
	if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) == 0 {
		return false
	}

	// Check if it's a relevant file type
	ext := strings.ToLower(filepath.Ext(event.Name))
	relevantExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".py": true, ".java": true, ".kt": true, ".rs": true, ".rb": true,
		".cs": true, ".cpp": true, ".c": true, ".h": true, ".hpp": true,
		".swift": true, ".php": true, ".vue": true, ".svelte": true,
		".json": true, ".yaml": true, ".yml": true, ".toml": true,
		".md": true, ".txt": true,
	}

	// Also watch config files
	name := filepath.Base(event.Name)
	configFiles := map[string]bool{
		"package.json": true, "go.mod": true, "Cargo.toml": true,
		"pyproject.toml": true, "requirements.txt": true,
		"pom.xml": true, "build.gradle": true,
		".argus.yaml": true,
	}

	return relevantExts[ext] || configFiles[name]
}

func isGeneratedFile(path string) bool {
	// Check if path is in .claude/ directory
	if strings.Contains(path, ".claude/") || strings.Contains(path, ".claude"+string(filepath.Separator)) {
		return true
	}

	name := filepath.Base(path)
	generated := map[string]bool{
		"CLAUDE.md":               true,
		".cursorrules":            true,
		"copilot-instructions.md": true,
	}
	return generated[name]
}

func runInsights(cmd *cobra.Command, args []string) error {
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	opts := usage.Options{
		IncludeSubagents: insightsSubagents,
	}

	if insightsSince != "" {
		since, err := parseSinceFlag(insightsSince)
		if err != nil {
			return fmt.Errorf("invalid --since value: %w", err)
		}
		opts.Since = since
	}

	fmt.Printf("🔍 Analyzing Claude Code usage for %s...\n", filepath.Base(absPath))

	insights, err := usage.AnalyzeUsage(ctx, absPath, opts)
	if err != nil {
		return fmt.Errorf("usage analysis failed: %w", err)
	}

	if insights == nil {
		fmt.Println("No Claude Code session data found for this project.")
		fmt.Printf("Expected logs at: ~/.claude/projects/%s/\n", usage.EncodeProjectPath(absPath))
		return nil
	}

	if insightsFormat == "json" {
		return printInsightsJSON(insights)
	}

	printInsightsText(insights, filepath.Base(absPath))
	return nil
}

func attachUsageInsights(ctx context.Context, absPath string, analysis *types.Analysis) error {
	opts := usage.Options{
		Since:            time.Now().AddDate(0, -1, 0), // Last 30 days
		IncludeSubagents: true,
	}

	insights, err := usage.AnalyzeUsage(ctx, absPath, opts)
	if err != nil {
		return err
	}

	if insights != nil {
		analysis.UsageInsights = insights
		if verbose {
			fmt.Printf("   📊 Usage: %d sessions, %d turns, $%.2f estimated cost\n",
				insights.SessionCount, insights.TotalTurns, insights.CostEstimate.TotalCost)
		}
	}

	return nil
}

func printInsightsJSON(insights *types.UsageInsights) error {
	data, err := json.MarshalIndent(insights, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printInsightsText(insights *types.UsageInsights, projectName string) {
	dateRange := fmt.Sprintf("%s - %s",
		insights.DateRange.Start.Format("Jan 2, 2006"),
		insights.DateRange.End.Format("Jan 2, 2006"))

	fmt.Printf("\nClaude Code Usage Insights for: %s\n", projectName)
	fmt.Printf("Period: %s (%d sessions, %d turns)\n", dateRange, insights.SessionCount, insights.TotalTurns)
	fmt.Println(strings.Repeat("─", 60))

	// Tool Usage
	if len(insights.ToolUsage) > 0 {
		fmt.Println("\nTOOL USAGE")
		maxCount := insights.ToolUsage[0].Count
		for _, ts := range insights.ToolUsage {
			barLen := 0
			if maxCount > 0 {
				barLen = ts.Count * 25 / maxCount
			}
			bar := strings.Repeat("█", barLen)
			fmt.Printf("  %-14s %5d (%4.1f%%)  %s\n", ts.Name, ts.Count, ts.Percentage, bar)
		}
	}

	// Hot Files
	if len(insights.HotFiles) > 0 {
		fmt.Println("\nHOT FILES (most AI interaction)")
		fmt.Printf("  %-4s %-45s %5s %5s %5s %5s\n", "Rank", "File", "Reads", "Edits", "Writes", "Total")
		limit := len(insights.HotFiles)
		if limit > 10 {
			limit = 10
		}
		for i, hf := range insights.HotFiles[:limit] {
			path := hf.Path
			if len(path) > 45 {
				path = "..." + path[len(path)-42:]
			}
			fmt.Printf("  %-4d %-45s %5d %5d %5d %5d\n",
				i+1, path, hf.ReadCount, hf.EditCount, hf.WriteCount, hf.TotalOps)
		}
	}

	// Pain Points
	if len(insights.PainPoints) > 0 {
		fmt.Println("\nAI PAIN POINTS")
		for _, pp := range insights.PainPoints {
			path := pp.File
			if len(path) > 40 {
				path = "..." + path[len(path)-37:]
			}
			fmt.Printf("  ! %-40s %s\n", path, pp.Description)
		}
	}

	// Token Usage
	if len(insights.ModelBreakdown) > 0 {
		fmt.Println("\nTOKEN USAGE")
		fmt.Printf("  %-25s %10s %10s %10s\n", "Model", "Input", "Output", "Total")
		for _, mu := range insights.ModelBreakdown {
			fmt.Printf("  %-25s %10s %10s %10s\n",
				mu.Model,
				formatTokens(mu.TokenUsage.InputTokens),
				formatTokens(mu.TokenUsage.OutputTokens),
				formatTokens(mu.TokenUsage.TotalTokens))
		}
		fmt.Printf("  %-25s %10s %10s %10s\n",
			"Total",
			formatTokens(insights.TokenUsage.InputTokens),
			formatTokens(insights.TokenUsage.OutputTokens),
			formatTokens(insights.TokenUsage.TotalTokens))
	}

	// Cost Estimate
	fmt.Printf("\nCOST ESTIMATE: $%.2f\n", insights.CostEstimate.TotalCost)
	fmt.Printf("  Input: $%.2f  Output: $%.2f  Cache: $%.2f\n",
		insights.CostEstimate.InputCost, insights.CostEstimate.OutputCost, insights.CostEstimate.CacheCost)
	fmt.Println()
}

func formatTokens(n int64) string {
	switch {
	case n >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", float64(n)/1_000_000_000)
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func parseSinceFlag(s string) (time.Time, error) {
	// Handle relative durations like "7d", "30d"
	if len(s) > 1 && s[len(s)-1] == 'd' {
		days := 0
		for _, c := range s[:len(s)-1] {
			if c < '0' || c > '9' {
				break
			}
			days = days*10 + int(c-'0')
		}
		if days > 0 {
			return time.Now().AddDate(0, 0, -days), nil
		}
	}

	// Try ISO date format
	t, err := time.Parse("2006-01-02", s)
	if err == nil {
		return t, nil
	}

	// Try YYYYMMDD format
	t, err = time.Parse("20060102", s)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("unsupported format %q (use 7d, 30d, or 2025-01-01)", s)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	fmt.Printf("🔍 Current version: %s\n", version)
	fmt.Println("📡 Checking for updates...")

	// Fetch latest release info
	release, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(version, "v")

	if latestVersion == currentVersion {
		fmt.Printf("✅ Already on the latest version (%s)\n", version)
		return nil
	}

	fmt.Printf("📦 New version available: %s → %s\n", version, release.TagName)

	// Find the right asset for this OS/arch
	assetName := getAssetName(latestVersion)
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no release found for %s/%s (looking for %s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	fmt.Printf("⬇️  Downloading %s...\n", assetName)

	// Download the asset
	tmpFile, err := downloadAsset(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile) }()

	// Extract the binary
	fmt.Println("📂 Extracting...")
	binaryPath, err := extractBinary(tmpFile, assetName)
	if err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}
	defer func() { _ = os.Remove(binaryPath) }()

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Replace the binary
	fmt.Println("🔄 Installing...")
	if err := replaceBinary(binaryPath, execPath); err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}

	fmt.Printf("✅ Successfully upgraded to %s!\n", release.TagName)
	return nil
}

func getLatestRelease() (*GitHubRelease, error) {
	resp, err := http.Get("https://api.github.com/repos/Priyans-hu/argus/releases/latest")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func getAssetName(ver string) string {
	goos := runtime.GOOS
	arch := runtime.GOARCH

	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}

	return fmt.Sprintf("argus_%s_%s_%s%s", ver, goos, arch, ext)
}

func downloadAsset(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "argus-upgrade-*")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if closeErr := tmpFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

func extractBinary(archivePath, assetName string) (string, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractZip(archivePath)
	}
	return extractTarGz(archivePath)
}

func extractTarGz(archivePath string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Typeflag == tar.TypeReg && (header.Name == "argus" || strings.HasSuffix(header.Name, "/argus")) {
			tmpFile, err := os.CreateTemp("", "argus-binary-*")
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(tmpFile, tr); err != nil {
				_ = tmpFile.Close()
				_ = os.Remove(tmpFile.Name())
				return "", err
			}
			_ = tmpFile.Close()

			if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
				_ = os.Remove(tmpFile.Name())
				return "", err
			}

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("argus binary not found in archive")
}

func extractZip(archivePath string) (string, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		if f.Name == "argus.exe" || strings.HasSuffix(f.Name, "/argus.exe") {
			rc, err := f.Open()
			if err != nil {
				return "", err
			}

			tmpFile, err := os.CreateTemp("", "argus-binary-*.exe")
			if err != nil {
				_ = rc.Close()
				return "", err
			}

			if _, err := io.Copy(tmpFile, rc); err != nil {
				_ = tmpFile.Close()
				_ = rc.Close()
				_ = os.Remove(tmpFile.Name())
				return "", err
			}
			_ = tmpFile.Close()
			_ = rc.Close()

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("argus.exe not found in archive")
}

func replaceBinary(newBinary, targetPath string) error {
	// On Windows, we can't replace a running executable directly
	// So we rename the old one first
	if runtime.GOOS == "windows" {
		oldPath := targetPath + ".old"
		_ = os.Remove(oldPath) // Remove any existing .old file
		if err := os.Rename(targetPath, oldPath); err != nil {
			return err
		}
	}

	// Read the new binary
	data, err := os.ReadFile(newBinary)
	if err != nil {
		return err
	}

	// Write to target
	if err := os.WriteFile(targetPath, data, 0755); err != nil {
		return err
	}

	return nil
}
