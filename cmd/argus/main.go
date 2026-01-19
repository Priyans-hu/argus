package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Priyans-hu/argus/internal/analyzer"
	"github.com/Priyans-hu/argus/internal/generator"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

// Flags
var (
	outputDir    string
	outputFormat string
	dryRun       bool
	verbose      bool
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
	Use:   "init",
	Short: "Initialize Argus in current directory",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing Argus...")
		// TODO: Implement initialization
	},
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
	Use:   "sync",
	Short: "Update context files with latest changes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncing context files...")
		// TODO: Implement sync
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("argus version %s\n", version)
	},
}

func init() {
	// Scan command flags
	scanCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated files")
	scanCmd.Flags().StringVarP(&outputFormat, "format", "f", "claude", "Output format: claude, cursor, copilot, all")
	scanCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be generated without writing files")
	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed output")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(versionCmd)
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

	fmt.Printf("🔍 Scanning %s...\n", absPath)

	// Run analysis
	a := analyzer.NewAnalyzer(absPath, nil)
	analysis, err := a.Analyze()
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	if verbose {
		fmt.Printf("\n📊 Analysis Results:\n")
		fmt.Printf("   Project: %s\n", analysis.ProjectName)
		fmt.Printf("   Languages: %d\n", len(analysis.TechStack.Languages))
		fmt.Printf("   Frameworks: %d\n", len(analysis.TechStack.Frameworks))
		fmt.Printf("   Directories: %d\n", len(analysis.Structure.Directories))
		fmt.Printf("   Key Files: %d\n", len(analysis.KeyFiles))
		fmt.Printf("   Commands: %d\n", len(analysis.Commands))
	}

	// Generate output
	var gen *generator.ClaudeGenerator
	var outputFile string

	switch outputFormat {
	case "claude", "all":
		gen = generator.NewClaudeGenerator()
		outputFile = gen.OutputFile()
	default:
		gen = generator.NewClaudeGenerator()
		outputFile = gen.OutputFile()
	}

	content, err := gen.Generate(analysis)
	if err != nil {
		return fmt.Errorf("generation failed: %w", err)
	}

	// Determine output path
	outPath := filepath.Join(outputDir, outputFile)
	if outputDir == "." {
		outPath = filepath.Join(absPath, outputFile)
	}

	if dryRun {
		fmt.Printf("\n📄 Would write to %s:\n", outPath)
		fmt.Println("---")
		fmt.Println(string(content))
		fmt.Println("---")
		return nil
	}

	// Write file
	if err := os.WriteFile(outPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("\n✅ Generated %s\n", outPath)

	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
