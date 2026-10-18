package detector

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestJSASTImportsUseLongestPrefix(t *testing.T) {
	root, files := writeProject(t, map[string]string{
		"src/app.js": `var router = require('next/router');
var nav = require('next/navigation');
var dom = require('react-dom');
var query = require('react-query');
function App() { useState(); useEffect(); useMemo(); }
`,
	})

	patterns := stableResult(t, func() []types.PatternInfo { return NewJSASTDetector(root, files).Detect() })

	frameworks := map[string]string{} // import path -> framework name
	var hooks []string
	for _, p := range patterns {
		switch p.Category {
		case "JavaScript Frameworks":
			frameworks[strings.TrimPrefix(p.Description, "Detected via import: ")] = p.Name
		case "React Hooks":
			hooks = append(hooks, p.Name)
		}
	}
	wantFrameworks := map[string]string{
		"next/router":     "Next.js Router",
		"next/navigation": "Next.js",
		"react-dom":       "React DOM",
		"react-query":     "React Query",
	}
	if !reflect.DeepEqual(frameworks, wantFrameworks) {
		t.Errorf("frameworks = %v, want %v", frameworks, wantFrameworks)
	}
	if want := []string{"useEffect", "useMemo", "useState"}; !reflect.DeepEqual(hooks, want) {
		t.Errorf("hooks = %v, want %v", hooks, want)
	}
}
