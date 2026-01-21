<!-- ARGUS:AUTO -->
# argus

The all-seeing code analyzer. Generates optimized context files for AI coding assistants.

## Quick Reference

```bash
# Development
make setup          # Install dev dependencies (golangci-lint, goimports)
make build          # Build binary to bin/argus
make test           # Run all tests
make lint           # Run golangci-lint
make fmt            # Format code (gofmt + goimports)

# Usage
argus init          # Initialize .argus.yaml config
argus scan          # Analyze codebase and generate context files
argus sync          # Regenerate using existing config
argus watch         # Watch mode - auto-regenerate on changes
argus upgrade       # Self-upgrade to latest version
```

## Architecture

### Overview

Argus follows the **Standard Go Layout** pattern with a clear separation between CLI, core logic, and output generation.

```
                         ┌─────────────────┐
                         │   CLI (cobra)   │
                         │  cmd/argus/     │
                         └────────┬────────┘
                                  │
                    ┌─────────────┼─────────────┐
                    │             │             │
                    ▼             ▼             ▼
             ┌──────────┐  ┌──────────┐  ┌──────────┐
             │ analyzer │  │  config  │  │  merger  │
             └────┬─────┘  └──────────┘  └──────────┘
                  │
     ┌────────────┼────────────┐
     │            │            │
     ▼            ▼            ▼
┌─────────┐ ┌──────────┐ ┌───────────┐
│detector │ │generator │ │  walker   │
└─────────┘ └──────────┘ └───────────┘
```

### Data Flow

1. **CLI** (`cmd/argus/main.go`) - Parses commands and flags via Cobra
2. **Config** (`internal/config/`) - Loads `.argus.yaml` configuration
3. **Analyzer** (`internal/analyzer/`) - Orchestrates the analysis pipeline:
   - Creates a `Walker` to traverse the file tree
   - Invokes multiple **Detectors** to extract information
   - Passes results to **Generators** for output
4. **Detector** (`internal/detector/`) - Specialized modules that detect:
   - Tech stack (languages, frameworks)
   - Project structure and architecture
   - Code patterns and conventions
   - Git conventions (commit style, branch naming)
   - API endpoints
5. **Generator** (`internal/generator/`) - Creates output files:
   - `claude.go` → `CLAUDE.md`
   - `cursor.go` → `.cursorrules`
   - `copilot.go` → `.github/copilot-instructions.md`
   - `claudecode*.go` → `.claude/` directory structure
6. **Merger** (`internal/merger/`) - Preserves custom sections during regeneration

## Project Structure

```
argus/
├── cmd/argus/              # CLI entry point
│   └── main.go             # Cobra commands: init, scan, sync, watch, upgrade
├── internal/
│   ├── analyzer/           # Core orchestration
│   │   ├── analyzer.go     # Main analysis pipeline
│   │   └── walker.go       # File tree traversal
│   ├── config/             # Configuration management
│   │   └── config.go       # .argus.yaml loading/saving
│   ├── detector/           # Detection modules
│   │   ├── architecture.go # Architecture style detection
│   │   ├── codepatterns.go # Code pattern scanning
│   │   ├── convention.go   # Convention detection
│   │   ├── endpoints.go    # API endpoint detection
│   │   ├── frameworks.go   # Framework-specific patterns
│   │   ├── git.go          # Git conventions
│   │   ├── monorepo.go     # Monorepo detection
│   │   ├── patterns.go     # General patterns
│   │   ├── readme.go       # README parsing
│   │   ├── structure.go    # Directory structure
│   │   └── techstack.go    # Language/framework detection
│   ├── generator/          # Output generators
│   │   ├── claude.go       # CLAUDE.md generator
│   │   ├── claudecode*.go  # .claude/ directory generators
│   │   ├── copilot.go      # GitHub Copilot instructions
│   │   ├── cursor.go       # Cursor rules generator
│   │   └── context_builder.go
│   └── merger/             # Content merging
│       └── merger.go       # Preserve custom sections
├── pkg/types/              # Shared type definitions
│   └── types.go            # Analysis, TechStack, Convention, etc.
├── .githooks/              # Git hooks
│   └── pre-commit          # Auto-format + lint + test
├── Makefile                # Build automation
├── go.mod                  # Go module definition
└── .argus.yaml             # Self-configuration (dogfooding)
```

