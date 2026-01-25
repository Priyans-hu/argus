package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParallelAnalyzer_Analyze(t *testing.T) {
	// Create a temp directory with minimal Go project
	tmpDir, err := os.MkdirTemp("", "parallel-test")
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

	// Create README.md
	readme := `# Test Project

This is a test project.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readme), 0644); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}

	// Test parallel analysis
	pa := NewParallelAnalyzer(tmpDir, nil)
	analysis, err := pa.Analyze(context.Background())

	if err != nil {
		t.Fatalf("parallel analysis failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("expected analysis to not be nil")
	}

	if analysis.ProjectName == "" {
		t.Error("expected project name to be set")
	}

	// Verify tech stack was detected
	if len(analysis.TechStack.Languages) == 0 {
		t.Error("expected at least one language to be detected")
	}

	// Verify structure was detected
	if len(analysis.Structure.Directories) == 0 && len(analysis.KeyFiles) == 0 {
		t.Error("expected structure or key files to be detected")
	}
}

func TestParallelAnalyzer_MatchesSequential(t *testing.T) {
	// Create a temp directory with Go project
	tmpDir, err := os.MkdirTemp("", "parallel-match-test")
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

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	// Run sequential analysis
	seqAnalyzer := NewAnalyzer(tmpDir, nil)
	seqAnalysis, err := seqAnalyzer.Analyze(context.Background())
	if err != nil {
		t.Fatalf("sequential analysis failed: %v", err)
	}

	// Run parallel analysis
	parAnalyzer := NewParallelAnalyzer(tmpDir, nil)
	parAnalysis, err := parAnalyzer.Analyze(context.Background())
	if err != nil {
		t.Fatalf("parallel analysis failed: %v", err)
	}

	// Compare key fields
	if seqAnalysis.ProjectName != parAnalysis.ProjectName {
		t.Errorf("project name mismatch: seq=%s, par=%s", seqAnalysis.ProjectName, parAnalysis.ProjectName)
	}

	if len(seqAnalysis.TechStack.Languages) != len(parAnalysis.TechStack.Languages) {
		t.Errorf("language count mismatch: seq=%d, par=%d",
			len(seqAnalysis.TechStack.Languages), len(parAnalysis.TechStack.Languages))
	}

	if len(seqAnalysis.TechStack.Frameworks) != len(parAnalysis.TechStack.Frameworks) {
		t.Errorf("framework count mismatch: seq=%d, par=%d",
			len(seqAnalysis.TechStack.Frameworks), len(parAnalysis.TechStack.Frameworks))
	}

	if len(seqAnalysis.Structure.Directories) != len(parAnalysis.Structure.Directories) {
		t.Errorf("directory count mismatch: seq=%d, par=%d",
			len(seqAnalysis.Structure.Directories), len(parAnalysis.Structure.Directories))
	}

	if len(seqAnalysis.KeyFiles) != len(parAnalysis.KeyFiles) {
		t.Errorf("key files count mismatch: seq=%d, par=%d",
			len(seqAnalysis.KeyFiles), len(parAnalysis.KeyFiles))
	}

	if len(seqAnalysis.Commands) != len(parAnalysis.Commands) {
		t.Errorf("commands count mismatch: seq=%d, par=%d",
			len(seqAnalysis.Commands), len(parAnalysis.Commands))
	}
}

func TestParallelAnalyzer_WithMakefile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parallel-makefile-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create go.mod
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create Makefile
	makefile := `build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Makefile"), []byte(makefile), 0644); err != nil {
		t.Fatalf("failed to create Makefile: %v", err)
	}

	pa := NewParallelAnalyzer(tmpDir, nil)
	analysis, err := pa.Analyze(context.Background())
	if err != nil {
		t.Fatalf("parallel analysis failed: %v", err)
	}

	// Should detect commands from Makefile
	if len(analysis.Commands) == 0 {
		t.Error("expected commands to be detected from Makefile")
	}
}

func TestParallelAnalyzer_WithREADME(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parallel-readme-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create go.mod
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create README
	readme := `# Test Project

This is a test project for parallel analysis.

## Features

- Feature 1
- Feature 2
`
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readme), 0644); err != nil {
		t.Fatalf("failed to create README.md: %v", err)
	}

	pa := NewParallelAnalyzer(tmpDir, nil)
	analysis, err := pa.Analyze(context.Background())
	if err != nil {
		t.Fatalf("parallel analysis failed: %v", err)
	}

	// Should detect README content
	if analysis.ReadmeContent == nil {
		t.Error("expected README content to be detected")
	}
}

