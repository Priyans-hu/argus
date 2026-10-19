package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/Priyans-hu/argus/internal/analyzer"
	"github.com/Priyans-hu/argus/internal/config"
	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [path]",
	Short: "Report generated files that scan would change, without writing",
	Long: `Render the same outputs as scan and compare them with the files on disk.
Nothing is written. Exit status: 0 when every file is up to date, 1 when scan
would write a file, 2 on an error.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheck,
}

// errDrift reports that scan would write; it exits 1 without an error line.
var errDrift = errors.New("generated files are out of date")

func runCheck(cmd *cobra.Command, args []string) error {
	target := "."
	if len(args) > 0 {
		target = args[0]
	}
	absPath, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	if info, err := os.Stat(absPath); err != nil || !info.IsDir() {
		return fmt.Errorf("not a directory: %s", absPath)
	}
	cfg, err := config.Load(absPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.AI != nil && cfg.AI.Enabled {
		return errors.New("ai is enabled in .argus.yaml: check cannot reproduce AI output")
	}
	// Commit and branch conventions come from history.
	if _, err := os.Stat(filepath.Join(absPath, ".git", "shallow")); err == nil {
		return errors.New("check needs full git history, this clone is shallow (use fetch-depth: 0)")
	}

	formats := cfg.Output
	if cmd.Flags().Changed("format") {
		formats = []string{outputFormat}
	}
	formats = expandFormats(formats)
	warnContinue(cfg, formats)

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	analysis, err := analyzer.NewParallelAnalyzer(absPath, analyzerConfig(cfg)).Analyze(ctx)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}
	for _, conv := range cfg.CustomConventions {
		analysis.Conventions = append(analysis.Conventions, types.Convention{Category: "custom", Description: conv})
	}

	// Render through scan's write layer: writeOutput records instead of
	// writing, and scan's own result lines are dropped.
	out := stdout
	checked, stdout = map[string]string{}, io.Discard
	defer func() { checked, stdout, skipped = nil, out, nil }()
	if err := writeOutputs(ctx, absPath, cfg, formats, analysis); err != nil {
		return err
	}

	drift, total := 0, 0
	for _, p := range slices.Sorted(maps.Keys(checked)) {
		state := checked[p]
		_, _ = fmt.Fprintf(out, "%s: %s\n", p, state)
		switch state {
		case "out of date", "missing":
			drift++
			total++
		case "up to date":
			total++
		}
	}
	if drift > 0 {
		errf("argus check: %d of %d files out of date (run argus scan to update)\n", drift, total)
		return errDrift
	}
	errf("argus check: all %d files up to date\n", total)
	return nil
}
