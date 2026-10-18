package main

import "testing"

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
	if err := scanCmd.ParseFlags([]string{"-v", "--dry-run"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, name := range []string{"verbose", "dry-run"} {
			f := scanCmd.Flags().Lookup(name)
			_ = f.Value.Set(f.DefValue)
			f.Changed = false
		}
	})
	if !verbose || !dryRun {
		t.Errorf("scan -v --dry-run parsed verbose=%v dryRun=%v, want both true", verbose, dryRun)
	}
}
