package generator

import (
	"fmt"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// generateSkills creates skill files based on detected commands
// Skills are the new format replacing commands in Claude Code
func (g *ClaudeCodeGenerator) generateSkills(analysis *types.Analysis) []types.GeneratedFile {
	var files []types.GeneratedFile

	// Build context for context-aware generation
	ctx := BuildContext(analysis)

	// Track which skills we've generated to avoid duplicates
	generated := make(map[string]bool)

	// Generate skills from detected commands
	for _, cmd := range analysis.Commands {
		skillType := classifyCommand(cmd.Name, cmd.Command)
		if skillType != "" && !generated[skillType] {
			generated[skillType] = true
			file := generateSkillFile(skillType, cmd, analysis, ctx)
			if file != nil {
				files = append(files, *file)
			}
		}
	}

	// Add framework-specific skills
	frameworkSkills := g.generateFrameworkSkills(analysis, ctx, generated)
	files = append(files, frameworkSkills...)

	// Add project-specific tool skills
	projectToolSkills := g.generateProjectToolSkills(analysis)
	files = append(files, projectToolSkills...)

	return files
}

// classifyCommand determines the skill type from its name or content
func classifyCommand(name, command string) string {
	nameLower := strings.ToLower(name)
	cmdLower := strings.ToLower(command)

	// Build commands
	if strings.Contains(nameLower, "build") || strings.Contains(cmdLower, "build") {
		if !strings.Contains(nameLower, "docker") {
			return "build"
		}
	}

	// Test commands
	if strings.Contains(nameLower, "test") || strings.Contains(cmdLower, "test") {
		return "test"
	}

	// Lint commands
	if strings.Contains(nameLower, "lint") || strings.Contains(cmdLower, "lint") ||
		strings.Contains(cmdLower, "eslint") || strings.Contains(cmdLower, "golangci") {
		return "lint"
	}

	// Dev/start commands
	if nameLower == "dev" || nameLower == "start" || strings.Contains(nameLower, "serve") {
		return "dev"
	}

	// Format commands
	if strings.Contains(nameLower, "fmt") || strings.Contains(nameLower, "format") ||
		strings.Contains(cmdLower, "prettier") || strings.Contains(cmdLower, "gofmt") {
		return "format"
	}

	// Database migration commands
	if strings.Contains(nameLower, "migrate") || strings.Contains(nameLower, "db:") ||
		strings.Contains(cmdLower, "migrate") {
		return "db-migrate"
	}

	// Docker commands
	if strings.Contains(nameLower, "docker") || strings.Contains(cmdLower, "docker") {
		return "docker-build"
	}

	return ""
}

// generateSkillFile creates a skill file for a specific skill type
func generateSkillFile(skillType string, cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) *types.GeneratedFile {
	// Normalize the command - use Name if Command is empty
	normalizedCmd := cmd
	if normalizedCmd.Command == "" {
		normalizedCmd.Command = normalizedCmd.Name
	}

	var content string

	switch skillType {
	case "build":
		content = buildSkillContent(normalizedCmd, analysis, ctx)
	case "test":
		content = testSkillContent(normalizedCmd, analysis, ctx)
	case "lint":
		content = lintSkillContent(normalizedCmd, analysis, ctx)
	case "dev":
		content = devSkillContent(normalizedCmd, analysis, ctx)
	case "format":
		content = formatSkillContent(normalizedCmd, analysis, ctx)
	case "db-migrate":
		content = dbMigrateSkillContent(normalizedCmd, analysis, ctx)
	case "docker-build":
		content = dockerBuildSkillContent(normalizedCmd, analysis, ctx)
	default:
		return nil
	}

	// Skills use directory structure: .claude/skills/skill-name/SKILL.md
	return &types.GeneratedFile{
		Path:    fmt.Sprintf(".claude/skills/%s/SKILL.md", skillType),
		Content: []byte(content),
	}
}

// generateFrameworkSkills adds framework-specific skills
func (g *ClaudeCodeGenerator) generateFrameworkSkills(analysis *types.Analysis, ctx *GeneratorContext, generated map[string]bool) []types.GeneratedFile {
	var files []types.GeneratedFile

	// Prisma skills
	if hasFramework(analysis, "Prisma") && !generated["db-migrate"] {
		files = append(files, types.GeneratedFile{
			Path:    ".claude/skills/db-migrate/SKILL.md",
			Content: []byte(prismaSkillContent()),
		})
	}

	// Django skills
	if hasFramework(analysis, "Django") && !generated["db-migrate"] {
		files = append(files, types.GeneratedFile{
			Path:    ".claude/skills/db-migrate/SKILL.md",
			Content: []byte(djangoMigrateSkillContent()),
		})
	}

	return files
}

// Skill content generators with YAML frontmatter

func buildSkillContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString("name: build\n")
	content.WriteString("description: Build the project. Use when compiling code, creating artifacts, or preparing for deployment.\n")
	content.WriteString("allowed-tools: Bash, Read, Glob\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Build - %s\n\n", ctx.ProjectName))
	content.WriteString("Build the project.\n\n")

	content.WriteString("## Command\n\n")
	content.WriteString("```bash\n")
	content.WriteString(cmd.Command)
	content.WriteString("\n```\n\n")

	content.WriteString("## Description\n\n")
	content.WriteString(getDescription(cmd, "Compile and build the project"))
	content.WriteString("\n\n")

	// Add entry point reference
	if ctx.EntryPoint != "" {
		content.WriteString("## Entry Point\n\n")
		content.WriteString(fmt.Sprintf("Main: `%s`\n\n", ctx.EntryPoint))
	}

	content.WriteString("## On Failure\n\n")
	content.WriteString("- Check for syntax errors in the output\n")
	content.WriteString("- Verify all dependencies are installed\n")
	content.WriteString("- Check for type errors (if applicable)\n")
	content.WriteString("- Review recent changes that might have broken the build\n\n")

	content.WriteString("## Success Criteria\n\n")
	content.WriteString("- Build completes without errors\n")
	content.WriteString("- Output artifacts are generated correctly\n")

	return content.String()
}

func testSkillContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString("name: test\n")
	content.WriteString("description: Run the test suite. Use when running tests, checking test coverage, or validating code changes.\n")
	content.WriteString("allowed-tools: Bash, Read, Glob, Grep\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Test - %s\n\n", ctx.ProjectName))
	content.WriteString("Run the project test suite.\n\n")

	content.WriteString("## Command\n\n")
	content.WriteString("```bash\n")
	content.WriteString(cmd.Command)
	content.WriteString("\n```\n\n")

	content.WriteString("## Description\n\n")
	content.WriteString(getDescription(cmd, "Run all tests in the project"))
	content.WriteString("\n\n")

	// Add test configuration reference
	if testConfig := ctx.GetTestConfig(); testConfig != "" {
		content.WriteString("## Configuration\n\n")
		content.WriteString(fmt.Sprintf("Test config: `%s`\n\n", testConfig))
	}

	// Add test file examples
	if len(ctx.TestFiles) > 0 {
		content.WriteString("## Test Examples\n\n")
		content.WriteString("Reference these files for test patterns:\n")
		for i, file := range ctx.TestFiles {
			if i >= 3 {
				break
			}
			content.WriteString(fmt.Sprintf("- `%s`\n", file))
		}
		content.WriteString("\n")
	}

	content.WriteString("## On Failure\n\n")
	content.WriteString("- Analyze the failing test output\n")
	content.WriteString("- Identify the root cause of the failure\n")
	content.WriteString("- Check if it's a test issue or a code issue\n")
	content.WriteString("- Suggest fixes for the failing tests\n\n")

	content.WriteString("## Options\n\n")
	content.WriteString(fmt.Sprintf("- Add %s to run specific test files\n", getTestFileFlag(analysis)))
	content.WriteString(fmt.Sprintf("- Add %s for verbose output\n\n", getVerboseFlag(analysis)))

	content.WriteString("## Success Criteria\n\n")
	content.WriteString("- All tests pass\n")
	content.WriteString("- No skipped tests without reason\n")
	content.WriteString("- Coverage meets project standards (if applicable)\n")

	return content.String()
}

func lintSkillContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString("name: lint\n")
	content.WriteString("description: Run code linting and static analysis. Use to check code quality, find issues, or auto-fix style problems.\n")
	content.WriteString("allowed-tools: Bash, Read, Edit, Glob\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Lint - %s\n\n", ctx.ProjectName))
	content.WriteString("Run code linting and static analysis.\n\n")

	content.WriteString("## Command\n\n")
	content.WriteString("```bash\n")
	content.WriteString(cmd.Command)
	content.WriteString("\n```\n\n")

	content.WriteString("## Description\n\n")
	content.WriteString(getDescription(cmd, "Check code for style and potential issues"))
	content.WriteString("\n\n")

	// Add lint configuration reference
	if lintConfig := ctx.GetLintConfig(); lintConfig != "" {
		content.WriteString("## Configuration\n\n")
		content.WriteString(fmt.Sprintf("Lint rules: `%s`\n\n", lintConfig))
	}

	content.WriteString("## On Failure\n\n")
	content.WriteString("- Review each linting error\n")
	content.WriteString(fmt.Sprintf("- Auto-fix when possible using %s flag\n", getAutoFixFlag(analysis)))
	content.WriteString("- For remaining issues, fix manually\n")
	content.WriteString("- Don't ignore warnings without good reason\n\n")

	content.WriteString("## Success Criteria\n\n")
	content.WriteString("- No linting errors\n")
	content.WriteString("- Warnings are reviewed and addressed\n")

	return content.String()
}

func devSkillContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString("name: dev\n")
	content.WriteString("description: Start the development server. Use when running the app locally for testing or development.\n")
	content.WriteString("allowed-tools: Bash, Read\n")
	content.WriteString("disable-model-invocation: true\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Dev - %s\n\n", ctx.ProjectName))
	content.WriteString("Start the development server.\n\n")

	content.WriteString("## Command\n\n")
	content.WriteString("```bash\n")
	content.WriteString(cmd.Command)
	content.WriteString("\n```\n\n")

	content.WriteString("## Description\n\n")
	content.WriteString(getDescription(cmd, "Start the development server with hot reload"))
	content.WriteString("\n\n")

	// Add entry point reference
	if ctx.EntryPoint != "" {
		content.WriteString("## Entry Point\n\n")
		content.WriteString(fmt.Sprintf("Main: `%s`\n\n", ctx.EntryPoint))
	}

	content.WriteString("## Usage\n\n")
	content.WriteString("- Server will start with hot reload enabled\n")
	content.WriteString("- Watch the console for errors\n")
	content.WriteString("- Access the application at the URL shown in output\n\n")

	content.WriteString("## Common Issues\n\n")
	content.WriteString("- Port already in use: kill existing process or use different port\n")
	content.WriteString("- Missing environment variables: check .env file\n")
	content.WriteString("- Dependency issues: run install command first\n")

	return content.String()
}

func formatSkillContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString("name: format\n")
	content.WriteString("description: Format code according to project standards. Use after making changes or before committing.\n")
	content.WriteString("allowed-tools: Bash, Read, Edit, Glob\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Format - %s\n\n", ctx.ProjectName))
	content.WriteString("Format code according to project standards.\n\n")

	content.WriteString("## Command\n\n")
	content.WriteString("```bash\n")
	content.WriteString(cmd.Command)
	content.WriteString("\n```\n\n")

	content.WriteString("## Description\n\n")
	content.WriteString(getDescription(cmd, "Auto-format code files"))
	content.WriteString("\n\n")

	// Add formatter configuration reference
	if prettier, ok := ctx.ConfigFiles["prettier"]; ok {
		content.WriteString("## Configuration\n\n")
		content.WriteString(fmt.Sprintf("Prettier config: `%s`\n\n", prettier))
	}

	content.WriteString("## Usage\n\n")
	content.WriteString("- Run before committing changes\n")
	content.WriteString(fmt.Sprintf("- Use %s to check without modifying files\n", getCheckOnlyFlag(analysis)))
	content.WriteString("- Format specific files by passing paths as arguments\n\n")

	content.WriteString("## Notes\n\n")
	content.WriteString("- This command modifies files in place\n")
	content.WriteString("- Formatting is enforced by pre-commit hooks (if configured)\n")

	return content.String()
}

func dbMigrateSkillContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString("name: db-migrate\n")
	content.WriteString("description: Run database migrations. Use when applying schema changes or setting up the database.\n")
	content.WriteString("allowed-tools: Bash, Read\n")
	content.WriteString("disable-model-invocation: true\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Database Migration - %s\n\n", ctx.ProjectName))
	content.WriteString("Run database migrations.\n\n")

	content.WriteString("## Command\n\n")
	content.WriteString("```bash\n")
	content.WriteString(cmd.Command)
	content.WriteString("\n```\n\n")

	content.WriteString("## Description\n\n")
	content.WriteString(getDescription(cmd, "Apply pending database migrations"))
	content.WriteString("\n\n")

	content.WriteString("## Before Running\n\n")
	content.WriteString("- Ensure database is running and accessible\n")
	content.WriteString("- Check connection settings in environment\n")
	content.WriteString("- Review pending migrations\n\n")

	content.WriteString("## On Failure\n\n")
	content.WriteString("- Check database connection\n")
	content.WriteString("- Review migration files for errors\n")
	content.WriteString("- Check for conflicting migrations\n")
	content.WriteString("- Consider rolling back if needed\n\n")

	content.WriteString("## Related Commands\n\n")
	content.WriteString("- Generate new migration: check project documentation\n")
	content.WriteString("- Rollback: check project documentation\n")
	content.WriteString("- Reset: check project documentation (destructive!)\n")

	return content.String()
}

func dockerBuildSkillContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString("name: docker-build\n")
	content.WriteString("description: Build Docker image for the project. Use for containerizing the application.\n")
	content.WriteString("allowed-tools: Bash, Read, Glob\n")
	content.WriteString("disable-model-invocation: true\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Docker Build - %s\n\n", ctx.ProjectName))
	content.WriteString("Build Docker image for the project.\n\n")

	content.WriteString("## Command\n\n")
	content.WriteString("```bash\n")
	content.WriteString(cmd.Command)
	content.WriteString("\n```\n\n")

	content.WriteString("## Description\n\n")
	content.WriteString(getDescription(cmd, "Build Docker container image"))
	content.WriteString("\n\n")

	content.WriteString("## Options\n\n")
	content.WriteString("- Add `-t image:tag` to tag the image\n")
	content.WriteString("- Add `--no-cache` for no cache build\n")
	content.WriteString("- Add `-f Dockerfile.custom` for specific Dockerfile\n\n")

	content.WriteString("## Success Criteria\n\n")
	content.WriteString("- Image builds successfully\n")
	content.WriteString("- Image size is reasonable\n")
	content.WriteString("- All required files are included\n")

	return content.String()
}

func prismaSkillContent() string {
	return `---
name: db-migrate
description: Manage database schema with Prisma. Use for migrations, schema changes, and database operations.
allowed-tools: Bash, Read
disable-model-invocation: true
---

# Database Migration (Prisma)

Manage database schema with Prisma.

## Commands

### Apply migrations
` + "```bash" + `
npx prisma migrate dev
` + "```" + `

### Generate client
` + "```bash" + `
npx prisma generate
` + "```" + `

### Push schema (dev only)
` + "```bash" + `
npx prisma db push
` + "```" + `

### View database
` + "```bash" + `
npx prisma studio
` + "```" + `

## On Failure

- Check DATABASE_URL environment variable
- Ensure database server is running
- Review migration files for conflicts
- Check Prisma schema for errors

## Notes

- Always run 'prisma generate' after schema changes
- Use 'migrate dev' for development, 'migrate deploy' for production
- 'db push' is for prototyping only, doesn't create migration files
`
}

func djangoMigrateSkillContent() string {
	return `---
name: db-migrate
description: Manage database schema with Django migrations. Use for database setup and schema changes.
allowed-tools: Bash, Read
disable-model-invocation: true
---

# Database Migration (Django)

Manage database schema with Django migrations.

## Commands

### Apply migrations
` + "```bash" + `
python manage.py migrate
` + "```" + `

### Create new migration
` + "```bash" + `
python manage.py makemigrations
` + "```" + `

### Show migration status
` + "```bash" + `
python manage.py showmigrations
` + "```" + `

### Rollback migration
` + "```bash" + `
python manage.py migrate app_name migration_name
` + "```" + `

## On Failure

- Check database connection in settings.py
- Ensure database server is running
- Review migration files for conflicts
- Check for circular dependencies

## Notes

- Always run 'makemigrations' after model changes
- Review generated migrations before applying
- Test migrations on staging before production
`
}

// Helper functions

func getDescription(cmd types.Command, fallback string) string {
	if cmd.Description != "" {
		return cmd.Description
	}
	return fallback
}

func getTestFileFlag(analysis *types.Analysis) string {
	if hasLanguage(analysis, "Go") {
		return "`-run TestName`"
	}
	if hasLanguage(analysis, "Python") {
		return "`-k test_name`"
	}
	return "`-- path/to/test`"
}

func getVerboseFlag(analysis *types.Analysis) string {
	if hasLanguage(analysis, "Go") {
		return "`-v`"
	}
	if hasLanguage(analysis, "Python") {
		return "`-v`"
	}
	return "`--verbose`"
}

func getAutoFixFlag(analysis *types.Analysis) string {
	if hasLanguage(analysis, "Go") {
		return "`--fix`"
	}
	return "`--fix`"
}

