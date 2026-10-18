package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestErrorsPrintOnce(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	empty := t.TempDir()

	tests := []struct {
		name  string
		args  []string
		want  string // the whole error report
		usage bool
	}{
		{"missing path", []string{"scan", missing}, "Error: path does not exist: " + missing + "\n", false},
		{"sync without config", []string{"sync", empty},
			"Error: no .argus.yaml found in " + empty + "\nRun 'argus init' first\n", false},
		{"too many arguments", []string{"scan", "a", "b"}, "Error: accepts at most 1 arg(s), received 2\n", true},
		{"unknown flag", []string{"scan", "--bogus"}, "Error: unknown flag: --bogus\n", true},
		{"unknown command", []string{"bogus"}, "Error: unknown command \"bogus\" for \"argus\"\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags(t)
			errOut, code := runCLI(t, tt.args...)

			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
			if n := strings.Count(errOut, "Error:"); n != 1 {
				t.Errorf("Error: printed %d times:\n%s", n, errOut)
			}
			if !strings.HasSuffix(errOut, tt.want) {
				t.Errorf("stderr ends with %q, want %q", errOut, tt.want)
			}
			wantUsage := 0
			if tt.usage {
				wantUsage = 1
			}
			if n := strings.Count(errOut, "Usage:"); n != wantUsage {
				t.Errorf("usage printed %d times, want %d:\n%s", n, wantUsage, errOut)
			}
		})
	}
}
