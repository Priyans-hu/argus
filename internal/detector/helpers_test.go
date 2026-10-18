package detector

import (
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// writeProject writes slash-path files into a temp dir and returns the dir
// with the file list the walker would produce for them (directories left out).
func writeProject(t *testing.T, files map[string]string) (string, []types.FileInfo) {
	t.Helper()
	root := t.TempDir()
	var infos []types.FileInfo
	for _, rel := range slices.Sorted(maps.Keys(files)) {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(files[rel]), 0644); err != nil {
			t.Fatal(err)
		}
		infos = append(infos, types.FileInfo{
			Path:      filepath.FromSlash(rel),
			Name:      filepath.Base(rel),
			Extension: strings.ToLower(filepath.Ext(rel)),
			Size:      int64(len(files[rel])),
		})
	}
	return root, infos
}

// stableResult calls f 50 times and fails unless every result equals the
// first. Go randomizes map iteration on every range, so an order that
// depends on it shows up within a few runs.
func stableResult[T any](t *testing.T, f func() T) T {
	t.Helper()
	first := f()
	for i := 1; i < 50; i++ {
		if got := f(); !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d differs from run 0:\n got %+v\nwant %+v", i, got, first)
		}
	}
	return first
}
