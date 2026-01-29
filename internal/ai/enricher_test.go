package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Priyans-hu/argus/pkg/types"
)

func newTestAnalysis() *types.Analysis {
	return &types.Analysis{
		ProjectName: "test-project",
		TechStack: types.TechStack{
			Languages: []types.Language{
				{Name: "Go", Version: "1.24"},
			},
			Frameworks: []types.Framework{
				{Name: "Cobra", Category: "cli"},
			},
		},
		Structure: types.ProjectStructure{
			Directories: []types.Directory{
				{Path: "cmd", Purpose: "Command entrypoints"},
				{Path: "internal", Purpose: "Private packages"},
			},
		},
		Conventions: []types.Convention{
			{Category: "naming", Description: "Use camelCase"},
		},
	}
}

func TestEnricher_Enrich_Success(t *testing.T) {
	var callCount atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)

		var req generateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var response string
		if strings.Contains(req.Prompt, "project summary") {
			response = "A Go CLI tool for code analysis."
		} else {
			response = `[{"title":"Test insight","description":"This is a test insight."}]`
		}

		resp := generateResponse{Response: response, Done: true}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := &Config{
		Enabled:  true,
		Endpoint: srv.URL,
		Model:    "test-model",
		Timeout:  5 * time.Second,
	}

	enricher := NewEnricher(cfg)
	analysis := newTestAnalysis()

	err := enricher.Enrich(context.Background(), analysis)
	if err != nil {
		t.Fatalf("Enrich failed: %v", err)
	}

	if analysis.AIEnrichment == nil {
		t.Fatal("expected AIEnrichment to be set")
	}

	if analysis.AIEnrichment.Model != "test-model" {
		t.Errorf("expected model 'test-model', got '%s'", analysis.AIEnrichment.Model)
	}

	if analysis.AIEnrichment.ProjectSummary == "" {
		t.Error("expected non-empty project summary")
	}

	// At least some of the enrichment calls should have succeeded
	if callCount.Load() == 0 {
		t.Error("expected at least one API call")
	}
}

func TestEnricher_Enrich_AllFail(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("error"))
	}))
	defer srv.Close()

	cfg := &Config{
		Enabled:  true,
		Endpoint: srv.URL,
		Model:    "test-model",
		Timeout:  5 * time.Second,
	}

	enricher := NewEnricher(cfg)
	analysis := newTestAnalysis()

	err := enricher.Enrich(context.Background(), analysis)
	if err == nil {
		t.Error("expected error when all enrichment calls fail")
	}
}

func TestEnricher_IsAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer srv.Close()

	cfg := &Config{
		Enabled:  true,
		Endpoint: srv.URL,
		Model:    "test-model",
		Timeout:  5 * time.Second,
	}

	enricher := NewEnricher(cfg)
	if !enricher.IsAvailable(context.Background()) {
		t.Error("expected IsAvailable to return true")
	}
}

func TestParseInsights_ValidJSON(t *testing.T) {
	response := `[{"title":"Insight 1","description":"Desc 1"},{"title":"Insight 2","description":"Desc 2"}]`
	insights := parseInsights(response)
	if len(insights) != 2 {
		t.Fatalf("expected 2 insights, got %d", len(insights))
	}
	if insights[0].Title != "Insight 1" {
		t.Errorf("expected 'Insight 1', got %q", insights[0].Title)
	}
}

func TestParseInsights_JSONInText(t *testing.T) {
	response := `Here are some insights:
[{"title":"Wrapped","description":"Found in text"}]
Hope that helps!`
	insights := parseInsights(response)
	if len(insights) != 1 {
		t.Fatalf("expected 1 insight, got %d", len(insights))
	}
	if insights[0].Title != "Wrapped" {
		t.Errorf("expected 'Wrapped', got %q", insights[0].Title)
	}
}

func TestParseInsights_InvalidJSON(t *testing.T) {
	response := "This is not JSON at all"
	insights := parseInsights(response)
	if len(insights) != 0 {
		t.Errorf("expected 0 insights for invalid JSON, got %d", len(insights))
	}
}

func TestParseInsights_EmptyEntries(t *testing.T) {
	response := `[{"title":"Good","description":"Valid"},{"title":"","description":""}]`
	insights := parseInsights(response)
	if len(insights) != 1 {
		t.Fatalf("expected 1 valid insight, got %d", len(insights))
	}
}

func TestConfig_Merge(t *testing.T) {
	cfg := &Config{Enabled: true}
	cfg.Merge()

	if cfg.Endpoint != "http://localhost:11434" {
		t.Errorf("expected default endpoint, got %q", cfg.Endpoint)
	}
	if cfg.Model != "llama3.2" {
		t.Errorf("expected default model, got %q", cfg.Model)
	}
	if cfg.Timeout != 120*time.Second {
		t.Errorf("expected default timeout, got %v", cfg.Timeout)
	}
}

func TestConfig_Merge_PreservesExisting(t *testing.T) {
	cfg := &Config{
		Enabled:  true,
		Endpoint: "http://custom:9999",
		Model:    "mistral",
		Timeout:  60 * time.Second,
	}
	cfg.Merge()

	if cfg.Endpoint != "http://custom:9999" {
		t.Errorf("expected custom endpoint, got %q", cfg.Endpoint)
	}
	if cfg.Model != "mistral" {
		t.Errorf("expected custom model, got %q", cfg.Model)
	}
	if cfg.Timeout != 60*time.Second {
		t.Errorf("expected custom timeout, got %v", cfg.Timeout)
	}
}
