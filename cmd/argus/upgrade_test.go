package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
