# Argus - AI Assistant Instructions

## Project Overview

Argus is a CLI tool that scans codebases and generates optimized context files for AI coding assistants (Claude Code, Cursor, Copilot, Continue).

**Core value:** Automate the creation of CLAUDE.md, .cursorrules, and similar files so developers don't have to manually write and maintain them.

## Local Tracking Files

**IMPORTANT:** Always maintain these local files while working on this project.

### Files to Update (gitignored, local only)

| File | Purpose | When to Update |
|------|---------|----------------|
| `.todos.md` | Active tasks and progress | Before starting, after completing, when planning |
| `.brainstorm.md` | Ideas, designs, decisions | When discussing new features or approaches |
| `.notes.md` | Session notes, blockers | During work sessions |

### Update Format

Always include timestamps in ISO format:

```markdown
## 2026-01-19T10:30:00

### Completed
- [x] Task description

### In Progress
- [ ] Task description

### Notes
- Decision made: xyz
```

## Project Structure

```
argus/
├── cmd/argus/          # CLI entry point (Cobra)
├── internal/
│   ├── analyzer/       # Codebase analysis logic
│   ├── detector/       # Tech stack & pattern detection
│   └── generator/      # Output file generation
├── pkg/types/          # Shared types
├── docs/               # Documentation
└── [local files]       # Not committed
    ├── .todos.md
    ├── .brainstorm.md
    └── .notes.md
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      CLI (Cobra)                        │
│                   cmd/argus/main.go                     │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                     Analyzer                            │
│              internal/analyzer/                         │
│  - Walks file tree                                      │
│  - Collects file metadata                               │
│  - Coordinates detectors                                │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                    Detectors                            │
│              internal/detector/                         │
│  - TechStackDetector (frameworks, languages)            │
│  - PatternDetector (conventions, naming)                │
│  - StructureDetector (directory layout)                 │
│  - DependencyDetector (package managers)                │
└─────────────────────┬───────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────┐
│                    Generator                            │
│              internal/generator/                        │
│  - ClaudeGenerator (CLAUDE.md)                          │
│  - CursorGenerator (.cursorrules)                       │
│  - CopilotGenerator (copilot-instructions.md)           │
│  - ContinueGenerator (config.json)                      │
└─────────────────────────────────────────────────────────┘
```

## Code Style

- Go 1.21+
- Use `gofmt` for formatting
- Follow Effective Go guidelines
- Keep functions small and focused
- Error wrapping with `fmt.Errorf("context: %w", err)`

## Build & Test

```bash
go build -o argus ./cmd/argus    # Build
go test ./...                     # Test
./argus --help                    # Run
```

## Git Workflow

### Branch Naming

| Prefix | Purpose | Example |
|--------|---------|---------|
| `feat/` | New feature | `feat/tech-detection` |
| `fix/` | Bug fix | `fix/go-mod-parsing` |
| `refactor/` | Code refactoring | `refactor/detector-interface` |
| `docs/` | Documentation | `docs/readme-examples` |
| `test/` | Adding tests | `test/analyzer-unit` |
| `chore/` | Maintenance | `chore/update-deps` |

### Commit Convention

```
<type>(<scope>): <description>

Types: feat, fix, docs, style, refactor, test, chore
Scopes: analyzer, detector, generator, cli, config
```

## Key Design Decisions

1. **No AI required for MVP** — Static analysis only, AI enhancement later
2. **Multiple output formats** — Generate for all major AI tools
3. **Convention over configuration** — Works out of box, config is optional
4. **Fast** — Should scan large codebases in seconds

## When Working on Tasks

### Before Starting
1. Read `.todos.md` to understand current state
2. Add new task with timestamp if not exists
3. Mark task as "In Progress"

### After Completing
1. Mark task as completed with timestamp
2. Add any follow-up tasks discovered
3. Update `.brainstorm.md` if new ideas emerged
