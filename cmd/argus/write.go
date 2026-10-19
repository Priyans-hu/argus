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
	"strings"

	"github.com/Priyans-hu/argus/internal/generator"
	"github.com/Priyans-hu/argus/internal/merger"
)

// The pair of context files: AGENTS.md holds the body, CLAUDE.md imports it.
const (
	agentsFile = "AGENTS.md"
	claudeFile = "CLAUDE.md"
	shimImport = "@AGENTS.md"
)

// shimNotes is the Claude-only CUSTOM block a new CLAUDE.md shim gets below
// its import of AGENTS.md.
const shimNotes = merger.CustomStartMarker + "\n<!-- Claude Code only: notes here are not shared with other tools. argus keeps this block when it regenerates. -->\n" + merger.CustomEndMarker

// tempPattern names the temp file writeFile renames over its target; scans
// and watch skip what it leaves behind (generator.IsGeneratedPath).
const tempPattern = generator.TempPrefix + "*"

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

// checked is set only by check: it maps each output's display path to its
// state, and writeOutput records instead of writing.
var checked map[string]string

// record notes an output's state for check.
func record(shown, state string) {
	if checked != nil {
		checked[shown] = state
	}
}

// fileStatus is what a write helper did with one output file.
type fileStatus int

const (
	fileWritten   fileStatus = iota // written, or previewed in dry-run mode
	fileUnchanged                   // already equal to the generated output
	fileSkipped                     // differs from the generated output, left alone
)

// writeMerged writes a merge file (see generator.IsMergeFile): the generated
// text replaces the file's ARGUS:AUTO block and everything outside the block
// is kept.
func writeMerged(root, label, rel string, generated []byte, dryRun bool) error {
	target := filepath.Join(root, filepath.FromSlash(rel))
	shown := displayPath(label, rel)
	if skipLinked(root, shown, rel, dryRun) {
		return nil
	}
	existing, err := os.ReadFile(target)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("failed to read %s: %w", shown, err)
	}

	unmarked := !isBlank(existing) && !bytes.Contains(existing, []byte(merger.AutoStartMarker)) &&
		!bytes.Contains(existing, []byte(merger.AutoEndMarker))
	switch {
	case rel == agentsFile && unmarked && !force:
		record(shown, "skipped")
		if notifyOnce(shown) {
			errf("⚠️  %s exists without argus markers, left unchanged (add an empty ARGUS:AUTO pair where the generated block should go, or use --force to overwrite)\n", shown)
		}
		return nil
	case rel == claudeFile && unmarked && hasLine(existing, shimImport):
		reportFile(shown, fileUnchanged, dryRun)
		return nil
	}
	replaced := rel == agentsFile && unmarked // only with --force

	front, body, generated := splitFront(existing, generated)
	content, outcome := merger.NewMerger(mergeMode).Merge(body, generated)
	if replaced {
		content = []byte(merger.WrapContent(string(generated)))
	}
	if outcome == merger.Malformed {
		if !force {
			record(shown, "skipped")
			if notifyOnce(shown) {
				errf("⚠️  %s: unbalanced or repeated ARGUS:AUTO markers, left unchanged (fix them or use --force to overwrite)\n", shown)
			}
			return nil
		}
		content = []byte(merger.WrapContent(string(generated)))
	}
	if rel == claudeFile && isBlank(existing) {
		content = append(content, "\n\n"+shimNotes...)
	}
	if addCustomBlock && !bytes.Contains(content, []byte(merger.CustomStartMarker)) {
		content = []byte(merger.AddCustomSectionPlaceholder(string(content)))
	}
	content = bytes.Join([][]byte{front, content}, nil)
	content = append(bytes.TrimRight(content, "\r\n"), '\n')
	// A file whose newlines are all CRLF stays CRLF; others keep an LF block.
	crlf, lf := []byte("\r\n"), []byte("\n")
	if n := bytes.Count(existing, lf); n > 0 && n == bytes.Count(existing, crlf) {
		content = bytes.ReplaceAll(bytes.ReplaceAll(content, crlf, lf), lf, crlf)
	}

	if sameText(existing, content) {
		reportFile(shown, fileUnchanged, dryRun)
		return nil
	}
	if err := writeOutput(target, shown, content, dryRun); err != nil {
		return err
	}
	reportFile(shown, fileWritten, dryRun)
	switch {
	case replaced:
		errf("⚠️  %s had no argus markers, replaced with a fresh generated file (--force); the old text was not kept\n", shown)
	case outcome == merger.Adopted:
		errf("ℹ️  %s had no argus markers: generated block added on top, your text kept below it (remove any part an older argus run wrote)\n", shown)
	case outcome == merger.Malformed:
		errf("⚠️  %s: unbalanced or repeated ARGUS:AUTO markers, replaced with a fresh generated file (--force); the old text was not kept\n", shown)
	case rel == claudeFile && mergeMode && outcome == merger.Replaced:
		noticeMigration(shown, string(existing))
	}
	return nil
}

