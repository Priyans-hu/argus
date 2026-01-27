package usage

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Priyans-hu/argus/pkg/types"
)

// Model pricing per million tokens (USD)
var modelPricing = map[string]struct {
	Input  float64
	Output float64
}{
	"claude-opus-4-5":   {15.0, 75.0},
	"claude-opus-4":     {15.0, 75.0},
	"claude-sonnet-4":   {3.0, 15.0},
	"claude-sonnet-4-5": {3.0, 15.0},
	"claude-haiku-4":    {0.80, 4.0},
	"claude-haiku-4-5":  {0.80, 4.0},
	// Fallback pricing for unrecognized models
	"default": {3.0, 15.0},
}

const (
	hotFileLimit   = 15
	painPointLimit = 10
	// Pain point thresholds
	repeatedReadThreshold = 10
)

// Aggregate combines parsed session data into UsageInsights
func Aggregate(sessions []*SessionData, projectRoot string) *types.UsageInsights {
	if len(sessions) == 0 {
		return nil
	}

	insights := &types.UsageInsights{
		SessionCount: len(sessions),
	}

	toolCounts := make(map[string]int)
	fileCounts := make(map[string]*types.HotFile)
	modelStats := make(map[string]*modelAccumulator)
	var totalTokens types.TokenSummary

	for _, s := range sessions {
		insights.TotalTurns += s.TurnCount

		// Track date range
		if insights.DateRange.Start.IsZero() || (!s.StartTime.IsZero() && s.StartTime.Before(insights.DateRange.Start)) {
			insights.DateRange.Start = s.StartTime
		}
		if s.EndTime.After(insights.DateRange.End) {
			insights.DateRange.End = s.EndTime
		}

		// Aggregate tool usage
		for _, event := range s.ToolEvents {
			toolCounts[event.Name]++

			// Track file interactions
			if event.FilePath != "" {
				relPath := makeRelative(event.FilePath, projectRoot)
				hf, ok := fileCounts[relPath]
				if !ok {
					hf = &types.HotFile{Path: relPath}
					fileCounts[relPath] = hf
				}
				switch event.Name {
				case "Read":
					hf.ReadCount++
				case "Edit":
					hf.EditCount++
				case "Write":
					hf.WriteCount++
				}
				hf.TotalOps++
			}
		}

		// Aggregate token stats
		for _, tr := range s.TokenStats {
			totalTokens.InputTokens += tr.InputTokens
			totalTokens.OutputTokens += tr.OutputTokens
			totalTokens.CacheCreationTokens += tr.CacheCreationTokens
			totalTokens.CacheReadTokens += tr.CacheReadTokens

			acc, ok := modelStats[tr.Model]
			if !ok {
				acc = &modelAccumulator{model: tr.Model}
				modelStats[tr.Model] = acc
			}
			acc.inputTokens += tr.InputTokens
			acc.outputTokens += tr.OutputTokens
			acc.cacheCreationTokens += tr.CacheCreationTokens
			acc.cacheReadTokens += tr.CacheReadTokens
			acc.turnCount++
		}
	}

	totalTokens.TotalTokens = totalTokens.InputTokens + totalTokens.OutputTokens +
		totalTokens.CacheCreationTokens + totalTokens.CacheReadTokens
	insights.TokenUsage = totalTokens

	// Build tool usage stats
	insights.ToolUsage = buildToolUsage(toolCounts)

	// Build hot files (sorted by total ops, limited)
	insights.HotFiles = buildHotFiles(fileCounts)

	// Detect pain points
	insights.PainPoints = detectPainPoints(fileCounts)

	// Build model breakdown and cost estimate
	insights.ModelBreakdown, insights.CostEstimate = buildModelBreakdown(modelStats)

	return insights
}

type modelAccumulator struct {
	model               string
	inputTokens         int64
	outputTokens        int64
	cacheCreationTokens int64
	cacheReadTokens     int64
	turnCount           int
}

