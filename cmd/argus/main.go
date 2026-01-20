package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Priyans-hu/argus/internal/analyzer"
	"github.com/Priyans-hu/argus/internal/config"
	"github.com/Priyans-hu/argus/internal/generator"
	"github.com/Priyans-hu/argus/internal/merger"
	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

// Flags
var (
	outputDir      string
	outputFormat   string
	dryRun         bool
	verbose        bool
	force          bool
	mergeMode      bool
	addCustomBlock bool
)

var rootCmd = &cobra.Command{
	Use:   "argus",
	Short: "Help AI grok your codebase",
	Long: `Argus - The All-Seeing Code Analyzer

Argus scans your codebase and generates optimized context files
for AI coding assistants (Claude Code, Cursor, Copilot, etc.).

No more manually writing CLAUDE.md or .cursorrules - Argus sees everything.`,
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

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("argus version %s\n", version)
	},
}

func init() {
	// Init command flags
	initCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing config file")

	// Scan command flags
	scanCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated files")
	scanCmd.Flags().StringVarP(&outputFormat, "format", "f", "claude", "Output format: claude, cursor, copilot, all")
	scanCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be generated without writing files")
	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	scanCmd.Flags().BoolVarP(&mergeMode, "merge", "m", true, "Preserve custom sections when regenerating (default: true)")
	scanCmd.Flags().BoolVar(&addCustomBlock, "add-custom", false, "Add a custom section placeholder to output")

	// Sync command flags
	syncCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be generated without writing files")
	syncCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	syncCmd.Flags().BoolVarP(&mergeMode, "merge", "m", true, "Preserve custom sections when regenerating (default: true)")
	syncCmd.Flags().BoolVar(&addCustomBlock, "add-custom", false, "Add a custom section placeholder to output")

	// Watch command flags
	watchCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")
	watchCmd.Flags().BoolVarP(&mergeMode, "merge", "m", true, "Preserve custom sections when regenerating (default: true)")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(versionCmd)
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
			formats = []string{"claude", "cursor", "copilot"}
		} else {
			formats = []string{outputFormat}
		}
	}

	fmt.Printf("🔍 Scanning %s...\n", absPath)

	// Run analysis
	a := analyzer.NewAnalyzer(absPath, nil)
	analysis, err := a.Analyze()
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

	// Generate output for each format
	for _, format := range formats {
		if err := generateOutput(absPath, format, analysis, dryRun); err != nil {
			return err
		}
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

	// Run analysis
	a := analyzer.NewAnalyzer(absPath, nil)
	analysis, err := a.Analyze()
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

	// Generate output for each format in config
	for _, format := range cfg.Output {
		if err := generateOutput(absPath, format, analysis, dryRun); err != nil {
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

func generateOutput(absPath, format string, analysis *types.Analysis, dryRun bool) error {
	var gen contextGenerator
	var outputFile string

	switch format {
	case "claude":
		g := generator.NewClaudeGenerator()
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
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println()

	// Do initial generation
	if err := regenerate(absPath, cfg); err != nil {
		fmt.Printf("⚠️  Initial generation failed: %v\n", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Debounce timer
	var debounceTimer *time.Timer
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

			// Debounce: reset timer on each event
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(debounceDelay, func() {
				fmt.Printf("🔄 Change detected: %s\n", filepath.Base(event.Name))
				if err := regenerate(absPath, cfg); err != nil {
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

func regenerate(absPath string, cfg *config.Config) error {
	// Run analysis
	a := analyzer.NewAnalyzer(absPath, nil)
	analysis, err := a.Analyze()
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
		if err := generateOutput(absPath, format, analysis, false); err != nil {
			return err
		}
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
	name := filepath.Base(path)
	generated := map[string]bool{
		"CLAUDE.md":                true,
		".cursorrules":             true,
		"copilot-instructions.md":  true,
	}
	return generated[name]
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
