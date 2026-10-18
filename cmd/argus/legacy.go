package main

import (
	"bytes"
	"os"
	"path/filepath"
)

// legacyNotices reports files under root that older argus versions wrote for
// format and that argus now leaves unchanged. Each line is printed once per
// run, and once per session in watch mode.
func legacyNotices(root, label, format string) {
	notice := func(icon, rel, text string) {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			return
		}
		if shown := displayPath(label, rel); notifyOnce("legacy " + shown) {
			errf("%s  %s%s\n", icon, shown, text)
		}
	}
	switch format {
	case "cursor":
		notice("ℹ️", ".cursorrules", " is no longer written by argus and was left unchanged (rules now go to .cursor/rules/argus.mdc; delete the old file once its rules are moved, since a stale one can shadow AGENTS.md in tools like Zed)")
	case "claude-code":
		notice("ℹ️", ".claude/mcp.json", " is not read by Claude Code (project servers live in .mcp.json) and argus no longer writes it, left unchanged (move the servers you use into .mcp.json, then delete it)")
		settings, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
		if err == nil && !force && bytes.Contains(settings, []byte("$FILE_PATH")) {
			notice("⚠️", ".claude/settings.json", ": hook commands use $FILE_PATH, which Claude Code never sets, so they fail on every edit (regenerate with --force, or read the path from stdin with jq as the Claude Code hooks guide shows)")
		}
	}
}
