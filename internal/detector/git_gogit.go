package detector

import (
	"cmp"
	"fmt"
	"maps"
	"regexp"
	"slices"
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

// detectRepository extracts git repository information from the origin
// remote, or the first remote by name when there is no origin
func (d *GitDetectorGoGit) detectRepository() *types.GitRepository {
	if d.repo == nil {
		return nil
	}

	remotes, err := d.repo.Remotes()
	if err != nil || len(remotes) == 0 {
		return nil
	}

	// Remotes come from a map, so pick by name rather than position
	slices.SortFunc(remotes, func(a, b *git.Remote) int {
		return cmp.Compare(a.Config().Name, b.Config().Name)
	})
	remote := remotes[0]
	for _, r := range remotes {
		if r.Config().Name == "origin" {
			remote = r
			break
		}
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return nil
	}
	return repositoryFromRemote(urls[0])
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

// commitStyles are checked in this order, and on equal counts the earlier
// style wins. Every angular commit is also conventional, so angular can only
// tie conventional and never wins.
var commitStyles = []struct {
	name string
	re   *regexp.Regexp
}{
	// submatches: type, scope
	{"conventional", regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(?:\(([^)]+)\))?!?:\s*.+`)},
	{"angular", regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore)\(([^)]+)\)!?:\s*.+`)},
	{"gitmoji", regexp.MustCompile(`^:[a-z0-9_]+:\s*.+`)},
	{"jira", regexp.MustCompile(`^[A-Z]+-\d+\s*.+`)},
}

// branchPrefixRegex captures the prefix of a branch such as feat/user-auth.
var branchPrefixRegex = regexp.MustCompile(`^(feat|fix|feature|bugfix|hotfix|release|chore|docs|test|refactor|ci|build)/`)

// analyzeCommitPatterns extracts commit message conventions
func analyzeCommitPatterns(commits []string) *types.CommitConvention {
	counts := make([]int, len(commitStyles))
	typeCount := make(map[string]int)
	scopeCount := make(map[string]int)

	for _, commit := range commits {
		for i, style := range commitStyles {
			m := style.re.FindStringSubmatch(commit)
			if m == nil {
				continue
			}
			counts[i]++
			if style.name == "conventional" {
				typeCount[m[1]]++
				if m[2] != "" {
					scopeCount[m[2]]++
				}
			}
		}
	}

	// Determine dominant convention
	best := 0
	for i, count := range counts {
		if count > counts[best] {
			best = i
		}
	}

	// Need at least 30% of commits to follow a pattern
	threshold := len(commits) * 30 / 100
	if counts[best] == 0 || counts[best] < threshold {
		return nil
	}

	bestPattern := commitStyles[best].name
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
	prefixCount := make(map[string]int)
	var matchedBranches int

	for _, branch := range branches {
		// Remove origin/ prefix for counting
		branch = strings.TrimPrefix(branch, "origin/")

		if m := branchPrefixRegex.FindStringSubmatch(branch); m != nil {
			// Normalize similar prefixes
			prefixCount[normalizeBranchPrefix(m[1])]++
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

// getTopKeysFromMap returns up to n keys of m, highest count first and ties
// in key order.
func getTopKeysFromMap(m map[string]int, n int) []string {
	keys := slices.Collect(maps.Keys(m))
	slices.SortFunc(keys, func(a, b string) int {
		return cmp.Or(cmp.Compare(m[b], m[a]), cmp.Compare(a, b))
	})
	if len(keys) > n {
		keys = keys[:n]
	}
	return keys
}
