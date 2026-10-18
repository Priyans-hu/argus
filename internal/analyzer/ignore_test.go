package analyzer

import "testing"

func TestIgnoreRuleMatches(t *testing.T) {
	tests := []struct {
		pattern string
		rel     string
		isDir   bool
		want    bool
	}{
		{"gen.go", "gen.go", false, true},
		{"gen.go", "a/b/gen.go", false, true},
		{"gen.go", "a/gen.go.bak", false, false},
		{"build", "src/build", true, true},
		{"build", "app/build.gradle.kts", false, false},
		{"bin", "internal/binder", true, false},
		{"Build", "build", true, false},
		{"*.tmp", "c/x.tmp", false, true},
		{"*.tmp", "c/x.tmpl", false, false},
		{"/scripts/old", "scripts/old", true, true},
		{"/scripts/old", "d/scripts/old", true, false},
		{"a/b", "a/b", true, true},
		{"a/b", "x/a/b", true, false},
		{"logs/", "logs", true, true},
		{"logs/", "x/logs", true, true},
		{"logs/", "logs", false, false},
		{"keep.txt/", "keep.txt", false, false},
		{"docs/legacy/", "docs/legacy", true, true},
		{"docs/legacy/", "x/docs/legacy", true, false},
		{"**/gen2.go", "x/y/gen2.go", false, true},
		{"**/x/y", "a/x/y", true, true},
		{"**/x/y", "a/x/z", true, false},
		{"docs2/**", "docs2", true, true},
		{"docs2/**", "docs2", false, false},
		{"docs2/**", "x/docs2", true, false},
		{"a/**/b", "a/x/b", false, true},
		{"gen.go", `a\b\gen.go`, false, true},
		{"a/b", `a\b`, true, true},
		{"/etc", "src/etc", true, false},
		{"../x", "x", true, false},
		{"..", "a", true, false},
		{"./src/../src", "src", true, false},
	}
	for _, tt := range tests {
		r, ok := parseIgnoreRule(tt.pattern)
		if !ok {
			t.Errorf("parseIgnoreRule(%q) gave no rule", tt.pattern)
			continue
		}
		if got := r.matches(tt.rel, tt.isDir); got != tt.want {
			t.Errorf("%q matches %q (dir %v) = %v, want %v", tt.pattern, tt.rel, tt.isDir, got, tt.want)
		}
	}
}

func TestIgnoreRuleSkipsNonPatterns(t *testing.T) {
	for _, line := range []string{"", "   ", "# comment", "#c", "!a/keep.txt", "/", "//", "/**"} {
		if r, ok := parseIgnoreRule(line); ok {
			t.Errorf("parseIgnoreRule(%q) = %+v, want no rule", line, r)
		}
	}
}
