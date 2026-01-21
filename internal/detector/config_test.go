package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDetector_Detect_GoProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-go-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create Go project config files
	files := map[string]string{
		"go.mod":        "module test\n\ngo 1.21",
		"Makefile":      ".PHONY: build\nbuild:\n\tgo build",
		".golangci.yml": "linters:\n  enable:\n    - gofmt",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	detector := NewConfigDetector(tmpDir, nil)
	configs := detector.Detect()

	if len(configs) == 0 {
		t.Fatal("expected config files to be detected")
	}

	// Check for specific configs
	hasConfig := func(path string) bool {
		for _, c := range configs {
			if c.Path == path {
				return true
			}
		}
		return false
	}

	if !hasConfig("go.mod") {
		t.Error("expected go.mod to be detected")
	}
	if !hasConfig("Makefile") {
		t.Error("expected Makefile to be detected")
	}
	if !hasConfig(".golangci.yml") {
		t.Error("expected .golangci.yml to be detected")
	}
}

func TestConfigDetector_Detect_NodeProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-node-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create Node.js project config files
	files := map[string]string{
		"package.json":       `{"name": "test"}`,
		"tsconfig.json":      `{"compilerOptions": {}}`,
		".eslintrc.json":     `{"rules": {}}`,
		".prettierrc":        `{"semi": true}`,
		"jest.config.js":     `module.exports = {}`,
		"tailwind.config.js": `module.exports = {}`,
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	detector := NewConfigDetector(tmpDir, nil)
	configs := detector.Detect()

	expectedConfigs := []string{
		"package.json",
		"tsconfig.json",
		".eslintrc.json",
		".prettierrc",
		"jest.config.js",
		"tailwind.config.js",
	}

	for _, expected := range expectedConfigs {
		found := false
		for _, c := range configs {
			if c.Path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s to be detected", expected)
		}
	}
}

func TestConfigDetector_Detect_PythonProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-python-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create Python project config files
	files := map[string]string{
		"pyproject.toml":   `[project]\nname = "test"`,
		"requirements.txt": "flask==2.0.0",
		".flake8":          "[flake8]\nmax-line-length = 100",
		"pytest.ini":       "[pytest]",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	detector := NewConfigDetector(tmpDir, nil)
	configs := detector.Detect()

	expectedConfigs := []string{
		"pyproject.toml",
		"requirements.txt",
		".flake8",
		"pytest.ini",
	}

	for _, expected := range expectedConfigs {
		found := false
		for _, c := range configs {
			if c.Path == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %s to be detected", expected)
		}
	}
}

func TestConfigDetector_Detect_DockerProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-docker-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create Docker config files
	files := map[string]string{
		"Dockerfile":         "FROM golang:1.21",
		"docker-compose.yml": "version: '3'",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	detector := NewConfigDetector(tmpDir, nil)
	configs := detector.Detect()

	hasConfig := func(path string) bool {
		for _, c := range configs {
			if c.Path == path {
				return true
			}
		}
		return false
	}

	if !hasConfig("Dockerfile") {
		t.Error("expected Dockerfile to be detected")
	}
	if !hasConfig("docker-compose.yml") {
		t.Error("expected docker-compose.yml to be detected")
	}
}

func TestConfigDetector_Detect_GitHubWorkflows(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-github-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .github/workflows directory
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatalf("failed to create workflows dir: %v", err)
	}

	// Create a workflow file
	if err := os.WriteFile(filepath.Join(workflowsDir, "ci.yml"), []byte("name: CI"), 0644); err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	// Create dependabot.yml
	if err := os.WriteFile(filepath.Join(tmpDir, ".github", "dependabot.yml"), []byte("version: 2"), 0644); err != nil {
		t.Fatalf("failed to create dependabot.yml: %v", err)
	}

	detector := NewConfigDetector(tmpDir, nil)
	configs := detector.Detect()

	hasConfig := func(path string) bool {
		for _, c := range configs {
			if c.Path == path {
				return true
			}
		}
		return false
	}

	if !hasConfig(".github/workflows/") {
		t.Error("expected .github/workflows/ to be detected")
	}
	if !hasConfig(".github/dependabot.yml") {
		t.Error("expected .github/dependabot.yml to be detected")
	}
}

func TestConfigDetector_Detect_EmptyDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-empty-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	detector := NewConfigDetector(tmpDir, nil)
	configs := detector.Detect()

	if len(configs) != 0 {
		t.Errorf("expected no configs for empty directory, got %d", len(configs))
	}
}

func TestConfigDetector_Detect_ConfigTypes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-types-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create various config files and check their types
	testCases := []struct {
		filename     string
		expectedType string
	}{
		{"go.mod", "Go Modules"},
		{"package.json", "npm"},
		{"Makefile", "Build"},
		{".goreleaser.yml", "Release"},
		{".editorconfig", "Editor"},
		{".codecov.yml", "Coverage"},
		{"turbo.json", "Monorepo"},
	}

	for _, tc := range testCases {
		if err := os.WriteFile(filepath.Join(tmpDir, tc.filename), []byte("test"), 0644); err != nil {
			t.Fatalf("failed to create %s: %v", tc.filename, err)
		}
	}

	detector := NewConfigDetector(tmpDir, nil)
	configs := detector.Detect()

	for _, tc := range testCases {
		found := false
		for _, c := range configs {
			if c.Path == tc.filename {
				found = true
				if c.Type != tc.expectedType {
					t.Errorf("expected type %q for %s, got %q", tc.expectedType, tc.filename, c.Type)
				}
				break
			}
		}
		if !found {
			t.Errorf("expected %s to be detected", tc.filename)
		}
	}
}

func TestConfigDetector_FileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-exists-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a file and a directory
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "testdir"), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	detector := &ConfigDetector{rootPath: tmpDir}

	if !detector.fileExists("test.txt") {
		t.Error("expected fileExists to return true for existing file")
	}

	if detector.fileExists("nonexistent.txt") {
		t.Error("expected fileExists to return false for non-existent file")
	}

	if detector.fileExists("testdir") {
		t.Error("expected fileExists to return false for directory")
	}
}

func TestConfigDetector_DirExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-dir-exists-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a file and a directory
	if err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tmpDir, "testdir"), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	detector := &ConfigDetector{rootPath: tmpDir}

	if !detector.dirExists("testdir") {
		t.Error("expected dirExists to return true for existing directory")
	}

	if detector.dirExists("nonexistent") {
		t.Error("expected dirExists to return false for non-existent directory")
	}

	if detector.dirExists("test.txt") {
		t.Error("expected dirExists to return false for file")
	}
}
