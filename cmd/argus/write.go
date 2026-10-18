package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/Priyans-hu/argus/internal/generator"
	"github.com/Priyans-hu/argus/internal/merger"
)

// Results go to stdout and notices to stderr, so `scan --dry-run > file`
// keeps only results. Tests swap both.
var (
	stdout io.Writer = os.Stdout
	stderr io.Writer = os.Stderr
)

func outf(format string, a ...any) { _, _ = fmt.Fprintf(stdout, format, a...) }
func errf(format string, a ...any) { _, _ = fmt.Fprintf(stderr, format, a...) }

// Reporting state for one run: skipped lists the create-only files left
// unchanged since the last reportSkipped; watchNotified is set only in watch
// mode and holds the paths already reported in the session.
var (
	skipped       []string
	watchNotified map[string]bool
)

// fileStatus is what a write helper did with one output file.
type fileStatus int

const (
	fileWritten   fileStatus = iota // written, or previewed in dry-run mode
	fileUnchanged                   // already equal to the generated output
	fileSkipped                     // differs from the generated output, left alone
)

// writePolicy is how an output that already exists is treated.
type writePolicy int

const (
	policyCreateOnly writePolicy = iota // written once, later differences skipped
	policyMerge                         // ARGUS:AUTO block replaced, other text kept
)

// policyFor returns the write policy for a slash path relative to its root.
func policyFor(rel string) writePolicy {
	if generator.IsMergeFile(rel) {
		return policyMerge
	}
	return policyCreateOnly
}

// writeMerged writes a merge-policy output: the generated text replaces the
// file's ARGUS:AUTO block and everything outside the block is kept.
func writeMerged(root, label, rel string, generated []byte, dryRun bool) error {
	target := filepath.Join(root, filepath.FromSlash(rel))
	shown := displayPath(label, rel)
	existing, err := os.ReadFile(target)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to read %s: %w", shown, err)
	}

	content, outcome := merger.NewMerger(mergeMode).Merge(existing, generated)
	if outcome == merger.Malformed {
		if !force {
			if notifyOnce(shown) {
				errf("⚠️  %s: unbalanced or repeated ARGUS:AUTO markers, left unchanged (fix them or use --force to overwrite)\n", shown)
			}
			return nil
		}
		content = []byte(merger.WrapContent(string(generated)))
	}
	if addCustomBlock && !bytes.Contains(content, []byte(merger.CustomStartMarker)) {
		content = []byte(merger.AddCustomSectionPlaceholder(string(content)))
	}
	content = append(bytes.TrimRight(content, "\r\n"), '\n')

	if sameText(existing, content) {
		reportFile(shown, fileUnchanged, dryRun)
		return nil
	}
	if err := writeOutput(target, shown, content, dryRun); err != nil {
		return err
	}
	reportFile(shown, fileWritten, dryRun)
	switch outcome {
	case merger.Adopted:
		errf("ℹ️  %s had no argus markers: generated block added on top, your text kept below it (remove any part an older argus run wrote)\n", shown)
	case merger.Malformed:
		errf("⚠️  %s: unbalanced or repeated ARGUS:AUTO markers, replaced with a fresh generated file (--force); the old text was not kept\n", shown)
	}
	return nil
}

// writeCreateOnly writes an output argus never merges into: a missing file
// is created, an equal one is left alone, and one that differs is skipped
// and reported unless --force is set.
func writeCreateOnly(root, label, rel string, content []byte, dryRun bool) (fileStatus, error) {
	target := filepath.Join(root, filepath.FromSlash(rel))
	shown := displayPath(label, rel)
	existing, err := os.ReadFile(target)
	switch {
	case errors.Is(err, fs.ErrNotExist):
	case err != nil:
		return 0, fmt.Errorf("failed to read %s: %w", shown, err)
	case sameText(existing, content):
		return fileUnchanged, nil
	case !force:
		skipped = append(skipped, shown)
		return fileSkipped, nil
	}
	return fileWritten, writeOutput(target, shown, content, dryRun)
}

// writeOutput previews content in dry-run mode and otherwise writes it with
// writeFile.
func writeOutput(target, shown string, content []byte, dryRun bool) error {
	if dryRun {
		outf("\n📄 Would write to %s:\n---\n%s\n---\n", target, content)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", shown, err)
	}
	if err := writeFile(target, content); err != nil {
		return fmt.Errorf("failed to write %s: %w", shown, err)
	}
	return nil
}

// writeFile replaces an existing regular file the owner can write through a
// temp file and a rename, so a failed write cannot truncate it. A missing
// file, a symlink (it stays one), a read-only file and a file whose temp copy
// cannot be created, chmodded or renamed are written in place.
func writeFile(target string, content []byte) error {
	fi, err := os.Lstat(target)
	if err != nil || !fi.Mode().IsRegular() || fi.Mode().Perm()&0200 == 0 {
		return os.WriteFile(target, content, 0644)
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".argus-*")
	if err != nil {
		return os.WriteFile(target, content, 0644)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	_, err = tmp.Write(content)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if os.Chmod(tmp.Name(), fi.Mode().Perm()) != nil || os.Rename(tmp.Name(), target) != nil {
		return os.WriteFile(target, content, 0644)
	}
	return nil
}

// reportFile prints the result line for a single-file output.
func reportFile(shown string, status fileStatus, dryRun bool) {
	switch {
	case status == fileUnchanged:
		outf("✅ %s is up to date\n", shown)
	case status == fileWritten && !dryRun:
		outf("✅ Generated %s\n", shown)
	}
}

// reportSkipped prints one line for the files skipped since the last call
// (-v lists them). In watch mode each path is reported once per session.
func reportSkipped() {
	var paths []string
	for _, p := range skipped {
		if notifyOnce(p) {
			paths = append(paths, p)
		}
	}
	skipped = nil
	if len(paths) == 0 {
		return
	}
	errf("⚠️  Skipped %d existing files that differ from the generated output (use --force to overwrite, -v to list)\n", len(paths))
	if verbose {
		for _, p := range paths {
			errf("  %s\n", p)
		}
	}
}

// notifyOnce reports whether a notice about path should be printed: always,
// except in watch mode for a path already reported this session.
func notifyOnce(shown string) bool {
	if watchNotified == nil {
		return true
	}
	if watchNotified[shown] {
		return false
	}
	watchNotified[shown] = true
	return true
}

// displayPath is the slash path of an output relative to the scanned
// directory; label is the workspace path of a monorepo workspace.
func displayPath(label, rel string) string {
	return path.Join(filepath.ToSlash(label), rel)
}

// sameText reports whether a and b are equal when CRLF is read as LF.
func sameText(a, b []byte) bool {
	crlf, lf := []byte("\r\n"), []byte("\n")
	return bytes.Equal(bytes.ReplaceAll(a, crlf, lf), bytes.ReplaceAll(b, crlf, lf))
}
