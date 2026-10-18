package detector

import (
	"cmp"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/Priyans-hu/argus/pkg/types"
	"golang.org/x/mod/modfile"
)

// ReadDependencies reads the dependencies declared in the manifests
// FindManifests lists, one entry per type and name (the shallowest manifest
// wins), sorted by type and name.
func ReadDependencies(root string, files []types.FileInfo) []types.Dependency {
	var deps []types.Dependency
	seen := map[[2]string]bool{}
	for _, m := range FindManifests(files) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(m.Path)))
		if err != nil {
			continue
		}
		for _, d := range parseManifest(m.Kind, data) {
			if key := [2]string{d.Type, d.Name}; !seen[key] {
				seen[key] = true
				deps = append(deps, d)
			}
		}
	}
	slices.SortFunc(deps, func(a, b types.Dependency) int {
		return cmp.Or(cmp.Compare(a.Type, b.Type), cmp.Compare(a.Name, b.Name))
	})
	return deps
}

func parseManifest(kind string, data []byte) []types.Dependency {
	switch kind {
	case "package.json":
		return npmDeps(data)
	case "go.mod":
		return goDeps(data)
	case "pyproject.toml":
		return pyprojectDeps(data)
	case "requirements.txt":
		return requirementsDeps(data)
	case "Cargo.toml":
		return cargoDeps(data)
	}
	return nil
}

