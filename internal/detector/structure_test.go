package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestDetectGoCommands(t *testing.T) {
	// Create temp directory with go.mod
	tmpDir, err := os.MkdirTemp("", "go-cmd-test")
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

	// Create a test file to trigger test command detection
	testFile := `package main

import "testing"

func TestSomething(t *testing.T) {}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte(testFile), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	commands := detectGoCommands(tmpDir)

	// Should have go build, go test, go test -v, go fmt
	if len(commands) < 3 {
		t.Errorf("expected at least 3 commands, got %d", len(commands))
	}

	// Check for specific commands
	hasCommand := func(name string) bool {
		for _, cmd := range commands {
			if cmd.Name == name {
				return true
			}
		}
		return false
	}

	if !hasCommand("go build ./...") {
		t.Error("expected 'go build ./...' command")
	}
	if !hasCommand("go test ./...") {
		t.Error("expected 'go test ./...' command")
	}
	if !hasCommand("go fmt ./...") {
		t.Error("expected 'go fmt ./...' command")
	}
}

func TestParseMakefileTargets(t *testing.T) {
	makefileContent := `
.PHONY: build test clean

build:
	go build ./...

test:
	go test ./...

clean:
	rm -rf bin/

lint:
	golangci-lint run

.internal-target:
	echo "internal"
`

	commands := parseMakefileTargets(makefileContent)

	// Should detect build, test, clean, lint (not .internal-target)
	if len(commands) != 4 {
		t.Errorf("expected 4 commands, got %d", len(commands))
	}

	hasTarget := func(name string) bool {
		for _, cmd := range commands {
			if cmd.Name == "make "+name {
				return true
			}
		}
		return false
	}

	if !hasTarget("build") {
		t.Error("expected 'make build' command")
	}
	if !hasTarget("test") {
		t.Error("expected 'make test' command")
	}
	if !hasTarget("clean") {
		t.Error("expected 'make clean' command")
	}
	if !hasTarget("lint") {
		t.Error("expected 'make lint' command")
	}

	// Should not include internal targets
	for _, cmd := range commands {
		if cmd.Name == "make .internal-target" {
			t.Error("should not include internal target starting with .")
		}
	}
}

func TestDetectCargoCommands(t *testing.T) {
	commands := detectCargoCommands()

	if len(commands) != 5 {
		t.Errorf("expected 5 cargo commands, got %d", len(commands))
	}

	hasCommand := func(name string) bool {
		for _, cmd := range commands {
			if cmd.Name == name {
				return true
			}
		}
		return false
	}

	expectedCommands := []string{
		"cargo build",
		"cargo build --release",
		"cargo test",
		"cargo fmt",
		"cargo clippy",
	}

	for _, expected := range expectedCommands {
		if !hasCommand(expected) {
			t.Errorf("expected '%s' command", expected)
		}
	}
}

func TestDetectPythonCommands(t *testing.T) {
	// Create temp directory with requirements.txt
	tmpDir, err := os.MkdirTemp("", "python-cmd-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create requirements.txt with pytest
	reqContent := `flask==2.0.0
pytest==7.0.0
requests==2.28.0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte(reqContent), 0644); err != nil {
		t.Fatalf("failed to create requirements.txt: %v", err)
	}

	commands := detectPythonCommands(tmpDir)

	hasCommand := func(name string) bool {
		for _, cmd := range commands {
			if cmd.Name == name {
				return true
			}
		}
		return false
	}

	if !hasCommand("pip install -r requirements.txt") {
		t.Error("expected pip install command")
	}
	if !hasCommand("pytest") {
		t.Error("expected pytest command (pytest is in requirements.txt)")
	}
}

func TestDetectPythonCommands_WithPyproject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "python-pyproject-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create pyproject.toml
	pyproject := `[project]
name = "mypackage"
version = "1.0.0"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644); err != nil {
		t.Fatalf("failed to create pyproject.toml: %v", err)
	}

	commands := detectPythonCommands(tmpDir)

	hasCommand := func(name string) bool {
		for _, cmd := range commands {
			if cmd.Name == name {
				return true
			}
		}
		return false
	}

	if !hasCommand("pip install -e .") {
		t.Error("expected 'pip install -e .' command for pyproject.toml")
	}
}

func TestDetectCobraCommands(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cobra-cmd-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create go.mod with cobra
	goMod := `module myapp

go 1.21

