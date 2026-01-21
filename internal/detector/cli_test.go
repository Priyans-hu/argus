package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestCLIDetector_Detect_WithCobraFramework(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cli-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create cmd/app/main.go with verbose and dry-run flags
	cmdDir := filepath.Join(tmpDir, "cmd", "app")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("failed to create cmd dir: %v", err)
	}

	mainContent := `package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var verbose bool
var dryRun bool

func main() {
	rootCmd := &cobra.Command{
		Use:   "app",
		Short: "Test app",
		Run: func(cmd *cobra.Command, args []string) {
			if verbose {
				fmt.Println("🔍 Scanning...")
			}
			fmt.Println("✅ Success")
			if dryRun {
				fmt.Println("📄 Dry run mode")
			}
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "n", false, "dry run mode")
	rootCmd.Execute()
}
`
	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	files := []types.FileInfo{
		{Path: "cmd/app/main.go", Extension: ".go"},
	}

	techStack := &types.TechStack{
		Frameworks: []types.Framework{
			{Name: "Cobra", Category: "CLI"},
		},
	}

	detector := NewCLIDetector(tmpDir, files, techStack)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected CLI info to be detected")
	}

	if result.VerboseFlag == "" {
		t.Error("expected verbose flag to be detected")
	}

	if result.DryRunFlag == "" {
		t.Error("expected dry-run flag to be detected")
	}

	if len(result.Indicators) == 0 {
		t.Error("expected indicators to be detected")
	}

	// Check for specific indicators
	hasIndicator := func(symbol string) bool {
		for _, ind := range result.Indicators {
			if ind.Symbol == symbol {
				return true
			}
		}
		return false
	}

	if !hasIndicator("✅") {
		t.Error("expected success indicator to be detected")
	}
	if !hasIndicator("🔍") {
		t.Error("expected scanning indicator to be detected")
	}
}

func TestCLIDetector_Detect_WithCmdDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cli-cmd-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create cmd/main.go
	cmdDir := filepath.Join(tmpDir, "cmd")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("failed to create cmd dir: %v", err)
	}

	mainContent := `package main

func main() {
	// Simple CLI with verbose flag
	verbose := "--verbose"
	_ = verbose
}
`
	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte(mainContent), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	files := []types.FileInfo{
		{Path: "cmd/main.go", Extension: ".go"},
	}

	// No CLI framework but has cmd/ directory
	techStack := &types.TechStack{
		Languages: []types.Language{{Name: "Go"}},
	}

	detector := NewCLIDetector(tmpDir, files, techStack)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected CLI info to be detected for cmd/ directory project")
	}

	if result.VerboseFlag == "" {
		t.Error("expected verbose flag to be detected")
	}
}

func TestCLIDetector_Detect_NotCLIProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "non-cli-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	files := []types.FileInfo{
		{Path: "main.go", Extension: ".go"},
	}

	// No CLI framework and no cmd/ directory
	techStack := &types.TechStack{
		Languages: []types.Language{{Name: "Go"}},
		Frameworks: []types.Framework{
			{Name: "Gin", Category: "Web"},
		},
	}

	detector := NewCLIDetector(tmpDir, files, techStack)
	result := detector.Detect()

	if result != nil {
		t.Error("expected no CLI info for non-CLI project")
	}
}

func TestCLIDetector_Detect_NilTechStack(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nil-techstack-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	detector := NewCLIDetector(tmpDir, nil, nil)
	result := detector.Detect()

	if result != nil {
		t.Error("expected no CLI info when techStack is nil")
	}
}

func TestCLIDetector_IsCLIProject(t *testing.T) {
	tests := []struct {
		name      string
		files     []types.FileInfo
		techStack *types.TechStack
		expectCLI bool
	}{
		{
			name:  "cobra framework",
			files: []types.FileInfo{},
			techStack: &types.TechStack{
				Frameworks: []types.Framework{{Name: "Cobra"}},
			},
			expectCLI: true,
		},
		{
			name:  "click framework",
			files: []types.FileInfo{},
			techStack: &types.TechStack{
				Frameworks: []types.Framework{{Name: "Click"}},
			},
			expectCLI: true,
		},
		{
			name:  "commander framework",
			files: []types.FileInfo{},
			techStack: &types.TechStack{
				Frameworks: []types.Framework{{Name: "Commander"}},
			},
			expectCLI: true,
		},
		{
			name: "cmd directory",
			files: []types.FileInfo{
				{Path: "cmd/app/main.go", Extension: ".go"},
			},
			techStack: &types.TechStack{},
			expectCLI: true,
		},
		{
			name: "web framework only",
			files: []types.FileInfo{
				{Path: "main.go", Extension: ".go"},
			},
			techStack: &types.TechStack{
				Frameworks: []types.Framework{{Name: "Express"}},
			},
			expectCLI: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector := &CLIDetector{
				rootPath:  "/tmp",
				files:     tt.files,
				techStack: tt.techStack,
			}
			result := detector.isCLIProject()
			if result != tt.expectCLI {
				t.Errorf("expected isCLIProject=%v, got %v", tt.expectCLI, result)
			}
		})
	}
}
