package generator

import (
	"fmt"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// generateAgents creates agent files based on the detected tech stack
func (g *ClaudeCodeGenerator) generateAgents(analysis *types.Analysis) []types.GeneratedFile {
	var files []types.GeneratedFile

	// Build context for context-aware generation
	ctx := BuildContext(analysis)

	// Generate tech-stack specific reviewers
	files = append(files, g.generateTechStackReviewers(analysis, ctx)...)

	// Generate generic agents (always included)
	files = append(files, g.generateGenericAgents(analysis, ctx)...)

	return files
}

// generateTechStackReviewers creates language-specific code reviewer agents
func (g *ClaudeCodeGenerator) generateTechStackReviewers(analysis *types.Analysis, ctx *GeneratorContext) []types.GeneratedFile {
	var files []types.GeneratedFile

	// Go reviewer
	if hasLanguage(analysis, "Go") {
		files = append(files, types.GeneratedFile{
			Path:    ".claude/agents/go-reviewer.md",
			Content: []byte(goReviewerContent(analysis, ctx)),
		})
	}

	// TypeScript/JavaScript reviewer
	if hasLanguage(analysis, "TypeScript") || hasLanguage(analysis, "JavaScript") {
		files = append(files, types.GeneratedFile{
			Path:    ".claude/agents/ts-reviewer.md",
			Content: []byte(tsReviewerContent(analysis, ctx)),
		})
	}

	// Python reviewer
	if hasLanguage(analysis, "Python") {
		files = append(files, types.GeneratedFile{
			Path:    ".claude/agents/python-reviewer.md",
			Content: []byte(pythonReviewerContent(analysis, ctx)),
		})
	}

	// Rust reviewer
	if hasLanguage(analysis, "Rust") {
		files = append(files, types.GeneratedFile{
			Path:    ".claude/agents/rust-reviewer.md",
			Content: []byte(rustReviewerContent(analysis, ctx)),
		})
	}

	// Java reviewer
	if hasLanguage(analysis, "Java") {
		files = append(files, types.GeneratedFile{
			Path:    ".claude/agents/java-reviewer.md",
			Content: []byte(javaReviewerContent(analysis, ctx)),
		})
	}

	return files
}

// generateGenericAgents creates agents that are useful for any project
func (g *ClaudeCodeGenerator) generateGenericAgents(analysis *types.Analysis, ctx *GeneratorContext) []types.GeneratedFile {
	return []types.GeneratedFile{
		{
			Path:    ".claude/agents/planner.md",
			Content: []byte(plannerAgentContent(analysis, ctx)),
		},
		{
			Path:    ".claude/agents/security-reviewer.md",
			Content: []byte(securityReviewerContent(analysis, ctx)),
		},
	}
}

func goReviewerContent(analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter with new Claude Code fields
	content.WriteString("---\n")
	content.WriteString("name: go-reviewer\n")
	content.WriteString("description: Expert Go code reviewer. Use when reviewing Go code for quality, patterns, and best practices.\n")
	content.WriteString("tools: Read, Grep, Glob, Bash\n")
	content.WriteString("model: haiku\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Go Code Reviewer for %s\n\n", ctx.ProjectName))
	content.WriteString("You are an expert Go code reviewer for this project. When reviewing Go code, focus on:\n\n")

	// Project-specific error handling section
	content.WriteString("## Error Handling\n\n")
	if len(ctx.ErrorPatterns) > 0 || hasErrorHandlingConvention(analysis) {
		content.WriteString("This project uses explicit error checking. See examples:\n")
		errorFiles := findFilesWithPattern(analysis, "err != nil")
		for i, file := range errorFiles {
			if i >= 3 {
				break
			}
			content.WriteString(fmt.Sprintf("- `%s`\n", file))
		}
		content.WriteString("\n")
	}
	content.WriteString("- All errors must be checked (no `_ = err`)\n")
	content.WriteString("- Use `%w` for error wrapping to preserve error chain\n")
	content.WriteString("- Add context when wrapping errors\n\n")

	// Project-specific testing section
	if ctx.HasTestingContext() {
		content.WriteString("## Testing\n\n")
		if len(ctx.TestingPatterns) > 0 {
			content.WriteString("Detected testing patterns:\n")
			for pattern, files := range ctx.TestingPatterns {
				content.WriteString(fmt.Sprintf("- **%s**", pattern))
				if len(files) > 0 {
					content.WriteString(fmt.Sprintf(" - see `%s`", files[0]))
				}
				content.WriteString("\n")
			}
			content.WriteString("\n")
		}
		if len(ctx.TestFiles) > 0 {
			content.WriteString("Example test files:\n")
			for i, file := range ctx.TestFiles {
				if i >= 3 {
					break
				}
				content.WriteString(fmt.Sprintf("- `%s`\n", file))
			}
			content.WriteString("\n")
		}
		if ctx.TestCommand != "" {
			content.WriteString(fmt.Sprintf("Run tests: `%s`\n\n", ctx.TestCommand))
		}
	} else {
		content.WriteString("## Testing\n\n")
		content.WriteString("- Use table-driven tests for multiple cases\n")
		content.WriteString("- Use `t.Run()` for subtests\n")
		content.WriteString("- Use `t.Parallel()` for independent tests\n\n")
	}

	// Linting section with actual config
	if ctx.HasLintingContext() {
		content.WriteString("## Linting\n\n")
		if lintConfig := ctx.GetLintConfig(); lintConfig != "" {
			content.WriteString(fmt.Sprintf("Config: `%s`\n", lintConfig))
		}
		if ctx.LintCommand != "" {
			content.WriteString(fmt.Sprintf("Run: `%s`\n", ctx.LintCommand))
		}
		content.WriteString("\n")
	}

	// Standard guidelines
	content.WriteString("## Code Quality\n\n")
	content.WriteString("- Verify consistent use of gofmt/goimports formatting\n")
	content.WriteString("- Look for proper use of interfaces and composition\n")
	content.WriteString("- Check for race conditions in concurrent code\n")
	content.WriteString("- Ensure proper resource cleanup with defer\n\n")

	content.WriteString("## Naming Conventions\n\n")
	content.WriteString("- Exported names: PascalCase\n")
	content.WriteString("- Unexported names: camelCase\n")
	content.WriteString("- Acronyms consistently cased (HTTP, URL)\n")
	content.WriteString("- Interface names describe behavior (-er suffix)\n\n")

	content.WriteString("## Common Issues to Flag\n\n")
	content.WriteString("- Using panic for regular error handling\n")
	content.WriteString("- Global mutable state\n")
	content.WriteString("- Not closing resources (files, connections)\n")
	content.WriteString("- Mixing pointer and value receivers on same type\n")

	return content.String()
}

