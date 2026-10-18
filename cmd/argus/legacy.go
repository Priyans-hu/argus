package main

import (
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
	if format == "cursor" {
		notice("ℹ️", ".cursorrules", " is no longer written by argus and was left unchanged (rules now go to .cursor/rules/argus.mdc; delete the old file once its rules are moved, since a stale one can shadow AGENTS.md in tools like Zed)")
	}
}
