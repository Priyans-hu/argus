package types

import "time"

// UsageInsights holds AI usage data extracted from Claude Code JSONL session logs
type UsageInsights struct {
	SessionCount   int             `json:"session_count"`
	TotalTurns     int             `json:"total_turns"`
	DateRange      DateRange       `json:"date_range"`
	ToolUsage      []ToolUsageStat `json:"tool_usage"`
	HotFiles       []HotFile       `json:"hot_files"`
	PainPoints     []PainPoint     `json:"pain_points,omitempty"`
	TokenUsage     TokenSummary    `json:"token_usage"`
	CostEstimate   CostEstimate    `json:"cost_estimate"`
	ModelBreakdown []ModelUsage    `json:"model_breakdown,omitempty"`
}

// DateRange represents a time period
type DateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ToolUsageStat represents usage statistics for a single tool
type ToolUsageStat struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// HotFile represents a file frequently accessed by AI
type HotFile struct {
	Path       string `json:"path"`
	ReadCount  int    `json:"read_count"`
	EditCount  int    `json:"edit_count"`
	WriteCount int    `json:"write_count"`
	TotalOps   int    `json:"total_ops"`
}

// PainPoint represents a file that causes AI difficulty
type PainPoint struct {
	File        string `json:"file"`
	Type        string `json:"type"` // "repeated_reads", "failed_edits"
	Count       int    `json:"count"`
	Description string `json:"description"`
}

// TokenSummary holds aggregate token counts
type TokenSummary struct {
	InputTokens         int64 `json:"input_tokens"`
	OutputTokens        int64 `json:"output_tokens"`
	CacheCreationTokens int64 `json:"cache_creation_tokens"`
	CacheReadTokens     int64 `json:"cache_read_tokens"`
	TotalTokens         int64 `json:"total_tokens"`
}

// CostEstimate holds estimated costs in USD
type CostEstimate struct {
	InputCost  float64 `json:"input_cost"`
	OutputCost float64 `json:"output_cost"`
	CacheCost  float64 `json:"cache_cost"`
	TotalCost  float64 `json:"total_cost"`
}

// ModelUsage holds per-model usage statistics
type ModelUsage struct {
	Model      string       `json:"model"`
	TokenUsage TokenSummary `json:"token_usage"`
	TurnCount  int          `json:"turn_count"`
}
