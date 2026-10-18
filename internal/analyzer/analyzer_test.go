package analyzer

import (
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
)

func TestDetectDependenciesAreSortedByName(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"package.json": `{"dependencies": {"react": "18", "next": "14", "axios": "1"}, "devDependencies": {"vitest": "1", "eslint": "8"}}`,
	})

	want := []types.Dependency{
		{Name: "axios", Version: "1", Type: "runtime"},
		{Name: "next", Version: "14", Type: "runtime"},
		{Name: "react", Version: "18", Type: "runtime"},
		{Name: "eslint", Version: "8", Type: "dev"},
		{Name: "vitest", Version: "1", Type: "dev"},
	}
	a := &Analyzer{rootPath: root}
	for i := range 50 {
		if got := a.detectDependencies(root); !reflect.DeepEqual(got, want) {
			t.Fatalf("run %d: dependencies = %+v, want %+v", i, got, want)
		}
	}
}

func TestPyProjectScriptsAreSortedByName(t *testing.T) {
	info := &detector.PyProjectInfo{Scripts: map[string]string{"serve": "app:serve", "cli": "app:main", "migrate": "app:migrate"}}

	for i := range 50 {
		var names []string
		for _, c := range detectPyProjectCommands(info) {
			names = append(names, c.Name)
		}
		if want := []string{"cli", "migrate", "serve"}; !reflect.DeepEqual(names, want) {
			t.Fatalf("run %d: commands = %v, want %v", i, names, want)
		}
	}
}
