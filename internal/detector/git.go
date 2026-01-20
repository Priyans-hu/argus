package detector

import (
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// GitDetector detects git conventions from repository history
type GitDetector struct {
	rootPath string
}

// NewGitDetector creates a new git detector
func NewGitDetector(rootPath string) *GitDetector {
	return &GitDetector{rootPath: rootPath}
}

// Detect analyzes git history and returns conventions
func (d *GitDetector) Detect() *types.GitConventions {
	conventions := &types.GitConventions{}

	// Check if this is a git repo
	if !d.isGitRepo() {
		return conventions
	}

	// Detect commit conventions
	conventions.CommitConvention = d.detectCommitConvention()

	// Detect branch naming conventions
	conventions.BranchConvention = d.detectBranchConvention()

	return conventions
}

// isGitRepo checks if the directory is a git repository
func (d *GitDetector) isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = d.rootPath
	err := cmd.Run()
	return err == nil
}

// detectCommitConvention analyzes commit history for patterns
func (d *GitDetector) detectCommitConvention() *types.CommitConvention {
	// Get recent commits
	cmd := exec.Command("git", "log", "--oneline", "-100", "--format=%s")
	cmd.Dir = d.rootPath
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(commits) == 0 {
		return nil
	}

	// Patterns to detect
	patterns := map[string]*regexp.Regexp{
		"conventional": regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?:\s*.+`),
		"gitmoji":      regexp.MustCompile(`^:[a-z_]+:\s*.+`),
		"jira":         regexp.MustCompile(`^[A-Z]+-\d+\s*.+`),
		"angular":      regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore)\(.+\):\s*.+`),
	}

	// Count matches for each pattern
	counts := make(map[string]int)
	typeCount := make(map[string]int)
	scopeCount := make(map[string]int)

	conventionalRegex := regexp.MustCompile(`^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(([^)]+)\))?:\s*(.+)`)

	for _, commit := range commits {
		commit = strings.TrimSpace(commit)
		if commit == "" {
			continue
		}

		for name, pattern := range patterns {
			if pattern.MatchString(commit) {
				counts[name]++
			}
		}

		// Extract types and scopes from conventional commits
		if matches := conventionalRegex.FindStringSubmatch(commit); len(matches) >= 2 {
			typeCount[matches[1]]++
			if len(matches) >= 4 && matches[3] != "" {
				scopeCount[matches[3]]++
			}
		}
	}

	// Determine the dominant convention
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

	// Build format string and examples based on detected style
	switch bestPattern {
	case "conventional", "angular":
		convention.Format = "<type>(<scope>): <description>"
		convention.Types = getTopKeys(typeCount, 7)
		convention.Scopes = getTopKeys(scopeCount, 5)

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

// detectBranchConvention analyzes branch names for patterns
func (d *GitDetector) detectBranchConvention() *types.BranchConvention {
	// Get all branches (local and remote)
	cmd := exec.Command("git", "branch", "-a", "--format=%(refname:short)")
	cmd.Dir = d.rootPath
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(branches) == 0 {
		return nil
	}

	// Common prefixes to look for
	prefixPattern := regexp.MustCompile(`^(origin/)?(feat|fix|feature|bugfix|hotfix|release|chore|docs|test|refactor|ci|build)/`)

	prefixCount := make(map[string]int)
	var matchedBranches int

	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if branch == "" || branch == "main" || branch == "master" || branch == "develop" || branch == "dev" {
			continue
		}

		// Remove origin/ prefix for counting
		branch = strings.TrimPrefix(branch, "origin/")

		if matches := prefixPattern.FindStringSubmatch(branch); len(matches) >= 3 {
			prefix := matches[2]
			// Normalize similar prefixes
			switch prefix {
			case "feature":
				prefix = "feat"
			case "bugfix", "hotfix":
				prefix = "fix"
			}
			prefixCount[prefix]++
			matchedBranches++
		}
	}

	// Need at least some branches with prefixes
	if matchedBranches < 2 {
		return nil
	}

	convention := &types.BranchConvention{
		Prefixes: getTopKeys(prefixCount, 6),
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

// getTopKeys returns the top N keys from a map sorted by count
func getTopKeys(m map[string]int, n int) []string {
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
