package analyzer

import (
	"context"
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/internal/detector"
	"github.com/Priyans-hu/argus/pkg/types"
)

func TestDependenciesMatchAcrossAnalyzers(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"package.json":         `{"dependencies": {"react": "18.2.0"}, "peerDependencies": {"react": ">=18"}}`,
		"web/package.json":     `{"dependencies": {"next": "14.0.0", "react": "17.0.0"}}`,
		"src-tauri/Cargo.toml": "[dependencies]\ntauri = \"2\"\n",
		"go.mod":               "module example.com/m\n\ngo 1.24\n\nrequire github.com/spf13/cobra v1.8.0\n",
	})

	want := []types.Dependency{
		{Name: "react", Version: ">=18", Type: "peer"},
		{Name: "github.com/spf13/cobra", Version: "v1.8.0", Type: "runtime"},
		{Name: "next", Version: "14.0.0", Type: "runtime"},
		{Name: "react", Version: "18.2.0", Type: "runtime"},
		{Name: "tauri", Version: "2", Type: "runtime"},
	}
	ctx := context.Background()
	sequential, err := NewAnalyzer(root, nil).Analyze(ctx)
	if err != nil {
		t.Fatal(err)
	}
	parallel, err := NewParallelAnalyzer(root, nil).Analyze(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(sequential.Dependencies, want) {
		t.Errorf("sequential dependencies = %+v, want %+v", sequential.Dependencies, want)
	}
	if !reflect.DeepEqual(parallel.Dependencies, want) || !reflect.DeepEqual(parallel.TechStack, sequential.TechStack) {
		t.Errorf("parallel = %+v %+v, sequential = %+v %+v", parallel.Dependencies, parallel.TechStack, sequential.Dependencies, sequential.TechStack)
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
