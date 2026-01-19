package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

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
	Use:   "scan",
	Short: "Scan codebase and generate context files",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Argus is scanning your codebase...")
		// TODO: Implement scanning
	},
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
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
