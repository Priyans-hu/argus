package detector

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestGitDetector_DetectCommitConvention(t *testing.T) {
	// Create a temp directory with a git repo
	tmpDir, err := os.MkdirTemp("", "git-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

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
	detector := NewGitDetectorGoGit(tmpDir)
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
	defer func() { _ = os.RemoveAll(tmpDir) }()

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
	detector := NewGitDetectorGoGit(tmpDir)
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
	defer func() { _ = os.RemoveAll(tmpDir) }()

	detector := NewGitDetectorGoGit(tmpDir)
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

func TestGetTopKeysFromMap(t *testing.T) {
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
		{
			name:     "ties sort by key",
			input:    map[string]int{"fix": 2, "docs": 1, "feat": 2, "chore": 1, "ci": 1},
			n:        4,
			expected: []string{"feat", "fix", "chore", "ci"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range 20 {
				if result := getTopKeysFromMap(tt.input, tt.n); !reflect.DeepEqual(result, tt.expected) {
					t.Fatalf("getTopKeysFromMap() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestAnalyzeCommitPatterns(t *testing.T) {
	conventional := func(commitTypes, scopes []string, example string) *types.CommitConvention {
		return &types.CommitConvention{
			Style: "conventional", Format: "<type>(<scope>): <description>",
			Types: commitTypes, Scopes: scopes, Example: example,
		}
	}
	gitmoji := &types.CommitConvention{Style: "gitmoji", Format: ":<emoji>: <description>", Example: ":sparkles: add new feature"}
	jira := &types.CommitConvention{Style: "jira", Format: "<TICKET-ID> <description>", Example: "PROJ-123 add new feature"}

	tests := []struct {
		name    string
		commits []string
		want    *types.CommitConvention
	}{
		{
			name:    "scoped conventional",
			commits: []string{"feat(api): add users", "fix(auth): handle expiry", "feat(cli): add flag", "fix(api): check nil"},
			want:    conventional([]string{"feat", "fix"}, []string{"api", "auth", "cli"}, "feat(api): add new feature"),
		},
		{
			name:    "unscoped conventional",
			commits: []string{"fix: handle nil", "docs: update readme", "feat: add search"},
			want:    conventional([]string{"docs", "feat", "fix"}, nil, "docs: add new feature"),
		},
		{
			name:    "breaking change marker",
			commits: []string{"feat(api)!: drop v1 endpoints", "refactor!: rename config keys"},
			want:    conventional([]string{"feat", "refactor"}, []string{"api"}, "feat(api): add new feature"),
		},
		{
			name:    "gitmoji",
			commits: []string{":sparkles: add search", ":bug: fix crash", ":pencil2: fix typo"},
			want:    gitmoji,
		},
		{
			name:    "jira",
			commits: []string{"PROJ-12 add search", "PROJ-13 fix crash", "OPS-7 bump image"},
			want:    jira,
		},
		{
			name: "merge commits mixed in",
			commits: []string{
				"Merge pull request #12 from org/feat/search", "feat(search): add index",
				"Merge branch 'main' into feat/search", "fix(search): escape query",
			},
			want: conventional([]string{"feat", "fix"}, []string{"search"}, "feat(search): add new feature"),
		},
		{
			name:    "conventional wins its tie with angular",
			commits: []string{"feat(ui): add button", "fix(ui): align label"},
			want:    conventional([]string{"feat", "fix"}, []string{"ui"}, "feat(ui): add new feature"),
		},
		{
			name:    "equal counts go to the earlier style",
			commits: []string{"PROJ-1 add search", ":sparkles: add search", "PROJ-2 fix crash", ":bug: fix crash"},
			want:    gitmoji,
		},
		{
			name: "below the 30% threshold",
			commits: []string{
				"feat: add search", "fix: crash", "update readme", "wip", "more work",
				"tweak", "cleanup", "bump", "try again", "final",
			},
			want: nil,
		},
		{
			name:    "no style in a short history",
			commits: []string{"initial commit", "wip"},
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range 20 {
				if got := analyzeCommitPatterns(tt.commits); !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("analyzeCommitPatterns() = %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func TestAnalyzeBranchPatterns(t *testing.T) {
	branches := []string{"feature/search", "origin/fix/crash", "bugfix/login", "chore/deps", "release/v1", "spike-idea", "origin/HEAD"}
	want := &types.BranchConvention{
		Prefixes: []string{"fix", "chore", "feat", "release"},
		Format:   "<prefix>/<description>",
		Examples: []string{"fix/user-auth", "chore/login-bug", "feat/update-deps"},
	}
	for range 20 {
		if got := analyzeBranchPatterns(branches); !reflect.DeepEqual(got, want) {
			t.Fatalf("analyzeBranchPatterns() = %+v, want %+v", got, want)
		}
	}
}

// runGitCommand runs a git command in the specified directory, ignoring any
// hooks or signing setup in the user's git config
func runGitCommand(dir string, args ...string) error {
	args = append([]string{"-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false"}, args...)
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2024-01-01T00:00:00", "GIT_COMMITTER_DATE=2024-01-01T00:00:00")
	return cmd.Run()
}
