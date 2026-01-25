---
sidebar_position: 2
title: Architecture Analysis
description: How argus understands project structure
---

# Architecture Analysis

argus analyzes your project structure to understand its architecture.

## Architecture Styles

argus detects common architectural patterns:

| Style | Detection |
|-------|-----------|
| Standard Go Layout | `cmd/`, `internal/`, `pkg/` directories |
| Next.js App Router | `app/` directory with page files |
| Monorepo | `packages/` or `apps/` directories |
| Microservices | Multiple service directories |
| MVC | `models/`, `views/`, `controllers/` |
| Clean Architecture | `domain/`, `usecase/`, `infrastructure/` |

## Package Dependencies

argus analyzes import statements to understand dependencies:

```
┌─────────────┐
│     cmd     │
└──────┬──────┘
       │
┌──────┴──────┐
│   internal  │
└──────┬──────┘
       │
┌──────┴──────┐
│     pkg     │
└─────────────┘
```

## Entry Points

Detected entry points:

| Type | Detection |
|------|-----------|
| Go | `main.go` or `cmd/*/main.go` |
| Node.js | `main` field in `package.json` |
| Python | `__main__.py` or `main.py` |

## Layer Analysis

For layered architectures, argus identifies:

- **Presentation layer** - API handlers, controllers
- **Business layer** - Services, use cases
- **Data layer** - Repositories, models

## Monorepo Detection

argus detects monorepo tools:

- Turborepo (`turbo.json`)
- Nx (`nx.json`)
- Lerna (`lerna.json`)
- pnpm workspaces (`pnpm-workspace.yaml`)
- npm/yarn workspaces (`workspaces` in `package.json`)

For monorepos, argus lists all packages/apps:

```markdown
## Monorepo Structure

**Tool:** Turborepo

**Packages:**
- `apps/web` - Next.js frontend
- `apps/api` - Express backend
- `packages/ui` - Shared components
- `packages/config` - Shared configuration
```
