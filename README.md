# Argus

[![CI](https://github.com/Priyans-hu/argus/actions/workflows/ci.yml/badge.svg)](https://github.com/Priyans-hu/argus/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Priyans-hu/argus)](https://goreportcard.com/report/github.com/Priyans-hu/argus)
[![Release](https://img.shields.io/github/v/release/Priyans-hu/argus)](https://github.com/Priyans-hu/argus/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Priyans-hu/argus)](https://go.dev/)
[![GitHub stars](https://img.shields.io/github/stars/Priyans-hu/argus?style=social)](https://github.com/Priyans-hu/argus)
[![Homebrew](https://img.shields.io/badge/homebrew-tap-orange)](https://github.com/Priyans-hu/homebrew-tap)

**The all-seeing code analyzer. Help AI grok your codebase.**

Argus scans your codebase and generates optimized context files for AI coding assistants — no more manually writing `CLAUDE.md` or `.cursorrules`.

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
# ✓ CLAUDE.md (for Claude Code)
# ✓ .cursorrules (for Cursor)
# ✓ .github/copilot-instructions.md (for Copilot)
# ✓ .continue/config.json (for Continue)
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

## Quick Start

```bash
# Navigate to your project
cd your-project

# Scan and generate context files
argus scan

# Keep files synced with changes
argus sync
```

## What Argus Detects

- **Tech Stack** — Frameworks, languages, databases
- **Project Structure** — Directory layout, key files
- **Conventions** — Naming patterns, code style, formatting
- **Dependencies** — Package managers, libraries
- **Commands** — Build, test, dev scripts
- **Patterns** — API shapes, error handling, state management

## Output Example

```markdown
# your-project

## Tech Stack
- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript (strict mode)
- **Database**: PostgreSQL + Prisma ORM
- **Styling**: TailwindCSS + shadcn/ui

## Project Structure
src/
├── app/           # Pages (App Router)
├── components/    # React components
├── lib/           # Utilities
└── types/         # TypeScript definitions

## Conventions
- Components: PascalCase (UserCard.tsx)
- Utilities: camelCase (formatDate.ts)
- Use @/ path alias for imports

## Commands
- npm run dev — Start dev server
- npm run build — Production build
- npm test — Run tests
```

## Supported Output Formats

| Format | File | AI Tool |
|--------|------|---------|
| Claude | `CLAUDE.md` | Claude Code |
| Cursor | `.cursorrules` | Cursor |
| Copilot | `.github/copilot-instructions.md` | GitHub Copilot |
| Continue | `.continue/config.json` | Continue.dev |

## Commands

```bash
argus init      # Initialize config (optional)
argus scan      # Analyze and generate files
argus sync      # Update files with changes
argus version   # Print version
```

## Configuration (Optional)

Create `.argus.yaml` to customize:

```yaml
output:
  - claude
  - cursor
  - copilot

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
