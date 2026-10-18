package detector

import (
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestLoggingTiesGoToTheEarlierLogger(t *testing.T) {
	tests := []struct {
		name string
		ext  string
		line string
		want string
	}{
		// the winston and pino patterns match the same calls
		{"winston over pino", ".ts", "logger.info('started');", "Winston logger with structured logging"},
		// zap's pattern also matches zerolog's chained calls
		{"zap over zerolog", ".go", `logger.Info().Msg("started")`, "Uber's Zap logger"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, files := writeProject(t, map[string]string{
				"a" + tt.ext: tt.line, "b" + tt.ext: tt.line, "c" + tt.ext: tt.line,
			})

			conventions := stableResult(t, func() []types.Convention {
				return NewPatternDetector(root, files).detectLoggingPatterns()
			})

			if len(conventions) != 1 || conventions[0].Description != tt.want {
				t.Errorf("logging conventions = %+v, want %q", conventions, tt.want)
			}
		})
	}
}
