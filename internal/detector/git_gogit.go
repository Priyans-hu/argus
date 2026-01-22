package detector

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitDetectorGoGit detects git conventions using go-git library
// This replaces shell-based git commands with pure Go implementation
type GitDetectorGoGit struct {
	rootPath string
	repo     *git.Repository
}

// NewGitDetectorGoGit creates a new go-git based detector
func NewGitDetectorGoGit(rootPath string) *GitDetectorGoGit {
	return &GitDetectorGoGit{rootPath: rootPath}
}

// Detect analyzes git history and returns conventions
func (d *GitDetectorGoGit) Detect() *types.GitConventions {
	conventions := &types.GitConventions{}

	// Try to open the repository
	repo, err := git.PlainOpen(d.rootPath)
	if err != nil {
		// Not a git repository or can't access
		return conventions
	}
	d.repo = repo

	// Detect commit conventions
	conventions.CommitConvention = d.detectCommitConvention()

	// Detect branch naming conventions
	conventions.BranchConvention = d.detectBranchConvention()

	// Extract repository information
	conventions.Repository = d.detectRepository()

	// Get recent commits
	conventions.RecentCommits = d.getRecentCommits(10)

	return conventions
}

// detectRepository extracts git repository information
func (d *GitDetectorGoGit) detectRepository() *types.GitRepository {
	if d.repo == nil {
		return nil
	}

	// Get remote URL
	remotes, err := d.repo.Remotes()
	if err != nil || len(remotes) == 0 {
		return nil
	}

	// Use origin remote, or first available
	var remoteURL string
	for _, remote := range remotes {
		if remote.Config().Name == "origin" {
			if len(remote.Config().URLs) > 0 {
				remoteURL = remote.Config().URLs[0]
			}
			break
		}
	}

	// Fallback to first remote if origin not found
	if remoteURL == "" && len(remotes) > 0 {
		if len(remotes[0].Config().URLs) > 0 {
			remoteURL = remotes[0].Config().URLs[0]
		}
	}

	if remoteURL == "" {
		return nil
	}

	repo := &types.GitRepository{
		RemoteURL: remoteURL,
	}

	// Parse URL to extract owner, name, and platform
	var owner, name, platform string

	// Remove .git suffix
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// Detect platform
	if strings.Contains(remoteURL, "github") {
		platform = "github"
	} else if strings.Contains(remoteURL, "gitlab") {
		platform = "gitlab"
	} else if strings.Contains(remoteURL, "bitbucket") {
		platform = "bitbucket"
	}

	// Parse HTTPS URL
	if strings.HasPrefix(remoteURL, "http") {
		parts := strings.Split(remoteURL, "/")
		if len(parts) >= 2 {
			name = parts[len(parts)-1]
			owner = parts[len(parts)-2]
		}
	} else if strings.Contains(remoteURL, "@") {
		// Parse SSH URL (git@host:owner/repo)
		parts := strings.Split(remoteURL, ":")
		if len(parts) >= 2 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 2 {
				owner = pathParts[0]
				name = pathParts[1]
			}
		}
	}

	repo.Owner = owner
	repo.Name = name
	repo.Platform = platform

	return repo
}

// getRecentCommits retrieves recent commit history
func (d *GitDetectorGoGit) getRecentCommits(limit int) []types.GitCommit {
	if d.repo == nil {
		return nil
	}

	// Get HEAD reference
	ref, err := d.repo.Head()
	if err != nil {
		return nil
	}

	// Get commit history
	commitIter, err := d.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil
	}
	defer commitIter.Close()

	var commits []types.GitCommit
	count := 0

	_ = commitIter.ForEach(func(c *object.Commit) error {
		if count >= limit {
			return fmt.Errorf("limit reached")
		}

		commits = append(commits, types.GitCommit{
			Hash:    c.Hash.String(),
			Message: strings.Split(c.Message, "\n")[0], // First line only
			Author:  c.Author.Name,
			Date:    c.Author.When.Format("2006-01-02 15:04:05 -0700"),
		})

		count++
		return nil
	})

	return commits
}

