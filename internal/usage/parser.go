package usage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxScanBufferSize = 4 * 1024 * 1024 // 4MB per line

// ParseOptions controls parsing behavior
type ParseOptions struct {
	Since time.Time // Skip entries before this time
}

// SessionData holds parsed data from a single JSONL session file
type SessionData struct {
	ID         string
	ToolEvents []ToolEvent
	TokenStats []TokenRecord
	TurnCount  int
	StartTime  time.Time
	EndTime    time.Time
}

// ToolEvent represents a single tool invocation extracted from assistant content
type ToolEvent struct {
	Name     string // Read, Edit, Bash, Write, Glob, Grep, Task, etc.
	FilePath string // extracted from input.file_path (empty for non-file tools)
	Model    string
}

// TokenRecord holds token counts from a single assistant message
type TokenRecord struct {
	Model               string
	InputTokens         int64
	OutputTokens        int64
	CacheCreationTokens int64
	CacheReadTokens     int64
}

// jsonlEntry is the top-level structure of each JSONL line
type jsonlEntry struct {
	Type      string          `json:"type"`
	Timestamp string          `json:"timestamp"`
	SessionID string          `json:"sessionId"`
	Message   json.RawMessage `json:"message"`
}

// assistantMessage holds the fields we extract from assistant messages
type assistantMessage struct {
	Model   string         `json:"model"`
	Usage   tokenUsage     `json:"usage"`
	Content []contentBlock `json:"content"`
}

type tokenUsage struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// toolInput captures the file_path field from tool inputs
type toolInput struct {
	FilePath string `json:"file_path"`
	Path     string `json:"path"`    // some tools use path instead
	Pattern  string `json:"pattern"` // Glob/Grep
	Command  string `json:"command"` // Bash
}

// ParseSession parses a single JSONL session file and extracts usage data.
// It streams the file line-by-line to handle large files efficiently.
func ParseSession(ctx context.Context, path string, opts ParseOptions) (*SessionData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	session := &SessionData{
		ID: sessionIDFromPath(path),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, maxScanBufferSize), maxScanBufferSize)

	var firstTime, lastTime time.Time

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry jsonlEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue // skip malformed lines
		}

		// Skip non-essential entry types early
		switch entry.Type {
		case "assistant":
			// process below
		case "user":
			// track timestamp only
			if ts := parseTimestamp(entry.Timestamp); !ts.IsZero() {
				if firstTime.IsZero() || ts.Before(firstTime) {
					firstTime = ts
				}
				if ts.After(lastTime) {
					lastTime = ts
				}
			}
			continue
		default:
			continue
		}

		// Parse timestamp and apply date filter
		ts := parseTimestamp(entry.Timestamp)
		if ts.IsZero() {
			continue
		}
		if !opts.Since.IsZero() && ts.Before(opts.Since) {
			continue
		}

		if firstTime.IsZero() || ts.Before(firstTime) {
			firstTime = ts
		}
		if ts.After(lastTime) {
			lastTime = ts
		}

		// Parse assistant message
		if len(entry.Message) == 0 {
			continue
		}

		var msg assistantMessage
		if err := json.Unmarshal(entry.Message, &msg); err != nil {
			continue
		}

		// Extract token stats
		if msg.Usage.InputTokens > 0 || msg.Usage.OutputTokens > 0 {
			session.TokenStats = append(session.TokenStats, TokenRecord{
				Model:               normalizeModel(msg.Model),
				InputTokens:         msg.Usage.InputTokens,
				OutputTokens:        msg.Usage.OutputTokens,
				CacheCreationTokens: msg.Usage.CacheCreationInputTokens,
				CacheReadTokens:     msg.Usage.CacheReadInputTokens,
			})
			session.TurnCount++
		}

		// Extract tool usage from content blocks
		for _, block := range msg.Content {
			if block.Type != "tool_use" || block.Name == "" {
				continue
			}

			event := ToolEvent{
				Name:  block.Name,
				Model: normalizeModel(msg.Model),
			}

			// Extract file path from tool input
			if len(block.Input) > 0 {
				var ti toolInput
				if err := json.Unmarshal(block.Input, &ti); err == nil {
					if ti.FilePath != "" {
						event.FilePath = ti.FilePath
					} else if ti.Path != "" {
						event.FilePath = ti.Path
					}
				}
			}

			session.ToolEvents = append(session.ToolEvents, event)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	session.StartTime = firstTime
	session.EndTime = lastTime

	return session, nil
}

// sessionIDFromPath extracts session UUID from file path
func sessionIDFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, ".jsonl")
}

// parseTimestamp parses an ISO-8601 timestamp string
func parseTimestamp(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return time.Time{}
		}
	}
	return t
}

// normalizeModel simplifies model names for grouping
// e.g., "claude-opus-4-5-20251101" → "claude-opus-4-5"
func normalizeModel(model string) string {
	if model == "" {
		return "unknown"
	}

	// Remove date suffixes like -20251101
	parts := strings.Split(model, "-")
	if len(parts) > 1 {
		last := parts[len(parts)-1]
		if len(last) == 8 && isAllDigits(last) {
			parts = parts[:len(parts)-1]
		}
	}

	return strings.Join(parts, "-")
}

func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
