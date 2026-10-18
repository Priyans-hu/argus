package detector

import (
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestDominantPatternTies(t *testing.T) {
	tests := []struct {
		counts map[NamingPattern]int
		want   NamingPattern
	}{
		{map[NamingPattern]int{PascalCase: 3, CamelCase: 3}, PascalCase},
		{map[NamingPattern]int{CamelCase: 4, KebabCase: 4, SnakeCase: 4}, CamelCase},
		{map[NamingPattern]int{KebabCase: 3, SnakeCase: 3}, KebabCase},
		{map[NamingPattern]int{SnakeCase: 5, PascalCase: 3}, SnakeCase},
	}
	for _, tt := range tests {
		got := stableResult(t, func() NamingPattern {
			pattern, _ := dominantPattern(tt.counts)
			return pattern
		})
		if got != tt.want {
			t.Errorf("dominantPattern(%v) = %s, want %s", tt.counts, got, tt.want)
		}
	}
}

func TestPathAliasesAreSorted(t *testing.T) {
	root, files := writeProject(t, map[string]string{
		"tsconfig.json": `{"compilerOptions": {"paths": {"~/*": ["./src/*"], "@components/*": ["./src/components/*"], "@/*": ["./src/*"], "#lib/*": ["./lib/*"]}}}`,
	})

	conventions := stableResult(t, func() []types.Convention {
		return NewConventionDetector(root, files).detectImportStyles()
	})

	want := types.Convention{
		Category:    "imports",
		Description: "Path aliases configured: #lib, @, @components, ~",
		Example:     "import { Button } from '#lib/components/Button'",
	}
	if len(conventions) == 0 || conventions[0] != want {
		t.Errorf("conventions = %+v, want first %+v", conventions, want)
	}
}
