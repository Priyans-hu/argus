package detector

import "testing"

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
