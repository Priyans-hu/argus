package analyzer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func TestResolveWorkspaces_GlobPatterns(t *testing.T) {
	root := t.TempDir()

	mkdirAll(t, filepath.Join(root, "packages", "a"))
	mkdirAll(t, filepath.Join(root, "packages", "b"))
	mkdirAll(t, filepath.Join(root, "apps", "web"))
	mkdirAll(t, filepath.Join(root, "apps", "api"))

	ma := NewMonorepoAnalyzer(root, false, 4)

	info := &types.MonorepoInfo{
		IsMonorepo:     true,
		WorkspacePaths: []string{"packages/*", "apps/*"},
	}

	resolved := ma.resolveWorkspaces(info)

	if len(resolved) != 4 {
		t.Fatalf("expected 4 workspaces, got %d: %v", len(resolved), resolved)
	}

	expected := map[string]bool{
		filepath.Join("packages", "a"): true,
		filepath.Join("packages", "b"): true,
		filepath.Join("apps", "web"):   true,
		filepath.Join("apps", "api"):   true,
	}

	for _, r := range resolved {
		if !expected[r] {
			t.Errorf("unexpected workspace: %s", r)
		}
	}
}

func TestResolveWorkspaces_FromPackages(t *testing.T) {
	root := t.TempDir()

	mkdirAll(t, filepath.Join(root, "apps", "web"))
	mkdirAll(t, filepath.Join(root, "apps", "api"))

	ma := NewMonorepoAnalyzer(root, false, 4)

	info := &types.MonorepoInfo{
		IsMonorepo: true,
		Packages: []types.WorkspacePackage{
			{
				Name:        "apps",
				Path:        "apps",
				SubPackages: []string{"web", "api"},
			},
		},
	}

	resolved := ma.resolveWorkspaces(info)
	if len(resolved) != 2 {
		t.Fatalf("expected 2 workspaces, got %d: %v", len(resolved), resolved)
	}
}

func TestResolveWorkspaces_Deduplication(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "packages", "shared"))

	ma := NewMonorepoAnalyzer(root, false, 4)

	info := &types.MonorepoInfo{
		IsMonorepo:     true,
		WorkspacePaths: []string{"packages/*"},
		Packages: []types.WorkspacePackage{
			{
				Name:        "packages",
				Path:        "packages",
				SubPackages: []string{"shared"},
			},
		},
	}

	resolved := ma.resolveWorkspaces(info)
	if len(resolved) != 1 {
		t.Fatalf("expected 1 workspace (deduplicated), got %d: %v", len(resolved), resolved)
	}
}

func TestResolveWorkspaces_Empty(t *testing.T) {
	root := t.TempDir()
	ma := NewMonorepoAnalyzer(root, false, 4)

	info := &types.MonorepoInfo{
		IsMonorepo: true,
	}

	resolved := ma.resolveWorkspaces(info)
	if len(resolved) != 0 {
		t.Fatalf("expected 0 workspaces, got %d", len(resolved))
	}
}

func TestAnalyzeWorkspaces_Sequential(t *testing.T) {
	root := t.TempDir()

	wsDir := filepath.Join(root, "packages", "mylib")
	mkdirAll(t, wsDir)
	writeFile(t, filepath.Join(wsDir, "main.go"), []byte("package main\n"))

	ma := NewMonorepoAnalyzer(root, false, 4)

	info := &types.MonorepoInfo{
		IsMonorepo:     true,
		WorkspacePaths: []string{"packages/*"},
	}

	results := ma.AnalyzeWorkspaces(context.Background(), info)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Error != nil {
		t.Fatalf("workspace analysis failed: %v", results[0].Error)
	}

	if results[0].Analysis == nil {
		t.Fatal("expected non-nil analysis")
	}

	if results[0].Name != "mylib" {
		t.Errorf("expected workspace name 'mylib', got '%s'", results[0].Name)
	}
}

func TestAnalyzeWorkspaces_Parallel(t *testing.T) {
	root := t.TempDir()

	for _, name := range []string{"app-a", "app-b"} {
		wsDir := filepath.Join(root, "apps", name)
		mkdirAll(t, wsDir)
		writeFile(t, filepath.Join(wsDir, "index.js"), []byte("console.log('hello');\n"))
	}

	ma := NewMonorepoAnalyzer(root, true, 2)

	info := &types.MonorepoInfo{
		IsMonorepo:     true,
		WorkspacePaths: []string{"apps/*"},
	}

	results := ma.AnalyzeWorkspaces(context.Background(), info)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Error != nil {
			t.Errorf("workspace %s failed: %v", r.Path, r.Error)
		}
		if r.Analysis == nil {
			t.Errorf("workspace %s: expected non-nil analysis", r.Path)
		}
	}
}

func TestAnalyzeWorkspaces_ContextCancellation(t *testing.T) {
	root := t.TempDir()
	mkdirAll(t, filepath.Join(root, "packages", "a"))
	writeFile(t, filepath.Join(root, "packages", "a", "main.go"), []byte("package main\n"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ma := NewMonorepoAnalyzer(root, false, 4)

	info := &types.MonorepoInfo{
		IsMonorepo:     true,
		WorkspacePaths: []string{"packages/*"},
	}

	results := ma.AnalyzeWorkspaces(ctx, info)
	if len(results) > 1 {
		t.Errorf("expected at most 1 result with cancelled context, got %d", len(results))
	}
}

func TestWorkspaceName_FromPackageJSON(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "apps", "web")
	mkdirAll(t, wsDir)

	pkgJSON, _ := json.Marshal(map[string]string{"name": "@myapp/web"})
	writeFile(t, filepath.Join(wsDir, "package.json"), pkgJSON)

	name := workspaceName(wsDir, "apps/web")
	if name != "@myapp/web" {
		t.Errorf("expected '@myapp/web', got '%s'", name)
	}
}

func TestWorkspaceName_FallbackToDir(t *testing.T) {
	root := t.TempDir()
	wsDir := filepath.Join(root, "packages", "utils")
	mkdirAll(t, wsDir)

	name := workspaceName(wsDir, "packages/utils")
	if name != "utils" {
		t.Errorf("expected 'utils', got '%s'", name)
	}
}

func TestExtractJSONField(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		field    string
		expected string
	}{
		{"simple", `{"name": "test"}`, "name", "test"},
		{"scoped", `{"name": "@scope/pkg", "version": "1.0.0"}`, "name", "@scope/pkg"},
		{"missing", `{"version": "1.0.0"}`, "name", ""},
		{"empty", `{}`, "name", ""},
		{"number value", `{"name": 123}`, "name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONField([]byte(tt.data), tt.field)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
