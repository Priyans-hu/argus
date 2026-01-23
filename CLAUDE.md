<!-- ARGUS:AUTO -->
# argus

## Project Overview

The all-seeing code analyzer. Help AI grok your codebase.

## Quick Reference

```bash
# Build
go build ./...           # Build all packages
make build               # Build the project

# Test
go test ./...            # Run all tests
go test -v ./...         # Run all tests with verbose output
make test                # Run tests
make test-v

# Lint
make lint                # Run linter
make lint-fix

# Format
go fmt ./...             # Format all Go files
make fmt                 # Format code

# Setup
make setup-hooks
make setup

```

## Architecture

**Style:** Standard Go Layout

**Entry Point:** `cmd/argus/main.go`

```
                    ┌─────────────┐
                    │    argus    │
                    └──────┬──────┘
                           │
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  analyzer   │  │   config    │  │  detector   │
└─────────────┘  └─────────────┘  └─────────────┘
       │              │
┌─────────────┐  ┌─────────────┐
│  generator  │  │   merger    │
└─────────────┘  └─────────────┘
                           │
                           ▼
                   [External Services]
```

**Package Dependencies:**
- `cmd` → `analyzer`, `config`, `generator`, `merger`
- `internal` → `config`, `detector`

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
├── .claude/
├── .continue/
├── cmd/          # Command entrypoints
├── internal/
│   ├── analyzer/          # Analysis logic
│   ├── config/          # Configuration
│   ├── detector/          # Detection logic
│   ├── generator/          # Code generation
│   └── merger/          # Merge utilities
├── pkg/
│   └── types/          # Type definitions
├── scripts/          # Scripts
├── .codecov.yml
├── .cursorrules
├── .goreleaser.yml
├── CHANGELOG.md
├── CLAUDE.md
├── CONTRIBUTING.md
├── LICENSE
├── Makefile
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

## Configuration

| File | Type | Purpose |
|------|------|--------|
| `Makefile` | Build | Make build automation |
| `.github/workflows/` | CI/CD | GitHub Actions workflows |
| `.codecov.yml` | Coverage | Codecov configuration |
| `.github/dependabot.yml` | Dependencies | Dependabot configuration |
| `go.mod` | Go Modules | Go dependencies and module path |
| `.goreleaser.yml` | Release | GoReleaser configuration |

## Development Setup

### Prerequisites

- Go 1.24+

### Initial Setup

```bash
make setup               # Run setup
```

### Git Hooks

- **pre-commit**: Format code, Organize imports, Run linter, Run tests

## Available Commands

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Run all tests with verbose output
go test -v ./...

# Format all Go files
go fmt ./...

make check-hooks

# Build the project
make build

# Run tests
make test

make test-v

# Run linter
make lint

make lint-fix

make setup-hooks

make setup

# Clean build artifacts
make clean

# Run the application
make run

# Format code
make fmt

