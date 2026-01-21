package detector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestDevelopmentDetector_Detect_GoProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-go-test")
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

	// Create Makefile with setup target
	makefile := `## Install development dependencies
setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Build the project
build:
	go build ./...
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Makefile"), []byte(makefile), 0644); err != nil {
		t.Fatalf("failed to create Makefile: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	// Check prerequisites
	hasPrereq := func(name string) bool {
		for _, p := range result.Prerequisites {
			if p.Name == name {
				return true
			}
		}
		return false
	}

	if !hasPrereq("Go") {
		t.Error("expected Go prerequisite to be detected")
	}

	// Check Go version
	for _, p := range result.Prerequisites {
		if p.Name == "Go" {
			if p.Version != "1.21+" {
				t.Errorf("expected Go version '1.21+', got '%s'", p.Version)
			}
		}
	}

	// Check setup steps
	if len(result.SetupSteps) == 0 {
		t.Error("expected setup steps to be detected")
	}

	hasSetupStep := func(cmd string) bool {
		for _, s := range result.SetupSteps {
			if s.Command == cmd {
				return true
			}
		}
		return false
	}

	if !hasSetupStep("make setup") {
		t.Error("expected 'make setup' step to be detected")
	}
}

func TestDevelopmentDetector_Detect_NodeProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-node-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create package.json with engines
	packageJSON := `{
  "name": "test",
  "engines": {
    "node": ">=18.0.0"
  }
}
`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	// Check Node.js prerequisite
	hasNode := false
	for _, p := range result.Prerequisites {
		if p.Name == "Node.js" {
			hasNode = true
			if p.Version != ">=18.0.0" {
				t.Errorf("expected Node version '>=18.0.0', got '%s'", p.Version)
			}
		}
	}

	if !hasNode {
		t.Error("expected Node.js prerequisite to be detected")
	}

	// Check inferred setup step
	hasNpmInstall := false
	for _, s := range result.SetupSteps {
		if s.Command == "npm install" {
			hasNpmInstall = true
		}
	}

	if !hasNpmInstall {
		t.Error("expected 'npm install' step to be inferred")
	}
}

func TestDevelopmentDetector_Detect_YarnProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-yarn-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create package.json and yarn.lock
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name": "test"}`), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "yarn.lock"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create yarn.lock: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	hasYarnInstall := false
	for _, s := range result.SetupSteps {
		if s.Command == "yarn install" {
			hasYarnInstall = true
		}
	}

	if !hasYarnInstall {
		t.Error("expected 'yarn install' step to be inferred")
	}
}

func TestDevelopmentDetector_Detect_PnpmProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-pnpm-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create package.json and pnpm-lock.yaml
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name": "test"}`), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "pnpm-lock.yaml"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create pnpm-lock.yaml: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	hasPnpmInstall := false
	for _, s := range result.SetupSteps {
		if s.Command == "pnpm install" {
			hasPnpmInstall = true
		}
	}

	if !hasPnpmInstall {
		t.Error("expected 'pnpm install' step to be inferred")
	}
}

func TestDevelopmentDetector_Detect_PythonProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-python-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create pyproject.toml with requires-python
	pyproject := `[project]
name = "test"
requires-python = ">=3.10"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "pyproject.toml"), []byte(pyproject), 0644); err != nil {
		t.Fatalf("failed to create pyproject.toml: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	hasPython := false
	for _, p := range result.Prerequisites {
		if p.Name == "Python" {
			hasPython = true
			if p.Version != ">=3.10" {
				t.Errorf("expected Python version '>=3.10', got '%s'", p.Version)
			}
		}
	}

	if !hasPython {
		t.Error("expected Python prerequisite to be detected")
	}
}

func TestDevelopmentDetector_Detect_DockerPrerequisite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-docker-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create Dockerfile
	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte("FROM alpine"), 0644); err != nil {
		t.Fatalf("failed to create Dockerfile: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	hasDocker := false
	for _, p := range result.Prerequisites {
		if p.Name == "Docker" {
			hasDocker = true
		}
	}

	if !hasDocker {
		t.Error("expected Docker prerequisite to be detected")
	}
}

func TestDevelopmentDetector_Detect_GitHooks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-hooks-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .githooks directory with pre-commit hook
	hooksDir := filepath.Join(tmpDir, ".githooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("failed to create .githooks dir: %v", err)
	}

	preCommit := `#!/bin/bash
gofmt -w .
golangci-lint run
go test ./...
`
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte(preCommit), 0755); err != nil {
		t.Fatalf("failed to create pre-commit: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	if len(result.GitHooks) == 0 {
		t.Fatal("expected git hooks to be detected")
	}

	// Check for pre-commit hook
	var preCommitHook *types.GitHook
	for i := range result.GitHooks {
		if result.GitHooks[i].Name == "pre-commit" {
			preCommitHook = &result.GitHooks[i]
			break
		}
	}

	if preCommitHook == nil {
		t.Fatal("expected pre-commit hook to be detected")
	}

	// Check actions
	hasAction := func(action string) bool {
		for _, a := range preCommitHook.Actions {
			if a == action {
				return true
			}
		}
		return false
	}

	if !hasAction("Format code") {
		t.Error("expected 'Format code' action")
	}
	if !hasAction("Run linter") {
		t.Error("expected 'Run linter' action")
	}
	if !hasAction("Run tests") {
		t.Error("expected 'Run tests' action")
	}
}

func TestDevelopmentDetector_Detect_HuskyHooks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-husky-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .husky directory with pre-commit hook
	huskyDir := filepath.Join(tmpDir, ".husky")
	if err := os.MkdirAll(huskyDir, 0755); err != nil {
		t.Fatalf("failed to create .husky dir: %v", err)
	}

	preCommit := `#!/bin/sh
