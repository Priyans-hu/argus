//go:build unix

package main

import (
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/Priyans-hu/argus/internal/merger"
)

// A write that fails halfway (here: past RLIMIT_FSIZE) must leave the
// existing file as it was and no temp file behind. The limit stays far above
// what the test binary itself writes meanwhile (its test log).
func TestMergeFailedTempWriteLeavesFileUntouched(t *testing.T) {
	resetFlags(t)
	captureOutput(t)
	root := t.TempDir()
	const stale = merger.AutoStartMarker + "\nstale\n" + merger.AutoEndMarker + "\n"
	writeFiles(t, root, map[string]string{"AGENTS.md": stale})

	var old syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_FSIZE, &old); err != nil {
		t.Skip(err)
	}
	limit := old
	limit.Cur = 1 << 20
	if err := syscall.Setrlimit(syscall.RLIMIT_FSIZE, &limit); err != nil {
		t.Skip(err)
	}
	t.Cleanup(func() { _ = syscall.Setrlimit(syscall.RLIMIT_FSIZE, &old) })

	err := writeMerged(root, "", "AGENTS.md", []byte(strings.Repeat("generated line\n", 1<<17)), false)
	if err == nil {
		t.Fatal("a write past the size limit succeeded")
	}
	if got := readFile(t, root, "AGENTS.md"); got != stale {
		t.Errorf("AGENTS.md changed after a failed write: %q", got)
	}
	if entries, _ := os.ReadDir(root); len(entries) != 1 {
		t.Errorf("directory holds %v, want only AGENTS.md", entries)
	}
}
