package detector

import (
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestIsProjectSource(t *testing.T) {
	tests := []struct {
		rel  string
		want bool
	}{
		{"src/app.ts", true},
		{"package.json", true},
		{"docs", true},
		{"src/docs.ts", true},
		{"gen/x.go", true},
		{"src/demo.go", true},
		{"docs/x.ts", false},
		{"apps/docs/x.tsx", false},
		{"packages/ui/examples/y.tsx", false},
		{"packages/ui/src/demo/q.ts", false},
		{"Docs/x.ts", false},
		{`vendor\x\y.go`, false},
		{"__mocks__/api.ts", false},
		{"pkg/testdata/a.go", false},
	}
	for dir := range nonSourceDirs {
		tests = append(tests, struct {
			rel  string
			want bool
		}{"a/" + dir + "/b.go", false})
	}
	for _, tt := range tests {
		if got := IsProjectSource(tt.rel); got != tt.want {
			t.Errorf("IsProjectSource(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		rel  string
		want bool
	}{
		{"handler_test.go", true},
		{"src/a.test.ts", true},
		{"src/Button.spec.tsx", true},
		{"test_api.py", true},
		{"pkg/api_test.py", true},
		{"conftest.py", true},
		{"tests/helpers.py", true},
		{"src/__tests__/a.tsx", true},
		{"spec/models/user.rb", true},
		{"e2e/home.cy.ts", true},
		{"test/x.go", true},
		{"Tests/X.swift", true},
		{`src\__tests__\a.ts`, true},
		{"handler.go", false},
		{"src/latest.go", false},
		{"src/contest.py", false},
		{"src/test_utils.ts", false},
		{"testing/x.go", false},
		{"testdata/x.go", false},
		{"src/attest.py", false},
	}
	for _, tt := range tests {
		if got := IsTestFile(tt.rel); got != tt.want {
			t.Errorf("IsTestFile(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}

func TestProjectSourceFiles(t *testing.T) {
	var files []types.FileInfo
	for _, rel := range []string{"src/a.go", "docs/b.md", "apps/docs/x.tsx", "packages/ui/examples/y.tsx", "gen/x.go", "src/a_test.go"} {
		files = append(files, types.FileInfo{Path: rel})
	}
	var got []string
	for _, f := range ProjectSourceFiles(files) {
		got = append(got, f.Path)
	}
	if want := []string{"src/a.go", "gen/x.go", "src/a_test.go"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ProjectSourceFiles = %v, want %v", got, want)
	}
}
