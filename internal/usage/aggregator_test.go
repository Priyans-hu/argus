package usage

import (
	"testing"
	"time"

	"github.com/Priyans-hu/argus/pkg/types"
)

func TestAggregate_Basic(t *testing.T) {
	sessions := []*SessionData{
		{
			ID:        "session-1",
			TurnCount: 3,
			StartTime: time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
			ToolEvents: []ToolEvent{
				{Name: "Read", FilePath: "/project/cmd/main.go", Model: "claude-sonnet-4-5"},
				{Name: "Read", FilePath: "/project/cmd/main.go", Model: "claude-sonnet-4-5"},
				{Name: "Edit", FilePath: "/project/cmd/main.go", Model: "claude-sonnet-4-5"},
				{Name: "Bash", FilePath: "", Model: "claude-sonnet-4-5"},
				{Name: "Read", FilePath: "/project/pkg/types.go", Model: "claude-sonnet-4-5"},
			},
			TokenStats: []TokenRecord{
				{Model: "claude-sonnet-4-5", InputTokens: 1000, OutputTokens: 200, CacheCreationTokens: 5000, CacheReadTokens: 3000},
				{Model: "claude-sonnet-4-5", InputTokens: 2000, OutputTokens: 500, CacheCreationTokens: 0, CacheReadTokens: 8000},
				{Model: "claude-sonnet-4-5", InputTokens: 1500, OutputTokens: 300, CacheCreationTokens: 0, CacheReadTokens: 5000},
			},
		},
	}

	insights := Aggregate(sessions, "/project")
	if insights == nil {
		t.Fatal("Aggregate returned nil")
	}

	// Session count
	if insights.SessionCount != 1 {
		t.Errorf("expected 1 session, got %d", insights.SessionCount)
	}

	// Total turns
	if insights.TotalTurns != 3 {
		t.Errorf("expected 3 turns, got %d", insights.TotalTurns)
	}

	// Tool usage
	if len(insights.ToolUsage) != 3 { // Read, Edit, Bash
		t.Fatalf("expected 3 tool types, got %d", len(insights.ToolUsage))
	}
	// Read should be first (3 count)
	if insights.ToolUsage[0].Name != "Read" || insights.ToolUsage[0].Count != 3 {
		t.Errorf("expected Read with count 3 as top tool, got %s with %d", insights.ToolUsage[0].Name, insights.ToolUsage[0].Count)
	}

	// Hot files
	if len(insights.HotFiles) != 2 {
		t.Fatalf("expected 2 hot files, got %d", len(insights.HotFiles))
	}
	// cmd/main.go should be first (3 ops)
	if insights.HotFiles[0].Path != "cmd/main.go" {
		t.Errorf("expected 'cmd/main.go' as top hot file, got %q", insights.HotFiles[0].Path)
	}
	if insights.HotFiles[0].ReadCount != 2 {
		t.Errorf("expected 2 reads for main.go, got %d", insights.HotFiles[0].ReadCount)
	}
	if insights.HotFiles[0].EditCount != 1 {
		t.Errorf("expected 1 edit for main.go, got %d", insights.HotFiles[0].EditCount)
	}

	// Token totals
	if insights.TokenUsage.InputTokens != 4500 {
		t.Errorf("expected 4500 input tokens, got %d", insights.TokenUsage.InputTokens)
	}
	if insights.TokenUsage.OutputTokens != 1000 {
		t.Errorf("expected 1000 output tokens, got %d", insights.TokenUsage.OutputTokens)
	}

	// Cost estimate should be non-zero
	if insights.CostEstimate.TotalCost <= 0 {
		t.Error("expected non-zero total cost")
	}

	// Model breakdown
	if len(insights.ModelBreakdown) != 1 {
		t.Fatalf("expected 1 model in breakdown, got %d", len(insights.ModelBreakdown))
	}
	if insights.ModelBreakdown[0].Model != "claude-sonnet-4-5" {
		t.Errorf("expected model 'claude-sonnet-4-5', got %q", insights.ModelBreakdown[0].Model)
	}
}

func TestAggregate_PainPoints(t *testing.T) {
	// Create a session where a file is read more than 10 times
	var events []ToolEvent
	for range 12 {
		events = append(events, ToolEvent{Name: "Read", FilePath: "/project/hard-file.go", Model: "claude-sonnet-4-5"})
	}

	sessions := []*SessionData{
		{
			ID:         "session-pain",
			TurnCount:  12,
			StartTime:  time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC),
			EndTime:    time.Date(2026, 1, 20, 11, 0, 0, 0, time.UTC),
			ToolEvents: events,
			TokenStats: []TokenRecord{
				{Model: "claude-sonnet-4-5", InputTokens: 12000, OutputTokens: 2000},
			},
		},
	}

	insights := Aggregate(sessions, "/project")
	if insights == nil {
		t.Fatal("Aggregate returned nil")
	}

	// Should detect repeated reads pain point
	if len(insights.PainPoints) == 0 {
		t.Fatal("expected at least 1 pain point")
	}

	found := false
	for _, pp := range insights.PainPoints {
		if pp.File == "hard-file.go" && pp.Type == "repeated_reads" {
			found = true
			if pp.Count != 12 {
				t.Errorf("expected count 12, got %d", pp.Count)
			}
		}
	}
	if !found {
		t.Error("expected 'repeated_reads' pain point for hard-file.go")
	}
}

