package analyzer

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func richAnalysis() *types.Analysis {
	return &types.Analysis{
		ProjectName: "shop",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "Python", Percentage: 20},
				{Name: "TypeScript", Percentage: 40},
				{Name: "Go", Version: "1.22", Percentage: 40},
			},
			Frameworks: []types.Framework{
				{Name: "React", Category: "frontend", Version: "18"},
				{Name: "Gin", Category: "backend"},
				{Name: "Vitest", Category: "testing"},
				{Name: "Echo", Category: "backend"},
				{Name: "Gin", Category: "backend"},
			},
			Databases: []string{"Redis", "PostgreSQL", "Redis", "MongoDB"},
			Tools:     []string{"Docker", "Make", "Docker", "GitHub Actions"},
		},
		Structure: types.ProjectStructure{
			Directories: []types.Directory{
				{Path: "web", Purpose: "Frontend", FileCount: 4},
				{Path: "api", Purpose: "API", FileCount: 9},
				{Path: "api/handlers", FileCount: 3},
			},
			RootFiles: []string{"go.mod", "Makefile", "README.md", ".gitignore"},
		},
		Conventions: []types.Convention{
			{Category: "testing", Description: "Tests are colocated"},
			{Category: "code-style", Description: "Use gofmt"},
		},
		Dependencies: []types.Dependency{
			{Name: "vitest", Version: "1", Type: "dev"},
			{Name: "react", Version: "18", Type: "runtime"},
			{Name: "axios", Version: "1", Type: "runtime"},
			{Name: "eslint", Version: "8", Type: "dev"},
			{Name: "axios", Version: "0.27", Type: "dev"},
		},
		Commands: []types.Command{{Name: "make test"}, {Name: "make build"}},
		KeyFiles: []types.KeyFile{{Path: "main.go"}, {Path: "go.mod"}},
		Endpoints: []types.Endpoint{
			{Method: "POST", Path: "/users", File: "api/users.go", Line: 30},
			{Method: "GET", Path: "/users", File: "api/users.go", Line: 12},
			{Method: "GET", Path: "/orders", File: "api/orders.go", Line: 12},
			{Method: "DELETE", Path: "/users/:id", File: "api/users.go", Line: 30},
		},
		CodePatterns: &types.CodePatterns{
			Testing: []types.PatternInfo{
				{Name: "t.Run", Category: "testing", FileCount: 8},
				{Name: "func Test", Category: "testing", FileCount: 13, Examples: []string{"a_test.go", "b_test.go"}},
				{Name: "t.Error", Category: "testing", FileCount: 13},
				{Name: "httptest", Category: "testing", FileCount: 1},
			},
			DataFetching: []types.PatternInfo{
				{Name: "axios", Category: "data-fetching", FileCount: 2},
				{Name: "useEffect", Category: "React Hooks", FileCount: 2},
				{Name: "fetch", Category: "data-fetching", FileCount: 5},
			},
			GoPatterns: []types.PatternInfo{
				{Name: "go func", Category: "concurrency", FileCount: 3},
				{Name: "protobuf", Category: "serialization", FileCount: 2},
				{Name: "context.Context", Category: "concurrency", FileCount: 6},
			},
			MLPatterns: []types.PatternInfo{
				{Name: "PyTorch", Category: "ml", FileCount: 1, Usage: "training"},
				{Name: "PyTorch", Category: "ml", FileCount: 1, Usage: "inference"},
			},
		},
		ArchitectureInfo: &types.ArchitectureInfo{
			Layers: []types.ArchitectureLayer{{Name: "internal"}, {Name: "cmd"}},
		},
	}
}

