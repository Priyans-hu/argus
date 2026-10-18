package detector

import (
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestPythonASTUsesMostSpecificMatch(t *testing.T) {
	root, files := writeProject(t, map[string]string{
		"app.py": `import functools
import google.cloud.storage
from fastapi import FastAPI


class Repo:
    @functools.cached_property
    def items(self):
        return []


@pydantic.field_validator("name")
def check(cls, value):
    return value
`,
	})

	patterns := stableResult(t, func() []types.PatternInfo { return NewPythonASTDetector(root, files).Detect() })

	got := map[string]string{} // name -> description
	for _, p := range patterns {
		if _, seen := got[p.Name]; !seen {
			got[p.Name] = p.Description
		}
	}
	want := map[string]string{
		"@functools.cached_property": "Cached property",
		"@pydantic.field_validator":  "Pydantic field validator",
		"FastAPI":                    "Detected via import: fastapi",
		"Google Cloud":               "Detected via import: google.cloud.storage",
	}
	for name, desc := range want {
		if got[name] != desc {
			t.Errorf("%s: description = %q, want %q", name, got[name], desc)
		}
	}
}
