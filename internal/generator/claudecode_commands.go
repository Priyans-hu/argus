package generator

import (
	"fmt"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// generateCommands creates command files based on detected commands
func (g *ClaudeCodeGenerator) generateCommands(analysis *types.Analysis) []types.GeneratedFile {
	var files []types.GeneratedFile

	// Build context for context-aware generation
	ctx := BuildContext(analysis)

	// Track which commands we've generated to avoid duplicates
	generated := make(map[string]bool)

	// Generate commands from detected commands
	for _, cmd := range analysis.Commands {
		cmdType := classifyCommand(cmd.Name, cmd.Command)
		if cmdType != "" && !generated[cmdType] {
			generated[cmdType] = true
			file := generateCommandFile(cmdType, cmd, analysis, ctx)
			if file != nil {
				files = append(files, *file)
			}
		}
	}

	// Add framework-specific commands
	frameworkCmds := g.generateFrameworkCommands(analysis, ctx, generated)
	files = append(files, frameworkCmds...)

	return files
}

// classifyCommand determines the command type from its name or content
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

// generateCommandFile creates a command file for a specific command type
func generateCommandFile(cmdType string, cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) *types.GeneratedFile {
	// Normalize the command - use Name if Command is empty
	normalizedCmd := cmd
	if normalizedCmd.Command == "" {
		normalizedCmd.Command = normalizedCmd.Name
	}

	var content string

	switch cmdType {
	case "build":
		content = buildCommandContent(normalizedCmd, analysis, ctx)
	case "test":
		content = testCommandContent(normalizedCmd, analysis, ctx)
	case "lint":
		content = lintCommandContent(normalizedCmd, analysis, ctx)
	case "dev":
		content = devCommandContent(normalizedCmd, analysis, ctx)
	case "format":
		content = formatCommandContent(normalizedCmd, analysis, ctx)
	case "db-migrate":
		content = dbMigrateCommandContent(normalizedCmd, analysis, ctx)
	case "docker-build":
		content = dockerBuildCommandContent(normalizedCmd, analysis, ctx)
	default:
		return nil
	}

	return &types.GeneratedFile{
		Path:    fmt.Sprintf(".claude/commands/%s.md", cmdType),
		Content: []byte(content),
	}
}

// generateFrameworkCommands adds framework-specific commands
func (g *ClaudeCodeGenerator) generateFrameworkCommands(analysis *types.Analysis, ctx *GeneratorContext, generated map[string]bool) []types.GeneratedFile {
	var files []types.GeneratedFile

	// Prisma commands
	if hasFramework(analysis, "Prisma") && !generated["db-migrate"] {
		files = append(files, types.GeneratedFile{
			Path:    ".claude/commands/db-migrate.md",
			Content: []byte(prismaCommandContent()),
		})
	}

	// Django commands
	if hasFramework(analysis, "Django") && !generated["db-migrate"] {
		files = append(files, types.GeneratedFile{
			Path:    ".claude/commands/db-migrate.md",
			Content: []byte(djangoMigrateCommandContent()),
		})
	}

	return files
}

func buildCommandContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

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

func testCommandContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

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

func lintCommandContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

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

func devCommandContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

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

func formatCommandContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

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

func dbMigrateCommandContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

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

func dockerBuildCommandContent(cmd types.Command, analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

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

func prismaCommandContent() string {
	return `# Database Migration (Prisma)

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

func djangoMigrateCommandContent() string {
	return `# Database Migration (Django)

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
