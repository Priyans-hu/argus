package detector

import (
	"cmp"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// Manifest is a dependency manifest in the walked file list.
type Manifest struct {
	Kind string // package.json, go.mod, pyproject.toml, requirements.txt or Cargo.toml
	Path string // slash path relative to the scan root
	Dir  string // slash directory relative to the scan root, empty for the root
}

// FindManifests lists the manifests at the root and up to two directories
// deep that are project source, shallowest first.
func FindManifests(files []types.FileInfo) []Manifest {
	var found []Manifest
	for _, f := range files {
		rel := filepath.ToSlash(f.Path)
		kind := manifestKind(rel)
		if f.IsDir || kind == "" || strings.Count(rel, "/") > 2 || !IsProjectSource(rel) {
			continue
		}
		dir := path.Dir(rel)
		if dir == "." {
			dir = ""
		}
		found = append(found, Manifest{Kind: kind, Path: rel, Dir: dir})
	}
	slices.SortFunc(found, func(a, b Manifest) int {
		return cmp.Or(cmp.Compare(strings.Count(a.Path, "/"), strings.Count(b.Path, "/")), cmp.Compare(a.Path, b.Path))
	})
	return found
}

// manifestKind returns the kind of the manifest at the slash path rel, or
// "" for none. Requirements files are requirements*.txt, or any .txt or .in
// file in a directory named requirements (requirements/base.txt).
func manifestKind(rel string) string {
	name, ext := path.Base(rel), path.Ext(rel)
	switch {
	case name == "package.json", name == "go.mod", name == "pyproject.toml", name == "Cargo.toml":
		return name
	case strings.HasPrefix(name, "requirements") && ext == ".txt",
		path.Base(path.Dir(rel)) == "requirements" && (ext == ".txt" || ext == ".in"):
		return "requirements.txt"
	}
	return ""
}