npm run lint
npm test
`
	if err := os.WriteFile(filepath.Join(huskyDir, "pre-commit"), []byte(preCommit), 0755); err != nil {
		t.Fatalf("failed to create pre-commit: %v", err)
	}

	// Husky internal files should be ignored
	if err := os.WriteFile(filepath.Join(huskyDir, "_"), []byte(""), 0644); err != nil {
		t.Fatalf("failed to create _: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	// Should have pre-commit but not _
	hasPreCommit := false
	hasUnderscore := false
	for _, h := range result.GitHooks {
		if h.Name == "pre-commit" {
			hasPreCommit = true
		}
		if h.Name == "_" {
			hasUnderscore = true
		}
	}

	if !hasPreCommit {
		t.Error("expected pre-commit hook to be detected")
	}
	if hasUnderscore {
		t.Error("expected underscore file to be ignored")
	}
}

func TestDevelopmentDetector_Detect_LefthookConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-lefthook-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create lefthook.yml
	if err := os.WriteFile(filepath.Join(tmpDir, "lefthook.yml"), []byte("pre-commit:\n  commands:\n    lint:\n      run: npm run lint"), 0644); err != nil {
		t.Fatalf("failed to create lefthook.yml: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	hasLefthook := false
	for _, h := range result.GitHooks {
		if h.Name == "lefthook" {
			hasLefthook = true
		}
	}

	if !hasLefthook {
		t.Error("expected lefthook to be detected")
	}
}

func TestDevelopmentDetector_Detect_ToolVersions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-asdf-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .tool-versions (asdf)
	toolVersions := `nodejs 18.17.0
python 3.11.4
# comment line
ruby 3.2.2
`
	if err := os.WriteFile(filepath.Join(tmpDir, ".tool-versions"), []byte(toolVersions), 0644); err != nil {
		t.Fatalf("failed to create .tool-versions: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	// Check prerequisites from .tool-versions
	hasPrereq := func(name, version string) bool {
		for _, p := range result.Prerequisites {
			if p.Name == name && p.Version == version {
				return true
			}
		}
		return false
	}

	if !hasPrereq("Nodejs", "18.17.0") {
		t.Error("expected Nodejs prerequisite from .tool-versions")
	}
	if !hasPrereq("Python", "3.11.4") {
		t.Error("expected Python prerequisite from .tool-versions")
	}
	if !hasPrereq("Ruby", "3.2.2") {
		t.Error("expected Ruby prerequisite from .tool-versions")
	}
}

func TestDevelopmentDetector_Detect_NvmrcFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-nvmrc-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .nvmrc
	if err := os.WriteFile(filepath.Join(tmpDir, ".nvmrc"), []byte("20.10.0"), 0644); err != nil {
		t.Fatalf("failed to create .nvmrc: %v", err)
	}

	// Create package.json to trigger Node detection
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name": "test"}`), 0644); err != nil {
		t.Fatalf("failed to create package.json: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	hasNode := false
	for _, p := range result.Prerequisites {
		if p.Name == "Node.js" {
			hasNode = true
			// .nvmrc should take precedence
			if p.Version != "20.10.0" {
				t.Errorf("expected Node version '20.10.0' from .nvmrc, got '%s'", p.Version)
			}
		}
	}

	if !hasNode {
		t.Error("expected Node.js prerequisite")
	}
}

func TestDevelopmentDetector_Detect_EnvExample(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-env-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .env.example
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.example"), []byte("API_KEY=xxx"), 0644); err != nil {
		t.Fatalf("failed to create .env.example: %v", err)
	}

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result == nil {
		t.Fatal("expected development info to be detected")
	}

	hasCpEnv := false
	for _, s := range result.SetupSteps {
		if s.Command == "cp .env.example .env" {
			hasCpEnv = true
		}
	}

	if !hasCpEnv {
		t.Error("expected 'cp .env.example .env' setup step")
	}
}

func TestDevelopmentDetector_Detect_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dev-empty-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	detector := NewDevelopmentDetector(tmpDir, nil)
	result := detector.Detect()

	if result != nil {
		t.Error("expected nil for empty directory")
	}
}

func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"go", "Go"},
		{"nodejs", "Nodejs"},
		{"", ""},
		{"A", "A"},
		{"already", "Already"},
	}

	for _, tt := range tests {
		result := capitalizeFirst(tt.input)
		if result != tt.expected {
			t.Errorf("capitalizeFirst(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
