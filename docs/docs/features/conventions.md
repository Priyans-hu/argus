---
sidebar_position: 3
title: Convention Detection
description: How argus discovers coding conventions
---

# Convention Detection

argus analyzes your codebase to discover coding conventions and practices.

## Naming Conventions

### Files

| Convention | Example | Languages |
|------------|---------|-----------|
| kebab-case | `user-service.ts` | JavaScript, TypeScript |
| snake_case | `user_service.py` | Python |
| PascalCase | `UserService.go` | Go (exported) |
| camelCase | `userService.ts` | TypeScript |

### Variables & Functions

argus detects naming patterns:

```markdown
## Conventions

- Variables use camelCase
- Constants use SCREAMING_SNAKE_CASE
- React components use PascalCase
```

## Code Style

### Formatting

Detected from config files:

- ESLint configuration
- Prettier settings
- EditorConfig
- Go fmt

### Import Organization

argus detects import patterns:

```markdown
## Import Order

1. Standard library imports
2. Third-party imports
3. Local imports
```

## Testing Conventions

| Pattern | Meaning |
|---------|---------|
| `*_test.go` | Go tests colocated |
| `__tests__/` | Jest test directory |
| `*.spec.ts` | TypeScript specs |
| `test_*.py` | pytest tests |

## Documentation Conventions

argus detects documentation patterns:

- JSDoc comments
- Go doc comments
- Python docstrings
- TypeDoc annotations

## Git Conventions

From commit history:

```markdown
## Git Conventions

- Commit style: Conventional Commits
- Branch naming: `feat/`, `fix/`, `chore/`
- PR template exists
```

## Error Handling

argus identifies error handling patterns:

| Pattern | Language |
|---------|----------|
| `if err != nil` | Go |
| `try/catch` | JavaScript/TypeScript |
| `try/except` | Python |
| `Result<T, E>` | Rust |
