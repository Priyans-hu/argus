package generator

import "github.com/Priyans-hu/argus/pkg/types"

// AgentsGenerator writes AGENTS.md, the main output every tool reads.
type AgentsGenerator struct {
	full bool
}

// NewAgentsGenerator returns a generator for the lean body, or for the
// detailed one when full is set.
func NewAgentsGenerator(full bool) *AgentsGenerator {
	return &AgentsGenerator{full: full}
}

// OutputFile returns the output filename
func (g *AgentsGenerator) OutputFile() string {
	return "AGENTS.md"
}

// Generate renders the AGENTS.md body.
func (g *AgentsGenerator) Generate(analysis *types.Analysis) ([]byte, error) {
	if g.full {
		return NewClaudeGenerator().Generate(analysis)
	}
	return renderLean(analysis), nil
}

// ShimGenerator writes CLAUDE.md as an import of AGENTS.md, so Claude Code
// reads the same body as every other tool.
type ShimGenerator struct{}

// OutputFile returns the output filename
func (ShimGenerator) OutputFile() string {
	return "CLAUDE.md"
}

// Generate returns the import line.
func (ShimGenerator) Generate(*types.Analysis) ([]byte, error) {
	return []byte("@AGENTS.md\n"), nil
}