require github.com/spf13/cobra v1.8.0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to create go.mod: %v", err)
	}

	// Create cmd/myapp/cmd/ structure
	cmdDir := filepath.Join(tmpDir, "cmd", "myapp", "cmd")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("failed to create cmd dir: %v", err)
	}

	// Create root.go
	rootGo := `package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "My application",
}
`
	if err := os.WriteFile(filepath.Join(cmdDir, "root.go"), []byte(rootGo), 0644); err != nil {
		t.Fatalf("failed to create root.go: %v", err)
	}

	// Create a command file
	serveGo := `package cmd

import "github.com/spf13/cobra"

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the server",
}
`
	if err := os.WriteFile(filepath.Join(cmdDir, "serve.go"), []byte(serveGo), 0644); err != nil {
		t.Fatalf("failed to create serve.go: %v", err)
	}

	// Create another command
	versionGo := `package cmd

import "github.com/spf13/cobra"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
}
`
	if err := os.WriteFile(filepath.Join(cmdDir, "version.go"), []byte(versionGo), 0644); err != nil {
		t.Fatalf("failed to create version.go: %v", err)
	}

	commands := detectCobraCommands(tmpDir)

	// Should detect serve and version (not root.go)
	if len(commands) != 2 {
		t.Errorf("expected 2 commands, got %d", len(commands))
	}

	hasCommand := func(name string) bool {
		for _, cmd := range commands {
			if cmd.Name == name {
				return true
			}
		}
		return false
	}

	if !hasCommand("myapp serve") {
		t.Error("expected 'myapp serve' command")
	}
	if !hasCommand("myapp version") {
		t.Error("expected 'myapp version' command")
	}
}

func TestParseCobraCommand(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cobra-parse-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	content := `package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Long description here",
}
`
	filePath := filepath.Join(tmpDir, "config.go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	name, desc := parseCobraCommand(filePath)

	if name != "config" {
		t.Errorf("expected name 'config', got '%s'", name)
	}
	if desc != "Manage configuration" {
		t.Errorf("expected desc 'Manage configuration', got '%s'", desc)
	}
}

func TestInferDirectoryPurpose(t *testing.T) {
	tests := []struct {
		dir      string
		expected string
	}{
		{"errors", "Error definitions"},
		{"history", "History tracking"},
		{"tui", "Terminal UI"},
		{"resolver", "Resolution logic"},
		{"internal/config", "Configuration"},
		{"src/components", "UI components"},
	}

	for _, tt := range tests {
		result := inferDirectoryPurpose(tt.dir)
		if result != tt.expected {
			t.Errorf("inferDirectoryPurpose(%q) = %q, expected %q", tt.dir, result, tt.expected)
		}
	}
}

func TestDetectKeyFiles_NoDuplicates(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "keyfiles-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create multiple README.md files (root and nested)
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Root"), 0644); err != nil {
		t.Fatalf("failed to create root README: %v", err)
	}

	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(docsDir, "README.md"), []byte("# Docs"), 0644); err != nil {
		t.Fatalf("failed to create docs README: %v", err)
	}

	files := []types.FileInfo{
		{Path: "README.md", Name: "README.md", IsDir: false},
		{Path: "docs/README.md", Name: "README.md", IsDir: false},
	}

	detector := NewStructureDetector(tmpDir, files)
	keyFiles := detector.DetectKeyFiles()

	// Should only have one README.md (root level)
	readmeCount := 0
	for _, kf := range keyFiles {
		if kf.Path == "README.md" {
			readmeCount++
		}
	}

	if readmeCount != 1 {
		t.Errorf("expected 1 README.md, got %d", readmeCount)
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"pytest", "pytest", true},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tt := range tests {
		result := containsString(tt.s, tt.substr)
		if result != tt.expected {
			t.Errorf("containsString(%q, %q) = %v, expected %v", tt.s, tt.substr, result, tt.expected)
		}
	}
}

func TestHasGoSuffix(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"main.go", true},
		{"test_test.go", true},
		{"file.txt", false},
		{".go", false},
		{"go", false},
	}

	for _, tt := range tests {
		result := hasGoSuffix(tt.name)
		if result != tt.expected {
			t.Errorf("hasGoSuffix(%q) = %v, expected %v", tt.name, result, tt.expected)
		}
	}
}

func TestHasTestSuffix(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"main_test.go", true},
		{"foo_test.go", true},
		{"main.go", false},
		{"test.go", false},
		{"_test.go", false},
	}

	for _, tt := range tests {
		result := hasTestSuffix(tt.name)
		if result != tt.expected {
			t.Errorf("hasTestSuffix(%q) = %v, expected %v", tt.name, result, tt.expected)
		}
	}
}
