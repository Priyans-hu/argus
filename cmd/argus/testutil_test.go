package main

import (
	"bytes"
	"os"
	"path/filepath"
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
