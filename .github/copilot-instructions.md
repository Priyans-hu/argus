# argus - Copilot Instructions

## Overview

This is a CLI project primarily written in Go.

## Tech Stack

**Languages:**
- Go 1.24

**Frameworks/Libraries:**
- Cobra

## Architecture

**Directory Structure:**
- `cmd/` - Command entrypoints
- `internal/analyzer/` - Analysis logic
- `internal/config/` - Configuration
- `internal/detector/` - Detection logic
- `internal/generator/` - Code generation
- `internal/merger/` - Merge utilities
- `pkg/types/` - Type definitions
- `scripts/` - Scripts

## Coding Standards

- Go project - use 'go fmt' or 'gofmt' for formatting
- Use `gofmt` for formatting
- Always handle errors explicitly
- Export only what needs to be public

## Patterns to Follow

**Error handling:**
- Go-style explicit error checking (if err != nil)

**Testing:**
- Test files use _test suffix (Go style)
- Tests are colocated with source files

## Avoid

- Don't add unnecessary comments for obvious code
- Don't ignore errors or use empty catch blocks
- Don't commit sensitive data or credentials
- Don't introduce breaking changes without discussion
- Don't use panic() for regular error handling
- Don't use global state unnecessarily

