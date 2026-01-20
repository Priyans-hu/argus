package generator

import (
	"encoding/json"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// MCPConfig represents the Claude Code MCP configuration
type MCPConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents an individual MCP server configuration
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// generateMCP creates MCP configuration based on detected tech stack
func (g *ClaudeCodeGenerator) generateMCP(analysis *types.Analysis) []types.GeneratedFile {
	mcpConfig := MCPConfig{
		MCPServers: make(map[string]MCPServer),
	}

	// Detect and add relevant MCP servers
	g.addDatabaseMCPServers(analysis, mcpConfig.MCPServers)
	g.addToolMCPServers(analysis, mcpConfig.MCPServers)
	g.addFrameworkMCPServers(analysis, mcpConfig.MCPServers)

	// Only generate if we have servers to configure
	if len(mcpConfig.MCPServers) == 0 {
		return nil
	}

	// Marshal to JSON with indentation
	content, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return nil
	}

	return []types.GeneratedFile{
		{
			Path:    ".claude/mcp.json",
			Content: content,
		},
	}
}

// addDatabaseMCPServers adds database-specific MCP servers
func (g *ClaudeCodeGenerator) addDatabaseMCPServers(analysis *types.Analysis, servers map[string]MCPServer) {
	// Check databases in tech stack
	for _, db := range analysis.TechStack.Databases {
		dbLower := strings.ToLower(db)

		switch {
		case strings.Contains(dbLower, "postgres") || strings.Contains(dbLower, "postgresql"):
			servers["postgres"] = MCPServer{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-postgres"},
				Env: map[string]string{
					"POSTGRES_CONNECTION_STRING": "${POSTGRES_CONNECTION_STRING}",
				},
			}

		case strings.Contains(dbLower, "sqlite"):
			servers["sqlite"] = MCPServer{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-sqlite", "--db-path", "./database.db"},
			}

		case strings.Contains(dbLower, "mysql") || strings.Contains(dbLower, "mariadb"):
			servers["mysql"] = MCPServer{
				Command: "npx",
				Args:    []string{"-y", "@benborber/mcp-server-mysql"},
				Env: map[string]string{
					"MYSQL_HOST":     "${MYSQL_HOST:-localhost}",
					"MYSQL_USER":     "${MYSQL_USER}",
					"MYSQL_PASSWORD": "${MYSQL_PASSWORD}",
					"MYSQL_DATABASE": "${MYSQL_DATABASE}",
				},
			}

		case strings.Contains(dbLower, "mongo"):
			servers["mongodb"] = MCPServer{
				Command: "npx",
				Args:    []string{"-y", "mcp-mongo-server"},
				Env: map[string]string{
					"MONGODB_URI": "${MONGODB_URI}",
				},
			}

		case strings.Contains(dbLower, "redis"):
			servers["redis"] = MCPServer{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-redis"},
				Env: map[string]string{
					"REDIS_URL": "${REDIS_URL:-redis://localhost:6379}",
				},
			}
		}
	}

	// Also check frameworks for ORM-based database detection
	for _, fw := range analysis.TechStack.Frameworks {
		fwLower := strings.ToLower(fw.Name)

		switch {
		case strings.Contains(fwLower, "prisma"):
			// Prisma typically uses PostgreSQL or MySQL
			if _, exists := servers["postgres"]; !exists {
				servers["postgres"] = MCPServer{
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server-postgres"},
					Env: map[string]string{
						"POSTGRES_CONNECTION_STRING": "${DATABASE_URL}",
					},
				}
			}

		case strings.Contains(fwLower, "drizzle"):
			// Drizzle often uses PostgreSQL
			if _, exists := servers["postgres"]; !exists {
				servers["postgres"] = MCPServer{
					Command: "npx",
					Args:    []string{"-y", "@modelcontextprotocol/server-postgres"},
					Env: map[string]string{
						"POSTGRES_CONNECTION_STRING": "${DATABASE_URL}",
					},
				}
			}

		case strings.Contains(fwLower, "mongoose"):
			if _, exists := servers["mongodb"]; !exists {
				servers["mongodb"] = MCPServer{
					Command: "npx",
					Args:    []string{"-y", "mcp-mongo-server"},
					Env: map[string]string{
						"MONGODB_URI": "${MONGODB_URI}",
					},
				}
			}
		}
	}
}

