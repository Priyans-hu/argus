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

### Code-style

- Go project - use 'go fmt' or 'gofmt' for formatting

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

- **resty.** - Resty HTTP client
- **http.Get** - Go standard HTTP GET
- **http.Post** - Go standard HTTP POST
- **http.Client** - Go HTTP client

### Routing

- **fiber.Ctx** - Fiber framework context (2 files)
- **chi.Router** - Chi router
- **http.HandleFunc** - Go standard HTTP handler
- **echo.Context** - Echo framework context
- **mux.NewRouter** - Gorilla Mux router
- **fiber.New** - Fiber app initialization
- **gin.Context** - Gin framework context (2 files)
- **chi.NewRouter** - Chi router initialization
- **http.Handle** - Go standard HTTP handler
- **gin.Default** - Gin default router
- **echo.New** - Echo router initialization

### Testing

- **t.Error** - Go test error (6 files)
  - Found in: `cmd/argus/main.go`, `internal/analyzer/analyzer.go`, `internal/config/config.go`
- **httptest** - Go HTTP testing
- **t.Run** - Go subtest (2 files)
- **func Test** - Go test function (2 files)
- **t.Fatal** - Go test fatal (2 files)
- **require.** - Testify require assertions
- **assert.** - Testify assert
- **gomock** - GoMock mocking

### Authentication

- **jwt.** - JWT handling
- **middleware** - Auth middleware (4 files)
  - Found in: `internal/detector/codepatterns.go`, `internal/detector/endpoints.go`, `internal/detector/frameworks.go`
- **Authorization** - Authorization header (2 files)
- **Bearer** - Bearer token

### API Patterns

- **protobuf** - Protocol Buffers
- **websocket** - WebSocket communication
- **useQuery** - GraphQL/React Query (2 files)
- **swagger** - Swagger documentation
- **grpc** - gRPC protocol (2 files)
- **REST** - RESTful API design (3 files)
  - Found in: `internal/detector/codepatterns.go`, `internal/detector/frameworks.go`, `internal/detector/structure.go`
- **GraphQL** - GraphQL API (2 files)
- **gql`** - GraphQL query
- **socket.io** - Socket.IO real-time
- **useMutation** - GraphQL/React Query mutation
- **tRPC** - tRPC type-safe API
- **OpenAPI** - OpenAPI/Swagger spec

### Database & ORM

- **gorm.Model** - GORM model embedding
- **pgx** - pgx PostgreSQL driver
- **sqlx** - sqlx database library (2 files)
- **sql.DB** - Go standard SQL
- **ent.** - Ent ORM (3 files)
  - Found in: `cmd/argus/main.go`, `internal/detector/codepatterns.go`, `internal/generator/claude.go`
- **mongo.** - MongoDB Go driver
- **bun.** - Bun ORM
- **gorm.** - GORM ORM (2 files)

## Dependencies

### Runtime

- `github.com/fsnotify/fsnotify` v1.9.0
- `github.com/inconshreveable/mousetrap` v1.1.0
- `github.com/spf13/pflag` v1.0.9
- `golang.org/x/sys` v0.13.0
- `gopkg.in/yaml.v3` v3.0.1
<!-- /ARGUS:AUTO -->