```

## CLI Output & Verbosity

| Flag | Output |
|------|--------|
| (none) | Progress indicators, success/error messages |
| `-v, --verbose` | Detailed analysis results, file-by-file processing |
| `-n, --dry-run` | Preview output without writing files |

### Output Indicators

- `✅` - Success
- `⚠️` - Warning
- `🔍` - Scanning/analyzing
- `→` - Progress indicator
- `🔄` - Processing/syncing
- `📊` - Analysis results
- `📄` - File output
- `👁️` - Watch mode
- `❌` - Error

## Coding Conventions

### Testing

- Test files use _test suffix (Go style)
  ```
  handler_test.go, utils_test.go
  ```
- Tests are colocated with source files

### Code-style

- Go project - use 'go fmt' or 'gofmt' for formatting

### Documentation

- Go doc comments (start with function name)
  ```
  // HandleRequest processes incoming HTTP requests
  ```

### Error-handling

- Go-style explicit error checking (if err != nil)
  ```
  if err != nil { return fmt.Errorf("context: %w", err) }
  ```

### Git

- Repository: [Priyans-hu/argus](https://github.com/Priyans-hu/argus.git)
- Branch naming uses prefixes: feat, fix, chore
  ```
  feat/user-auth, fix/login-bug, chore/update-deps
  ```

## Guidelines

### Do

- Use `gofmt` or `goimports` for consistent formatting
- Handle all errors explicitly with `if err != nil`
- Use meaningful variable names; short names for short scopes
- Write doc comments for exported functions starting with function name
- Prefer composition over inheritance

### Don't

- Don't use `panic()` for regular error handling
- Don't ignore errors with `_`
- Don't use global state unnecessarily

## Detected Patterns

*The following patterns were detected by scanning the codebase:*

### Data Fetching

- **http.Client** - Go HTTP client
- **resty.** - Resty HTTP client
- **http.Get** - Go standard HTTP GET (4 files)
  - Found in: `cmd/argus/main.go`, `internal/detector/codepatterns.go`, `internal/detector/codepatterns_test.go`
- **http.Post** - Go standard HTTP POST (2 files)

### Routing

- **http.HandleFunc** - Go standard HTTP handler (2 files)
- **gin.Context** - Gin framework context (2 files)
- **fiber.New** - Fiber app initialization
- **mux.NewRouter** - Gorilla Mux router
- **gin.Default** - Gin default router
- **chi.NewRouter** - Chi router initialization
- **http.Handle** - Go standard HTTP handler (2 files)
- **fiber.Ctx** - Fiber framework context (2 files)
- **chi.Router** - Chi router
- **echo.Context** - Echo framework context
- **echo.New** - Echo router initialization

### Testing

- **t.Run** - Go subtest (8 files)
  - Found in: `internal/analyzer/incremental_test.go`, `internal/analyzer/parallel_test.go`, `internal/detector/architecture_test.go`
- **t.Error** - Go test assertions (13 files)
  - Found in: `internal/analyzer/incremental_test.go`, `internal/analyzer/parallel_test.go`, `internal/detector/architecture_test.go`
- **t.Fatal** - Go test fatal assertions (13 files)
  - Found in: `internal/analyzer/incremental_test.go`, `internal/analyzer/parallel_test.go`, `internal/detector/architecture_test.go`
- **func Test** - Go test function (13 files)
  - Found in: `internal/analyzer/incremental_test.go`, `internal/analyzer/parallel_test.go`, `internal/detector/architecture_test.go`
- **gomock** - GoMock mocking (2 files)
- **require** - Testify require assertions (2 files)
- **assert** - Testify assert (2 files)
- **httptest** - Go HTTP testing

### Authentication

- **Authorization** - Authorization header (4 files)
  - Found in: `internal/detector/codepatterns.go`, `internal/detector/structure.go`, `internal/generator/claudecode_agents.go`
- **Bearer** - Bearer token
- **jwt.** - JWT handling (2 files)
- **middleware** - Auth middleware (5 files)
  - Found in: `internal/detector/architecture.go`, `internal/detector/codepatterns.go`, `internal/detector/endpoints.go`

### API Patterns

- **REST** - RESTful API design (3 files)
  - Found in: `internal/detector/codepatterns.go`, `internal/detector/frameworks.go`, `internal/detector/structure.go`
- **useMutation** - GraphQL/React Query mutation
- **websocket** - WebSocket communication
- **socket.io** - Socket.IO real-time (2 files)
- **OpenAPI** - OpenAPI/Swagger spec
- **protobuf** - Protocol Buffers (2 files)
- **useQuery** - GraphQL/React Query (2 files)
- **swagger** - Swagger documentation
- **tRPC** - tRPC type-safe API (2 files)
- **GraphQL** - GraphQL API (3 files)
  - Found in: `internal/detector/ast_treesitter.go`, `internal/detector/codepatterns.go`, `internal/detector/structure.go`
- **gql`** - GraphQL query
- **grpc** - gRPC protocol (3 files)
  - Found in: `internal/detector/codepatterns.go`, `internal/detector/codepatterns_test.go`, `internal/detector/structure.go`

### Database & ORM

- **sql.Open** - Go standard SQL
- **pgx.Connect** - pgx PostgreSQL driver
- **mongo.Connect** - MongoDB Go driver
- **bun.NewDB** - Bun ORM
- **gorm.Open** - GORM ORM (2 files)
- **sqlx.Connect** - sqlx database library
- **sqlx.Open** - sqlx database library
- **gorm.Model** - GORM model embedding (2 files)

### Go Patterns

- **cobra.Command** - Cobra CLI framework (4 files)
  - Found in: `cmd/argus/main.go`, `internal/detector/cli_test.go`, `internal/detector/codepatterns.go`
