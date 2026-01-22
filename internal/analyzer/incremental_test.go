package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetermineImpact_ConfigFile(t *testing.T) {
	impacts := DetermineImpact(".argus.yaml")

	if len(impacts) != 1 || impacts[0] != ImpactAll {
		t.Errorf("expected ImpactAll for .argus.yaml, got %v", impacts)
	}
}

func TestDetermineImpact_PackageManager(t *testing.T) {
	tests := []struct {
		file     string
		expected []string
	}{
		{"go.mod", []string{ImpactTechStack, ImpactDevelopment}},
		{"package.json", []string{ImpactTechStack, ImpactDevelopment}},
		{"Cargo.toml", []string{ImpactTechStack, ImpactDevelopment}},
		{"pyproject.toml", []string{ImpactTechStack, ImpactDevelopment}},
		{"requirements.txt", []string{ImpactTechStack, ImpactDevelopment}},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			impacts := DetermineImpact(tt.file)
			if len(impacts) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, impacts)
				return
			}
			for i, imp := range impacts {
				if imp != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected, impacts)
				}
			}
		})
	}
}

func TestDetermineImpact_Makefile(t *testing.T) {
	impacts := DetermineImpact("Makefile")

	hasCommands := false
	hasDevelopment := false
	for _, imp := range impacts {
		if imp == ImpactCommands {
			hasCommands = true
		}
		if imp == ImpactDevelopment {
			hasDevelopment = true
		}
	}

	if !hasCommands || !hasDevelopment {
		t.Errorf("expected commands and development impact for Makefile, got %v", impacts)
	}
}

func TestDetermineImpact_README(t *testing.T) {
	tests := []string{"README.md", "readme.md", "README", "Readme.md"}

	for _, file := range tests {
		t.Run(file, func(t *testing.T) {
			impacts := DetermineImpact(file)
			if len(impacts) != 1 || impacts[0] != ImpactReadme {
				t.Errorf("expected ImpactReadme for %s, got %v", file, impacts)
			}
		})
	}
}

func TestDetermineImpact_GitHooks(t *testing.T) {
	tests := []string{
		".githooks/pre-commit",
		".husky/pre-commit",
		"lefthook.yml",
		".lefthook.yml",
	}

	for _, file := range tests {
		t.Run(file, func(t *testing.T) {
			impacts := DetermineImpact(file)
			hasDevelopment := false
			for _, imp := range impacts {
				if imp == ImpactDevelopment {
					hasDevelopment = true
				}
			}
			if !hasDevelopment {
				t.Errorf("expected ImpactDevelopment for %s, got %v", file, impacts)
			}
		})
	}
}

func TestDetermineImpact_SourceCode(t *testing.T) {
	tests := []string{
		"main.go",
		"app.ts",
		"index.js",
		"handler.py",
		"service.java",
	}

	for _, file := range tests {
		t.Run(file, func(t *testing.T) {
			impacts := DetermineImpact(file)
			hasConventions := false
			hasEndpoints := false
			for _, imp := range impacts {
				if imp == ImpactConventions {
					hasConventions = true
				}
				if imp == ImpactEndpoints {
					hasEndpoints = true
				}
			}
			if !hasConventions || !hasEndpoints {
				t.Errorf("expected conventions and endpoints impact for %s, got %v", file, impacts)
			}
		})
	}
}

func TestDetermineImpact_ConfigFiles(t *testing.T) {
	tests := []string{
		".golangci.yml",
		"tsconfig.json",
		"Dockerfile",
		".editorconfig",
	}

	for _, file := range tests {
		t.Run(file, func(t *testing.T) {
			impacts := DetermineImpact(file)
			hasConfig := false
			for _, imp := range impacts {
				if imp == ImpactConfig {
					hasConfig = true
				}
			}
			if !hasConfig {
				t.Errorf("expected ImpactConfig for %s, got %v", file, impacts)
			}
		})
	}
}

func TestDetermineImpact_GitHubWorkflows(t *testing.T) {
	impacts := DetermineImpact(".github/workflows/ci.yml")

	hasConfig := false
	for _, imp := range impacts {
		if imp == ImpactConfig {
			hasConfig = true
		}
	}
	if !hasConfig {
		t.Errorf("expected ImpactConfig for GitHub workflow, got %v", impacts)
	}
}

func TestImpactDescription(t *testing.T) {
	tests := []struct {
		impacts  []string
		expected string
	}{
		{[]string{ImpactAll}, "all sections"},
		{[]string{ImpactTechStack}, "tech stack"},
		{[]string{ImpactCommands, ImpactDevelopment}, "commands, development"},
		{[]string{}, "no sections"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := ImpactDescription(tt.impacts)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIncrementalAnalyzer_FullAnalysis(t *testing.T) {
	// Create a temp directory with minimal Go project
	tmpDir, err := os.MkdirTemp("", "incremental-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create go.mod
	goMod := `module test

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create main.go
	mainGo := `package main

func main() {
	println("hello")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	// Test full analysis
	ia := NewIncrementalAnalyzer(tmpDir)
	analysis, err := ia.AnalyzeFull()

	if err != nil {
		t.Fatalf("full analysis failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("expected analysis to not be nil")
	}

	if analysis.ProjectName == "" {
		t.Error("expected project name to be set")
	}

	// Check cache is populated
	if ia.GetCache() == nil {
		t.Error("expected cache to be populated after full analysis")
	}
}

func TestIncrementalAnalyzer_IncrementalAnalysis(t *testing.T) {
	// Create a temp directory with minimal Go project
	tmpDir, err := os.MkdirTemp("", "incremental-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create go.mod
	goMod := `module test

go 1.21
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create main.go
	mainGo := `package main

func main() {
	println("hello")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	// Do initial full analysis
	ia := NewIncrementalAnalyzer(tmpDir)
	_, err = ia.AnalyzeFull()
	if err != nil {
		t.Fatalf("initial analysis failed: %v", err)
	}

	// Now do incremental analysis for a source file change
	analysis, impacts, err := ia.AnalyzeIncremental(filepath.Join(tmpDir, "main.go"))
	if err != nil {
		t.Fatalf("incremental analysis failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("expected analysis to not be nil")
	}

	// Should have conventions and endpoints impact
	hasConventions := false
	for _, imp := range impacts {
		if imp == ImpactConventions {
			hasConventions = true
		}
	}
	if !hasConventions {
		t.Errorf("expected ImpactConventions for main.go change, got %v", impacts)
	}
}

func TestIncrementalAnalyzer_NoCacheFallsBackToFull(t *testing.T) {
	// Create a temp directory with minimal Go project
	tmpDir, err := os.MkdirTemp("", "incremental-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create go.mod
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create new analyzer without running full analysis first
	ia := NewIncrementalAnalyzer(tmpDir)

	// Incremental analysis should fall back to full when no cache
	analysis, impacts, err := ia.AnalyzeIncremental(filepath.Join(tmpDir, "main.go"))
	if err != nil {
		t.Fatalf("analysis failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("expected analysis to not be nil")
	}

	// Should have done full analysis
	if len(impacts) != 1 || impacts[0] != ImpactAll {
		t.Errorf("expected ImpactAll for no-cache scenario, got %v", impacts)
	}
}
