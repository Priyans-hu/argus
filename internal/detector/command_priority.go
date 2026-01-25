package detector

import (
	"regexp"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// CommandCategory represents a category of commands with priority
type CommandCategory int

const (
	CategoryBuild CommandCategory = iota
	CategoryTest
	CategoryLint
	CategoryFormat
	CategoryRun
	CategoryInstall
	CategoryClean
	CategoryGenerate
	CategoryDeploy
	CategoryDocker
	CategoryDatabase
	CategoryOther
)

// categoryPriority maps categories to their priority (lower = higher priority)
var categoryPriority = map[CommandCategory]int{
	CategoryBuild:    1,
	CategoryTest:     2,
	CategoryLint:     3,
	CategoryFormat:   4,
	CategoryRun:      5,
	CategoryInstall:  6,
	CategoryClean:    7,
	CategoryGenerate: 8,
	CategoryDeploy:   9,
	CategoryDocker:   10,
	CategoryDatabase: 11,
	CategoryOther:    99,
}

// categoryPatterns maps regex patterns to categories
// These work across all languages
var categoryPatterns = []struct {
	pattern  *regexp.Regexp
	category CommandCategory
}{
	// Build commands
	{regexp.MustCompile(`(?i)^(make\s+)?(build|compile|dist|release|bundle)$`), CategoryBuild},
	{regexp.MustCompile(`(?i)^go\s+build`), CategoryBuild},
	{regexp.MustCompile(`(?i)^cargo\s+build`), CategoryBuild},
	{regexp.MustCompile(`(?i)^(npm|yarn|pnpm|bun)\s+(run\s+)?build`), CategoryBuild},
	{regexp.MustCompile(`(?i)^python\s+setup\.py\s+build`), CategoryBuild},
	{regexp.MustCompile(`(?i)^poetry\s+build`), CategoryBuild},
	{regexp.MustCompile(`(?i)^gradle\s+build`), CategoryBuild},
	{regexp.MustCompile(`(?i)^mvn\s+(compile|package)`), CategoryBuild},
	{regexp.MustCompile(`(?i)^dotnet\s+build`), CategoryBuild},

	// Test commands
	{regexp.MustCompile(`(?i)^(make\s+)?test`), CategoryTest},
	{regexp.MustCompile(`(?i)^go\s+test`), CategoryTest},
	{regexp.MustCompile(`(?i)^cargo\s+test`), CategoryTest},
	{regexp.MustCompile(`(?i)^(npm|yarn|pnpm|bun)\s+(run\s+)?test`), CategoryTest},
	{regexp.MustCompile(`(?i)^pytest`), CategoryTest},
	{regexp.MustCompile(`(?i)^python\s+-m\s+(pytest|unittest)`), CategoryTest},
	{regexp.MustCompile(`(?i)^poetry\s+run\s+(pytest|python\s+-m\s+pytest)`), CategoryTest},
	{regexp.MustCompile(`(?i)^jest`), CategoryTest},
	{regexp.MustCompile(`(?i)^vitest`), CategoryTest},
	{regexp.MustCompile(`(?i)^gradle\s+test`), CategoryTest},
	{regexp.MustCompile(`(?i)^mvn\s+test`), CategoryTest},
	{regexp.MustCompile(`(?i)^dotnet\s+test`), CategoryTest},
	{regexp.MustCompile(`(?i)^rspec`), CategoryTest},
	{regexp.MustCompile(`(?i)^bundle\s+exec\s+rspec`), CategoryTest},
	{regexp.MustCompile(`(?i)coverage`), CategoryTest},

	// Lint commands
	{regexp.MustCompile(`(?i)^(make\s+)?lint`), CategoryLint},
	{regexp.MustCompile(`(?i)^golangci-lint`), CategoryLint},
	{regexp.MustCompile(`(?i)^cargo\s+clippy`), CategoryLint},
	{regexp.MustCompile(`(?i)^(npm|yarn|pnpm|bun)\s+(run\s+)?lint`), CategoryLint},
	{regexp.MustCompile(`(?i)^eslint`), CategoryLint},
	{regexp.MustCompile(`(?i)^(ruff|flake8|pylint|mypy)\s+check`), CategoryLint},
	{regexp.MustCompile(`(?i)^poetry\s+run\s+(ruff|flake8|pylint|mypy)`), CategoryLint},
	{regexp.MustCompile(`(?i)^rubocop`), CategoryLint},
	{regexp.MustCompile(`(?i)^check`), CategoryLint},

	// Format commands
	{regexp.MustCompile(`(?i)^(make\s+)?(format|fmt)$`), CategoryFormat},
	{regexp.MustCompile(`(?i)^go\s+fmt`), CategoryFormat},
	{regexp.MustCompile(`(?i)^gofmt`), CategoryFormat},
	{regexp.MustCompile(`(?i)^goimports`), CategoryFormat},
	{regexp.MustCompile(`(?i)^cargo\s+fmt`), CategoryFormat},
	{regexp.MustCompile(`(?i)^(npm|yarn|pnpm|bun)\s+(run\s+)?format`), CategoryFormat},
	{regexp.MustCompile(`(?i)^prettier`), CategoryFormat},
	{regexp.MustCompile(`(?i)^(black|ruff\s+format|autopep8|yapf)`), CategoryFormat},
	{regexp.MustCompile(`(?i)^poetry\s+run\s+(black|ruff\s+format)`), CategoryFormat},

	// Run/Dev commands
	{regexp.MustCompile(`(?i)^(make\s+)?(run|start|serve|dev)$`), CategoryRun},
	{regexp.MustCompile(`(?i)^go\s+run`), CategoryRun},
	{regexp.MustCompile(`(?i)^cargo\s+run`), CategoryRun},
	{regexp.MustCompile(`(?i)^(npm|yarn|pnpm|bun)\s+(run\s+)?(start|dev|serve)`), CategoryRun},
	{regexp.MustCompile(`(?i)^python\s+(app|main|run|manage)\.py`), CategoryRun},
	{regexp.MustCompile(`(?i)^(flask|uvicorn|gunicorn|django)`), CategoryRun},
	{regexp.MustCompile(`(?i)^poetry\s+run\s+(python|flask|uvicorn)`), CategoryRun},
	{regexp.MustCompile(`(?i)^rails\s+s`), CategoryRun},
	{regexp.MustCompile(`(?i)^bundle\s+exec\s+rails`), CategoryRun},
	{regexp.MustCompile(`(?i)runserver`), CategoryRun},

	// Install commands
	{regexp.MustCompile(`(?i)^(make\s+)?install$`), CategoryInstall},
	{regexp.MustCompile(`(?i)^(npm|yarn|pnpm|bun)\s+install`), CategoryInstall},
	{regexp.MustCompile(`(?i)^(npm|yarn|pnpm|bun)\s+ci`), CategoryInstall},
	{regexp.MustCompile(`(?i)^pip\s+install`), CategoryInstall},
	{regexp.MustCompile(`(?i)^poetry\s+install`), CategoryInstall},
	{regexp.MustCompile(`(?i)^cargo\s+install`), CategoryInstall},
	{regexp.MustCompile(`(?i)^bundle\s+install`), CategoryInstall},
	{regexp.MustCompile(`(?i)^go\s+mod\s+(download|tidy)`), CategoryInstall},
	{regexp.MustCompile(`(?i)^setup`), CategoryInstall},

	// Clean commands
	{regexp.MustCompile(`(?i)^(make\s+)?clean`), CategoryClean},
	{regexp.MustCompile(`(?i)^cargo\s+clean`), CategoryClean},
	{regexp.MustCompile(`(?i)^go\s+clean`), CategoryClean},
	{regexp.MustCompile(`(?i)^(npm|yarn|pnpm|bun)\s+(run\s+)?clean`), CategoryClean},

	// Generate commands
	{regexp.MustCompile(`(?i)^(make\s+)?(generate|gen|codegen|proto)`), CategoryGenerate},
	{regexp.MustCompile(`(?i)^go\s+generate`), CategoryGenerate},
	{regexp.MustCompile(`(?i)^protoc`), CategoryGenerate},

	// Deploy commands
	{regexp.MustCompile(`(?i)^(make\s+)?deploy`), CategoryDeploy},
	{regexp.MustCompile(`(?i)^(npm|yarn|pnpm|bun)\s+(run\s+)?deploy`), CategoryDeploy},
	{regexp.MustCompile(`(?i)^kubectl`), CategoryDeploy},
	{regexp.MustCompile(`(?i)^helm`), CategoryDeploy},
	{regexp.MustCompile(`(?i)^terraform`), CategoryDeploy},

	// Docker commands
	{regexp.MustCompile(`(?i)^(make\s+)?docker`), CategoryDocker},
	{regexp.MustCompile(`(?i)^docker\s+(build|compose|run)`), CategoryDocker},

	// Database commands
	{regexp.MustCompile(`(?i)^(make\s+)?migrate`), CategoryDatabase},
	{regexp.MustCompile(`(?i)^(make\s+)?seed`), CategoryDatabase},
	{regexp.MustCompile(`(?i)migrations`), CategoryDatabase},
	{regexp.MustCompile(`(?i)^(npx\s+)?prisma`), CategoryDatabase},
	{regexp.MustCompile(`(?i)^alembic`), CategoryDatabase},
	{regexp.MustCompile(`(?i)^rails\s+db:`), CategoryDatabase},
}

// PrioritizedCommand extends Command with priority info
type PrioritizedCommand struct {
	types.Command
	Category CommandCategory
	Priority int
}

// categorizeCommand determines the category of a command
func categorizeCommand(cmd types.Command) CommandCategory {
	// Check command name against patterns
	cmdName := strings.TrimSpace(cmd.Name)

	for _, cp := range categoryPatterns {
		if cp.pattern.MatchString(cmdName) {
			return cp.category
		}
	}

	// Check description for hints
	desc := strings.ToLower(cmd.Description)
	if strings.Contains(desc, "build") || strings.Contains(desc, "compile") {
		return CategoryBuild
	}
	if strings.Contains(desc, "test") {
		return CategoryTest
	}
	if strings.Contains(desc, "lint") || strings.Contains(desc, "check") {
		return CategoryLint
	}
	if strings.Contains(desc, "format") {
		return CategoryFormat
	}
	if strings.Contains(desc, "install") || strings.Contains(desc, "dependencies") {
		return CategoryInstall
	}
	if strings.Contains(desc, "clean") {
		return CategoryClean
	}
	if strings.Contains(desc, "run") || strings.Contains(desc, "start") || strings.Contains(desc, "dev") {
		return CategoryRun
	}

	return CategoryOther
}

// PrioritizeCommands sorts commands by category priority and removes duplicates
func PrioritizeCommands(commands []types.Command) []types.Command {
	if len(commands) == 0 {
		return commands
	}

	// Categorize all commands
	prioritized := make([]PrioritizedCommand, 0, len(commands))
	for _, cmd := range commands {
		cat := categorizeCommand(cmd)
		prioritized = append(prioritized, PrioritizedCommand{
			Command:  cmd,
			Category: cat,
			Priority: categoryPriority[cat],
		})
	}

	// Sort by priority, then alphabetically within same priority
	sort.SliceStable(prioritized, func(i, j int) bool {
		if prioritized[i].Priority != prioritized[j].Priority {
			return prioritized[i].Priority < prioritized[j].Priority
		}
		return prioritized[i].Name < prioritized[j].Name
	})

	// Remove duplicates (keep first occurrence which has higher priority)
	seen := make(map[string]bool)
	result := make([]types.Command, 0, len(prioritized))

	for _, pc := range prioritized {
		// Normalize command for dedup
		normalized := normalizeCommand(pc.Name)
		if seen[normalized] {
			continue
		}
		seen[normalized] = true
		result = append(result, pc.Command)
	}

	return result
}

// GetQuickReferenceCommands returns the top N most important commands
func GetQuickReferenceCommands(commands []types.Command, maxCommands int) []types.Command {
	prioritized := PrioritizeCommands(commands)

	// Take top N commands, ensuring we get diversity across categories
	if len(prioritized) <= maxCommands {
		return prioritized
	}

	// Ensure at least one from each important category if available
	result := make([]types.Command, 0, maxCommands)
	categoryCount := make(map[CommandCategory]int)
	importantCategories := []CommandCategory{
		CategoryBuild, CategoryTest, CategoryLint, CategoryFormat, CategoryRun, CategoryInstall,
	}

	// First pass: get one from each important category
	for _, cmd := range prioritized {
		cat := categorizeCommand(cmd)
		for _, ic := range importantCategories {
			if cat == ic && categoryCount[cat] == 0 {
				result = append(result, cmd)
				categoryCount[cat]++
				break
			}
		}
		if len(result) >= maxCommands {
			break
		}
	}

	// Second pass: fill remaining slots with highest priority commands not yet added
	added := make(map[string]bool)
	for _, cmd := range result {
		added[normalizeCommand(cmd.Name)] = true
	}

	for _, cmd := range prioritized {
		if len(result) >= maxCommands {
			break
		}
		normalized := normalizeCommand(cmd.Name)
		if !added[normalized] {
			result = append(result, cmd)
			added[normalized] = true
		}
	}

	return result
}

// normalizeCommand normalizes a command string for deduplication
func normalizeCommand(cmd string) string {
	// Remove common variations
	cmd = strings.ToLower(cmd)
	cmd = strings.TrimSpace(cmd)

	// Remove "(in subdir)" suffixes
	if idx := strings.Index(cmd, " (in "); idx > 0 {
		cmd = cmd[:idx]
	}

	// Normalize package managers
	cmd = strings.ReplaceAll(cmd, "yarn ", "npm ")
	cmd = strings.ReplaceAll(cmd, "pnpm ", "npm ")
	cmd = strings.ReplaceAll(cmd, "bun ", "npm ")

	// Remove "run" from npm run commands
	cmd = strings.ReplaceAll(cmd, "npm run ", "npm ")

	return cmd
}

// GetCategoryName returns a human-readable category name
func GetCategoryName(cat CommandCategory) string {
	names := map[CommandCategory]string{
		CategoryBuild:    "Build",
		CategoryTest:     "Test",
		CategoryLint:     "Lint",
		CategoryFormat:   "Format",
		CategoryRun:      "Run",
		CategoryInstall:  "Setup",
		CategoryClean:    "Clean",
		CategoryGenerate: "Generate",
		CategoryDeploy:   "Deploy",
		CategoryDocker:   "Docker",
		CategoryDatabase: "Database",
		CategoryOther:    "Other",
	}
	if name, ok := names[cat]; ok {
		return name
	}
	return "Other"
}

// GroupCommandsByCategory groups commands by their category
func GroupCommandsByCategory(commands []types.Command) map[string][]types.Command {
	groups := make(map[string][]types.Command)

	for _, cmd := range commands {
		cat := categorizeCommand(cmd)
		catName := GetCategoryName(cat)
		groups[catName] = append(groups[catName], cmd)
	}

	return groups
}