// addToolMCPServers adds tool-specific MCP servers
func (g *ClaudeCodeGenerator) addToolMCPServers(analysis *types.Analysis, servers map[string]MCPServer) {
	// Check for Git (common in most projects)
	if isGitRepo(analysis) {
		servers["git"] = MCPServer{
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-git"},
		}
	}

	// Check for GitHub integration
	if hasGitHubIntegration(analysis) {
		servers["github"] = MCPServer{
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-github"},
			Env: map[string]string{
				"GITHUB_TOKEN": "${GITHUB_TOKEN}",
			},
		}
	}

	// Check for Docker
	if hasDocker(analysis) {
		servers["docker"] = MCPServer{
			Command: "npx",
			Args:    []string{"-y", "mcp-server-docker"},
		}
	}

	// Check for Kubernetes
	if hasKubernetes(analysis) {
		servers["kubernetes"] = MCPServer{
			Command: "npx",
			Args:    []string{"-y", "mcp-server-kubernetes"},
		}
	}

	// File system server (useful for most projects)
	servers["filesystem"] = MCPServer{
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
	}

	// Memory server for context persistence
	servers["memory"] = MCPServer{
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-memory"},
	}
}

// addFrameworkMCPServers adds framework-specific MCP servers
func (g *ClaudeCodeGenerator) addFrameworkMCPServers(analysis *types.Analysis, servers map[string]MCPServer) {
	for _, fw := range analysis.TechStack.Frameworks {
		fwLower := strings.ToLower(fw.Name)

		// Puppeteer/Playwright for browser automation projects
		if strings.Contains(fwLower, "puppeteer") || strings.Contains(fwLower, "playwright") {
			servers["puppeteer"] = MCPServer{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-puppeteer"},
			}
		}

		// Slack integration
		if strings.Contains(fwLower, "slack") {
			servers["slack"] = MCPServer{
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-slack"},
				Env: map[string]string{
					"SLACK_BOT_TOKEN": "${SLACK_BOT_TOKEN}",
					"SLACK_TEAM_ID":   "${SLACK_TEAM_ID}",
				},
			}
		}
	}

	// Check for web projects that might benefit from fetch
	if isWebProject(analysis) {
		servers["fetch"] = MCPServer{
			Command: "npx",
			Args:    []string{"-y", "@modelcontextprotocol/server-fetch"},
		}
	}
}

// Helper functions

func isGitRepo(analysis *types.Analysis) bool {
	// Check if git conventions were detected (indicates git repo)
	return analysis.GitConventions != nil
}

func hasGitHubIntegration(analysis *types.Analysis) bool {
	// Check for .github directory or GitHub-related files
	for _, dir := range analysis.Structure.Directories {
		if strings.Contains(dir.Path, ".github") {
			return true
		}
	}
	for _, file := range analysis.Structure.RootFiles {
		if strings.Contains(strings.ToLower(file), "github") {
			return true
		}
	}
	return false
}

func hasDocker(analysis *types.Analysis) bool {
	// Check for Dockerfile or docker-compose
	for _, file := range analysis.Structure.RootFiles {
		fileLower := strings.ToLower(file)
		if strings.Contains(fileLower, "dockerfile") || strings.Contains(fileLower, "docker-compose") {
			return true
		}
	}
	// Check for Docker in tools
	for _, tool := range analysis.TechStack.Tools {
		if strings.Contains(strings.ToLower(tool), "docker") {
			return true
		}
	}
	return false
}

func hasKubernetes(analysis *types.Analysis) bool {
	// Check for k8s manifests or helm charts
	for _, dir := range analysis.Structure.Directories {
		dirLower := strings.ToLower(dir.Path)
		if strings.Contains(dirLower, "kubernetes") ||
			strings.Contains(dirLower, "k8s") ||
			strings.Contains(dirLower, "helm") ||
			strings.Contains(dirLower, "charts") {
			return true
		}
	}
	return false
}

func isWebProject(analysis *types.Analysis) bool {
	// Check for web frameworks
	for _, fw := range analysis.TechStack.Frameworks {
		cat := strings.ToLower(fw.Category)
		if cat == "frontend" || cat == "backend" || cat == "fullstack" {
			return true
		}
	}
	// Check for web-related languages
	for _, lang := range analysis.TechStack.Languages {
		langLower := strings.ToLower(lang.Name)
		if langLower == "javascript" || langLower == "typescript" {
			return true
		}
	}
	return false
}
