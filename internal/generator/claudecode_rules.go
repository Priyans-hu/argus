package generator

import (
	"fmt"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// generateRules creates rule files based on detected conventions and patterns
func (g *ClaudeCodeGenerator) generateRules(analysis *types.Analysis) []types.GeneratedFile {
	var files []types.GeneratedFile

	// Build context for context-aware generation
	ctx := BuildContext(analysis)

	// Git workflow rules (from GitConventions)
	if analysis.GitConventions != nil {
		if file := g.generateGitWorkflowRule(analysis, ctx); file != nil {
			files = append(files, *file)
		}
	}

	// Testing rules (from CodePatterns)
	if analysis.CodePatterns != nil && len(analysis.CodePatterns.Testing) > 0 {
		if file := g.generateTestingRule(analysis, ctx); file != nil {
			files = append(files, *file)
		}
	}

	// Coding style rules (from Conventions)
	if len(analysis.Conventions) > 0 {
		if file := g.generateCodingStyleRule(analysis, ctx); file != nil {
			files = append(files, *file)
		}
	}

	// Architecture rules (if architecture is detected)
	if ctx.HasArchitectureContext() {
		if file := g.generateArchitectureRule(analysis, ctx); file != nil {
			files = append(files, *file)
		}
	}

	// Security rules (always generated)
	files = append(files, types.GeneratedFile{
		Path:    ".claude/rules/security.md",
		Content: []byte(g.securityRuleContent(analysis, ctx)),
	})

	return files
}

// generateGitWorkflowRule creates git workflow rules from detected conventions
func (g *ClaudeCodeGenerator) generateGitWorkflowRule(analysis *types.Analysis, ctx *GeneratorContext) *types.GeneratedFile {
	if analysis.GitConventions == nil {
		return nil
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("# Git Workflow Rules for %s\n\n", ctx.ProjectName))
	content.WriteString("Follow these git conventions for this project.\n\n")

	// Commit conventions
	if cc := analysis.GitConventions.CommitConvention; cc != nil {
		content.WriteString("## Commit Messages\n\n")
		content.WriteString(fmt.Sprintf("- **Style**: %s\n", cc.Style))
		if cc.Format != "" {
			content.WriteString(fmt.Sprintf("- **Format**: `%s`\n", cc.Format))
		}
		if len(cc.Types) > 0 {
			content.WriteString(fmt.Sprintf("- **Types**: %s\n", strings.Join(cc.Types, ", ")))
		}
		if cc.Example != "" {
			content.WriteString(fmt.Sprintf("- **Example**: `%s`\n", cc.Example))
		}
		content.WriteString("\n")

		// Add type descriptions for conventional commits
		if cc.Style == "conventional" {
			content.WriteString("### Commit Types\n\n")
			content.WriteString("| Type | Description |\n")
			content.WriteString("|------|-------------|\n")
			content.WriteString("| feat | New feature |\n")
			content.WriteString("| fix | Bug fix |\n")
			content.WriteString("| docs | Documentation only |\n")
			content.WriteString("| style | Formatting, no code change |\n")
			content.WriteString("| refactor | Code restructuring |\n")
			content.WriteString("| test | Adding/fixing tests |\n")
			content.WriteString("| chore | Maintenance tasks |\n")
			content.WriteString("\n")
		}
	}

	// Branch conventions
	if bc := analysis.GitConventions.BranchConvention; bc != nil {
		content.WriteString("## Branch Naming\n\n")
		if bc.Format != "" {
			content.WriteString(fmt.Sprintf("- **Format**: `%s`\n", bc.Format))
		}
		if len(bc.Prefixes) > 0 {
			content.WriteString(fmt.Sprintf("- **Prefixes**: %s\n", strings.Join(bc.Prefixes, ", ")))
		}
		if len(bc.Examples) > 0 {
			content.WriteString("- **Examples**:\n")
			for _, ex := range bc.Examples {
				content.WriteString(fmt.Sprintf("  - `%s`\n", ex))
			}
		}
		content.WriteString("\n")
	}

	// Workflow
	content.WriteString("## Workflow\n\n")
	content.WriteString("1. Create a feature branch from main\n")
	content.WriteString("2. Make small, focused commits\n")
	content.WriteString("3. Push and open a pull request\n")
	content.WriteString("4. Address review comments\n")
	content.WriteString("5. Squash and merge when approved\n")

	return &types.GeneratedFile{
		Path:    ".claude/rules/git-workflow.md",
		Content: []byte(content.String()),
	}
}

// generateTestingRule creates testing rules from detected patterns
func (g *ClaudeCodeGenerator) generateTestingRule(analysis *types.Analysis, ctx *GeneratorContext) *types.GeneratedFile {
	if analysis.CodePatterns == nil || len(analysis.CodePatterns.Testing) == 0 {
		return nil
	}

	var content strings.Builder

	// Add path-specific frontmatter for test files
	testPatterns := getTestFilePatterns(analysis)
	if len(testPatterns) > 0 {
		content.WriteString("---\n")
		content.WriteString("paths:\n")
		for _, pattern := range testPatterns {
			content.WriteString(fmt.Sprintf("  - \"%s\"\n", pattern))
		}
		content.WriteString("---\n\n")
	}

	content.WriteString(fmt.Sprintf("# Testing Rules for %s\n\n", ctx.ProjectName))
	content.WriteString("Follow these testing conventions for this project.\n\n")

	// Detected testing patterns with file examples
	content.WriteString("## Detected Testing Patterns\n\n")
	for _, pattern := range analysis.CodePatterns.Testing {
		content.WriteString(fmt.Sprintf("- **%s**: %s", pattern.Name, pattern.Description))
		if pattern.FileCount > 0 {
			content.WriteString(fmt.Sprintf(" (%d files)", pattern.FileCount))
		}
		content.WriteString("\n")
		// Add example file if available
		if len(pattern.Examples) > 0 {
			content.WriteString(fmt.Sprintf("  - See: `%s`\n", pattern.Examples[0]))
		}
	}
	content.WriteString("\n")

	// Test file examples section
	if len(ctx.TestFiles) > 0 {
		content.WriteString("## Test File Examples\n\n")
		content.WriteString("Reference these files for test patterns:\n")
		for _, file := range ctx.TestFiles {
			content.WriteString(fmt.Sprintf("- `%s`\n", file))
		}
		content.WriteString("\n")
	}

	// Test command section
	if ctx.TestCommand != "" {
		content.WriteString("## Running Tests\n\n")
		content.WriteString("```bash\n")
		content.WriteString(ctx.TestCommand)
		content.WriteString("\n```\n\n")
	}

	// Test configuration
	if testConfig := ctx.GetTestConfig(); testConfig != "" {
		content.WriteString("## Test Configuration\n\n")
		content.WriteString(fmt.Sprintf("Config file: `%s`\n\n", testConfig))
	}

	// Language-specific testing guidelines
	content.WriteString("## Testing Guidelines\n\n")

	if hasLanguage(analysis, "Go") {
		content.WriteString("### Go Testing\n\n")
		content.WriteString("- Use table-driven tests for multiple cases\n")
		content.WriteString("- Use `t.Run()` for subtests\n")
		content.WriteString("- Name test functions as `TestFunctionName_Scenario`\n")
		content.WriteString("- Use `t.Parallel()` for independent tests\n")
		content.WriteString("- Mock external dependencies\n")
		content.WriteString("\n")
	}

	if hasLanguage(analysis, "TypeScript") || hasLanguage(analysis, "JavaScript") {
		content.WriteString("### JavaScript/TypeScript Testing\n\n")
		content.WriteString("- Use describe/it blocks for organization\n")
		content.WriteString("- Follow Arrange-Act-Assert pattern\n")
		content.WriteString("- Mock external services and APIs\n")
		content.WriteString("- Test both success and error cases\n")
		content.WriteString("- Use meaningful test descriptions\n")
		content.WriteString("\n")
	}

	if hasLanguage(analysis, "Python") {
		content.WriteString("### Python Testing\n\n")
		content.WriteString("- Use pytest as the test framework\n")
		content.WriteString("- Use fixtures for test setup\n")
		content.WriteString("- Use parametrize for multiple test cases\n")
		content.WriteString("- Name test files as `test_*.py`\n")
		content.WriteString("- Name test functions as `test_*`\n")
		content.WriteString("\n")
	}

	// General best practices
	content.WriteString("## Best Practices\n\n")
	content.WriteString("- Write tests before or alongside code\n")
	content.WriteString("- Keep tests focused and independent\n")
	content.WriteString("- Test edge cases and error conditions\n")
	content.WriteString("- Don't test implementation details\n")
	content.WriteString("- Maintain test coverage for critical paths\n")

	return &types.GeneratedFile{
		Path:    ".claude/rules/testing.md",
		Content: []byte(content.String()),
	}
}

// generateCodingStyleRule creates coding style rules from detected conventions
func (g *ClaudeCodeGenerator) generateCodingStyleRule(analysis *types.Analysis, ctx *GeneratorContext) *types.GeneratedFile {
	var content strings.Builder
	content.WriteString(fmt.Sprintf("# Coding Style Rules for %s\n\n", ctx.ProjectName))
	content.WriteString("Follow these coding conventions for this project.\n\n")

	// Configuration files section
	if len(ctx.ConfigFiles) > 0 {
		content.WriteString("## Configuration Files\n\n")
		if eslint, ok := ctx.ConfigFiles["eslint"]; ok {
			content.WriteString(fmt.Sprintf("- **ESLint**: `%s`\n", eslint))
		}
		if prettier, ok := ctx.ConfigFiles["prettier"]; ok {
			content.WriteString(fmt.Sprintf("- **Prettier**: `%s`\n", prettier))
		}
		if golangci, ok := ctx.ConfigFiles["golangci"]; ok {
			content.WriteString(fmt.Sprintf("- **golangci-lint**: `%s`\n", golangci))
		}
		if tsconfig, ok := ctx.ConfigFiles["tsconfig"]; ok {
			content.WriteString(fmt.Sprintf("- **TypeScript**: `%s`\n", tsconfig))
		}
		if pyproject, ok := ctx.ConfigFiles["pyproject"]; ok {
			content.WriteString(fmt.Sprintf("- **Python**: `%s`\n", pyproject))
		}
		content.WriteString("\n")
	}

	// Lint/format commands
	if ctx.LintCommand != "" || ctx.FormatCommand != "" {
		content.WriteString("## Commands\n\n")
		if ctx.LintCommand != "" {
			content.WriteString(fmt.Sprintf("- Lint: `%s`\n", ctx.LintCommand))
		}
		if ctx.FormatCommand != "" {
			content.WriteString(fmt.Sprintf("- Format: `%s`\n", ctx.FormatCommand))
		}
		content.WriteString("\n")
	}

	// Group conventions by category
	categories := make(map[string][]types.Convention)
	for _, conv := range analysis.Conventions {
		categories[conv.Category] = append(categories[conv.Category], conv)
	}

	// Write conventions by category
	categoryOrder := []string{"code-style", "naming", "imports", "structure", "error-handling", "custom"}
	for _, cat := range categoryOrder {
		if convs, ok := categories[cat]; ok && len(convs) > 0 {
			content.WriteString(fmt.Sprintf("## %s\n\n", formatCategoryName(cat)))
			for _, conv := range convs {
				content.WriteString(fmt.Sprintf("- %s\n", conv.Description))
				if conv.Example != "" {
					content.WriteString(fmt.Sprintf("  ```\n  %s\n  ```\n", conv.Example))
				}
			}
			content.WriteString("\n")
		}
	}

	// Language-specific style guidelines
	if hasLanguage(analysis, "Go") {
		content.WriteString("## Go Style\n\n")
		content.WriteString("- Run `gofmt` or `goimports` before committing\n")
		content.WriteString("- Follow Effective Go guidelines\n")
		content.WriteString("- Use meaningful variable names\n")
		content.WriteString("- Keep functions focused and small\n")
		content.WriteString("\n")
	}

	if hasLanguage(analysis, "TypeScript") || hasLanguage(analysis, "JavaScript") {
		content.WriteString("## TypeScript/JavaScript Style\n\n")
		content.WriteString("- Use Prettier for formatting\n")
		content.WriteString("- Follow ESLint rules\n")
		content.WriteString("- Prefer const over let\n")
		content.WriteString("- Use meaningful variable and function names\n")
		content.WriteString("\n")
	}

	if hasLanguage(analysis, "Python") {
		content.WriteString("## Python Style\n\n")
		content.WriteString("- Follow PEP 8 style guide\n")
		content.WriteString("- Use type hints for function signatures\n")
		content.WriteString("- Write docstrings for public functions\n")
		content.WriteString("- Use Black or similar formatter\n")
		content.WriteString("\n")
	}

	return &types.GeneratedFile{
		Path:    ".claude/rules/coding-style.md",
		Content: []byte(content.String()),
	}
}

// generateArchitectureRule creates architecture rules from detected patterns
func (g *ClaudeCodeGenerator) generateArchitectureRule(analysis *types.Analysis, ctx *GeneratorContext) *types.GeneratedFile {
	var content strings.Builder

	arch := analysis.ArchitectureInfo
	styleName := "Modular"
	if arch != nil && arch.Style != "" {
		styleName = arch.Style
	}

	content.WriteString(fmt.Sprintf("# Architecture Rules for %s\n\n", ctx.ProjectName))
	content.WriteString(fmt.Sprintf("This project follows a **%s** architecture.\n\n", styleName))

	// Entry point
	if ctx.EntryPoint != "" {
		content.WriteString("## Entry Point\n\n")
		content.WriteString(fmt.Sprintf("Main entry: `%s`\n\n", ctx.EntryPoint))
	}

	// Layer structure
	if len(ctx.Layers) > 0 {
		content.WriteString("## Layer Structure\n\n")
		content.WriteString("Follow the dependency rules between layers:\n\n")

		for _, layer := range ctx.Layers {
			content.WriteString(fmt.Sprintf("### %s\n\n", layer.Name))
			if layer.Purpose != "" {
				content.WriteString(fmt.Sprintf("%s\n\n", layer.Purpose))
			}
			if len(layer.Packages) > 0 {
				content.WriteString("**Packages**: ")
				content.WriteString("`" + strings.Join(layer.Packages, "`, `") + "`\n\n")
			}
			if len(layer.DependsOn) > 0 {
				content.WriteString("**Can depend on**: ")
				content.WriteString("`" + strings.Join(layer.DependsOn, "`, `") + "`\n\n")
			}
		}
	}

	// Key directories from structure
	if len(analysis.Structure.Directories) > 0 {
		content.WriteString("## Key Directories\n\n")
		count := 0
		for _, dir := range analysis.Structure.Directories {
			if count >= 8 {
				break
			}
			if dir.Purpose != "" {
				content.WriteString(fmt.Sprintf("- `%s/` - %s\n", dir.Path, dir.Purpose))
				count++
			}
		}
		if count > 0 {
			content.WriteString("\n")
		}
	}

	// Architecture diagram
	if ctx.Diagram != "" {
		content.WriteString("## Diagram\n\n")
		content.WriteString("```\n")
		content.WriteString(ctx.Diagram)
		content.WriteString("\n```\n\n")
	}

	// Guidelines
	content.WriteString("## Guidelines\n\n")
	content.WriteString("- Respect layer boundaries - lower layers should not depend on higher layers\n")
	content.WriteString("- Keep business logic in the appropriate layer\n")
	content.WriteString("- Use dependency injection for cross-layer dependencies\n")
	content.WriteString("- New features should follow the existing architectural patterns\n")

	return &types.GeneratedFile{
		Path:    ".claude/rules/architecture.md",
		Content: []byte(content.String()),
	}
}

// securityRuleContent generates security rules
func (g *ClaudeCodeGenerator) securityRuleContent(analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// Add path-specific frontmatter for security-sensitive files
	securityPatterns := getSecuritySensitivePatterns(analysis)
	if len(securityPatterns) > 0 {
		content.WriteString("---\n")
		content.WriteString("paths:\n")
		for _, pattern := range securityPatterns {
			content.WriteString(fmt.Sprintf("  - \"%s\"\n", pattern))
		}
		content.WriteString("---\n\n")
	}

	content.WriteString(fmt.Sprintf("# Security Rules for %s\n\n", ctx.ProjectName))
	content.WriteString("Follow these security practices for all code changes.\n\n")

	// Project-specific auth patterns
	if len(ctx.AuthPatterns) > 0 {
		content.WriteString("## Authentication in This Project\n\n")
		content.WriteString("Detected patterns:\n")
		for pattern, files := range ctx.AuthPatterns {
			content.WriteString(fmt.Sprintf("- **%s**", pattern))
			if len(files) > 0 {
				content.WriteString(fmt.Sprintf(" - see `%s`", files[0]))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	content.WriteString("## General Security\n\n")
	content.WriteString("- Never commit secrets, API keys, or credentials\n")
	content.WriteString("- Validate all user input\n")
	content.WriteString("- Use parameterized queries for database operations\n")
	content.WriteString("- Sanitize output to prevent XSS\n")
	content.WriteString("- Use HTTPS for all external communications\n")
	content.WriteString("- Keep dependencies updated\n")
	content.WriteString("\n")

	content.WriteString("## Authentication & Authorization\n\n")
	content.WriteString("- Implement proper authentication for protected resources\n")
	content.WriteString("- Use secure session management\n")
	content.WriteString("- Implement rate limiting for authentication endpoints\n")
	content.WriteString("- Use secure password hashing (bcrypt, argon2)\n")
	content.WriteString("- Validate authorization for all protected actions\n")
	content.WriteString("\n")

	content.WriteString("## Data Protection\n\n")
	content.WriteString("- Encrypt sensitive data at rest\n")
	content.WriteString("- Don't log sensitive information\n")
	content.WriteString("- Use secure random number generation\n")
	content.WriteString("- Implement proper error handling (don't leak details)\n")
	content.WriteString("\n")

	// Language-specific security
	if hasLanguage(analysis, "Go") {
		content.WriteString("## Go Security\n\n")
		content.WriteString("- Use `html/template` for HTML output (auto-escaping)\n")
		content.WriteString("- Validate and sanitize all exec.Command inputs\n")
		content.WriteString("- Use `crypto/rand` not `math/rand` for security\n")
		content.WriteString("- Configure TLS properly (min TLS 1.2)\n")
		content.WriteString("\n")
	}

	if hasLanguage(analysis, "TypeScript") || hasLanguage(analysis, "JavaScript") {
		content.WriteString("## JavaScript Security\n\n")
		content.WriteString("- Use Content Security Policy headers\n")
		content.WriteString("- Implement CSRF protection\n")
		content.WriteString("- Avoid eval() and innerHTML with user data\n")
		content.WriteString("- Run `npm audit` regularly\n")
		content.WriteString("- Use secure cookie settings (httpOnly, secure, sameSite)\n")
		content.WriteString("\n")
	}

	if hasLanguage(analysis, "Python") {
		content.WriteString("## Python Security\n\n")
		content.WriteString("- Use `yaml.safe_load()` not `yaml.load()`\n")
		content.WriteString("- Avoid `pickle` with untrusted data\n")
		content.WriteString("- Use parameterized queries (SQLAlchemy, Django ORM)\n")
		content.WriteString("- Validate file paths to prevent traversal\n")
		content.WriteString("- Run `pip-audit` or `safety` regularly\n")
		content.WriteString("\n")
	}

	content.WriteString("## Sensitive Files\n\n")
	content.WriteString("Never commit these files:\n")
	content.WriteString("- `.env` files with secrets\n")
	content.WriteString("- Private keys (*.pem, *.key)\n")
	content.WriteString("- Credentials files\n")
	content.WriteString("- Database connection strings with passwords\n")

	return content.String()
}

// getTestFilePatterns returns glob patterns for test files based on detected languages
func getTestFilePatterns(analysis *types.Analysis) []string {
	var patterns []string

	if hasLanguage(analysis, "Go") {
		patterns = append(patterns, "**/*_test.go")
	}
	if hasLanguage(analysis, "TypeScript") {
		patterns = append(patterns, "**/*.test.ts", "**/*.spec.ts", "**/*.test.tsx", "**/*.spec.tsx")
	}
	if hasLanguage(analysis, "JavaScript") {
		patterns = append(patterns, "**/*.test.js", "**/*.spec.js", "**/*.test.jsx", "**/*.spec.jsx")
	}
	if hasLanguage(analysis, "Python") {
		patterns = append(patterns, "**/test_*.py", "**/*_test.py")
	}
	if hasLanguage(analysis, "Java") {
		patterns = append(patterns, "**/*Test.java", "**/*Tests.java")
	}
	if hasLanguage(analysis, "Rust") {
		patterns = append(patterns, "**/tests/**/*.rs")
	}
	if hasLanguage(analysis, "Ruby") {
		patterns = append(patterns, "**/*_spec.rb", "**/test_*.rb")
	}

	return patterns
}

// getSecuritySensitivePatterns returns glob patterns for security-sensitive files
func getSecuritySensitivePatterns(analysis *types.Analysis) []string {
	var patterns []string

	// Auth-related files
	patterns = append(patterns, "**/auth/**/*", "**/authentication/**/*", "**/authorization/**/*")

	// API and handlers
	patterns = append(patterns, "**/api/**/*", "**/handlers/**/*", "**/controllers/**/*")

	// Database and models
	patterns = append(patterns, "**/db/**/*", "**/database/**/*", "**/models/**/*")

	// Config files
	patterns = append(patterns, "**/*.env", "**/*.env.*", "**/config/**/*")

	return patterns
}

// formatCategoryName formats a category name for display
func formatCategoryName(category string) string {
	switch category {
	case "code-style":
		return "Code Style"
	case "naming":
		return "Naming Conventions"
	case "imports":
		return "Import Style"
	case "structure":
		return "Project Structure"
	case "error-handling":
		return "Error Handling"
	case "custom":
		return "Custom Conventions"
	default:
		// Simple title case without deprecated strings.Title
		words := strings.Split(strings.ReplaceAll(category, "-", " "), " ")
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
		return strings.Join(words, " ")
	}
}