// detectCommitConvention analyzes commit history for patterns
func (d *GitDetectorGoGit) detectCommitConvention() *types.CommitConvention {
	if d.repo == nil {
		return nil
	}

	// Get HEAD reference
	ref, err := d.repo.Head()
	if err != nil {
		return nil
	}

	// Get commit history (last 100 commits)
	commitIter, err := d.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil
	}
	defer commitIter.Close()

	var commits []string
	count := 0
	maxCommits := 100

	_ = commitIter.ForEach(func(c *object.Commit) error {
		if count >= maxCommits {
			return fmt.Errorf("limit reached")
		}
		// Get first line of commit message
		message := strings.Split(c.Message, "\n")[0]
		commits = append(commits, strings.TrimSpace(message))
		count++
		return nil
	})

	if len(commits) == 0 {
		return nil
	}

	// Analyze patterns (reuse existing pattern detection logic)
	return analyzeCommitPatterns(commits)
}

// detectBranchConvention analyzes branch names for patterns
func (d *GitDetectorGoGit) detectBranchConvention() *types.BranchConvention {
	if d.repo == nil {
		return nil
	}

	// Get all branches (local and remote)
	refs, err := d.repo.References()
	if err != nil {
		return nil
	}
	defer refs.Close()

	var branches []string
	_ = refs.ForEach(func(ref *plumbing.Reference) error {
		// Only process branch references
		if ref.Name().IsBranch() || ref.Name().IsRemote() {
			branchName := ref.Name().Short()
			// Skip main/master/develop branches
			if branchName != "main" && branchName != "master" &&
				branchName != "develop" && branchName != "dev" {
				branches = append(branches, branchName)
			}
		}
		return nil
	})

	if len(branches) == 0 {
		return nil
	}

	// Analyze patterns (reuse existing pattern detection logic)
	return analyzeBranchPatterns(branches)
}

// analyzeCommitPatterns extracts commit message conventions
// This is extracted from the original implementation for reuse
func analyzeCommitPatterns(commits []string) *types.CommitConvention {
	// Pattern matching logic (from original git.go)
	conventionalRegex := `^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(([^)]+)\))?:\s*(.+)`
	gitmojiRegex := `^:[a-z_]+:\s*.+`
	jiraRegex := `^[A-Z]+-\d+\s*.+`
	angularRegex := `^(feat|fix|docs|style|refactor|test|chore)\(.+\):\s*.+`

	counts := make(map[string]int)
	typeCount := make(map[string]int)
	scopeCount := make(map[string]int)

	for _, commit := range commits {
		// Check patterns
		if matched, _ := matchPattern(commit, conventionalRegex); matched {
			counts["conventional"]++
			// Extract type and scope
			if parts := extractConventionalParts(commit); len(parts) >= 2 {
				typeCount[parts[1]]++
				if len(parts) >= 4 && parts[3] != "" {
					scopeCount[parts[3]]++
				}
			}
		}
		if matched, _ := matchPattern(commit, gitmojiRegex); matched {
			counts["gitmoji"]++
		}
		if matched, _ := matchPattern(commit, jiraRegex); matched {
			counts["jira"]++
		}
		if matched, _ := matchPattern(commit, angularRegex); matched {
			counts["angular"]++
		}
	}

	// Determine dominant convention
	var bestPattern string
	var bestCount int
	for name, count := range counts {
		if count > bestCount {
			bestCount = count
			bestPattern = name
		}
	}

	// Need at least 30% of commits to follow a pattern
	threshold := len(commits) * 30 / 100
	if bestCount < threshold {
		return nil
	}

	convention := &types.CommitConvention{
		Style: bestPattern,
	}

	// Build format string and examples
	switch bestPattern {
	case "conventional", "angular":
		convention.Format = "<type>(<scope>): <description>"
		convention.Types = getTopKeysFromMap(typeCount, 7)
		convention.Scopes = getTopKeysFromMap(scopeCount, 5)

		// Generate example
		if len(convention.Types) > 0 {
			exampleType := convention.Types[0]
			if len(convention.Scopes) > 0 {
				convention.Example = exampleType + "(" + convention.Scopes[0] + "): add new feature"
			} else {
				convention.Example = exampleType + ": add new feature"
			}
		}

	case "gitmoji":
		convention.Format = ":<emoji>: <description>"
		convention.Example = ":sparkles: add new feature"

	case "jira":
		convention.Format = "<TICKET-ID> <description>"
		convention.Example = "PROJ-123 add new feature"
	}

	return convention
}