func TestAggregate_WriteThenEdit(t *testing.T) {
	sessions := []*SessionData{
		{
			ID:        "session-wte",
			TurnCount: 3,
			StartTime: time.Date(2026, 1, 20, 10, 0, 0, 0, time.UTC),
			EndTime:   time.Date(2026, 1, 20, 11, 0, 0, 0, time.UTC),
			ToolEvents: []ToolEvent{
				{Name: "Write", FilePath: "/project/new.go", Model: "claude-sonnet-4-5"},
				{Name: "Edit", FilePath: "/project/new.go", Model: "claude-sonnet-4-5"},
				{Name: "Edit", FilePath: "/project/new.go", Model: "claude-sonnet-4-5"},
			},
			TokenStats: []TokenRecord{
				{Model: "claude-sonnet-4-5", InputTokens: 3000, OutputTokens: 500},
			},
		},
	}

	insights := Aggregate(sessions, "/project")
	if insights == nil {
		t.Fatal("Aggregate returned nil")
	}

	found := false
	for _, pp := range insights.PainPoints {
		if pp.File == "new.go" && pp.Type == "write_then_edit" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'write_then_edit' pain point for new.go")
	}
}

func TestAggregate_Empty(t *testing.T) {
	insights := Aggregate(nil, "/project")
	if insights != nil {
		t.Error("expected nil for empty sessions")
	}

	insights = Aggregate([]*SessionData{}, "/project")
	if insights != nil {
		t.Error("expected nil for empty sessions slice")
	}
}

func TestMakeRelative(t *testing.T) {
	tests := []struct {
		filePath    string
		projectRoot string
		expected    string
	}{
		{"/Users/test/project/cmd/main.go", "/Users/test/project", "cmd/main.go"},
		{"/Users/test/project/pkg/types.go", "/Users/test/project", "pkg/types.go"},
		{"/other/path/file.go", "/Users/test/project", "/other/path/file.go"}, // outside project
		{"relative/path.go", "", "relative/path.go"},                          // no project root
	}

	for _, tt := range tests {
		result := makeRelative(tt.filePath, tt.projectRoot)
		if result != tt.expected {
			t.Errorf("makeRelative(%q, %q) = %q, want %q", tt.filePath, tt.projectRoot, result, tt.expected)
		}
	}
}

func TestBuildToolUsage_Sorting(t *testing.T) {
	counts := map[string]int{
		"Read": 10,
		"Bash": 5,
		"Edit": 8,
	}

	stats := buildToolUsage(counts)
	if len(stats) != 3 {
		t.Fatalf("expected 3 stats, got %d", len(stats))
	}

	// Should be sorted by count descending
	if stats[0].Name != "Read" {
		t.Errorf("expected Read first, got %s", stats[0].Name)
	}
	if stats[1].Name != "Edit" {
		t.Errorf("expected Edit second, got %s", stats[1].Name)
	}
	if stats[2].Name != "Bash" {
		t.Errorf("expected Bash third, got %s", stats[2].Name)
	}

	// Check percentage calculation
	// Read: 10/23 ≈ 43.48%
	if stats[0].Percentage < 43 || stats[0].Percentage > 44 {
		t.Errorf("expected ~43.5%% for Read, got %.1f%%", stats[0].Percentage)
	}
}

func TestCostEstimation(t *testing.T) {
	modelStats := map[string]*modelAccumulator{
		"claude-sonnet-4-5": {
			model:        "claude-sonnet-4-5",
			inputTokens:  1_000_000, // 1M tokens
			outputTokens: 100_000,   // 100K tokens
			turnCount:    10,
		},
	}

	breakdown, cost := buildModelBreakdown(modelStats)

	if len(breakdown) != 1 {
		t.Fatalf("expected 1 model, got %d", len(breakdown))
	}

	// Sonnet: $3/MTok input, $15/MTok output
	// Input cost: 1M * $3/M = $3.00
	// Output cost: 0.1M * $15/M = $1.50
	expectedInput := 3.0
	expectedOutput := 1.5

	if cost.InputCost < expectedInput-0.01 || cost.InputCost > expectedInput+0.01 {
		t.Errorf("expected input cost ~$%.2f, got $%.2f", expectedInput, cost.InputCost)
	}
	if cost.OutputCost < expectedOutput-0.01 || cost.OutputCost > expectedOutput+0.01 {
		t.Errorf("expected output cost ~$%.2f, got $%.2f", expectedOutput, cost.OutputCost)
	}
}

// TestHotFileLimit ensures we cap at hotFileLimit
func TestHotFileLimit(t *testing.T) {
	fileCounts := make(map[string]*types.HotFile)
	for i := range 20 {
		path := "/project/file" + string(rune('A'+i)) + ".go"
		fileCounts[path] = &types.HotFile{
			Path:     path,
			TotalOps: 20 - i, // Descending ops
		}
	}

	files := buildHotFiles(fileCounts)
	if len(files) > hotFileLimit {
		t.Errorf("expected max %d hot files, got %d", hotFileLimit, len(files))
	}
}
