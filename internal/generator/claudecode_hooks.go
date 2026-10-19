package generator

import (
	"bytes"
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

	// Only generate if we have hooks to add
	if len(settings.Hooks) == 0 {
		return nil
	}

	content, err := marshalJSON(settings)
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
			formatCmd = formatHook("*.go", "gofmt -w")
		} else if hasLanguage(analysis, "TypeScript") || hasLanguage(analysis, "JavaScript") {
			if strings.Contains(ctx.FormatCommand, "prettier") {
				formatCmd = formatHook("*.ts|*.tsx|*.js|*.jsx|*.mjs|*.cjs", "npx prettier --write")
			}
		} else if hasLanguage(analysis, "Python") {
			if strings.Contains(ctx.FormatCommand, "black") {
				formatCmd = formatHook("*.py", "black")
			} else if strings.Contains(ctx.FormatCommand, "ruff") {
				formatCmd = formatHook("*.py", "ruff format")
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

// formatHook returns a PostToolUse command that runs formatter on the edited
// file when its name matches globs (a shell case pattern). Claude Code passes
// the path in the hook's stdin JSON; without jq or the formatter the hook
// exits quietly.
func formatHook(globs, formatter string) string {
	bin, _, _ := strings.Cut(formatter, " ")
	return "command -v jq >/dev/null 2>&1 && command -v " + bin + " >/dev/null 2>&1 || exit 0; " +
		`f=$(jq -r '.tool_input.file_path // empty'); case "$f" in ` + globs + ") " + formatter + ` "$f";; esac`
}

// marshalJSON indents v and ends it with a newline. Shell characters such as
// > and & stay readable instead of being escaped.
func marshalJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err := enc.Encode(v)
	return buf.Bytes(), err
}
