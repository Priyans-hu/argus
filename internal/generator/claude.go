package generator

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// ClaudeGenerator generates CLAUDE.md files
type ClaudeGenerator struct{}

// NewClaudeGenerator creates a new Claude generator
func NewClaudeGenerator() *ClaudeGenerator {
	return &ClaudeGenerator{}
}

// Name returns the generator name
func (g *ClaudeGenerator) Name() string {
	return "claude"
}

// OutputFile returns the output filename
func (g *ClaudeGenerator) OutputFile() string {
	return "CLAUDE.md"
}

// Generate creates the CLAUDE.md content
// Produces a lean AUTO section: endpoints, key dependencies, and rule imports.
// Domain-specific context belongs in the CUSTOM section.
func (g *ClaudeGenerator) Generate(analysis *types.Analysis) ([]byte, error) {
	var buf bytes.Buffer

	// Header
	fmt.Fprintf(&buf, "# %s\n\n", analysis.ProjectName)

	// API Endpoints
	g.writeEndpoints(&buf, analysis.Endpoints)

	// Key Dependencies (one-liner)
	g.writeDependencies(&buf, analysis.Dependencies)

	// Import references to .claude/ rules
	g.writeImports(&buf, analysis)

	return buf.Bytes(), nil
}

// writeEndpoints writes the API endpoints section grouped by resource
func (g *ClaudeGenerator) writeEndpoints(buf *bytes.Buffer, endpoints []types.Endpoint) {
	if len(endpoints) == 0 {
		return
	}

	buf.WriteString("## API Endpoints\n\n")

	// Group endpoints by resource (first path segment after /)
	grouped := groupEndpointsByResource(endpoints)

	// Get sorted resource names
	var resources []string
	for resource := range grouped {
		resources = append(resources, resource)
	}
	sort.Strings(resources)

	totalShown := 0
	maxEndpoints := 100 // Show up to 100 endpoints total

	for _, resource := range resources {
		if totalShown >= maxEndpoints {
			break
		}

		eps := grouped[resource]
		if len(eps) == 0 {
			continue
		}

		// Write resource header
		displayResource := resource
		if displayResource == "" || displayResource == "/" {
			displayResource = "Root"
		}
		fmt.Fprintf(buf, "### %s\n\n", displayResource)
		buf.WriteString("| Method | Path | File |\n")
		buf.WriteString("|--------|------|------|\n")

		for _, ep := range eps {
			if totalShown >= maxEndpoints {
				break
			}

			file := ep.File
			if ep.Line > 0 {
				file = fmt.Sprintf("%s:%d", ep.File, ep.Line)
			}

			// Add auth indicator if present
			path := ep.Path
			if ep.Auth != "" {
				path = fmt.Sprintf("%s 🔒", ep.Path)
			}

			fmt.Fprintf(buf, "| %s | `%s` | `%s` |\n", ep.Method, path, file)
			totalShown++
		}

		buf.WriteString("\n")
	}

	remaining := len(endpoints) - totalShown
	if remaining > 0 {
		fmt.Fprintf(buf, "*...and %d more endpoints*\n\n", remaining)
	}
}

// groupEndpointsByResource groups endpoints by their resource path prefix
func groupEndpointsByResource(endpoints []types.Endpoint) map[string][]types.Endpoint {
	grouped := make(map[string][]types.Endpoint)

	for _, ep := range endpoints {
		resource := extractResourcePrefix(ep.Path)
		grouped[resource] = append(grouped[resource], ep)
	}

	// Sort endpoints within each group by path, then method
	for resource := range grouped {
		eps := grouped[resource]
		sort.Slice(eps, func(i, j int) bool {
			if eps[i].Path == eps[j].Path {
				return methodPriority(eps[i].Method) < methodPriority(eps[j].Method)
			}
			return eps[i].Path < eps[j].Path
		})
		grouped[resource] = eps
	}

	return grouped
}