func shuffle[T any](r *rand.Rand, s []T) {
	r.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

func TestNormalizeOrdersArbitrarySlices(t *testing.T) {
	want := richAnalysis()
	normalize(want)

	r := rand.New(rand.NewPCG(1, 2))
	for i := range 20 {
		a := richAnalysis()
		shuffle(r, a.TechStack.Languages)
		shuffle(r, a.TechStack.Frameworks)
		shuffle(r, a.TechStack.Databases)
		shuffle(r, a.TechStack.Tools)
		shuffle(r, a.Structure.Directories)
		shuffle(r, a.Structure.RootFiles)
		shuffle(r, a.Dependencies)
		shuffle(r, a.Endpoints)
		shuffle(r, a.CodePatterns.Testing)
		shuffle(r, a.CodePatterns.DataFetching)
		shuffle(r, a.CodePatterns.GoPatterns)
		shuffle(r, a.CodePatterns.MLPatterns)

		normalize(a)
		if !reflect.DeepEqual(a, want) {
			t.Fatalf("run %d: normalize gave\n%+v\nwant\n%+v", i, a, want)
		}
	}

	again := richAnalysis()
	normalize(again)
	normalize(again)
	if !reflect.DeepEqual(again, want) {
		t.Errorf("normalize is not idempotent:\n%+v\nwant\n%+v", again, want)
	}
}

func TestNormalizeResult(t *testing.T) {
	a := richAnalysis()
	normalize(a)

	names := func(n int, name func(int) string) []string {
		var out []string
		for i := range n {
			out = append(out, name(i))
		}
		return out
	}
	checks := []struct {
		field string
		got   []string
		want  []string
	}{
		{"Languages", names(len(a.TechStack.Languages), func(i int) string { return a.TechStack.Languages[i].Name }),
			[]string{"Go", "TypeScript", "Python"}},
		{"Frameworks", names(len(a.TechStack.Frameworks), func(i int) string { return a.TechStack.Frameworks[i].Name }),
			[]string{"Echo", "Gin", "React", "Vitest"}},
		{"Databases", a.TechStack.Databases, []string{"MongoDB", "PostgreSQL", "Redis"}},
		{"Tools", a.TechStack.Tools, []string{"Docker", "GitHub Actions", "Make"}},
		{"Directories", names(len(a.Structure.Directories), func(i int) string { return a.Structure.Directories[i].Path }),
			[]string{"api", "api/handlers", "web"}},
		{"RootFiles", a.Structure.RootFiles, []string{".gitignore", "Makefile", "README.md", "go.mod"}},
		{"Dependencies", names(len(a.Dependencies), func(i int) string { return a.Dependencies[i].Type + ":" + a.Dependencies[i].Name }),
			[]string{"dev:axios", "dev:eslint", "dev:vitest", "runtime:axios", "runtime:react"}},
		{"Endpoints", names(len(a.Endpoints), func(i int) string { return a.Endpoints[i].Method + " " + a.Endpoints[i].Path }),
			[]string{"GET /orders", "GET /users", "DELETE /users/:id", "POST /users"}},
		{"Testing", names(len(a.CodePatterns.Testing), func(i int) string { return a.CodePatterns.Testing[i].Name }),
			[]string{"func Test", "t.Error", "t.Run", "httptest"}},
		{"DataFetching", names(len(a.CodePatterns.DataFetching), func(i int) string { return a.CodePatterns.DataFetching[i].Name }),
			[]string{"fetch", "useEffect", "axios"}},
		{"MLPatterns", names(len(a.CodePatterns.MLPatterns), func(i int) string { return a.CodePatterns.MLPatterns[i].Usage }),
			[]string{"inference", "training"}},
		// ordered at the source, so normalize leaves these alone
		{"Commands", names(len(a.Commands), func(i int) string { return a.Commands[i].Name }),
			[]string{"make test", "make build"}},
		{"Conventions", names(len(a.Conventions), func(i int) string { return a.Conventions[i].Category }),
			[]string{"testing", "code-style"}},
		{"KeyFiles", names(len(a.KeyFiles), func(i int) string { return a.KeyFiles[i].Path }),
			[]string{"main.go", "go.mod"}},
		{"Layers", names(len(a.ArchitectureInfo.Layers), func(i int) string { return a.ArchitectureInfo.Layers[i].Name }),
			[]string{"internal", "cmd"}},
	}
	for _, c := range checks {
		if !reflect.DeepEqual(c.got, c.want) {
			t.Errorf("%s = %q, want %q", c.field, c.got, c.want)
		}
	}
}

func TestNormalizeNil(t *testing.T) {
	normalize(nil)
	a := &types.Analysis{}
	normalize(a)
	if !reflect.DeepEqual(a, &types.Analysis{}) {
		t.Errorf("normalize changed an empty analysis: %+v", a)
	}
}