func buildToolUsage(counts map[string]int) []types.ToolUsageStat {
	total := 0
	for _, c := range counts {
		total += c
	}
	if total == 0 {
		return nil
	}

	var stats []types.ToolUsageStat
	for name, count := range counts {
		stats = append(stats, types.ToolUsageStat{
			Name:       name,
			Count:      count,
			Percentage: float64(count) / float64(total) * 100,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}

func buildHotFiles(fileCounts map[string]*types.HotFile) []types.HotFile {
	var files []types.HotFile
	for _, hf := range fileCounts {
		files = append(files, *hf)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].TotalOps > files[j].TotalOps
	})

	if len(files) > hotFileLimit {
		files = files[:hotFileLimit]
	}

	return files
}

func detectPainPoints(fileCounts map[string]*types.HotFile) []types.PainPoint {
	var points []types.PainPoint

	for path, hf := range fileCounts {
		// Files read too many times
		if hf.ReadCount >= repeatedReadThreshold {
			points = append(points, types.PainPoint{
				File:        path,
				Type:        "repeated_reads",
				Count:       hf.ReadCount,
				Description: fmt.Sprintf("Read %dx across sessions — consider inlining key info in context", hf.ReadCount),
			})
		}

		// Files with writes followed by edits (suggesting initial write was wrong)
		if hf.WriteCount > 0 && hf.EditCount >= 2 {
			points = append(points, types.PainPoint{
				File:        path,
				Type:        "write_then_edit",
				Count:       hf.EditCount,
				Description: fmt.Sprintf("Written then edited %dx — AI may be struggling with this file's structure", hf.EditCount),
			})
		}
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Count > points[j].Count
	})

	if len(points) > painPointLimit {
		points = points[:painPointLimit]
	}

	return points
}

func buildModelBreakdown(modelStats map[string]*modelAccumulator) ([]types.ModelUsage, types.CostEstimate) {
	var breakdown []types.ModelUsage
	var totalCost types.CostEstimate

	for _, acc := range modelStats {
		total := acc.inputTokens + acc.outputTokens + acc.cacheCreationTokens + acc.cacheReadTokens
		mu := types.ModelUsage{
			Model: acc.model,
			TokenUsage: types.TokenSummary{
				InputTokens:         acc.inputTokens,
				OutputTokens:        acc.outputTokens,
				CacheCreationTokens: acc.cacheCreationTokens,
				CacheReadTokens:     acc.cacheReadTokens,
				TotalTokens:         total,
			},
			TurnCount: acc.turnCount,
		}
		breakdown = append(breakdown, mu)

		// Calculate cost
		pricing, ok := modelPricing[acc.model]
		if !ok {
			pricing = modelPricing["default"]
		}

		inputCost := float64(acc.inputTokens) / 1_000_000 * pricing.Input
		outputCost := float64(acc.outputTokens) / 1_000_000 * pricing.Output
		// Cache creation costs same as input, cache reads are discounted (10% of input)
		cacheCost := float64(acc.cacheCreationTokens)/1_000_000*pricing.Input +
			float64(acc.cacheReadTokens)/1_000_000*pricing.Input*0.1

		totalCost.InputCost += inputCost
		totalCost.OutputCost += outputCost
		totalCost.CacheCost += cacheCost
	}

	totalCost.TotalCost = totalCost.InputCost + totalCost.OutputCost + totalCost.CacheCost

	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].TokenUsage.TotalTokens > breakdown[j].TokenUsage.TotalTokens
	})

	return breakdown, totalCost
}

// makeRelative converts an absolute file path to relative path from projectRoot.
// If the path is already relative or doesn't start with projectRoot, returns as-is.
func makeRelative(filePath, projectRoot string) string {
	if projectRoot == "" {
		return filePath
	}

	// Clean both paths
	filePath = filepath.Clean(filePath)
	projectRoot = filepath.Clean(projectRoot)

	rel, err := filepath.Rel(projectRoot, filePath)
	if err != nil {
		return filePath
	}

	// Don't return paths that go above project root
	if strings.HasPrefix(rel, "..") {
		return filePath
	}

	return rel
}
