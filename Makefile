.PHONY: build test lint lint-fix setup-hooks clean

# Build the binary
build:
	go build -o bin/argus ./cmd/argus

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-v:
	go test -v ./...

# Run linter
lint:
	golangci-lint run ./...

# Run linter with auto-fix
lint-fix:
	gofmt -w .
	goimports -w . 2>/dev/null || true
	golangci-lint run --fix ./...

# Setup git hooks
setup-hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks configured to use .githooks/"

# Install development dependencies
setup:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(MAKE) setup-hooks
	@echo "Development environment ready!"

# Clean build artifacts
clean:
	rm -rf bin/

# Run the application
run:
	go run ./cmd/argus

# Format code
fmt:
	gofmt -w .
	goimports -w . 2>/dev/null || true
