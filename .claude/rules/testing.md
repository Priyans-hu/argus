---
paths:
  - "**/*_test.go"
---

# Testing Rules for argus

Follow these testing conventions for this project.

## Detected Testing Patterns

- **func Test**: Go test function (9 files)
  - See: `internal/detector/architecture_test.go`
- **t.Run**: Go subtest (5 files)
  - See: `internal/detector/architecture_test.go`
- **t.Error**: Go test assertions (9 files)
  - See: `internal/detector/architecture_test.go`
- **t.Fatal**: Go test fatal assertions (9 files)
  - See: `internal/detector/architecture_test.go`
- **require**: Testify require assertions (1 files)
  - See: `internal/detector/codepatterns.go`
- **assert**: Testify assert (1 files)
  - See: `internal/detector/codepatterns.go`
- **gomock**: GoMock mocking (1 files)
  - See: `internal/detector/codepatterns.go`
- **httptest**: Go HTTP testing (1 files)
  - See: `internal/detector/codepatterns.go`

## Test File Examples

Reference these files for test patterns:
- `internal/detector/architecture_test.go`
- `internal/detector/cli_test.go`
- `internal/detector/codepatterns_test.go`
- `internal/detector/codepatterns.go`

## Running Tests

```bash
go test ./...
```

## Testing Guidelines

### Go Testing

- Use table-driven tests for multiple cases
- Use `t.Run()` for subtests
- Name test functions as `TestFunctionName_Scenario`
- Use `t.Parallel()` for independent tests
- Mock external dependencies

## Best Practices

- Write tests before or alongside code
- Keep tests focused and independent
- Test edge cases and error conditions
- Don't test implementation details
- Maintain test coverage for critical paths
