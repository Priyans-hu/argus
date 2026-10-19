package main

import (
	"strings"
	"testing"
)

func TestClaudeCodeLegacyNotices(t *testing.T) {
	resetFlags(t)
	_, errOut := captureOutput(t)
	root := t.TempDir()
	mcp := `{"mcpServers":{"memory":{"command":"npx","args":["-y","@modelcontextprotocol/server-memory"]}}}` + "\n"
	settings := `{"hooks":{"PostToolUse":[{"matcher":"Edit|Write","hooks":[{"type":"command","command":"gofmt -w $FILE_PATH"}]}]}}` + "\n"
	writeFiles(t, root, map[string]string{".claude/mcp.json": mcp, ".claude/settings.json": settings})
	mcpLine := "ℹ️  .claude/mcp.json is not read by Claude Code (project servers live in .mcp.json) and argus no longer writes it, " +
		"left unchanged (move the servers you use into .mcp.json, then delete it)\n"
	hookLine := "⚠️  .claude/settings.json: hook commands use $FILE_PATH, which Claude Code never sets, so they fail on every edit " +
		"(regenerate with --force, or read the path from stdin with jq as the Claude Code hooks guide shows)\n"

	for _, dry := range []bool{true, false, false} {
		dryRun = dry
		errOut.Reset()
		generate(t, root, "claude-code")
		if errOut.String() != mcpLine+hookLine {
			t.Errorf("dry-run %v: stderr =\n%q\nwant\n%q", dry, errOut, mcpLine+hookLine)
		}
	}
	if readFile(t, root, ".claude/mcp.json") != mcp || readFile(t, root, ".claude/settings.json") != settings {
		t.Error("legacy files were changed")
	}

	dryRun, force = false, true
	errOut.Reset()
	generate(t, root, "claude-code")
	if errOut.String() != mcpLine || strings.Contains(readFile(t, root, ".claude/settings.json"), "FILE_PATH") {
		t.Errorf("--force: stderr = %q, settings.json:\n%s", errOut, readFile(t, root, ".claude/settings.json"))
	}
	if readFile(t, root, ".claude/mcp.json") != mcp {
		t.Error("--force changed .claude/mcp.json")
	}
}
