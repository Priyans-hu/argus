---
sidebar_position: 1
title: Tech Stack Detection
description: How argus identifies your tech stack
---

# Tech Stack Detection

argus automatically detects languages, frameworks, libraries, and tools in your project.

## Languages

Detected by analyzing file extensions and their prevalence:

- **Go** - `.go` files
- **JavaScript/TypeScript** - `.js`, `.jsx`, `.ts`, `.tsx`
- **Python** - `.py`
- **Rust** - `.rs`
- **Java** - `.java`
- **Ruby** - `.rb`
- And many more...

## Frameworks

### Web Frameworks

| Framework | Detection Method |
|-----------|------------------|
| React | `react` in dependencies |
| Vue | `vue` in dependencies |
| Angular | `@angular/core` in dependencies |
| Next.js | `next` in dependencies |
| Express | `express` in dependencies |
| FastAPI | `fastapi` in imports |
| Django | `django` in imports |
| Flask | `flask` in imports |
| Gin | `github.com/gin-gonic/gin` |
| Echo | `github.com/labstack/echo` |

### Testing Frameworks

| Framework | Detection Method |
|-----------|------------------|
| Jest | `jest` in dependencies |
| Vitest | `vitest` in dependencies |
| pytest | `pytest` in imports |
| Go testing | `testing` package imports |

## Libraries

argus identifies common libraries:

- State management (Redux, Zustand, MobX)
- Data fetching (React Query, SWR, Axios)
- UI components (Material-UI, Chakra, Tailwind)
- ORMs (Prisma, Drizzle, GORM, SQLAlchemy)
- And more...

## Tools

Detected from configuration files:

| Tool | Config File |
|------|-------------|
| ESLint | `.eslintrc.*` |
| Prettier | `.prettierrc` |
| Docker | `Dockerfile` |
| GitHub Actions | `.github/workflows/` |
| CircleCI | `.circleci/config.yml` |

## AST-Based Detection

For JavaScript/TypeScript and Python, argus uses AST parsing for accurate detection:

```go
// Detects actual imports, not just file patterns
import { useState } from 'react';  // ✓ React detected
const react = 'not a framework';   // ✗ Not detected
```

This avoids false positives from comments or string literals.
