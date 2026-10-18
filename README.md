# Argus

[![CI](https://github.com/Priyans-hu/argus/actions/workflows/ci.yml/badge.svg)](https://github.com/Priyans-hu/argus/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Priyans-hu/argus)](https://goreportcard.com/report/github.com/Priyans-hu/argus)
[![Release](https://img.shields.io/github/v/release/Priyans-hu/argus)](https://github.com/Priyans-hu/argus/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Priyans-hu/argus)](https://go.dev/)
[![GitHub stars](https://img.shields.io/github/stars/Priyans-hu/argus?style=social)](https://github.com/Priyans-hu/argus)
[![Homebrew](https://img.shields.io/badge/homebrew-tap-orange)](https://github.com/Priyans-hu/homebrew-tap)

**The all-seeing code analyzer. Help AI grok your codebase.**

Argus scans your codebase and generates optimized context files for AI coding assistants — no more manually writing `AGENTS.md` or `CLAUDE.md`.

> If you find this useful, consider giving it a [⭐ star on GitHub](https://github.com/Priyans-hu/argus) — it helps others discover the project!

## The Problem

```
You: *opens Cursor/Claude Code*
AI: "I don't know your project structure, conventions, or patterns"
You: *spends 30 mins writing CLAUDE.md*
You: *file gets stale in 2 weeks*
You: *repeat*
```

## The Solution

```bash
argus scan

# Argus analyzes your codebase and generates:
# ✓ AGENTS.md (for Codex, Cursor, Copilot, Gemini CLI and other agents)
# ✓ CLAUDE.md (for Claude Code, imports AGENTS.md)
# and with --format: .claude/, .cursor/rules/argus.mdc, .github/copilot-instructions.md
```

## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/Priyans-hu/argus/main/install.sh | bash
```

### Go

```bash
go install github.com/Priyans-hu/argus/cmd/argus@latest
```

### Homebrew

```bash
brew install Priyans-hu/tap/argus
```

### Binary

Download from [Releases](https://github.com/Priyans-hu/argus/releases).

## Use as an Agent Skill

Let your coding agent run argus for you (Claude Code, Codex, Cursor, Copilot, Gemini CLI and more):

```bash
npx skills add Priyans-hu/argus --skill argus   # add -g for all projects
```

In Claude Code you can also run `/plugin install argus --marketplace Priyans-hu/argus` (older versions: `/plugin marketplace add Priyans-hu/argus`, then `/plugin install argus@argus`). The skill drives the CLI, so install argus as well.

## Quick Start

```bash
# Navigate to your project
cd your-project

# Scan and generate context files
argus scan

# Keep files synced with changes
argus sync
```

## Re-running is safe

Run `argus scan` or `argus sync` as often as you like:

- `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/argus.mdc` and `.github/copilot-instructions.md` hold the generated content between `<!-- ARGUS:AUTO -->` markers. A rescan replaces only that block; anything you write above or below it is kept.
- A file without markers (hand-written, or from an older argus) keeps its text below a new generated block, and argus tells you once. The exception is an `AGENTS.md` without markers: argus leaves it unchanged and warns; add an empty `ARGUS:AUTO` pair to opt in, or use `--force`.
- `.claude/**` and `.continue/config.yaml` are written once. If you edit them, later runs leave them alone and report them; `--force` overwrites them.
- A file whose content would not change is not rewritten.

## What Argus Detects

- **Tech Stack** — Frameworks, languages, databases
- **Project Structure** — Directory layout, key files
- **Conventions** — Naming patterns, code style, formatting
- **Dependencies** — Package managers, libraries
- **Commands** — Build, test, dev scripts
- **Patterns** — API shapes, error handling, state management

## Output Example
<!-- verify -->

````markdown
# your-project

## Commands

```bash
# Build
npm run build

# Test
npm test
npx vitest run <file>  # one test

# Run
npm run dev
```

## Stack

- Languages: TypeScript 94%, CSS 6%
- Frameworks: Next.js, React, Prisma
- Databases: PostgreSQL

## Conventions

- TypeScript strict mode is enabled
- Use the @/ path alias for imports
````

`argus scan --full` adds the project structure, key files, detected patterns and dependencies.

## Supported Output Formats

| Format | File | AI Tool |
|--------|------|---------|
| Agents | `AGENTS.md` | Codex, Cursor, Copilot, Gemini CLI and other agents |
| Claude | `CLAUDE.md` (imports `AGENTS.md`) | Claude Code |
| Claude Code | `.claude/` (agents, rules, skills, hooks) | Claude Code |
| Cursor | `.cursor/rules/argus.mdc` | Cursor |
| Copilot | `.github/copilot-instructions.md` | GitHub Copilot |
| Continue (deprecated) | `.continue/config.yaml` | Continue.dev |

`argus scan` writes `AGENTS.md` and `CLAUDE.md`. Choose formats with `--format` or `output:` in `.argus.yaml`; `--format all` writes every format except Continue. Claude Code only creates missing files, its formatter hooks need `jq`, and argus writes no MCP config.

## Commands

```bash
argus init      # Initialize config (optional)
argus scan      # Analyze and generate files
argus sync      # Update files with changes
argus scan --full    # Detailed AGENTS.md (structure, key files, patterns)
argus scan --force   # Also overwrite files argus would skip (edited .claude/ files, an AGENTS.md without markers)
argus --version # Print version (same as argus version)
```

`--compact` (lean output is now the default) and `-o`/`--output` (no effect) are deprecated.

`argus check` exits 1 when a generated file is out of date, for CI. <!-- verify -->

## Configuration (Optional)

Create `.argus.yaml` to customize:

```yaml
output:
  - agents
  - claude
  - cursor

ignore:
  - node_modules
  - .git
  - dist

custom_conventions:
  - "Use React Query for data fetching"
  - "All API routes return { success, data, error }"
```

## Why "Argus"?

In Greek mythology, **Argus Panoptes** was a giant with 100 eyes — the "all-seeing."

Argus sees your entire codebase so AI doesn't have to guess.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE).
