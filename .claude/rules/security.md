# Security Rules for argus

Follow these security practices for all code changes.

## Authentication in This Project

Detected patterns:
- **Authorization** - see `internal/detector/codepatterns.go`
- **Bearer** - see `internal/detector/codepatterns.go`
- **jwt.** - see `internal/detector/codepatterns.go`
- **middleware** - see `internal/detector/codepatterns.go`

## General Security

- Never commit secrets, API keys, or credentials
- Validate all user input
- Use parameterized queries for database operations
- Sanitize output to prevent XSS
- Use HTTPS for all external communications
- Keep dependencies updated

## Authentication & Authorization

- Implement proper authentication for protected resources
- Use secure session management
- Implement rate limiting for authentication endpoints
- Use secure password hashing (bcrypt, argon2)
- Validate authorization for all protected actions

## Data Protection

- Encrypt sensitive data at rest
- Don't log sensitive information
- Use secure random number generation
- Implement proper error handling (don't leak details)

## Go Security

- Use `html/template` for HTML output (auto-escaping)
- Validate and sanitize all exec.Command inputs
- Use `crypto/rand` not `math/rand` for security
- Configure TLS properly (min TLS 1.2)

## Sensitive Files

Never commit these files:
- `.env` files with secrets
- Private keys (*.pem, *.key)
- Credentials files
- Database connection strings with passwords