func tsReviewerContent(analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter with new Claude Code fields
	content.WriteString("---\n")
	content.WriteString("name: ts-reviewer\n")
	content.WriteString("description: Expert TypeScript/JavaScript code reviewer. Use when reviewing TS/JS code for quality and patterns.\n")
	content.WriteString("tools: Read, Grep, Glob, Bash\n")
	content.WriteString("model: haiku\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# TypeScript/JavaScript Code Reviewer for %s\n\n", ctx.ProjectName))
	content.WriteString("You are an expert TypeScript/JavaScript code reviewer for this project. When reviewing code, focus on:\n\n")

	// Project-specific state management
	if analysis.CodePatterns != nil && len(analysis.CodePatterns.StateManagement) > 0 {
		content.WriteString("## State Management\n\n")
		content.WriteString("This project uses:\n")
		for _, pattern := range analysis.CodePatterns.StateManagement {
			content.WriteString(fmt.Sprintf("- **%s**", pattern.Name))
			if len(pattern.Examples) > 0 {
				content.WriteString(fmt.Sprintf(" - see `%s`", pattern.Examples[0]))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Project-specific data fetching
	if analysis.CodePatterns != nil && len(analysis.CodePatterns.DataFetching) > 0 {
		content.WriteString("## Data Fetching\n\n")
		content.WriteString("Detected patterns:\n")
		for _, pattern := range analysis.CodePatterns.DataFetching {
			content.WriteString(fmt.Sprintf("- **%s**", pattern.Name))
			if len(pattern.Examples) > 0 {
				content.WriteString(fmt.Sprintf(" - see `%s`", pattern.Examples[0]))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Testing section
	if ctx.HasTestingContext() {
		content.WriteString("## Testing\n\n")
		if len(ctx.TestingPatterns) > 0 {
			content.WriteString("Detected testing patterns:\n")
			for pattern, files := range ctx.TestingPatterns {
				content.WriteString(fmt.Sprintf("- **%s**", pattern))
				if len(files) > 0 {
					content.WriteString(fmt.Sprintf(" - see `%s`", files[0]))
				}
				content.WriteString("\n")
			}
			content.WriteString("\n")
		}
		if testConfig := ctx.GetTestConfig(); testConfig != "" {
			content.WriteString(fmt.Sprintf("Test config: `%s`\n", testConfig))
		}
		if ctx.TestCommand != "" {
			content.WriteString(fmt.Sprintf("Run tests: `%s`\n", ctx.TestCommand))
		}
		content.WriteString("\n")
	}

	// Linting section
	if ctx.HasLintingContext() {
		content.WriteString("## Linting & Formatting\n\n")
		if eslint, ok := ctx.ConfigFiles["eslint"]; ok {
			content.WriteString(fmt.Sprintf("ESLint config: `%s`\n", eslint))
		}
		if prettier, ok := ctx.ConfigFiles["prettier"]; ok {
			content.WriteString(fmt.Sprintf("Prettier config: `%s`\n", prettier))
		}
		if ctx.LintCommand != "" {
			content.WriteString(fmt.Sprintf("Run lint: `%s`\n", ctx.LintCommand))
		}
		content.WriteString("\n")
	}

	// Standard type safety guidelines
	content.WriteString("## Type Safety (TypeScript)\n\n")
	content.WriteString("- Avoid using 'any' type - use 'unknown' or proper types\n")
	content.WriteString("- Ensure strict null checks are handled\n")
	content.WriteString("- Use discriminated unions for complex state\n")
	content.WriteString("- Prefer interfaces over type aliases for object shapes\n\n")

	content.WriteString("## Code Quality\n\n")
	content.WriteString("- Use const by default, let when reassignment needed\n")
	content.WriteString("- Prefer arrow functions for callbacks\n")
	content.WriteString("- Use async/await over raw promises\n")
	content.WriteString("- Handle promise rejections properly\n\n")

	// Add React-specific guidelines if detected
	if hasFramework(analysis, "React") || hasFramework(analysis, "Next.js") {
		content.WriteString("## React Specific\n\n")
		content.WriteString("- Use functional components with hooks\n")
		content.WriteString("- Follow Rules of Hooks (no conditional hooks)\n")
		content.WriteString("- Memoize expensive computations with useMemo\n")
		content.WriteString("- Use useCallback for event handlers passed to children\n")
		content.WriteString("- Keep components focused and small\n\n")
	}

	// Add Vue-specific guidelines if detected
	if hasFramework(analysis, "Vue.js") || hasFramework(analysis, "Nuxt.js") {
		content.WriteString("## Vue Specific\n\n")
		content.WriteString("- Use Composition API with <script setup>\n")
		content.WriteString("- Keep reactive state minimal\n")
		content.WriteString("- Use computed properties for derived state\n")
		content.WriteString("- Use watchEffect sparingly\n\n")
	}

	content.WriteString("## Common Issues to Flag\n\n")
	content.WriteString("- Memory leaks from uncleared intervals/subscriptions\n")
	content.WriteString("- Missing dependency arrays in useEffect/useMemo\n")
	content.WriteString("- Mutating state directly\n")
	content.WriteString("- Not handling loading/error states\n")

	return content.String()
}

func pythonReviewerContent(analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter with new Claude Code fields
	content.WriteString("---\n")
	content.WriteString("name: python-reviewer\n")
	content.WriteString("description: Expert Python code reviewer. Use when reviewing Python code for quality and best practices.\n")
	content.WriteString("tools: Read, Grep, Glob, Bash\n")
	content.WriteString("model: haiku\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Python Code Reviewer for %s\n\n", ctx.ProjectName))
	content.WriteString("You are an expert Python code reviewer for this project. When reviewing Python code, focus on:\n\n")

	// Testing section with context
	if ctx.HasTestingContext() {
		content.WriteString("## Testing\n\n")
		if len(ctx.TestingPatterns) > 0 {
			content.WriteString("Detected testing patterns:\n")
			for pattern, files := range ctx.TestingPatterns {
				content.WriteString(fmt.Sprintf("- **%s**", pattern))
				if len(files) > 0 {
					content.WriteString(fmt.Sprintf(" - see `%s`", files[0]))
				}
				content.WriteString("\n")
			}
			content.WriteString("\n")
		}
		if len(ctx.TestFiles) > 0 {
			content.WriteString("Example test files:\n")
			for i, file := range ctx.TestFiles {
				if i >= 3 {
					break
				}
				content.WriteString(fmt.Sprintf("- `%s`\n", file))
			}
			content.WriteString("\n")
		}
		if ctx.TestCommand != "" {
			content.WriteString(fmt.Sprintf("Run tests: `%s`\n\n", ctx.TestCommand))
		}
	}

	// Linting section
	if ctx.HasLintingContext() {
		content.WriteString("## Linting & Formatting\n\n")
		if pyproject, ok := ctx.ConfigFiles["pyproject"]; ok {
			content.WriteString(fmt.Sprintf("Config: `%s`\n", pyproject))
		}
		if ctx.LintCommand != "" {
			content.WriteString(fmt.Sprintf("Run lint: `%s`\n", ctx.LintCommand))
		}
		if ctx.FormatCommand != "" {
			content.WriteString(fmt.Sprintf("Format: `%s`\n", ctx.FormatCommand))
		}
		content.WriteString("\n")
	}

	// Standard guidelines
	content.WriteString("## Code Quality\n\n")
	content.WriteString("- Follow PEP 8 style guidelines\n")
	content.WriteString("- Use type hints for function signatures\n")
	content.WriteString("- Write docstrings for public functions/classes\n")
	content.WriteString("- Keep functions focused and small\n\n")

	content.WriteString("## Type Hints\n\n")
	content.WriteString("- Add type hints to function parameters and return values\n")
	content.WriteString("- Use Optional[] for nullable values\n")
	content.WriteString("- Use Union[] or | for multiple types (Python 3.10+)\n")
	content.WriteString("- Consider using TypedDict for complex dictionaries\n\n")

	content.WriteString("## Error Handling\n\n")
	content.WriteString("- Use specific exception types\n")
	content.WriteString("- Don't use bare except clauses\n")
	content.WriteString("- Use context managers (with statements) for resources\n")
	content.WriteString("- Provide helpful error messages\n\n")

	content.WriteString("## Common Issues to Flag\n\n")
	content.WriteString("- Mutable default arguments\n")
	content.WriteString("- Global variables\n")
	content.WriteString("- Overly long functions\n")
	content.WriteString("- Missing error handling\n")
	content.WriteString("- Not using context managers for files/connections\n")
	content.WriteString("- Circular imports\n")

	return content.String()
}

func rustReviewerContent(analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter with new Claude Code fields
	content.WriteString("---\n")
	content.WriteString("name: rust-reviewer\n")
	content.WriteString("description: Expert Rust code reviewer. Use when reviewing Rust code for safety, ownership, and patterns.\n")
	content.WriteString("tools: Read, Grep, Glob, Bash\n")
	content.WriteString("model: haiku\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Rust Code Reviewer for %s\n\n", ctx.ProjectName))
	content.WriteString("You are an expert Rust code reviewer for this project. When reviewing Rust code, focus on:\n\n")

	// Testing section
	if ctx.HasTestingContext() {
		content.WriteString("## Testing\n\n")
		if len(ctx.TestFiles) > 0 {
			content.WriteString("Example test files:\n")
			for i, file := range ctx.TestFiles {
				if i >= 3 {
					break
				}
				content.WriteString(fmt.Sprintf("- `%s`\n", file))
			}
			content.WriteString("\n")
		}
		if ctx.TestCommand != "" {
			content.WriteString(fmt.Sprintf("Run tests: `%s`\n\n", ctx.TestCommand))
		}
	}

	// Standard Rust guidelines
	content.WriteString("## Ownership and Borrowing\n\n")
	content.WriteString("- Minimize cloning when borrowing would suffice\n")
	content.WriteString("- Use references appropriately\n")
	content.WriteString("- Understand lifetime annotations\n")
	content.WriteString("- Avoid unnecessary Box/Rc/Arc usage\n\n")

	content.WriteString("## Error Handling\n\n")
	content.WriteString("- Use Result<T, E> for recoverable errors\n")
	content.WriteString("- Use the ? operator for error propagation\n")
	content.WriteString("- Create custom error types when appropriate\n")
	content.WriteString("- Use thiserror or anyhow for error handling\n\n")

	content.WriteString("## Safety\n\n")
	content.WriteString("- Minimize unsafe code blocks\n")
	content.WriteString("- Document safety invariants for unsafe code\n")
	content.WriteString("- Prefer safe abstractions\n\n")

	content.WriteString("## Common Issues to Flag\n\n")
	content.WriteString("- Unnecessary cloning\n")
	content.WriteString("- Not using pattern matching effectively\n")
	content.WriteString("- Ignoring compiler warnings\n")
	content.WriteString("- Missing documentation on public items\n")
	content.WriteString("- Using unwrap() in library code\n")

	return content.String()
}

func javaReviewerContent(analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter with new Claude Code fields
	content.WriteString("---\n")
	content.WriteString("name: java-reviewer\n")
	content.WriteString("description: Expert Java code reviewer. Use when reviewing Java code for quality and patterns.\n")
	content.WriteString("tools: Read, Grep, Glob, Bash\n")
	content.WriteString("model: haiku\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Java Code Reviewer for %s\n\n", ctx.ProjectName))
	content.WriteString("You are an expert Java code reviewer for this project. When reviewing Java code, focus on:\n\n")

	// Testing section
	if ctx.HasTestingContext() {
		content.WriteString("## Testing\n\n")
		if len(ctx.TestingPatterns) > 0 {
			content.WriteString("Detected testing patterns:\n")
			for pattern, files := range ctx.TestingPatterns {
				content.WriteString(fmt.Sprintf("- **%s**", pattern))
				if len(files) > 0 {
					content.WriteString(fmt.Sprintf(" - see `%s`", files[0]))
				}
				content.WriteString("\n")
			}
			content.WriteString("\n")
		}
		if ctx.TestCommand != "" {
			content.WriteString(fmt.Sprintf("Run tests: `%s`\n\n", ctx.TestCommand))
		}
	}

	// Standard Java guidelines
	content.WriteString("## Code Quality\n\n")
	content.WriteString("- Follow Java naming conventions\n")
	content.WriteString("- Use appropriate access modifiers\n")
	content.WriteString("- Prefer composition over inheritance\n")
	content.WriteString("- Keep methods focused and small\n\n")

	content.WriteString("## Null Safety\n\n")
	content.WriteString("- Use Optional for potentially null returns\n")
	content.WriteString("- Add null checks for parameters\n")
	content.WriteString("- Consider using @Nullable/@NonNull annotations\n")
	content.WriteString("- Return empty collections instead of null\n\n")

	content.WriteString("## Error Handling\n\n")
	content.WriteString("- Use specific exception types\n")
	content.WriteString("- Don't catch Exception broadly\n")
	content.WriteString("- Clean up resources with try-with-resources\n")
	content.WriteString("- Document thrown exceptions in Javadoc\n\n")

	content.WriteString("## Common Issues to Flag\n\n")
	content.WriteString("- Mutable static fields\n")
	content.WriteString("- Not closing resources\n")
	content.WriteString("- Catching and swallowing exceptions\n")
	content.WriteString("- Using raw types instead of generics\n")
	content.WriteString("- God classes or methods\n")
	content.WriteString("- Missing input validation\n")

	return content.String()
}

func plannerAgentContent(analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter with new Claude Code fields
	content.WriteString("---\n")
	content.WriteString("name: planner\n")
	content.WriteString("description: Planning agent for breaking down complex tasks. Use when designing features or planning implementations.\n")
	content.WriteString("tools: Read, Grep, Glob\n")
	content.WriteString("model: sonnet\n")
	content.WriteString("permissionMode: plan\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Planner Agent for %s\n\n", ctx.ProjectName))
	content.WriteString("You are a planning agent that helps break down complex tasks into manageable steps.\n\n")

	// Project-specific context
	if ctx.HasArchitectureContext() {
		content.WriteString("## Project Structure\n\n")
		if ctx.EntryPoint != "" {
			content.WriteString(fmt.Sprintf("Entry point: `%s`\n", ctx.EntryPoint))
		}
		if len(ctx.Layers) > 0 {
			content.WriteString("\nArchitectural layers:\n")
			for _, layer := range ctx.Layers {
				content.WriteString(fmt.Sprintf("- **%s**", layer.Name))
				if layer.Purpose != "" {
					content.WriteString(fmt.Sprintf(": %s", layer.Purpose))
				}
				content.WriteString("\n")
			}
		}
		content.WriteString("\n")
	}

	// Key files for planning
	if len(analysis.KeyFiles) > 0 {
		content.WriteString("## Key Files\n\n")
		content.WriteString("Important files to consider when planning:\n")
		count := 0
		for _, kf := range analysis.KeyFiles {
			if count >= 5 {
				break
			}
			content.WriteString(fmt.Sprintf("- `%s` - %s\n", kf.Path, kf.Purpose))
			count++
		}
		content.WriteString("\n")
	}

	// Available commands for verification
	if ctx.BuildCommand != "" || ctx.TestCommand != "" || ctx.LintCommand != "" {
		content.WriteString("## Verification Commands\n\n")
		if ctx.BuildCommand != "" {
			content.WriteString(fmt.Sprintf("- Build: `%s`\n", ctx.BuildCommand))
		}
		if ctx.TestCommand != "" {
			content.WriteString(fmt.Sprintf("- Test: `%s`\n", ctx.TestCommand))
		}
		if ctx.LintCommand != "" {
			content.WriteString(fmt.Sprintf("- Lint: `%s`\n", ctx.LintCommand))
		}
		content.WriteString("\n")
	}

	// Standard planning process
	content.WriteString("## Planning Process\n\n")
	content.WriteString("1. **Understand**: Clarify the requirements and goals\n")
	content.WriteString("2. **Research**: Identify relevant files and patterns in the codebase\n")
	content.WriteString("3. **Design**: Outline the high-level approach\n")
	content.WriteString("4. **Decompose**: Break into specific implementation steps\n")
	content.WriteString("5. **Validate**: Review the plan for completeness\n\n")

	content.WriteString("## Output Format\n\n")
	content.WriteString("When planning, provide:\n")
	content.WriteString("- Clear numbered steps\n")
	content.WriteString("- Files that need to be modified/created\n")
	content.WriteString("- Dependencies between steps\n")
	content.WriteString("- Potential risks or considerations\n")
	content.WriteString("- Testing strategy\n\n")

	content.WriteString("## Best Practices\n\n")
	content.WriteString("- Start with the simplest approach that could work\n")
	content.WriteString("- Consider backward compatibility\n")
	content.WriteString("- Think about error handling early\n")
	content.WriteString("- Plan for testing alongside implementation\n")
	content.WriteString("- Identify when to ask for clarification\n")

	return content.String()
}

func securityReviewerContent(analysis *types.Analysis, ctx *GeneratorContext) string {
	var content strings.Builder

	// YAML frontmatter with new Claude Code fields
	content.WriteString("---\n")
	content.WriteString("name: security-reviewer\n")
	content.WriteString("description: Security-focused code reviewer. Use when auditing code for vulnerabilities and security issues.\n")
	content.WriteString("tools: Read, Grep, Glob, Bash\n")
	content.WriteString("model: sonnet\n")
	content.WriteString("skills:\n")
	content.WriteString("  - lint\n")
	content.WriteString("  - test\n")
	content.WriteString("---\n\n")

	content.WriteString(fmt.Sprintf("# Security Reviewer for %s\n\n", ctx.ProjectName))
	content.WriteString("You are a security-focused code reviewer for this project. When reviewing code, focus on:\n\n")

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

	// API patterns that may have security implications
	if len(ctx.APIPatterns) > 0 {
		content.WriteString("## API Security\n\n")
		content.WriteString("Detected API patterns to review:\n")
		for pattern, files := range ctx.APIPatterns {
			content.WriteString(fmt.Sprintf("- **%s**", pattern))
			if len(files) > 0 {
				content.WriteString(fmt.Sprintf(" - see `%s`", files[0]))
			}
			content.WriteString("\n")
		}
		content.WriteString("\n")
	}

	// Standard security guidelines
	content.WriteString("## Input Validation\n\n")
	content.WriteString("- Validate all user input\n")
	content.WriteString("- Sanitize data before using in queries or output\n")
	content.WriteString("- Use allowlists over blocklists when possible\n")
	content.WriteString("- Check for proper encoding/escaping\n\n")

	content.WriteString("## Authentication & Authorization\n\n")
	content.WriteString("- Verify authentication is required where needed\n")
	content.WriteString("- Check authorization for all protected resources\n")
	content.WriteString("- Look for privilege escalation vulnerabilities\n")
	content.WriteString("- Ensure secure session management\n\n")

	content.WriteString("## Data Protection\n\n")
	content.WriteString("- Check for sensitive data exposure in logs\n")
	content.WriteString("- Verify encryption for sensitive data at rest\n")
	content.WriteString("- Ensure secure transmission (HTTPS)\n")
	content.WriteString("- Look for hardcoded secrets or credentials\n\n")

	content.WriteString("## Common Vulnerabilities (OWASP Top 10)\n\n")
	content.WriteString("- SQL Injection\n")
	content.WriteString("- Cross-Site Scripting (XSS)\n")
	content.WriteString("- Broken Authentication\n")
	content.WriteString("- Sensitive Data Exposure\n")
	content.WriteString("- Broken Access Control\n")
	content.WriteString("- Security Misconfiguration\n")
	content.WriteString("- Insecure Deserialization\n\n")

	// Language-specific security
	if hasLanguage(analysis, "Go") {
		content.WriteString("## Go-Specific Security\n\n")
		content.WriteString("- Check for SQL injection in database queries\n")
		content.WriteString("- Verify proper use of html/template for HTML output\n")
		content.WriteString("- Check for command injection in os/exec calls\n")
		content.WriteString("- Ensure proper TLS configuration\n\n")
	}

	if hasLanguage(analysis, "TypeScript") || hasLanguage(analysis, "JavaScript") {
		content.WriteString("## JavaScript/TypeScript Security\n\n")
		content.WriteString("- Check for XSS vulnerabilities\n")
		content.WriteString("- Verify CSRF protection\n")
		content.WriteString("- Look for prototype pollution\n")
		content.WriteString("- Check for insecure dependencies (npm audit)\n")
		content.WriteString("- Verify Content Security Policy usage\n\n")
	}

	if hasLanguage(analysis, "Python") {
		content.WriteString("## Python Security\n\n")
		content.WriteString("- Check for SQL injection (use parameterized queries)\n")
		content.WriteString("- Verify pickle usage (avoid untrusted data)\n")
		content.WriteString("- Check for command injection\n")
		content.WriteString("- Verify YAML safe_load usage\n")
		content.WriteString("- Check for path traversal vulnerabilities\n\n")
	}

	content.WriteString("## Sensitive Files\n\n")
	content.WriteString("Never commit these files:\n")
	content.WriteString("- `.env` files with secrets\n")
	content.WriteString("- Private keys (*.pem, *.key)\n")
	content.WriteString("- Credentials files\n")
	content.WriteString("- Database connection strings with passwords\n")

	return content.String()
}

// Helper functions

// hasErrorHandlingConvention checks if error handling convention is detected
func hasErrorHandlingConvention(analysis *types.Analysis) bool {
	for _, conv := range analysis.Conventions {
		if conv.Category == "error-handling" {
			return true
		}
	}
	return false
}

// findFilesWithPattern finds files that contain a specific pattern
// Uses the KeyFiles and CodePatterns as proxy since we don't have direct file search
func findFilesWithPattern(analysis *types.Analysis, pattern string) []string {
	var files []string

	// Use KeyFiles as they're likely to have important patterns
	for _, kf := range analysis.KeyFiles {
		if strings.HasSuffix(kf.Path, ".go") {
			files = append(files, kf.Path)
		}
	}

	// Add entry points
	if analysis.ArchitectureInfo != nil && analysis.ArchitectureInfo.EntryPoint != "" {
		files = append(files, analysis.ArchitectureInfo.EntryPoint)
	}

	// Deduplicate and limit
	seen := make(map[string]bool)
	unique := []string{}
	for _, f := range files {
		if !seen[f] {
			seen[f] = true
			unique = append(unique, f)
		}
	}

	if len(unique) > 3 {
		unique = unique[:3]
	}

	return unique
}
