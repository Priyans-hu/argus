---
sidebar_position: 2
title: Watch Command
description: Automatically update context when files change
---

# Watch Command

The `watch` command monitors your project for changes and automatically regenerates the context file.

## Basic Usage

```bash
argus watch .
```

## How It Works

1. Performs an initial scan
2. Watches for file changes using filesystem events
3. Re-analyzes and updates the generated files when changes are detected
4. Uses debouncing to avoid excessive regeneration

## Options

| Flag | Short | Description |
|------|-------|-------------|
| `--verbose` | `-v` | Show detailed output |
| `--full` | | Write the detailed sections |

## Example

```bash
$ argus watch .
👁️ Watching /my-project for changes...
🔍 Initial scan complete
📄 Generated AGENTS.md

# Make some changes to your code...

🔄 Changes detected, re-scanning...
✅ Updated AGENTS.md
```

## When to Use Watch

- During active development when project structure changes frequently
- When onboarding to understand how the codebase evolves
- When experimenting with different architectures

## Stopping Watch

Press `Ctrl+C` to stop watching.

```bash
^C
👁️ Stopped watching
```
