<!-- ARGUS:AUTO -->
# argus

## Project Overview

The all-seeing code analyzer. Help AI grok your codebase.

## Tech Stack

### Languages

- **Go** 1.24 (100.0%)

### Frameworks & Libraries

**CLI:**
- Cobra

### Tools

- GitHub Actions

## Project Structure

```
.
├── cmd/          # Command entrypoints
├── internal/analyzer/          # Analysis logic
├── internal/config/          # Configuration
├── internal/detector/          # Detection logic
├── internal/generator/          # Code generation
├── internal/merger/          # Merge utilities
├── pkg/types/          # Type definitions
├── .goreleaser.yml
├── CHANGELOG.md
├── CLAUDE.md
├── CONTRIBUTING.md
├── LICENSE
├── README.md
├── SETUP.md
├── go.mod
└── llms.txt
```

## Key Files

| File | Purpose | Description |
|------|---------|-------------|
| `CONTRIBUTING.md` | Contributing | Contribution guidelines |
| `README.md` | Documentation | Project documentation |
| `cmd/argus/main.go` | Entry point | Go application entry |
| `go.mod` | Go module | Go dependencies |

## Coding Conventions

### Git

- Branch naming uses prefixes: feat, chore
  ```
  feat/user-auth, fix/login-bug, chore/update-deps
  ```

### Error-handling

- Go-style explicit error checking (if err != nil)
  ```
  if err != nil { return fmt.Errorf("context: %w", err) }
  ```

### Code-style

- Go project - use 'go fmt' or 'gofmt' for formatting

## Dependencies

### Runtime

- `github.com/fsnotify/fsnotify` v1.9.0
- `github.com/inconshreveable/mousetrap` v1.1.0
- `github.com/spf13/pflag` v1.0.9
- `golang.org/x/sys` v0.13.0
- `gopkg.in/yaml.v3` v3.0.1
<!-- /ARGUS:AUTO -->