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
| `--output` | `-o` | Custom output file path |
| `--config` | `-c` | Path to config file |

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

### Custom Output Path

```bash
argus scan . -o docs/AI_CONTEXT.md
```

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
