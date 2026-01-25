---
sidebar_position: 10
title: Contributing
description: How to contribute to argus
---

# Contributing

We welcome contributions to argus! Here's how to get started.

## Development Setup

```bash
# Clone the repo
git clone https://github.com/Priyans-hu/argus.git
cd argus

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run linter
make lint
```

## Project Structure

```
argus/
├── cmd/argus/         # CLI entry point
├── internal/
│   ├── analyzer/      # Core analysis logic
│   ├── detector/      # Pattern detectors
│   ├── generator/     # Output generators
│   └── config/        # Configuration
├── pkg/types/         # Public types
└── docs/              # Documentation (Docusaurus)
```

## Adding a New Detector

1. Create a new file in `internal/detector/`
2. Implement the detector logic
3. Add tests in `*_test.go`
4. Integrate in `internal/analyzer/parallel.go`

Example detector:

```go
type MyDetector struct {
    rootPath string
    files    []types.FileInfo
}

func NewMyDetector(rootPath string, files []types.FileInfo) *MyDetector {
    return &MyDetector{rootPath: rootPath, files: files}
}

func (d *MyDetector) Detect() *MyInfo {
    // Detection logic
}
```

## Pull Request Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit with conventional commits: `git commit -m "feat: add feature"`
7. Push and create a PR

## Commit Convention

We use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `test:` - Tests
- `chore:` - Maintenance

## Code Style

- Run `go fmt` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Add tests for new functionality
- Document exported functions

## Reporting Issues

- Use GitHub Issues
- Include reproduction steps
- Share relevant `argus scan` output
- Mention your OS and Go version
