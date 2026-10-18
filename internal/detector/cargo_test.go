package detector

import (
	"reflect"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestCargoFeaturesAndDependenciesAreSorted(t *testing.T) {
	root, _ := writeProject(t, map[string]string{
		"Cargo.toml": `[package]
name = "demo"

[dependencies]
tokio = "1"
anyhow = "1"
serde = "1"

[dev-dependencies]
cargo-watch = "8"
cargo-nextest = "0.9"

[features]
tls = []
default = ["tls"]
metrics = []
`,
	})

	info := stableResult(t, func() *CargoInfo { return NewCargoDetector(root).Detect() })
	if want := []string{"default", "metrics", "tls"}; !reflect.DeepEqual(info.Features, want) {
		t.Errorf("Features = %v, want %v", info.Features, want)
	}
	if want := []string{"anyhow", "serde", "tokio"}; !reflect.DeepEqual(info.Dependencies, want) {
		t.Errorf("Dependencies = %v, want %v", info.Dependencies, want)
	}

	commands := stableResult(t, func() []types.Command { return NewCargoDetector(root).DetectCargoCommands() })
	var tail []string
	for _, c := range commands[len(commands)-3:] {
		tail = append(tail, c.Name)
	}
	want := []string{"cargo nextest run", "cargo watch -x run", "cargo build --features default,metrics,tls"}
	if !reflect.DeepEqual(tail, want) {
		t.Errorf("last commands = %v, want %v", tail, want)
	}
}
