# Contributing to Argus

Thanks for your interest in contributing to Argus!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone git@github.com:YOUR_USERNAME/argus.git`
3. Create a branch: `git checkout -b feat/your-feature`
4. Make your changes
5. Test: `go test ./...`
6. Commit: `git commit -m "feat: your feature"`
7. Push: `git push origin feat/your-feature`
8. Open a Pull Request

## Development Setup

```bash
# Clone
git clone git@github.com:Priyans-hu/argus.git
cd argus

# Install dependencies
go mod download

# Build
go build -o argus ./cmd/argus

# Run
./argus --help

# Test
go test ./...
```

## Project Structure

```
argus/
├── cmd/argus/          # CLI entry point
├── internal/
│   ├── analyzer/       # Codebase analysis
│   ├── detector/       # Tech stack detection
│   └── generator/      # Output generation
├── pkg/types/          # Shared types
└── docs/               # Documentation
```

## Commit Convention

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

Types: feat, fix, docs, style, refactor, test, chore
Scopes: analyzer, detector, generator, cli, config
```

Examples:
- `feat(detector): add Python support`
- `fix(generator): handle empty directories`
- `docs(readme): add installation section`

## Branch Naming

- `feat/` — New features
- `fix/` — Bug fixes
- `docs/` — Documentation
- `refactor/` — Code refactoring
- `test/` — Tests
- `chore/` — Maintenance

## Code Style

- Run `gofmt` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Keep functions focused and small
- Add comments for exported functions

## Adding a New Detector

1. Create file in `internal/detector/`
2. Implement the `Detector` interface
3. Register in detector registry
4. Add tests

```go
type Detector interface {
    Name() string
    Detect(ctx *AnalysisContext) (*DetectionResult, error)
}
```

## Adding a New Generator

1. Create file in `internal/generator/`
2. Implement the `Generator` interface
3. Register in generator registry
4. Add tests

```go
type Generator interface {
    Name() string
    OutputFile() string
    Generate(analysis *Analysis) ([]byte, error)
}
```

## Questions?

Open an issue or reach out to [@Priyans-hu](https://github.com/Priyans-hu).
