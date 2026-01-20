package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitDetector_DetectCommitConvention(t *testing.T) {
	// Create a temp directory with a git repo
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	if err := runGitCommand(tmpDir, "init"); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user for commits
	if err := runGitCommand(tmpDir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("failed to config git email: %v", err)
	}
	if err := runGitCommand(tmpDir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("failed to config git name: %v", err)
	}

	// Create some conventional commits
	commits := []string{
		"feat(api): add user endpoint",
		"fix(auth): resolve login issue",
		"docs: update readme",
		"feat(cli): add new command",
		"chore: update dependencies",
		"fix(api): handle edge case",
		"feat: add logging",
		"test: add unit tests",
	}

	for i, msg := range commits {
		// Create a file for each commit
		filePath := filepath.Join(tmpDir, "file"+string(rune('0'+i))+".txt")
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}
		if err := runGitCommand(tmpDir, "add", "."); err != nil {
			t.Fatalf("failed to git add: %v", err)
		}
		if err := runGitCommand(tmpDir, "commit", "-m", msg); err != nil {
			t.Fatalf("failed to git commit: %v", err)
		}
	}

	// Test detection
	detector := NewGitDetector(tmpDir)
	conventions := detector.Detect()

	if conventions == nil {
		t.Fatal("expected conventions to not be nil")
	}

	if conventions.CommitConvention == nil {
		t.Fatal("expected commit convention to be detected")
	}

	cc := conventions.CommitConvention
	if cc.Style != "conventional" {
		t.Errorf("expected style 'conventional', got '%s'", cc.Style)
	}

	if cc.Format == "" {
		t.Error("expected format to be set")
	}

	// Check that types were detected
	if len(cc.Types) == 0 {
		t.Error("expected types to be detected")
	}

	// Check for common types
	hasType := func(types []string, target string) bool {
		for _, t := range types {
			if t == target {
				return true
			}
		}
		return false
	}

	if !hasType(cc.Types, "feat") {
		t.Error("expected 'feat' type to be detected")
	}
	if !hasType(cc.Types, "fix") {
		t.Error("expected 'fix' type to be detected")
	}
}

func TestGitDetector_DetectBranchConvention(t *testing.T) {
	// Create a temp directory with a git repo
	tmpDir, err := os.MkdirTemp("", "git-branch-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize git repo
	if err := runGitCommand(tmpDir, "init"); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Configure git user
	if err := runGitCommand(tmpDir, "config", "user.email", "test@test.com"); err != nil {
		t.Fatalf("failed to config git email: %v", err)
	}
	if err := runGitCommand(tmpDir, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("failed to config git name: %v", err)
	}

	// Create initial commit on main
	filePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(filePath, []byte("# Test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := runGitCommand(tmpDir, "add", "."); err != nil {
		t.Fatalf("failed to git add: %v", err)
	}
	if err := runGitCommand(tmpDir, "commit", "-m", "initial commit"); err != nil {
		t.Fatalf("failed to git commit: %v", err)
	}

	// Create branches with prefixes
	branches := []string{
		"feat/user-auth",
		"feat/api-endpoints",
		"fix/login-bug",
		"fix/memory-leak",
		"chore/update-deps",
		"chore/cleanup",
	}

	for _, branch := range branches {
		if err := runGitCommand(tmpDir, "branch", branch); err != nil {
			t.Fatalf("failed to create branch %s: %v", branch, err)
		}
	}

	// Test detection
	detector := NewGitDetector(tmpDir)
	conventions := detector.Detect()

	if conventions == nil {
		t.Fatal("expected conventions to not be nil")
	}

	if conventions.BranchConvention == nil {
		t.Fatal("expected branch convention to be detected")
	}

	bc := conventions.BranchConvention
	if len(bc.Prefixes) == 0 {
		t.Error("expected prefixes to be detected")
	}

	// Check for common prefixes
	hasPrefix := func(prefixes []string, target string) bool {
		for _, p := range prefixes {
			if p == target {
				return true
			}
		}
		return false
	}

	if !hasPrefix(bc.Prefixes, "feat") {
		t.Error("expected 'feat' prefix to be detected")
	}
	if !hasPrefix(bc.Prefixes, "fix") {
		t.Error("expected 'fix' prefix to be detected")
	}
	if !hasPrefix(bc.Prefixes, "chore") {
		t.Error("expected 'chore' prefix to be detected")
	}
}

func TestGitDetector_NotGitRepo(t *testing.T) {
	// Create a temp directory without git
	tmpDir, err := os.MkdirTemp("", "no-git-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	detector := NewGitDetector(tmpDir)
	conventions := detector.Detect()

	if conventions == nil {
		t.Fatal("expected conventions to not be nil")
	}

	// Should return empty conventions for non-git directory
	if conventions.CommitConvention != nil {
		t.Error("expected no commit convention for non-git repo")
	}
	if conventions.BranchConvention != nil {
		t.Error("expected no branch convention for non-git repo")
	}
}

func TestGetTopKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		n        int
		expected []string
	}{
		{
			name:     "empty map",
			input:    map[string]int{},
			n:        3,
			expected: nil,
		},
		{
			name:     "fewer items than n",
			input:    map[string]int{"a": 5, "b": 3},
			n:        5,
			expected: []string{"a", "b"},
		},
		{
			name:     "more items than n",
			input:    map[string]int{"a": 5, "b": 10, "c": 3, "d": 8},
			n:        2,
			expected: []string{"b", "d"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTopKeys(tt.input, tt.n)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d items, got %d", len(tt.expected), len(result))
				return
			}
			// For the sorted case, check exact match
			if tt.name == "more items than n" {
				for i, v := range tt.expected {
					if result[i] != v {
						t.Errorf("expected %s at position %d, got %s", v, i, result[i])
					}
				}
			}
		})
	}
}

// runGitCommand runs a git command in the specified directory
func runGitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2024-01-01T00:00:00", "GIT_COMMITTER_DATE=2024-01-01T00:00:00")
	return cmd.Run()
}