// noticeMigration tells the user once that an old CLAUDE.md, whose AUTO
// block held the generated body, now imports AGENTS.md.
func noticeMigration(shown, existing string) {
	start := strings.Index(existing, merger.AutoStartMarker)
	end := strings.Index(existing, merger.AutoEndMarker)
	if strings.TrimSpace(existing[start+len(merger.AutoStartMarker):end]) == shimImport {
		return
	}
	kept := 0
	for _, line := range strings.Split(existing[:start]+"\n"+existing[end+len(merger.AutoEndMarker):], "\n") {
		if strings.TrimSpace(line) != "" {
			kept++
		}
	}
	msg := shown + ": the generated body moved to AGENTS.md and this file now imports it"
	if kept > 0 {
		msg += fmt.Sprintf(", %d lines of your other text were kept (delete anything an older argus run wrote)", kept)
	}
	errf("ℹ️  %s\n", msg)
}

// skipLinked reports whether rel must be left alone because AGENTS.md or
// CLAUDE.md is a symlink: argus never writes through or replaces those, not
// even with --force. A CLAUDE.md linked to AGENTS.md already is a shim.
func skipLinked(root, shown, rel string, dryRun bool) bool {
	if rel != agentsFile && rel != claudeFile {
		return false
	}
	agents, claude := filepath.Join(root, agentsFile), filepath.Join(root, claudeFile)
	switch {
	case isSymlink(agents) && linksTo(agents, claude):
		record(shown, "skipped")
		if rel == agentsFile && notifyOnce(shown) {
			errf("⚠️  %s is a symlink to CLAUDE.md, both left unchanged (remove the symlink: argus then writes AGENTS.md and makes CLAUDE.md import it)\n", shown)
		}
	case rel == claudeFile && isSymlink(claude) && linksTo(claude, agents):
		reportFile(shown, fileUnchanged, dryRun)
	case isSymlink(filepath.Join(root, rel)):
		record(shown, "skipped")
		if notifyOnce(shown) {
			errf("⚠️  %s is a symlink, left unchanged (replace it with a regular file to let argus write it)\n", shown)
		}
	default:
		return false
	}
	return true
}

func isSymlink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

// linksTo reports whether the symlink link resolves to other: the same file
// when both exist, else the same cleaned path (a dangling link).
func linksTo(link, other string) bool {
	a, errA := os.Stat(link)
	b, errB := os.Stat(other)
	if errA == nil && errB == nil {
		return os.SameFile(a, b)
	}
	dest, err := os.Readlink(link)
	if err != nil {
		return false
	}
	if !filepath.IsAbs(dest) {
		dest = filepath.Join(filepath.Dir(link), dest)
	}
	return filepath.Clean(dest) == filepath.Clean(other)
}

// hasLine reports whether text has a line equal to want, surrounding
// whitespace aside.
func hasLine(text []byte, want string) bool {
	for _, line := range strings.Split(string(text), "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
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
		record(shown, "up to date")
		return fileUnchanged, nil
	case !force:
		record(shown, "skipped")
		skipped = append(skipped, shown)
		return fileSkipped, nil
	}
	return fileWritten, writeOutput(target, shown, content, dryRun)
}

// writeOutput previews content in dry-run mode and otherwise writes it with
// writeFile.
func writeOutput(target, shown string, content []byte, dryRun bool) error {
	if checked != nil {
		state := "out of date"
		if _, err := os.Lstat(target); err != nil {
			state = "missing"
		}
		record(shown, state)
		return nil
	}
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
	tmp, err := os.CreateTemp(filepath.Dir(target), tempPattern)
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
		record(shown, "up to date")
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

// isBlank reports whether a file holds nothing but whitespace and a BOM.
func isBlank(b []byte) bool {
	return len(bytes.TrimSpace(bytes.TrimPrefix(b, []byte("\ufeff")))) == 0
}

// sameText reports whether a and b are equal when CRLF is read as LF.
func sameText(a, b []byte) bool {
	crlf, lf := []byte("\r\n"), []byte("\n")
	return bytes.Equal(bytes.ReplaceAll(a, crlf, lf), bytes.ReplaceAll(b, crlf, lf))
}