func TestParallelAnalyzer_WithSubdirectories(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parallel-subdir-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create directory structure
	dirs := []string{"cmd/app", "internal/handler", "internal/service", "pkg/types"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Create go.mod
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create main.go in cmd/app
	mainGo := `package main

func main() {}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "cmd/app/main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	pa := NewParallelAnalyzer(tmpDir, nil)
	analysis, err := pa.Analyze(context.Background())
	if err != nil {
		t.Fatalf("parallel analysis failed: %v", err)
	}

	// Should detect directory structure
	if len(analysis.Structure.Directories) == 0 {
		t.Error("expected directories to be detected")
	}

	// Should detect architecture info
	if analysis.ArchitectureInfo == nil {
		t.Error("expected architecture info to be detected")
	}
}

func TestParallelAnalyzer_MultipleRuns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "parallel-multi-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create minimal project
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n\ngo 1.21"), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	pa := NewParallelAnalyzer(tmpDir, nil)

	// Run analysis multiple times to ensure no race conditions
	for i := 0; i < 5; i++ {
		analysis, err := pa.Analyze(context.Background())
		if err != nil {
			t.Fatalf("analysis run %d failed: %v", i, err)
		}
		if analysis == nil {
			t.Fatalf("analysis run %d returned nil", i)
		}
	}
}

func TestGetProjectName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/project", "project"},
		{"/single", "single"},
		{"relative/path", "path"},
		{"simple", "simple"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := getProjectName(tc.path)
			if result != tc.expected {
				t.Errorf("getProjectName(%q) = %q, expected %q", tc.path, result, tc.expected)
			}
		})
	}
}

func TestBaseFileName(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/file.go", "file.go"},
		{"file.go", "file.go"},
		{"/root", "root"},
		{"a/b/c/d", "d"},
		{"", ""},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			result := baseFileName(tc.path)
			if result != tc.expected {
				t.Errorf("baseFileName(%q) = %q, expected %q", tc.path, result, tc.expected)
			}
		})
	}
}

func BenchmarkSequentialAnalyzer(b *testing.B) {
	// Use the current directory for a more realistic benchmark
	tmpDir, err := os.MkdirTemp("", "bench-seq")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a more realistic project structure
	setupBenchmarkProject(b, tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a := NewAnalyzer(tmpDir, nil)
		_, err := a.Analyze(context.Background())
		if err != nil {
			b.Fatalf("analysis failed: %v", err)
		}
	}
}

func BenchmarkParallelAnalyzer(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-par")
	if err != nil {
		b.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a more realistic project structure
	setupBenchmarkProject(b, tmpDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pa := NewParallelAnalyzer(tmpDir, nil)
		_, err := pa.Analyze(context.Background())
		if err != nil {
			b.Fatalf("analysis failed: %v", err)
		}
	}
}

func setupBenchmarkProject(b *testing.B, tmpDir string) {
	b.Helper()

	// Create go.mod
	goMod := `module benchmark

go 1.21

require (
	github.com/spf13/cobra v1.8.0
	github.com/fsnotify/fsnotify v1.7.0
)
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		b.Fatalf("failed to create go.mod: %v", err)
	}

	// Create directories
	dirs := []string{"cmd/app", "internal/handler", "internal/service", "pkg/types"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			b.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Create main.go
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "cmd/app/main.go"), []byte(mainGo), 0644); err != nil {
		b.Fatalf("failed to create main.go: %v", err)
	}

	// Create handler.go
	handler := `package handler

import "net/http"

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "internal/handler/handler.go"), []byte(handler), 0644); err != nil {
		b.Fatalf("failed to create handler.go: %v", err)
	}

	// Create README.md
	readme := `# Benchmark Project

A test project for benchmarking.

## Getting Started

go run cmd/app/main.go
`
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte(readme), 0644); err != nil {
		b.Fatalf("failed to create README.md: %v", err)
	}

	// Create Makefile
	makefile := `build:
	go build -o app ./cmd/app

test:
	go test ./...

run:
	go run ./cmd/app
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Makefile"), []byte(makefile), 0644); err != nil {
		b.Fatalf("failed to create Makefile: %v", err)
	}
}
