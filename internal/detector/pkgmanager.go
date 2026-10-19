package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// lockfiles in the order DetectPackageManager checks them.
var lockfiles = []struct{ name, manager string }{
	{"bun.lock", "bun"}, {"bun.lockb", "bun"}, {"pnpm-lock.yaml", "pnpm"}, {"yarn.lock", "yarn"},
	{"package-lock.json", "npm"}, {"npm-shrinkwrap.json", "npm"},
}

// DetectPackageManager returns the package manager of the package in dir;
// dir and root are absolute, dir at or below root. The nearest package.json
// from dir up to root that declares packageManager or
// devEngines.packageManager.name (npm, pnpm, yarn or bun) wins, as with
// Corepack; else the nearest directory holding a lockfile; else pnpm when
// one of those directories has pnpm-workspace.yaml; else npm.
func DetectPackageManager(dir, root string) string {
	dirs := dirsUpTo(dir, root)
	for _, d := range dirs {
		if name, _ := declaredManager(d); name != "" {
			return name
		}
	}
	for _, d := range dirs {
		for _, l := range lockfiles {
			if fileExists(filepath.Join(d, l.name)) {
				return l.manager
			}
		}
	}
	for _, d := range dirs {
		if fileExists(filepath.Join(d, "pnpm-workspace.yaml")) {
			return "pnpm"
		}
	}
	return "npm"
}

// installCommands returns `<pm> install` and, when that manager's lockfile
// exists from dir up to root, its clean install. yarn ci and pnpm ci do not
// exist.
func installCommands(pm, dir, root string) []string {
	commands := []string{pm + " install"}
	dirs := dirsUpTo(dir, root)
	hasLock := false
	for _, d := range dirs {
		for _, l := range lockfiles {
			if l.manager == pm && fileExists(filepath.Join(d, l.name)) {
				hasLock = true
			}
		}
	}
	if !hasLock {
		return commands
	}
	switch pm {
	case "npm":
		return append(commands, "npm ci")
	case "yarn":
		if isYarnBerry(dirs) {
			return append(commands, "yarn install --immutable")
		}
		return append(commands, "yarn install --frozen-lockfile")
	}
	return append(commands, pm+" install --frozen-lockfile")
}

// scriptCommand runs a package.json script; bun test is Bun's own runner,
// so only npm, yarn and pnpm get the start and test shortcuts.
func scriptCommand(pm, script string) string {
	if pm != "bun" && (script == "start" || script == "test") {
		return pm + " " + script
	}
	return pm + " run " + script
}

// isYarnBerry reports Yarn 2 or later: a .yarnrc.yml, or a packageManager
// declaration of yarn@2 or later.
func isYarnBerry(dirs []string) bool {
	for _, d := range dirs {
		if fileExists(filepath.Join(d, ".yarnrc.yml")) {
			return true
		}
		if name, version := declaredManager(d); name == "yarn" {
			major, _, _ := strings.Cut(version, ".")
			n, err := strconv.Atoi(major)
			return err == nil && n >= 2
		}
	}
	return false
}

// declaredManager returns the manager a package.json in dir declares and
// its version, or "" for none, an unknown name or the array form of
// devEngines.packageManager.
func declaredManager(dir string) (name, version string) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return "", ""
	}
	var pkg struct {
		PackageManager string `json:"packageManager"`
		DevEngines     struct {
			PackageManager json.RawMessage `json:"packageManager"`
		} `json:"devEngines"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return "", ""
	}
	spec := pkg.PackageManager
	if spec == "" {
		var dev struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		if json.Unmarshal(pkg.DevEngines.PackageManager, &dev) == nil {
			spec = dev.Name + "@" + dev.Version
		}
	}
	name, version, _ = strings.Cut(spec, "@")
	switch name {
	case "npm", "pnpm", "yarn", "bun":
		return name, version
	}
	return "", ""
}

// dirsUpTo lists dir and its parents up to root.
func dirsUpTo(dir, root string) []string {
	dir, root = filepath.Clean(dir), filepath.Clean(root)
	dirs := []string{dir}
	for dir != root && strings.HasPrefix(dir, root+string(filepath.Separator)) {
		dir = filepath.Dir(dir)
		dirs = append(dirs, dir)
	}
	return dirs
}
