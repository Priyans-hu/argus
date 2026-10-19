---
sidebar_position: 1
title: Getting Started
description: Quick start guide for argus
---

# Getting Started

**argus** is a CLI tool that generates context files (`AGENTS.md`, plus a `CLAUDE.md` that imports it) to help AI assistants understand your codebase.

## Quick Start

```bash
# Install via Homebrew
brew install Priyans-hu/tap/argus

# Or using Go
go install github.com/Priyans-hu/argus/cmd/argus@latest

# Scan your project
argus scan .
```

That's it! argus will analyze your codebase and generate an `AGENTS.md` file with:

- Project description and tech stack
- Build, test, and single-test commands
- Coding and git conventions

`argus scan --full` adds the architecture, project structure, key files and detected patterns.

## What Problems Does argus Solve?

When working with AI coding assistants, you often need to provide context about your project:

- "We use React with TypeScript"
- "Run `npm test` to test"
- "Follow the existing naming conventions"

**argus automates this** by analyzing your codebase and generating a structured context file that AI assistants can read and understand.

## Example Output

After running `argus scan .`, you'll get an `AGENTS.md` file like this:

<!-- verify -->
````markdown
# your-project

A Next.js app for booking meeting rooms.

## Commands

```bash
# Setup
npm install
npm ci

# Build
npm run build

# Test
npm test

# Lint
npm run lint

# Run
npm run dev
npm start
```

## Stack

- Languages: TypeScript 100%
- Frameworks: Prisma, Next.js, React, TypeScript, Vitest, ESLint

## Conventions

- Path aliases configured: @
- Absolute imports enabled with baseUrl: .
- TypeScript strict mode enabled - maintain strict type safety

## Git

- Repository: acme/your-project
- Commit style: Conventional Commits
- Commit format: `<type>(<scope>): <description>`
````

## Next Steps

- [Installation](/docs/installation) - Detailed installation options
- [Scan Command](/docs/usage/scan) - Learn about scan options
- [Configuration](/docs/configuration) - Customize argus behavior
