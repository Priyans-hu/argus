package generator

import (
	"github.com/Priyans-hu/argus/internal/config"
	"github.com/Priyans-hu/argus/pkg/types"
)

// ClaudeCodeGenerator generates Claude Code configuration files
// including agents, skills, and rules in the .claude/ directory
type ClaudeCodeGenerator struct {
	config *config.ClaudeCodeConfig
}

// NewClaudeCodeGenerator creates a new ClaudeCodeGenerator
func NewClaudeCodeGenerator(cfg *config.ClaudeCodeConfig) *ClaudeCodeGenerator {
	// Apply defaults if config is nil
	if cfg == nil {
		cfg = &config.ClaudeCodeConfig{
			Agents: true,
			Skills: true,
			Rules:  true,
			MCP:    true,
			Hooks:  true,
		}
	}
	return &ClaudeCodeGenerator{config: cfg}
}

// Name returns the generator name
func (g *ClaudeCodeGenerator) Name() string {
	return "claude-code"
}

// Generate creates all Claude Code configuration files
func (g *ClaudeCodeGenerator) Generate(analysis *types.Analysis) ([]types.GeneratedFile, error) {
	var files []types.GeneratedFile

	// Generate agents
	if g.config.Agents {
		agentFiles := g.generateAgents(analysis)
		files = append(files, agentFiles...)
	}

	// Generate skills (replaces commands in Claude Code)
	if g.config.Skills {
		skillFiles := g.generateSkills(analysis)
		files = append(files, skillFiles...)
	}

	// Generate rules
	if g.config.Rules {
		ruleFiles := g.generateRules(analysis)
		files = append(files, ruleFiles...)
	}

	// Generate MCP configuration
	if g.config.MCP {
		mcpFiles := g.generateMCP(analysis)
		files = append(files, mcpFiles...)
	}

	// Generate hooks in settings.json
	if g.config.Hooks {
		hookFiles := g.generateHooks(analysis)
		files = append(files, hookFiles...)
	}

	return files, nil
}

// hasLanguage checks if the analysis has a specific language
func hasLanguage(analysis *types.Analysis, lang string) bool {
	for _, l := range analysis.TechStack.Languages {
		if l.Name == lang {
			return true
		}
	}
	return false
}

// hasFramework checks if the analysis has a specific framework
func hasFramework(analysis *types.Analysis, framework string) bool {
	for _, f := range analysis.TechStack.Frameworks {
		if f.Name == framework {
			return true
		}
	}
	return false
}