// extractResourcePrefix extracts the resource name from a path
// /api/users/123 -> /api/users
// /users/:id -> /users
// /v1/products/categories -> /v1/products
func extractResourcePrefix(path string) string {
	// Clean the path
	path = strings.Trim(path, "/")
	if path == "" {
		return "/"
	}

	parts := strings.Split(path, "/")

	// Handle API versioning prefixes
	startIdx := 0
	if len(parts) > 0 {
		first := strings.ToLower(parts[0])
		// Check for common prefixes like api, v1, v2
		if first == "api" || (len(first) >= 2 && first[0] == 'v' && isDigit(first[1])) {
			startIdx = 1
		}
	}

	// Build resource path
	var resourceParts []string

	// Include prefix (api, v1, etc.)
	for i := 0; i < startIdx && i < len(parts); i++ {
		resourceParts = append(resourceParts, parts[i])
	}

	// Add the main resource (first non-prefix, non-param segment)
	for i := startIdx; i < len(parts) && i < startIdx+2; i++ {
		part := parts[i]
		// Skip dynamic segments
		if strings.HasPrefix(part, ":") || strings.HasPrefix(part, "{") ||
			strings.HasPrefix(part, "[") || part == "*" {
			break
		}
		resourceParts = append(resourceParts, part)
		// Stop at 2 meaningful segments
		if len(resourceParts)-startIdx >= 1 {
			break
		}
	}

	if len(resourceParts) == 0 {
		return "/"
	}

	return "/" + strings.Join(resourceParts, "/")
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func methodPriority(method string) int {
	priority := map[string]int{
		"GET": 1, "POST": 2, "PUT": 3, "PATCH": 4, "DELETE": 5, "ALL": 6,
	}
	if p, ok := priority[method]; ok {
		return p
	}
	return 99
}

// writeDependencies writes only notable dependencies with context
func (g *ClaudeGenerator) writeDependencies(buf *bytes.Buffer, deps []types.Dependency) {
	if len(deps) == 0 {
		return
	}

	// Only include notable dependencies that help understand the project
	// Skip generic/common packages that don't add context
	notablePatterns := map[string]string{
		// Databases
		"gorm":     "ORM",
		"sqlx":     "SQL",
		"mongo":    "MongoDB",
		"redis":    "Redis",
		"postgres": "PostgreSQL",
		"mysql":    "MySQL",
		"sqlite":   "SQLite",
		// Messaging
		"kafka":    "Kafka",
		"rabbitmq": "RabbitMQ",
		"nats":     "NATS",
		"pubsub":   "Pub/Sub",
		// Cloud
		"aws-sdk": "AWS",
		"azure":   "Azure",
		"gcloud":  "GCP",
		// Observability
		"prometheus":    "Metrics",
		"sentry":        "Error tracking",
		"opentelemetry": "Tracing",
		"elastic":       "APM",
		// Testing
		"testify":  "Testing",
		"gomock":   "Mocking",
		"httptest": "HTTP testing",
	}

	var notable []string
	seen := make(map[string]bool)

	for _, d := range deps {
		nameLower := strings.ToLower(d.Name)
		for pattern, category := range notablePatterns {
			if strings.Contains(nameLower, pattern) && !seen[category] {
				notable = append(notable, category)
				seen[category] = true
				break
			}
		}
	}

	// Only write section if we have notable dependencies
	if len(notable) > 0 {
		buf.WriteString("## Key Dependencies\n\n")
		sort.Strings(notable)
		buf.WriteString(strings.Join(notable, ", ") + "\n\n")
	}
}

// writeImports writes import references to .claude/ rules
// This uses Claude Code's @import syntax to reference external files
func (g *ClaudeGenerator) writeImports(buf *bytes.Buffer, analysis *types.Analysis) {
	// Build list of available rule imports
	var imports []string

	// Add standard rules that are typically generated
	if analysis.GitConventions != nil {
		imports = append(imports, "@.claude/rules/git-workflow.md")
	}
	if analysis.CodePatterns != nil && len(analysis.CodePatterns.Testing) > 0 {
		imports = append(imports, "@.claude/rules/testing.md")
	}
	if len(analysis.Conventions) > 0 {
		imports = append(imports, "@.claude/rules/coding-style.md")
	}
	if analysis.ArchitectureInfo != nil && analysis.ArchitectureInfo.Style != "" {
		imports = append(imports, "@.claude/rules/architecture.md")
	}
	// Security rules are always generated
	imports = append(imports, "@.claude/rules/security.md")

	if len(imports) == 0 {
		return
	}

	buf.WriteString("## Additional Rules\n\n")
	buf.WriteString("*The following rules are imported from `.claude/rules/` for context-specific guidance:*\n\n")

	for _, imp := range imports {
		fmt.Fprintf(buf, "- %s\n", imp)
	}
	buf.WriteString("\n")
}
