package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/Priyans-hu/argus/pkg/types"
)

// Enricher uses an AI model to generate additional insights about a codebase
type Enricher struct {
	client *Client
	model  string
}

// NewEnricher creates a new AI enricher
func NewEnricher(cfg *Config) *Enricher {
	cfg.Merge()
	client := NewClient(cfg.Endpoint, cfg.Model, cfg.Timeout)
	return &Enricher{
		client: client,
		model:  cfg.Model,
	}
}

// IsAvailable checks if the AI backend is reachable
func (e *Enricher) IsAvailable(ctx context.Context) bool {
	return e.client.IsAvailable(ctx)
}

// Enrich runs all enrichment prompts and attaches results to the analysis
func (e *Enricher) Enrich(ctx context.Context, analysis *types.Analysis) error {
	enrichment := &types.AIEnrichment{
		Model: e.model,
	}

	type result struct {
		name string
		err  error
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	results := make(chan result, 5)

	// Run enrichment calls concurrently
	enrich := func(name string, fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := fn()
			results <- result{name: name, err: err}
		}()
	}

	enrich("summary", func() error {
		prompt := buildProjectSummaryPrompt(analysis)
		resp, err := e.client.Generate(ctx, prompt)
		if err != nil {
			return err
		}
		mu.Lock()
		enrichment.ProjectSummary = strings.TrimSpace(resp)
		mu.Unlock()
		return nil
	})

	enrich("conventions", func() error {
		prompt := buildConventionsPrompt(analysis)
		resp, err := e.client.Generate(ctx, prompt)
		if err != nil {
			return err
		}
		insights := parseInsights(resp)
		mu.Lock()
		enrichment.Conventions = insights
		mu.Unlock()
		return nil
	})

	enrich("architecture", func() error {
		prompt := buildArchitecturePrompt(analysis)
		resp, err := e.client.Generate(ctx, prompt)
		if err != nil {
			return err
		}
		insights := parseInsights(resp)
		mu.Lock()
		enrichment.Architecture = insights
		mu.Unlock()
		return nil
	})

	enrich("best_practices", func() error {
		prompt := buildBestPracticesPrompt(analysis)
		resp, err := e.client.Generate(ctx, prompt)
		if err != nil {
			return err
		}
		insights := parseInsights(resp)
		mu.Lock()
		enrichment.BestPractices = insights
		mu.Unlock()
		return nil
	})

	enrich("patterns", func() error {
		prompt := buildPatternsPrompt(analysis)
		resp, err := e.client.Generate(ctx, prompt)
		if err != nil {
			return err
		}
		insights := parseInsights(resp)
		mu.Lock()
		enrichment.Patterns = insights
		mu.Unlock()
		return nil
	})

	// Wait for all goroutines, then close channel
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results — log failures but don't fail overall
	var succeeded, failed int
	for r := range results {
		if r.err != nil {
			slog.Warn("AI enrichment failed", "call", r.name, "error", r.err)
			failed++
		} else {
			succeeded++
		}
	}

	if succeeded == 0 && failed > 0 {
		return fmt.Errorf("all AI enrichment calls failed")
	}

	analysis.AIEnrichment = enrichment
	return nil
}

// parseInsights extracts JSON insight objects from an AI response
func parseInsights(response string) []types.EnrichedInsight {
	response = strings.TrimSpace(response)

	// Try to find JSON array in the response
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")
	if start == -1 || end == -1 || end <= start {
		return nil
	}

	jsonStr := response[start : end+1]

	var insights []types.EnrichedInsight
	if err := json.Unmarshal([]byte(jsonStr), &insights); err != nil {
		slog.Debug("failed to parse AI insights JSON", "error", err, "response", jsonStr)
		return nil
	}

	// Filter out empty entries
	var valid []types.EnrichedInsight
	for _, i := range insights {
		if i.Title != "" && i.Description != "" {
			valid = append(valid, i)
		}
	}
	return valid
}
