package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"v0.5.0", "0.4.0", true},
		{"v0.5.0", "0.5.0", false},
		{"0.5.0", "v0.5.0", false},
		{"v0.10.0", "0.9.0", true},
		{"v0.4.0", "0.5.0", false},      // a newer local build never downgrades
		{"v0.5.0-rc.1", "0.5.0", false}, // a prerelease is older than its release
		{"v0.5.0", "0.5.0-rc.1", true},
		{"v0.5.0", "dev", true},
		{"v0.5.0", "", true},
		{"nightly", "0.4.0", false},
		{"nightly", "dev", false},
		{"v0.5.0", "0.4.1-0.20261018160112-850ef45a1376", true},
		{"v0.4.1", "0.4.1-0.20261018160112-850ef45a1376+dirty", true},
		{"v0.4.0", "0.4.1-0.20261018160112-850ef45a1376", false},
	}
	for _, tt := range tests {
		if got := isNewer(tt.latest, tt.current); got != tt.want {
			t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

// serveRelease points the upgrade requests at handler for one test.
func serveRelease(t *testing.T, timeout time.Duration, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	oldURL, oldClient := releaseURL, httpClient
	releaseURL, httpClient = srv.URL, &http.Client{Timeout: timeout}
	t.Cleanup(func() {
		releaseURL, httpClient = oldURL, oldClient
		srv.Close()
	})
	return srv
}

func TestGetLatestRelease(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		wantTag string
		wantErr string
	}{
		{"ok", http.StatusOK, `{"tag_name": "v0.5.0", "assets": [{"name": "argus_0.5.0_linux_amd64.tar.gz", "browser_download_url": "https://example.com/a"}]}`, "v0.5.0", ""},
		{"not found", http.StatusNotFound, `{"message": "Not Found"}`, "", "GitHub API returned status 404"},
		{"malformed JSON", http.StatusOK, `{"tag_name": `, "", "unexpected EOF"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serveRelease(t, 5*time.Second, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = fmt.Fprint(w, tt.body)
			})

			release, err := getLatestRelease()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if release.TagName != tt.wantTag || len(release.Assets) != 1 {
				t.Errorf("release = %+v, want tag %s with one asset", release, tt.wantTag)
			}
		})
	}
}

func TestGetLatestReleaseTimesOut(t *testing.T) {
	serveRelease(t, 100*time.Millisecond, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done() // stall until the client gives up
	})

	start := time.Now()
	if _, err := getLatestRelease(); err == nil {
		t.Fatal("a stalled request succeeded")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("stalled request took %v", elapsed)
	}
}

func TestDownloadAsset(t *testing.T) {
	srv := serveRelease(t, 5*time.Second, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/asset" {
			http.NotFound(w, r)
			return
		}
		_, _ = fmt.Fprint(w, "archive bytes")
	})

	if _, err := downloadAsset(srv.URL + "/missing"); err == nil || !strings.Contains(err.Error(), "download failed with status 404") {
		t.Errorf("missing asset: err = %v", err)
	}

	path, err := downloadAsset(srv.URL + "/asset")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })
	if data, err := os.ReadFile(path); err != nil || string(data) != "archive bytes" {
		t.Errorf("downloaded %q, %v", data, err)
	}
}

// installFixture writes an old target binary and a new one in separate
// directories and returns their paths.
func installFixture(t *testing.T) (target, newBinary string) {
	t.Helper()
	target = filepath.Join(t.TempDir(), "argus")
	newBinary = filepath.Join(t.TempDir(), "argus-new")
	if err := os.WriteFile(target, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newBinary, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	return target, newBinary
}

// checkInstall fails unless target holds want and its directory holds
// nothing else.
func checkInstall(t *testing.T, target, want string) {
	t.Helper()
	if data, err := os.ReadFile(target); err != nil || string(data) != want {
		t.Errorf("target = %q, %v; want %q", data, err, want)
	}
	entries, err := os.ReadDir(filepath.Dir(target))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if runtime.GOOS == "windows" {
		names = slices.DeleteFunc(names, func(n string) bool { return n == "argus.old" })
	}
	if !slices.Equal(names, []string{"argus"}) {
		t.Errorf("directory holds %v, want only the target", names)
	}
}

func TestReplaceBinary(t *testing.T) {
	target, newBinary := installFixture(t)
	if err := replaceBinary(newBinary, target); err != nil {
		t.Fatal(err)
	}
	checkInstall(t, target, "new")
	if info, err := os.Stat(target); err != nil || (runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0) {
		t.Errorf("target is not executable: %v", err)
	}
}

func TestReplaceBinaryMissingNewBinary(t *testing.T) {
	target, _ := installFixture(t)
	if err := replaceBinary(filepath.Join(t.TempDir(), "missing"), target); err == nil {
		t.Fatal("installing a missing binary succeeded")
	}
	checkInstall(t, target, "old")
}

func TestReplaceBinaryUnwritableDirectory(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("needs a directory the test user cannot write")
	}
	target, newBinary := installFixture(t)
	dir := filepath.Dir(target)
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	if err := replaceBinary(newBinary, target); err == nil {
		t.Fatal("installing into a read-only directory succeeded")
	}
	checkInstall(t, target, "old")
}

// TestReplaceBinaryWhileRunning replaces a copy of the test binary while it
// runs; writing it in place fails with ETXTBSY on Linux.
func TestReplaceBinaryWhileRunning(t *testing.T) {
	if os.Getenv("ARGUS_TEST_SLEEP") == "1" {
		time.Sleep(time.Minute) // the parent kills this process
		return
	}
	if runtime.GOOS == "windows" {
		t.Skip("Windows moves the running binary aside instead")
	}
	self, err := os.Executable()
	if err != nil {
		t.Skip(err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatal(err)
	}
	target, newBinary := installFixture(t)
	if err := os.WriteFile(target, data, 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(target, "-test.run=^TestReplaceBinaryWhileRunning$")
	cmd.Env = append(os.Environ(), "ARGUS_TEST_SLEEP=1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })

	if err := replaceBinary(newBinary, target); err != nil {
		t.Fatalf("replacing a running binary: %v", err)
	}
	checkInstall(t, target, "new")
}
