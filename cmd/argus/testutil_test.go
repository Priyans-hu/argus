package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// resetFlags puts the flag and reporting globals at their defaults for one
// test and again when it ends.
func resetFlags(t *testing.T) {
	t.Helper()
	reset := func() {
		outputFormat, dryRun, verbose, force = "claude", false, false, false
		mergeMode, addCustomBlock, compactMode, parallel = true, false, false, true
		usageMode, monorepoMode, aiMode = false, false, false
		skipped, watchNotified = nil, nil
	}
	reset()
	t.Cleanup(reset)
}

// captureOutput sends stdout and stderr writes to buffers for one test.
func captureOutput(t *testing.T) (out, errOut *bytes.Buffer) {
	t.Helper()
	out, errOut = &bytes.Buffer{}, &bytes.Buffer{}
	oldOut, oldErr := stdout, stderr
	stdout, stderr = out, errOut
	t.Cleanup(func() { stdout, stderr = oldOut, oldErr })
	return out, errOut
}

// executeRoot runs rootCmd with args and returns what cobra printed to its
// output and error streams. Only use it for commands that return before
// doing real work (version, help, argument and path errors).
func executeRoot(t *testing.T, args ...string) (out, errOut string, err error) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	rootCmd.SetOut(&outBuf)
	rootCmd.SetErr(&errBuf)
	rootCmd.SetArgs(args)
	t.Cleanup(resetRoot)
	err = rootCmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

// runCLI runs the command line as main does and returns everything that
// would reach stderr (cobra's usage included) and the exit code. Only use it
// for commands that fail before doing real work.
func runCLI(t *testing.T, args ...string) (errOut string, code int) {
	t.Helper()
	var buf bytes.Buffer
	rootCmd.SetOut(&buf) // cobra prints usage to stderr unless an output is set
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs(args)
	t.Cleanup(resetRoot)
	code = execute(&buf)
	return buf.String(), code
}

// resetRoot undoes what executing rootCmd leaves behind.
func resetRoot() {
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	rootCmd.SetArgs(nil)
	if f := rootCmd.Flags().Lookup("version"); f != nil {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	}
	rootCmd.SilenceUsage = false
	for _, c := range rootCmd.Commands() {
		c.SilenceUsage = false
	}
}

// writeFiles writes slash-path files under root.
func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

// snapshotTree maps the slash path of every file under root, except .git,
// to its content.
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	tree := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case d.IsDir() && d.Name() == ".git":
			return filepath.SkipDir
		case d.IsDir():
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		tree[filepath.ToSlash(rel)] = string(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

// treeDiff lists the paths whose content differs between two snapshots.
func treeDiff(a, b map[string]string) []string {
	var diff []string
	for path, content := range a {
		if other, ok := b[path]; !ok || other != content {
			diff = append(diff, path)
		}
	}
	for path := range b {
		if _, ok := a[path]; !ok {
			diff = append(diff, path)
		}
	}
	slices.Sort(diff)
	return diff
}

// readFile returns the content of a slash path under root.
func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// goAnalysis is a small Go project for which every format renders output,
// claude-code included (agents, rules, skills and a settings.json hook).
func goAnalysis() *types.Analysis {
	return &types.Analysis{
		ProjectName: "demo",
		TechStack:   types.TechStack{Languages: []types.Language{{Name: "Go", Percentage: 100}}},
		Commands: []types.Command{
			{Name: "go build ./...", Description: "Build all packages"},
			{Name: "go test ./...", Description: "Run all tests"},
			{Name: "go fmt ./...", Description: "Format all Go files"},
		},
	}
}
