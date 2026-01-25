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
3. Re-analyzes and updates `CLAUDE.md` when changes are detected
4. Uses debouncing to avoid excessive regeneration

## Options

| Flag | Short | Description |
|------|-------|-------------|
| `--verbose` | `-v` | Show detailed output |
| `--output` | `-o` | Custom output file path |

## Example

```bash
$ argus watch .
👁️ Watching /my-project for changes...
🔍 Initial scan complete
📄 Generated CLAUDE.md

# Make some changes to your code...

🔄 Changes detected, re-scanning...
✅ Updated CLAUDE.md
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