func getCheckOnlyFlag(analysis *types.Analysis) string {
	if hasLanguage(analysis, "Go") {
		return "`-d` (diff only)"
	}
	if hasLanguage(analysis, "Python") {
		return "`--check`"
	}
	return "`--check`"
}

// generateProjectToolSkills creates skills for project-specific tools
func (g *ClaudeCodeGenerator) generateProjectToolSkills(analysis *types.Analysis) []types.GeneratedFile {
	var files []types.GeneratedFile

	for _, tool := range analysis.ProjectTools {
		skillFile := generateProjectToolSkillFile(tool, analysis)
		if skillFile != nil {
			files = append(files, *skillFile)
		}
	}

	return files
}

// generateProjectToolSkillFile creates a skill file for a project-specific tool
func generateProjectToolSkillFile(tool types.ProjectTool, analysis *types.Analysis) *types.GeneratedFile {
	var content strings.Builder

	// YAML frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("name: %s\n", tool.Name))
	content.WriteString(fmt.Sprintf("description: %s\n", tool.Description))

	// Add replacement warning if applicable
	if tool.ReplacesTool != "" {
		content.WriteString(fmt.Sprintf("# CRITICAL: This skill replaces the %s tool for this project\n", tool.ReplacesTool))
	}

	content.WriteString("---\n\n")

	// Title - capitalize first letter
	toolTitle := tool.Name
	if len(toolTitle) > 0 {
		toolTitle = strings.ToUpper(toolTitle[:1]) + toolTitle[1:]
	}
	content.WriteString(fmt.Sprintf("# %s - Project Tool\n\n", toolTitle))

	// Description
	content.WriteString(fmt.Sprintf("%s\n\n", tool.Description))

	// When to use
	if tool.WhenToUse != "" {
		content.WriteString("## When to Invoke This Skill\n\n")
		content.WriteString(fmt.Sprintf("%s\n\n", tool.WhenToUse))
	}

	// Setup instructions if needed
	if tool.RequiresSetup && tool.SetupInstructions != "" {
		content.WriteString("## Setup Required\n\n")
		content.WriteString(fmt.Sprintf("%s\n\n", tool.SetupInstructions))
	}

	// Usage examples
	if len(tool.UsageExamples) > 0 {
		content.WriteString("## Usage Examples\n\n")
		for _, example := range tool.UsageExamples {
			content.WriteString("```bash\n")
			content.WriteString(example + "\n")
			content.WriteString("```\n\n")
		}
	}

	// Binary path if available
	if tool.BinaryPath != "" {
		content.WriteString("## Binary Location\n\n")
		content.WriteString(fmt.Sprintf("The tool binary is located at: `%s`\n\n", tool.BinaryPath))

		// Check if needs building
		needsBuild := strings.Contains(tool.BinaryPath, "bin/") || strings.Contains(tool.BinaryPath, "build/")
		if needsBuild {
			content.WriteString("You may need to build the tool first:\n\n")
			content.WriteString("```bash\n")

			// Suggest build command based on project type
			if hasLanguage(analysis, "Go") {
				content.WriteString("make build  # or: go build -o " + tool.BinaryPath + "\n")
			} else if hasLanguage(analysis, "Rust") {
				content.WriteString("cargo build --release\n")
			} else {
				content.WriteString("make build\n")
			}
			content.WriteString("```\n\n")
		}
	}

	// Tool replacement guidance
	if tool.ReplacesTool != "" {
		content.WriteString(fmt.Sprintf("## IMPORTANT: Replaces %s\n\n", tool.ReplacesTool))
		content.WriteString(fmt.Sprintf("**This tool replaces the built-in %s for this project.**\n\n", tool.ReplacesTool))
		content.WriteString("When working on this codebase:\n")
		content.WriteString(fmt.Sprintf("- ✅ Use `%s` for project-specific searches\n", tool.Name))
		content.WriteString(fmt.Sprintf("- ❌ Avoid using built-in %s unless for exact text matching\n\n", tool.ReplacesTool))
	}

	// Best practices
	content.WriteString("## Best Practices\n\n")
	content.WriteString(fmt.Sprintf("- Use `%s` when working with this codebase\n", tool.Name))
	content.WriteString("- Check that the tool is available before use\n")
	if tool.ReplacesTool != "" {
		content.WriteString(fmt.Sprintf("- Fall back to %s only if this tool fails\n", tool.ReplacesTool))
	}
	content.WriteString("\n")

	return &types.GeneratedFile{
		Path:    fmt.Sprintf(".claude/skills/%s/SKILL.md", tool.Name),
		Content: []byte(content.String()),
	}
}
