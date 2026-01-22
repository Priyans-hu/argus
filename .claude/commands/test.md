# Test - argus

Run the project test suite.

## Command

```bash
go test ./...
```

## Description

Run all tests

## Test Examples

Reference these files for test patterns:
- `internal/analyzer/incremental_test.go`
- `internal/analyzer/parallel_test.go`
- `internal/detector/architecture_test.go`

## On Failure

- Analyze the failing test output
- Identify the root cause of the failure
- Check if it's a test issue or a code issue
- Suggest fixes for the failing tests

## Options

- Add `-run TestName` to run specific test files
- Add `-v` for verbose output

## Success Criteria

- All tests pass
- No skipped tests without reason
- Coverage meets project standards (if applicable)
