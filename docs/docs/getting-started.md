---
sidebar_position: 1
title: Getting Started
description: Quick start guide for argus
---

# Getting Started

**argus** is a CLI tool that generates comprehensive context files (`CLAUDE.md`) to help AI assistants understand your codebase.

## Quick Start

```bash
# Install via Homebrew
brew install Priyans-hu/tap/argus

# Or using Go
go install github.com/Priyans-hu/argus/cmd/argus@latest

# Scan your project
argus scan .
```

That's it! argus will analyze your codebase and generate a `CLAUDE.md` file with:

- Project overview and tech stack
- Architecture diagram and patterns
- Build, test, and run commands
- Coding conventions and practices
- Key files and their purposes

## What Problems Does argus Solve?

When working with AI coding assistants, you often need to provide context about your project:

- "We use React with TypeScript"
- "Run `npm test` to test"
- "Follow the existing naming conventions"

**argus automates this** by analyzing your codebase and generating a structured context file that AI assistants can read and understand.

## Example Output

After running `argus scan .`, you'll get a `CLAUDE.md` file like this:

```markdown
# my-project

## Tech Stack
- **Languages**: TypeScript (85%), CSS (15%)
- **Framework**: React 18, Next.js 14
- **Testing**: Jest, React Testing Library

## Quick Reference
```bash
npm run dev      # Start development server
npm run build    # Build for production
npm test         # Run tests
```

## Architecture
Standard Next.js App Router structure...

## Conventions
- Component files use PascalCase
- Tests colocated with source files
- ...
```

## Next Steps

- [Installation](/docs/installation) - Detailed installation options
- [Scan Command](/docs/usage/scan) - Learn about scan options
- [Configuration](/docs/configuration) - Customize argus behavior