func npmDeps(data []byte) []types.Dependency {
	var pkg struct {
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		PeerDependencies     map[string]string `json:"peerDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return nil
	}
	var deps []types.Dependency
	for _, g := range []struct {
		typ  string
		deps map[string]string
	}{{"runtime", pkg.Dependencies}, {"dev", pkg.DevDependencies}, {"peer", pkg.PeerDependencies}, {"optional", pkg.OptionalDependencies}} {
		for _, name := range slices.Sorted(maps.Keys(g.deps)) {
			deps = append(deps, types.Dependency{Name: name, Version: g.deps[name], Type: g.typ})
		}
	}
	return deps
}

// goDeps returns the direct requirements of a go.mod.
func goDeps(data []byte) []types.Dependency {
	mod, err := modfile.ParseLax("go.mod", data, nil)
	if err != nil {
		return nil
	}
	var deps []types.Dependency
	for _, r := range mod.Require {
		if !r.Indirect && !isInternalPackage(r.Mod.Path) {
			deps = append(deps, types.Dependency{Name: r.Mod.Path, Version: r.Mod.Version, Type: "runtime"})
		}
	}
	return deps
}

// isInternalPackage checks if a package path is internal/vendor
func isInternalPackage(pkg string) bool {
	return strings.Contains(pkg, "/internal/") || strings.Contains(pkg, "/feature/")
}

func pyprojectDeps(data []byte) []types.Dependency {
	var py struct {
		Project struct {
			Dependencies         []string            `toml:"dependencies"`
			OptionalDependencies map[string][]string `toml:"optional-dependencies"`
		} `toml:"project"`
		Tool struct {
			Poetry struct {
				Dependencies    map[string]any `toml:"dependencies"`
				DevDependencies map[string]any `toml:"dev-dependencies"`
				Group           map[string]struct {
					Dependencies map[string]any `toml:"dependencies"`
				} `toml:"group"`
			} `toml:"poetry"`
		} `toml:"tool"`
		DependencyGroups map[string][]any `toml:"dependency-groups"`
	}
	if _, err := toml.Decode(string(data), &py); err != nil {
		return nil
	}
	var deps []types.Dependency
	add := func(spec any, typ string) {
		if s, ok := spec.(string); ok {
			if d, ok := pythonDep(s, typ); ok {
				deps = append(deps, d)
			}
		}
	}
	poetry := func(m map[string]any, typ string) {
		for _, name := range slices.Sorted(maps.Keys(m)) {
			if pythonName(name) != "python" {
				deps = append(deps, types.Dependency{Name: pythonName(name), Version: tableVersion(m[name]), Type: typ})
			}
		}
	}

	for _, spec := range py.Project.Dependencies {
		add(spec, "runtime")
	}
	for _, group := range slices.Sorted(maps.Keys(py.Project.OptionalDependencies)) {
		for _, spec := range py.Project.OptionalDependencies[group] {
			add(spec, "optional")
		}
	}
	poetry(py.Tool.Poetry.Dependencies, "runtime")
	poetry(py.Tool.Poetry.DevDependencies, "dev")
	for _, group := range slices.Sorted(maps.Keys(py.Tool.Poetry.Group)) {
		poetry(py.Tool.Poetry.Group[group].Dependencies, "dev")
	}
	for _, group := range slices.Sorted(maps.Keys(py.DependencyGroups)) {
		for _, spec := range py.DependencyGroups[group] {
			add(spec, "dev")
		}
	}
	return deps
}

// requirementsDeps reads a requirements*.txt; options, URLs and paths are skipped.
func requirementsDeps(data []byte) []types.Dependency {
	var deps []types.Dependency
	for _, line := range strings.Split(string(data), "\n") {
		line, _, _ = strings.Cut(line, "#")
		line, _, _ = strings.Cut(line, " -")
		if d, ok := pythonDep(strings.TrimSuffix(strings.TrimSpace(line), `\`), "runtime"); ok {
			deps = append(deps, d)
		}
	}
	return deps
}

var pythonReq = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9._-]*)\s*(?:\[[^\]]*\])?\s*([(<>=!~].*)?$`)

// pythonDep parses a PEP 508 requirement such as `Django[argon2]>=4.2 ;
// python_version>"3.8"`: extras and markers are dropped, `==x` gives the
// version x and any other specifier is kept as written. URLs and paths do
// not parse.
func pythonDep(spec, typ string) (types.Dependency, bool) {
	spec, _, _ = strings.Cut(spec, ";")
	m := pythonReq.FindStringSubmatch(strings.TrimSpace(spec))
	if m == nil {
		return types.Dependency{}, false
	}
	version := strings.TrimSpace(strings.Trim(strings.TrimSpace(m[2]), "()"))
	if v, ok := strings.CutPrefix(version, "=="); ok && !strings.ContainsAny(v, ",=") {
		version = strings.TrimSpace(v)
	}
	return types.Dependency{Name: pythonName(m[1]), Version: version, Type: typ}, true
}

var pythonNameSeparators = regexp.MustCompile(`[-_.]+`)

// pythonName normalizes a package name per PEP 503.
func pythonName(name string) string {
	return strings.ToLower(pythonNameSeparators.ReplaceAllString(name, "-"))
}

// tableVersion returns a dependency's version from `name = "1"` or
// `name = { version = "1" }`.
func tableVersion(v any) string {
	if t, ok := v.(map[string]any); ok {
		v = t["version"]
	}
	s, _ := v.(string)
	return s
}

func cargoDeps(data []byte) []types.Dependency {
	var cargo struct {
		Dependencies      map[string]any `toml:"dependencies"`
		DevDependencies   map[string]any `toml:"dev-dependencies"`
		BuildDependencies map[string]any `toml:"build-dependencies"`
		Workspace         struct {
			Dependencies map[string]any `toml:"dependencies"`
		} `toml:"workspace"`
	}
	if _, err := toml.Decode(string(data), &cargo); err != nil {
		return nil
	}
	var deps []types.Dependency
	for _, g := range []struct {
		typ  string
		deps map[string]any
	}{{"runtime", cargo.Dependencies}, {"dev", cargo.DevDependencies}, {"build", cargo.BuildDependencies}, {"runtime", cargo.Workspace.Dependencies}} {
		for _, key := range slices.Sorted(maps.Keys(g.deps)) {
			name := key
			if t, ok := g.deps[key].(map[string]any); ok {
				if real, ok := t["package"].(string); ok {
					name = real
				}
			}
			name = strings.ReplaceAll(strings.ToLower(name), "_", "-")
			deps = append(deps, types.Dependency{Name: name, Version: tableVersion(g.deps[key]), Type: g.typ})
		}
	}
	return deps
}

// DepSet answers whether a dependency is declared, by exact name. It maps a
// name to its version, empty when the name is declared only as a peer or
// without a version.
type DepSet map[string]string

// NewDepSet indexes deps; the first non-empty version of a name wins.
func NewDepSet(deps []types.Dependency) DepSet {
	set := DepSet{}
	for _, d := range deps {
		version := d.Version
		if d.Type == "peer" {
			version = ""
		}
		if old, ok := set[d.Name]; !ok || old == "" {
			set[d.Name] = version
		}
	}
	return set
}

// Has reports whether any of names is declared.
func (s DepSet) Has(names ...string) bool {
	for _, name := range names {
		if _, ok := s[name]; ok {
			return true
		}
	}
	return false
}

// HasPrefix reports whether a declared name starts with any of prefixes,
// such as the npm scope "@trpc/".
func (s DepSet) HasPrefix(prefixes ...string) bool {
	for name := range s {
		for _, p := range prefixes {
			if strings.HasPrefix(name, p) {
				return true
			}
		}
	}
	return false
}

// HasModule reports whether the Go module path, or a path below it such as
// its /v2 major version, is declared.
func (s DepSet) HasModule(path string) bool {
	for name := range s {
		if name == path || strings.HasPrefix(name, path+"/") {
			return true
		}
	}
	return false
}

// declares reports whether one of packages is declared: by exact name, a Go
// module also by a path below it, and a name ending in "/", such as the npm
// scope "@trpc/", as a prefix.
func (s DepSet) declares(packages []string) bool {
	for _, p := range packages {
		if strings.HasSuffix(p, "/") && s.HasPrefix(p) || s.HasModule(p) {
			return true
		}
	}
	return false
}
