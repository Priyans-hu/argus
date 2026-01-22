# Coding Style Rules for argus

Follow these coding conventions for this project.

## Commands

- Lint: `make lint`
- Format: `go fmt ./...`

## Code Style

- Go project - use 'go fmt' or 'gofmt' for formatting

## Error Handling

- Go-style explicit error checking (if err != nil)
  ```
  if err != nil { return fmt.Errorf("context: %w", err) }
  ```

## Go Style

- Run `gofmt` or `goimports` before committing
- Follow Effective Go guidelines
- Use meaningful variable names
- Keep functions focused and small

