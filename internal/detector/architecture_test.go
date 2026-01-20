package detector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestArchitectureDetector_DetectStyle(t *testing.T) {
	tests := []struct {
		name     string
		dirs     []string
		expected string
	}{
		{
			name:     "standard go layout",
			dirs:     []string{"cmd", "internal", "pkg"},
			expected: "Standard Go Layout",
		},
		{
			name:     "clean architecture",
			dirs:     []string{"domain", "infrastructure", "adapters"},
			expected: "Clean Architecture",
		},
		{
			name:     "hexagonal architecture",
			dirs:     []string{"ports", "adapters"},
			expected: "Hexagonal Architecture",
		},
		{
			name:     "mvc pattern",
			dirs:     []string{"models", "views", "controllers"},
			expected: "MVC",
		},
		{
			name:     "feature-based",
			dirs:     []string{"features", "shared"},
			expected: "Feature-based",
		},
		{
			name:     "go package layout",
			dirs:     []string{"pkg", "lib"},
			expected: "Go Package Layout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir, err := os.MkdirTemp("", "arch-test")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Create directory structure
			var files []types.FileInfo
			for _, dir := range tt.dirs {
				dirPath := filepath.Join(tmpDir, dir)
				if err := os.MkdirAll(dirPath, 0755); err != nil {
					t.Fatalf("failed to create dir %s: %v", dir, err)
				}
				files = append(files, types.FileInfo{
					Path:  dir,
					Name:  dir,
					IsDir: true,
				})
				// Create a file in each dir
				filePath := filepath.Join(dirPath, "file.go")
				if err := os.WriteFile(filePath, []byte("package "+dir), 0644); err != nil {
					t.Fatalf("failed to create file: %v", err)
				}
			}

			detector := NewArchitectureDetector(tmpDir, files)
			info := detector.Detect()

			if info.Style != tt.expected {
				t.Errorf("expected style '%s', got '%s'", tt.expected, info.Style)
			}
		})
	}
}

func TestArchitectureDetector_DetectEntryPoint(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "entry-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create cmd/myapp/main.go structure
	cmdDir := filepath.Join(tmpDir, "cmd", "myapp")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("failed to create cmd dir: %v", err)
	}
	mainPath := filepath.Join(cmdDir, "main.go")
	if err := os.WriteFile(mainPath, []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	files := []types.FileInfo{
		{Path: "cmd", Name: "cmd", IsDir: true},
		{Path: "cmd/myapp", Name: "myapp", IsDir: true},
	}

	detector := NewArchitectureDetector(tmpDir, files)
	info := detector.Detect()

	if info.EntryPoint != "cmd/myapp/main.go" {
		t.Errorf("expected entry point 'cmd/myapp/main.go', got '%s'", info.EntryPoint)
	}
}

func TestArchitectureDetector_DetectLayers(t *testing.T) {
	// Create temp directory with internal packages
	tmpDir, err := os.MkdirTemp("", "layers-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create internal structure
	internalDirs := []string{"config", "handler", "service", "repository"}
	for _, dir := range internalDirs {
		dirPath := filepath.Join(tmpDir, "internal", dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		// Create a Go file
		filePath := filepath.Join(dirPath, dir+".go")
		if err := os.WriteFile(filePath, []byte("package "+dir), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	var files []types.FileInfo
	files = append(files, types.FileInfo{Path: "internal", Name: "internal", IsDir: true})
	for _, dir := range internalDirs {
		files = append(files, types.FileInfo{
			Path:  "internal/" + dir,
			Name:  dir,
			IsDir: true,
		})
	}

	detector := NewArchitectureDetector(tmpDir, files)
	info := detector.Detect()

	// Check layers were detected
	var internalLayer *types.ArchitectureLayer
	for i := range info.Layers {
		if info.Layers[i].Name == "internal" {
			internalLayer = &info.Layers[i]
			break
		}
	}

	if internalLayer == nil {
		t.Fatal("expected internal layer to be detected")
	}

	if len(internalLayer.Packages) != len(internalDirs) {
		t.Errorf("expected %d packages, got %d", len(internalDirs), len(internalLayer.Packages))
	}
}

func TestArchitectureDetector_GenerateDiagram(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "diagram-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create cmd and internal structure
	cmdDir := filepath.Join(tmpDir, "cmd", "app")
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		t.Fatalf("failed to create cmd dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create main.go: %v", err)
	}

	internalDirs := []string{"api", "config", "db"}
	for _, dir := range internalDirs {
		dirPath := filepath.Join(tmpDir, "internal", dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(dirPath, dir+".go"), []byte("package "+dir), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
	}

	var files []types.FileInfo
	files = append(files, types.FileInfo{Path: "cmd", Name: "cmd", IsDir: true})
	files = append(files, types.FileInfo{Path: "cmd/app", Name: "app", IsDir: true})
	files = append(files, types.FileInfo{Path: "internal", Name: "internal", IsDir: true})
	for _, dir := range internalDirs {
		files = append(files, types.FileInfo{
			Path:  "internal/" + dir,
			Name:  dir,
			IsDir: true,
		})
	}

	detector := NewArchitectureDetector(tmpDir, files)
	info := detector.Detect()

	// Check diagram was generated
	if info.Diagram == "" {
		t.Error("expected diagram to be generated")
	}

	// Check diagram contains key elements
	if !strings.Contains(info.Diagram, "```") {
		t.Error("expected diagram to be in code block")
	}

	if !strings.Contains(info.Diagram, "app") {
		t.Error("expected diagram to contain entry point name")
	}
}

func TestCenterPad(t *testing.T) {
	tests := []struct {
		input    string
		width    int
		expected string
	}{
		{"api", 7, "  api  "},
		{"config", 6, "config"},
		{"db", 10, "    db    "},
		{"longername", 5, "longe"},
	}

	for _, tt := range tests {
		result := centerPad(tt.input, tt.width)
		if result != tt.expected {
			t.Errorf("centerPad(%q, %d) = %q, expected %q", tt.input, tt.width, result, tt.expected)
		}
	}
}

func TestGroupStrings(t *testing.T) {
	tests := []struct {
		input    []string
		size     int
		expected [][]string
	}{
		{
			input:    []string{"a", "b", "c", "d", "e"},
			size:     2,
			expected: [][]string{{"a", "b"}, {"c", "d"}, {"e"}},
		},
		{
			input:    []string{"a", "b", "c"},
			size:     3,
			expected: [][]string{{"a", "b", "c"}},
		},
		{
			input:    []string{},
			size:     3,
			expected: nil,
		},
	}

	for _, tt := range tests {
		result := groupStrings(tt.input, tt.size)
		if len(result) != len(tt.expected) {
			t.Errorf("groupStrings got %d groups, expected %d", len(result), len(tt.expected))
			continue
		}
		for i := range tt.expected {
			if len(result[i]) != len(tt.expected[i]) {
				t.Errorf("group %d: got %d items, expected %d", i, len(result[i]), len(tt.expected[i]))
			}
		}
	}
}
