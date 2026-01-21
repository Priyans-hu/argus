<!-- ARGUS:AUTO -->
# argus

## Project Overview

The all-seeing code analyzer. Help AI grok your codebase.

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
- `internal` → `
		matches := importLineRegex.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) >= 3 {
				pkg := match[2]
				if pkg != `, `([^`, `config`, `detector`

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
├── internal/
│   ├── analyzer/          # Analysis logic
│   ├── config/          # Configuration
│   ├── detector/          # Detection logic
│   ├── generator/          # Code generation
│   └── merger/          # Merge utilities
├── pkg/
│   └── types/          # Type definitions
├── .codecov.yml
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

### Async

- Async/await pattern for asynchronous operations

### Git

- Commit style: **Conventional Commits**
  - Format: `<type>(<scope>): <description>`
  - Types: `feat`, `chore`, `fix`, `docs`
  - Scopes: `detector`, `generator`, `ci`, `analyzer`, `cli`
  - Example: `feat(detector): add new feature`
- Branch naming uses prefixes: feat, chore, fix
  ```
  feat/user-auth, chore/login-bug, fix/update-deps
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

- **http.Get** - Go standard HTTP GET (2 files)
- **http.Post** - Go standard HTTP POST
- **http.Client** - Go HTTP client
- **resty.** - Resty HTTP client

### Routing

- **mux.NewRouter** - Gorilla Mux router
- **gin.Context** - Gin framework context (2 files)
- **http.HandleFunc** - Go standard HTTP handler (2 files)
- **gin.Default** - Gin default router
- **echo.Context** - Echo framework context
- **echo.New** - Echo router initialization
- **fiber.Ctx** - Fiber framework context (2 files)
- **chi.NewRouter** - Chi router initialization
- **fiber.New** - Fiber app initialization
- **chi.Router** - Chi router
- **http.Handle** - Go standard HTTP handler (2 files)

### Testing

- **t.Error** - Go test assertions (6 files)
  - Found in: `internal/detector/architecture_test.go`, `internal/detector/codepatterns_test.go`, `internal/detector/endpoints_test.go`
- **t.Fatal** - Go test fatal assertions (6 files)
  - Found in: `internal/detector/architecture_test.go`, `internal/detector/codepatterns_test.go`, `internal/detector/endpoints_test.go`
- **func Test** - Go test function (6 files)
  - Found in: `internal/detector/architecture_test.go`, `internal/detector/codepatterns_test.go`, `internal/detector/endpoints_test.go`
- **t.Run** - Go subtest (4 files)
  - Found in: `internal/detector/architecture_test.go`, `internal/detector/codepatterns_test.go`, `internal/detector/git_test.go`
- **httptest** - Go HTTP testing
- **require** - Testify require assertions
- **assert** - Testify assert
- **gomock** - GoMock mocking

### Authentication

- **Authorization** - Authorization header (4 files)
  - Found in: `internal/detector/codepatterns.go`, `internal/detector/structure.go`, `internal/generator/claudecode_agents.go`
- **Bearer** - Bearer token
- **jwt.** - JWT handling (2 files)
- **middleware** - Auth middleware (4 files)
  - Found in: `internal/detector/codepatterns.go`, `internal/detector/endpoints.go`, `internal/detector/frameworks.go`

### API Patterns

- **OpenAPI** - OpenAPI/Swagger spec
- **swagger** - Swagger documentation
- **useMutation** - GraphQL/React Query mutation
- **tRPC** - tRPC type-safe API
- **websocket** - WebSocket communication
- **socket.io** - Socket.IO real-time
- **GraphQL** - GraphQL API (2 files)
- **protobuf** - Protocol Buffers (2 files)
- **REST** - RESTful API design (3 files)
  - Found in: `internal/detector/codepatterns.go`, `internal/detector/frameworks.go`, `internal/detector/structure.go`
- **gql`** - GraphQL query
- **useQuery** - GraphQL/React Query (2 files)
- **grpc** - gRPC protocol (3 files)
  - Found in: `internal/detector/codepatterns.go`, `internal/detector/codepatterns_test.go`, `internal/detector/structure.go`

### Database & ORM

- **sql.Open** - Go standard SQL
- **bun.NewDB** - Bun ORM
- **gorm.Open** - GORM ORM (2 files)
- **gorm.Model** - GORM model embedding (2 files)
- **sqlx.Connect** - sqlx database library
- **sqlx.Open** - sqlx database library
- **pgx.Connect** - pgx PostgreSQL driver
- **mongo.Connect** - MongoDB Go driver
<!-- /ARGUS:AUTO -->

<!-- ARGUS:CUSTOM -->
## My Custom Notes

This is a test custom section that should be preserved.
<!-- /ARGUS:CUSTOM -->