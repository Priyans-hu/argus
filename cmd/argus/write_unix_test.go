//go:build unix

package main

import (
	"os"
	"syscall"
	"testing"

	"github.com/Priyans-hu/argus/internal/config"
	"github.com/Priyans-hu/argus/internal/merger"
)

// A write that fails halfway (here: past RLIMIT_FSIZE) must leave the
// existing file as it was and no temp file behind.
func TestMergeFailedTempWriteLeavesFileUntouched(t *testing.T) {
	resetFlags(t)
	captureOutput(t)
	root := t.TempDir()
	const stale = merger.AutoStartMarker + "\nstale\n" + merger.AutoEndMarker + "\n"
	writeFiles(t, root, map[string]string{"CLAUDE.md": stale})

	var old syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &old); err != nil {
		t.Skip(err)
	}
	limit := old
	limit.Cur = 64
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit); err != nil {
		t.Skip(err)
	}
	t.Cleanup(func() { _ = syscall.Setrlimit(syscall.RLIMIT_FSIZE, &old) })

	err := generateOutput(root, "", "claude", goAnalysis(), config.DefaultConfig(), false, false)
	if err == nil {
		t.Fatal("a write past the size limit succeeded")
	}
	if got := readFile(t, root, "CLAUDE.md"); got != stale {
		t.Errorf("CLAUDE.md changed after a failed write: %q", got)
	}
	if entries, _ := os.ReadDir(root); len(entries) != 1 {
		t.Errorf("directory holds %v, want only CLAUDE.md", entries)
	}
}
