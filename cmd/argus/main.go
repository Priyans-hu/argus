package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Priyans-hu/argus/internal/analyzer"
	"github.com/Priyans-hu/argus/internal/config"
	"github.com/Priyans-hu/argus/internal/generator"
	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

// Flags
var (
	outputDir    string
	outputFormat string
	dryRun       bool
	verbose      bool
	force        bool
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

	// Sync command flags
	syncCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be generated without writing files")
	syncCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(syncCmd)
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

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
