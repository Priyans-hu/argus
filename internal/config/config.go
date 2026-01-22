package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = ".argus.yaml"
)

// ClaudeCodeConfig controls what Claude Code configs to generate
type ClaudeCodeConfig struct {
	Agents bool `yaml:"agents"`
	Skills bool `yaml:"skills"` // Skills replace commands in Claude Code
	Rules  bool `yaml:"rules"`
	MCP    bool `yaml:"mcp"`
	Hooks  bool `yaml:"hooks"` // Generate .claude/settings.json with hooks
}

// Config represents Argus configuration
type Config struct {
	// Output formats to generate
	Output []string `yaml:"output,omitempty"`

	// Patterns to ignore (in addition to .gitignore)
	Ignore []string `yaml:"ignore,omitempty"`

	// Custom conventions to include in output
	CustomConventions []string `yaml:"custom_conventions,omitempty"`

	// Override detected values
	Overrides map[string]string `yaml:"overrides,omitempty"`

	// Claude Code specific configuration
	ClaudeCode *ClaudeCodeConfig `yaml:"claude_code,omitempty"`
}

// DefaultConfig returns a config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Output: []string{"claude"},
		Ignore: []string{
			"node_modules",
			".git",
			"dist",
			"build",
			"vendor",
			"*.log",
		},
		CustomConventions: []string{},
		Overrides:         map[string]string{},
	}
}

// Load reads config from .argus.yaml in the given directory
func Load(dir string) (*Config, error) {
	configPath := filepath.Join(dir, ConfigFileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply defaults for empty fields
	if len(cfg.Output) == 0 {
		cfg.Output = []string{"claude"}
	}

	return &cfg, nil
}

// Save writes config to .argus.yaml in the given directory
func Save(dir string, cfg *Config) error {
	configPath := filepath.Join(dir, ConfigFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add header comment
	header := []byte(`# Argus Configuration
# https://github.com/Priyans-hu/argus

`)
	content := append(header, data...)

	if err := os.WriteFile(configPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Exists checks if .argus.yaml exists in the given directory
func Exists(dir string) bool {
	configPath := filepath.Join(dir, ConfigFileName)
	_, err := os.Stat(configPath)
	return err == nil
}

// ConfigWithComments returns a commented config string for init
func ConfigWithComments() string {
	return `# Argus Configuration
# https://github.com/Priyans-hu/argus

# Output formats to generate
# Options: claude, claude-code, cursor, copilot, all
output:
  - claude
  # - claude-code  # Generate .claude/ directory with agents, commands, rules
  # - cursor
  # - copilot

# Additional patterns to ignore (beyond .gitignore)
ignore:
  - node_modules
  - .git
  - dist
  - build
  - vendor
  - "*.log"

# Custom conventions to include in generated files
# These are added to the auto-detected conventions
custom_conventions:
  # - "Use React Query for data fetching"
  # - "All API routes return { success, data, error }"
  # - "Components should be under 200 lines"

# Override auto-detected values
# overrides:
#   project_name: "My Project"
#   framework: "Next.js 14"

# Claude Code configuration (for --format claude-code)
# Controls which configs are generated in .claude/ directory
# claude_code:
#   agents: true    # Generate .claude/agents/*.md
#   skills: true    # Generate .claude/skills/*/SKILL.md (replaces commands)
#   rules: true     # Generate .claude/rules/*.md
#   mcp: true       # Generate .claude/mcp.json (MCP server configs)
#   hooks: true     # Generate .claude/settings.json with automation hooks
`
}
