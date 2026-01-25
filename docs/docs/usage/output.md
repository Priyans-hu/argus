---
sidebar_position: 3
title: Output Format
description: Understanding the CLAUDE.md output
---

# Output Format

argus generates a `CLAUDE.md` file with structured sections that AI assistants can easily parse and understand.

## File Structure

The output file has two main regions:

```markdown
<!-- ARGUS:AUTO -->
... auto-generated content ...
<!-- /ARGUS:AUTO -->

<!-- ARGUS:CUSTOM -->
... your custom notes (preserved on re-scan) ...
<!-- /ARGUS:CUSTOM -->
```

## Auto-Generated Sections

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

Add your own notes that persist across scans:

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
