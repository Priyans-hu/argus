package generator

import (
	"fmt"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// generateCommands creates command files based on detected commands
func (g *ClaudeCodeGenerator) generateCommands(analysis *types.Analysis) []types.GeneratedFile {
	var files []types.GeneratedFile

	// Track which commands we've generated to avoid duplicates
	generated := make(map[string]bool)

	// Generate commands from detected commands
	for _, cmd := range analysis.Commands {
		cmdType := classifyCommand(cmd.Name, cmd.Command)
		if cmdType != "" && !generated[cmdType] {
			generated[cmdType] = true
			file := generateCommandFile(cmdType, cmd, analysis)
			if file != nil {
				files = append(files, *file)
			}
		}
	}

	// Add framework-specific commands
	frameworkCmds := g.generateFrameworkCommands(analysis, generated)
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
func generateCommandFile(cmdType string, cmd types.Command, analysis *types.Analysis) *types.GeneratedFile {
	// Normalize the command - use Name if Command is empty
	normalizedCmd := cmd
	if normalizedCmd.Command == "" {
		normalizedCmd.Command = normalizedCmd.Name
	}

	var content string

	switch cmdType {
	case "build":
		content = buildCommandContent(normalizedCmd, analysis)
	case "test":
		content = testCommandContent(normalizedCmd, analysis)
	case "lint":
		content = lintCommandContent(normalizedCmd, analysis)
	case "dev":
		content = devCommandContent(normalizedCmd, analysis)
	case "format":
		content = formatCommandContent(normalizedCmd, analysis)
	case "db-migrate":
		content = dbMigrateCommandContent(normalizedCmd, analysis)
	case "docker-build":
		content = dockerBuildCommandContent(normalizedCmd, analysis)
	default:
		return nil
	}

	return &types.GeneratedFile{
		Path:    fmt.Sprintf(".claude/commands/%s.md", cmdType),
		Content: []byte(content),
	}
}

// generateFrameworkCommands adds framework-specific commands
func (g *ClaudeCodeGenerator) generateFrameworkCommands(analysis *types.Analysis, generated map[string]bool) []types.GeneratedFile {
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

func buildCommandContent(cmd types.Command, analysis *types.Analysis) string {
	return fmt.Sprintf(`# Build

Build the project.

## Command

%s
%s
%s

## Description

%s

## On Failure

- Check for syntax errors in the output
- Verify all dependencies are installed
- Check for type errors (if applicable)
- Review recent changes that might have broken the build

## Success Criteria

- Build completes without errors
- Output artifacts are generated correctly
`, "```bash", cmd.Command, "```", getDescription(cmd, "Compile and build the project"))
}

func testCommandContent(cmd types.Command, analysis *types.Analysis) string {
	return fmt.Sprintf(`# Test

Run the project test suite.

## Command

%s
%s
%s

## Description

%s

## On Failure

- Analyze the failing test output
- Identify the root cause of the failure
- Check if it's a test issue or a code issue
- Suggest fixes for the failing tests

## Options

- Add %s to run specific test files
- Add %s for verbose output

## Success Criteria

- All tests pass
- No skipped tests without reason
- Coverage meets project standards (if applicable)
`, "```bash", cmd.Command, "```", getDescription(cmd, "Run all tests in the project"),
		getTestFileFlag(analysis), getVerboseFlag(analysis))
}

func lintCommandContent(cmd types.Command, analysis *types.Analysis) string {
	return fmt.Sprintf(`# Lint

Run code linting and static analysis.

## Command

%s
%s
%s

## Description

%s

## On Failure

- Review each linting error
- Auto-fix when possible using %s flag
- For remaining issues, fix manually
- Don't ignore warnings without good reason

## Success Criteria

- No linting errors
- Warnings are reviewed and addressed
`, "```bash", cmd.Command, "```", getDescription(cmd, "Check code for style and potential issues"),
		getAutoFixFlag(analysis))
}

func devCommandContent(cmd types.Command, analysis *types.Analysis) string {
	return fmt.Sprintf(`# Dev

Start the development server.

## Command

%s
%s
%s

## Description

%s

## Usage

- Server will start with hot reload enabled
- Watch the console for errors
- Access the application at the URL shown in output

## Common Issues

- Port already in use: kill existing process or use different port
- Missing environment variables: check .env file
- Dependency issues: run install command first
`, "```bash", cmd.Command, "```", getDescription(cmd, "Start the development server with hot reload"))
}

func formatCommandContent(cmd types.Command, analysis *types.Analysis) string {
	return fmt.Sprintf(`# Format

Format code according to project standards.

## Command

%s
%s
%s

## Description

%s

## Usage

- Run before committing changes
- Use %s to check without modifying files
- Format specific files by passing paths as arguments

## Notes

- This command modifies files in place
- Formatting is enforced by pre-commit hooks (if configured)
`, "```bash", cmd.Command, "```", getDescription(cmd, "Auto-format code files"),
		getCheckOnlyFlag(analysis))
}

func dbMigrateCommandContent(cmd types.Command, analysis *types.Analysis) string {
	return fmt.Sprintf(`# Database Migration

Run database migrations.

## Command

%s
%s
%s

## Description

%s

## Before Running

- Ensure database is running and accessible
- Check connection settings in environment
- Review pending migrations

## On Failure

- Check database connection
- Review migration files for errors
- Check for conflicting migrations
- Consider rolling back if needed

## Related Commands

- Generate new migration: check project documentation
- Rollback: check project documentation
- Reset: check project documentation (destructive!)
`, "```bash", cmd.Command, "```", getDescription(cmd, "Apply pending database migrations"))
}

func dockerBuildCommandContent(cmd types.Command, analysis *types.Analysis) string {
	return fmt.Sprintf(`# Docker Build

Build Docker image for the project.

## Command

%s
%s
%s

## Description

%s

## Options

- Add %s to tag the image
- Add %s for no cache build
- Add %s for specific Dockerfile

## Success Criteria

- Image builds successfully
- Image size is reasonable
- All required files are included
`, "```bash", cmd.Command, "```", getDescription(cmd, "Build Docker container image"),
		"`-t image:tag`", "`--no-cache`", "`-f Dockerfile.custom`")
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
