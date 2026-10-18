package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		ldflag, mainVersion, want string
	}{
		{"0.5.0", "v0.4.0", "0.5.0"},
		{"dev", "v0.5.0", "0.5.0"},
		{"", "v0.5.0", "0.5.0"},
		{"dev", "(devel)", "dev"},
		{"dev", "", "dev"},
		{"dev", "v0.4.1-0.20261018160112-850ef45a1376", "0.4.1-0.20261018160112-850ef45a1376"},
	}
	for _, tt := range tests {
		if got := resolveVersion(tt.ldflag, tt.mainVersion); got != tt.want {
			t.Errorf("resolveVersion(%q, %q) = %q, want %q", tt.ldflag, tt.mainVersion, got, tt.want)
		}
	}
}

func TestVersionFlagMatchesVersionCommand(t *testing.T) {
	want := "argus version " + version + "\n"
	for _, arg := range []string{"--version", "-v", "version"} {
		t.Run(arg, func(t *testing.T) {
			out, errOut, err := executeRoot(t, arg)
			if err != nil {
				t.Fatalf("argus %s: %v", arg, err)
			}
			if out != want || errOut != "" {
				t.Errorf("argus %s printed %q (stderr %q), want %q", arg, out, errOut, want)
			}
		})
	}
}

func TestScanShortVIsVerbose(t *testing.T) {
	resetFlags(t)
	restoreFlags(t, scanCmd, "verbose", "dry-run")
	if err := scanCmd.ParseFlags([]string{"-v", "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	if !verbose || !dryRun {
		t.Errorf("scan -v --dry-run parsed verbose=%v dryRun=%v, want both true", verbose, dryRun)
	}
}

func TestOutputFlagIsHiddenAndDeprecated(t *testing.T) {
	resetFlags(t)
	help, _, err := executeRoot(t, "scan", "--help")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(help, "--output") || strings.Contains(help, "-o,") {
		t.Errorf("scan --help lists -o/--output:\n%s", help)
	}

	var warnings bytes.Buffer
	rootCmd.SetOut(&warnings) // cobra prints flag warnings to stderr unless an output is set
	t.Cleanup(resetRoot)
	restoreFlags(t, scanCmd, "output", "dry-run")
	if err := scanCmd.ParseFlags([]string{"-o", "x", "--dry-run"}); err != nil {
		t.Fatalf("scan -o x --dry-run: %v", err)
	}
	// cobra keeps earlier warnings in its buffer when the test runs again in
	// the same process, so check for the line rather than the whole output
	want := "Flag --output has been deprecated, it has no effect and will be removed in a future release\n"
	if !strings.HasSuffix(warnings.String(), want) {
		t.Errorf("warnings = %q, want them to end with %q", warnings.String(), want)
	}
	if outputDir != "x" || !dryRun {
		t.Errorf("parsed outputDir=%q dryRun=%v, want x and true", outputDir, dryRun)
	}
}