## Configuration System

### File: `.argus.yaml`

```yaml
# Output formats to generate
output:
  - claude          # CLAUDE.md
  - cursor          # .cursorrules
  - copilot         # .github/copilot-instructions.md
  - claude-code     # .claude/ directory

# Patterns to ignore (in addition to .gitignore)
ignore:
  - node_modules
  - vendor
  - dist
  - "*.log"

# Custom conventions to include in output
custom_conventions:
  - "Use table-driven tests"
  - "All exported functions must have doc comments"

# Override auto-detected values
overrides:
  project_name: "My Project"
  framework: "Custom Framework"

# Claude Code specific settings
claude_code:
  agents: true      # Generate .claude/agents/*.md
  commands: true    # Generate .claude/commands/*.md
  rules: true       # Generate .claude/rules/*.md
  mcp: true         # Generate .claude/mcp.json
```

### Config Loading

- Config is loaded from `.argus.yaml` in the target directory
- Falls back to sensible defaults if file doesn't exist
- CLI flags override config file values

## Development Setup

### Prerequisites

- Go 1.24+
- Git

### Initial Setup

```bash
# Clone repository
git clone https://github.com/Priyans-hu/argus.git
cd argus

# Install development tools and configure git hooks
make setup

# This installs:
# - golangci-lint (linter)
# - goimports (import organizer)
# - Configures .githooks/pre-commit
```

### Git Hooks

The pre-commit hook (`.githooks/pre-commit`) automatically:
1. Runs `gofmt` on staged Go files
2. Runs `goimports` if available
3. Runs `golangci-lint --fix` for auto-fixable issues
4. Runs tests with `go test ./... -short`

Hooks are configured via:
```bash
git config core.hooksPath .githooks
```

### Linting

Uses `golangci-lint` with default configuration (no `.golangci.yml`).

```bash
make lint           # Check for issues
make lint-fix       # Auto-fix what's possible
```

## CLI Output & Verbosity

Argus is a CLI tool that outputs to stdout/stderr. No logging library is used.

### Output Levels

| Flag | Output |
|------|--------|
| (none) | Progress indicators, success/error messages |
| `-v, --verbose` | Detailed analysis results, file-by-file processing |
| `-n, --dry-run` | Preview output without writing files |

### Output Indicators

- `🔍` - Scanning/analyzing
- `🔄` - Syncing/regenerating
- `✅` - Success
- `⚠️` - Warning
- `👁️` - Watch mode active
- `📊` - Analysis results (verbose)
- `📄` - File preview (dry-run)

### Example Verbose Output

```bash
$ argus scan -v

🔍 Scanning /path/to/project...

📊 Analysis Results:
   Project: myproject
   Languages: 2
   Frameworks: 3
   Directories: 15
   Key Files: 8
   Commands: 12
   Conventions: 25

✅ Generated CLAUDE.md
```

## Tech Stack

- **Language:** Go 1.24
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **File Watching:** [fsnotify](https://github.com/fsnotify/fsnotify)
- **Config Parsing:** [yaml.v3](https://gopkg.in/yaml.v3)
- **CI/CD:** GitHub Actions
- **Releases:** GoReleaser

## Testing

```bash
make test           # Run all tests
make test-v         # Verbose test output
go test ./... -run TestName  # Run specific test
```

Tests are colocated with source files (`*_test.go`).

## Coding Conventions

### Style
- Format with `gofmt` / `goimports`
- No custom linter config - use golangci-lint defaults

### Error Handling
```go
if err != nil {
    return fmt.Errorf("context: %w", err)
}
```

### Documentation
```go
// FunctionName does something specific.
// It returns an error if something goes wrong.
func FunctionName() error { ... }
```

### Git Commits
- Style: **Conventional Commits**
- Format: `<type>(<scope>): <description>`
- Types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`
- Scopes: `cli`, `analyzer`, `detector`, `generator`, `config`
- Example: `feat(detector): add monorepo detection`

### Branch Naming
- `feat/description` - New features
- `fix/description` - Bug fixes
- `chore/description` - Maintenance tasks
<!-- /ARGUS:AUTO -->

<!-- ARGUS:CUSTOM -->
## My Custom Notes

This is a test custom section that should be preserved.
<!-- /ARGUS:CUSTOM -->
