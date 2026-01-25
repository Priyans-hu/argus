---
sidebar_position: 3
title: Configuration
description: Customize argus behavior
---

# Configuration

argus works out of the box without configuration, but you can customize its behavior.

## Configuration File

Create a `.argus.yaml` or `.argus.json` in your project root:

```yaml
# .argus.yaml
output: CLAUDE.md
verbose: false

# Directories to exclude from scanning
exclude:
  - node_modules
  - .git
  - dist
  - build
  - vendor

# Custom sections to preserve
preserve_custom: true
```

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `output` | string | `CLAUDE.md` | Output file path |
| `verbose` | boolean | `false` | Enable verbose logging |
| `exclude` | string[] | (built-in list) | Directories to exclude |
| `preserve_custom` | boolean | `true` | Keep custom section on re-scan |

## Default Exclusions

argus automatically excludes common non-source directories:

- `node_modules/`
- `.git/`
- `vendor/`
- `dist/`, `build/`, `out/`
- `__pycache__/`, `.venv/`, `venv/`
- `.idea/`, `.vscode/`
- `coverage/`

## Environment Variables

| Variable | Description |
|----------|-------------|
| `ARGUS_VERBOSE` | Enable verbose mode |
| `ARGUS_CONFIG` | Path to config file |

## Gitignore Integration

argus respects your `.gitignore` file. Any patterns in `.gitignore` are automatically excluded from scanning.

## Example Configurations

### Monorepo

```yaml
output: docs/AI_CONTEXT.md
exclude:
  - node_modules
  - .turbo
  - apps/*/dist
  - packages/*/dist
```

### Go Project

```yaml
output: CLAUDE.md
exclude:
  - vendor
  - bin
  - testdata
```

### Python Project

```yaml
output: AI_CONTEXT.md
exclude:
  - .venv
  - __pycache__
  - .pytest_cache
  - dist
  - "*.egg-info"
```
