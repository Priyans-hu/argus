---
sidebar_position: 1
title: Scan Command
description: Analyze your codebase with argus scan
---

# Scan Command

The `scan` command analyzes your codebase and generates a `CLAUDE.md` context file.

## Basic Usage

```bash
# Scan current directory
argus scan .

# Scan a specific path
argus scan /path/to/project
```

## Options

| Flag | Short | Description |
|------|-------|-------------|
| `--verbose` | `-v` | Show detailed analysis output |
| `--dry-run` | `-n` | Preview output without writing files |
| `--merge` | `-m` | Keep your text around the generated block when regenerating (default `true`; `--merge=false` replaces the whole file) |
| `--add-custom` | | Add an empty CUSTOM block for your notes to CLAUDE.md, `.cursorrules` and the Copilot file |
| `--force` | | Overwrite files argus would otherwise skip: edited `.claude/` and `.continue/config.yaml` files, and files with unbalanced or repeated `ARGUS:AUTO` markers |

See [Output Format](./output.md#re-running-argus) for what a re-scan keeps and what it skips.

## Examples

### Preview Output

```bash
argus scan . --dry-run
```

This shows what would be written without creating any files.

### Verbose Mode

```bash
argus scan . -v
```

Shows detailed progress:
- Files being scanned
- Patterns detected
- Time taken for each phase

### Scan Specific Directory

```bash
argus scan ./backend
```

## What Gets Analyzed

argus performs a comprehensive analysis:

1. **File Discovery** - Walks the directory tree, respecting `.gitignore`
2. **Tech Stack Detection** - Identifies languages, frameworks, libraries
3. **Architecture Analysis** - Understands project structure and patterns
4. **Command Discovery** - Finds build, test, lint, and run commands
5. **Convention Detection** - Discovers coding standards and practices
6. **Git Analysis** - Examines commit patterns and branch conventions

## Output Structure

The generated `CLAUDE.md` includes:

```
├── Project Overview
├── Quick Reference (commands)
├── Architecture
│   ├── Style (monolith, microservices, etc.)
│   ├── Entry Point
│   └── Package Dependencies
├── Tech Stack
│   ├── Languages
│   ├── Frameworks
│   └── Tools
├── Project Structure
├── Key Files
├── Configuration
├── Development Setup
├── Coding Conventions
├── Detected Patterns
└── Dependencies
```

## Performance

argus uses parallel analysis for better performance:

- Phase 1: Essential detectors (tech stack, structure)
- Phase 2: All other detectors run concurrently
- Typical scan time: 1-5 seconds for most projects
