package generator

import (
	"encoding/json"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// HookConfig represents a single hook configuration
type HookConfig struct {
	Type    string `json:"type"`
	Command string `json:"command,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
}

// HookMatcher represents a hook with its matcher pattern
type HookMatcher struct {
	Matcher string       `json:"matcher,omitempty"`
	Hooks   []HookConfig `json:"hooks"`
}

// SettingsJSON represents the .claude/settings.json structure
type SettingsJSON struct {
	Hooks       map[string][]HookMatcher `json:"hooks,omitempty"`
	Permissions *PermissionsConfig       `json:"permissions,omitempty"`
}

// PermissionsConfig represents permission settings
type PermissionsConfig struct {
	Allow []string `json:"allow,omitempty"`
	Deny  []string `json:"deny,omitempty"`
}

// generateHooks creates the .claude/settings.json file with useful hooks
func (g *ClaudeCodeGenerator) generateHooks(analysis *types.Analysis) []types.GeneratedFile {
	ctx := BuildContext(analysis)

	settings := SettingsJSON{
		Hooks: make(map[string][]HookMatcher),
	}

	// Add PostToolUse hooks for auto-linting after file edits
	postToolUseHooks := g.buildPostToolUseHooks(analysis, ctx)
	if len(postToolUseHooks) > 0 {
		settings.Hooks["PostToolUse"] = postToolUseHooks
	}

	// Add PreToolUse hooks for validation
	preToolUseHooks := g.buildPreToolUseHooks(analysis, ctx)
	if len(preToolUseHooks) > 0 {
		settings.Hooks["PreToolUse"] = preToolUseHooks
	}

	// Only generate if we have hooks to add
	if len(settings.Hooks) == 0 {
		return nil
	}

	// Marshal to JSON with indentation
	content, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return nil
	}

	return []types.GeneratedFile{
		{
			Path:    ".claude/settings.json",
			Content: content,
		},
	}
}

// buildPostToolUseHooks creates hooks that run after tool execution
func (g *ClaudeCodeGenerator) buildPostToolUseHooks(analysis *types.Analysis, ctx *GeneratorContext) []HookMatcher {
	var hooks []HookMatcher

	// Auto-format after file edits
	if ctx.FormatCommand != "" {
		// Determine which file patterns to format based on language
		var formatCmd string

		if hasLanguage(analysis, "Go") {
			formatCmd = "gofmt -w $FILE_PATH"
		} else if hasLanguage(analysis, "TypeScript") || hasLanguage(analysis, "JavaScript") {
			if strings.Contains(ctx.FormatCommand, "prettier") {
				formatCmd = "npx prettier --write $FILE_PATH"
			} else if strings.Contains(ctx.FormatCommand, "eslint") {
				formatCmd = "npx eslint --fix $FILE_PATH"
			}
		} else if hasLanguage(analysis, "Python") {
			if strings.Contains(ctx.FormatCommand, "black") {
				formatCmd = "black $FILE_PATH"
			} else if strings.Contains(ctx.FormatCommand, "ruff") {
				formatCmd = "ruff format $FILE_PATH"
			}
		}

		if formatCmd != "" {
			hooks = append(hooks, HookMatcher{
				Matcher: "Edit|Write",
				Hooks: []HookConfig{
					{
						Type:    "command",
						Command: formatCmd,
						Timeout: 30,
					},
				},
			})
		}
	}

	return hooks
}

// buildPreToolUseHooks creates hooks that run before tool execution
func (g *ClaudeCodeGenerator) buildPreToolUseHooks(analysis *types.Analysis, ctx *GeneratorContext) []HookMatcher {
	var hooks []HookMatcher

	// Add security validation for bash commands if security patterns detected
	if len(ctx.AuthPatterns) > 0 || hasSecurityPatterns(analysis) {
		hooks = append(hooks, HookMatcher{
			Matcher: "Bash",
			Hooks: []HookConfig{
				{
					Type:    "command",
					Command: buildSecurityValidationScript(),
					Timeout: 10,
				},
			},
		})
	}

	return hooks
}

// hasSecurityPatterns checks if the project has security-related patterns
func hasSecurityPatterns(analysis *types.Analysis) bool {
	if analysis.CodePatterns == nil {
		return false
	}

	// Check for auth patterns
	for _, pattern := range analysis.CodePatterns.Authentication {
		if pattern.FileCount > 0 {
			return true
		}
	}

	return false
}

// buildSecurityValidationScript returns an inline validation script
// that checks for potentially dangerous bash commands
func buildSecurityValidationScript() string {
	// This is a simple inline script that validates bash commands
	// In a real project, this would be a separate script file
	return `python3 -c "
import json
import sys
import re

DANGEROUS_PATTERNS = [
    r'rm\s+-rf\s+/',
    r'chmod\s+777',
    r'curl.*\|.*sh',
    r'wget.*\|.*sh',
    r'eval\s+',
]

try:
    data = json.load(sys.stdin)
    cmd = data.get('tool_input', {}).get('command', '')
    for pattern in DANGEROUS_PATTERNS:
        if re.search(pattern, cmd, re.IGNORECASE):
            print(f'Potentially dangerous command detected: {pattern}', file=sys.stderr)
            sys.exit(2)
    sys.exit(0)
except Exception as e:
    sys.exit(0)
"`
}
