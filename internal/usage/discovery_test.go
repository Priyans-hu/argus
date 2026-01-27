package usage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEncodeProjectPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/Users/foo/bar", "-Users-foo-bar"},
		{"/Users/priyanshu/markLab/argus", "-Users-priyanshu-markLab-argus"},
		{"/Users/priyanshu.garg/markLab/argus", "-Users-priyanshu-garg-markLab-argus"},
		{"/home/user/project", "-home-user-project"},
	}

	for _, tt := range tests {
		result := EncodeProjectPath(tt.input)
		if result != tt.expected {
			t.Errorf("EncodeProjectPath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDiscoverSessions(t *testing.T) {
	// Create temp dir with mock session files
	tmpDir := t.TempDir()

	// Create session files
	files := []struct {
		name    string
		modTime time.Time
	}{
		{"abc-123.jsonl", time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)},
		{"def-456.jsonl", time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC)},
		{"ghi-789.jsonl", time.Date(2026, 1, 25, 10, 0, 0, 0, time.UTC)},
		{"not-a-session.txt", time.Date(2026, 1, 25, 10, 0, 0, 0, time.UTC)},
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)
		if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, f.modTime, f.modTime); err != nil {
			t.Fatal(err)
		}
	}

	// Create a subdirectory (should be skipped)
	if err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	// Test without date filter
	sessions, err := DiscoverSessions(tmpDir, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions without filter, got %d", len(sessions))
	}

	// Test with date filter (only files after Jan 18)
	since := time.Date(2026, 1, 18, 0, 0, 0, 0, time.UTC)
	sessions, err = DiscoverSessions(tmpDir, since)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions after filter, got %d", len(sessions))
	}
}

func TestDiscoverSubagents(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subagents directory
	sessionID := "abc-123"
	subagentDir := filepath.Join(tmpDir, sessionID, "subagents")
	if err := os.MkdirAll(subagentDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create subagent files
	for _, name := range []string{"agent-a1.jsonl", "agent-b2.jsonl"} {
		if err := os.WriteFile(filepath.Join(subagentDir, name), []byte("{}"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := DiscoverSubagents(tmpDir, sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 subagent files, got %d", len(files))
	}

	// Non-existent session should return nil
	files, err = DiscoverSubagents(tmpDir, "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if files != nil {
		t.Errorf("expected nil for nonexistent session, got %v", files)
	}
}
