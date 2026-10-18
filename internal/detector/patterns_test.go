package detector

import (
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// winston and pino share their call pattern, so each needs its declared
// dependency, and a tie between evidenced rows goes to the earlier row.
func TestLoggingLibrariesNeedEvidence(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		line string
		deps []string
		want string
	}{
		{"winston declared", ".ts", "logger.info('started');", []string{"winston"}, "Winston logger with structured logging"},
		{"pino declared", ".ts", "logger.info('started');", []string{"pino"}, "Pino logger (fast JSON logging)"},
		{"both declared", ".ts", "logger.info('started');", []string{"pino", "winston"}, "Winston logger with structured logging"},
		{"neither declared", ".ts", "logger.info('started');", nil, ""},
		{"console needs none", ".ts", "console.log('started');", nil, "console.log/warn/error for logging"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, files := writeProject(t, map[string]string{
				"a" + tt.ext: tt.line, "b" + tt.ext: tt.line, "c" + tt.ext: tt.line,
			})
			var deps []types.Dependency
			for _, name := range tt.deps {
				deps = append(deps, types.Dependency{Name: name, Version: "1.0.0", Type: "runtime"})
			}

			conventions := stableResult(t, func() []types.Convention {
				return NewPatternDetector(root, files, deps).detectLoggingPatterns()
			})

			var got string
			if len(conventions) == 1 {
				got = conventions[0].Description
			}
			if len(conventions) > 1 || got != tt.want {
				t.Errorf("logging conventions = %+v, want %q", conventions, tt.want)
			}
		})
	}
}
