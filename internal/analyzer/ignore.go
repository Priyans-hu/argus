package analyzer

import (
	"path"
	"strings"
)

// ignoreRule is one gitignore-style pattern from the built-in defaults, the
// root .gitignore or the ignore list in .argus.yaml.
type ignoreRule struct {
	pattern  string
	anchored bool // matched against the path from the scan root
	dirOnly  bool
}

// parseIgnoreRule parses one ignore line. Blank lines, comments and
// negations ("!", unsupported) yield no rule.
func parseIgnoreRule(line string) (ignoreRule, bool) {
	p := strings.TrimSpace(line)
	if p == "" || p[0] == '#' || p[0] == '!' {
		return ignoreRule{}, false
	}
	var r ignoreRule
	p, floating := strings.CutPrefix(p, "**/")
	hasSlash := strings.Contains(strings.TrimSuffix(p, "/"), "/")
	if rest, ok := strings.CutSuffix(p, "/**"); ok {
		p, r.dirOnly = rest, true
	}
	if rest, ok := strings.CutSuffix(p, "/"); ok {
		p, r.dirOnly = rest, true
	}
	p, rooted := strings.CutPrefix(p, "/")
	r.anchored = rooted || (hasSlash && !floating)
	r.pattern = strings.ReplaceAll(p, "**", "*")
	return r, r.pattern != ""
}

// matches reports whether the rule covers rel, a path relative to the scan
// root. A pattern without a slash matches the entry name at any depth.
func (r ignoreRule) matches(rel string, isDir bool) bool {
	if r.dirOnly && !isDir {
		return false
	}
	rel = strings.ReplaceAll(rel, `\`, "/")
	if r.anchored {
		ok, _ := path.Match(r.pattern, rel)
		return ok
	}
	if !strings.Contains(r.pattern, "/") {
		ok, _ := path.Match(r.pattern, path.Base(rel))
		return ok
	}
	for {
		if ok, _ := path.Match(r.pattern, rel); ok {
			return true
		}
		i := strings.IndexByte(rel, '/')
		if i < 0 {
			return false
		}
		rel = rel[i+1:]
	}
}
