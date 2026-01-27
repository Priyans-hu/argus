package usage

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestParseSession_Basic(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join("testdata", "session-basic.jsonl")

	session, err := ParseSession(ctx, path, ParseOptions{})
	if err != nil {
		t.Fatalf("ParseSession failed: %v", err)
	}

	if session.ID != "session-basic" {
		t.Errorf("expected session ID 'session-basic', got %q", session.ID)
	}

	// Should have 4 assistant turns with tokens
	if session.TurnCount != 4 {
		t.Errorf("expected 4 turns, got %d", session.TurnCount)
	}

	// Should have 4 tool events: Read, Edit, Bash, Read
	if len(session.ToolEvents) != 4 {
		t.Fatalf("expected 4 tool events, got %d", len(session.ToolEvents))
	}

	// Verify tool names
	expectedTools := []string{"Read", "Edit", "Bash", "Read"}
	for i, expected := range expectedTools {
		if session.ToolEvents[i].Name != expected {
			t.Errorf("tool %d: expected %q, got %q", i, expected, session.ToolEvents[i].Name)
		}
	}

	// Verify file paths extracted
	if session.ToolEvents[0].FilePath != "/Users/test/project/cmd/main.go" {
		t.Errorf("expected file path for Read, got %q", session.ToolEvents[0].FilePath)
	}
	if session.ToolEvents[2].FilePath != "" {
		t.Errorf("Bash should have empty file path, got %q", session.ToolEvents[2].FilePath)
	}

	// Verify token stats
	if len(session.TokenStats) != 4 {
		t.Fatalf("expected 4 token records, got %d", len(session.TokenStats))
	}

	totalInput := int64(0)
	for _, ts := range session.TokenStats {
		totalInput += ts.InputTokens
	}
	if totalInput != 5300 { // 1000 + 2000 + 1500 + 800
		t.Errorf("expected total input tokens 5300, got %d", totalInput)
	}

	// Verify timestamps
	if session.StartTime.IsZero() {
		t.Error("start time should not be zero")
	}
	if session.EndTime.IsZero() {
		t.Error("end time should not be zero")
	}
}

func TestParseSession_MultiTool(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join("testdata", "session-multi-tool.jsonl")

	session, err := ParseSession(ctx, path, ParseOptions{})
	if err != nil {
		t.Fatalf("ParseSession failed: %v", err)
	}

	// 3 assistant messages, but 2 with opus usage + 1 with haiku
	if session.TurnCount != 3 {
		t.Errorf("expected 3 turns, got %d", session.TurnCount)
	}

	// Tool events: Glob, Grep, Read, Write, Edit, Bash = 6
	if len(session.ToolEvents) != 6 {
		t.Fatalf("expected 6 tool events, got %d", len(session.ToolEvents))
	}

	// Verify model normalization
	if session.ToolEvents[0].Model != "claude-opus-4-5" {
		t.Errorf("expected normalized model 'claude-opus-4-5', got %q", session.ToolEvents[0].Model)
	}
	if session.ToolEvents[5].Model != "claude-haiku-4-5" {
		t.Errorf("expected normalized model 'claude-haiku-4-5', got %q", session.ToolEvents[5].Model)
	}
}

func TestParseSession_DateFilter(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join("testdata", "session-basic.jsonl")

	// Filter to only include entries after the last one
	since := time.Date(2026, 1, 15, 10, 0, 22, 0, time.UTC)

	session, err := ParseSession(ctx, path, ParseOptions{Since: since})
	if err != nil {
		t.Fatalf("ParseSession failed: %v", err)
	}

	// Only the entries at 10:00:25 should pass, but that's a system entry (skipped)
	// So we should have 0 turns
	if session.TurnCount != 0 {
		t.Errorf("expected 0 turns after date filter, got %d", session.TurnCount)
	}
}

func TestNormalizeModel(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"claude-opus-4-5-20251101", "claude-opus-4-5"},
		{"claude-sonnet-4-5-20251101", "claude-sonnet-4-5"},
		{"claude-haiku-4-5-20251001", "claude-haiku-4-5"},
		{"claude-opus-4-5", "claude-opus-4-5"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		result := normalizeModel(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeModel(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSessionIDFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/.claude/projects/foo/abc-123.jsonl", "abc-123"},
		{"session.jsonl", "session"},
		{"/path/to/3dda144e-5c30-481a-a89a-a667d137b922.jsonl", "3dda144e-5c30-481a-a89a-a667d137b922"},
	}

	for _, tt := range tests {
		result := sessionIDFromPath(tt.path)
		if result != tt.expected {
			t.Errorf("sessionIDFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
		}
	}
}
