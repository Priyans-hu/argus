package generator

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Priyans-hu/argus/pkg/types"
)

// formatHookCommand renders settings.json for one language and format
// command and returns the PostToolUse formatter command.
func formatHookCommand(t *testing.T, lang, format string) string {
	t.Helper()
	analysis := &types.Analysis{
		TechStack: types.TechStack{Languages: []types.Language{{Name: lang}}},
		Commands:  []types.Command{{Name: "format", Command: format}},
	}
	content, ok := renderClaudeCode(t, analysis)[".claude/settings.json"]
	if !ok {
		t.Fatal("no .claude/settings.json generated")
	}
	if strings.Contains(content, `\u00`) || strings.Contains(content, "FILE_PATH") {
		t.Errorf("settings.json escapes shell characters or uses $FILE_PATH:\n%s", content)
	}
	if !strings.HasSuffix(content, "}\n") {
		t.Errorf("settings.json lacks a final newline:\n%s", content)
	}
	var settings SettingsJSON
	if err := json.Unmarshal([]byte(content), &settings); err != nil {
		t.Fatal(err)
	}
	post := settings.Hooks["PostToolUse"]
	if len(settings.Hooks) != 1 || len(post) != 1 || post[0].Matcher != "Edit|Write" || len(post[0].Hooks) != 1 || post[0].Hooks[0].Timeout != 30 {
		t.Fatalf("unexpected hooks: %+v", settings.Hooks)
	}
	return post[0].Hooks[0].Command
}

func TestFormatHookCommand(t *testing.T) {
	want := `command -v jq >/dev/null 2>&1 && command -v gofmt >/dev/null 2>&1 || exit 0; ` +
		`f=$(jq -r '.tool_input.file_path // empty'); case "$f" in *.go) gofmt -w "$f";; esac`
	if got := formatHookCommand(t, "Go", "gofmt -l ."); got != want {
		t.Errorf("gofmt hook:\n got %s\nwant %s", got, want)
	}
}

// runHook runs a hook command the way Claude Code does, through sh with the
// PostToolUse payload on stdin. PATH holds only jq (when withJQ is set) and
// the stub formatters named in stubs; each stub logs its name and arguments,
// one per line, to the returned string.
func runHook(t *testing.T, command, payload string, withJQ bool, stubs ...string) string {
	t.Helper()
	sh, err := exec.LookPath("sh")
	if err != nil {
		t.Skip("no sh")
	}
	bin := t.TempDir()
	if withJQ {
		jq, err := exec.LookPath("jq")
		if err != nil {
			t.Skip("jq is not installed")
		}
		if err := os.Symlink(jq, filepath.Join(bin, "jq")); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range stubs {
		stub := "#!" + sh + "\nprintf '%s\\n' \"${0##*/}\" \"$@\" >> \"$LOG\"\n"
		if err := os.WriteFile(filepath.Join(bin, name), []byte(stub), 0755); err != nil {
			t.Fatal(err)
		}
	}
	log := filepath.Join(t.TempDir(), "calls")

	cmd := exec.Command(sh, "-c", command)
	cmd.Env = []string{"PATH=" + bin, "LOG=" + log}
	cmd.Stdin = strings.NewReader(payload)
	if out, err := cmd.CombinedOutput(); err != nil || len(out) != 0 {
		t.Fatalf("hook exited with %v, output %q", err, out)
	}
	calls, err := os.ReadFile(log)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(calls)
}

// editPayload is a PostToolUse stdin payload for an Edit of file.
func editPayload(t *testing.T, file string) string {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"session_id":      "abc123",
		"hook_event_name": "PostToolUse",
		"tool_name":       "Edit",
		"tool_input":      map[string]any{"file_path": file, "old_string": "a", "new_string": "b"},
		"tool_response":   map[string]any{"filePath": file, "success": true},
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(payload)
}

func TestFormatHookReadsPathFromStdin(t *testing.T) {
	tests := []struct {
		name, lang, format, file, want string
		withJQ                         bool
		stub                           string
	}{
		{"gofmt", "Go", "gofmt -l .", "/repo/my dir/main.go", "gofmt\n-w\n/repo/my dir/main.go\n", true, "gofmt"},
		{"gofmt other file", "Go", "gofmt -l .", "/repo/README.md", "", true, "gofmt"},
		{"prettier", "TypeScript", "prettier --write .", "/repo/src/app.tsx", "npx\nprettier\n--write\n/repo/src/app.tsx\n", true, "npx"},
		{"prettier other file", "JavaScript", "prettier --write .", "/repo/styles.css", "", true, "npx"},
		{"black", "Python", "black .", "/repo/app.py", "black\n/repo/app.py\n", true, "black"},
		{"ruff", "Python", "ruff format .", "/repo/app.py", "ruff\nformat\n/repo/app.py\n", true, "ruff"},
		{"ruff other file", "Python", "ruff format .", "/repo/pyproject.toml", "", true, "ruff"},
		{"no file_path", "Go", "gofmt -l .", "", "", true, "gofmt"},
		{"no formatter", "Go", "gofmt -l .", "/repo/main.go", "", true, ""},
		{"no jq", "Go", "gofmt -l .", "/repo/main.go", "", false, "gofmt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := "{}"
			if tt.file != "" {
				payload = editPayload(t, tt.file)
			}
			var stubs []string
			if tt.stub != "" {
				stubs = append(stubs, tt.stub)
			}
			got := runHook(t, formatHookCommand(t, tt.lang, tt.format), payload, tt.withJQ, stubs...)
			if got != tt.want {
				t.Errorf("formatter calls = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNoBashGuardHook(t *testing.T) {
	analysis := &types.Analysis{
		TechStack: types.TechStack{Languages: []types.Language{{Name: "Go"}}},
		CodePatterns: &types.CodePatterns{Authentication: []types.PatternInfo{
			{Name: "jwt.", FileCount: 2, Examples: []string{"api/auth.go"}},
		}},
	}
	if content, ok := renderClaudeCode(t, analysis)[".claude/settings.json"]; ok {
		t.Errorf("settings.json generated without a formatter:\n%s", content)
	}
	analysis.Commands = []types.Command{{Name: "fmt", Command: "gofmt -l ."}}
	if content := renderClaudeCode(t, analysis)[".claude/settings.json"]; !strings.Contains(content, "PostToolUse") || strings.Contains(content, "PreToolUse") {
		t.Errorf("want only the formatter hook:\n%s", content)
	}
}
