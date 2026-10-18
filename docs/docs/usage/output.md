---
sidebar_position: 3
title: Output Format
description: Understanding the AGENTS.md and CLAUDE.md output
---

# Output Format

argus generates an `AGENTS.md` file with structured sections that AI assistants can easily parse and understand, and a `CLAUDE.md` that imports it.

## File Structure

The generated content sits between `ARGUS:AUTO` markers. Everything outside them is yours and is kept on every re-scan:

```markdown
... your notes (kept on re-scan) ...

<!-- ARGUS:AUTO -->
... auto-generated content ...
<!-- /ARGUS:AUTO -->

<!-- ARGUS:CUSTOM -->
... more of your notes (kept on re-scan) ...
<!-- /ARGUS:CUSTOM -->
```

## CLAUDE.md

`CLAUDE.md` imports `AGENTS.md`, so Claude Code reads the same content as other tools. A new one looks like this:

```markdown
<!-- ARGUS:AUTO -->
@AGENTS.md
<!-- /ARGUS:AUTO -->

<!-- ARGUS:CUSTOM -->
<!-- Claude Code only: notes here are not shared with other tools. argus keeps this block when it regenerates. -->
<!-- /ARGUS:CUSTOM -->
```

An existing `CLAUDE.md` keeps all text outside the generated block. When its generated body moves to `AGENTS.md`, argus says so once.

## Other Formats

- `--format claude-code` writes `.claude/` agents, rules, skills and `settings.json` hooks. Its formatter hooks read the edited file with `jq` and do nothing without it. argus writes no MCP config.
- `--format cursor` writes `.cursor/rules/argus.mdc` with `alwaysApply: true` frontmatter; your edits to the frontmatter are kept. `.cursorrules` is no longer written.
- `--format copilot` writes `.github/copilot-instructions.md`.
- `--format continue` is deprecated and not part of `--format all`.

## Re-running argus

`argus scan`, `argus sync` and `argus watch` treat existing files in one of two ways.

**Merged files:** `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/argus.mdc` and `.github/copilot-instructions.md`

- A new file is written with `ARGUS:AUTO` markers around the generated content.
- With one `ARGUS:AUTO` pair, only the text between the markers is replaced. Text above, between and below (CUSTOM blocks or plain notes) stays where it is, and you may move the block.
- A file without markers, for example one you wrote yourself or one from an older argus version, gets the generated block on top and keeps its text below it. argus prints a one-time notice; remove anything an older argus run wrote that you no longer want.
- An `AGENTS.md` without markers counts as hand-written: it is left unchanged and argus warns. Add an empty `ARGUS:AUTO` pair where the generated block should go, or use `--force` to replace it.
- A symlinked `AGENTS.md` or `CLAUDE.md` is never written through.
- With unbalanced or repeated `ARGUS:AUTO` markers the file is left unchanged and argus warns. Fix the markers, or use `--force` to replace the file with a fresh generated one.
- `--merge=false` replaces the whole file with a fresh generated one.

**Create-only files:** everything under `.claude/` and `.continue/config.yaml`

- A missing file is written. A file that already matches the generated output is reported as up to date.
- A file that differs (for example one you edited) is left alone, and argus reports how many files it skipped (`-v` lists them). Run `argus scan --force` or `argus sync --force`, or delete the file, to regenerate it.

In both cases a file whose content would not change is not rewritten, and `--dry-run` shows the same decisions without writing anything.

## Auto-Generated Sections

The default `AGENTS.md` is lean: project description, Commands, Stack, Conventions, Git and API. `--full` writes the detailed sections below.

### Project Overview

Basic project information extracted from `package.json`, `go.mod`, `pyproject.toml`, etc.

### Quick Reference

Prioritized commands for common tasks:

```markdown
## Quick Reference

```bash
# Build
npm run build

# Test
npm test

# Lint
npm run lint

# Run
npm start
```
```

### Architecture

Project structure analysis:

```markdown
## Architecture

**Style:** Monorepo with Turborepo

**Entry Point:** `apps/web/src/main.tsx`

```
┌─────────────┐
│    apps     │
└──────┬──────┘
       │
┌──────┴──────┐
│  packages   │
└─────────────┘
```
```

### Tech Stack

Languages, frameworks, and tools:

```markdown
## Tech Stack

### Languages
- **TypeScript** 89.5%
- **CSS** 10.5%

### Frameworks
- React 18.2.0
- Next.js 14.0.0

### Tools
- ESLint
- Prettier
- GitHub Actions
```

### Detected Patterns

Code patterns found in the codebase:

```markdown
## Detected Patterns

### Data Fetching
- **React Query** - TanStack Query hooks (12 files)
- **fetch** - Native fetch API (8 files)

### State Management
- **useState** - React state (45 files)
- **Zustand** - Global state store (3 files)
```

## Custom Section

Any text outside the `ARGUS:AUTO` block persists across scans. A CUSTOM block (`argus scan --add-custom` adds an empty one) marks your notes explicitly:

```markdown
<!-- ARGUS:CUSTOM -->
## My Notes

- Remember to run `docker-compose up` before testing
- API keys are in 1Password vault "Dev"
- Contact @john for database access
<!-- /ARGUS:CUSTOM -->
```

## Output Indicators

argus uses these indicators in the output:

| Indicator | Meaning |
|-----------|---------|
| `✅` | Success |
| `⚠️` | Warning |
| `🔍` | Scanning/analyzing |
| `→` | Progress |
| `🔄` | Processing |
| `📊` | Analysis results |
| `📄` | File output |
| `👁️` | Watch mode |
| `❌` | Error |