// analyzeBranchPatterns extracts branch naming conventions
func analyzeBranchPatterns(branches []string) *types.BranchConvention {
	prefixPattern := `^(feat|fix|feature|bugfix|hotfix|release|chore|docs|test|refactor|ci|build)/`

	prefixCount := make(map[string]int)
	var matchedBranches int

	for _, branch := range branches {
		// Remove origin/ prefix for counting
		branch = strings.TrimPrefix(branch, "origin/")

		if matched, prefix := matchPattern(branch, prefixPattern); matched && prefix != "" {
			// Normalize similar prefixes
			normalizedPrefix := normalizeBranchPrefix(prefix)
			prefixCount[normalizedPrefix]++
			matchedBranches++
		}
	}

	// Need at least 2 branches with prefixes
	if matchedBranches < 2 {
		return nil
	}

	convention := &types.BranchConvention{
		Prefixes: getTopKeysFromMap(prefixCount, 6),
		Format:   "<prefix>/<description>",
	}

	// Generate examples
	if len(convention.Prefixes) > 0 {
		convention.Examples = []string{}
		exampleDescs := []string{"user-auth", "login-bug", "update-deps"}
		for i, prefix := range convention.Prefixes {
			if i >= 3 {
				break
			}
			descIdx := i % len(exampleDescs)
			convention.Examples = append(convention.Examples, prefix+"/"+exampleDescs[descIdx])
		}
	}

	return convention
}

// Helper functions

func matchPattern(text, pattern string) (bool, string) {
	// Simple pattern matching - can be enhanced with regexp
	if strings.Contains(pattern, "(feat|fix|") {
		// Branch pattern
		for _, prefix := range []string{"feat/", "fix/", "feature/", "bugfix/", "hotfix/", "release/", "chore/", "docs/", "test/", "refactor/", "ci/", "build/"} {
			if strings.HasPrefix(text, prefix) {
				return true, strings.TrimSuffix(prefix, "/")
			}
		}
	}
	// Commit patterns - simple check
	if strings.HasPrefix(text, "feat:") || strings.HasPrefix(text, "fix:") ||
		strings.HasPrefix(text, "docs:") || strings.HasPrefix(text, "chore:") {
		return true, ""
	}
	return false, ""
}

func extractConventionalParts(commit string) []string {
	// Extract type(scope): message
	parts := strings.SplitN(commit, ":", 2)
	if len(parts) != 2 {
		return nil
	}

	typeScope := parts[0]
	result := []string{"", "", "", ""}

	// Check for scope
	if strings.Contains(typeScope, "(") && strings.Contains(typeScope, ")") {
		startScope := strings.Index(typeScope, "(")
		endScope := strings.Index(typeScope, ")")
		result[1] = typeScope[:startScope]             // type
		result[3] = typeScope[startScope+1 : endScope] // scope
	} else {
		result[1] = typeScope // type only
	}

	return result
}

func normalizeBranchPrefix(prefix string) string {
	switch prefix {
	case "feature":
		return "feat"
	case "bugfix", "hotfix":
		return "fix"
	default:
		return prefix
	}
}

func getTopKeysFromMap(m map[string]int, n int) []string {
	type kv struct {
		Key   string
		Value int
	}

	var sorted []kv
	for k, v := range m {
		sorted = append(sorted, kv{k, v})
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Value > sorted[j].Value
	})

	var result []string
	for i, item := range sorted {
		if i >= n {
			break
		}
		result = append(result, item.Key)
	}

	return result
}
