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
		{"loguru declared", ".py", "from loguru import logger\n\nlogger.info('started')\n", []string{"loguru"}, "Loguru logging"},
		{"loguru import alone", ".py", "from loguru import logger\n\nlogger.info('started')\n", nil, "Loguru logging"},
		{"logger. without an import", ".py", "logger.info('started')\n", nil, ""},
		{"logger. with logging imported", ".py", "import logging\n\nlogger = logging.getLogger(__name__)\nlogger.info('started')\n", nil, "Python logging module"},
		{"rust macros without log or tracing", ".rs", "fn main() { info!(\"started\"); }\n", nil, ""},
		{"rust macros with tracing", ".rs", "fn main() { info!(\"started\"); }\n", []string{"tracing"}, "Rust log/tracing macros"},
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
