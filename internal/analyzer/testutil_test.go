package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/Priyans-hu/argus/pkg/types"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// writeTree writes slash-path files under root.
func writeTree(t *testing.T, root string, files map[string]string) {
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

// commitTree makes root a git repository with one commit of everything in it.
func commitTree(t *testing.T, root string) {
	t.Helper()
	repo, err := git.PlainInit(root, false)
	if err != nil {
		t.Fatal(err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
		t.Fatal(err)
	}
	sig := &object.Signature{Name: "Dev", Email: "dev@example.com", When: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}
	if _, err := wt.Commit("feat: add the fixture", &git.CommitOptions{Author: sig, Committer: sig}); err != nil {
		t.Fatal(err)
	}
}

// analyzeBoth runs the sequential and the parallel analyzer on root and
// fails unless they agree.
func analyzeBoth(t *testing.T, root string, cfg *types.Config) (sequential, parallel *types.Analysis) {
	t.Helper()
	ctx := context.Background()
	sequential, err := NewAnalyzer(root, cfg).Analyze(ctx)
	if err != nil {
		t.Fatal(err)
	}
	parallel, err = NewParallelAnalyzer(root, cfg).Analyze(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sequential, parallel) {
		t.Fatalf("sequential and parallel analyses differ:\nsequential %+v\nparallel   %+v", sequential, parallel)
	}
	return sequential, parallel
}