- **urfave/cli** - urfave/cli framework (4 files)
  - Found in: `internal/detector/ast_go.go`, `internal/detector/cli.go`, `internal/detector/codepatterns.go`
- **spf13/pflag** - spf13 pflag for flags (2 files)
- **kingpin.** - Kingpin CLI framework
- **spf13/viper** - Viper config library (2 files)
- **viper.** - Viper configuration
- **godotenv** - GoDotEnv environment variables (2 files)
- **envconfig** - Kelseyhightower envconfig (2 files)
- **go func** - Goroutines (3 files)
  - Found in: `cmd/argus/main.go`, `internal/analyzer/parallel.go`, `internal/detector/codepatterns.go`
- **make(chan** - Channels (3 files)
  - Found in: `cmd/argus/main.go`, `internal/analyzer/parallel.go`, `internal/detector/codepatterns.go`
- **context.WithCancel** - Cancelable context (3 files)
  - Found in: `cmd/argus/main.go`, `internal/detector/ast_go.go`, `internal/detector/codepatterns.go`
- **context.Context** - Context for cancellation and deadlines (6 files)
  - Found in: `cmd/argus/main.go`, `internal/analyzer/analyzer.go`, `internal/analyzer/incremental.go`
- **sync.RWMutex** - Read-write mutex (2 files)
- **sync.WaitGroup** - WaitGroup for goroutine coordination (2 files)
- **sync.Mutex** - Mutex synchronization (2 files)
- **select {** - Select statement for channel operations (5 files)
  - Found in: `cmd/argus/main.go`, `internal/analyzer/analyzer.go`, `internal/analyzer/parallel.go`
- **context.WithTimeout** - Context with timeout (2 files)
- **sync.Once** - sync.Once for one-time initialization
- **slog.** - Go 1.21+ structured logging (4 files)
  - Found in: `cmd/argus/main.go`, `internal/analyzer/analyzer.go`, `internal/analyzer/parallel.go`
- **logrus.** - Logrus structured logging
- **zap.** - Zap high-performance logging
- **zerolog.** - Zerolog logging
- **fmt.Errorf** - Error formatting (8 files)
  - Found in: `cmd/argus/main.go`, `internal/analyzer/analyzer.go`, `internal/analyzer/parallel.go`
- **errors.Is** - Go 1.13+ error comparison (2 files)
- **errors.As** - Go 1.13+ error type assertion (2 files)
- **errors.Wrap** - pkg/errors error wrapping
- **wire.** - Google Wire dependency injection
- **dig.** - Uber Dig dependency injection
- **fx.** - Uber Fx application framework
- **yaml.Unmarshal** - YAML deserialization (2 files)
- **encoding/xml** - XML encoding (2 files)
- **json.Marshal** - JSON serialization (4 files)
  - Found in: `internal/detector/ast_go.go`, `internal/detector/codepatterns.go`, `internal/generator/claudecode_hooks.go`
- **protobuf** - Protocol Buffers (2 files)
- **json.Unmarshal** - JSON deserialization (8 files)
  - Found in: `internal/analyzer/analyzer.go`, `internal/detector/ast_go.go`, `internal/detector/codepatterns.go`
- **yaml.Marshal** - YAML serialization (2 files)

## Dependencies

### Runtime

- `github.com/fsnotify/fsnotify` v1.9.0
- `github.com/go-git/go-git/v5` v5.16.4
- `github.com/go-playground/validator/v10` v10.30.1
- `github.com/smacker/go-tree-sitter` v0.0.0-20240827094217-dd81d9e9be82
- `github.com/spf13/cobra` v1.10.2
- `golang.org/x/mod` v0.30.0
- `gopkg.in/yaml.v3` v3.0.1

## Additional Rules

*The following rules are imported from `.claude/rules/` for context-specific guidance:*

- @.claude/rules/git-workflow.md
- @.claude/rules/testing.md
- @.claude/rules/coding-style.md
- @.claude/rules/architecture.md
- @.claude/rules/security.md
<!-- /ARGUS:AUTO -->

<!-- ARGUS:CUSTOM -->
## My Custom Notes

This is a test custom section that should be preserved.
<!-- /ARGUS:CUSTOM -->