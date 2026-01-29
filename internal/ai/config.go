package ai

import "time"

// Config controls the AI enrichment behavior
type Config struct {
	Enabled  bool          `yaml:"enabled"`
	Endpoint string        `yaml:"endpoint,omitempty"` // Ollama API endpoint (default: http://localhost:11434)
	Model    string        `yaml:"model,omitempty"`    // Model name (default: llama3.2)
	Timeout  time.Duration `yaml:"timeout,omitempty"`  // Request timeout (default: 120s)
}

// DefaultConfig returns an AI config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Enabled:  false,
		Endpoint: "http://localhost:11434",
		Model:    "llama3.2",
		Timeout:  120 * time.Second,
	}
}

// Merge fills in missing values from defaults
func (c *Config) Merge() {
	defaults := DefaultConfig()
	if c.Endpoint == "" {
		c.Endpoint = defaults.Endpoint
	}
	if c.Model == "" {
		c.Model = defaults.Model
	}
	if c.Timeout == 0 {
		c.Timeout = defaults.Timeout
	}
}
