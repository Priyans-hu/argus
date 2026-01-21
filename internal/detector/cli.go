package detector

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// CLIDetector detects CLI-specific information
type CLIDetector struct {
	rootPath  string
	files     []types.FileInfo
	techStack *types.TechStack
}

// NewCLIDetector creates a new CLI detector
func NewCLIDetector(rootPath string, files []types.FileInfo, techStack *types.TechStack) *CLIDetector {
	return &CLIDetector{
		rootPath:  rootPath,
		files:     files,
		techStack: techStack,
	}
}

// Detect extracts CLI-specific information
func (d *CLIDetector) Detect() *types.CLIInfo {
	// Only process if this is a CLI project
	if !d.isCLIProject() {
		return nil
	}

	info := &types.CLIInfo{}

	// Detect flags from source code
	info.VerboseFlag, info.DryRunFlag = d.detectFlags()

	// Detect output indicators
	info.Indicators = d.detectIndicators()

	// Return nil if nothing detected
	if info.VerboseFlag == "" && info.DryRunFlag == "" && len(info.Indicators) == 0 {
		return nil
	}

	return info
}

// isCLIProject checks if this is a CLI project
func (d *CLIDetector) isCLIProject() bool {
	if d.techStack == nil {
		return false
	}

	// Check for CLI frameworks
	cliFrameworks := []string{
		"cobra", "urfave/cli", "kingpin", "pflag", // Go
		"click", "typer", "argparse", // Python
		"clap", "structopt", // Rust
		"commander", "yargs", "meow", // Node.js
	}

	for _, fw := range d.techStack.Frameworks {
		fwLower := strings.ToLower(fw.Name)
		for _, cli := range cliFrameworks {
			if strings.Contains(fwLower, cli) {
				return true
			}
		}
	}

	// Check for cmd/ directory (common Go CLI pattern)
	for _, f := range d.files {
		if strings.HasPrefix(f.Path, "cmd/") && f.Extension == ".go" {
			return true
		}
	}

	return false
}

// detectFlags detects verbose and dry-run flags from source code
func (d *CLIDetector) detectFlags() (verboseFlag, dryRunFlag string) {
	// Patterns for flag definitions
	verbosePatterns := []string{
		`-v.*verbose`,
		`--verbose`,
		`-verbose`,
		`"verbose"`,
		`'verbose'`,
	}

	dryRunPatterns := []string{
		`-n.*dry-run`,
		`--dry-run`,
		`-dry-run`,
		`"dry-run"`,
		`'dry-run'`,
		`"dryRun"`,
	}

	// Find main.go or cmd/*.go files
	mainFiles := d.findMainFiles()

	for _, mainFile := range mainFiles {
		content, err := os.ReadFile(mainFile)
		if err != nil {
			continue
		}

		contentStr := string(content)

		// Check for verbose flag
		if verboseFlag == "" {
			for _, pattern := range verbosePatterns {
				if matched, _ := regexp.MatchString(pattern, contentStr); matched {
					verboseFlag = "-v, --verbose"
					break
				}
			}
		}

		// Check for dry-run flag
		if dryRunFlag == "" {
			for _, pattern := range dryRunPatterns {
				if matched, _ := regexp.MatchString(pattern, contentStr); matched {
					dryRunFlag = "-n, --dry-run"
					break
				}
			}
		}

		if verboseFlag != "" && dryRunFlag != "" {
			break
		}
	}

	return verboseFlag, dryRunFlag
}

// detectIndicators detects output indicators (emojis or symbols)
func (d *CLIDetector) detectIndicators() []types.Indicator {
	var indicators []types.Indicator

	// Common CLI indicator patterns
	indicatorPatterns := map[string]string{
		"✅":  "Success",
		"❌":  "Error",
		"⚠️": "Warning",
		"🔍":  "Scanning/analyzing",
		"🔄":  "Processing/syncing",
		"📊":  "Analysis results",
		"📄":  "File output",
		"👁️": "Watch mode",
		"✓":  "Success",
		"✗":  "Error",
		"→":  "Progress indicator",
	}

	// Find source files
	mainFiles := d.findMainFiles()

	foundIndicators := make(map[string]bool)

	for _, mainFile := range mainFiles {
		content, err := os.ReadFile(mainFile)
		if err != nil {
			continue
		}

		contentStr := string(content)

		for symbol, meaning := range indicatorPatterns {
			if strings.Contains(contentStr, symbol) && !foundIndicators[symbol] {
				foundIndicators[symbol] = true
				indicators = append(indicators, types.Indicator{
					Symbol:  symbol,
					Meaning: meaning,
				})
			}
		}
	}

	return indicators
}

// findMainFiles finds main.go and cmd/*.go files
func (d *CLIDetector) findMainFiles() []string {
	var files []string

	// Check main.go
	mainPath := filepath.Join(d.rootPath, "main.go")
	if _, err := os.Stat(mainPath); err == nil {
		files = append(files, mainPath)
	}

	// Check cmd/ directory
	cmdDir := filepath.Join(d.rootPath, "cmd")
	if entries, err := os.ReadDir(cmdDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				// Check for main.go in subdirectory
				subMain := filepath.Join(cmdDir, entry.Name(), "main.go")
				if _, err := os.Stat(subMain); err == nil {
					files = append(files, subMain)
				}
			} else if strings.HasSuffix(entry.Name(), ".go") {
				files = append(files, filepath.Join(cmdDir, entry.Name()))
			}
		}
	}

	// Also check internal/generator for output patterns (for this project specifically)
	genDir := filepath.Join(d.rootPath, "internal", "generator")
	if entries, err := os.ReadDir(genDir); err == nil {
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".go") {
				files = append(files, filepath.Join(genDir, entry.Name()))
			}
		}
	}

	return files
}
